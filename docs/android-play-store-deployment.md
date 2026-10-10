# Panduan Deploy Android ke Google Play Store

Panduan ini berlaku untuk project Flutter **Nihongo Master** dan diperbarui
pada **10 Oktober 2026**. Menu Play Console dapat berubah; gunakan
tautan dokumentasi resmi pada bagian akhir jika nama menu berbeda.

## 1. Status project saat ini

Jangan mengunggah build sebelum signing key dan API production dikonfigurasi.
Status project:

- `applicationId` sudah `com.codeverta.nihongomaster` dan nama aplikasi sudah `Nihongo Master`.
- Release Gradle memakai upload key jika dikonfigurasi; fallback debug hanya
  untuk preview lokal. Workflow GitHub menolak build tanpa upload key.
- Default API Flutter masih `http://10.0.2.2:8080/v1`, yang hanya berlaku pada
  emulator Android.
- Account, pembelian, progress, dan achievement sekarang persistent di SQLite;
  sesi login tetap perlu login ulang setelah restart API.
- Hapus akun tersedia di Profil; sediakan URL permintaan penghapusan untuk listing.
- Lifetime sudah terhubung ke billing client dan server verification; produk console,
  kredensial, notifications, serta tes license tester perlu dikonfigurasi sesuai
  [panduan lifetime billing](lifetime-billing.md).

Upload key dan API HTTPS production belum tersedia di workspace. Konfigurasi
store, billing, policy, dan pengujian native juga perlu diselesaikan sebelum
production.

## 2. Prasyarat

Siapkan:

