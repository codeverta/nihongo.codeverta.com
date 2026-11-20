package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"nihongo-master-backend/internal/unlock"
)

// Google's public AdMob test units. They never serve real ads or earn revenue.
const (
	testBannerAdUnit       = "ca-app-pub-3940256099942544/6300978111"
	testNativeAdUnit       = "ca-app-pub-3940256099942544/2247696110"
	testRewardedAdUnit     = "ca-app-pub-3940256099942544/5224354917"
	testInterstitialAdUnit = "ca-app-pub-3940256099942544/1033173712"
)

// defaultInterstitialMinutes is the foreground usage between interstitials.
const defaultInterstitialMinutes = 10

var adUnitPattern = regexp.MustCompile(`^ca-app-pub-[0-9]{16}/[0-9]{10}$`)

// AdPlacement controls one ad slot. Every is the item spacing for native ads
// and the minutes of foreground use between interstitials.
type AdPlacement struct {
	Enabled bool   `json:"enabled"`
	UnitID  string `json:"unitId"`
	Every   int    `json:"every,omitempty"`
}

// AdsConfig is the remote on/off switch for in-app ads. Enabled is the master
// switch; each placement can also be turned off individually. The app never
// shows ads to Premium or Lifetime users regardless of this configuration.
type AdsConfig struct {
	Enabled    bool        `json:"enabled"`
	Banner     AdPlacement `json:"banner"`
	NewsNative AdPlacement `json:"newsNative"`
	JobsNative AdPlacement `json:"jobsNative"`
	// AIChatRewarded is the opt-in rewarded ad that unlocks one AI chat
	// message after the daily free quota.
	AIChatRewarded AdPlacement `json:"aiChatRewarded"`
	// Interstitial is shown at the next natural break (screen or tab change)
	// once the user has used the app for Every minutes since the last
	// full-screen ad. Never mid-activity.
	Interstitial AdPlacement `json:"interstitial"`
	// DictionaryNative is inserted every N dictionary search results.
	DictionaryNative AdPlacement `json:"dictionaryNative"`
	// QuizResultBanner sits on lesson quiz and exam result screens.
	QuizResultBanner AdPlacement `json:"quizResultBanner"`
	// UnlockRewarded is the opt-in rewarded ad that opens one locked item
	// (see docs/rewarded-unlock-spec.md). The unit must have server-side
	// verification pointed at /v1/ads/ssv.
	UnlockRewarded AdPlacement `json:"unlockRewarded"`
	// UnlockDailyCap is the most items an account can open per day, across
	// every kind of item. Zero means the default.
	UnlockDailyCap int `json:"unlockDailyCap"`
	// UnlockFallbackPerDay is how many of those may be opened without a
	// verified ad when no ad could be loaded. Zero turns the fallback off.
	UnlockFallbackPerDay int `json:"unlockFallbackPerDay"`
	// UnlockKinds switches the ad offer per kind of item. An item opened
	// earlier stays open when its kind is switched off later.
	UnlockKinds UnlockKinds `json:"unlockKinds"`
	// ContentGates locks content that used to be free. Both start off so a
	// deploy changes nothing; switch them on only after the app that can
	// unlock them is out (older apps cannot read a locked article or apply).
	ContentGates ContentGates `json:"contentGates"`
	// InterstitialKinds switches the full-screen ad per completion moment.
	// The ad only ever follows a finished lesson, exam, practice or Anki
	// session, when the user leaves its result screen.
	InterstitialKinds CompletionKinds `json:"interstitialKinds"`
	// ExamTabBanner sits above the bottom navigation on the Exam tab.
	ExamTabBanner AdPlacement `json:"examTabBanner"`
	// PracticeResultBanner sits on the end screen of an Anki session, an SSW
	// practice and a Kaigo lesson.
	PracticeResultBanner AdPlacement `json:"practiceResultBanner"`
	// NewsArticleNative is one native ad at the end of a news article.
	NewsArticleNative AdPlacement `json:"newsArticleNative"`
	// AnkiSessionMinCards is how many cards an Anki session needs before it
	// counts as finished. Zero means the default.
	AnkiSessionMinCards int `json:"ankiSessionMinCards"`
	// XPRepeatDailyCap is the XP a user can earn per study day from repeat
	// completions and SSW practice. Zero means the default.
	XPRepeatDailyCap int `json:"xpRepeatDailyCap"`
	// XPBoostRewarded is the opt-in rewarded ad on result screens that
	// doubles the XP of that activity. Like UnlockRewarded it must use
	// server-side verification pointed at /v1/ads/ssv.
	XPBoostRewarded AdPlacement `json:"xpBoostRewarded"`
	// XPBoostKinds says which activities offer double XP.
	XPBoostKinds CompletionKinds `json:"xpBoostKinds"`
	// XPBoostDailyCap is how many activities a user can double per day, with
	// an ad or (Lifetime) without. Zero means the default.
	XPBoostDailyCap int `json:"xpBoostDailyCap"`
}

