package api

import (
	"errors"
	"net/http"
	"nihongo-master-backend/internal/anki"
	"nihongo-master-backend/internal/store"
	"strconv"
	"strings"
)

func (s *Server) registerAnki(mux *http.ServeMux) {
	s.registerAnkiTransfer(mux)
	mux.HandleFunc("GET /v1/anki/decks", s.optionalAuth(s.ankiDecks))
	mux.HandleFunc("GET /v1/anki/credits", s.ankiCredits)
	mux.HandleFunc("GET /v1/anki/decks/{deckID}", s.optionalAuth(s.ankiDeck))
	mux.HandleFunc("GET /v1/anki/decks/{deckID}/study", s.requireAuth(s.ankiStudy))
	mux.HandleFunc("POST /v1/anki/decks/{deckID}/review", s.requireAuth(s.ankiReview))
}
func (s *Server) ankiDecks(w http.ResponseWriter, r *http.Request) {
	decks, err := anki.List(r.URL.Query().Get("q"), r.URL.Query().Get("category"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	items := make([]any, 0, len(decks))
	for _, deck := range decks {
		items = append(items, deck)
	}
	if userID(r) != "" {
		personal, err := s.store.UserAnkiDecks(userID(r))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, deck := range personal {
			category, query := r.URL.Query().Get("category"), strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
			if category != "" && category != deck.Category {
				continue
			}
			if query != "" && !strings.Contains(strings.ToLower(deck.Title+" "+deck.Description+" "+deck.Level), query) {
				continue
			}
			items = append(items, deck)
		}
	}
	w.Header().Set("Cache-Control", "private, no-store")
	writeData(w, 200, map[string]any{"decks": items, "revision": anki.Revision(), "source": "Kotoba / mistval"})
}
func (s *Server) ankiCredits(w http.ResponseWriter, r *http.Request) {
	writeData(w, 200, map[string]any{"repository": "https://github.com/mistval/kotoba", "revision": anki.Revision(), "license": anki.License(), "notice": "Data dan arti mengikuti sumber Kotoba. Sebagian deck memuat atribusi sumber pihak ketiga pada detailnya. Tautan media eksternal tetap mengikuti ketersediaan penyedianya."})
}
func (s *Server) ankiDeck(w http.ResponseWriter, r *http.Request) {
	view, err := s.store.AnkiDeck(userID(r), r.PathValue("deckID"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	writeData(w, 200, view)
}
func (s *Server) ankiStudy(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if value := r.URL.Query().Get("limit"); value != "" {
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > 20 {
			writeStoreError(w, store.ErrInvalid)
			return
		}
		limit = n
	}
	result, err := s.store.AnkiStudy(userID(r), r.PathValue("deckID"), limit)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	writeData(w, 200, result)
}
func (s *Server) ankiReview(w http.ResponseWriter, r *http.Request) {
	var input store.AnkiReviewInput
	if err := decodeJSON(r, &input); err != nil {
		writeStoreError(w, store.ErrInvalid)
		return
	}
	result, err := s.store.AnkiReview(userID(r), r.PathValue("deckID"), input)
	if errors.Is(err, store.ErrAnkiConflict) {
		writeError(w, 409, "review_conflict", err.Error())
		return
	}
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	writeData(w, 200, result)
}
