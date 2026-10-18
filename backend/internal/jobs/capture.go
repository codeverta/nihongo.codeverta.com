package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/jobfetch"
)

// ImportCaptured reprocesses a recent real response without any HTTP access.
// It keeps the domain's weekly deadline and uses the original observation time.
func (r *Repository) ImportCaptured(ctx context.Context, id string, result jobfetch.Result, observedAt time.Time) (int, error) {
	if observedAt.IsZero() || observedAt.Before(r.now().Add(-24*time.Hour)) || observedAt.After(r.now().Add(5*time.Minute)) {
		return 0, ErrInvalid
	}
	var source jobfetch.Source
	for _, s := range r.sources {
		if s.ID == id {
			source = s
			break
		}
	}
	if source.ID == "" {
		return 0, ErrInvalid
	}
	owner := uuid.NewString()
	started := r.now().UTC()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var domain domainRow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("domain = ?", source.Domain).First(&domain).Error; err != nil {
			return err
		}
		if domain.LeaseUntil != nil && domain.LeaseUntil.After(started) {
			return ErrLeaseLost
		}
		until := started.Add(15 * time.Minute)
		updates := map[string]any{"lease_owner": owner, "lease_until": until, "next_fetch_at": maxTime(domain.NextFetchAt, observedAt.Add(r.interval))}
		if domain.LastStartedAt == nil {
			updates["last_started_at"] = observedAt
		}
		return tx.Model(&domain).Updates(updates).Error
	})
	if err != nil {
		return 0, err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r.db.WithContext(cleanup).Model(&domainRow{}).Where("domain = ? AND lease_owner = ?", source.Domain, owner).Updates(map[string]any{"lease_owner": "", "lease_until": nil})
	}()
	var state sourceRow
	if err := r.db.WithContext(ctx).Where("code = ?", id).First(&state).Error; err != nil {
		return 0, err
	}
	var n int
	// An older capture may repair previously unknown normalized metadata, but
	// must never roll back current descriptions, observations, or availability.
	if state.LastSuccessAt != nil && state.LastSuccessAt.After(observedAt) {
		n, err = r.repairCapturedMetadata(ctx, source, owner, result, observedAt)
	} else {
		n, err = r.ingest(ctx, source, source.Domain, owner, result, observedAt)
	}
	status := "replayed"
	if err != nil {
		status = "error"
	}
	auditErr := r.db.WithContext(ctx).Create(&runRow{ID: uuid.NewString(), Domain: source.Domain, StartedAt: started, FinishedAt: timePointer(r.now().UTC()), JobsSeen: n, Status: status}).Error
	if err != nil {
		return n, err
	}
	return n, auditErr
}

func (r *Repository) repairCapturedMetadata(ctx context.Context, source jobfetch.Source, owner string, result jobfetch.Result, observedAt time.Time) (int, error) {
	n := 0
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.lease(tx, source.Domain, owner); err != nil {
			return err
		}
		for _, job := range result.Jobs {
			normal, err := normalJob(source, job, observedAt)
			if err != nil {
				return err
			}
			var observation observationRow
			if err := tx.Where("source_id = ? AND external_id = ?", source.ID, job.ExternalID).First(&observation).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					continue
				}
				return err
			}
			var row listingRow
			if err := tx.Where("id = ?", observation.ListingID).First(&row).Error; err != nil {
				return err
			}
			if row.CanonicalKey != normal.CanonicalKey || row.Title != normal.Title {
				continue
			}
			updates := map[string]any{}
			if row.Category == "general" && normal.Category != "general" {
				updates["category"] = normal.Category
			}
			if row.Prefecture == "" && normal.Prefecture != "" {
				updates["prefecture"] = normal.Prefecture
			}
			if len(updates) > 0 {
				if err := tx.Model(&row).Updates(updates).Error; err != nil {
					return err
				}
				n++
			}
		}
		return nil
	})
	return n, err
}

func timePointer(t time.Time) *time.Time { return &t }