- Akun [Google Play Console](https://play.google.com/console/), identitas
  developer sudah diverifikasi, dan 2-Step Verification aktif.
- Java 17 atau JBR Android Studio.
- Flutter SDK dan Android SDK terbaru yang digunakan project.
- Domain HTTPS untuk API production, privacy policy, support, dan halaman
  penghapusan akun.
- Email support yang aktif.
- Minimal dua perangkat/emulator Android untuk smoke test.

Periksa toolchain:

```powershell
flutter doctor -v
flutter --version
java -version
```

## 3. Tentukan package ID final

Package ID bersifat unik dan praktis permanen setelah aplikasi dibuat di Play
Console. Saat ini project memakai:

```text
com.codeverta.nihongomaster
```

Pilih ID final milik developer/organisasi, misalnya pola:

```text
com.namaorganisasi.nihongomaster
```

Kemudian ubah `applicationId` dan `namespace` di
`android/app/build.gradle.kts`, serta sesuaikan package folder dan deklarasi
package `MainActivity.kt`.

> Jangan mengganti package ID setelah app pertama dibuat di Play Console.
> Package ID tidak dapat digunakan ulang secara bebas. Jangan pula mengganti
> package tanpa memperbarui konfigurasi Google OAuth.

Setelah package ID final dipilih:

1. Buka Google Cloud Console > Google Auth Platform > Clients.
2. Buat Android OAuth client baru dengan package ID final.
3. Gunakan SHA-1 debug untuk development.
4. Setelah Play App Signing aktif, buat Android OAuth client lain menggunakan
   package ID yang sama dan **SHA-1 App signing key certificate** dari Play
   Console.
5. Web OAuth Client ID tetap dipakai sebagai `GOOGLE_CLIENT_ID` backend dan
   `GOOGLE_SERVER_CLIENT_ID` Flutter.

Panduan OAuth lengkap tersedia di [authentication.md](./authentication.md).

## 4. Ubah nama aplikasi

Pada `android/app/src/main/AndroidManifest.xml`, ubah:

```xml
<application
    android:label="Nihongo Master"
    ...>
```

Pastikan launcher icon final sudah tersedia untuk seluruh density Android dan
berbeda dari Play Store icon 512x512 yang diunggah melalui Play Console.

## 5. Siapkan backend production

Build Play Store tidak dapat memakai `localhost` atau `10.0.2.2`. Sebelum
membuat AAB:

1. Deploy backend Go ke server/container production.
2. Gunakan database persisten seperti PostgreSQL untuk akun, progress,
   entitlement, achievement, dan session.
3. Pasang HTTPS dengan sertifikat valid.
4. Simpan `GOOGLE_CLIENT_SECRET` dan credential database sebagai secret milik
   platform deployment, bukan di image, repository, Flutter, atau AAB.
5. Batasi CORS untuk web client jika backend juga digunakan dari web.
6. Tambahkan monitoring, backup database, migration, rate limit, dan log tanpa
   token/password.
7. Verifikasi endpoint berikut dari internet:

```powershell
Invoke-RestMethod https://api.domain-anda.com/healthz
```

Contoh URL API yang diberikan ke Flutter:

```text
https://api.domain-anda.com/v1
```

## 6. Buat upload key

Gunakan **Play App Signing**. Google menyimpan app signing key, sedangkan tim
menyimpan upload key untuk menandatangani AAB yang diunggah.

Jalankan di PowerShell. Simpan file di lokasi aman di luar repository:

```powershell
keytool -genkeypair -v `
  -keystore C:\secure\nihongo-master-upload.jks `
  -keyalg RSA `
  -keysize 2048 `
  -validity 10000 `
  -alias upload
```

Gunakan password kuat dan simpan di password manager organisasi. Backup
keystore secara terenkripsi. Jangan mengirim keystore atau password melalui
chat/email dan jangan commit ke Git.

Lihat fingerprint upload key:

```powershell
keytool -list -v `
  -keystore C:\secure\nihongo-master-upload.jks `
  -alias upload
```

## 7. Konfigurasi signing Flutter

Konfigurasi release Gradle sudah membaca upload key dari environment atau file
lokal `android/key.properties`. Buat file lokal seperti ini:

```properties
storePassword=PASSWORD_KEYSTORE
keyPassword=PASSWORD_KEY
keyAlias=upload
storeFile=C:/secure/nihongo-master-upload.jks
```

File `key.properties` dan semua file `.jks` sudah diabaikan oleh `.gitignore`.
Release lokal tanpa signing key memakai debug key hanya untuk preview; jangan
unggah hasil itu ke Play Console.

Untuk build GitHub, jangan commit keystore atau password. Di repository buka
**Settings > Secrets and variables > Actions**, lalu tambahkan Actions secrets:

- `ANDROID_KEYSTORE_BASE64`: keystore `.jks` yang diubah ke Base64.
- `ANDROID_KEYSTORE_PASSWORD`
- `ANDROID_KEY_ALIAS`
- `ANDROID_KEY_PASSWORD`

Tambahkan Actions variables berikut:

- `API_BASE_URL`: URL backend production yang diawali `https://`.
- `GOOGLE_SERVER_CLIENT_ID`: Web OAuth Client ID berakhiran
  `.apps.googleusercontent.com`.

Workflow menolak build bila salah satu variable/secret ini belum diisi. Client
ID tersebut harus sama dengan `GOOGLE_CLIENT_ID` di backend production. Untuk
mengubah keystore ke Base64 di Windows PowerShell tanpa menampilkan nilainya ke
terminal:

```powershell
[Convert]::ToBase64String([IO.File]::ReadAllBytes('C:\secure\nihongo-master-upload.jks')) | Set-Clipboard
```

Paste hasil clipboard langsung ke secret `ANDROID_KEYSTORE_BASE64` di GitHub.

## 8. Versioning

Versi Flutter berada di `pubspec.yaml`:

```yaml
version: 1.0.0+1
```

- `1.0.0` menjadi `versionName` yang terlihat pengguna.
- `1` menjadi `versionCode` internal Play Store.
- Setiap AAB update wajib memiliki `versionCode` lebih tinggi.

Contoh update berikutnya:

```yaml
version: 1.0.1+2
```

Version dapat dioverride saat build, tetapi satu sumber di `pubspec.yaml` lebih
mudah diaudit.

## 9. Target API level

Project mengikuti `flutter.targetSdkVersion`. Periksa nilai efektif dengan:

```powershell
cd android
$env:JAVA_HOME = 'C:\Program Files\Android\Android Studio\jbr'
.\gradlew.bat app:properties
cd ..
```

Mulai **31 Agustus 2026**, aplikasi dan update baru untuk Android mobile harus
menargetkan Android 16/API 36 atau lebih tinggi. Karena persyaratan berubah
setiap tahun, periksa kembali halaman target API sebelum release.

## 10. Build dan validasi AAB

Dari root project:

```powershell
flutter clean
flutter pub get
flutter analyze
flutter test
flutter build appbundle --release `
  --dart-define=API_BASE_URL=https://api.domain-anda.com/v1 `
  --dart-define=GOOGLE_SERVER_CLIENT_ID=WEB_OAUTH_CLIENT_ID
```

Hasil build:

```text
build/app/outputs/bundle/release/app-release.aab
```

`GOOGLE_SERVER_CLIENT_ID` adalah identifier publik. Jangan pernah memberikan
`GOOGLE_CLIENT_SECRET` ke `--dart-define`; nilai compile-time Flutter dapat
diekstrak dari aplikasi.

Verifikasi signing bundle:

```powershell
jarsigner -verify -verbose -certs `
  build\app\outputs\bundle\release\app-release.aab
```

Sebelum upload, pastikan:

- Tidak ada URL `10.0.2.2`, localhost, atau HTTP pada build production.
- Login email, Google Sign-In, logout, guest mode, dan refresh Beranda bekerja.
- Backend restart tidak menghapus user atau entitlement.
- Tidak ada credential pada log aplikasi/backend.
- Deep link, icon, splash screen, dark mode, offline/error state, dan rotasi
  layar diuji.
- Crash-free pada build release, bukan hanya debug.

## Build dan download dari GitHub

Workflow ada di `.github/workflows/android-release.yml`. Workflow membuat AAB
untuk Play Console dan APK untuk uji install, lalu memuat keduanya ke **GitHub
Release** saat tag versi `v*` dibuat dan didorong. Jalankan manual dari tab
**Actions** untuk mendapat workflow artifact selama 30 hari.

Setelah secrets dan variables pada bagian 7 tersimpan:

1. Commit dan push konfigurasi build ke GitHub.
2. Buat tag versi baru seperti `v1.0.0` melalui **Releases > Draft a new
   release** atau push tag ke repository. Jangan pakai ulang tag yang sudah
   pernah dirilis.
3. Buka **Actions** dan tunggu **Android release build** selesai.
4. Unduh `nihongo-master.aab` dari bagian Assets di GitHub Release, lalu upload
   AAB tersebut ke Play Console. `nihongo-master.apk` hanya untuk install/test.

Setiap update wajib menaikkan `versionCode` pada `pubspec.yaml` sebelum tag
baru dibuat. Keystore tidak dimasukkan ke artifact atau repository.

## 11. Buat aplikasi di Play Console

1. Buka Play Console > **Home > Create app**.
2. Pilih bahasa default dan nama **Nihongo Master**.
3. Pilih **App**, lalu tentukan gratis/berbayar.
4. Masukkan email support.
5. Setujui deklarasi kebijakan, export laws, dan Play App Signing.
6. Pastikan package pada AAB sama dengan package ID final yang dipilih.

Gunakan Android App Bundle (`.aab`), bukan APK, untuk release Play Store.

## 12. Store listing

Siapkan minimal:

- App name: maksimal 30 karakter.
- Short description: maksimal 80 karakter.
- Full description: jelaskan fungsi yang benar-benar sudah tersedia.
- Play Store icon: PNG 32-bit, 512x512, maksimal 1 MB.
- Feature graphic: JPEG atau PNG 24-bit tanpa alpha, 1024x500.
- Minimal dua screenshot ponsel yang benar-benar berasal dari versi aplikasi
  yang diunggah; Play Console mendukung sampai delapan per device type.
- Email support, website, dan URL privacy policy.

Jangan menulis klaim seperti “nomor satu”, “terbaik”, atau fitur yang belum
tersedia. Jika premium/billing belum aktif, jangan menampilkan seolah pengguna
sudah dapat membelinya.

## 13. App content dan compliance

Selesaikan seluruh menu **Policy and programs > App content**:

### Privacy policy

Privacy policy wajib tersedia melalui URL HTTPS publik dan juga dapat diakses
dari dalam aplikasi. Jelaskan minimal:

- Data akun: nama, email, foto profil Google jika digunakan.
- Data belajar: progress, exam, XP, streak, achievement.
- Tujuan pemrosesan, masa retensi, penghapusan, keamanan, dan kontak.
- Penggunaan Google Sign-In serta pihak ketiga lain.
- Nama aplikasi atau badan developer sesuai store listing.

### Data safety

Isi berdasarkan perilaku nyata app, backend, dan semua SDK. Untuk kondisi
project saat ini, audit setidaknya kategori **Personal info** dan **App
activity**. Jangan menyalin jawaban generik; developer bertanggung jawab atas
akurasi deklarasi.

### Account deletion

Karena login Google dapat membuat akun di backend, sebelum production harus
tersedia:

1. Jalur penghapusan akun yang mudah ditemukan di dalam aplikasi, misalnya
   Profil > Pengaturan > Hapus akun.
2. Endpoint backend yang menghapus akun dan data terkait.
3. URL web publik agar mantan pengguna dapat meminta penghapusan tanpa harus
   menginstal ulang aplikasi.
4. Penjelasan data yang mungkin dipertahankan karena kewajiban hukum.

Logout atau menonaktifkan akun tidak dihitung sebagai penghapusan.

### App access

Pilih bahwa sebagian fitur dibatasi login dan berikan akun review yang aktif:

```text
Email: demo@nihongomaster.app
Password: password khusus review yang stabil
Langkah: buka Profil/Belajar/Ujian untuk mengakses fitur login.
```

Jangan menggunakan password production pribadi. Akun review harus tersedia
selama seluruh proses review dan tidak memerlukan OTP/manual approval.

### Deklarasi lain

- Isi content rating IARC secara akurat.
- Tentukan target audience. Jangan menargetkan anak jika produk dan compliance
  belum memenuhi Families Policy.
- Deklarasikan apakah aplikasi berisi iklan.
- Tinjau permission. Manifest saat ini hanya meminta `INTERNET`; pertahankan
  prinsip permission minimum.
- Isi kategori aplikasi, contact details, dan deklarasi penggunaan government,
  health, financial, news, atau fitur khusus lain bila nantinya relevan.

## 14. Premium, subscription, dan lifetime access

Premium memberi akses ke konten digital, sehingga transaksi Play-distributed
app pada umumnya harus memakai Google Play Billing, kecuali program/pengecualian
regional yang berlaku dipenuhi.

Sebelum mengaktifkan tombol pembelian:

1. Buat one-time product `nihongo_master_lifetime`, opsi Buy, harga Indonesia Rp99.000.
2. Integrasi Flutter dan verifikasi backend sudah tersedia; isi kredensial server dan aktifkan RTDN sesuai [panduan lifetime billing](lifetime-billing.md).
3. Uji cancel, pending, refund, restore, acknowledgment, pindah akun, dan account deletion memakai license tester serta testing track.
4. Lifetime tidak mempunyai renewal atau expiry subscription; paket bulanan/tahunan sudah dihapus.

Jangan mempercayai status premium yang hanya disimpan di perangkat.

## 15. Upload pertama ke Internal testing

1. Play Console > **Test and release > Testing > Internal testing**.
2. Buat release baru.
3. Aktifkan Play App Signing dengan opsi Google-generated app signing key yang
   direkomendasikan, kecuali ada kebutuhan multi-store khusus.
4. Upload `app-release.aab`.
5. Tambahkan release notes Bahasa Indonesia.
6. Tambahkan tester melalui email list/Google Group.
7. Review dan rollout internal release.
8. Install hanya melalui opt-in link Play Store untuk menguji artifact yang
   benar-benar ditandatangani Google.

Setelah upload pertama:

1. Buka **Test and release > Setup > App integrity** atau menu App signing.
2. Salin SHA-1 pada **App signing key certificate**, bukan hanya upload key.
3. Daftarkan package ID + SHA-1 tersebut sebagai Android OAuth client di
   Google Cloud project yang sama.
4. Tunggu propagasi konfigurasi, uninstall build lama, lalu install ulang dari
   Internal testing.
5. Uji Google Sign-In lagi. Debug SHA-1 tidak cukup untuk build dari Play.

## 16. Closed testing dan production

Untuk akun developer personal yang dibuat setelah 13 November 2023, Google
saat ini mensyaratkan closed test dengan minimal **12 tester** yang tetap
opted-in selama **14 hari berturut-turut** sebelum dapat mengajukan akses
production.

Alur yang disarankan:

1. Internal testing untuk tim developer.
2. Closed testing untuk pengguna nyata dan compliance testing.
3. Pantau Android vitals, crash, ANR, feedback, dan pre-launch report.
4. Perbaiki masalah, naikkan versionCode, lalu upload AAB baru.
5. Ajukan production access jika akun terkena syarat 12/14.
6. Buat Production release dan gunakan staged rollout, misalnya 10% lalu
   meningkat bertahap setelah vitals sehat.

Jangan mengandalkan internal testing sebagai satu-satunya validasi. Internal
track juga tidak menggantikan persyaratan closed testing bagi akun yang terkena
aturan tersebut.

## 17. Checklist final

### Build

- [ ] Package ID final dan bukan `com.example`.
- [ ] App label, icon, splash, dan versi final.
- [ ] Release menggunakan upload key, bukan debug key.
- [ ] targetSdk memenuhi requirement terbaru.
- [ ] `flutter analyze` dan `flutter test` lulus.
- [ ] AAB berhasil diverifikasi dan diuji melalui Play testing track.

### Backend dan auth

- [ ] API production memakai HTTPS dan storage persisten.
- [ ] Tidak ada secret dalam AAB/repository/log.
- [ ] Google Web Client ID sama pada Flutter dan backend.
- [ ] Android OAuth client memakai package final + SHA-1 Play App Signing.
- [ ] Akun review Play Console aktif.

### Policy

- [ ] Privacy policy tersedia di web dan dalam app.
- [ ] Data safety sesuai perilaku app dan SDK.
- [ ] Penghapusan akun tersedia di app dan web.
- [ ] Content rating, target audience, ads, dan app access selesai.
- [ ] Premium memakai billing dan verifikasi server-side sebelum dijual.

### Store listing dan release

- [ ] Description hanya menjanjikan fitur yang sudah aktif.
- [ ] Icon 512x512, feature graphic 1024x500, dan screenshot siap.
- [ ] Internal/closed testing selesai.
- [ ] Pre-launch report, crash, dan ANR diperiksa.
- [ ] Release notes serta staged rollout disiapkan.

## 18. Membuat update berikutnya

1. Naikkan `version`/`versionCode` di `pubspec.yaml`.
2. Jalankan test dan build AAB dengan production dart-defines yang sama.
3. Upload ke Internal atau Closed testing terlebih dahulu.
4. Pastikan migration backend/database backward-compatible dengan versi app
   lama selama rollout.
5. Promote release yang sama ke Production jika memungkinkan.
6. Pantau vitals dan hentikan staged rollout jika ada regresi.

Package ID dan signing identity harus tetap konsisten. Play Console akan
menolak update dengan package berbeda, versionCode yang tidak meningkat, atau
signature upload yang salah.

## Referensi resmi

- [Flutter: Build and release an Android app](https://docs.flutter.dev/deployment/android)
- [Android: Sign your app](https://developer.android.com/studio/publish/app-signing)
- [Play Console: Create and set up an app](https://support.google.com/googleplay/android-developer/answer/9859152)
- [Target API requirements](https://developer.android.com/google/play/requirements/target-sdk)
- [Testing requirements for new personal accounts](https://support.google.com/googleplay/android-developer/answer/14151465)
- [Store listing asset requirements](https://support.google.com/googleplay/android-developer/answer/9866151)
- [Data safety form](https://support.google.com/googleplay/android-developer/answer/10787469)
- [Account deletion requirements](https://support.google.com/googleplay/android-developer/answer/13327111)
- [Google Play payments policy](https://support.google.com/googleplay/android-developer/answer/9858738)
- [Google Play Billing integration](https://developer.android.com/google/play/billing/integrate)
