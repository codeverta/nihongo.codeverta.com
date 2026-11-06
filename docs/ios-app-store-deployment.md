# Panduan Build iOS dan Upload ke App Store

Panduan ini berlaku untuk project Flutter **Nihongo Master**. GitHub Actions
membuat IPA release yang ditandatangani, menyimpan artifact 30 hari, dan
menambahkan IPA ke GitHub Release saat tag versi `v*` didorong.

## Status project saat ini

- Workflow ada di `.github/workflows/ios-release.yml` dan memakai runner
  macOS GitHub-hosted.
- Bundle ID iOS di project sudah `com.codeverta.nihongomaster` (sama dengan
  Android). Pastikan ID ini terdaftar sebagai App ID pada Apple Developer
  Account dan sama dengan variable `IOS_BUNDLE_ID` sebelum build.
- Workflow Android sudah ada terpisah. Kedua workflow menggunakan lock rilis
  yang sama supaya tidak berebut membuat GitHub Release yang sama.
- Build membutuhkan sertifikat Apple Distribution dan App Store provisioning
  profile. File IPA tidak dapat ditandatangani sebelum credential tersebut
  disiapkan.

## 1. Yang perlu disiapkan

Siapkan:

- Keanggotaan Apple Developer Program aktif.
- App Store Connect dan app record Nihongo Master.
- Bundle ID permanen, misalnya `com.perusahaan.nihongomaster`.
- Apple Distribution certificate yang memiliki private key dan diekspor sebagai
  `.p12`.
- App Store provisioning profile untuk Bundle ID dan certificate tersebut.
- URL HTTPS backend production.
- Web OAuth Client ID Google yang sama dengan client ID yang digunakan backend.

Bundle ID tidak boleh `com.example...` dan harus sama pada Xcode, Apple Developer
portal, provisioning profile, dan App Store Connect.

## 2. Tetapkan Bundle ID di project

Ubah `PRODUCT_BUNDLE_IDENTIFIER` untuk target **Runner** pada konfigurasi
Debug, Profile, dan Release di `ios/Runner.xcodeproj/project.pbxproj` menjadi
Bundle ID final, lalu commit file tersebut. Di Mac, perubahan yang sama bisa
dilakukan lewat Xcode: buka `ios/Runner.xcworkspace`, pilih project **Runner**
dan target **Runner**, lalu ganti **Bundle Identifier** di **Signing &
Capabilities**.

Pastikan hanya target app yang menjadi identitas App Store; target test dapat
memakai Bundle ID dengan akhiran `.RunnerTests`.

Workflow memeriksa bahwa `IOS_BUNDLE_ID` cocok dengan Bundle ID target Runner.
Jika masih placeholder atau nilainya berbeda, build berhenti dengan pesan yang
menunjukkan pengaturan yang perlu diperbaiki.

## 3. Buat certificate dan provisioning profile

