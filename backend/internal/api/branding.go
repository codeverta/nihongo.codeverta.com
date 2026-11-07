package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type Branding struct {
	AppName         string `json:"appName"`
	Tagline         string `json:"tagline"`
	IconURL         string `json:"iconUrl"`
	SplashURL       string `json:"splashUrl"`
	BackgroundColor string `json:"backgroundColor"`
}

type BrandingStore struct {
	mu          sync.RWMutex
	value       Branding
	path, token string
}

func defaultBranding() Branding {
	return Branding{AppName: "Nihongo Master", Tagline: "Selangkah lebih dekat dengan Jepang", BackgroundColor: "#FAFAFA"}
}

// ConfigureBranding loads durable settings. An empty token disables admin writes.
func (s *Server) ConfigureBranding(path, token string) error {
	b := &BrandingStore{value: defaultBranding(), path: path, token: token}
	data, err := os.ReadFile(path)
	if err == nil {
		if err = json.Unmarshal(data, &b.value); err != nil {
			return err
		}
		if err = validateBranding(b.value); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	s.branding = b
	return nil
}

func validateBranding(b Branding) error {
	if strings.TrimSpace(b.AppName) == "" || len([]rune(b.AppName)) > 50 || len([]rune(b.Tagline)) > 120 {
		return fmt.Errorf("nama wajib diisi (maks 50 karakter), tagline maks 120 karakter")
	}
	if !regexp.MustCompile(`^#[0-9a-fA-F]{6}$`).MatchString(b.BackgroundColor) {
		return fmt.Errorf("warna harus #RRGGBB")
	}
	for _, raw := range []string{b.IconURL, b.SplashURL} {
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || len(raw) > 2048 {
			return fmt.Errorf("gambar harus URL HTTPS publik, kosongkan untuk default")
		}
	}
	return nil
}

func (s *Server) registerBranding(mux *http.ServeMux) {
	if s.branding == nil {
		s.branding = &BrandingStore{value: defaultBranding()}
	}
	mux.HandleFunc("GET /v1/branding", func(w http.ResponseWriter, r *http.Request) {
		s.branding.mu.RLock()
		defer s.branding.mu.RUnlock()
		w.Header().Set("Cache-Control", "no-cache")
		writeData(w, http.StatusOK, s.branding.value)
	})
	mux.HandleFunc("PUT /v1/admin/branding", s.updateBranding)
	mux.HandleFunc("GET /admin/branding", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(brandingAdminHTML))
	})
}

func (s *Server) updateBranding(w http.ResponseWriter, r *http.Request) {
	b := s.branding
	if !adminRequestAuthorized(r, b.token) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "admin token tidak valid atau belum dikonfigurasi")
		return
	}
	var input Branding
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, 400, "invalid_json", err.Error())
		return
	}
	if err := validateBranding(input); err != nil {
		writeError(w, 422, "invalid_branding", err.Error())
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	// Write a sibling file and rename before publishing the new value.
	if b.path != "" {
		if err := os.MkdirAll(filepath.Dir(b.path), 0755); err != nil {
			writeError(w, 500, "save_failed", "gagal menyimpan branding")
			return
		}
		data, _ := json.MarshalIndent(input, "", "  ")
		if err := os.WriteFile(b.path+".tmp", data, 0600); err != nil {
			writeError(w, 500, "save_failed", "gagal menyimpan branding")
			return
		}
		if err := os.Rename(b.path+".tmp", b.path); err != nil {
			writeError(w, 500, "save_failed", "gagal menyimpan branding")
			return
		}
	}
	b.value = input
	writeData(w, 200, input)
}

const brandingAdminHTML = `<!doctype html><html lang="id"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Branding • Nihongo Master</title>
<style>body{font:16px system-ui;background:#fafafa;color:#181933;max-width:640px;margin:40px auto;padding:24px}h1{color:#1d2088}label{display:block;margin-top:20px}input{box-sizing:border-box;width:100%;padding:12px;border:1px solid #ccc;border-radius:8px}button{margin-top:24px;padding:14px 24px;background:#1d2088;color:white;border:0;border-radius:8px}small{color:#666}img{max-width:120px;max-height:180px;margin:16px}#preview{border-radius:12px;margin-top:24px;text-align:center;padding:20px}</style>
<h1>Branding aplikasi</h1><p>Upload PNG/JPG ke COS atau CDN, lalu masukkan URL HTTPS publik. Kosongkan URL untuk memakai gambar default.</p>
<form id="form"><label>Admin token<input id="token" type="password" required autocomplete="off"></label><label>Nama aplikasi<input id="appName" maxlength="50" required></label><label>Tagline<input id="tagline" maxlength="120"></label><label>URL ikon<input id="iconUrl" type="url" placeholder="https://..."></label><label>URL splash<input id="splashUrl" type="url" placeholder="https://..."></label><label>Warna latar splash<input id="backgroundColor" pattern="#[0-9a-fA-F]{6}" required></label><button>Simpan branding</button></form>
<p id="status" role="status"></p><div id="preview"><img id="icon" alt="Preview ikon"><img id="splash" alt="Preview splash"><h2 id="name"></h2><p id="tag"></p></div><p><small>Perubahan muncul saat aplikasi dibuka ulang. Ikon launcher dan splash native diperbarui lewat build aplikasi berikutnya.</small></p>
<script>const keys=['appName','tagline','iconUrl','splashUrl','backgroundColor'];const el=id=>document.getElementById(id);function preview(){el('preview').style.backgroundColor=el('backgroundColor').value;el('name').textContent=el('appName').value;el('tag').textContent=el('tagline').value;for(const [img,key] of [['icon','iconUrl'],['splash','splashUrl']]){el(img).hidden=!el(key).value;if(el(key).value)el(img).src=el(key).value;}}fetch('/v1/branding').then(r=>r.json()).then(j=>{keys.forEach(k=>el(k).value=j.data[k]);preview()}).catch(()=>el('status').textContent='Gagal memuat branding');keys.forEach(k=>el(k).addEventListener('input',preview));el('form').onsubmit=async e=>{e.preventDefault();el('status').textContent='Menyimpan…';try{const r=await fetch('/v1/admin/branding',{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+el('token').value},body:JSON.stringify(Object.fromEntries(keys.map(k=>[k,el(k).value])))});const j=await r.json();el('status').textContent=r.ok?'Branding tersimpan.':j.error.message;}catch(e){el('status').textContent='Koneksi gagal. Coba lagi.'}};</script></html>`
