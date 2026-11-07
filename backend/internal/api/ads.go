package api

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// Google's public AdMob test units. They never serve real ads or earn revenue.
const (
	testBannerAdUnit = "ca-app-pub-3940256099942544/6300978111"
	testNativeAdUnit = "ca-app-pub-3940256099942544/2247696110"
)

var adUnitPattern = regexp.MustCompile(`^ca-app-pub-[0-9]{16}/[0-9]{10}$`)

// AdPlacement controls one ad slot. Every is the item spacing for native ads.
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
}

type AdsStore struct {
	mu          sync.RWMutex
	value       AdsConfig
	path, token string
}

func defaultAdsConfig() AdsConfig {
	return AdsConfig{
		Banner:     AdPlacement{Enabled: true, UnitID: testBannerAdUnit},
		NewsNative: AdPlacement{Enabled: true, UnitID: testNativeAdUnit, Every: 6},
		JobsNative: AdPlacement{Enabled: true, UnitID: testNativeAdUnit, Every: 8},
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
	return nil
}

func validateAdsConfig(config AdsConfig) error {
	for name, placement := range map[string]AdPlacement{"banner": config.Banner, "newsNative": config.NewsNative, "jobsNative": config.JobsNative} {
		if placement.UnitID != "" && !adUnitPattern.MatchString(placement.UnitID) {
			return fmt.Errorf("%s: ID unit iklan harus seperti ca-app-pub-1234567890123456/1234567890", name)
		}
		if placement.Enabled && placement.UnitID == "" {
			return fmt.Errorf("%s: isi ID unit iklan sebelum mengaktifkan", name)
		}
		if name != "banner" && (placement.Every < 3 || placement.Every > 30) {
			return fmt.Errorf("%s: jarak iklan harus 3 sampai 30 item", name)
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
	return s.ads.token != "" && subtle.ConstantTimeCompare([]byte(bearerToken(r)), []byte(s.ads.token)) == 1
}

func (s *Server) updateAdsConfig(w http.ResponseWriter, r *http.Request) {
	if !s.adsAuthorized(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Kunci admin salah atau belum diatur di server.")
		return
	}
	var input AdsConfig
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "format pengaturan iklan tidak valid")
		return
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
<p><small>ID uji Google (ca-app-pub-3940256099942544/...) tidak menghasilkan pendapatan. Ganti dengan ID unit iklan asli dari akun AdMob Anda. Build debug selalu memakai ID uji.</small></p>
<button id="save" type="submit">Simpan</button> <span id="status" role="status"></span></form>
<script>const el=id=>document.getElementById(id),P=['banner','newsNative','jobsNative'];let loaded=false;
const auth=()=>({Authorization:'Bearer '+el('token').value});
el('load').onclick=async()=>{el('status').textContent='';try{const r=await fetch('/v1/admin/ads-config',{cache:'no-store',headers:auth()});if(!r.ok){alert(r.status===401?'Kunci admin salah atau belum diatur di server.':'Gagal membuka pengaturan.');return}const c=(await r.json()).data;el('enabled').checked=c.enabled;for(const p of P){el(p+'-enabled').checked=c[p].enabled;el(p+'-unit').value=c[p].unitId||'';if(p!=='banner')el(p+'-every').value=c[p].every||6}loaded=true;el('form').hidden=false}catch(e){alert('Koneksi gagal.')}};
el('form').onsubmit=async e=>{e.preventDefault();if(!loaded)return;const d={enabled:el('enabled').checked};for(const p of P){d[p]={enabled:el(p+'-enabled').checked,unitId:el(p+'-unit').value.trim()};if(p!=='banner')d[p].every=Number(el(p+'-every').value)}el('save').disabled=true;el('status').textContent='Menyimpan…';try{const r=await fetch('/v1/admin/ads-config',{method:'PUT',headers:{'Content-Type':'application/json',...auth()},body:JSON.stringify(d)});const j=await r.json();el('status').textContent=r.ok?'Tersimpan. Perubahan berlaku saat pengguna membuka aplikasi berikutnya.':(j.error&&j.error.message||'Gagal menyimpan.')}catch(e){el('status').textContent='Koneksi gagal.'}finally{el('save').disabled=false}};</script></html>`
