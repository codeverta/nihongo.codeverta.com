package store

import "nihongo-master-backend/data/kanji"

// seedKanjiN1 loads the independently maintained N1 curriculum. Shared level
// registration is kept in seedKanji() so the N2/N1 integration can be ordered
// together without duplicating or conflicting edits.
func seedKanjiN1() []kanjiLessonContent {
	return seedKanjiLevel("N1", kanji.N1, 247, 1232)
}
