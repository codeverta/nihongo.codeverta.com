package api

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"nihongo-master-backend/internal/notifications"
)

//go:embed notifications_admin.html
var notificationsAdminHTML string

//go:embed admin.html
var adminHomeHTML string

type Announcement struct {
	ID string `json:"id"`
	notifications.Message
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	SentAt    *time.Time `json:"sentAt,omitempty"`
	FCMName   string     `json:"fcmName,omitempty"`
	Error     string     `json:"error,omitempty"`
}
type NotificationStore struct {
	mu          sync.Mutex
	path, token string
	sender      notifications.Sender
	items       []Announcement
}

func (s *Server) ConfigureNotifications(path, token string, sender notifications.Sender) error {
	n := &NotificationStore{path: path, token: token, sender: sender, items: []Announcement{}}
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err = json.Unmarshal(data, &n.items); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	// A crash after FCM accepted a request must never cause an automatic resend.
	for i := range n.items {
		if n.items[i].Status == "sending" {
			n.items[i].Status = "uncertain"
			n.items[i].Error = "Server dimulai ulang saat pengiriman. Jangan kirim ulang sebelum memastikan pesan sebelumnya."
		}
	}
	s.notifications = n
	return nil
}

var announcementRoutes = map[string]bool{"/home": true, "/dictionary": true, "/anki": true, "/ssw": true, "/premium": true, "/learn/vocabulary": true, "/learn/grammar": true, "/learn/kanji": true, "/learn/reading": true, "/exams/n5": true}

