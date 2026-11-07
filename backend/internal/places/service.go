package places

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
	"unicode"
)

// ImageStore keeps optimized JPEG copies of spot photos (COS or local disk).
type ImageStore interface {
	Put(ctx context.Context, jpeg []byte) (string, error)
}

type Service struct {
	Store   Store
	Fetcher *Fetcher
	Images  ImageStore
	// Reading converts Japanese text to hiragana ("" when unknown).
	Reading func(string) string
	Now     func() time.Time

	mu sync.Mutex // imports and refreshes run one at a time
}

// Import adds a spot from a Wikidata ID, en.wikipedia URL, or title. It is
// saved unpublished so the admin can review it first.
func (s *Service) Import(ctx context.Context, input, category string) (Spot, error) {
	if !ValidCategory(category) {
		return Spot{}, ErrInvalid
	}
	qid, title, err := ParseInput(input)
	if err != nil {
		return Spot{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if qid != "" {
		if existing, err := s.Store.ByWikidata(ctx, qid); err == nil {
			return existing, ErrExists
		} else if !errors.Is(err, ErrNotFound) {
			return Spot{}, err
		}
	}
	fetched, err := s.Fetcher.Fetch(ctx, qid, title)
	if err != nil {
		return Spot{}, err
	}
	if existing, err := s.Store.ByWikidata(ctx, fetched.WikidataID); err == nil {
		return existing, ErrExists
	} else if !errors.Is(err, ErrNotFound) {
		return Spot{}, err
	}
	spot := Spot{WikidataID: fetched.WikidataID, Category: category}
	if err := s.apply(ctx, &spot, fetched); err != nil {
		return Spot{}, err
	}
	if err := s.Store.Save(ctx, &spot); err != nil {
		return Spot{}, err
	}
	return spot, nil
}

// Refresh re-fetches sources. Manual text, phrases, category, published state,
// admin-set location, and uploaded photos are kept.
func (s *Service) Refresh(ctx context.Context, id string) (Spot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	spot, err := s.Store.Get(ctx, id)
	if err != nil {
		return Spot{}, err
	}
	fetched, err := s.Fetcher.Fetch(ctx, spot.WikidataID, "")
	if err != nil {
		return Spot{}, err
	}
	if err := s.apply(ctx, &spot, fetched); err != nil {
		return Spot{}, err
	}
	if spot.Published && !spot.Publishable() {
		spot.Published = false
	}
	if err := s.Store.Save(ctx, &spot); err != nil {
		return Spot{}, err
	}
	return spot, nil
}

func (s *Service) apply(ctx context.Context, spot *Spot, f Fetched) error {
	spot.NameEN = f.Labels["en"]
	spot.NameID = f.Labels["id"]
	spot.NameJA = f.Labels["ja"]
	spot.Reading = ""
	if s.Reading != nil && hasJapanese(spot.NameJA) {
		spot.Reading = s.Reading(spot.NameJA)
	}
	spot.SummaryEN, spot.WikipediaEN = f.Wikipedia["en"].Text, f.Wikipedia["en"].URL
	spot.SummaryID, spot.WikipediaID = f.Wikipedia["id"].Text, f.Wikipedia["id"].URL
	spot.WikipediaJA = f.Wikipedia["ja"].URL
	spot.TipsEN, spot.WikivoyageEN = f.Wikivoyage["en"].Text, f.Wikivoyage["en"].URL
	spot.TipsID, spot.WikivoyageID = f.Wikivoyage["id"].Text, f.Wikivoyage["id"].URL
	if !spot.ManualLocation {
		spot.Lat, spot.Lng, spot.Prefecture = f.Lat, f.Lng, ""
		if f.Lat != nil && f.Lng != nil {
			spot.Prefecture = PrefectureAt(*f.Lat, *f.Lng)
		}
	}
	if !spot.ManualImage {
		spot.ImageURL, spot.ImageAuthor, spot.ImageLicense, spot.ImageLicenseURL, spot.ImageSourceURL = "", "", "", "", ""
		if f.Image != nil && s.Images != nil {
			stored, err := s.Images.Put(ctx, f.Image.JPEG)
			if err != nil {
				return err
			}
			spot.ImageURL = stored
			spot.ImageAuthor, spot.ImageLicense = f.Image.Author, f.Image.License
			spot.ImageLicenseURL, spot.ImageSourceURL = f.Image.LicenseURL, f.Image.SourceURL
		}
	}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	spot.FetchedAt = now().UnixMilli()
	return nil
}

func hasJapanese(text string) bool {
	for _, r := range text {
		if unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) {
			return true
		}
	}
	return false
}

// Edit is what the admin form can change.
type Edit struct {
	Category        string   `json:"category"`
	Prefecture      string   `json:"prefecture"`
	Lat             *float64 `json:"lat"`
	Lng             *float64 `json:"lng"`
	ManualLocation  bool     `json:"manualLocation"`
	NameEN          string   `json:"nameEn"`
	NameID          string   `json:"nameId"`
	NameJA          string   `json:"nameJa"`
	Reading         string   `json:"reading"`
	ManualSummaryEN string   `json:"manualSummaryEn"`
	ManualSummaryID string   `json:"manualSummaryId"`
	ManualTipsEN    string   `json:"manualTipsEn"`
	ManualTipsID    string   `json:"manualTipsId"`
	Phrases         []Phrase `json:"phrases"`
	Published       bool     `json:"published"`
}

func (s *Service) Update(ctx context.Context, id string, e Edit) (Spot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	spot, err := s.Store.Get(ctx, id)
	if err != nil {
		return Spot{}, err
	}
	if !ValidCategory(e.Category) {
		return Spot{}, ErrInvalid
	}
	for _, field := range []struct {
		value string
		limit int
	}{{e.NameEN, 200}, {e.NameID, 200}, {e.NameJA, 200}, {e.Reading, 200}, {e.ManualSummaryEN, 5000}, {e.ManualSummaryID, 5000}, {e.ManualTipsEN, 5000}, {e.ManualTipsID, 5000}} {
		if !within(field.value, field.limit) {
			return Spot{}, ErrInvalid
		}
	}
	spot.Category = e.Category
	if e.ManualLocation {
		if e.Lat == nil || e.Lng == nil || *e.Lat < 20 || *e.Lat > 46 || *e.Lng < 122 || *e.Lng > 154 {
			return Spot{}, ErrInvalid
		}
		prefecture := strings.TrimSpace(e.Prefecture)
		if prefecture == "" {
			prefecture = PrefectureAt(*e.Lat, *e.Lng)
		}
		if !ValidPrefecture(prefecture) {
			return Spot{}, ErrInvalid
		}
		lat, lng := *e.Lat, *e.Lng
		spot.Lat, spot.Lng, spot.Prefecture, spot.ManualLocation = &lat, &lng, prefecture, true
	} else {
		spot.ManualLocation = false
	}
	spot.NameEN, spot.NameID, spot.NameJA = strings.TrimSpace(e.NameEN), strings.TrimSpace(e.NameID), strings.TrimSpace(e.NameJA)
	spot.Reading = strings.TrimSpace(e.Reading)
	spot.ManualSummaryEN, spot.ManualSummaryID = strings.TrimSpace(e.ManualSummaryEN), strings.TrimSpace(e.ManualSummaryID)
	spot.ManualTipsEN, spot.ManualTipsID = strings.TrimSpace(e.ManualTipsEN), strings.TrimSpace(e.ManualTipsID)
	if err := spot.SetPhrases(e.Phrases); err != nil {
		return Spot{}, err
	}
	if e.Published && !spot.Publishable() {
		return Spot{}, ErrInvalid
	}
	spot.Published = e.Published
	if err := s.Store.Save(ctx, &spot); err != nil {
		return Spot{}, err
	}
	return spot, nil
}

// SetImage stores an admin-uploaded photo; refreshes keep it.
func (s *Service) SetImage(ctx context.Context, id string, jpeg []byte, author, license string) (Spot, error) {
	author, license = strings.TrimSpace(author), strings.TrimSpace(license)
	if s.Images == nil || author == "" || license == "" || !within(author, 300) || !within(license, 120) {
		return Spot{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	spot, err := s.Store.Get(ctx, id)
	if err != nil {
		return Spot{}, err
	}
	stored, err := s.Images.Put(ctx, jpeg)
	if err != nil {
		return Spot{}, err
	}
	spot.ImageURL, spot.ImageAuthor, spot.ImageLicense = stored, author, license
	spot.ImageLicenseURL, spot.ImageSourceURL, spot.ManualImage = "", "", true
	if err := s.Store.Save(ctx, &spot); err != nil {
		return Spot{}, err
	}
	return spot, nil
}

// ClearImage drops an uploaded photo so the next refresh uses Commons again.
func (s *Service) ClearImage(ctx context.Context, id string) (Spot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	spot, err := s.Store.Get(ctx, id)
	if err != nil {
		return Spot{}, err
	}
	spot.ImageURL, spot.ImageAuthor, spot.ImageLicense, spot.ImageLicenseURL, spot.ImageSourceURL = "", "", "", "", ""
	spot.ManualImage = false
	if err := s.Store.Save(ctx, &spot); err != nil {
		return Spot{}, err
	}
	return spot, nil
}
