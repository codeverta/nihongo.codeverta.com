package api

import (
	"errors"
	"net/http"
	"nihongo-master-backend/internal/dictionary"
	"strconv"
)

// ConfigureDictionaryCache must run before Handler starts serving requests.
func (s *Server) ConfigureDictionaryCache(cache *dictionary.SQLCache) {
	s.dictionary.Persistence = cache
}

func (s *Server) dictionarySearch(w http.ResponseWriter, r *http.Request) {
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		var err error
		page, err = strconv.Atoi(raw)
		if err != nil {
			dictionaryError(w, dictionary.ErrInvalid)
			return
		}
	}
	common := r.URL.Query().Get("common")
	if common != "" && common != "true" && common != "false" {
		dictionaryError(w, dictionary.ErrInvalid)
		return
	}
	result, err := s.dictionary.Search(r.Context(), r.URL.Query().Get("q"), page, r.URL.Query().Get("jlpt"), common == "true")
	if err != nil {
		dictionaryError(w, err)
		return
	}
	writeData(w, 200, result)
}
func (s *Server) dictionaryKanji(w http.ResponseWriter, r *http.Request) {
	result, err := s.dictionary.Kanji(r.Context(), r.PathValue("character"))
	if err != nil {
		dictionaryError(w, err)
		return
	}
	writeData(w, 200, result)
}
func dictionaryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dictionary.ErrInvalid):
		writeError(w, 400, "invalid_query", "Masukkan kata maksimal 100 karakter, halaman 1–100, atau satu karakter kanji.")
	case errors.Is(err, dictionary.ErrNotFound):
		writeError(w, 404, "kanji_not_found", "Detail kanji belum tersedia dari sumber kamus.")
	default:
		writeError(w, 503, "dictionary_unavailable", "Sumber kamus sedang tidak tersedia. Coba lagi sebentar atau buka Jisho.")
	}
}