const defaultXPBoostDailyCap = 10

func (c AdsConfig) xpBoostDailyCap() int {
	if c.XPBoostDailyCap > 0 {
		return c.XPBoostDailyCap
	}
	return defaultXPBoostDailyCap
}

// boostAllows reports whether double XP is switched on for a completion
// group (lessons, exams, practice) as the store reports it.
func (c AdsConfig) boostAllows(group string) bool {
	switch group {
	case "lessons":
		return c.XPBoostKinds.Lessons
	case "exams":
		return c.XPBoostKinds.Exams
	case "practice":
		return c.XPBoostKinds.Practice
	}
	return false
}

const defaultXPRepeatDailyCap = 100

// CompletionKinds are the "finished studying" moments.
type CompletionKinds struct {
	Lessons  bool `json:"lessons"`  // JLPT lesson quizzes
	Exams    bool `json:"exams"`    // simulation, diagnostic, JFT, NAT, SSW exams
	Practice bool `json:"practice"` // SSW practice and Kaigo lessons
	Anki     bool `json:"anki"`     // Anki sessions past the minimum card count
}

const defaultAnkiSessionMinCards = 10

// UnlockKinds says which kinds of item may be opened with a rewarded ad.
type UnlockKinds struct {
	Lessons    bool `json:"lessons"`    // N3–N1 grammar, kanji, reading, vocabulary
	Anki       bool `json:"anki"`       // built-in N3–N1 Anki decks
	Simulation bool `json:"simulation"` // JLPT N3 simulation paper 01, per attempt
	News       bool `json:"news"`       // one news article
	Jobs       bool `json:"jobs"`       // one job's application link
}

// ContentGates locks news articles and job application links for free
// accounts. Lessons, decks and simulations are locked by NIHONGO_OPEN_ACCESS.
type ContentGates struct {
	News bool `json:"news"`
	Jobs bool `json:"jobs"`
}

// unlockOffered reports whether an ad may open an item of kind.
func (c AdsConfig) unlockOffered(kind unlock.Kind) bool {
	if kind == unlock.KindXPBoost {
		return c.Enabled && c.XPBoostRewarded.Enabled && c.XPBoostRewarded.UnitID != ""
	}
	if !c.unlockAvailable() {
		return false
	}
	switch {
	case kind.IsLesson():
		return c.UnlockKinds.Lessons
	case kind == unlock.KindAnki:
		return c.UnlockKinds.Anki
	case kind == unlock.KindSimulation:
		return c.UnlockKinds.Simulation
	case kind == unlock.KindNews:
		return c.UnlockKinds.News
	case kind == unlock.KindJob:
		return c.UnlockKinds.Jobs
	}
	return false
}

const (
	defaultUnlockDailyCap       = 20
	defaultUnlockFallbackPerDay = 2
)

// unlockAvailable reports whether free users can be offered an ad to open a
// locked item.
func (c AdsConfig) unlockAvailable() bool {
	return c.Enabled && c.UnlockRewarded.Enabled && c.UnlockRewarded.UnitID != ""
}

