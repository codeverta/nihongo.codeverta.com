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
	// MessageEn is shown instead of Message to English users. When empty,
	// English users see the built-in English explanation only.
	MessageEn string `json:"messageEn"`
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
		if len([]rune(p.Message)) > 500 || len([]rune(p.MessageEn)) > 500 || len(p.LatestVersion) > 30 {
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
	mux.HandleFunc("GET /v1/admin/app-update-policy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		u := s.appUpdates
		if !adminRequestAuthorized(r, u.token) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Kunci admin salah atau belum diatur di server.")
			return
		}
		u.mu.RLock()
		defer u.mu.RUnlock()
		writeData(w, http.StatusOK, u.value)
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
	if !adminRequestAuthorized(r, u.token) {
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
<style>body{font:16px system-ui;background:#fafafa;color:#181933;max-width:760px;margin:32px auto;padding:24px}fieldset{border:1px solid #ddd;border-radius:12px;margin:24px 0;padding:20px}label{display:block;margin:16px 0 6px}input:not([type=checkbox]),select,textarea{box-sizing:border-box;width:100%;padding:12px;border:1px solid #bbb;border-radius:8px;font:inherit}button{padding:14px 24px;background:#1d2088;color:white;border:0;border-radius:8px;cursor:pointer}small{display:block;color:#555;margin-top:6px;line-height:1.5}summary{cursor:pointer;font-weight:650;padding:12px 0}.hint{background:#eef0ff;padding:16px;border-radius:10px;line-height:1.6}.secret{display:flex;gap:8px}.secret input{flex:1;min-width:0}.secret button{padding:10px 14px;font-size:14px}legend{font-weight:650}#status{white-space:pre-wrap}</style>
<h1>Update aplikasi</h1><p>Atur kapan pengguna harus memperbarui Nihongo Master.</p>
<p class="hint"><strong>Contoh:</strong> rilis <strong>1.0.2+3</strong> berarti nama versi <strong>1.0.2</strong> dan nomor build <strong>3</strong>. Jika build minimum diisi 3, pengguna dengan build 1 atau 2 akan diminta update. Isi <strong>0</strong> jika belum ingin mewajibkan update.<br><strong>Pastikan rilis tersedia di Play Store untuk semua pengguna terkait sebelum menaikkan batas minimum.</strong></p>
<form id="form"><fieldset><legend>Akses pengaturan</legend><label for="token">Kunci admin (password pengaturan)</label><div class="secret"><input id="token" type="password" required autocomplete="off" spellcheck="false" autocapitalize="none" aria-describedby="token-help" placeholder="Masukkan kunci admin backend"><button id="toggle-token" type="button" aria-controls="token" aria-pressed="false">Tampilkan</button></div><small id="token-help">Kunci rahasia untuk menyimpan pengaturan ini. Bukan password Google, Play Store, atau akun aplikasi. Saat ini memakai kunci admin yang sama dengan halaman branding backend. Kunci disimpan di konfigurasi server (APP_UPDATE_ADMIN_TOKEN, atau BRANDING_ADMIN_TOKEN jika belum diisi). Kunci tidak ditampilkan oleh halaman ini dan tidak disimpan di browser.</small><button id="unlock" type="button">Buka pengaturan</button><button id="lock" type="button" hidden>Kunci halaman</button></fieldset><div id="settings" hidden><div id="platforms"></div><button id="save" disabled>Simpan pengaturan</button></div></form><p id="status" role="status"></p>
<script>const el=id=>document.getElementById(id);const fields=['minimumBuild','latestBuild','latestVersion','minimumAndroidApi','minimumIOSVersion','storeUrl','unsupportedDevicePolicy','message','messageEn'];const numberFields=['minimumBuild','latestBuild','minimumAndroidApi'];let loaded=false;
el('toggle-token').onclick=()=>{const show=el('token').type==='password';el('token').type=show?'text':'password';el('toggle-token').textContent=show?'Sembunyikan':'Tampilkan';el('toggle-token').setAttribute('aria-pressed',String(show))};
window.addEventListener('blur',()=>{el('token').type='password';el('toggle-token').textContent='Tampilkan';el('toggle-token').setAttribute('aria-pressed','false')});
const labels={minimumBuild:'Nomor build minimum',latestBuild:'Nomor build terbaru di store',latestVersion:'Nama versi terbaru di store',minimumAndroidApi:'Android minimum untuk rilis yang diwajibkan (API level)',minimumIOSVersion:'iOS minimum untuk rilis yang diwajibkan',storeUrl:'Tautan aplikasi di store',unsupportedDevicePolicy:'Jika perangkat terlalu lama untuk update',message:'Pesan tambahan (opsional)',messageEn:'Pesan tambahan untuk pengguna English (opsional)'};
const help={minimumBuild:'Angka setelah + pada versi rilis. Contoh 1.0.2+3 → isi 3. Build di bawah angka ini wajib update. Isi 0 untuk mematikan update wajib; jangan naikkan sebelum rilis tersedia di store.',latestBuild:'Nomor build rilis yang SUDAH bisa diunduh dari store, bukan yang baru selesai di GitHub. Contoh 1.0.2+3 → isi 3. Harus sama atau lebih besar dari build minimum.',latestVersion:'Nama versi yang dibaca pengguna. Contoh 1.0.2 (tanpa +3). Nomor build di atas yang menentukan apakah update wajib.',minimumAndroidApi:'Syarat Android dari rilis tujuan. Isi API level, bukan versi Android: Android 8 = 26, Android 9 = 28, Android 10 = 29, Android 11 = 30, Android 12 = 31, Android 13 = 33, Android 14 = 34, Android 15 = 35, Android 16 = 36. Isi 0 jika belum ingin memeriksa OS. Sesuaikan dengan minSdk rilis tersebut.',minimumIOSVersion:'Versi iOS paling lama yang dapat memasang rilis tujuan. Contoh 15.0. Sesuaikan dengan deployment target rilis iOS. Diperiksa hanya jika update diwajibkan.',storeUrl:'android',unsupportedDevicePolicy:'Izinkan versi lama: pengguna diberi penjelasan lalu boleh melanjutkan. Hentikan akses: pengguna harus memperbarui OS atau memakai perangkat lain; pilih ini hanya jika versi lama memang tidak dapat dilayani lagi.',messageEn:'Versi English dari pesan tambahan. Jika kosong, pengguna English hanya melihat penjelasan bawaan. Maksimal 500 karakter.',message:'Teks tambahan pada dialog update. Contoh: “Perbarui aplikasi untuk mendapatkan perbaikan login.” Maksimal 500 karakter. Boleh dikosongkan; penjelasan bawaan tetap tampil.'};
for(const [p,label] of [['android','Android'],['ios','iPhone / iPad']]){const section=document.createElement('details');section.open=p==='android';const summary=document.createElement('summary');summary.textContent=label+(p==='ios'?' — nonaktif sampai App Store siap':'');section.append(summary);const f=document.createElement('fieldset');section.append(f);const enabledLabel=document.createElement('label');const enabled=document.createElement('input');enabled.type='checkbox';enabled.id=p+'-enabled';enabled.setAttribute('aria-describedby',p+'-enabled-help');enabledLabel.append(enabled,document.createTextNode(' Aktifkan pemeriksaan versi '+label));f.append(enabledLabel);const enabledHelp=document.createElement('small');enabledHelp.id=p+'-enabled-help';enabledHelp.textContent=p==='android'?'Jika dicentang, aplikasi membaca aturan Android di bawah. Update baru diwajibkan jika build minimum lebih dari 0.':'Biarkan tidak dicentang sampai aplikasi tersedia di App Store dan tautan resminya diisi. Aturan iOS tidak memengaruhi Android.';f.append(enabledHelp);
for(const k of fields){if((p==='android'&&k==='minimumIOSVersion')||(p==='ios'&&k==='minimumAndroidApi'))continue;const l=document.createElement('label');l.textContent=labels[k];l.htmlFor=p+'-'+k;f.append(l);let input;if(k==='unsupportedDevicePolicy'){input=document.createElement('select');for(const [v,t] of [['allow','Izinkan versi lama tetap digunakan (disarankan)'],['block','Hentikan akses; minta pembaruan OS / perangkat lain']]){const o=document.createElement('option');o.value=v;o.textContent=t;input.append(o)}}else{input=document.createElement(k==='message'||k==='messageEn'?'textarea':'input');if(k!=='message'&&k!=='messageEn')input.type=numberFields.includes(k)?'number':k==='storeUrl'?'url':'text';if(numberFields.includes(k)){input.min='0';input.step='1'}if(k==='message'||k==='messageEn'){input.maxLength=500;input.rows=3}if(k==='latestVersion')input.placeholder='Contoh: 1.0.2';if(k==='minimumIOSVersion')input.placeholder='Contoh: 15.0';}input.id=p+'-'+k;input.setAttribute('aria-describedby',input.id+'-help');f.append(input);const note=document.createElement('small');note.id=input.id+'-help';note.textContent=k==='storeUrl'?(p==='android'?'Tautan resmi Play Store: https://play.google.com/store/apps/details?id=com.codeverta.nihongomaster. Dipakai tombol “Update aplikasi”.':'Tautan resmi App Store, contoh https://apps.apple.com/id/app/id1234567890. Ganti dengan ID aplikasi asli. Biarkan kosong selama iOS nonaktif.'):help[k];f.append(note)}el('platforms').append(section)}
el('unlock').onclick=async()=>{if(!el('token').reportValidity())return;el('unlock').disabled=true;el('status').textContent='Membuka pengaturan…';try{const r=await fetch('/v1/admin/app-update-policy',{cache:'no-store',headers:{Authorization:'Bearer '+el('token').value}});if(!r.ok){el('status').textContent=r.status===401?'Kunci admin salah atau belum diatur di server. Ini bukan password Google.':'Gagal membuka pengaturan. Coba lagi.';return}const j=await r.json();for(const p of ['android','ios']){el(p+'-enabled').checked=j.data[p].enabled;for(const k of fields){const field=el(p+'-'+k);if(field)field.value=j.data[p][k]}}loaded=true;el('settings').hidden=false;el('save').disabled=false;el('unlock').hidden=true;el('lock').hidden=false;el('token').readOnly=true;el('status').textContent='Pengaturan terbuka. Kunci hanya dipakai selama halaman ini dibuka.'}catch(e){el('status').textContent='Koneksi gagal. Coba lagi.'}finally{el('unlock').disabled=false;el('token').type='password';el('toggle-token').textContent='Tampilkan';el('toggle-token').setAttribute('aria-pressed','false')}};
el('lock').onclick=()=>{loaded=false;el('settings').hidden=true;el('save').disabled=true;el('token').value='';el('token').readOnly=false;el('unlock').hidden=false;el('lock').hidden=true;el('status').textContent='Halaman dikunci.';el('token').focus()};
el('form').onsubmit=async e=>{e.preventDefault();if(!loaded){el('unlock').click();return;}const data={};for(const p of ['android','ios']){data[p]={enabled:el(p+'-enabled').checked};for(const k of fields){const f=el(p+'-'+k);data[p][k]=numberFields.includes(k)?Number(f?.value||0):(f?.value||'')}}el('save').disabled=true;el('status').textContent='Menyimpan…';try{const r=await fetch('/v1/admin/app-update-policy',{method:'PUT',headers:{'Content-Type':'application/json','Authorization':'Bearer '+el('token').value},body:JSON.stringify(data)});const j=await r.json();el('status').textContent=r.ok?'Pengaturan tersimpan. Aplikasi memeriksa lagi saat dibuka, kembali dari store, atau setiap lima menit.':(r.status===401?'Kunci admin salah atau belum diatur di server. Gunakan kunci admin backend, bukan password Google.':j.error.message)}catch(e){el('status').textContent='Koneksi gagal. Coba lagi.'}finally{el('token').type='password';el('toggle-token').textContent='Tampilkan';el('toggle-token').setAttribute('aria-pressed','false');el('save').disabled=false}};</script></html>`
