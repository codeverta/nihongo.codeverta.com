// Package places stores curated Japan map spots. Content is fetched once from
// Wikidata, Wikipedia, Wikivoyage, and Wikimedia Commons when an admin adds or
// refreshes a spot, then served from the database only.
package places

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalid  = errors.New("invalid map spot")
	ErrNotFound = errors.New("map spot not found")
	ErrExists   = errors.New("map spot already exists")
	ErrSource   = errors.New("map source unavailable")
)

// Categories in display order. Labels live in the app's localization.
var Categories = []string{"culture", "shrine", "nature", "food", "city", "history"}

func ValidCategory(category string) bool {
	for _, c := range Categories {
		if c == category {
			return true
		}
	}
	return false
}

// Spot is the stored record. Fetched fields are overwritten by a refresh;
// Manual* fields, category, phrases, and the published flag never are.
type Spot struct {
	ID         string   `gorm:"type:char(36);primaryKey" json:"id"`
	WikidataID string   `gorm:"type:varchar(32);uniqueIndex:map_spot_wikidata" json:"wikidataId"`
	Category   string   `gorm:"type:varchar(32);index" json:"category"`
	Prefecture string   `gorm:"type:varchar(2);index" json:"prefecture"`
	Lat        *float64 `json:"lat"`
	Lng        *float64 `json:"lng"`
	// ManualLocation keeps admin-set coordinates and prefecture on refresh.
	ManualLocation bool `json:"manualLocation"`

	NameEN  string `gorm:"column:name_en;type:varchar(255)" json:"nameEn"`
	NameID  string `gorm:"column:name_id;type:varchar(255)" json:"nameId"`
	NameJA  string `gorm:"column:name_ja;type:varchar(255)" json:"nameJa"`
	Reading string `gorm:"type:varchar(255)" json:"reading"`

	SummaryEN    string `gorm:"column:summary_en;type:text" json:"summaryEn"`
	SummaryID    string `gorm:"column:summary_id;type:text" json:"summaryId"`
	WikipediaEN  string `gorm:"column:wikipedia_en;type:varchar(512)" json:"wikipediaEn"`
	WikipediaID  string `gorm:"column:wikipedia_id;type:varchar(512)" json:"wikipediaId"`
	WikipediaJA  string `gorm:"column:wikipedia_ja;type:varchar(512)" json:"wikipediaJa"`
	TipsEN       string `gorm:"column:tips_en;type:text" json:"tipsEn"`
	TipsID       string `gorm:"column:tips_id;type:text" json:"tipsId"`
	WikivoyageEN string `gorm:"column:wikivoyage_en;type:varchar(512)" json:"wikivoyageEn"`
	WikivoyageID string `gorm:"column:wikivoyage_id;type:varchar(512)" json:"wikivoyageId"`

	ImageURL        string `gorm:"type:varchar(1024)" json:"imageUrl"`
	ImageAuthor     string `gorm:"type:varchar(512)" json:"imageAuthor"`
	ImageLicense    string `gorm:"type:varchar(128)" json:"imageLicense"`
	ImageLicenseURL string `gorm:"type:varchar(512)" json:"imageLicenseUrl"`
	ImageSourceURL  string `gorm:"type:varchar(1024)" json:"imageSourceUrl"`
	// ManualImage keeps an admin-uploaded photo on refresh.
	ManualImage bool `json:"manualImage"`

	ManualSummaryEN string `gorm:"column:manual_summary_en;type:text" json:"manualSummaryEn"`
	ManualSummaryID string `gorm:"column:manual_summary_id;type:text" json:"manualSummaryId"`
	ManualTipsEN    string `gorm:"column:manual_tips_en;type:text" json:"manualTipsEn"`
	ManualTipsID    string `gorm:"column:manual_tips_id;type:text" json:"manualTipsId"`
	PhrasesJSON     []byte `gorm:"column:phrases;type:blob" json:"-"`

	Published bool  `gorm:"index" json:"published"`
	FetchedAt int64 `json:"fetchedAt"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli" json:"updatedAt"`
}

func (Spot) TableName() string { return "map_spots" }

type Phrase struct {
	Japanese  string `json:"japanese"`
	Reading   string `json:"reading"`
	MeaningID string `json:"meaningId"`
	MeaningEN string `json:"meaningEn"`
}

func (s *Spot) Phrases() []Phrase {
	var phrases []Phrase
	if len(s.PhrasesJSON) > 0 && json.Unmarshal(s.PhrasesJSON, &phrases) == nil {
		return phrases
	}
	return []Phrase{}
}

func (s *Spot) SetPhrases(phrases []Phrase) error {
	if len(phrases) > 30 {
		return ErrInvalid
	}
	clean := make([]Phrase, 0, len(phrases))
	for _, p := range phrases {
		p = Phrase{strings.TrimSpace(p.Japanese), strings.TrimSpace(p.Reading), strings.TrimSpace(p.MeaningID), strings.TrimSpace(p.MeaningEN)}
		if p.Japanese == "" || p.MeaningID == "" || !within(p.Japanese, 120) || !within(p.Reading, 200) || !within(p.MeaningID, 300) || !within(p.MeaningEN, 300) {
			return ErrInvalid
		}
		clean = append(clean, p)
	}
	data, err := json.Marshal(clean)
	if err != nil {
		return err
	}
	s.PhrasesJSON = data
	return nil
}

