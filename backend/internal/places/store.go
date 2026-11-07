package places

import "context"

// Store is the persistence the service needs. Repository (MariaDB) and
// MemoryStore (tests) implement it.
type Store interface {
	Published(ctx context.Context) ([]Spot, error)
	All(ctx context.Context) ([]Spot, error)
	Get(ctx context.Context, id string) (Spot, error)
	ByWikidata(ctx context.Context, wikidataID string) (Spot, error)
	Save(ctx context.Context, spot *Spot) error
	Delete(ctx context.Context, id string) error
}