func (c AdsConfig) unlockLimits() unlock.Limits {
	limits := unlock.Limits{DailyCap: c.UnlockDailyCap, FallbackPerDay: c.UnlockFallbackPerDay}
	if limits.DailyCap <= 0 {
		limits.DailyCap = defaultUnlockDailyCap
	}
	if limits.FallbackPerDay < 0 {
		limits.FallbackPerDay = 0
	}
	return limits
}

// aiChatRewardedAvailable reports whether free users can be offered an ad.
func (c AdsConfig) aiChatRewardedAvailable() bool {
	return c.Enabled && c.AIChatRewarded.Enabled && c.AIChatRewarded.UnitID != ""
}

type AdsStore struct {
	mu          sync.RWMutex
	value       AdsConfig
	path, token string
}

func defaultAdsConfig() AdsConfig {
	return AdsConfig{
		Banner:           AdPlacement{Enabled: true, UnitID: testBannerAdUnit},
		NewsNative:       AdPlacement{Enabled: true, UnitID: testNativeAdUnit, Every: 6},
		JobsNative:       AdPlacement{Enabled: true, UnitID: testNativeAdUnit, Every: 8},
		AIChatRewarded:   AdPlacement{Enabled: true, UnitID: testRewardedAdUnit},
		Interstitial:     AdPlacement{Enabled: true, UnitID: testInterstitialAdUnit, Every: defaultInterstitialMinutes},
		DictionaryNative: AdPlacement{Enabled: true, UnitID: testNativeAdUnit, Every: 8},
		QuizResultBanner: AdPlacement{Enabled: true, UnitID: testBannerAdUnit},
		UnlockRewarded:   AdPlacement{Enabled: true, UnitID: testRewardedAdUnit},
		UnlockDailyCap:   defaultUnlockDailyCap, UnlockFallbackPerDay: defaultUnlockFallbackPerDay,
		UnlockKinds:          UnlockKinds{Lessons: true, Anki: true, Simulation: true, News: true, Jobs: true},
		InterstitialKinds:    CompletionKinds{Lessons: true, Exams: true, Practice: true, Anki: true},
		ExamTabBanner:        AdPlacement{Enabled: true, UnitID: testBannerAdUnit},
		PracticeResultBanner: AdPlacement{Enabled: true, UnitID: testBannerAdUnit},
		NewsArticleNative:    AdPlacement{Enabled: true, UnitID: testNativeAdUnit},
		AnkiSessionMinCards:  defaultAnkiSessionMinCards,
		XPRepeatDailyCap:     defaultXPRepeatDailyCap,
		XPBoostRewarded:      AdPlacement{Enabled: true, UnitID: testRewardedAdUnit},
		XPBoostKinds:         CompletionKinds{Lessons: true, Exams: true, Practice: true},
		XPBoostDailyCap:      defaultXPBoostDailyCap,
	}
}

