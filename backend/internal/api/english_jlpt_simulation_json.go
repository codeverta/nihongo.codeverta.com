package api

import "nihongo-master-backend/internal/store"

// englishJLPTSimulationJSONText returns the English explanation of a JSON-backed
// JLPT practice question (papers 06 and later), keyed by the Indonesian text.
func englishJLPTSimulationJSONText(field, source string) (string, bool) {
	if field != "simulation.explanation" {
		return "", false
	}
	text, ok := store.SimulationJSONExplanationsEN()[source]
	return text, ok
}
