package dictionary

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/url"
	"nihongo-master-backend/internal/database"
	"strconv"
	"strings"
	"time"
)

// SQLCache uses the application's MariaDB connection; its owner closes the DB.
// Responses are shared across users, while entries form a retained source archive.
type SQLCache struct{ db *gorm.DB }

func NewSQLCache(db *gorm.DB) (*SQLCache, error) {
	if db == nil {
		return nil, fmt.Errorf("dictionary persistence requires a database")
	}
	if err := database.Migrate(db, &database.DictionaryResponse{}, &database.DictionaryEntry{}); err != nil {
		return nil, fmt.Errorf("initialize GORM dictionary schema: %w", err)
	}

	return &SQLCache{db: db}, nil
}
func responseKey(endpoint string) string {
	sum := sha256.Sum256([]byte(endpoint))
	return hex.EncodeToString(sum[:])
}
func (s *SQLCache) load(ctx context.Context, endpoint string) (cached, bool, error) {
	var record database.DictionaryResponse
	err := s.db.WithContext(ctx).Where("cache_key = ?", responseKey(endpoint)).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cached{}, false, nil
	}
	if err != nil {
		return cached{}, false, err
	}
	return cached{body: record.Body, at: time.UnixMilli(record.FetchedAt)}, true, nil
}

type archiveEntry struct {
	id   string
	body json.RawMessage
}

// Validate before caching, so a 200 response containing an error never enters DB.
func responseEntries(endpoint string, body []byte) (source, kind string, entries []archiveEntry, err error) {
	if len(body) > 1<<20 || !json.Valid(body) {
		return "", "", nil, ErrUnavailable
	}
	if strings.HasPrefix(endpoint, "https://jisho.org/api/v1/search/words?") {
		var data struct {
			Meta struct {
				Status int `json:"status"`
			} `json:"meta"`
			Data []json.RawMessage `json:"data"`
		}
		if json.Unmarshal(body, &data) != nil || data.Meta.Status != 200 || data.Data == nil {
			return "", "", nil, ErrUnavailable
		}
		for _, raw := range data.Data {
			var entry struct {
				Slug     string            `json:"slug"`
				Japanese []json.RawMessage `json:"japanese"`
				Senses   []json.RawMessage `json:"senses"`
			}
			if json.Unmarshal(raw, &entry) != nil || entry.Slug == "" || len(entry.Japanese) == 0 || len(entry.Senses) == 0 {
				return "", "", nil, ErrUnavailable
			}
			entries = append(entries, archiveEntry{id: entry.Slug, body: raw})
		}
		return "jisho.org", "word", entries, nil
	}
	if strings.HasPrefix(endpoint, "https://kanjiapi.dev/v1/kanji/") {
		character, e := url.PathUnescape(strings.TrimPrefix(endpoint, "https://kanjiapi.dev/v1/kanji/"))
		var entry struct {
			Kanji   string `json:"kanji"`
			Strokes int    `json:"stroke_count"`
		}
		if e != nil || json.Unmarshal(body, &entry) != nil || entry.Kanji != character || entry.Strokes < 1 {
			return "", "", nil, ErrUnavailable
		}
		return "kanjiapi.dev", "kanji", []archiveEntry{{id: character, body: body}}, nil
	}
	return "", "", nil, ErrInvalid
}
func (s *SQLCache) save(ctx context.Context, endpoint string, hit cached) error {
	source, kind, entries, err := responseEntries(endpoint, hit.body)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		response := database.DictionaryResponse{CacheKey: responseKey(endpoint), Source: source, Body: hit.body, FetchedAt: hit.at.UnixMilli()}
		response.Endpoint = endpoint
		u, _ := url.Parse(endpoint)
		if kind == "word" {
			response.QueryText = u.Query().Get("keyword")
			response.Page, _ = strconv.Atoi(u.Query().Get("page"))
			parts := strings.Split(u.Fragment, "|")
			if len(parts) == 4 {
				response.Mode = parts[1]
				response.JLPT = parts[2]
				response.Common = parts[3] == "true"
				if response.Common {
					response.QueryText = strings.TrimSuffix(response.QueryText, " #common")
				}
				if response.JLPT != "" {
					response.QueryText = strings.TrimSuffix(response.QueryText, " #jlpt-"+strings.ToLower(response.JLPT))
				}
				if response.Mode == "english" {
					response.QueryText = strings.Trim(response.QueryText, "\"")
				}
			} else {
				response.Mode = "raw"
			}
		} else {
			response.Mode = "kanji"
			response.QueryText = strings.TrimPrefix(u.Path, "/v1/kanji/")
		}
		upsert := clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"body", "fetched_at"})}
		responseUpsert := clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"body", "fetched_at", "endpoint", "query_text", "mode", "page", "jlpt", "common"})}
		if err := tx.Clauses(responseUpsert).Create(&response).Error; err != nil {
			return err
		}
		for _, entry := range entries {
			record := database.DictionaryEntry{Source: source, EntryID: entry.id, Kind: kind, Body: []byte(entry.body), FetchedAt: hit.at.UnixMilli()}
			if err := tx.Clauses(upsert).Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

type CacheRecord struct {
	ID        string `json:"id"`
	QueryText string `json:"query"`
	Mode      string `json:"mode"`
	Page      int    `json:"page"`
	JLPT      string `json:"jlpt"`
	Common    bool   `json:"common"`
	Source    string `json:"source"`
	FetchedAt int64  `json:"fetchedAt"`
}

func (s *SQLCache) List(ctx context.Context, page int, query string) ([]CacheRecord, int64, error) {
	db := s.db.WithContext(ctx).Model(&database.DictionaryResponse{})
	if query != "" {
		db = db.Where("query_text LIKE ?", "%"+query+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []CacheRecord
	err := db.Select("id, query_text, mode, page, jlpt, common, source, fetched_at").Order("fetched_at DESC, id").Limit(50).Offset((page - 1) * 50).Scan(&records).Error
	if records == nil {
		records = []CacheRecord{}
	}
	return records, total, err
}

func (c *Client) Refresh(ctx context.Context, id string) error {
	if c.Persistence == nil {
		return ErrUnavailable
	}
	var record database.DictionaryResponse
	err := c.Persistence.db.WithContext(ctx).Select("endpoint").Where("id = ?", id).Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return ErrUnavailable
	}
	u, err := url.Parse(record.Endpoint)
	if err != nil || u.Scheme != "https" || u.User != nil || !((u.Host == "jisho.org" && u.Path == "/api/v1/search/words") || (u.Host == "kanjiapi.dev" && strings.HasPrefix(u.Path, "/v1/kanji/"))) {
		return ErrInvalid
	}
	_, err = c.getRequest(ctx, record.Endpoint, true)
	return err
}