func (s *Server) ConfigureAds(path, token string) error {
	store := &AdsStore{value: defaultAdsConfig(), path: path, token: token}
	data, err := os.ReadFile(path)
	if err == nil {
		if err = json.Unmarshal(data, &store.value); err != nil {
			return err
		}
		if err = validateAdsConfig(store.value); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	s.ads = store
	s.applyXPRules(store.value)
	return nil
}

func validateAdsConfig(config AdsConfig) error {
	placements := map[string]AdPlacement{
		"banner": config.Banner, "newsNative": config.NewsNative, "jobsNative": config.JobsNative,
		"aiChatRewarded": config.AIChatRewarded, "interstitial": config.Interstitial,
		"dictionaryNative": config.DictionaryNative, "quizResultBanner": config.QuizResultBanner,
		"unlockRewarded": config.UnlockRewarded, "examTabBanner": config.ExamTabBanner,
		"practiceResultBanner": config.PracticeResultBanner, "newsArticleNative": config.NewsArticleNative,
		"xpBoostRewarded": config.XPBoostRewarded,
	}
	if config.XPBoostDailyCap != 0 && (config.XPBoostDailyCap < 1 || config.XPBoostDailyCap > 50) {
		return fmt.Errorf("xpBoostDailyCap: batas 2× XP harus 1 sampai 50 per hari")
	}
	if config.XPRepeatDailyCap != 0 && (config.XPRepeatDailyCap < 10 || config.XPRepeatDailyCap > 1000) {
		return fmt.Errorf("xpRepeatDailyCap: batas XP pengulangan harus 10 sampai 1000 per hari")
	}
	if config.AnkiSessionMinCards != 0 && (config.AnkiSessionMinCards < 1 || config.AnkiSessionMinCards > 100) {
		return fmt.Errorf("ankiSessionMinCards: ambang sesi Anki harus 1 sampai 100 kartu")
	}
	// Zero means "default" so an older admin client that omits these can save.
	if config.UnlockDailyCap != 0 && (config.UnlockDailyCap < 1 || config.UnlockDailyCap > 200) {
		return fmt.Errorf("unlockDailyCap: batas harian harus 1 sampai 200")
	}
	if config.UnlockFallbackPerDay < 0 || config.UnlockFallbackPerDay > 10 {
		return fmt.Errorf("unlockFallbackPerDay: jatah cadangan harus 0 sampai 10")
	}
	for name, placement := range placements {
		if placement.UnitID != "" && !adUnitPattern.MatchString(placement.UnitID) {
			return fmt.Errorf("%s: ID unit iklan harus seperti ca-app-pub-1234567890123456/1234567890", name)
		}
		if placement.Enabled && placement.UnitID == "" {
			return fmt.Errorf("%s: isi ID unit iklan sebelum mengaktifkan", name)
		}
		if (name == "newsNative" || name == "jobsNative") && (placement.Every < 3 || placement.Every > 30) {
			return fmt.Errorf("%s: jarak iklan harus 3 sampai 30 item", name)
		}
		// Newer placements only need spacing while they are switched on, so an
		// older admin client that omits them can still save.
		if name == "dictionaryNative" && placement.Enabled && (placement.Every < 3 || placement.Every > 30) {
			return fmt.Errorf("%s: jarak iklan harus 3 sampai 30 item", name)
		}
		if name == "interstitial" && placement.Enabled && (placement.Every < 5 || placement.Every > 60) {
			return fmt.Errorf("%s: jeda iklan layar penuh harus 5 sampai 60 menit", name)
		}
	}
	return nil
}

func (s *Server) registerAds(mux *http.ServeMux) {
	if s.ads == nil {
		s.ads = &AdsStore{value: defaultAdsConfig()}
	}
	mux.HandleFunc("GET /v1/ads-config", func(w http.ResponseWriter, r *http.Request) {
		s.ads.mu.RLock()
		defer s.ads.mu.RUnlock()
		w.Header().Set("Cache-Control", "no-store")
		writeData(w, http.StatusOK, s.ads.value)
	})
	mux.HandleFunc("GET /v1/admin/ads-config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !s.adsAuthorized(r) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Kunci admin salah atau belum diatur di server.")
			return
		}
		s.ads.mu.RLock()
		defer s.ads.mu.RUnlock()
		writeData(w, http.StatusOK, s.ads.value)
	})
	mux.HandleFunc("PUT /v1/admin/ads-config", s.updateAdsConfig)
	mux.HandleFunc("GET /admin/ads", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(adsAdminHTML))
	})
}

func (s *Server) adsAuthorized(r *http.Request) bool {
	return adminRequestAuthorized(r, s.ads.token)
}

