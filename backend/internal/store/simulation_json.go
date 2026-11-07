package store

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Practice papers 06 and later are stored as JSON files instead of Go literals so
// that content can be edited, reviewed and diffed without touching code.
// File name: simulation_data/jlpt_<level>_paper<NN>.json (level n5 or n4).
// Choices are listed with the correct answer first; the server shuffles them per attempt.
//
//go:embed simulation_data/*.json
var simulationDataFS embed.FS

// simulationPaperFile is the on-disk JSON shape of one practice paper.
type simulationPaperFile struct {
	Level string                `json:"level"`
	Paper int                   `json:"paper"`
	Items []simulationItemEntry `json:"items"`
}

type simulationItemEntry struct {
	Section       string   `json:"section"`
	Group         string   `json:"group"`
	Instructions  string   `json:"instructions"`
	Passage       string   `json:"passage"`
	Prompt        string   `json:"prompt"`
	Choices       []string `json:"choices"`
	Explanation   string   `json:"explanation"`
	ExplanationEn string   `json:"explanationEn"`
}

type simulationJSONCatalog struct {
	items map[string][]simulationItem // key "N5-6"
	en    map[string]string           // Indonesian explanation -> English explanation
}

var (
	simulationJSONOnce  sync.Once
	simulationJSONData  simulationJSONCatalog
	simulationJSONError error
)

func loadSimulationJSON() (simulationJSONCatalog, error) {
	simulationJSONOnce.Do(func() {
		cat := simulationJSONCatalog{items: map[string][]simulationItem{}, en: map[string]string{}}
		entries, err := simulationDataFS.ReadDir("simulation_data")
		if err != nil {
			simulationJSONError = err
			return
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			raw, err := simulationDataFS.ReadFile("simulation_data/" + entry.Name())
			if err != nil {
				simulationJSONError = err
				return
			}
			var file simulationPaperFile
			if err := json.Unmarshal(raw, &file); err != nil {
				simulationJSONError = fmt.Errorf("%s: %w", entry.Name(), err)
				return
			}
			key := fmt.Sprintf("%s-%d", strings.ToUpper(file.Level), file.Paper)
			if _, dup := cat.items[key]; dup {
				simulationJSONError = fmt.Errorf("%s: duplicate paper %s", entry.Name(), key)
				return
			}
			items := make([]simulationItem, 0, len(file.Items))
			for i, it := range file.Items {
				if len(it.Choices) != 4 || strings.TrimSpace(it.Explanation) == "" || strings.TrimSpace(it.ExplanationEn) == "" {
					simulationJSONError = fmt.Errorf("%s: item %d is incomplete", entry.Name(), i+1)
					return
				}
				items = append(items, simulationItem{it.Section, it.Group, it.Instructions, it.Passage, it.Prompt,
					[4]string{it.Choices[0], it.Choices[1], it.Choices[2], it.Choices[3]}, it.Explanation})
				if prev, ok := cat.en[it.Explanation]; ok && prev != it.ExplanationEn {
					simulationJSONError = fmt.Errorf("%s: item %d has conflicting English text", entry.Name(), i+1)
					return
				}
				cat.en[it.Explanation] = it.ExplanationEn
			}
			cat.items[key] = items
		}
		simulationJSONData = cat
	})
	return simulationJSONData, simulationJSONError
}

// simulationJSONContent returns the JSON-backed items for a paper such as ("N5", 6),
// or nil when no file exists. A malformed embedded file is a build-time bug and panics.
func simulationJSONContent(base string, paperNo int) []simulationItem {
	cat, err := loadSimulationJSON()
	if err != nil {
		panic("store: invalid simulation_data: " + err.Error())
	}
	return cat.items[fmt.Sprintf("%s-%d", base, paperNo)]
}

// SimulationJSONExplanationsEN maps the Indonesian explanation of every JSON-backed
// practice question to its English translation.
func SimulationJSONExplanationsEN() map[string]string {
	cat, err := loadSimulationJSON()
	if err != nil {
		panic("store: invalid simulation_data: " + err.Error())
	}
	return cat.en
}
