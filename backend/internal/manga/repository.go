package manga

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/database"
)

type Repository struct {
	db     *gorm.DB
	now    func() time.Time
	images ImageStore
}

// Seed only editorially selected non-ecchi titles, never the upstream homepage.
// Japanese page language must be reviewed before adding further sources.
var seedURLs = []string{
	"https://rawkuma.net/manga/ruri-dragon/",
	"https://rawkuma.net/manga/ao-no-hako/",
	"https://rawkuma.net/manga/medalist/",
}

func New(db *gorm.DB, images ImageStore) (*Repository, error) {
	if db == nil {
		return nil, errors.New("manga requires MariaDB")
	}
	if err := database.Migrate(db, &Series{}, &Chapter{}, &Asset{}, &syncState{}); err != nil {
		return nil, err
	}
	r := &Repository{db: db, images: images, now: time.Now}
	epoch := time.Unix(1, 0).UTC()
	for _, raw := range seedURLs {
		s := Series{ID: ID(raw), URL: raw, Language: "ja", Enabled: true, Reviewed: true, Status: "pending", Genres: []string{}, NextFetchAt: epoch}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&s).Error; err != nil {
			return nil, err
		}
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&syncState{ID: "rawkuma", LeaseUntil: epoch, NextRunAt: epoch}).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Repository) CacheEnabled() bool { return r.images != nil }
func visible(db *gorm.DB) *gorm.DB {
	return db.Where("enabled = ? AND reviewed = ? AND language = ? AND status = ?", true, true, "ja", "ready")
}

