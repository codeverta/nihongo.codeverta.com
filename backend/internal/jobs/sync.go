package jobs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/jobfetch"
)

type Fetcher interface {
	Fetch(context.Context, string, jobfetch.Validators) (jobfetch.Result, error)
}
type SyncReport struct {
	Domains  int      `json:"domains"`
	Requests int      `json:"requests"`
	JobsSeen int      `json:"jobsSeen"`
	Errors   []string `json:"errors"`
}

// SyncSelected is used to replay captured results without contacting other hosts.
// A selection must include every configured feed on its selected domains, so a
// partial manual run cannot consume a shared weekly deadline for sibling feeds.
func (r *Repository) SyncSelected(ctx context.Context, fetcher Fetcher, ids []string) (SyncReport, error) {
	selected := map[string]bool{}
	domains := map[string]bool{}
	for _, id := range ids {
		selected[id] = true
	}
	for _, s := range r.sources {
		if selected[s.ID] {
			domains[s.Domain] = true
			delete(selected, s.ID)
		}
	}
	if len(selected) > 0 || len(domains) == 0 {
		return SyncReport{}, ErrInvalid
	}
	selected = map[string]bool{}
	for _, id := range ids {
		selected[id] = true
	}
	copyRepo := *r
	copyRepo.sources = nil
	for _, s := range r.sources {
		if domains[s.Domain] {
			if !selected[s.ID] {
				return SyncReport{}, ErrInvalid
			}
			copyRepo.sources = append(copyRepo.sources, s)
		}
	}
	return copyRepo.SyncDue(ctx, fetcher)
}

