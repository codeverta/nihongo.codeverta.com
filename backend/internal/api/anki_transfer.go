package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"nihongo-master-backend/internal/anki"
	"nihongo-master-backend/internal/ankitransfer"
	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

const ankiUploadLimit = 20 << 20

var ankiTransferSlots = make(chan struct{}, 2)

func (s *Server) registerAnkiTransfer(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/anki/import/preview", s.requireAuth(s.previewAnkiImport))
	mux.HandleFunc("POST /v1/anki/import", s.requireAuth(s.commitAnkiImport))
	mux.HandleFunc("GET /v1/anki/decks/{deckID}/export", s.requireAuth(s.exportAnkiDeck))
	mux.HandleFunc("DELETE /v1/anki/decks/{deckID}", s.requireAuth(s.deleteAnkiDeck))
	mux.HandleFunc("GET /v1/anki/decks/{deckID}/media/{name}", s.requireAuth(s.ankiDeckMedia))
}

func reserveAnkiTransfer(w http.ResponseWriter) bool {
	select {
	case ankiTransferSlots <- struct{}{}:
		return true
	default:
		w.Header().Set("Retry-After", "5")
		writeError(w, 429, "transfer_busy", "Pemrosesan deck sedang penuh. Coba lagi beberapa detik.")
		return false
	}
}

func (s *Server) readAnkiImport(w http.ResponseWriter, r *http.Request) (ankitransfer.Document, string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, ankiUploadLimit+(1<<20))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
		writeError(w, 413, "invalid_upload", "File tidak terbaca atau melebihi batas 20 MB.")
		return ankitransfer.Document{}, "", false
	}
	defer r.MultipartForm.RemoveAll()
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 422, "file_required", "Pilih file deck APKG, CSV, atau JSON.")
		return ankitransfer.Document{}, "", false
	}
	defer file.Close()
	if header.Size <= 0 || header.Size > ankiUploadLimit {
		writeError(w, 413, "file_too_large", "Ukuran file maksimal 20 MB.")
		return ankitransfer.Document{}, "", false
	}
	raw, err := io.ReadAll(io.LimitReader(file, ankiUploadLimit+1))
	if err != nil || len(raw) > ankiUploadLimit {
		writeError(w, 413, "file_too_large", "Ukuran file maksimal 20 MB.")
		return ankitransfer.Document{}, "", false
	}
	sum := sha256.Sum256(raw)
	digest := hex.EncodeToString(sum[:])
	if expected := r.FormValue("sha256"); expected != "" && expected != digest {
		writeError(w, 409, "file_changed", "File berubah setelah pratinjau. Pilih ulang file.")
		return ankitransfer.Document{}, "", false
	}
	document, err := ankitransfer.Parse(header.Filename, raw)
	if err != nil {
		writeError(w, 422, "invalid_deck", err.Error())
		return ankitransfer.Document{}, "", false
	}
	if document.SourceID != "" {
		if original, lookupErr := anki.Get(document.SourceID); lookupErr == nil {
			document.Level = original.Level
		}
	}
	document.Level = strings.ToUpper(strings.TrimSpace(document.Level))
	if document.Level != "" && document.Level != "N1" && document.Level != "N2" && document.Level != "N3" && document.Level != "N4" && document.Level != "N5" {
		writeError(w, 422, "invalid_level", "Level deck tidak dikenal.")
		return ankitransfer.Document{}, "", false
	}
	if document.Level == "N1" || document.Level == "N2" || document.Level == "N3" {
		if s.store.Profile(userID(r)).Entitlement.Tier != domain.TierPremium {
			writeStoreError(w, store.ErrPremium)
			return ankitransfer.Document{}, "", false
		}
	}
	if title := strings.TrimSpace(r.FormValue("title")); title != "" {
		document.Title = title
	}
	if !utf8.ValidString(document.Title) || len([]rune(document.Title)) < 1 || len([]rune(document.Title)) > 140 {
		writeError(w, 422, "invalid_title", "Nama deck wajib diisi, maksimal 140 karakter.")
		return ankitransfer.Document{}, "", false
	}
	return document, digest, true
}

