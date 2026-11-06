# Konfigurasi Build iOS (GitHub Actions)

Dokumen ini menjelaskan arti setiap Variable dan Secret yang dibaca workflow
`.github/workflows/ios-release.yml`, dari mana nilainya didapat, dan cara
mengisinya. Langkah deploy lengkap ada di
[ios-app-store-deployment.md](ios-app-store-deployment.md).

Build iOS berjalan di runner macOS milik GitHub, jadi Xcode tidak perlu
terpasang di laptop lokal.

## Variables vs Secrets

- **Variables** bukan rahasia. Nilainya tertanam di dalam aplikasi atau tidak
  berbahaya jika terlihat. Tampil terbuka di log dan di halaman settings.
- **Secrets** rahasia. Setelah disimpan, nilainya tidak bisa dibaca lagi dan
  disamarkan di log. Jangan pernah commit nilainya ke repo.

Lokasi pengisian: repository GitHub, **Settings > Secrets and variables >
Actions**, tab **Variables** atau **Secrets**.

## Variables

| Nama | Arti | Contoh | Asal nilai |
|---|---|---|---|
| `API_BASE_URL` | Alamat API backend production. Dipasang ke aplikasi lewat `--dart-define`. Wajib diawali `https://`. | `https://nihongo.codeverta.com/v1` | URL backend production yang sudah live. Dipakai juga oleh workflow Android. |
| `GOOGLE_SERVER_CLIENT_ID` | Web OAuth Client ID Google. Aplikasi memakainya untuk meminta ID token Google Sign-In yang audience-nya adalah backend. Wajib berakhiran `.apps.googleusercontent.com`. | `1234567890-abc.apps.googleusercontent.com` | Google Cloud Console > APIs & Services > Credentials, client bertipe **Web application**. Harus sama persis dengan `GOOGLE_CLIENT_ID` di `backend/.env` production. |
| `OFFLINE_ACCESS_PUBLIC_KEY` | Public key Ed25519 (Base64, tepat 32 byte) untuk memverifikasi lease akses offline premium. Ini bagian publik, aman disimpan sebagai variable. Jika kosong atau salah, fitur offline premium ditolak. | `q3Fh...=` (44 karakter Base64) | Jalankan `go run ./cmd/offline-keys` di folder `backend`. Salin nilai `OFFLINE_ACCESS_PUBLIC_KEY=` yang dicetak. Harus pasangan dari private key di server. |
| `IOS_BUNDLE_ID` | ID unik aplikasi iOS, seperti `applicationId` di Android. Harus sama di Xcode project, Apple Developer, provisioning profile, dan App Store Connect. Tidak boleh `com.example...`. | `com.codeverta.nihongomaster` | Ditentukan sendiri, lalu didaftarkan sebagai App ID di developer.apple.com. Lihat catatan penting di bawah. |
| `IOS_TEAM_ID` | ID tim Apple Developer, 10 karakter huruf besar dan angka. | `A1B2C3D4E5` | developer.apple.com > Account > Membership details > Team ID. |

## Secrets

Ketiganya adalah paket tanda tangan (code signing). Apple hanya menerima IPA
yang ditandatangani, dan workflow butuh ketiga file ini untuk melakukannya di
runner yang masih kosong.

| Nama | Arti | Asal nilai |
|---|---|---|
| `IOS_CERTIFICATE_BASE64` | File sertifikat **Apple Distribution** beserta private key-nya (`.p12`), diubah ke Base64. Ini bukti bahwa akun Anda yang menandatangani aplikasi. | Buat sertifikat di developer.apple.com > Certificates, lalu ekspor dari Keychain Access (Mac) sebagai `.p12`. Jika belum punya Mac sama sekali, buat CSR dengan OpenSSL lalu bangun `.p12` sendiri. |
| `IOS_CERTIFICATE_PASSWORD` | Password yang Anda tentukan saat mengekspor `.p12`. | Dibuat sendiri saat ekspor. Pakai password acak yang kuat. |
| `IOS_PROVISIONING_PROFILE_BASE64` | File provisioning profile `.mobileprovision` bertipe **App Store Connect**, diubah ke Base64. Mengikat Bundle ID + Team ID + sertifikat di atas. | developer.apple.com > Profiles > tambah baru, tipe App Store Connect, pilih App ID dan sertifikat Distribution tadi, lalu unduh. |

