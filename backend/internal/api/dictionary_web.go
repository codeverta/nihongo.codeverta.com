package api

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"nihongo-master-backend/internal/dictionary"
)

// Public, server-rendered dictionary pages on the landing domain. Cached words
// are served from MariaDB without limits. Lookups that would reach Jisho or
// kanjiapi.dev are capped per visitor and for the whole website, so the public
// web cannot flood the upstream sources or starve the mobile app.

const (
	publicSiteURL              = "https://nihongo.codeverta.com"
	playStoreURL               = "https://play.google.com/store/apps/details?id=com.codeverta.nihongomaster"
	webDictionaryPerIPPerHour  = 20
	webDictionaryGlobalPerHour = 300
	sitemapWordLimit           = 22000
	sitemapKanjiLimit          = 2500
)

//go:embed site/dictionary.html
var dictionaryWebHTML string

type dictionaryBadges struct {
	Word *dictionaryWordView
	C    *dictionaryCopy
}

var dictionaryWebTemplate = template.Must(template.New("dictionary").Funcs(template.FuncMap{
	"badgeData": func(word *dictionaryWordView, c *dictionaryCopy) dictionaryBadges { return dictionaryBadges{word, c} },
}).Parse(dictionaryWebHTML))

// webDictionaryLimiter counts upstream lookups in fixed one-hour windows.
// Only allowed lookups are recorded, so the map never exceeds the global cap.
type webDictionaryLimiter struct {
	mu      sync.Mutex
	window  time.Time
	global  int
	clients map[string]int
	perIP   int // zero means webDictionaryPerIPPerHour
	total   int // zero means webDictionaryGlobalPerHour
}

func (l *webDictionaryLimiter) allow(client string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	perIP, total := l.perIP, l.total
	if perIP <= 0 {
		perIP = webDictionaryPerIPPerHour
	}
	if total <= 0 {
		total = webDictionaryGlobalPerHour
	}
	if start := now.Truncate(time.Hour); !l.window.Equal(start) || l.clients == nil {
		l.window, l.global, l.clients = start, 0, map[string]int{}
	}
	if l.global >= total || l.clients[client] >= perIP {
		return false
	}
	l.global++
	l.clients[client]++
	return true
}

// webDictionaryClient identifies a visitor. CF-Connecting-IP is trusted only
// when the connection comes from the local reverse proxy (Cloudflare Tunnel and
// Nginx run on the same host); direct connections cannot spoof it. IPv6
// visitors are grouped per /64 so rotating addresses does not bypass the cap.
func webDictionaryClient(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		if forwarded := net.ParseIP(strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))); forwarded != nil {
			ip = forwarded
		}
	}
	if ip == nil {
		return "unknown"
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	return ip.Mask(net.CIDRMask(64, 128)).String() + "/64"
}

func (s *Server) webDictionaryGate(r *http.Request) func() bool {
	client := webDictionaryClient(r)
	return func() bool { return s.webDictionaryLimiter.allow(client, time.Now()) }
}

type dictionaryCopy struct {
	Lang, Base, WordBase, KanjiBase, Home                   string
	Skip, NavLabel, HomeLabel, HomeNav, Dictionary          string
	IndexTitle, IndexDescription, Heading, Lede             string
	SearchLabel, SearchPlaceholder, SearchButton            string
	ModeLabel, ModeAuto, ModeJapanese, ModeEnglish          string
	LevelLabel, AllLevels, CommonOnly, Common               string
	SearchTitle, ResultsHeading, PageSuffix, NoResults      string
	SearchJisho, More, Invalid, Limited, Unavailable        string
	WordTitle, WordDescription, WordMissing, SearchThis     string
	Meanings, OtherForms, KanjiInWord, SeeAlso, Antonyms    string
	OpenJisho, KanjiTitle, KanjiDescription, Meaning        string
	Strokes, Onyomi, Kunyomi, NameReadings, Grade, OldJLPT  string
	Frequency, WordsWithKanji, KanjiMissing, AppCTA         string
	AppButton, Privacy, Terms, Support, FooterLabel, Source string
}