func (s *Server) previewAnkiImport(w http.ResponseWriter, r *http.Request) {
	if !reserveAnkiTransfer(w) {
		return
	}
	defer func() { <-ankiTransferSlots }()
	doc, digest, ok := s.readAnkiImport(w, r)
	if !ok {
		return
	}
	preview := doc.Cards
	if len(preview) > 3 {
		preview = preview[:3]
	}
	// Media bytes are intentionally not echoed in a preview response.
	writeData(w, 200, map[string]any{"title": doc.Title, "description": doc.Description, "cardCount": len(doc.Cards), "mediaCount": len(doc.Media), "preview": preview, "sha256": digest, "level": doc.Level, "notice": "Kartu diimpor sebagai deck pribadi baru. Jadwal review dan XP dari aplikasi lain tidak diimpor. Media yang dibundel ikut disimpan; media eksternal tetap berupa tautan."})
}
func (s *Server) commitAnkiImport(w http.ResponseWriter, r *http.Request) {
	if !reserveAnkiTransfer(w) {
		return
	}
	defer func() { <-ankiTransferSlots }()
	doc, _, ok := s.readAnkiImport(w, r)
	if !ok {
		return
	}
	requestID := r.FormValue("requestId")
	if id, err := uuid.Parse(requestID); err != nil || id == uuid.Nil {
		writeError(w, 422, "invalid_request_id", "ID impor tidak valid. Pilih ulang file.")
		return
	}
	if r.FormValue("sha256") == "" {
		writeError(w, 422, "preview_required", "Pratinjau file sebelum mengimpor.")
		return
	}
	deck, err := s.store.ImportAnkiDeck(userID(r), doc, requestID)
	if errors.Is(err, store.ErrAnkiQuota) {
		writeError(w, 422, "anki_quota", "Batas deck pribadi tercapai: maksimal 50 deck dan 100 MB per akun. Hapus deck yang tidak diperlukan.")
		return
	}
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	writeData(w, 201, deck)
}
func (s *Server) exportAnkiDeck(w http.ResponseWriter, r *http.Request) {
	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "" {
		format = "apkg"
	}
	if format != "apkg" && format != "csv" && format != "json" {
		writeError(w, 422, "invalid_format", "Pilih format APKG, CSV, atau JSON.")
		return
	}
	if !reserveAnkiTransfer(w) {
		return
	}
	defer func() { <-ankiTransferSlots }()
	doc, err := s.store.ExportAnkiDeck(userID(r), r.PathValue("deckID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	raw, err := ankitransfer.Export(doc, format)
	if err != nil {
		writeError(w, 422, "export_failed", err.Error())
		return
	}
	mime := "application/octet-stream"
	if format == "csv" {
		mime = "text/csv; charset=utf-8"
	}
	if format == "json" {
		mime = "application/json; charset=utf-8"
	}
	filename := "nihongo-deck-" + r.PathValue("deckID") + "." + format
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(filename))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(raw)
}
func (s *Server) deleteAnkiDeck(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteUserAnkiDeck(userID(r), r.PathValue("deckID")); err != nil {
		writeStoreError(w, err)
		return
	}
	writeData(w, 200, map[string]bool{"deleted": true})
}
func (s *Server) ankiDeckMedia(w http.ResponseWriter, r *http.Request) {
	raw, mime, err := s.store.AnkiDeckMedia(userID(r), r.PathValue("deckID"), r.PathValue("name"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !strings.HasPrefix(mime, "image/") && !strings.HasPrefix(mime, "audio/") {
		writeStoreError(w, store.ErrNotFound)
		return
	}
	if mime == "image/svg+xml" {
		writeStoreError(w, store.ErrNotFound)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(raw)
}
