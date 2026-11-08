package store

import "testing"

func TestVocabularySeedIncludesAllJLPTLevels(t *testing.T) {
	levels := seedVocabulary()
	want := map[string]struct {
		lessons int
		words   int
	}{
		"N5": {lessons: 20, words: 160},
		"N4": {lessons: 34, words: 630},
		"N3": {lessons: 6, words: 72},
		"N2": {lessons: 5, words: 40},
		"N1": {lessons: 8, words: 64},
	}
	if len(levels) != len(want) {
		t.Fatalf("level count = %d, want %d", len(levels), len(want))
	}
	for _, level := range levels {
		expected, ok := want[level.Level]
		if !ok {
			t.Fatalf("unexpected vocabulary level %q", level.Level)
		}
		wordCount := 0
		ids := map[string]bool{}
		for _, lesson := range level.Lessons {
			for _, word := range lesson.Words {
				wordCount++
				if ids[word.ID] {
					t.Errorf("duplicate %s word ID %q", level.Level, word.ID)
				}
				ids[word.ID] = true
			}
		}
		if len(level.Lessons) != expected.lessons || wordCount != expected.words {
			t.Errorf("%s has %d lessons/%d words, want %d/%d", level.Level, len(level.Lessons), wordCount, expected.lessons, expected.words)
		}
	}
}

func TestN5LessonSummariesKeepImageAssetURLs(t *testing.T) {
	store := NewMemoryStore()
	levels := store.VocabularyLevels("vocabulary-test-user")
	if got := levels[0].Lessons[0].ImageAsset; got != "assets/images/vocabulary/n5-theme-01.webp" {
		t.Fatalf("N5 thumbnail asset = %q", got)
	}
	if got := levels[0].Lessons[0].ImageURL; got != "/assets/images/vocabulary/n5-theme-01.webp?v="+assetImageVersion {
		t.Fatalf("N5 thumbnail URL = %q", got)
	}
}

func TestN4LevelIncludesOpenJLPTAttribution(t *testing.T) {
	levels := NewMemoryStore().VocabularyLevels("vocabulary-test-user")
	level := levels[1]
	if level.Level != "N4" || level.SourceName != "OpenJLPT" || level.SourceURL == "" || level.SourceLicense == "" {
		t.Fatalf("N4 attribution is incomplete: %#v", level)
	}
}