// Claim writes the next weekly deadline BEFORE any request. A process crash or
// second replica cannot turn retries/restarts into additional website requests.
func (r *Repository) claim(ctx context.Context, domain string) (string, string, bool, error) {
	owner := uuid.NewString()
	runID := uuid.NewString()
	claimed := false
	now := r.now().UTC()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var d domainRow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("domain = ?", domain).First(&d).Error; err != nil {
			return err
		}
		if d.NextFetchAt.After(now) || (d.LeaseUntil != nil && d.LeaseUntil.After(now)) {
			return nil
		}
		until := now.Add(15 * time.Minute)
		if err := tx.Model(&d).Updates(map[string]any{"next_fetch_at": now.Add(r.interval), "last_started_at": now, "lease_owner": owner, "lease_until": until}).Error; err != nil {
			return err
		}
		if err := tx.Create(&runRow{ID: runID, Domain: domain, StartedAt: now, Status: "running"}).Error; err != nil {
			return err
		}
		claimed = true
		return nil
	})
	return owner, runID, claimed, err
}
func (r *Repository) lease(tx *gorm.DB, domain, owner string) error {
	var d domainRow
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("domain = ?", domain).First(&d).Error; err != nil {
		return err
	}
	if d.LeaseOwner != owner || d.LeaseUntil == nil || !d.LeaseUntil.After(r.now().UTC()) {
		return ErrLeaseLost
	}
	return nil
}
func (r *Repository) SyncDue(ctx context.Context, fetcher Fetcher) (SyncReport, error) {
	report := SyncReport{Errors: []string{}}
	groups := map[string][]jobfetch.Source{}
	for _, s := range r.sources {
		groups[s.Domain] = append(groups[s.Domain], s)
	}
	domains := []string{}
	for domain := range groups {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	for _, domain := range domains {
		if ctx.Err() != nil {
			return report, ctx.Err()
		}
		owner, runID, claimed, err := r.claim(ctx, domain)
		if err != nil {
			return report, err
		}
		if !claimed {
			continue
		}
		report.Domains++
		requests, seen := 0, 0
		runStatus := "ok"
		runErrors := []string{}
		sources := groups[domain]
		sort.Slice(sources, func(i, j int) bool { return sources[i].ID < sources[j].ID })
		for index, s := range sources {
			if index >= MaxRequestsPerDomain {
				break
			}
			if index > 0 {
				timer := time.NewTimer(r.requestSpacing)
				select {
				case <-ctx.Done():
					timer.Stop()
					runStatus = "cancelled"
				case <-timer.C:
				}
				if ctx.Err() != nil {
					break
				}
			}
			var state sourceRow
			if err := r.db.WithContext(ctx).Where("code = ?", s.ID).First(&state).Error; err != nil {
				runErrors = append(runErrors, "load source state failed")
				runStatus = "error"
				break
			}
			// Reserve request count atomically under the lease before outbound access.
			now := r.now().UTC()
			budget := requestBudget(s)
			err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				if err := r.lease(tx, domain, owner); err != nil {
					return err
				}
				updated := tx.Model(&runRow{}).Where("id = ? AND requests <= ?", runID, MaxRequestsPerDomain-budget).Update("requests", gorm.Expr("requests + ?", budget))
				if updated.Error != nil {
					return updated.Error
				}
				if updated.RowsAffected != 1 {
					return errors.New("weekly request budget exhausted")
				}
				return tx.Model(&sourceRow{}).Where("code = ?", s.ID).Update("last_attempt_at", now).Error
			})
			if err != nil {
				runErrors = append(runErrors, "reserve request failed")
				runStatus = "error"
				break
			}
			fetchCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
			result, fetchErr := fetcher.Fetch(fetchCtx, s.ID, jobfetch.Validators{ETag: state.ETag, LastModified: state.LastModified})
			cancel()
			actualRequests := result.Requests
			if actualRequests <= 0 || actualRequests > budget {
				actualRequests = budget
			}
			requests += actualRequests
			report.Requests += actualRequests
			if fetchErr != nil {
				runStatus = "error"
				runErrors = append(runErrors, s.ID+": "+short(fetchErr.Error(), 200))
				report.Errors = append(report.Errors, s.ID+": collection failed")
				failureCtx, cancelFailure := context.WithTimeout(context.Background(), 5*time.Second)
				err = r.fail(failureCtx, s, domain, owner, result, fetchErr)
				cancelFailure()
				if err != nil {
					return report, err
				}
				// Forbidden/rate limited on a host stops the whole domain batch, no fallback.
				if result.Blocked || !result.RetryAfter.IsZero() {
					break
				}
				continue
			}
			observedAt := r.now().UTC()
			if cached, ok := fetcher.(interface{ ObservationTime(string) time.Time }); ok {
				observedAt = cached.ObservationTime(s.ID).UTC()
				if observedAt.IsZero() || observedAt.Before(r.now().Add(-24*time.Hour)) || observedAt.After(r.now().Add(5*time.Minute)) {
					return report, ErrInvalid
				}
			}
			n, err := r.ingest(ctx, s, domain, owner, result, observedAt)
			if err != nil {
				runStatus = "error"
				runErrors = append(runErrors, s.ID+": import failed")
				report.Errors = append(report.Errors, s.ID+": import failed")
				failureCtx, cancelFailure := context.WithTimeout(context.Background(), 5*time.Second)
				_ = r.fail(failureCtx, s, domain, owner, jobfetch.Result{}, err)
				cancelFailure()
				break
			}
			seen += n
			report.JobsSeen += n
			if !result.Complete && !result.NotModified {
				runStatus = "partial"
			}
		}
		finishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = r.db.WithContext(finishCtx).Transaction(func(tx *gorm.DB) error {
			var d domainRow
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("domain = ?", domain).First(&d).Error; err != nil {
				return err
			}
			if d.LeaseOwner != owner {
				return ErrLeaseLost
			}
			if err := tx.Model(&d).Updates(map[string]any{"lease_owner": "", "lease_until": nil}).Error; err != nil {
				return err
			}
			return tx.Model(&runRow{}).Where("id = ?", runID).Updates(map[string]any{"finished_at": r.now().UTC(), "requests": requests, "jobs_seen": seen, "status": runStatus, "error": short(strings.Join(runErrors, "; "), 512)}).Error
		})
		cancel()
		if err != nil {
			return report, err
		}
	}
	return report, nil
}
func (r *Repository) fail(ctx context.Context, s jobfetch.Source, domain, owner string, result jobfetch.Result, cause error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.lease(tx, domain, owner); err != nil {
			return err
		}
		status := "error"
		if result.Unsupported {
			status = "unsupported"
		}
		if result.Blocked {
			status = "blocked"
		}
		if err := tx.Model(&sourceRow{}).Where("code = ?", s.ID).Updates(map[string]any{"status": status, "last_error": short(cause.Error(), 512), "consecutive_errors": gorm.Expr("consecutive_errors + 1")}).Error; err != nil {
			return err
		}
		if !result.RetryAfter.IsZero() {
			var d domainRow
			if err := tx.Where("domain = ?", domain).First(&d).Error; err != nil {
				return err
			}
			if err := tx.Model(&d).Update("next_fetch_at", maxTime(d.NextFetchAt, result.RetryAfter.UTC())).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *Repository) ingest(ctx context.Context, s jobfetch.Source, domain, owner string, result jobfetch.Result, observedAt time.Time) (int, error) {
	if len(result.Jobs) > 2000 {
		return 0, ErrInvalid
	}
	at := observedAt.UTC()
	rows := []listingRow{}
	external := map[string]bool{}
	for _, job := range result.Jobs {
		if external[job.ExternalID] {
			return 0, fmt.Errorf("duplicate feed identity: %w", ErrInvalid)
		}
		external[job.ExternalID] = true
		row, err := normalJob(s, job, at)
		if err != nil {
			return 0, err
		}
		rows = append(rows, row)
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.lease(tx, domain, owner); err != nil {
			return err
		}
		affected := map[string]bool{}
		var previous []observationRow
		if err := tx.Where("source_id = ?", s.ID).Find(&previous).Error; err != nil {
			return err
		}
		for _, o := range previous {
			affected[o.ListingID] = true
		}
		if result.NotModified {
			if err := tx.Model(&observationRow{}).Where("source_id = ? AND active = ?", s.ID, true).Update("last_seen_at", at).Error; err != nil {
				return err
			}
		} else {
			for i, row := range rows {
				var existing listingRow
				err := tx.Where("id = ?", row.ID).First(&existing).Error
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if err == nil {
					row.FirstSeenAt = existing.FirstSeenAt
					if row.PostedAt == nil {
						row.SortAt = existing.SortAt
					}
				}
				if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&row).Error; err != nil {
					return err
				}
				observation := observationRow{ID: stableID("observation:" + s.ID + ":" + result.Jobs[i].ExternalID), SourceID: s.ID, ExternalID: result.Jobs[i].ExternalID, ListingID: row.ID, LastSeenAt: at, Active: row.Status == "active"}
				if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&observation).Error; err != nil {
					return err
				}
				affected[row.ID] = true
			}
			if result.Complete {
				if err := tx.Model(&observationRow{}).Where("source_id = ? AND last_seen_at < ?", s.ID, at).Update("active", false).Error; err != nil {
					return err
				}
			}
		}
		for id := range affected {
			var count int64
			if err := tx.Model(&observationRow{}).Where("listing_id = ? AND active = ? AND last_seen_at >= ?", id, true, at.Add(-StaleAfter)).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				if err := tx.Model(&listingRow{}).Where("id = ? AND status = ?", id, "active").Update("status", "closed").Error; err != nil {
					return err
				}
			} else {
				var latest observationRow
				if err := tx.Where("listing_id = ? AND active = ?", id, true).Order("last_seen_at DESC").First(&latest).Error; err != nil {
					return err
				}
				if err := tx.Model(&listingRow{}).Where("id = ?", id).Updates(map[string]any{"status": "active", "last_seen_at": latest.LastSeenAt}).Error; err != nil {
					return err
				}
			}
		}
		status := "ok"
		if !result.Complete && !result.NotModified {
			status = "partial"
		}
		changes := map[string]any{"last_success_at": at, "last_error": "", "status": status, "consecutive_errors": 0}
		if result.ETag != "" {
			changes["etag"] = short(result.ETag, 512)
		}
		if result.LastModified != "" {
			changes["last_modified"] = short(result.LastModified, 128)
		}
		return tx.Model(&sourceRow{}).Where("code = ?", s.ID).Updates(changes).Error
	})
	return len(rows), err
}

// The timer only checks database deadlines. Opening the app never starts a crawl.
func (r *Repository) Run(ctx context.Context, fetcher Fetcher) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if r.errorHandler == nil {
				panic(recovered)
			}
			r.reportError(fmt.Errorf("panic: %v\n%s", recovered, debug.Stack()))
		}
	}()
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		syncCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		report, err := r.SyncDue(syncCtx, fetcher)
		cancel()
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("jobs sync failed: %v", err)
			r.reportError(err)
		} else if len(report.Errors) > 0 {
			r.reportError(errors.New(strings.Join(report.Errors, "; ")))
		} else if report.Requests > 0 {
			log.Printf("jobs sync: domains=%d requests=%d jobs=%d errors=%d", report.Domains, report.Requests, report.JobsSeen, len(report.Errors))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
