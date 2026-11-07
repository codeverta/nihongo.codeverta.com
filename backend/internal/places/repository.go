package places

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"nihongo-master-backend/internal/database"
)

type Repository struct{ db *gorm.DB }

func New(db *gorm.DB) (*Repository, error) {
	if db == nil {
		return nil, fmt.Errorf("map spots require a database")
	}
	if err := database.Migrate(db, &Spot{}); err != nil {
		return nil, fmt.Errorf("initialize map spot schema: %w", err)
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Published(ctx context.Context) ([]Spot, error) {
	var spots []Spot
	err := r.db.WithContext(ctx).Where("published = ?", true).Order("prefecture").Order("name_en").Find(&spots).Error
	return spots, err
}

func (r *Repository) All(ctx context.Context) ([]Spot, error) {
	var spots []Spot
	err := r.db.WithContext(ctx).Order("prefecture").Order("name_en").Find(&spots).Error
	return spots, err
}

func (r *Repository) Get(ctx context.Context, id string) (Spot, error) {
	return r.take(ctx, "id = ?", id)
}

func (r *Repository) ByWikidata(ctx context.Context, wikidataID string) (Spot, error) {
	return r.take(ctx, "wikidata_id = ?", wikidataID)
}

func (r *Repository) take(ctx context.Context, where string, value string) (Spot, error) {
	var spot Spot
	err := r.db.WithContext(ctx).Where(where, value).Take(&spot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Spot{}, ErrNotFound
	}
	return spot, err
}

func (r *Repository) Save(ctx context.Context, spot *Spot) error {
	if spot.ID == "" {
		spot.ID = uuid.NewString()
		return r.db.WithContext(ctx).Create(spot).Error
	}
	return r.db.WithContext(ctx).Save(spot).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&Spot{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
