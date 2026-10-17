package api

import (
	"crypto/subtle"
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

// Minimum OS describes the release users would have to install, not the OS
// requirement of an already installed, still supported legacy build.
type AppUpdatePlatform struct {
	Enabled                 bool   `json:"enabled"`
	MinimumBuild            int64  `json:"minimumBuild"`
	LatestBuild             int64  `json:"latestBuild"`
	LatestVersion           string `json:"latestVersion"`
	MinimumAndroidAPI       int    `json:"minimumAndroidApi"`
	MinimumIOSVersion       string `json:"minimumIOSVersion"`
	StoreURL                string `json:"storeUrl"`
	UnsupportedDevicePolicy string `json:"unsupportedDevicePolicy"`
	Message                 string `json:"message"`
}

type AppUpdatePolicy struct {
	Android AppUpdatePlatform `json:"android"`
	IOS     AppUpdatePlatform `json:"ios"`
}

type AppUpdateStore struct {
	mu          sync.RWMutex
	value       AppUpdatePolicy
	path, token string
}

func defaultAppUpdatePolicy() AppUpdatePolicy {
	return AppUpdatePolicy{
		Android: AppUpdatePlatform{Enabled: true, LatestBuild: 2, LatestVersion: "1.0.1", StoreURL: "https://play.google.com/store/apps/details?id=com.codeverta.nihongomaster", UnsupportedDevicePolicy: "allow"},
		IOS:     AppUpdatePlatform{MinimumIOSVersion: "15.0", UnsupportedDevicePolicy: "allow"},
	}
}

func (s *Server) ConfigureAppUpdates(path, token string) error {
	u := &AppUpdateStore{value: defaultAppUpdatePolicy(), path: path, token: token}
	data, err := os.ReadFile(path)
	if err == nil {
		if err = json.Unmarshal(data, &u.value); err != nil {
			return err
		}
		if err = validateAppUpdatePolicy(u.value); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	s.appUpdates = u
	return nil
}

func validateAppUpdatePolicy(policy AppUpdatePolicy) error {
	for name, p := range map[string]AppUpdatePlatform{"android": policy.Android, "ios": policy.IOS} {
		if p.MinimumBuild < 0 || p.LatestBuild < p.MinimumBuild || p.LatestBuild > 2100000000 || p.MinimumAndroidAPI < 0 || p.MinimumAndroidAPI > 1000 {
			return fmt.Errorf("%s: build minimum harus 0 sampai build terbaru; API Android harus 0 sampai 1000", name)
		}
		if p.UnsupportedDevicePolicy != "allow" && p.UnsupportedDevicePolicy != "block" {
			return fmt.Errorf("%s: kebijakan perangkat harus allow atau block", name)
		}
		if p.MinimumIOSVersion != "" && !regexp.MustCompile(`^[0-9]{1,3}(\.[0-9]{1,3}){0,2}$`).MatchString(p.MinimumIOSVersion) {
			return fmt.Errorf("%s: versi iOS harus seperti 15.0 atau 16.4.1", name)
		}
		if len([]rune(p.Message)) > 500 || len(p.LatestVersion) > 30 {
			return fmt.Errorf("%s: pesan maksimal 500 karakter dan versi maksimal 30 karakter", name)
		}
		if p.Enabled && (p.LatestBuild < 1 || strings.TrimSpace(p.LatestVersion) == "" || p.StoreURL == "") {
			return fmt.Errorf("%s: isi build terbaru, nama versi, dan URL store sebelum mengaktifkan", name)
		}
		if p.StoreURL == "" && !p.Enabled {
			continue
		}
		u, err := url.Parse(p.StoreURL)
		if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.Fragment != "" {
			return fmt.Errorf("%s: URL store harus HTTPS resmi", name)
		}
		if name == "android" {
			if u.Host != "play.google.com" || u.Path != "/store/apps/details" || u.Query().Get("id") != "com.codeverta.nihongomaster" || len(u.Query()) != 1 || len(u.Query()["id"]) != 1 {
				return fmt.Errorf("android: gunakan halaman Play Store com.codeverta.nihongomaster")
			}
		} else if u.Host != "apps.apple.com" || !regexp.MustCompile(`^/(?:[a-z]{2}/)?app/(?:[^/]+/)?id[0-9]+$`).MatchString(u.Path) || u.RawQuery != "" {
			return fmt.Errorf("ios: gunakan URL App Store dengan Apple ID angka aplikasi")
		}
	}
	return nil
}

func (s *Server) registerAppUpdates(mux *http.ServeMux) {
	if s.appUpdates == nil {
		s.appUpdates = &AppUpdateStore{value: defaultAppUpdatePolicy()}
	}
	mux.HandleFunc("GET /v1/app-update-policy", func(w http.ResponseWriter, r *http.Request) {
		s.appUpdates.mu.RLock()
		defer s.appUpdates.mu.RUnlock()
		w.Header().Set("Cache-Control", "no-store")
		writeData(w, http.StatusOK, s.appUpdates.value)
	})
	mux.HandleFunc("PUT /v1/admin/app-update-policy", s.updateAppUpdatePolicy)
	mux.HandleFunc("GET /admin/app-updates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(appUpdatesAdminHTML))
	})
}

func (s *Server) updateAppUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	u := s.appUpdates
	if u.token == "" || subtle.ConstantTimeCompare([]byte(bearerToken(r)), []byte(u.token)) != 1 {
		writeError(w, http.StatusUnauthorized, "unauthorized", "admin token tidak valid atau belum dikonfigurasi")
		return
	}
	var input AppUpdatePolicy
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "format kebijakan tidak valid")
		return
	}
	if err := validateAppUpdatePolicy(input); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_update_policy", err.Error())
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.path != "" {
		if err := os.MkdirAll(filepath.Dir(u.path), 0700); err != nil {
			writeError(w, 500, "save_failed", "gagal menyimpan pengaturan")
			return
		}
		data, _ := json.MarshalIndent(input, "", "  ")
		if err := os.WriteFile(u.path+".tmp", data, 0600); err != nil {
			writeError(w, 500, "save_failed", "gagal menyimpan pengaturan")
			return
		}
		if err := os.Rename(u.path+".tmp", u.path); err != nil {
			writeError(w, 500, "save_failed", "gagal menyimpan pengaturan")
			return
		}
	}
	u.value = input
	writeData(w, 200, input)
}

