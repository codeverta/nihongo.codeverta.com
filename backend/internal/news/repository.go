package news

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"
	_ "time/tzdata"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/database"
)

var ErrInvalid = errors.New("invalid news input")
var ErrNotFound = errors.New("news not found")
var ErrBusy = errors.New("news source is synchronizing or was just checked")
var ErrConflict = errors.New("news changed while being edited")

type FetchResult struct {
	Articles           []Article
	ETag, LastModified string
	NotModified        bool
	Skipped            int
}
type Fetcher interface {
	Fetch(context.Context, Source, time.Time) (FetchResult, error)
}
type Repository struct {
	db           *gorm.DB
	now          func() time.Time
	errorHandler func(error)
}

func (r *Repository) SetErrorHandler(handler func(error)) { r.errorHandler = handler }

func (r *Repository) reportError(err error) {
	if err != nil && r.errorHandler != nil {
		r.errorHandler(err)
	}
}

func New(db *gorm.DB) (*Repository, error) {
	if db == nil {
		return nil, errors.New("news requires MariaDB")
	}
	if err := database.Migrate(db, &Article{}, &Source{}); err != nil {
		return nil, err
	}
	r := &Repository{db: db, now: time.Now}
	for _, s := range []Source{
		{ID: "global-voices", Name: "Global Voices 日本語", Kind: "regular", URL: "https://jp.globalvoices.org/feed/", Timezone: "Asia/Jakarta", Days: []int{2, 5}, Hour: 18, Enabled: true},
		{ID: "osaka", Name: "大阪市", Kind: "easy", URL: "https://www.city.osaka.lg.jp/shisei/category/3054-1-2-21-3-4-0-0-0-0.html", Timezone: "Asia/Jakarta", Days: []int{0, 4}, Hour: 18, Enabled: true},
	} {
		s.NextFetchAt = time.Unix(0, 0).UTC() // One initial import; survives subsequent restarts.
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&s).Error; err != nil {
			return nil, err
		}
	}
	return r, nil
}

