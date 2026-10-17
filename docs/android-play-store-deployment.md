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
- Account, pembelian, progress, dan achievement sekarang persistent di MariaDB;
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
2. Gunakan MariaDB dengan konfigurasi `DB_*` untuk akun, progress,
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
menyimpan upload key untuk menandatangani AAB yang diunggah. Buat upload key
sekali dari PowerShell pada folder utama repository:

```powershell
.\scripts\create_android_upload_key.ps1
```

Script membuat JKS, password acak, alias `nihongo-master-upload`, fingerprint,
dan konfigurasi lokal `android/key.properties`. File signing tersimpan di
`android/.signing-secrets/`; Git sudah mengabaikan folder ini. Simpan backup JKS
dan password di password manager atau penyimpanan terenkripsi. Jika keystore
hilang setelah upload pertama, ikuti prosedur reset upload key di Play Console;
jangan membuat key baru untuk menandatangani update biasa.

## 7. Konfigurasi signing Flutter

Konfigurasi release Gradle membaca upload key dari environment atau file
`android/key.properties` yang dibuat oleh script di bagian sebelumnya. File ini
dan folder `android/.signing-secrets/` diabaikan oleh `.gitignore`.
Release lokal tanpa signing key memakai debug key hanya untuk preview; jangan
unggah hasil itu ke Play Console.

Untuk build GitHub, jangan commit keystore atau password. Di repository buka
**Settings > Secrets and variables > Actions**, lalu tambahkan Actions secrets:

- `ANDROID_KEYSTORE_BASE64`: keystore `.jks` yang diubah ke Base64.
- `ANDROID_KEYSTORE_PASSWORD`
- `ANDROID_KEY_ALIAS`
- `ANDROID_KEY_PASSWORD`

Salin nilai masing-masing dari file bernama sama di
`android/.signing-secrets/`, atau buka `github-actions-secrets.txt` untuk
melihat seluruh pasangan nama dan nilai. Tambahkan nilainya di GitHub satu per
satu; jangan commit atau mengirim file tersebut.

Tambahkan Actions variables berikut:

- `API_BASE_URL`: URL backend production yang diawali `https://`.
- `GOOGLE_SERVER_CLIENT_ID`: Web OAuth Client ID berakhiran
  `.apps.googleusercontent.com`.

Workflow menolak build bila salah satu variable/secret ini belum diisi. Client
ID tersebut harus sama dengan `GOOGLE_CLIENT_ID` di backend production. Secret
`ANDROID_KEYSTORE_BASE64` berisi teks Base64, bukan file JKS; script di bagian 6
sudah membuat teksnya di `android/.signing-secrets/ANDROID_KEYSTORE_BASE64.txt`.

## 8. Versioning

Versi Flutter berada di `pubspec.yaml`:

```yaml
version: 1.0.0+1
```

- `1.0.0` menjadi `versionName` yang terlihat pengguna.
- `1` menjadi `versionCode` internal Play Store.
- Setiap AAB update wajib memiliki `versionCode` lebih tinggi. Workflow release
  GitHub bisa menaikkan versi dan build number ini otomatis.

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

Workflow `.github/workflows/android-release.yml` menyediakan mode build biasa
dan release otomatis. Setelah workflow tersedia pada default branch serta
secrets/variables bagian 7 terisi:

1. Buka repository > **Actions > Android release build > Run workflow**.
2. Pilih branch default dan pilih jenis proses:
   - `build`: build AAB/APK tanpa menaikkan versi atau membuat Release.
   - `patch`: naikkan patch, misalnya `1.0.0` menjadi `1.0.1`.
   - `minor`: naikkan minor dan reset patch, misalnya `1.0.1` menjadi `1.1.0`.
   - `major`: naikkan major dan reset minor/patch, misalnya `1.1.0` menjadi `2.0.0`.
3. Klik **Run workflow**. Untuk patch/minor/major, workflow menaikkan build
   number, memperbarui `pubspec.yaml`, mengirim commit versi dan tag `vX.Y.Z`,
   lalu membuat GitHub Release dengan AAB dan APK terlampir.
4. Buka **Releases**, unduh `nihongo-master.aab`, lalu unggah ke Play Console.

Pilihan `build` menghasilkan artifact yang tersedia di halaman run selama 30
hari. Tag versi `v*` yang dibuat manual tetap menjalankan workflow; pastikan
versinya sama dengan `pubspec.yaml`. Workflow otomatis meminta build iOS pada
tag release baru jika sertifikat, provisioning profile, dan variable iOS sudah
disiapkan. Keystore tidak dimasukkan ke artifact atau repository.

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

- Data akun: nama, email, foto profil Google jika digunakan, serta foto profil yang pengguna pilih dan unggah (opsional). Foto unggahan disimpan di Tencent COS agar dapat ditampilkan melalui URL publik.
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

Pilih bahwa sebagian fitur dibatasi login. Untuk memberi akses reviewer ke
seluruh materi premium tanpa pembelian, siapkan akun reviewer khusus di
`backend/.env` pada server production:

```text
PLAY_REVIEW_EMAIL=alamat-email-review-yang-kamu-tentukan
PLAY_REVIEW_PASSWORD=password-kuat-yang-kamu-buat
```

Jangan memakai akun demo development atau akun pribadi. Backend membuat akun
ini setiap kali API dimulai, memberinya tier premium, dan tidak menyimpan
password mentah maupun akun reviewer di database. Akun harus dikonfigurasi dan
login-nya diverifikasi pada backend production sebelum kredensial dimasukkan ke
Play Console. Perubahan konfigurasi memerlukan restart/rebuild backend.

Isi detail Play Console dalam bahasa Inggris:

```text
Name: Nihongo Master review account
Username: nilai PLAY_REVIEW_EMAIL
Password: nilai PLAY_REVIEW_PASSWORD
Any other information required to access your app:
Sign in with the email and password above. No one-time password, two-step verification, or purchase is required. This account has full access to premium content and features. After signing in, open the Profile, Learn, or Exams tabs to review restricted areas.
```

Akun reviewer harus tetap tersedia selama proses review. Akun ini khusus login
email/password, tidak dapat memakai Google Sign-In, dan tidak dapat membeli atau
restore pembelian. Jangan commit kredensial ke repository.

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

1. Jalankan workflow Android release dan pilih `patch`, `minor`, atau `major`.
2. Workflow menaikkan `version`/`versionCode`, membuat tag, dan melampirkan
   AAB baru ke GitHub Release.
3. Upload AAB ke Internal atau Closed testing terlebih dahulu.
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
