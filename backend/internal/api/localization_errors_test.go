package api

import (
	"strings"
	"testing"
)

const testAPIErrorFallback = "The request could not be completed. Please try again."

func TestEnglishAPIErrorMessageCoversKnownCodes(t *testing.T) {
	for _, code := range []string{
		"invalid_json",
		"invalid_input",
		"not_found",
		"premium_required",
		"unauthorized",
		"internal_error",
		"invalid_credentials",
		"billing_unavailable",
		"manga_busy",
		"news_unavailable",
		"file_too_large",
		"photo_too_large",
		"save_failed",
	} {
		t.Run(code, func(t *testing.T) {
			message := englishAPIErrorMessage(code)
			if strings.TrimSpace(message) == "" {
				t.Fatalf("englishAPIErrorMessage(%q) is empty", code)
			}
			if message == testAPIErrorFallback {
				t.Fatalf("englishAPIErrorMessage(%q) returned the generic fallback", code)
			}
		})
	}
}

func TestEnglishAPIErrorMessageFallsBackForUnknownCode(t *testing.T) {
	if got := englishAPIErrorMessage("unknown-code"); got != testAPIErrorFallback {
		t.Fatalf("unknown code message = %q, want %q", got, testAPIErrorFallback)
	}
	if got := englishAPIErrorMessageFor("unknown-code", "Pesan tidak dikenal."); got != testAPIErrorFallback {
		t.Fatalf("unknown code with unknown source = %q, want %q", got, testAPIErrorFallback)
	}
}

func TestEnglishAPIErrorMessageForPrefersExactSource(t *testing.T) {
	got := englishAPIErrorMessageFor("file_too_large", "Ukuran file maksimal 20 MB.")
	if got != "The file size limit is 20 MB." {
		t.Fatalf("exact source message = %q", got)
	}
	got = englishAPIErrorMessageFor("file_too_large", "something else")
	if got != englishAPIErrorMessage("file_too_large") {
		t.Fatalf("unmatched source should use code message, got %q", got)
	}
}

func TestEnglishAPIErrorMessagesContainNoIndonesianMarkers(t *testing.T) {
	markers := map[string]bool{"tidak": true, "silakan": true, "gagal": true, "belum": true, "dengan": true, "yang": true}
	check := func(name, key, value string) {
		if strings.TrimSpace(value) == "" {
			t.Errorf("%s[%q] has an empty message", name, key)
		}
		words := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
			return (r < 'a' || r > 'z') && (r < '0' || r > '9')
		})
		for _, word := range words {
			if markers[word] {
				t.Errorf("%s[%q] contains Indonesian word %q: %q", name, key, word, value)
			}
		}
	}
	for code, message := range apiErrorMessagesEnglish {
		check("apiErrorMessagesEnglish", code, message)
	}
	for source, message := range apiErrorMessagesEnglishBySource {
		check("apiErrorMessagesEnglishBySource", source, message)
	}
}