var dictionaryIndonesian = dictionaryCopy{
	Lang: "id", Base: "/kamus", WordBase: "/kamus/kata/", KanjiBase: "/kamus/kanji/", Home: "/",
	Skip: "Langsung ke isi", NavLabel: "Navigasi utama", HomeLabel: "Nihongo Master, beranda", HomeNav: "Beranda", Dictionary: "Kamus",
	IndexTitle:        "Kamus bahasa Jepang online gratis · Nihongo Master",
	IndexDescription:  "Cari arti kosakata dan kanji Jepang dengan kanji, kana, romaji, atau kata Inggris. Gratis, tanpa login.",
	Heading:           "Kamus bahasa Jepang.",
	Lede:              "Cari dengan kanji, kana, romaji, atau kata Inggris. Definisi berbahasa Inggris dari Jisho.org.",
	SearchLabel:       "Kata yang dicari",
	SearchPlaceholder: "猫, neko, atau cat",
	SearchButton:      "Cari",
	ModeLabel:         "Mode", ModeAuto: "Otomatis", ModeJapanese: "Jepang (romaji)", ModeEnglish: "Inggris",
	LevelLabel: "Level", AllLevels: "Semua level", CommonOnly: "Kata umum saja", Common: "Umum",
	SearchTitle:     "“%s” · Kamus Jepang · Nihongo Master",
	ResultsHeading:  "Hasil untuk “%s”",
	PageSuffix:      " · halaman %d",
	NoResults:       "Tidak ada hasil untuk “%s”.",
	SearchJisho:     "Cari di Jisho.org",
	More:            "Muat lebih banyak",
	Invalid:         "Kata kunci tidak valid. Gunakan maksimal 100 karakter. Untuk mode Inggris, jangan pakai tanda kutip.",
	Limited:         "Kata ini belum ada di kamus kami dan batas pencarian baru sedang tercapai. Coba lagi nanti, cari di aplikasi, atau buka di Jisho.",
	Unavailable:     "Sumber kamus sedang tidak tersedia. Coba lagi sebentar atau buka di Jisho.",
	WordTitle:       "Arti %s · Kamus Jepang · Nihongo Master",
	WordDescription: "Arti %s dalam bahasa Jepang: %s. Lihat semua makna, bentuk lain, dan kanjinya.",
	WordMissing:     "Kata ini belum ada di kamus kami.",
	SearchThis:      "Cari “%s”",
	Meanings:        "Makna", OtherForms: "Bentuk lain", KanjiInWord: "Kanji dalam kata ini", SeeAlso: "Lihat juga", Antonyms: "Antonim",
	OpenJisho:        "Buka di Jisho.org",
	KanjiTitle:       "Kanji %s: arti, onyomi, kunyomi · Kamus Jepang · Nihongo Master",
	KanjiDescription: "Kanji %s: %s. Jumlah goresan, onyomi, kunyomi, dan kata yang memakainya.",
	Meaning:          "Arti", Strokes: "Jumlah goresan", Onyomi: "Onyomi", Kunyomi: "Kunyomi", NameReadings: "Bacaan nama",
	Grade: "Grade sekolah", OldJLPT: "JLPT lama", Frequency: "Peringkat frekuensi koran",
	WordsWithKanji: "Cari kata dengan kanji ini",
	KanjiMissing:   "Kanji ini tidak ditemukan.",
	AppCTA:         "Simpan kata favorit dan belajar kosakata per level di aplikasi Nihongo Master.",
	AppButton:      "Download gratis",
	Privacy:        "Privasi", Terms: "Ketentuan", Support: "Bantuan", FooterLabel: "Informasi dan kebijakan",
	Source: "Sumber",
}