func (r *Repository) List(ctx context.Context, q, genre string, page, size int) (Catalog, error) {
	result := Catalog{Items: []Series{}, Page: page, PageSize: size, SourceName: "Rawkuma", SourceURL: "https://rawkuma.net/", CacheEnabled: r.CacheEnabled()}
	if page < 1 || page > 10000 || size < 1 || size > 40 || len(q) > 100 || len(genre) > 50 {
		return result, ErrInvalid
	}
	db := visible(r.db.WithContext(ctx).Model(&Series{}))
	if q != "" {
		q = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(q, `\`, `\\`), "%", `\%`), "_", `\_`)
		db = db.Where("LOWER(title) LIKE LOWER(?) OR LOWER(japanese_title) LIKE LOWER(?)", "%"+q+"%", "%"+q+"%")
	}
	if genre != "" {
		db = db.Where("JSON_CONTAINS(genres, JSON_QUOTE(?))", genre)
	}
	if err := db.Count(&result.Total).Error; err != nil {
		return result, err
	}
	err := db.Order("title ASC").Limit(size).Offset((page - 1) * size).Find(&result.Items).Error
	return result, err
}

func (r *Repository) Detail(ctx context.Context, id string) (Detail, error) {
	result := Detail{Chapters: []Chapter{}, SourceName: "Rawkuma"}
	if err := visible(r.db.WithContext(ctx)).First(&result.Series, "id = ?", id).Error; err != nil {
		return result, notFound(err)
	}
	err := r.db.WithContext(ctx).Select("id", "series_id", "url", "title", "number", "position", "status", "next_attempt_at", "fetched_at").Where("series_id = ?", id).Order("position ASC, id ASC").Find(&result.Chapters).Error
	for i := range result.Chapters {
		result.Chapters[i].Pages = []Page{}
	}
	return result, err
}

func (r *Repository) Chapter(ctx context.Context, id string) (Chapter, error) {
	var ch Chapter
	if err := r.db.WithContext(ctx).First(&ch, "id = ?", id).Error; err != nil {
		return ch, notFound(err)
	}
	var count int64
	if err := visible(r.db.WithContext(ctx).Model(&Series{})).Where("id = ?", ch.SeriesID).Count(&count).Error; err != nil {
		return ch, err
	}
	if count == 0 {
		return Chapter{}, ErrNotFound
	}
	if ch.Status != "ready" || ch.Pages == nil {
		ch.Pages = []Page{}
	}
	return ch, nil
}

// Prepare only queues an existing chapter. No network I/O occurs on any API path.
// The shared queue is bounded and each ID is idempotent across all users.
func (r *Repository) Prepare(ctx context.Context, id string) (Chapter, error) {
	if !r.CacheEnabled() {
		return Chapter{}, errors.New("manga COS cache not configured")
	}
	var result Chapter
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lock syncState
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&lock, "id = ?", "rawkuma").Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&result, "id = ?", id).Error; err != nil {
			return notFound(err)
		}
		var parent Series
		if err := visible(tx).First(&parent, "id = ?", result.SeriesID).Error; err != nil {
			return notFound(err)
		}
		if result.Status == "ready" || result.Status == "queued" || result.Status == "processing" || result.NextAttemptAt.After(r.now()) {
			return nil
		}
		var count int64
		if err := tx.Model(&Chapter{}).Where("status = ?", "queued").Count(&count).Error; err != nil {
			return err
		}
		if count >= 64 {
			return ErrBusy
		}
		now := r.now().UTC()
		result.Status = "queued"
		result.RequestedAt = &now
		return tx.Model(&Chapter{}).Where("id = ?", id).Updates(map[string]any{"status": "queued", "requested_at": now, "failure_count": 0}).Error
	})
	if result.Pages == nil {
		result.Pages = []Page{}
	}
	return result, err
}

func (r *Repository) AdminList(ctx context.Context) ([]AdminSeries, error) {
	items := []Series{}
	err := r.db.WithContext(ctx).Order("title ASC").Find(&items).Error
	result := make([]AdminSeries, 0, len(items))
	for _, item := range items {
		result = append(result, AdminSeries{Series: item, ImportError: item.LastError})
	}
	return result, err
}
func (r *Repository) Submit(ctx context.Context, in Submission) error {
	if !ValidSeriesURL(in.URL) || (in.Enabled && (!in.JapaneseVerified || !in.NonAdultReviewed)) {
		return ErrInvalid
	}
	id := ID(in.URL)
	epoch := time.Unix(1, 0).UTC()
	s := Series{ID: id, URL: in.URL, Language: "ja", Enabled: in.Enabled, Reviewed: in.JapaneseVerified && in.NonAdultReviewed, Status: "pending", Genres: []string{}, NextFetchAt: epoch}
	// Editors may hide a series immediately; enabling never bypasses the durable
	// seven-day deadline, including across process restarts and admin submissions.
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&s).Error; err != nil {
			return err
		}
		return tx.Model(&Series{}).Where("id = ?", id).Updates(map[string]any{"enabled": in.Enabled, "reviewed": s.Reviewed}).Error
	})
}

func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (r *Repository) Run(ctx context.Context, f Fetcher) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		work, cancel := context.WithTimeout(ctx, 15*time.Minute)
		if err := r.Sync(work, f); err != nil && !errors.Is(err, ErrBusy) && ctx.Err() == nil {
			log.Printf("manga sync: %v", err)
		}
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// Sync serializes the whole source across API replicas/CLI workers. Attempt
// deadlines are persisted BEFORE HTTP, so crashes cannot trigger repeat fetches.
func (r *Repository) Sync(ctx context.Context, f Fetcher) error {
	if r.images == nil {
		return errors.New("manga COS cache not configured")
	}
	now := r.now().UTC()
	owner := uuid.NewString()
	claimed := r.db.WithContext(ctx).Model(&syncState{}).Where("id = ? AND lease_until <= ? AND next_run_at <= ?", "rawkuma", now, now).Updates(map[string]any{"lease_owner": owner, "lease_until": now.Add(20 * time.Minute), "next_run_at": now.Add(5 * time.Minute)})
	if claimed.Error != nil {
		return claimed.Error
	}
	if claimed.RowsAffected == 0 {
		return ErrBusy
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r.db.WithContext(cleanup).Model(&syncState{}).Where("id = ? AND lease_owner = ?", "rawkuma", owner).Updates(map[string]any{"lease_until": r.now().UTC(), "lease_owner": ""})
	}()
	if reset, ok := f.(interface{ ResetRobots() }); ok {
		reset.ResetRobots()
	}
	// A previous worker may have crashed after claiming a chapter. Keep the
	// cooldown and expose failure rather than leaving readers polling forever.
	if err := r.db.WithContext(ctx).Model(&Chapter{}).Where("status = ?", "processing").Updates(map[string]any{"status": "failed", "last_error": "previous import interrupted"}).Error; err != nil {
		return err
	}
	var sources []Series
	if err := r.db.WithContext(ctx).Where("enabled = ? AND reviewed = ? AND next_fetch_at <= ?", true, true, now).Order("next_fetch_at ASC").Limit(8).Find(&sources).Error; err != nil {
		return err
	}
	var failures []error
	for _, s := range sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		res := r.db.WithContext(ctx).Model(&Series{}).Where("id = ? AND next_fetch_at <= ? AND enabled = ?", s.ID, now, true).Update("next_fetch_at", now.Add(RefreshInterval))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			continue
		}
		if err := r.importSeries(ctx, f, s, now); err != nil {
			status := s.Status
			if status != "ready" {
				status = "failed"
			}
			if errors.Is(err, ErrBlocked) {
				status = "blocked"
			}
			if e := r.db.WithContext(ctx).Model(&Series{}).Where("id = ?", s.ID).Updates(map[string]any{"status": status, "last_error": err.Error()}).Error; e != nil {
				return e
			}
			failures = append(failures, fmt.Errorf("series %s: %w", s.ID, err))
		}
	}
	var chapters []Chapter
	const eligible = "(status = 'queued' OR (status = 'failed' AND JSON_LENGTH(source_pages) > 0 AND failure_count < 3))"
	if err := r.db.WithContext(ctx).Where(eligible).Where("next_attempt_at <= ?", now).Order("requested_at ASC").Limit(2).Find(&chapters).Error; err != nil {
		return err
	}
	for _, ch := range chapters {
		var parent Series
		if err := visible(r.db.WithContext(ctx)).First(&parent, "id = ?", ch.SeriesID).Error; err != nil {
			if e := r.db.WithContext(ctx).Model(&Chapter{}).Where("id = ?", ch.ID).Update("status", "available").Error; e != nil {
				return e
			}
			continue
		}
		res := r.db.WithContext(ctx).Model(&Chapter{}).Where("id = ? AND next_attempt_at <= ?", ch.ID, now).Where(eligible).Updates(map[string]any{"status": "processing", "next_attempt_at": now.Add(RefreshInterval)})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			continue
		}
		if err := r.importChapter(ctx, f, ch, now); err != nil {
			var failed Chapter
			if e := r.db.WithContext(ctx).Select("source_pages", "failure_count").First(&failed, "id = ?", ch.ID).Error; e != nil {
				return e
			}
			retryAt := now.Add(RefreshInterval)
			var upstream HTTPError
			denied := errors.As(err, &upstream) && upstream.Status >= 400 && upstream.Status < 500 && upstream.Status != 408
			// With a durable manifest, transient page/COS failures can resume
			// missing assets without requesting the chapter detail a second time.
			if len(failed.SourcePages) > 0 && !denied && failed.FailureCount < 2 {
				backoff := 5 * time.Minute
				if failed.FailureCount == 1 {
					backoff = 30 * time.Minute
				}
				retryAt = now.Add(backoff)
			}
			if e := r.db.WithContext(ctx).Model(&Chapter{}).Where("id = ?", ch.ID).Updates(map[string]any{"status": "failed", "last_error": err.Error(), "next_attempt_at": retryAt, "failure_count": failed.FailureCount + 1}).Error; e != nil {
				return e
			}
			failures = append(failures, fmt.Errorf("chapter %s: %w", ch.ID, err))
		}
	}
	if len(failures) > 0 {
		return errors.Join(failures...)
	}
	return r.db.WithContext(ctx).Model(&syncState{}).Where("id = ? AND lease_owner = ?", "rawkuma", owner).Update("last_success_at", now).Error
}

func (r *Repository) importSeries(ctx context.Context, f Fetcher, existing Series, now time.Time) error {
	data, err := f.Series(ctx, existing.URL)
	if err != nil {
		return err
	}
	if data.Series.URL != existing.URL || len(data.Chapters) == 0 || len(data.Chapters) > 5000 || !safeGenres(data.Series.Genres) {
		return ErrBlocked
	}
	cover, err := r.cacheImage(ctx, f, data.Cover, existing.URL)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Preserve editorial flags if an editor hides this while it downloads.
		s := data.Series
		s.ID = existing.ID
		s.CoverURL = cover.URL
		s.Status = "ready"
		s.Enabled = existing.Enabled
		s.Reviewed = existing.Reviewed
		s.Language = "ja"
		s.ChapterCount = len(data.Chapters)
		s.FetchedAt = &now
		s.NextFetchAt = now.Add(RefreshInterval)
		if err := tx.Model(&Series{}).Where("id = ?", existing.ID).Select("title", "japanese_title", "summary", "author", "genres", "cover_url", "publication_status", "status", "chapter_count", "fetched_at", "last_error").Updates(&s).Error; err != nil {
			return err
		}
		for i := range data.Chapters {
			ch := data.Chapters[i]
			if ch.SeriesID != existing.ID || ch.ID != ID(ch.URL) || !strings.HasPrefix(ch.URL, existing.URL+"chapter-") || !allowedURL(ch.URL) {
				return ErrInvalid
			}
			ch.NextAttemptAt = time.Unix(1, 0).UTC()
			ch.Pages = []Page{}
			ch.Status = "available"
			ch.RequestedAt = nil
			if err := tx.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"title", "number", "position"})}).Create(&ch).Error; err != nil {
				return err
			}
		}
		// Cache the first available chapter for each initial title. The rest is
		// queued on demand; browsing the library never downloads whole series.
		first := data.Chapters[0]
		return tx.Model(&Chapter{}).Where("id = ? AND status = ?", first.ID, "available").Updates(map[string]any{"status": "queued", "requested_at": now}).Error
	})
}

func (r *Repository) cacheImage(ctx context.Context, f Fetcher, raw, referer string) (Asset, error) {
	if !allowedURL(raw) {
		return Asset{}, ErrInvalid
	}
	var cached Asset
	id := ID(raw)
	err := r.db.WithContext(ctx).First(&cached, "id = ?", id).Error
	if err == nil {
		return cached, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Asset{}, err
	}
	body, err := f.Image(ctx, raw, referer)
	if err != nil {
		return Asset{}, err
	}
	if _, _, err = inspectImage(body); err != nil {
		return Asset{}, err
	}
	cached, err = r.images.Put(ctx, body)
	if err != nil {
		return Asset{}, err
	}
	cached.ID = id
	cached.SourceURL = raw
	if err = r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&cached).Error; err != nil {
		return Asset{}, err
	}
	return cached, nil
}

func (r *Repository) importChapter(ctx context.Context, f Fetcher, ch Chapter, now time.Time) error {
	urls := ch.SourcePages
	if len(urls) == 0 {
		var err error
		urls, err = f.Pages(ctx, ch)
		if err != nil {
			return err
		}
		if len(urls) == 0 || len(urls) > 120 {
			return ErrInvalid
		}
		for _, raw := range urls {
			if !allowedURL(raw) {
				return ErrInvalid
			}
		}
		manifest, err := json.Marshal(urls)
		if err != nil {
			return err
		}
		if err = r.db.WithContext(ctx).Model(&Chapter{}).Where("id = ? AND status = ?", ch.ID, "processing").Update("source_pages", string(manifest)).Error; err != nil {
			return err
		}
	}
	if len(urls) == 0 || len(urls) > 120 {
		return ErrInvalid
	}
	pages := make([]Page, 0, len(urls))
	totalBytes := 0
	for i, raw := range urls {
		asset, err := r.cacheImage(ctx, f, raw, ch.URL)
		if err != nil {
			return err
		}
		totalBytes += asset.Bytes
		if totalBytes > 160<<20 {
			return errors.New("chapter exceeds 160 MiB import budget")
		}
		pages = append(pages, Page{Index: i + 1, URL: asset.URL, Width: asset.Width, Height: asset.Height})
	}
	// Publish only after every page is durable in COS. Partial chapters never
	// leak upstream URLs or appear as ready to a second reader.
	return r.db.WithContext(ctx).Model(&Chapter{}).Where("id = ? AND status = ?", ch.ID, "processing").Updates(map[string]any{"status": "ready", "pages": mustJSON(pages), "fetched_at": now, "last_error": "", "failure_count": 0}).Error
}

func mustJSON(pages []Page) string { body, _ := json.Marshal(pages); return string(body) }