// NextScheduled returns the first future wall-clock occurrence, including DST.
func NextScheduled(s Source, after time.Time) (time.Time, error) {
	if s.Timezone == "Local" || s.Timezone == "" || len(s.Timezone) > 80 {
		return time.Time{}, ErrInvalid
	}
	if len(s.Days) < 1 || len(s.Days) > 2 || s.Hour < 0 || s.Hour > 23 || s.Minute < 0 || s.Minute > 59 {
		return time.Time{}, ErrInvalid
	}
	loc, err := time.LoadLocation(s.Timezone)
	if err != nil {
		return time.Time{}, ErrInvalid
	}
	seen := map[int]bool{}
	for _, d := range s.Days {
		if d < 0 || d > 6 || seen[d] {
			return time.Time{}, ErrInvalid
		}
		seen[d] = true
	}
	t := after.In(loc)
	for i := 0; i < 9; i++ {
		day := t.AddDate(0, 0, i)
		candidate := time.Date(day.Year(), day.Month(), day.Day(), s.Hour, s.Minute, 0, 0, loc)
		if candidate.Hour() == s.Hour && candidate.Minute() == s.Minute && seen[int(candidate.Weekday())] && candidate.After(after) {
			return candidate.UTC(), nil
		}
	}
	return time.Time{}, ErrInvalid
}
func (r *Repository) Sources(ctx context.Context) ([]Source, error) {
	rows := []Source{}
	err := r.db.WithContext(ctx).Order("id").Find(&rows).Error
	for i := range rows {
		s := &rows[i]
		s.Status = "idle"
		if s.LastSuccessAt != nil {
			s.Status = "ready"
		}
		if s.LastError != "" {
			s.Status = "error"
		}
		if s.LeaseOwner != "" {
			s.Status = "interrupted"
			if s.LeaseUntil != nil && s.LeaseUntil.After(r.now()) {
				s.Status = "running"
			}
		}
	}
	return rows, err
}
func (r *Repository) SetSchedule(ctx context.Context, id string, input Schedule) error {
	var source Source
	if err := r.db.WithContext(ctx).First(&source, "id = ?", id).Error; err != nil {
		return ErrNotFound
	}
	source.Timezone = strings.TrimSpace(input.Timezone)
	source.Days = input.Days
	source.Hour = input.Hour
	source.Minute = input.Minute
	source.Enabled = input.Enabled
	next, err := NextScheduled(source, r.now())
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&source).Select("Timezone", "Days", "Hour", "Minute", "Enabled", "NextFetchAt").Updates(&Source{Timezone: source.Timezone, Days: source.Days, Hour: source.Hour, Minute: source.Minute, Enabled: source.Enabled, NextFetchAt: next}).Error
}
func (r *Repository) List(ctx context.Context, page, size int, source string, admin bool) (Page, error) {
	if page < 1 || page > 100000 || size < 1 || size > 50 {
		return Page{}, ErrInvalid
	}
	result := Page{Items: []Article{}, Page: page, PageSize: size}
	q := r.db.WithContext(ctx).Model(&Article{}).Where("deleted = ?", false)
	if !admin {
		q = q.Where("hidden = ?", false)
	}
	if source != "" {
		q = q.Where("source_id = ?", source)
	}
	if err := q.Count(&result.Total).Error; err != nil {
		return result, err
	}
	if err := q.Omit("paragraphs", "references").Order("published_at DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&result.Items).Error; err != nil {
		return result, err
	}
	for i := range result.Items {
		result.Items[i].OriginalDate = sourceDate(result.Items[i].PublishedAt)
	}
	var err error
	result.Sources, err = r.Sources(ctx)
	if !admin {
		for i := range result.Sources {
			result.Sources[i].LastError = ""
		}
	}
	return result, err
}
func (r *Repository) Detail(ctx context.Context, id string, admin bool) (Article, error) {
	var a Article
	q := r.db.WithContext(ctx).Where("id = ? AND deleted = ?", id, false)
	if !admin {
		q = q.Where("hidden = ?", false)
	}
	err := q.First(&a).Error
	a.OriginalDate = sourceDate(a.PublishedAt)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrNotFound
	}
	return a, err
}