var dictionaryEnglish = dictionaryCopy{
	Lang: "en", Base: "/en/dictionary", WordBase: "/en/dictionary/word/", KanjiBase: "/en/dictionary/kanji/", Home: "/en/",
	Skip: "Skip to content", NavLabel: "Main navigation", HomeLabel: "Nihongo Master, home", HomeNav: "Home", Dictionary: "Dictionary",
	IndexTitle:        "Free online Japanese dictionary · Nihongo Master",
	IndexDescription:  "Look up Japanese words and kanji by kanji, kana, romaji, or English. Free, no sign-in needed.",
	Heading:           "Japanese dictionary.",
	Lede:              "Search by kanji, kana, romaji, or English. Definitions from Jisho.org.",
	SearchLabel:       "Word to look up",
	SearchPlaceholder: "猫, neko, or cat",
	SearchButton:      "Search",
	ModeLabel:         "Mode", ModeAuto: "Automatic", ModeJapanese: "Japanese (romaji)", ModeEnglish: "English",
	LevelLabel: "Level", AllLevels: "All levels", CommonOnly: "Common words only", Common: "Common",
	SearchTitle:     "“%s” · Japanese dictionary · Nihongo Master",
	ResultsHeading:  "Results for “%s”",
	PageSuffix:      " · page %d",
	NoResults:       "No results for “%s”.",
	SearchJisho:     "Search on Jisho.org",
	More:            "Load more",
	Invalid:         "Invalid search. Use at most 100 characters. In English mode, leave out quotation marks.",
	Limited:         "This word is not in our dictionary yet and the limit for new searches has been reached. Try again later, search in the app, or open Jisho.",
	Unavailable:     "The dictionary source is unavailable right now. Try again shortly or open Jisho.",
	WordTitle:       "%s meaning · Japanese dictionary · Nihongo Master",
	WordDescription: "%s in Japanese: %s. See every meaning, other forms, and its kanji.",
	WordMissing:     "This word is not in our dictionary yet.",
	SearchThis:      "Search “%s”",
	Meanings:        "Meanings", OtherForms: "Other forms", KanjiInWord: "Kanji in this word", SeeAlso: "See also", Antonyms: "Antonyms",
	OpenJisho:        "Open on Jisho.org",
	KanjiTitle:       "Kanji %s: meaning and readings · Japanese dictionary · Nihongo Master",
	KanjiDescription: "Kanji %s: %s. Stroke count, on'yomi, kun'yomi, and words that use it.",
	Meaning:          "Meaning", Strokes: "Strokes", Onyomi: "On'yomi", Kunyomi: "Kun'yomi", NameReadings: "Name readings",
	Grade: "School grade", OldJLPT: "Old JLPT", Frequency: "Newspaper frequency rank",
	WordsWithKanji: "Search words with this kanji",
	KanjiMissing:   "This kanji was not found.",
	AppCTA:         "Save favourite words and study vocabulary by level in the Nihongo Master app.",
	AppButton:      "Download free",
	Privacy:        "Privacy", Terms: "Terms", Support: "Support", FooterLabel: "Information and policies",
	Source: "Source",
}

type dictionaryOption struct {
	Value, Label string
	Selected     bool
}

type dictionaryLink struct{ Text, URL string }

type dictionarySenseView struct {
	Parts, Definitions, Notes string
	SeeAlso, Antonyms         []dictionaryLink
}

type dictionaryWordView struct {
	Slug, Title, Reading, URL, JishoURL, Summary string
	Common                                       bool
	Levels                                       []string
	Senses                                       []dictionarySenseView
	Others                                       string
	Kanji                                        []dictionaryLink
}

type dictionaryFact struct {
	Label, Value string
	Japanese     bool
}

type dictionaryKanjiView struct {
	Character, Meanings, SearchURL, JishoURL string
	Facts                                    []dictionaryFact
}

type dictionaryPage struct {
	C                                  *dictionaryCopy
	Title, Description, Canonical      string
	NoIndex, ShowHeading               bool
	PlayURL                            string
	Query, JLPT                        string
	Common                             bool
	Modes, Levels                      []dictionaryOption
	Notice, NoticeLink, NoticeLinkText string
	ResultsHeading, NextURL            string
	Results                            []*dictionaryWordView
	Word                               *dictionaryWordView
	Kanji                              *dictionaryKanjiView
}