func within(text string, limit int) bool {
	return utf8.ValidString(text) && utf8.RuneCountInString(text) <= limit
}

// Publishable reports whether the spot can appear on the map.
func (s *Spot) Publishable() bool {
	return s.Lat != nil && s.Lng != nil && ValidPrefecture(s.Prefecture) && ValidCategory(s.Category) && (s.NameEN != "" || s.NameID != "")
}

// MissingIndonesian is shown in the admin list: neither an Indonesian
// Wikipedia summary nor a manual Indonesian summary exists.
func (s *Spot) MissingIndonesian() bool {
	return strings.TrimSpace(s.ManualSummaryID) == "" && strings.TrimSpace(s.SummaryID) == ""
}

// LocalizedText is one resolved text for a language. SourceLang differs from
// the requested language when the English fallback is used.
type LocalizedText struct {
	Text       string `json:"text"`
	SourceLang string `json:"sourceLang"`
	Manual     bool   `json:"manual"`
}

func resolve(manual, local, english, lang string) LocalizedText {
	switch {
	case strings.TrimSpace(manual) != "":
		return LocalizedText{manual, lang, true}
	case strings.TrimSpace(local) != "":
		return LocalizedText{local, lang, false}
	case lang != "en" && strings.TrimSpace(english) != "":
		return LocalizedText{english, "en", false}
	}
	return LocalizedText{}
}

type Source struct {
	Kind    string `json:"kind"` // wikipedia, wikivoyage, wikidata
	Lang    string `json:"lang"`
	URL     string `json:"url"`
	License string `json:"license"`
}

type PublicImage struct {
	URL        string `json:"url"`
	Author     string `json:"author"`
	License    string `json:"license"`
	LicenseURL string `json:"licenseUrl"`
	SourceURL  string `json:"sourceUrl"`
}

// PublicSpot is what the app receives. Text is pre-resolved per language so
// the app can switch languages offline.
type PublicSpot struct {
	ID         string                   `json:"id"`
	Category   string                   `json:"category"`
	Prefecture string                   `json:"prefecture"`
	Lat        float64                  `json:"lat"`
	Lng        float64                  `json:"lng"`
	Name       map[string]string        `json:"name"`
	NameJa     string                   `json:"nameJa"`
	Reading    string                   `json:"reading"`
	Summary    map[string]LocalizedText `json:"summary"`
	Tips       map[string]LocalizedText `json:"tips"`
	Phrases    []Phrase                 `json:"phrases"`
	Image      *PublicImage             `json:"image"`
	Sources    []Source                 `json:"sources"`
	UpdatedAt  int64                    `json:"updatedAt"`
}

const textLicense = "CC BY-SA 4.0"

func (s *Spot) Public() PublicSpot {
	nameID := s.NameID
	if nameID == "" {
		nameID = s.NameEN
	}
	nameEN := s.NameEN
	if nameEN == "" {
		nameEN = s.NameID
	}
	out := PublicSpot{
		ID: s.ID, Category: s.Category, Prefecture: s.Prefecture,
		Name:    map[string]string{"id": nameID, "en": nameEN},
		NameJa:  s.NameJA,
		Reading: s.Reading,
		Summary: map[string]LocalizedText{
			"id": resolve(s.ManualSummaryID, s.SummaryID, s.SummaryEN, "id"),
			"en": resolve(s.ManualSummaryEN, s.SummaryEN, "", "en"),
		},
		Tips: map[string]LocalizedText{
			"id": resolve(s.ManualTipsID, s.TipsID, s.TipsEN, "id"),
			"en": resolve(s.ManualTipsEN, s.TipsEN, "", "en"),
		},
		Phrases:   s.Phrases(),
		Sources:   []Source{},
		UpdatedAt: s.UpdatedAt,
	}
	if s.Lat != nil && s.Lng != nil {
		out.Lat, out.Lng = *s.Lat, *s.Lng
	}
	if s.ImageURL != "" {
		out.Image = &PublicImage{s.ImageURL, s.ImageAuthor, s.ImageLicense, s.ImageLicenseURL, s.ImageSourceURL}
	}
	for _, src := range []Source{
		{"wikipedia", "id", s.WikipediaID, textLicense},
		{"wikipedia", "en", s.WikipediaEN, textLicense},
		{"wikivoyage", "id", s.WikivoyageID, textLicense},
		{"wikivoyage", "en", s.WikivoyageEN, textLicense},
	} {
		if src.URL != "" {
			out.Sources = append(out.Sources, src)
		}
	}
	if s.WikidataID != "" {
		out.Sources = append(out.Sources, Source{"wikidata", "", "https://www.wikidata.org/wiki/" + s.WikidataID, "CC0"})
	}
	return out
}
