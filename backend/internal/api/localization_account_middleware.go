package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

// accountLocalizedPrefixes lists public JSON routes that are not covered by the
// learning or news localizers. Admin routes stay in Indonesian.
var accountLocalizedPrefixes = []string{
	"/v1/me", "/v1/billing", "/v1/plans", "/v1/pricing", "/v1/auth",
	"/v1/jobs", "/v1/manga", "/v1/anki", "/v1/app-update-policy", "/v1/branding",
}

func isAccountLocalizedPath(path string) bool {
	if strings.HasPrefix(path, "/v1/admin/") || isLocalizedLearningPath(path) || isPublicNewsPath(path) {
		return false
	}
	for _, prefix := range accountLocalizedPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

// localizedAccountResponses translates app-authored success text and error
// messages for English clients on account, billing, jobs, manga, Anki and
// branding routes. Indonesian clients and non-JSON responses (file downloads,
// images) pass through untouched.
func localizedAccountResponses(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAccountLocalizedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		locale := requestContentLanguage(r.Header.Get("Accept-Language"))
		if locale == localeEnglish && !englishLearningContentEnabled() {
			locale = localeIndonesian
		}
		addVaryHeader(w.Header(), "Accept-Language")
		if locale != localeEnglish {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Language", locale)
		writer := &accountLocalizingWriter{ResponseWriter: w}
		next.ServeHTTP(writer, r)
		if !writer.buffering {
			return
		}
		body := writer.body.Bytes()
		if localized, ok := localizeAccountJSON(r.URL.Path, body); ok {
			body = localized
			w.Header().Del("Content-Length")
		}
		w.WriteHeader(writer.status)
		if len(body) > 0 {
			_, _ = w.Write(body)
		}
	})
}

// localizeAccountJSON returns the translated body and true when anything changed.
func localizeAccountJSON(path string, body []byte) ([]byte, bool) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return body, false
	}
	changed := false
	if root, ok := payload.(map[string]any); ok {
		if errorObject, ok := root["error"].(map[string]any); ok {
			code, _ := errorObject["code"].(string)
			source, _ := errorObject["message"].(string)
			if message := englishAPIErrorMessageFor(code, source); message != source {
				errorObject["message"] = message
				changed = true
			}
		}
	}
	if path == "/v1/app-update-policy" && localizeAppUpdateMessage(payload) {
		changed = true
	}
	if translateAccountPayload(path, payload) {
		changed = true
	}
	if !changed {
		return body, false
	}
	localized, err := json.Marshal(payload)
	if err != nil {
		return body, false
	}
	return localized, true
}

// localizeAppUpdateMessage swaps the admin-written Indonesian note for its
// English counterpart; without one, English users see only the built-in text.
func localizeAppUpdateMessage(payload any) bool {
	root, _ := payload.(map[string]any)
	data, _ := root["data"].(map[string]any)
	changed := false
	for _, platform := range []string{"android", "ios"} {
		policy, ok := data[platform].(map[string]any)
		if !ok {
			continue
		}
		english, _ := policy["messageEn"].(string)
		if existing, _ := policy["message"].(string); existing != english {
			policy["message"] = english
			changed = true
		}
	}
	return changed
}

// accountLocalizingWriter buffers JSON responses and streams everything else.
type accountLocalizingWriter struct {
	http.ResponseWriter
	status    int
	decided   bool
	buffering bool
	body      bytes.Buffer
}

func (w *accountLocalizingWriter) WriteHeader(status int) {
	if w.decided {
		return
	}
	w.decided = true
	w.status = status
	w.buffering = strings.HasPrefix(w.Header().Get("Content-Type"), "application/json")
	if !w.buffering {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *accountLocalizingWriter) Write(data []byte) (int, error) {
	if !w.decided {
		w.WriteHeader(http.StatusOK)
	}
	if w.buffering {
		return w.body.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

func (w *accountLocalizingWriter) Flush() {
	if w.buffering {
		return
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *accountLocalizingWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
