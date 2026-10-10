package dictionary

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// SQLCache uses the application's SQLite connection; its owner closes the DB.
// Responses are shared across users, while entries form a retained source archive.
type SQLCache struct{ db *sql.DB }

const persistentTTL = 30 * 24 * time.Hour
const maxPersistentResponses = 5000

func NewSQLCache(db *sql.DB) (*SQLCache, error) {
	if db == nil {
		return nil, fmt.Errorf("dictionary persistence requires a database")
	}
	for _, statement := range []string{
		`CREATE TABLE IF NOT EXISTS dictionary_responses(cache_key TEXT PRIMARY KEY, source TEXT NOT NULL, body BLOB NOT NULL, fetched_at INTEGER NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS dictionary_responses_age ON dictionary_responses(fetched_at)`,
		`CREATE TABLE IF NOT EXISTS dictionary_entries(source TEXT NOT NULL, entry_id TEXT NOT NULL, kind TEXT NOT NULL, body BLOB NOT NULL, fetched_at INTEGER NOT NULL, PRIMARY KEY(source,entry_id))`,
	} {
		if _, err := db.Exec(statement); err != nil {
			return nil, fmt.Errorf("initialize dictionary persistence: %w", err)
		}
	}
	return &SQLCache{db: db}, nil
}
func responseKey(endpoint string) string {
	sum := sha256.Sum256([]byte(endpoint))
	return hex.EncodeToString(sum[:])
}
func (s *SQLCache) load(ctx context.Context, endpoint string) (cached, bool, error) {
	var hit cached
	var at int64
	err := s.db.QueryRowContext(ctx, `SELECT body,fetched_at FROM dictionary_responses WHERE cache_key=?`, responseKey(endpoint)).Scan(&hit.body, &at)
	if err == sql.ErrNoRows {
		return hit, false, nil
	}
	if err != nil {
		return hit, false, err
	}
	hit.at = time.UnixMilli(at)
	return hit, true, nil
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO dictionary_responses(cache_key,source,body,fetched_at) VALUES(?,?,?,?) ON CONFLICT(cache_key) DO UPDATE SET body=excluded.body,fetched_at=excluded.fetched_at`, responseKey(endpoint), source, hit.body, hit.at.UnixMilli()); err != nil {
		return err
	}
	for _, entry := range entries {
		if _, err = tx.ExecContext(ctx, `INSERT INTO dictionary_entries(source,entry_id,kind,body,fetched_at) VALUES(?,?,?,?,?) ON CONFLICT(source,entry_id) DO UPDATE SET body=excluded.body,fetched_at=excluded.fetched_at`, source, entry.id, kind, []byte(entry.body), hit.at.UnixMilli()); err != nil {
			return err
		}
	}
	// Limit query combinations in the response cache; collected entries are retained.
	if _, err = tx.ExecContext(ctx, `DELETE FROM dictionary_responses WHERE cache_key IN (SELECT cache_key FROM dictionary_responses ORDER BY fetched_at DESC,cache_key LIMIT -1 OFFSET ?)`, maxPersistentResponses); err != nil {
		return err
	}
	return tx.Commit()
}
