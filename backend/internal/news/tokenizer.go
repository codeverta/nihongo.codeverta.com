package news

import (
	"strings"
	"sync"
	"unicode"
	"unicode/utf16"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

var analyzerOnce sync.Once
var analyzer *tokenizer.Tokenizer
var analyzerError error
var analyzerMu sync.Mutex

func prepareAnalyzer() error {
	analyzerOnce.Do(func() { analyzer, analyzerError = tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos()) })
	return analyzerError
}
func utf16len(s string) int { return len(utf16.Encode([]rune(s))) }
func hiragana(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'ァ' && r <= 'ヶ' {
			return r - 0x60
		}
		return r
	}, s)
}
func hasKanji(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// Reading returns the hiragana reading of Japanese text, or "" when any part
// containing kanji has no dictionary reading (for example rare place names).
func Reading(text string) string {
	if strings.TrimSpace(text) == "" || prepareAnalyzer() != nil {
		return ""
	}
	analyzerMu.Lock()
	defer analyzerMu.Unlock()
	var out strings.Builder
	for _, part := range analyzer.Tokenize(text) {
		reading, ok := part.Reading()
		if !ok || reading == "" || reading == "*" {
			if hasKanji(part.Surface) {
				return ""
			}
			reading = part.Surface
		}
		out.WriteString(hiragana(reading))
	}
	return out.String()
}

// Token positions always address the untouched selectable base text, never ruby.
func tokenize(text string) []Token {
	tokens := []Token{}
	if prepareAnalyzer() != nil {
		return tokens
	}
	analyzerMu.Lock()
	defer analyzerMu.Unlock()
	position := 0
	for _, part := range analyzer.Tokenize(text) {
		// Include skipped spaces in offsets. Kagome's own positions are rune-based.
		index := strings.Index(text[position:], part.Surface)
		if index < 0 {
			continue
		}
		start := position + index
		end := start + len(part.Surface)
		base, ok := part.BaseForm()
		if !ok || base == "*" {
			base = part.Surface
		}
		reading, ok := part.Reading()
		origin := "unknown"
		if ok && reading != "*" && part.Class != tokenizer.UNKNOWN && hasKanji(part.Surface) {
			reading = hiragana(reading)
			origin = "automatic"
		} else {
			reading = ""
		}
		tokens = append(tokens, Token{Start: utf16len(text[:start]), End: utf16len(text[:end]), Surface: part.Surface, Base: base, Reading: reading, ReadingSource: origin})
		position = end
	}
	return tokens
}