func (s *Server) updateAdsConfig(w http.ResponseWriter, r *http.Request) {
	if !s.adsAuthorized(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Kunci admin salah atau belum diatur di server.")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "format pengaturan iklan tidak valid")
		return
	}
	var input AdsConfig
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "format pengaturan iklan tidak valid")
		return
	}
	// A dashboard built before the unlock settings existed sends a config
	// without them. Saving it must not switch the unlock ad off or reset its
	// limits, so keep the stored values for any key the client left out.
	var sent map[string]json.RawMessage
	_ = json.Unmarshal(body, &sent)
	s.ads.mu.RLock()
	current := s.ads.value
	s.ads.mu.RUnlock()
	if _, ok := sent["unlockRewarded"]; !ok {
		input.UnlockRewarded = current.UnlockRewarded
	}
	if _, ok := sent["unlockDailyCap"]; !ok {
		input.UnlockDailyCap = current.UnlockDailyCap
	}
	if _, ok := sent["unlockFallbackPerDay"]; !ok {
		input.UnlockFallbackPerDay = current.UnlockFallbackPerDay
	}
	if _, ok := sent["unlockKinds"]; !ok {
		input.UnlockKinds = current.UnlockKinds
	}
	if _, ok := sent["contentGates"]; !ok {
		input.ContentGates = current.ContentGates
	}
	if _, ok := sent["interstitialKinds"]; !ok {
		input.InterstitialKinds = current.InterstitialKinds
	}
	if _, ok := sent["examTabBanner"]; !ok {
		input.ExamTabBanner = current.ExamTabBanner
	}
	if _, ok := sent["practiceResultBanner"]; !ok {
		input.PracticeResultBanner = current.PracticeResultBanner
	}
	if _, ok := sent["newsArticleNative"]; !ok {
		input.NewsArticleNative = current.NewsArticleNative
	}
	if _, ok := sent["ankiSessionMinCards"]; !ok {
		input.AnkiSessionMinCards = current.AnkiSessionMinCards
	}
	if _, ok := sent["xpRepeatDailyCap"]; !ok {
		input.XPRepeatDailyCap = current.XPRepeatDailyCap
	}
	if _, ok := sent["xpBoostRewarded"]; !ok {
		input.XPBoostRewarded = current.XPBoostRewarded
	}
	if _, ok := sent["xpBoostKinds"]; !ok {
		input.XPBoostKinds = current.XPBoostKinds
	}
	if _, ok := sent["xpBoostDailyCap"]; !ok {
		input.XPBoostDailyCap = current.XPBoostDailyCap
	}
	if err := validateAdsConfig(input); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_ads_config", err.Error())
		return
	}
	s.ads.mu.Lock()
	defer s.ads.mu.Unlock()
	if s.ads.path != "" {
		if err := os.MkdirAll(filepath.Dir(s.ads.path), 0700); err != nil {
			writeError(w, 500, "save_failed", "gagal menyimpan pengaturan")
			return
		}
		data, _ := json.MarshalIndent(input, "", "  ")
		if err := os.WriteFile(s.ads.path+".tmp", data, 0600); err != nil {
			writeError(w, 500, "save_failed", "gagal menyimpan pengaturan")
			return
		}
		if err := os.Rename(s.ads.path+".tmp", s.ads.path); err != nil {
			writeError(w, 500, "save_failed", "gagal menyimpan pengaturan")
			return
		}
	}
	s.ads.value = input
	s.applyXPRules(input)
	writeData(w, http.StatusOK, input)
}