func (s *Server) registerDictionaryWeb(mux *http.ServeMux) {
	register := func(c *dictionaryCopy) {
		mux.HandleFunc("GET "+c.Base, func(w http.ResponseWriter, r *http.Request) { s.dictionarySearchPage(w, r, c) })
		mux.HandleFunc("GET "+c.Base+"/{$}", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, c.Base, http.StatusPermanentRedirect)
		})
		mux.HandleFunc("GET "+c.WordBase+"{slug}", func(w http.ResponseWriter, r *http.Request) { s.dictionaryWordPage(w, r, c) })
		mux.HandleFunc("GET "+c.KanjiBase+"{character}", func(w http.ResponseWriter, r *http.Request) { s.dictionaryKanjiPage(w, r, c) })
	}
	register(&dictionaryIndonesian)
	if publicEnglishContentEnabled() {
		register(&dictionaryEnglish)
	}
	mux.HandleFunc("GET /robots.txt", s.robotsTxt)
	mux.HandleFunc("GET /sitemap.xml", s.sitemapXML)
}

func newDictionaryPage(c *dictionaryCopy, query, mode, jlpt string, common bool) dictionaryPage {
	page := dictionaryPage{C: c, PlayURL: playStoreURL, Query: query, JLPT: jlpt, Common: common}
	for _, option := range []dictionaryOption{{"auto", c.ModeAuto, false}, {"japanese", c.ModeJapanese, false}, {"english", c.ModeEnglish, false}} {
		option.Selected = option.Value == mode
		page.Modes = append(page.Modes, option)
	}
	page.Levels = append(page.Levels, dictionaryOption{"", c.AllLevels, jlpt == ""})
	for _, level := range []string{"N5", "N4", "N3", "N2", "N1"} {
		page.Levels = append(page.Levels, dictionaryOption{level, level, jlpt == level})
	}
	return page
}

func (c *dictionaryCopy) searchURL(query, mode, jlpt string, common bool, page int) string {
	values := url.Values{"q": {query}}
	if mode != "" && mode != "auto" {
		values.Set("mode", mode)
	}
	if jlpt != "" {
		values.Set("jlpt", jlpt)
	}
	if common {
		values.Set("common", "1")
	}
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	return c.Base + "?" + values.Encode()
}

func jishoSearchURL(query string) string { return "https://jisho.org/search/" + url.PathEscape(query) }

func (s *Server) dictionarySearchPage(w http.ResponseWriter, r *http.Request, c *dictionaryCopy) {
	q := r.URL.Query()
	query := strings.TrimSpace(q.Get("q"))
	mode := q.Get("mode")
	if mode != "japanese" && mode != "english" {
		mode = "auto"
	}
	jlpt := strings.ToUpper(strings.TrimSpace(q.Get("jlpt")))
	if len(jlpt) != 2 || jlpt[0] != 'N' || jlpt[1] < '1' || jlpt[1] > '5' {
		jlpt = ""
	}
	common := q.Get("common") == "1"
	pageNumber, err := strconv.Atoi(q.Get("page"))
	if err != nil || pageNumber < 1 || pageNumber > 100 {
		pageNumber = 1
	}
	page := newDictionaryPage(c, query, mode, jlpt, common)
	page.ShowHeading = true
	if query == "" {
		page.Title, page.Description, page.Canonical = c.IndexTitle, c.IndexDescription, publicSiteURL+c.Base
		renderDictionaryPage(w, http.StatusOK, page)
		return
	}
	page.NoIndex = true
	page.Title = fmt.Sprintf(c.SearchTitle, query)
	page.Description = c.IndexDescription
	ctx := dictionary.WithUpstreamGate(r.Context(), s.webDictionaryGate(r))
	result, err := s.dictionary.Search(ctx, query, pageNumber, jlpt, common, mode)
	if err != nil {
		status := http.StatusServiceUnavailable
		page.Notice, page.NoticeLink, page.NoticeLinkText = c.Unavailable, jishoSearchURL(query), c.SearchJisho
		switch {
		case errors.Is(err, dictionary.ErrInvalid):
			status, page.Notice, page.NoticeLink = http.StatusBadRequest, c.Invalid, ""
		case errors.Is(err, dictionary.ErrRateLimited):
			status, page.Notice = http.StatusTooManyRequests, c.Limited
		}
		renderDictionaryPage(w, status, page)
		return
	}
	page.ResultsHeading = fmt.Sprintf(c.ResultsHeading, query)
	if pageNumber > 1 {
		page.ResultsHeading += fmt.Sprintf(c.PageSuffix, pageNumber)
	}
	for _, raw := range result.Entries {
		if view, ok := parseDictionaryWord(raw, c); ok {
			page.Results = append(page.Results, &view)
		}
	}
	if len(page.Results) == 0 {
		page.Notice = fmt.Sprintf(c.NoResults, query)
		page.NoticeLink, page.NoticeLinkText = jishoSearchURL(query), c.SearchJisho
	}
	if result.HasMore {
		page.NextURL = c.searchURL(query, mode, jlpt, common, pageNumber+1)
	}
	renderDictionaryPage(w, http.StatusOK, page)
}

