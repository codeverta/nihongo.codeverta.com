package store

import "testing"

func TestAssetURLs(t *testing.T) {
	s := NewMemoryStore()
	key := "assets/images/reading/story-01.webp"
	if got := s.assetURLLocked(key); got != "/"+key {
		t.Fatalf("local URL = %q", got)
	}
	s.SetAssetBaseURL("https://cdn.example.com/")
	if got := s.assetURLLocked(key); got != "https://cdn.example.com/"+key {
		t.Fatalf("CDN URL = %q", got)
	}
	if got := s.assetURLLocked(""); got != "" {
		t.Fatalf("empty asset URL = %q", got)
	}
	levels := s.VocabularyLevels("user-demo")
	if levels[0].Lessons[0].ImageURL == "" {
		t.Fatal("lesson thumbnails require remote URLs")
	}
}
