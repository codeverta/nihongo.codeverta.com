package api

import (
	"bytes"
	"context"
	"crypto/subtle"
	_ "embed"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"nihongo-master-backend/internal/diagnostics"
)

//go:embed diagnostics_admin.html
var diagnosticsAdminHTML string

var (
	diagnosticEmailPattern  = regexp.MustCompile(`(?i)[\w.+-]+@[\w.-]+\.[a-z]{2,}`)
	diagnosticBearerPattern = regexp.MustCompile(`(?i)(bearer\s+)[a-z0-9._~+/-]+=*`)
	diagnosticJWTPattern    = regexp.MustCompile(`\beyJ[a-zA-Z0-9_-]{8,}\.[a-zA-Z0-9_-]{8,}\.[a-zA-Z0-9_-]{8,}\b`)
	diagnosticSecretPattern = regexp.MustCompile(`(?i)(token|password|secret|authorization)([=: ]+)[^\s&]+`)
	diagnosticURLPattern    = regexp.MustCompile(`(?i)https?://[^\s"'<>]+`)
	diagnosticUUIDPattern   = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f-]{27,}$`)
)

type diagnosticRateWindow struct {
	started time.Time
	count   int
}

type diagnosticRateLimiter struct {
	mu      sync.Mutex
	clients map[string]diagnosticRateWindow
}

func (s *Server) ConfigureDiagnostics(repository *diagnostics.Repository, token string) {
	s.diagnostics = repository
	s.diagnosticsAdminToken = token
	s.diagnosticLimiter.clients = make(map[string]diagnosticRateWindow)
}

func (s *Server) registerDiagnostics(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/diagnostics/errors", s.receiveFlutterDiagnostic)
	mux.HandleFunc("GET /admin/errors", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(diagnosticsAdminHTML))
	})
	mux.HandleFunc("GET /v1/admin/errors", s.diagnosticsAdmin(s.listDiagnosticEvents))
}

func (s *Server) diagnosticsAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if s.diagnosticsAdminToken == "" || subtle.ConstantTimeCompare([]byte(bearerToken(r)), []byte(s.diagnosticsAdminToken)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Kunci admin salah atau belum dikonfigurasi.")
			return
		}
		if s.diagnostics == nil {
			writeError(w, http.StatusServiceUnavailable, "database_unavailable", "Penyimpanan diagnostik belum tersedia.")
			return
		}
		next(w, r)
	}
}

func (s *Server) receiveFlutterDiagnostic(w http.ResponseWriter, r *http.Request) {
	if s.diagnostics == nil {
		writeError(w, http.StatusServiceUnavailable, "diagnostics_unavailable", "Pelaporan diagnostik belum tersedia.")
		return
	}
	if !s.diagnosticLimiter.allow(r.RemoteAddr, time.Now()) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "Terlalu banyak laporan diagnostik. Coba lagi nanti.")
		return
	}
	var input struct {
		Severity   string `json:"severity"`
		ErrorType  string `json:"errorType"`
		Message    string `json:"message"`
		StackTrace string `json:"stackTrace"`
		Platform   string `json:"platform"`
		AppVersion string `json:"appVersion"`
		Route      string `json:"route"`
	}
	payload, err := io.ReadAll(io.LimitReader(r.Body, 16*1024+1))
	if err != nil || len(payload) > 16*1024 {
		writeError(w, http.StatusRequestEntityTooLarge, "report_too_large", "Laporan diagnostik terlalu besar.")
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_report", "Laporan error tidak valid.")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_report", "Laporan error tidak valid.")
		return
	}
	input.Severity = strings.ToLower(strings.TrimSpace(input.Severity))
	if input.Severity != "error" && input.Severity != "crash" && input.Severity != "fatal" {
		input.Severity = "error"
	}
	if strings.TrimSpace(input.Message) == "" || len(input.Message) > 6000 || len(input.StackTrace) > 12000 || len(input.Route) > 512 {
		writeError(w, http.StatusUnprocessableEntity, "invalid_report", "Laporan wajib berisi pesan dan tidak boleh melebihi batas ukuran.")
		return
	}
	event := diagnostics.Event{
		Source:     "flutter",
		Severity:   input.Severity,
		ErrorType:  sanitizeDiagnosticText(input.ErrorType),
		Message:    sanitizeDiagnosticText(input.Message),
		StackTrace: sanitizeDiagnosticText(input.StackTrace),
		Platform:   safeDiagnosticLabel(sanitizeDiagnosticText(input.Platform), 32),
		AppVersion: safeDiagnosticLabel(sanitizeDiagnosticText(input.AppVersion), 48),
		Route:      diagnosticRoute(sanitizeDiagnosticText(input.Route)),
		Method:     "APP",
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.diagnostics.Record(ctx, event); err != nil {
		writeError(w, http.StatusServiceUnavailable, "diagnostics_write_failed", "Laporan diagnostik belum dapat disimpan.")
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) listDiagnosticEvents(w http.ResponseWriter, r *http.Request) {
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 100000 {
			writeError(w, http.StatusBadRequest, "invalid_page", "Halaman tidak valid.")
			return
		}
		page = parsed
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) > 120 {
		writeError(w, http.StatusBadRequest, "invalid_query", "Pencarian maksimal 120 karakter.")
		return
	}
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	if source != "" && source != "flutter" && source != "backend" {
		writeError(w, http.StatusBadRequest, "invalid_source", "Sumber harus flutter atau backend.")
		return
	}
	events, total, err := s.diagnostics.List(r.Context(), page, query, source)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "diagnostics_read_failed", "Daftar error belum dapat dibaca.")
		return
	}
	stats, err := s.diagnostics.Stats(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "diagnostics_read_failed", "Ringkasan error belum dapat dibaca.")
		return
	}
	writeData(w, http.StatusOK, map[string]any{"items": events, "total": total, "page": page, "pageSize": 50, "stats": stats})
}

// ReportBackendWorkerError records failures from scheduled tasks that do not
// produce an HTTP response, such as an interrupted import or source fetch.
func (s *Server) ReportBackendWorkerError(component string, err error) {
	if err == nil || s.diagnostics == nil {
		return
	}
	component = safeDiagnosticLabel(component, 80)
	message := err.Error()
	event := diagnostics.Event{
		Source: "backend", Severity: "error", ErrorType: "worker/" + component,
		Message: sanitizeDiagnosticText(message), Platform: "server", Route: "worker/" + component,
	}
	if strings.HasPrefix(message, "panic: ") {
		event.Severity = "crash"
		event.ErrorType = "worker/" + component + "/panic"
		panicMessage, stack, hasStack := strings.Cut(message, "\n")
		event.Message = sanitizeDiagnosticText(panicMessage)
		if hasStack {
			event.StackTrace = sanitizeDiagnosticText(stack)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if recordErr := s.diagnostics.Record(ctx, event); recordErr != nil {
		log.Printf("store worker diagnostic (%s): %v", component, recordErr)
	}
}

func (l *diagnosticRateLimiter) allow(remoteAddr string, now time.Time) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	if host == "" {
		host = "unknown"
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.clients == nil {
		l.clients = make(map[string]diagnosticRateWindow)
	}
	for address, window := range l.clients {
		if now.Sub(window.started) >= time.Minute {
			delete(l.clients, address)
		}
	}
	window := l.clients[host]
	if window.started.IsZero() || now.Sub(window.started) >= time.Minute {
		if len(l.clients) >= 10000 {
			return false
		}
		l.clients[host] = diagnosticRateWindow{started: now, count: 1}
		return true
	}
	if window.count >= 120 {
		return false
	}
	window.count++
	l.clients[host] = window
	return true
}

func safeDiagnosticLabel(value string, max int) string {
	value = strings.TrimSpace(value)
	var clean strings.Builder
	for _, char := range value {
		if char >= 32 && char != 127 {
			clean.WriteRune(char)
		}
	}
	return truncateDiagnostic(clean.String(), max)
}

func sanitizeDiagnosticText(value string) string {
	value = diagnosticEmailPattern.ReplaceAllString(value, "[redacted-email]")
	value = diagnosticBearerPattern.ReplaceAllString(value, "${1}[redacted]")
	value = diagnosticJWTPattern.ReplaceAllString(value, "[redacted-token]")
	value = diagnosticSecretPattern.ReplaceAllString(value, "${1}${2}[redacted]")
	value = diagnosticURLPattern.ReplaceAllString(value, "[redacted-url]")
	return truncateDiagnostic(value, 12000)
}

func diagnosticRoute(path string) string {
	path = sanitizeDiagnosticText(strings.SplitN(strings.TrimSpace(path), "?", 2)[0])
	if path == "" {
		return "unknown"
	}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "" {
			continue
		}
		if diagnosticUUIDPattern.MatchString(part) || len(part) > 80 || (i > 0 && isDynamicRouteSegment(parts[i-1])) {
			parts[i] = ":id"
		}
	}
	return truncateDiagnostic(strings.Join(parts, "/"), 512)
}

func isDynamicRouteSegment(previous string) bool {
	switch previous {
	case "attempts", "lessons", "news", "manga", "series", "chapters", "users", "accounts", "errors":
		return true
	default:
		return false
	}
}

func truncateDiagnostic(value string, max int) string {
	runes := []rune(value)
	if len(runes) > max {
		return string(runes[:max])
	}
	return value
}