const appUpdatesAdminHTML = `<!doctype html><html lang="id"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Update aplikasi • Nihongo Master</title>
<style>body{font:16px system-ui;background:#fafafa;color:#181933;max-width:760px;margin:32px auto;padding:24px}fieldset{border:1px solid #ddd;border-radius:12px;margin:24px 0;padding:20px}label{display:block;margin:16px 0 6px}input:not([type=checkbox]),select,textarea{box-sizing:border-box;width:100%;padding:12px;border:1px solid #bbb;border-radius:8px;font:inherit}button{padding:14px 24px;background:#1d2088;color:white;border:0;border-radius:8px;cursor:pointer}small{display:block;color:#555;margin-top:6px}#status{white-space:pre-wrap}</style>
<h1>Update aplikasi</h1><p>Atur versi minimum yang masih didukung. Pengguna di bawah batas ini akan diminta memperbarui aplikasi.</p><p>Publikasikan rilis baru untuk semua pengguna terkait sebelum menaikkan build minimum. Nomor build adalah angka setelah tanda + pada versi rilis. Gunakan build minimum 0 untuk menonaktifkan kewajiban update.</p>
<form id="form"><label>Admin token<input id="token" type="password" required autocomplete="off"></label><div id="platforms"></div><button id="save" disabled>Simpan pengaturan</button></form><p id="status" role="status"></p>
<script>const el=id=>document.getElementById(id);const fields=['minimumBuild','latestBuild','latestVersion','minimumAndroidApi','minimumIOSVersion','storeUrl','unsupportedDevicePolicy','message'];const numberFields=['minimumBuild','latestBuild','minimumAndroidApi'];let loaded=false;
for(const [p,label] of [['android','Android'],['ios','iPhone / iPad']]){const f=document.createElement('fieldset');f.innerHTML='<legend>'+label+'</legend><label><input type="checkbox" id="'+p+'-enabled"> Aktifkan pemeriksaan versi</label>';const labels={minimumBuild:'Build minimum yang masih didukung',latestBuild:'Build terbaru yang sudah tersedia di store',latestVersion:'Nama versi terbaru',minimumAndroidApi:'API Android minimum untuk rilis yang diwajibkan',minimumIOSVersion:'Versi iOS minimum untuk rilis yang diwajibkan',storeUrl:'URL halaman aplikasi di store',unsupportedDevicePolicy:'Jika OS perangkat tidak mendukung rilis yang diwajibkan',message:'Pesan tambahan untuk pengguna'};for(const k of fields){if((p==='android'&&k==='minimumIOSVersion')||(p==='ios'&&k==='minimumAndroidApi'))continue;const l=document.createElement('label');l.textContent=labels[k];f.append(l);let input;if(k==='unsupportedDevicePolicy'){input=document.createElement('select');for(const [v,t] of [['allow','Izinkan versi lama tetap digunakan (disarankan)'],['block','Hentikan akses; minta pembaruan OS / perangkat lain']]){const o=document.createElement('option');o.value=v;o.textContent=t;input.append(o)}}else{input=document.createElement(k==='message'?'textarea':'input');if(k!=='message')input.type=numberFields.includes(k)?'number':k==='storeUrl'?'url':'text';if(numberFields.includes(k)){input.min='0';input.step='1'}if(k==='message')input.maxLength=500;}input.id=p+'-'+k;f.append(input)}const note=document.createElement('small');note.textContent=p==='android'?'Gunakan API level (contoh Android 8 = API 26), bukan nomor versi Android. Nilai 0 mengabaikan syarat OS.':'iOS tetap nonaktif sampai URL App Store resminya tersedia. Contoh versi OS: 15.0.';f.append(note);el('platforms').append(f)}
fetch('/v1/app-update-policy',{cache:'no-store'}).then(r=>{if(!r.ok)throw Error();return r.json()}).then(j=>{for(const p of ['android','ios']){el(p+'-enabled').checked=j.data[p].enabled;for(const k of fields){const field=el(p+'-'+k);if(field)field.value=j.data[p][k]}}loaded=true;el('save').disabled=false}).catch(()=>el('status').textContent='Gagal memuat pengaturan. Muat ulang halaman sebelum mengubahnya.');
el('form').onsubmit=async e=>{e.preventDefault();if(!loaded)return;const data={};for(const p of ['android','ios']){data[p]={enabled:el(p+'-enabled').checked};for(const k of fields){const f=el(p+'-'+k);data[p][k]=numberFields.includes(k)?Number(f?.value||0):(f?.value||'')}}el('save').disabled=true;el('status').textContent='Menyimpan…';try{const r=await fetch('/v1/admin/app-update-policy',{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+el('token').value},body:JSON.stringify(data)});const j=await r.json();el('status').textContent=r.ok?'Pengaturan tersimpan. Aplikasi memeriksa lagi saat dibuka, kembali dari store, atau setiap lima menit.':j.error.message}catch(e){el('status').textContent='Koneksi gagal. Coba lagi.'}finally{el('save').disabled=false}};</script></html>`