const adsAdminHTML = `<!doctype html><html lang="id"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Iklan • Nihongo Master</title>
<style>body{font:16px system-ui;background:#fafafa;color:#181933;max-width:720px;margin:32px auto;padding:24px}fieldset{border:1px solid #ddd;border-radius:12px;margin:20px 0;padding:16px 20px}label{display:block;margin:12px 0 6px}input[type=text],input[type=password],input[type=number]{width:100%;box-sizing:border-box;padding:10px;border:1px solid #bbb;border-radius:8px}button{padding:10px 18px;border:0;border-radius:8px;background:#4f46e5;color:#fff;font-size:16px;cursor:pointer}button:disabled{opacity:.5}.master{font-size:18px;font-weight:700}small{color:#555}</style>
<h1>Iklan</h1><p>Nyalakan atau matikan iklan di aplikasi tanpa rilis ulang. Pengguna Premium dan Lifetime tidak pernah melihat iklan.</p>
<label for="token">Kunci admin</label><input id="token" type="password" autocomplete="off"><p><button id="load">Buka pengaturan</button></p>
<form id="form" hidden>
<p class="master"><label><input type="checkbox" id="enabled"> Iklan aktif untuk semua pengguna gratis</label></p>
<fieldset><legend>Banner (bawah layar Beranda, Belajar, Kamus)</legend><label><input type="checkbox" id="banner-enabled"> Aktif</label><label for="banner-unit">ID unit iklan</label><input type="text" id="banner-unit"></fieldset>
<fieldset><legend>Native di daftar Berita</legend><label><input type="checkbox" id="newsNative-enabled"> Aktif</label><label for="newsNative-unit">ID unit iklan</label><input type="text" id="newsNative-unit"><label for="newsNative-every">Tampil setiap N berita (3-30)</label><input type="number" id="newsNative-every" min="3" max="30"></fieldset>
<fieldset><legend>Native di daftar Lowongan</legend><label><input type="checkbox" id="jobsNative-enabled"> Aktif</label><label for="jobsNative-unit">ID unit iklan</label><input type="text" id="jobsNative-unit"><label for="jobsNative-every">Tampil setiap N lowongan (3-30)</label><input type="number" id="jobsNative-every" min="3" max="30"></fieldset>
<fieldset><legend>Rewarded di AI Chat (setelah kuota gratis harian habis)</legend><label><input type="checkbox" id="aiChatRewarded-enabled"> Aktif</label><label for="aiChatRewarded-unit">ID unit iklan rewarded</label><input type="text" id="aiChatRewarded-unit"><small>Jika nonaktif, pengguna gratis yang kuotanya habis diarahkan ke Lifetime.</small></fieldset>
<p><small>ID uji Google (ca-app-pub-3940256099942544/...) tidak menghasilkan pendapatan. Ganti dengan ID unit iklan asli dari akun AdMob Anda. Build debug selalu memakai ID uji.</small></p>
<button id="save" type="submit">Simpan</button> <span id="status" role="status"></span></form>
<script>const el=id=>document.getElementById(id),P=['banner','newsNative','jobsNative','aiChatRewarded'],E=['newsNative','jobsNative'];let loaded=false;
const auth=()=>({Authorization:'Bearer '+el('token').value});
el('load').onclick=async()=>{el('status').textContent='';try{const r=await fetch('/v1/admin/ads-config',{cache:'no-store',headers:auth()});if(!r.ok){alert(r.status===401?'Kunci admin salah atau belum diatur di server.':'Gagal membuka pengaturan.');return}const c=(await r.json()).data;el('enabled').checked=c.enabled;for(const p of P){const v=c[p]||{};el(p+'-enabled').checked=!!v.enabled;el(p+'-unit').value=v.unitId||'';if(E.includes(p))el(p+'-every').value=v.every||6}loaded=true;el('form').hidden=false}catch(e){alert('Koneksi gagal.')}};
el('form').onsubmit=async e=>{e.preventDefault();if(!loaded)return;const d={enabled:el('enabled').checked};for(const p of P){d[p]={enabled:el(p+'-enabled').checked,unitId:el(p+'-unit').value.trim()};if(E.includes(p))d[p].every=Number(el(p+'-every').value)}el('save').disabled=true;el('status').textContent='Menyimpan…';try{const r=await fetch('/v1/admin/ads-config',{method:'PUT',headers:{'Content-Type':'application/json',...auth()},body:JSON.stringify(d)});const j=await r.json();el('status').textContent=r.ok?'Tersimpan. Perubahan berlaku saat pengguna membuka aplikasi berikutnya.':(j.error&&j.error.message||'Gagal menyimpan.')}catch(e){el('status').textContent='Koneksi gagal.'}finally{el('save').disabled=false}};</script></html>`

// applyXPRules hands the admin-tunable XP limits to the store: the daily
// repeat cap, and the Anki review count that makes a day count for the
// streak (the same threshold that ends an Anki session for ads).
func (s *Server) applyXPRules(config AdsConfig) {
	s.store.SetXPRules(config.XPRepeatDailyCap, config.AnkiSessionMinCards)
}
