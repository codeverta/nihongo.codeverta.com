package store

import (
	"strings"
	"testing"
)

func TestSimulationJSONPapersLoad(t *testing.T) {
	if _, err := loadSimulationJSON(); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		base  string
		paper int
		total int
	}{{"N5", 6, 38}, {"N5", 7, 38}, {"N4", 6, 57}, {"N4", 7, 57},
		{"N5", 8, 38}, {"N5", 9, 38}, {"N5", 10, 38}, {"N4", 8, 57}, {"N4", 9, 57}, {"N4", 10, 57}} {
		items := simulationJSONContent(c.base, c.paper)
		if len(items) != c.total {
			t.Fatalf("%s-%d has %d items, want %d", c.base, c.paper, len(items), c.total)
		}
		for i, item := range items {
			seen := map[string]bool{}
			for _, choice := range item.choices {
				if strings.TrimSpace(choice) == "" || seen[choice] {
					t.Fatalf("%s-%d item %d has empty or duplicate choices: %v", c.base, c.paper, i+1, item.choices)
				}
				seen[choice] = true
			}
			if item.explanation == "" {
				t.Fatalf("%s-%d item %d has no explanation", c.base, c.paper, i+1)
			}
		}
	}
	if simulationJSONContent("N5", 11) != nil {
		t.Fatal("unexpected paper")
	}
	if len(SimulationJSONExplanationsEN()) == 0 {
		t.Fatal("missing English explanations")
	}
}