type ArticleEdit struct {
	Revision     string        `json:"revision"`
	Hidden       bool          `json:"hidden"`
	ImageURL     string        `json:"imageUrl"`
	ImageCredit  string        `json:"imageCredit"`
	ImageLicense string        `json:"imageLicense"`
	Readings     []ReadingEdit `json:"readings"`
}
type ReadingEdit struct {
	Paragraph int    `json:"paragraph"`
	Token     int    `json:"token"`
	Reading   string `json:"reading"`
} // Paragraph -1 = title.
func (r *Repository) Edit(ctx context.Context, id string, input ArticleEdit) error {
	if len(input.Readings) > 1000 || len(input.ImageCredit) > 8000 || len(input.ImageLicense) > 200 {
		return ErrInvalid
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var a Article
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "id = ? AND deleted = ?", id, false).Error; err != nil {
			return ErrNotFound
		}
		if len(input.Readings) > 0 && input.Revision != a.ContentHash {
			return ErrConflict
		}
		for _, change := range input.Readings {
			if utf8.RuneCountInString(change.Reading) > 100 || strings.ContainsAny(change.Reading, "<>\r\n\x00") {
				return ErrInvalid
			}
			tokens := a.TitleTokens
			if change.Paragraph >= 0 {
				if change.Paragraph >= len(a.Paragraphs) {
					return ErrInvalid
				}
				tokens = a.Paragraphs[change.Paragraph].Tokens
			} else if change.Paragraph != -1 {
				return ErrInvalid
			}
			if change.Token < 0 || change.Token >= len(tokens) {
				return ErrInvalid
			}
			tokens[change.Token].Reading = strings.TrimSpace(change.Reading)
			tokens[change.Token].ReadingSource = "editorial"
		}
		a.Hidden = input.Hidden
		if a.ImageURL != input.ImageURL {
			a.ImageSourceURL = ""
			a.ImageLicenseURL = ""
		}
		a.ImageURL = input.ImageURL
		a.ImageCredit = input.ImageCredit
		a.ImageLicense = input.ImageLicense
		return tx.Save(&a).Error
	})
}
func (r *Repository) Delete(ctx context.Context, id string) error {
	// Tombstone prevents a deleted URL from being silently republished next sync.
	res := r.db.WithContext(ctx).Model(&Article{}).Where("id = ? AND deleted = ?", id, false).Updates(map[string]any{"deleted": true, "hidden": true, "paragraphs": "[]", "title_tokens": "[]", "references": "[]"})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *Repository) Run(ctx context.Context, f Fetcher) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if r.errorHandler == nil {
				panic(recovered)
			}
			r.reportError(fmt.Errorf("panic: %v\n%s", recovered, debug.Stack()))
		}
	}()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		sources, err := r.Sources(ctx)
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("news scheduler: %v", err)
				r.reportError(err)
			}
		} else {
			for _, s := range sources {
				if s.Enabled && !s.NextFetchAt.After(r.now()) {
					if _, err := r.Sync(ctx, f, s.ID, false); err != nil && !errors.Is(err, ErrBusy) && ctx.Err() == nil {
						log.Printf("news %s: %v", s.ID, err)
						r.reportError(fmt.Errorf("news source %s: %w", s.ID, err))
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (r *Repository) Sync(ctx context.Context, f Fetcher, id string, force bool) (int, error) {
	var s Source
	owner := uuid.NewString()
	now := r.now().UTC()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&s, "id = ?", id).Error; err != nil {
			return ErrNotFound
		}
		if (s.LeaseUntil != nil && s.LeaseUntil.After(now)) || (!force && (!s.Enabled || s.NextFetchAt.After(now))) || (force && s.LastStartedAt != nil && now.Sub(*s.LastStartedAt) < time.Minute) {
			return ErrBusy
		}
		next, err := NextScheduled(s, now)
		if err != nil {
			return err
		}
		until := now.Add(10 * time.Minute)
		return tx.Model(&s).Updates(map[string]any{"lease_owner": owner, "lease_until": until, "last_started_at": now, "next_fetch_at": next}).Error
	})
	if err != nil {
		return 0, err
	}
	workCtx, cancel := context.WithTimeout(ctx, 8*time.Minute)
	defer cancel()
	if force {
		s.ETag = ""
		s.LastModified = ""
	}
	result, fetchErr := f.Fetch(workCtx, s, now.AddDate(0, -6, 0))
	if fetchErr == nil && !result.NotModified {
		fetchErr = r.db.WithContext(workCtx).Transaction(func(tx *gorm.DB) error {
			var current Source
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, "id = ?", id).Error; err != nil {
				return err
			}
			if current.LeaseOwner != owner || current.LeaseUntil == nil || !current.LeaseUntil.After(r.now()) {
				return ErrBusy
			}
			if len(result.Articles) > 20 {
				return fmt.Errorf("news import exceeds batch bound")
			}
			for _, a := range result.Articles {
				if err := validateArticle(a, s, now); err != nil {
					return err
				}
				a.SourceID = s.ID
				a.SourceName = s.Name
				a.Kind = s.Kind
				a.FetchedAt = now
				raw, _ := json.Marshal(struct {
					Title      string
					Paragraphs []Paragraph
				}{a.Title, a.Paragraphs})
				sum := sha256.Sum256(raw)
				a.ContentHash = hex.EncodeToString(sum[:])
				var old Article
				err := tx.First(&old, "url = ?", a.URL).Error
				if err == nil {
					if old.Deleted {
						continue
					}
					if old.SourceID != id {
						return ErrInvalid
					}
					a.ID = old.ID
					a.Hidden = old.Hidden
					if old.ImageURL != "" {
						a.ImageURL = old.ImageURL
						a.ImageCredit = old.ImageCredit
						a.ImageLicense = old.ImageLicense
						a.ImageSourceURL = old.ImageSourceURL
						a.ImageLicenseURL = old.ImageLicenseURL
					}
					if a.ContentHash == old.ContentHash {
						a.Paragraphs = old.Paragraphs
						a.TitleTokens = old.TitleTokens
					} else {
						preserveReadings(&a, old)
					}
				} else if errors.Is(err, gorm.ErrRecordNotFound) {
					a.ID = uuid.NewString()
				} else {
					return err
				}
				if err := tx.Save(&a).Error; err != nil {
					return err
				}
			}
			return nil
		})
	}
	// Use a bounded detached context to release the lease even on cancellation.
	finishCtx, finishCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer finishCancel()
	updates := map[string]any{"lease_owner": "", "lease_until": nil}
	if fetchErr != nil {
		message := fetchErr.Error()
		if len(message) > 2000 {
			message = message[:2000]
		}
		updates["last_error"] = message
	} else {
		updates["last_error"] = ""
		updates["last_success_at"] = now
		updates["last_imported"] = len(result.Articles)
		updates["last_skipped"] = result.Skipped
		if !result.NotModified {
			updates["e_tag"] = result.ETag
			updates["last_modified"] = result.LastModified
		}
	}
	finishErr := r.db.WithContext(finishCtx).Model(&Source{}).Where("id = ? AND lease_owner = ?", id, owner).Updates(updates).Error
	if fetchErr != nil {
		return 0, fetchErr
	}
	return len(result.Articles), finishErr
}
func preserveReadings(a *Article, old Article) {
	if a.Title == old.Title {
		preserveEditorial(a.TitleTokens, old.TitleTokens)
	}
	for i := range a.Paragraphs {
		for _, p := range old.Paragraphs {
			if a.Paragraphs[i].Text == p.Text {
				preserveEditorial(a.Paragraphs[i].Tokens, p.Tokens)
				break
			}
		}
	}
}

