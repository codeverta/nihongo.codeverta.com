package prefinfo

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
)

// Prefecture profiles: Wikipedia-style text (intro and a few sections from
// id/en Wikipedia, CC BY-SA 4.0) and facts from Wikidata (CC0) for each of
// the 47 prefectures. The data is fetched offline by a script and bundled.

//go:embed profiles.json
var profilesJSON []byte

// Profiles is the bundled document, served to the app as is.
type Profiles struct {
	Raw  json.RawMessage
	ETag string
}

// LoadProfiles returns the bundled profiles, parsed and checked once.
func LoadProfiles() (Profiles, error) { return loadProfiles() }

var loadProfiles = sync.OnceValues(parseProfiles)

func parseProfiles() (Profiles, error) {
	var doc struct {
		Prefectures []struct {
			Code string `json:"code"`
			Text map[string]struct {
				Intro string `json:"intro"`
			} `json:"text"`
		} `json:"prefectures"`
	}
	if err := json.Unmarshal(profilesJSON, &doc); err != nil {
		return Profiles{}, fmt.Errorf("parse bundled prefecture profiles: %w", err)
	}
	if len(doc.Prefectures) != 47 {
		return Profiles{}, fmt.Errorf("prefecture profiles: want 47 prefectures, got %d", len(doc.Prefectures))
	}
	seen := map[string]bool{}
	for _, p := range doc.Prefectures {
		if !codePattern.MatchString(p.Code) || seen[p.Code] {
			return Profiles{}, fmt.Errorf("prefecture profiles: bad or duplicate code %q", p.Code)
		}
		seen[p.Code] = true
		if p.Text["id"].Intro == "" && p.Text["en"].Intro == "" {
			return Profiles{}, fmt.Errorf("prefecture profiles: %s has no text", p.Code)
		}
	}
	sum := sha256.Sum256(profilesJSON)
	return Profiles{Raw: profilesJSON, ETag: `"` + hex.EncodeToString(sum[:8]) + `"`}, nil
}