Di [Apple Developer Certificates, Identifiers & Profiles](https://developer.apple.com/account/resources/):

1. Buat **App ID** eksplisit menggunakan Bundle ID final.
2. Buat **Apple Distribution** certificate. Pastikan private key tersedia di
   Keychain Mac yang dipakai untuk membuat certificate.
3. Ekspor certificate beserta private key dari Keychain Access sebagai file
   `.p12`, lalu tetapkan password export yang kuat.
4. Buat provisioning profile bertipe **App Store Connect** untuk App ID dan Apple Distribution certificate tadi. Unduh file `.mobileprovision`.
5. Catat **Team ID** 10 karakter dari keanggotaan Apple Developer.

Provisioning profile harus cocok tepat dengan Team ID dan Bundle ID di GitHub.
Jangan commit file `.p12`, `.mobileprovision`, private key, atau password ke
repository.

Sebelum upload pertama, buat app record di App Store Connect dengan platform iOS
dan Bundle ID yang sama. [Apple mewajibkan app record sebelum build dapat diunggah.](https://developer.apple.com/help/app-store-connect/create-an-app-record/add-a-new-app/)

## 4. Masukkan konfigurasi ke GitHub

Di repository buka **Settings > Secrets and variables > Actions**.

Tambahkan **Variables**:

| Nama | Nilai |
|---|---|
| `IOS_BUNDLE_ID` | Bundle ID final yang sama dengan project dan Apple Developer |
| `IOS_TEAM_ID` | Team ID Apple 10 karakter |
| `API_BASE_URL` | URL API production, wajib diawali `https://` |
| `GOOGLE_SERVER_CLIENT_ID` | Web OAuth Client ID berakhiran `.apps.googleusercontent.com` |

`API_BASE_URL` dan `GOOGLE_SERVER_CLIENT_ID` juga digunakan workflow Android.
Nilai Google harus sama dengan `GOOGLE_CLIENT_ID` pada backend production.
Lihat [panduan Authentication dan Google OAuth](authentication.md).

Tambahkan **Secrets**:

| Nama | Nilai |
|---|---|
| `IOS_CERTIFICATE_BASE64` | Isi file Apple Distribution `.p12` dalam Base64 |
| `IOS_CERTIFICATE_PASSWORD` | Password yang dipakai saat export `.p12` |
| `IOS_PROVISIONING_PROFILE_BASE64` | Isi file `.mobileprovision` dalam Base64 |

Ubah file certificate dan provisioning profile menjadi Base64. Jalankan satu
perintah untuk masing-masing file, lalu paste hasilnya langsung ke GitHub
Secret yang sesuai.

PowerShell — jalankan satu perintah, tempel clipboard ke GitHub Secret yang
sesuai, lalu jalankan perintah berikutnya:

```powershell
[Convert]::ToBase64String([IO.File]::ReadAllBytes('C:\secure\ios-distribution.p12')) | Set-Clipboard
```

Setelah menyimpan secret pertama, jalankan ini untuk secret profile:

```powershell
[Convert]::ToBase64String([IO.File]::ReadAllBytes('C:\secure\nihongo-master-appstore.mobileprovision')) | Set-Clipboard
```

Di macOS, jalankan satu perintah lalu paste hasil clipboard sebelum perintah
berikutnya:

```bash
base64 -i ios-distribution.p12 | pbcopy
```

```bash
base64 -i nihongo-master-appstore.mobileprovision | pbcopy
```

Di Linux dengan GNU coreutils, perintah `base64 -w 0 path-ke-file` mencetak satu
baris Base64; paste hasilnya ke secret GitHub. Base64 bukan enkripsi, jadi
jangan membagikan nilainya atau menambahkannya ke source control.

## 5. Jalankan build

### Build dan siapkan GitHub Release

1. Pastikan semua konfigurasi pada bagian sebelumnya sudah di-commit dan push.
2. Naikkan version di `pubspec.yaml`, misalnya `1.0.0+1` menjadi `1.0.1+2`.
   Build number setelah tanda `+` harus lebih tinggi untuk setiap upload baru.
3. Buat tag versi baru seperti `v1.0.1` melalui **Releases > Draft a new
   release**, atau push tag tersebut ke GitHub.
4. Buka tab **Actions**, tunggu **iOS release build** selesai.
5. Unduh `nihongo-master.ipa` dari Assets pada GitHub Release. File
   `SHA256SUMS` berisi checksum IPA.

Tag yang sama juga memulai workflow Android. Android mengunggah `.aab` dan `.apk`,
sedangkan iOS mengunggah `.ipa` ke release tersebut. Workflow iOS membuat GitHub
Release jika belum ada.

### Jalankan manual

Di **Actions > iOS release build > Run workflow**, pilih branch lalu jalankan.
File IPA tersedia pada bagian **Artifacts** di hasil run selama 30 hari. Run
manual tidak membuat atau mengubah GitHub Release.

## 6. Upload ke App Store Connect

GitHub hanya membuat dan menyimpan IPA; workflow ini tidak mengirim file ke
akun Apple secara otomatis.

1. Unduh `nihongo-master.ipa` dari GitHub Release atau artifact Actions.
2. Unggah ke App Store Connect memakai Apple Transporter atau tool upload Apple.
3. Tunggu Apple memproses build. Setelah tersedia, pilih build di versi app,
   lengkapi metadata, privacy, export compliance, dan persyaratan review.
4. Kirim versi app untuk App Review atau bagikan build lebih dahulu melalui
   TestFlight.

Flutter menghasilkan `.ipa` di `build/ios/ipa`. [Panduan Flutter iOS](https://docs.flutter.dev/deployment/ios)
menjelaskan proses build IPA dan pilihan upload ke App Store Connect.

## 7. Google Sign-In di iOS

Build IPA tidak otomatis menyiapkan OAuth khusus iOS. Setelah Bundle ID final
ditentukan, daftarkan iOS OAuth client yang memakai Bundle ID tersebut dan
konfigurasikan URL scheme/client ID iOS sesuai panduan Google Sign-In. Untuk
autentikasi backend, workflow tetap memberikan Web OAuth Client ID melalui
`GOOGLE_SERVER_CLIENT_ID`.

## Referensi

- [Flutter: Build and release an iOS app](https://docs.flutter.dev/deployment/ios)
- [GitHub: Hosted runners](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)
- [Apple: Create an App Store provisioning profile](https://developer.apple.com/help/account/provisioning-profiles/create-an-app-store-provisioning-profile/)
- [Apple: Create an app record](https://developer.apple.com/help/app-store-connect/create-an-app-record/add-a-new-app/)
- [Apple: Upload builds](https://developer.apple.com/help/app-store-connect/manage-builds/upload-builds/)