// A date-only Osaka notice at midnight JST must not become yesterday in WIB.
func sourceDate(at time.Time) string {
	return at.In(time.FixedZone("JST", 9*3600)).Format("2006-01-02")
}
func preserveEditorial(tokens, old []Token) {
	for i := range tokens {
		for _, previous := range old {
			if previous.ReadingSource == "editorial" && previous.Start == tokens[i].Start && previous.End == tokens[i].End && previous.Surface == tokens[i].Surface {
				tokens[i].Reading = previous.Reading
				tokens[i].ReadingSource = "editorial"
				break
			}
		}
	}
}
func validateArticle(a Article, s Source, now time.Time) error {
	if !allowedArticleURL(s.ID, a.URL) || a.Title == "" || utf8.RuneCountInString(a.Title) > 500 || len(a.Paragraphs) < 1 || len(a.Paragraphs) > 400 || a.PublishedAt.Before(now.AddDate(0, -6, 0)) || a.PublishedAt.After(now.Add(24*time.Hour)) || a.License == "" {
		return ErrInvalid
	}
	n := 0
	for _, p := range a.Paragraphs {
		n += utf8.RuneCountInString(p.Text)
	}
	if n < 20 || n > 60000 {
		return ErrInvalid
	}
	return nil
}

// HasImage reports whether the stored article already has an image, so imports
// do not request it from the publisher again.
func (r *Repository) HasImage(ctx context.Context, articleURL string) bool {
	var count int64
	err := r.db.WithContext(ctx).Model(&Article{}).Where("url = ? AND image_url <> ''", articleURL).Count(&count).Error
	return err == nil && count > 0
}
