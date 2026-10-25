package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

// localizedNewsResponses translates app-authored metadata and errors on public
// news reads. The Japanese article content and source-provided fields remain unchanged.
func localizedNewsResponses(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !isPublicNewsPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		locale := requestContentLanguage(r.Header.Get("Accept-Language"))
		if locale == localeEnglish && !englishLearningContentEnabled() {
			locale = localeIndonesian
		}
		addVaryHeader(w.Header(), "Accept-Language")
		w.Header().Set("Content-Language", locale)
		w.Header().Set("Cache-Control", "private, no-store")

		buffered := &localizedResponseBuffer{ResponseWriter: w}
		next.ServeHTTP(buffered, r)
		body := buffered.body.Bytes()
		status := buffered.status
		if status == 0 {
			status = http.StatusOK
		}
		if locale == localeEnglish {
			var localized []byte
			var changed bool
			var err error
			switch {
			case status == http.StatusOK:
				localized, changed, err = translateNewsJSON(body, translateNewsMetadataValue)
			case status >= http.StatusBadRequest:
				localized, changed, err = translateNewsJSON(body, translateNewsErrorValue)
			}
			if err == nil && changed {
				body = localized
				w.Header().Del("Content-Length")
			}
		}

		w.WriteHeader(status)
		if len(body) > 0 {
			_, _ = w.Write(body)
		}
	})
}

func isPublicNewsPath(path string) bool {
	return path == "/v1/news" || strings.HasPrefix(path, "/v1/news/")
}

func translateNewsErrorJSON(body []byte) ([]byte, bool, error) {
	return translateNewsJSON(body, translateNewsErrorValue)
}

func translateNewsJSON(body []byte, translate func(any) bool) ([]byte, bool, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return body, false, err
	}
	if !translate(payload) {
		return body, false, nil
	}
	localized, err := json.Marshal(payload)
	if err != nil {
		return body, false, err
	}
	return localized, true, nil
}

func translateNewsMetadataValue(value any) bool {
	changed := false
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			if key == "changes" {
				if text, ok := child.(string); ok {
					if translated, found := newsChangesTranslations[text]; found {
						value[key] = translated
						changed = true
					}
				}
				continue
			}
			changed = translateNewsMetadataValue(child) || changed
		}
	case []any:
		for _, child := range value {
			changed = translateNewsMetadataValue(child) || changed
		}
	}
	return changed
}

func translateNewsErrorValue(value any) bool {
	changed := false
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			if key == "error" {
				if errorObject, ok := child.(map[string]any); ok {
					code, _ := errorObject["code"].(string)
					message := englishNewsErrorMessage(code)
					if existing, ok := errorObject["message"].(string); !ok || existing != message {
						errorObject["message"] = message
						changed = true
					}
				}
				continue
			}
			changed = translateNewsErrorValue(child) || changed
		}
	case []any:
		for _, child := range value {
			changed = translateNewsErrorValue(child) || changed
		}
	}
	return changed
}

func englishNewsErrorMessage(code string) string {
	if message, ok := newsErrorTranslations[code]; ok {
		return message
	}
	return "The news request could not be completed."
}

var newsChangesTranslations = map[string]string{
	"Furigana ditambahkan oleh Nihongo Master; media eksternal tidak disertakan.":                                                                     "Furigana added by Nihongo Master; external media is not included.",
	"Furigana ditambahkan oleh Nihongo Master. Gambar berizin disalin dan dikonversi ke JPEG bila tersedia.":                                          "Furigana added by Nihongo Master. Licensed images are copied and converted to JPEG when available.",
	"Teks HTML Osaka City; furigana ditambahkan oleh Nihongo Master. Hanya gambar dengan lisensi tersendiri yang disalin; lampiran tidak disertakan.": "Osaka City HTML text; furigana added by Nihongo Master. Only images with a separate license are copied; attachments are not included.",
}

var newsErrorTranslations = map[string]string{
	"invalid_news":     "The news request is invalid. Check the settings or article details.",
	"news_not_found":   "The article or source could not be found.",
	"news_busy":        "This source is being processed or was checked recently. Please try again shortly.",
	"news_changed":     "This article changed since the editor was opened. Reload it before editing readings.",
	"news_unavailable": "News is temporarily unavailable. Saved articles are still available.",
}