func (s *Server) dictionaryWordPage(w http.ResponseWriter, r *http.Request, c *dictionaryCopy) {
	slug := r.PathValue("slug")
	page := newDictionaryPage(c, "", "auto", "", false)
	// Database only: a word page never triggers an upstream request.
	raw, err := s.dictionary.CachedWord(r.Context(), slug)
	if err == nil {
		if view, ok := parseDictionaryWord(raw, c); ok {
			page.Word = &view
			title := view.Title
			if view.Reading != "" {
				title += " (" + view.Reading + ")"
			}
			page.Title = fmt.Sprintf(c.WordTitle, title)
			page.Description = clip(fmt.Sprintf(c.WordDescription, title, view.Summary), 160)
			page.Canonical = publicSiteURL + view.URL
			renderDictionaryPage(w, http.StatusOK, page)
			return
		}
		err = dictionary.ErrNotFound
	}
	page.NoIndex = true
	page.Title = c.IndexTitle
	status := http.StatusServiceUnavailable
	page.Notice = c.Unavailable
	if errors.Is(err, dictionary.ErrNotFound) {
		status, page.Notice = http.StatusNotFound, c.WordMissing
		if utf8.ValidString(slug) && slug != "" && utf8.RuneCountInString(slug) <= 100 {
			page.NoticeLink, page.NoticeLinkText = c.searchURL(slug, "auto", "", false, 1), fmt.Sprintf(c.SearchThis, slug)
		}
	}
	renderDictionaryPage(w, status, page)
}

func (s *Server) dictionaryKanjiPage(w http.ResponseWriter, r *http.Request, c *dictionaryCopy) {
	character := r.PathValue("character")
	page := newDictionaryPage(c, "", "auto", "", false)
	ctx := dictionary.WithUpstreamGate(r.Context(), s.webDictionaryGate(r))
	raw, err := s.dictionary.Kanji(ctx, character)
	if err == nil {
		if view, ok := parseDictionaryKanji(raw, c); ok {
			page.Kanji = &view
			page.Title = fmt.Sprintf(c.KanjiTitle, view.Character)
			page.Description = clip(fmt.Sprintf(c.KanjiDescription, view.Character, view.Meanings), 160)
			page.Canonical = publicSiteURL + c.KanjiBase + url.PathEscape(view.Character)
			renderDictionaryPage(w, http.StatusOK, page)
			return
		}
		err = dictionary.ErrNotFound
	}
	page.NoIndex = true
	page.Title = c.IndexTitle
	status := http.StatusServiceUnavailable
	page.Notice = c.Unavailable
	switch {
	case errors.Is(err, dictionary.ErrInvalid), errors.Is(err, dictionary.ErrNotFound):
		status, page.Notice = http.StatusNotFound, c.KanjiMissing
	case errors.Is(err, dictionary.ErrRateLimited):
		status, page.Notice = http.StatusTooManyRequests, c.Limited
		page.NoticeLink, page.NoticeLinkText = jishoSearchURL(character+" #kanji"), c.SearchJisho
	}
	renderDictionaryPage(w, status, page)
}