func validateAnnouncement(m *notifications.Message) error {
	m.Title = strings.TrimSpace(m.Title)
	m.Body = strings.TrimSpace(m.Body)
	if len([]rune(m.Title)) < 1 || len([]rune(m.Title)) > 80 || len([]rune(m.Body)) < 1 || len([]rune(m.Body)) > 500 {
		return fmt.Errorf("judul wajib 1–80 karakter dan isi wajib 1–500 karakter")
	}
	if !announcementRoutes[m.Route] {
		return fmt.Errorf("pilih halaman tujuan dari daftar")
	}
	return notifications.ValidateSize(*m)
}
func (n *NotificationStore) save(items []Announcement) error {
	if n.path != "" {
		if err := os.MkdirAll(filepath.Dir(n.path), 0700); err != nil {
			return err
		}
		raw, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		if err = os.WriteFile(n.path+".tmp", raw, 0600); err != nil {
			return err
		}
		if err = os.Rename(n.path+".tmp", n.path); err != nil {
			return err
		}
	}
	n.items = items
	return nil
}
func (s *Server) registerNotifications(mux *http.ServeMux) {
	if s.notifications == nil {
		s.notifications = &NotificationStore{items: []Announcement{}}
	}
	page := func(html string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
			w.Write([]byte(html))
		}
	}
	mux.HandleFunc("GET /admin/notifications", page(notificationsAdminHTML))
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			n := s.notifications
			if !adminRequestAuthorized(r, n.token) {
				writeError(w, 401, "unauthorized", "Kunci admin salah atau belum diatur di server.")
				return
			}
			next(w, r)
		}
	}
	mux.HandleFunc("GET /v1/admin/notifications", auth(s.listAnnouncements))
	mux.HandleFunc("POST /v1/admin/notifications", auth(s.createAnnouncement))
	mux.HandleFunc("PUT /v1/admin/notifications/{id}", auth(s.editAnnouncement))
	mux.HandleFunc("POST /v1/admin/notifications/{id}/send", auth(s.sendAnnouncement))
	mux.HandleFunc("POST /v1/admin/notifications/check", auth(s.checkNotificationSender))
}
func (s *Server) listAnnouncements(w http.ResponseWriter, r *http.Request) {
	n := s.notifications
	n.mu.Lock()
	defer n.mu.Unlock()
	writeData(w, 200, map[string]any{"ready": n.sender != nil, "items": n.items, "audience": "Pengguna Android yang mengaktifkan pengumuman Nihongo Master"})
}
func (s *Server) createAnnouncement(w http.ResponseWriter, r *http.Request) {
	var m notifications.Message
	if err := decodeJSON(r, &m); err != nil {
		writeError(w, 400, "invalid_json", "Format pesan tidak valid.")
		return
	}
	if err := validateAnnouncement(&m); err != nil {
		writeError(w, 422, "invalid_message", err.Error())
		return
	}
	n := s.notifications
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.items) >= 500 {
		writeError(w, 409, "history_full", "Riwayat mencapai 500 pesan. Minta operator mengarsipkan riwayat sebelum membuat pesan baru.")
		return
	}
	now := time.Now().UTC()
	a := Announcement{ID: uuid.NewString(), Message: m, Status: "draft", CreatedAt: now, UpdatedAt: now}
	items := append([]Announcement{a}, n.items...)
	if err := n.save(items); err != nil {
		writeError(w, 500, "save_failed", "Draft gagal disimpan.")
		return
	}
	writeData(w, 201, a)
}
func (s *Server) editAnnouncement(w http.ResponseWriter, r *http.Request) {
	var m notifications.Message
	if decodeJSON(r, &m) != nil {
		writeError(w, 400, "invalid_json", "Format pesan tidak valid.")
		return
	}
	if err := validateAnnouncement(&m); err != nil {
		writeError(w, 422, "invalid_message", err.Error())
		return
	}
	n := s.notifications
	n.mu.Lock()
	defer n.mu.Unlock()
	for i, a := range n.items {
		if a.ID != r.PathValue("id") {
			continue
		}
		if a.Status != "draft" && a.Status != "failed" {
			writeError(w, 409, "not_draft", "Hanya draft atau pesan yang ditolak Firebase yang bisa diubah.")
			return
		}
		items := append([]Announcement(nil), n.items...)
		a.Message = m
		a.Status = "draft"
		a.Error = ""
		a.UpdatedAt = time.Now().UTC()
		items[i] = a
		if n.save(items) != nil {
			writeError(w, 500, "save_failed", "Draft gagal disimpan.")
			return
		}
		writeData(w, 200, a)
		return
	}
	writeError(w, 404, "not_found", "Pesan tidak ditemukan.")
}
func (s *Server) sendAnnouncement(w http.ResponseWriter, r *http.Request) {
	n := s.notifications
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.sender == nil {
		writeError(w, 503, "not_configured", "Pengirim Firebase belum diaktifkan di server.")
		return
	}
	for i, a := range n.items {
		if a.ID != r.PathValue("id") {
			continue
		}
		if a.Status != "draft" {
			writeError(w, 409, "already_processed", "Pesan sudah diproses. Pengiriman ulang otomatis dicegah.")
			return
		}
		var input struct {
			ConfirmID string `json:"confirmId"`
		}
		if decodeJSON(r, &input) != nil || input.ConfirmID != a.ID {
			writeError(w, 422, "confirm_required", "Konfirmasi pesan sebelum mengirim.")
			return
		}
		items := append([]Announcement(nil), n.items...)
		a.Status = "sending"
		a.UpdatedAt = time.Now().UTC()
		items[i] = a
		if n.save(items) != nil {
			writeError(w, 500, "save_failed", "Status pengiriman gagal disimpan. Pesan belum dikirim.")
			return
		}
		name, err := n.sender.Send(r.Context(), a.Message, a.ID, false)
		now := time.Now().UTC()
		a.UpdatedAt = now
		if err != nil {
			a.Status = "failed"
			if notifications.IsUncertain(err) {
				a.Status = "uncertain"
			}
			a.Error = err.Error()
		} else {
			a.Status = "sent"
			a.FCMName = name
			a.SentAt = &now
		}
		items = append([]Announcement(nil), n.items...)
		items[i] = a
		if n.save(items) != nil {
			writeError(w, 500, "save_failed", "Hasil pengiriman belum tersimpan. Jangan kirim ulang; minta operator memeriksa riwayat.")
			return
		}
		// Also return rejection/uncertainty as data so the admin always sees the saved status.
		writeData(w, 200, a)
		return
	}
	writeError(w, 404, "not_found", "Pesan tidak ditemukan.")
}
func (s *Server) checkNotificationSender(w http.ResponseWriter, r *http.Request) {
	n := s.notifications
	if n.sender == nil {
		writeError(w, 503, "not_configured", "Pengirim Firebase belum diaktifkan di server.")
		return
	}
	_, err := n.sender.Send(r.Context(), notifications.Message{Title: "Pemeriksaan koneksi", Body: "Validasi konfigurasi pengirim Nihongo Master.", Route: "/home"}, "configuration-check", true)
	if err != nil {
		writeError(w, 502, "fcm_error", err.Error())
		return
	}
	writeData(w, 200, map[string]any{"valid": true, "message": "Konfigurasi Firebase valid. Tidak ada notifikasi yang dikirim."})
}