Workflow memeriksa bahwa Team ID dan Bundle ID di dalam profile cocok dengan
`IOS_TEAM_ID` dan `IOS_BUNDLE_ID`. Jika tidak cocok, build berhenti.

Mengubah file ke Base64 (jalankan lalu tempel hasil clipboard ke GitHub):

```bash
base64 -i ios-distribution.p12 | pbcopy
base64 -i nihongo-master-appstore.mobileprovision | pbcopy
```

Linux: `base64 -w 0 nama-file`. Windows PowerShell:
`[Convert]::ToBase64String([IO.File]::ReadAllBytes('file')) | Set-Clipboard`.

## Catatan: Bundle ID project

Target **Runner** di `ios/Runner.xcodeproj/project.pbxproj` memakai
`com.codeverta.nihongomaster` (sama dengan `applicationId` Android), dan target
test memakai `com.codeverta.nihongomaster.RunnerTests`. Workflow menolak
`com.example...` dan memeriksa bahwa `IOS_BUNDLE_ID` sama persis dengan Bundle ID
target Runner. Jika Bundle ID diganti lagi, ubah `PRODUCT_BUNDLE_IDENTIFIER`
di file tersebut dan variable `IOS_BUNDLE_ID` bersamaan.

## Template `.env` untuk persiapan lokal

GitHub Actions tidak membaca file `.env`. Template ini hanya catatan kerja
lokal agar semua nilai terkumpul sebelum disalin ke GitHub. Simpan sebagai
`ios-release.env` di luar repo, atau di dalam repo karena `.env.*` sudah
masuk `.gitignore`, namun sebaiknya tetap di luar repo.

```dotenv
# ---------- GitHub Actions > Variables ----------
API_BASE_URL=https://nihongo.codeverta.com/v1
GOOGLE_SERVER_CLIENT_ID=YOUR_WEB_CLIENT_ID.apps.googleusercontent.com
OFFLINE_ACCESS_PUBLIC_KEY=BASE64_32_BYTE_ED25519_PUBLIC_KEY
IOS_BUNDLE_ID=com.codeverta.nihongomaster
IOS_TEAM_ID=ABCDE12345

# ---------- GitHub Actions > Secrets (jangan commit) ----------
# Isi dengan hasil base64 dari file; lebih aman ditempel langsung ke GitHub.
IOS_CERTIFICATE_BASE64=
IOS_CERTIFICATE_PASSWORD=
IOS_PROVISIONING_PROFILE_BASE64=
```

Nilai yang sama bisa dimasukkan lewat GitHub CLI:

```bash
gh variable set API_BASE_URL --body "https://nihongo.codeverta.com/v1"
gh variable set IOS_BUNDLE_ID --body "com.codeverta.nihongomaster"
base64 -i ios-distribution.p12 | gh secret set IOS_CERTIFICATE_BASE64
gh secret set IOS_CERTIFICATE_PASSWORD
```

## Hubungan dengan `backend/.env`

| Di GitHub (Flutter) | Di `backend/.env` (server) | Hubungan |
|---|---|---|
| `GOOGLE_SERVER_CLIENT_ID` | `GOOGLE_CLIENT_ID` | Nilai harus identik (Web Client ID). |
| `OFFLINE_ACCESS_PUBLIC_KEY` | `OFFLINE_ACCESS_PRIVATE_KEY` / `_FILE` | Pasangan public dan private key. Ganti satu berarti ganti keduanya. |
| `API_BASE_URL` | (server tidak memakainya) | Harus mengarah ke server yang menjalankan backend tersebut. |

Variable `IOS_*` dan semua Secret hanya relevan untuk build iOS, tidak ada di
backend.

## Checklist sebelum menjalankan workflow

- [ ] Bundle ID di `project.pbxproj` sudah final dan bukan `com.example...`.
- [ ] App ID dan app record dengan Bundle ID yang sama sudah dibuat di Apple.
- [ ] Lima Variables terisi.
- [ ] Tiga Secrets terisi.
- [ ] Team ID dan Bundle ID di provisioning profile cocok dengan Variables.
- [ ] Version di `pubspec.yaml` dinaikkan (build number setelah `+` lebih besar).
- [ ] Jalankan **Actions > iOS release build > Run workflow**.
