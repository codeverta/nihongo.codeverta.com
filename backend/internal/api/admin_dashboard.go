package api

import (
	"crypto/rand"
	"crypto/subtle"
	_ "embed"
	"encoding/base64"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:embed admin_dashboard.html
var adminDashboardHTML string

const (
	adminSessionCookie = "nm_admin_session"
	adminSessionTTL    = 12 * time.Hour
	defaultAdminEmail  = "demo@nihongomaster.app"
)

type adminSession struct {
	email   string
	name    string
	expires time.Time
}

// adminAuthState keeps dashboard sessions in memory. Sessions end when the
// API restarts; admins simply sign in again.
type adminAuthState struct {
	mu       sync.Mutex
	sessions map[string]adminSession
	attempts map[string][]time.Time
	emails   map[string]bool
	explicit bool
}

var adminAuth = &adminAuthState{
	sessions: map[string]adminSession{},
	attempts: map[string][]time.Time{},
	emails:   map[string]bool{defaultAdminEmail: true},
}

// ConfigureAdminAccounts sets which app accounts may sign in to /admin/.
// The value is a comma-separated email list (ADMIN_EMAILS). When empty, only
// the development demo account is allowed, and only from a local or private
// network request that did not pass through a reverse proxy.
func (s *Server) ConfigureAdminAccounts(list string) {
	adminAuth.mu.Lock()
	defer adminAuth.mu.Unlock()
	adminAuth.emails = map[string]bool{}
	for _, email := range strings.Split(list, ",") {
		if email = strings.ToLower(strings.TrimSpace(email)); email != "" {
			adminAuth.emails[email] = true
		}
	}
	adminAuth.explicit = len(adminAuth.emails) > 0
	if !adminAuth.explicit {
		adminAuth.emails[defaultAdminEmail] = true
	}
}

func adminEmailAllowed(r *http.Request, email string) bool {
	adminAuth.mu.Lock()
	allowed, explicit := adminAuth.emails[email], adminAuth.explicit
	adminAuth.mu.Unlock()
	if !allowed {
		return false
	}
	return explicit || isLocalAdminRequest(r)
}

func isLocalAdminRequest(r *http.Request) bool {
	if r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Real-IP") != "" || r.Header.Get("Forwarded") != "" {
		return false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsPrivate())
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// adminLoginAllowed limits sign-in attempts to 10 per 15 minutes per address.
func adminLoginAllowed(ip string, now time.Time) bool {
	adminAuth.mu.Lock()
	defer adminAuth.mu.Unlock()
	recent := adminAuth.attempts[ip][:0]
	for _, at := range adminAuth.attempts[ip] {
		if now.Sub(at) < 15*time.Minute {
			recent = append(recent, at)
		}
	}
	if len(recent) >= 10 {
		adminAuth.attempts[ip] = recent
		return false
	}
	adminAuth.attempts[ip] = append(recent, now)
	return true
}

func adminSessionFromRequest(r *http.Request) (adminSession, bool) {
	cookie, err := r.Cookie(adminSessionCookie)
	if err != nil || cookie.Value == "" {
		return adminSession{}, false
	}
	adminAuth.mu.Lock()
	defer adminAuth.mu.Unlock()
	session, ok := adminAuth.sessions[cookie.Value]
	if !ok {
		return adminSession{}, false
	}
	if time.Now().After(session.expires) {
		delete(adminAuth.sessions, cookie.Value)
		return adminSession{}, false
	}
	return session, true
}

// adminRequestAuthorized accepts a signed-in dashboard session, or the legacy
// per-tool admin token sent as a Bearer header.
func adminRequestAuthorized(r *http.Request, token string) bool {
	if _, ok := adminSessionFromRequest(r); ok {
		return true
	}
	return token != "" && subtle.ConstantTimeCompare([]byte(bearerToken(r)), []byte(token)) == 1
}

func secureRequest(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func (s *Server) registerAdminDashboard(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/{$}", func(w http.ResponseWriter, _ *http.Request) {
		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("Cache-Control", "no-store")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; img-src 'self' https: data: blob:; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		_, _ = w.Write([]byte(adminDashboardHTML))
	})
	mux.HandleFunc("POST /v1/admin/session", s.adminLogin)
	mux.HandleFunc("GET /v1/admin/session", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		session, ok := adminSessionFromRequest(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Silakan masuk ke dashboard admin.")
			return
		}
		writeData(w, http.StatusOK, map[string]string{"email": session.email, "displayName": session.name})
	})
	mux.HandleFunc("DELETE /v1/admin/session", func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(adminSessionCookie); err == nil {
			adminAuth.mu.Lock()
			delete(adminAuth.sessions, cookie.Value)
			adminAuth.mu.Unlock()
		}
		http.SetCookie(w, &http.Cookie{Name: adminSessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: secureRequest(r)})
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /v1/admin/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if _, ok := adminSessionFromRequest(r); !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Silakan masuk ke dashboard admin.")
			return
		}
		writeData(w, http.StatusOK, map[string]any{
			"users":       s.store.AdminUsers(),
			"generatedAt": time.Now().UTC(),
		})
	})
}

func (s *Server) adminLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Format permintaan tidak valid.")
		return
	}
	if !adminLoginAllowed(clientIP(r), time.Now()) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "Terlalu banyak percobaan. Coba lagi dalam 15 menit.")
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	profile, err := s.store.AuthenticateEmail(email, input.Password)
	if err != nil || !adminEmailAllowed(r, email) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "Email atau password salah, atau akun ini tidak punya akses admin.")
		return
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Sesi belum dapat dibuat.")
		return
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	adminAuth.mu.Lock()
	now := time.Now()
	for key, session := range adminAuth.sessions {
		if now.After(session.expires) {
			delete(adminAuth.sessions, key)
		}
	}
	adminAuth.sessions[token] = adminSession{email: email, name: profile.DisplayName, expires: now.Add(adminSessionTTL)}
	adminAuth.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name: adminSessionCookie, Value: token, Path: "/",
		MaxAge: int(adminSessionTTL.Seconds()), HttpOnly: true,
		SameSite: http.SameSiteStrictMode, Secure: secureRequest(r),
	})
	writeData(w, http.StatusOK, map[string]string{"email": email, "displayName": profile.DisplayName})
}

// adminToolRedirects sends the old standalone /admin/<tool> pages to the
// matching native view inside the dashboard.
var adminToolRoutes = map[string]string{
	"/admin/news": "news", "/admin/manga": "manga", "/admin/dictionary": "dictionary",
	"/admin/map": "map", "/admin/home-banners": "home-banners", "/admin/notifications": "notifications",
	"/admin/pricing": "pricing", "/admin/ads": "ads", "/admin/app-updates": "app-updates",
	"/admin/branding": "branding", "/admin/errors": "errors",
}

func adminToolRedirects(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if id, ok := adminToolRoutes[strings.TrimSuffix(r.URL.Path, "/")]; ok {
				http.Redirect(w, r, "/admin/#/tools/"+id, http.StatusFound)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
