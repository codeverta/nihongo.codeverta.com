// Package anki serves the pinned, local Kotoba mirror. It never fetches upstream.
package anki

import (
	"bytes"
	"compress/gzip"
	"container/list"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

//go:embed resources/catalog.json resources/cards resources/source resources/manifest.json resources/KOTOBA-LICENSE.txt
var files embed.FS

var ErrNotFound = errors.New("deck or card not found")

type Card struct {
	Index       int      `json:"index"`
	SourceIndex int      `json:"sourceIndex"`
	Front       string   `json:"front"`
	Answers     []string `json:"answers"`
	Meaning     string   `json:"meaning"`
	ImageURL    string   `json:"imageUrl"`
}
type Deck struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	Category            string `json:"category"`
	Level               string `json:"level"`
	Description         string `json:"description"`
	Instructions        string `json:"instructions"`
	OriginalName        string `json:"originalName"`
	OriginalDescription string `json:"originalDescription"`
	SourceFile          string `json:"sourceFile"`
	SourceURL           string `json:"sourceUrl"`
	MeaningLanguage     string `json:"meaningLanguage"`
	CardCount           int    `json:"cardCount"`
	SourceCount         int    `json:"sourceCount"`
	SkippedCount        int    `json:"skippedCount"`
	ChunkSize           int    `json:"chunkSize"`
	Available           bool   `json:"available"`
	UnavailableReason   string `json:"unavailableReason"`
	Preview             []Card `json:"preview,omitempty"`
}
type Catalog struct {
	Revision string `json:"revision"`
	Decks    []Deck `json:"decks"`
}

var catalog Catalog
var catalogOnce sync.Once
var catalogError error

func load() error {
	catalogOnce.Do(func() {
		raw, err := files.ReadFile("resources/catalog.json")
		if err != nil {
			catalogError = err
			return
		}
		catalogError = json.Unmarshal(raw, &catalog)
	})
	return catalogError
}
func List(query, category string) ([]Deck, error) {
	if err := load(); err != nil {
		return nil, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	result := []Deck{}
	for _, deck := range catalog.Decks {
		if category != "" && deck.Category != category {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(deck.ID+" "+deck.Title+" "+deck.OriginalName+" "+deck.Description+" "+deck.Level), query) {
			continue
		}
		deck.Preview = nil
		deck.OriginalDescription = ""
		result = append(result, deck)
	}
	return result, nil
}
func Get(id string) (Deck, error) {
	if err := load(); err != nil {
		return Deck{}, err
	}
	for _, deck := range catalog.Decks {
		if deck.ID == id {
			return deck, nil
		}
	}
	return Deck{}, ErrNotFound
}
func Revision() string { _ = load(); return catalog.Revision }
func License() string  { raw, _ := files.ReadFile("resources/KOTOBA-LICENSE.txt"); return string(raw) }

type cacheEntry struct {
	key   string
	cards []Card
}

var cache = struct {
	sync.Mutex
	order *list.List
	items map[string]*list.Element
}{order: list.New(), items: map[string]*list.Element{}}

// ReadCard uses a bounded chunk cache, including for the largest decks.
func ReadCard(id string, index int) (Card, error) {
	deck, err := Get(id)
	if err != nil {
		return Card{}, err
	}
	if index < 0 || index >= deck.CardCount || deck.ChunkSize < 1 {
		return Card{}, ErrNotFound
	}
	key := fmt.Sprintf("resources/cards/%s/%06d.json.gz", id, index/deck.ChunkSize)
	cache.Lock()
	if entry := cache.items[key]; entry != nil {
		cache.order.MoveToFront(entry)
		card := entry.Value.(cacheEntry).cards[index%deck.ChunkSize]
		cache.Unlock()
		return card, nil
	}
	cache.Unlock()
	raw, err := files.ReadFile(key)
	if err != nil {
		return Card{}, err
	}
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return Card{}, err
	}
	defer reader.Close()
	var cards []Card
	if err = json.NewDecoder(io.LimitReader(reader, 4<<20)).Decode(&cards); err != nil {
		return Card{}, err
	}
	if index%deck.ChunkSize >= len(cards) {
		return Card{}, ErrNotFound
	}
	cache.Lock()
	if cache.items[key] == nil {
		cache.items[key] = cache.order.PushFront(cacheEntry{key, cards})
		if cache.order.Len() > 16 {
			old := cache.order.Back()
			delete(cache.items, old.Value.(cacheEntry).key)
			cache.order.Remove(old)
		}
	}
	cache.Unlock()
	return cards[index%deck.ChunkSize], nil
}
