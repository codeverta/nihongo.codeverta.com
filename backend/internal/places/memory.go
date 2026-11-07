package places

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
)

// MemoryStore is an in-memory Store for tests and local development.
type MemoryStore struct {
	mu    sync.Mutex
	spots map[string]Spot
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{spots: map[string]Spot{}} }

func (m *MemoryStore) list(keep func(Spot) bool) []Spot {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []Spot{}
	for _, spot := range m.spots {
		if keep(spot) {
			out = append(out, spot)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Prefecture != out[j].Prefecture {
			return out[i].Prefecture < out[j].Prefecture
		}
		return out[i].NameEN < out[j].NameEN
	})
	return out
}

func (m *MemoryStore) Published(context.Context) ([]Spot, error) {
	return m.list(func(s Spot) bool { return s.Published }), nil
}

func (m *MemoryStore) All(context.Context) ([]Spot, error) {
	return m.list(func(Spot) bool { return true }), nil
}

func (m *MemoryStore) Get(_ context.Context, id string) (Spot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	spot, ok := m.spots[id]
	if !ok {
		return Spot{}, ErrNotFound
	}
	return spot, nil
}

func (m *MemoryStore) ByWikidata(_ context.Context, wikidataID string) (Spot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, spot := range m.spots {
		if spot.WikidataID == wikidataID {
			return spot, nil
		}
	}
	return Spot{}, ErrNotFound
}

func (m *MemoryStore) Save(_ context.Context, spot *Spot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if spot.ID == "" {
		spot.ID = uuid.NewString()
	}
	m.spots[spot.ID] = *spot
	return nil
}

func (m *MemoryStore) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.spots[id]; !ok {
		return ErrNotFound
	}
	delete(m.spots, id)
	return nil
}