func renderDictionaryPage(w http.ResponseWriter, status int, page dictionaryPage) {
	var body bytes.Buffer
	if err := dictionaryWebTemplate.ExecuteTemplate(&body, "page", page); err != nil {
		log.Printf("render dictionary page: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	h.Set("Content-Security-Policy", landingCSP)
	if status == http.StatusOK || status == http.StatusNotFound {
		h.Set("Cache-Control", "public, max-age=300, no-transform")
	} else {
		h.Set("Cache-Control", "no-store")
	}
	w.WriteHeader(status)
	_, _ = w.Write(body.Bytes())
}

type jishoWord struct {
	Slug     string   `json:"slug"`
	IsCommon bool     `json:"is_common"`
	JLPT     []string `json:"jlpt"`
	Japanese []struct {
		Word    string `json:"word"`
		Reading string `json:"reading"`
	} `json:"japanese"`
	Senses []struct {
		EnglishDefinitions []string `json:"english_definitions"`
		PartsOfSpeech      []string `json:"parts_of_speech"`
		Tags               []string `json:"tags"`
		Info               []string `json:"info"`
		SeeAlso            []string `json:"see_also"`
		Antonyms           []string `json:"antonyms"`
	} `json:"senses"`
}

func parseDictionaryWord(raw json.RawMessage, c *dictionaryCopy) (dictionaryWordView, bool) {
	var entry jishoWord
	if json.Unmarshal(raw, &entry) != nil || entry.Slug == "" || len(entry.Japanese) == 0 || len(entry.Senses) == 0 {
		return dictionaryWordView{}, false
	}
	first := entry.Japanese[0]
	view := dictionaryWordView{
		Slug:     entry.Slug,
		Title:    first.Word,
		Reading:  first.Reading,
		Common:   entry.IsCommon,
		URL:      c.WordBase + url.PathEscape(entry.Slug),
		JishoURL: "https://jisho.org/word/" + url.PathEscape(entry.Slug),
	}
	if view.Title == "" {
		view.Title, view.Reading = first.Reading, ""
	}
	if view.Title == "" {
		view.Title = entry.Slug
	}
	for _, level := range entry.JLPT {
		if l := strings.ToUpper(strings.TrimPrefix(level, "jlpt-")); len(l) == 2 && l[0] == 'N' {
			view.Levels = append(view.Levels, l)
		}
	}
	var summary []string
	for _, sense := range entry.Senses {
		if len(sense.EnglishDefinitions) == 0 {
			continue
		}
		definitions := strings.Join(sense.EnglishDefinitions, "; ")
		if len(summary) < 3 {
			summary = append(summary, definitions)
		}
		notes := append(append([]string{}, sense.Tags...), sense.Info...)
		view.Senses = append(view.Senses, dictionarySenseView{
			Parts:       strings.Join(sense.PartsOfSpeech, ", "),
			Definitions: definitions,
			Notes:       strings.Join(notes, " · "),
			SeeAlso:     dictionaryReferences(sense.SeeAlso, c),
			Antonyms:    dictionaryReferences(sense.Antonyms, c),
		})
	}
	if len(view.Senses) == 0 {
		return dictionaryWordView{}, false
	}
	view.Summary = strings.Join(summary, " / ")
	var others []string
	for _, form := range entry.Japanese[1:] {
		switch {
		case form.Word != "" && form.Reading != "":
			others = append(others, form.Word+"【"+form.Reading+"】")
		case form.Word != "":
			others = append(others, form.Word)
		case form.Reading != "":
			others = append(others, form.Reading)
		}
	}
	view.Others = strings.Join(others, "、")
	seen := map[rune]bool{}
	for _, r := range first.Word {
		if unicode.Is(unicode.Han, r) && !seen[r] {
			seen[r] = true
			view.Kanji = append(view.Kanji, dictionaryLink{Text: string(r), URL: c.KanjiBase + url.PathEscape(string(r))})
		}
	}
	return view, true
}

// Jisho references look like "食べる 1"; link to a search for the word itself.
func dictionaryReferences(items []string, c *dictionaryCopy) []dictionaryLink {
	var links []dictionaryLink
	for _, item := range items {
		item = strings.TrimSpace(item)
		word := strings.Fields(item)
		if len(word) == 0 || utf8.RuneCountInString(word[0]) > 100 {
			continue
		}
		links = append(links, dictionaryLink{Text: item, URL: c.searchURL(word[0], "auto", "", false, 1)})
	}
	return links
}

type kanjiAPIEntry struct {
	Kanji     string   `json:"kanji"`
	Grade     *int     `json:"grade"`
	Strokes   int      `json:"stroke_count"`
	Meanings  []string `json:"meanings"`
	Kun       []string `json:"kun_readings"`
	On        []string `json:"on_readings"`
	Name      []string `json:"name_readings"`
	JLPT      *int     `json:"jlpt"`
	Frequency *int     `json:"freq_mainichi_shinbun"`
}

func parseDictionaryKanji(raw json.RawMessage, c *dictionaryCopy) (dictionaryKanjiView, bool) {
	var entry kanjiAPIEntry
	if json.Unmarshal(raw, &entry) != nil || utf8.RuneCountInString(entry.Kanji) != 1 || entry.Strokes < 1 {
		return dictionaryKanjiView{}, false
	}
	view := dictionaryKanjiView{
		Character: entry.Kanji,
		Meanings:  strings.Join(entry.Meanings, ", "),
		SearchURL: c.searchURL(entry.Kanji, "auto", "", false, 1),
		JishoURL:  jishoSearchURL(entry.Kanji + " #kanji"),
	}
	add := func(label, value string, japanese bool) {
		if value != "" {
			view.Facts = append(view.Facts, dictionaryFact{label, value, japanese})
		}
	}
	add(c.Strokes, strconv.Itoa(entry.Strokes), false)
	add(c.Onyomi, strings.Join(entry.On, "、"), true)
	add(c.Kunyomi, strings.Join(entry.Kun, "、"), true)
	add(c.NameReadings, strings.Join(entry.Name, "、"), true)
	if entry.Grade != nil {
		add(c.Grade, strconv.Itoa(*entry.Grade), false)
	}
	if entry.JLPT != nil {
		add(c.OldJLPT, strconv.Itoa(*entry.JLPT), false)
	}
	if entry.Frequency != nil {
		add(c.Frequency, strconv.Itoa(*entry.Frequency), false)
	}
	return view, true
}

func clip(text string, limit int) string {
	if utf8.RuneCountInString(text) <= limit {
		return text
	}
	runes := []rune(text)
	return strings.TrimSpace(string(runes[:limit-1])) + "…"
}

func (s *Server) robotsTxt(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	// Search result URLs are noindex and may reach upstream sources, so
	// crawlers stay on word and kanji pages that are served from the database.
	_, _ = fmt.Fprintf(w, "User-agent: *\nDisallow: /v1/\nDisallow: /admin/\nDisallow: /kamus?\nDisallow: /en/dictionary?\nAllow: /\n\nSitemap: %s/sitemap.xml\n", publicSiteURL)
}

func (s *Server) sitemapXML(w http.ResponseWriter, r *http.Request) {
	words, err := s.dictionary.ArchivedIDs(r.Context(), "word", sitemapWordLimit)
	if err != nil {
		http.Error(w, "sitemap unavailable", http.StatusServiceUnavailable)
		return
	}
	kanji, err := s.dictionary.ArchivedIDs(r.Context(), "kanji", sitemapKanjiLimit)
	if err != nil {
		http.Error(w, "sitemap unavailable", http.StatusServiceUnavailable)
		return
	}
	copies := []*dictionaryCopy{&dictionaryIndonesian}
	if publicEnglishContentEnabled() {
		copies = append(copies, &dictionaryEnglish)
	}
	var body bytes.Buffer
	body.WriteString(xml.Header + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	write := func(path string) {
		body.WriteString("<url><loc>")
		_ = xml.EscapeText(&body, []byte(publicSiteURL+path))
		body.WriteString("</loc></url>\n")
	}
	for _, c := range copies {
		write(c.Home)
		write(c.Base)
		for _, slug := range words {
			if validSitemapID(slug) {
				write(c.WordBase + url.PathEscape(slug))
			}
		}
		for _, character := range kanji {
			if validSitemapID(character) {
				write(c.KanjiBase + url.PathEscape(character))
			}
		}
	}
	body.WriteString("</urlset>\n")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write(body.Bytes())
}

func validSitemapID(id string) bool {
	if id == "" || !utf8.ValidString(id) {
		return false
	}
	for _, r := range id {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
