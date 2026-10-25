# Authentication and Google OAuth Setup

Nihongo Master menggunakan guest-first navigation. Pengguna dapat membuka
Beranda tanpa akun, tetapi materi, exam, premium action, tab Belajar, Ujian, dan
Profil dilindungi auth gate.

## Alur autentikasi

### Email/password

1. Flutter mengirim email dan password ke `POST /v1/auth/login`.
2. Backend membandingkan password dengan bcrypt hash.
3. Backend membuat opaque random session token 256-bit.
4. Flutter menyimpan token pada Keychain/Keystore melalui secure storage;
   token lama di SharedPreferences dipindahkan pada pemulihan sesi berikutnya.
5. Dio mengirim `Authorization: Bearer <token>` pada request selanjutnya.
6. Startup menampilkan profil akun tersimpan saat jaringan gagal, lalu mencoba
   memvalidasi sesi melalui `GET /v1/auth/me`. Jawaban jaringan atau HTTP 5xx
   mempertahankan sesi; HTTP 401 yang menandakan token dicabut menghapus token.
7. Logout memanggil backend untuk mencabut token lalu membersihkan storage lokal
   milik akun, termasuk profil cache dan bukti akses offline.

### Pemulihan akun dan akses premium offline

Profil akun terakhir disimpan untuk memulihkan identitas saat aplikasi dibuka
tanpa jaringan. Profil cache tidak menyimpan entitlement. Materi premium offline
memerlukan bukti Ed25519 dari backend yang terikat pada ID akun dan
berlaku 30 hari sejak verifikasi Google Play/App Store terakhir yang berhasil.
Server tidak memperpanjang bukti dari flag akun lama; lease baru diterbitkan
hanya setelah purchase aktif diverifikasi store. Jawaban billing yang berhasil
menegaskan status free atau refund menghapus lease lokal. Jika server/store
sementara gagal, aplikasi mempertahankan lease yang telah terverifikasi sampai
expiry.

Jalankan `go run ./cmd/offline-keys` dari folder `backend` untuk membuat atau
menggunakan kembali key stabil. Helper menyimpan private Ed25519 key dan
practice HMAC secret di `backend/secrets/`, tidak pernah mencetak nilainya,
serta menampilkan public pin dan path konfigurasi. Simpan path private pada
`backend/.env` dan public pin pada `OFFLINE_ACCESS_PUBLIC_KEY`. Build Flutter
menerima public pin melalui `--dart-define=OFFLINE_ACCESS_PUBLIC_KEY=<base64>`.
Launcher `dart run.dart` meneruskan public pin dari environment atau
`backend/.env`; CI release Android/iOS memakai variable public yang sama.
Tanpa key server atau public pin yang cocok, offline premium ditolak dengan
aman. Rotasi key memerlukan build aplikasi yang mem-pin public key baru.

Rincian klaim lease dan konfigurasi deployment ada di
[panduan offline access](offline-access.md).

Demo development account:

```text
Email: demo@nihongomaster.app
Password: Demo123!
```

Account demo hanya untuk local development dan harus dihapus sebelum production.

### Google OAuth

1. Flutter memulai Google Sign-In melalui package `google_sign_in`.
2. Google mengembalikan OpenID Connect ID token.
3. Flutter mengirim ID token ke `POST /v1/auth/google`.
4. Backend memvalidasi signature, audience, issuer, expiry, subject, dan verified
   email menggunakan official Google API Go client.
5. Backend membuat atau menghubungkan user lokal.
6. Backend mengembalikan opaque session token milik Nihongo Master.

Backend tidak mempercayai Google user ID atau email yang dikirim terpisah oleh
client. Identitas hanya dibaca dari ID token yang sudah terverifikasi.

## Memahami dua environment variable

Kedua variable berikut menggunakan **nilai Web OAuth Client ID yang sama**:

| Variable | Dipakai oleh | Cara pemberian |
|---|---|---|
| `GOOGLE_CLIENT_ID` | Backend Go | Environment variable saat server dijalankan |
| `GOOGLE_SERVER_CLIENT_ID` | Flutter | Compile-time `--dart-define` |

Contoh bentuk nilainya:

```text
123456789012-abcdefghijklmnopqrstuvwxyz.apps.googleusercontent.com
```

Walaupun namanya berbeda, relasinya adalah:

```text
GOOGLE_CLIENT_ID == GOOGLE_SERVER_CLIENT_ID == Web OAuth Client ID
```

Android OAuth Client ID adalah credential kedua. Nilainya tidak ditempel ke
source code atau command Flutter; Google mencocokkannya otomatis menggunakan
package name dan SHA-1 aplikasi.

Client ID bukan secret. Namun **OAuth Client Secret tidak digunakan oleh mobile
app ini dan tidak boleh dimasukkan ke Flutter, Git, atau APK**.

## Langkah 1: buat Google Cloud project

1. Buka [Google Cloud Console](https://console.cloud.google.com/).
2. Gunakan project selector di bagian atas.
3. Pilih **New Project**.
4. Beri nama, misalnya `Nihongo Master Development`.
5. Setelah selesai, pastikan project tersebut sedang aktif.

Disarankan membuat project Google Cloud terpisah untuk development dan
production supaya package, fingerprint, user testing, dan metrics tidak
tercampur.

## Langkah 2: konfigurasi Google Auth Platform

1. Buka menu **Google Auth Platform**.
2. Jika tampil tombol **Get Started**, tekan tombol tersebut.
3. Pada halaman **Branding**, isi:
   - App name: `Nihongo Master`
   - User support email: email developer/support
   - Developer contact information: email developer
4. Pada **Audience**, pilih:
   - **External** untuk aplikasi yang boleh digunakan akun Google umum; atau
   - **Internal** hanya jika aplikasi khusus satu organisasi Google Workspace.
5. Untuk development, pertahankan publishing status **Testing**.
6. Tambahkan akun Google developer pada bagian **Test users** jika ditampilkan.
7. Pada **Data Access**, gunakan scope identitas minimum:
   - `openid`
   - `email`
   - `profile`

Nihongo Master tidak meminta akses Gmail, Drive, Calendar, atau data Google lain.

## Langkah 3: buat Web OAuth Client ID

Web client inilah yang menghasilkan nilai untuk kedua environment variable.

1. Buka **Google Auth Platform > Clients**.
2. Tekan **Create Client**.
3. Application type: **Web application**.
4. Name: `Nihongo Master Backend`.
5. Untuk Android-only development, Authorized JavaScript origins dan Authorized
   redirect URIs dapat dibiarkan kosong.
6. Tekan **Create**.
7. Salin nilai **Client ID** yang berakhiran
   `.apps.googleusercontent.com`.

Simpan sebagai contoh berikut:

```text
WEB_CLIENT_ID=123456789012-abc.apps.googleusercontent.com
```

Nilai **Client Secret** yang mungkin ditampilkan tidak digunakan oleh alur ID
token ini.

## Langkah 4: dapatkan package name Android

Package name dapat dilihat pada `android/app/build.gradle.kts`, bagian
`applicationId`.

Nilai proyek saat ini:

```text
com.codeverta.nihongomaster
```

Sebelum production, ganti package contoh tersebut dengan application ID resmi,
misalnya `com.perusahaan.nihongomaster`, lalu buat Android OAuth client baru.

## Langkah 5: dapatkan SHA-1 Android

### Nilai debug pada mesin proyek saat ini

```text
SHA-1: 0D:26:28:A4:7C:B4:16:68:1D:1F:F5:0A:00:57:9C:06:EC:C9:40:58
SHA-256: B9:7D:B7:F9:4D:CF:8B:82:39:44:10:18:63:C6:31:0D:51:85:F1:E4:9E:F7:28:1C:9C:42:1E:3C:AE:67:CA:79
```

Fingerprint ini berasal dari `C:\Users\rabih\.android\debug.keystore` dan bisa
berbeda pada komputer developer lain.

### Menghitung melalui Gradle

Pastikan terminal menggunakan Java 17 atau lebih baru. Pada mesin ini JDK
Android Studio berada di `C:\Program Files\Android\Android Studio\jbr`.

```powershell
$env:JAVA_HOME = 'C:\Program Files\Android\Android Studio\jbr'
cd android
.\gradlew.bat signingReport
```

Cari output variant `debug`, lalu salin `SHA1`.

### Menghitung langsung melalui keytool

```powershell
& 'C:\Program Files\Android\Android Studio\jbr\bin\keytool.exe' `
  -list -v `
  -alias androiddebugkey `
  -keystore "$env:USERPROFILE\.android\debug.keystore" `
  -storepass android `
  -keypass android
```

Untuk release, gunakan keystore release milik tim dan jangan memakai debug
fingerprint. Jika Google Play App Signing diaktifkan, tambahkan juga SHA-1 App
Signing Certificate dari **Play Console > App Integrity**.

Saat ini `android/app/build.gradle.kts` masih mengonfigurasi variant release
dengan debug signing agar development build mudah dijalankan. Karena itu output
`signingReport` sementara menunjukkan SHA-1 yang sama untuk debug dan release.
Ganti signing configuration tersebut sebelum distribusi production, lalu buat
Android OAuth client untuk SHA-1 release yang baru.

## Langkah 6: buat Android OAuth Client ID

1. Kembali ke **Google Auth Platform > Clients**.
2. Tekan **Create Client**.
3. Application type: **Android**.
4. Name: `Nihongo Master Android Debug`.
5. Package name: `com.codeverta.nihongomaster`.
6. SHA-1: masukkan fingerprint debug di atas.
7. Tekan **Create**.

Untuk production, buat client terpisah untuk kombinasi package dan SHA-1 release.
Satu Google Cloud project boleh memiliki beberapa Android OAuth client.

## Menyimpan credential backend pada `.env`

File lokal sudah disiapkan pada `backend/.env` dan diabaikan oleh Git. Formatnya:

```dotenv
PORT=8080
GOOGLE_CLIENT_ID=YOUR_WEB_CLIENT_ID.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=YOUR_WEB_CLIENT_SECRET
GOOGLE_ANDROID_CLIENT_ID=YOUR_ANDROID_CLIENT_ID.apps.googleusercontent.com
```

Jangan menjalankan Flutter dengan `--dart-define-from-file=backend/.env`.
File tersebut mengandung `GOOGLE_CLIENT_SECRET`, sedangkan semua compile-time
variable Flutter dapat ikut terbawa ke APK. Terus gunakan `--dart-define`
eksplisit seperti contoh di bawah agar hanya Client ID publik yang masuk ke app.

Backend otomatis memuat `.env` saat dijalankan dari folder backend maupun root
repository. Environment variable yang sudah diberikan oleh OS/deployment tetap
memiliki prioritas lebih tinggi daripada isi file.

`GOOGLE_CLIENT_SECRET` disimpan untuk kebutuhan konfigurasi server di masa
depan, tetapi tidak digunakan oleh endpoint ID-token login saat ini. Jangan
pernah mengirim nilainya ke Flutter.

## Menjalankan backend dengan Google OAuth

PowerShell:

```powershell
cd backend
go run ./cmd/api
```

Bash/zsh:

```bash
cd backend
go run ./cmd/api
```

Backend akan mengembalikan `503 google_not_configured` jika variable tersebut
belum diisi. Token dengan audience berbeda akan ditolak dengan `401`.

## Menjalankan Flutter dengan Google OAuth

Dari folder utama project di Windows, Linux, atau macOS, jalankan:

```text
dart run.dart
```

Launcher membaca `GOOGLE_SERVER_CLIENT_ID` dari `backend/.env` (atau memakai
`GOOGLE_CLIENT_ID` jika key itu kosong), menjalankan backend, lalu meneruskan
Web Client ID tersebut ke build Flutter. Jika kedua key diisi, nilainya harus
sama. Lihat [panduan menjalankan project](running-project.md) untuk memilih
device atau memakai ponsel fisik.

Jika menggunakan Firebase `google-services.json`, pastikan file tersebut berisi
OAuth web client (`client_type: 3`). Implementasi proyek saat ini tidak
mewajibkan Firebase dan dapat berjalan hanya dengan `serverClientId`.

## Checklist sebelum menguji

- [ ] Google Cloud project yang benar sedang aktif.
- [ ] OAuth Branding dan Audience sudah disimpan.
- [ ] Akun developer ditambahkan sebagai Test user jika diperlukan.
- [ ] Web OAuth Client ID sudah dibuat.
- [ ] Android OAuth client memakai package yang tepat.
- [ ] Android OAuth client memakai SHA-1 debug yang tepat.
- [ ] `GOOGLE_CLIENT_ID` backend berisi Web Client ID.
- [ ] `GOOGLE_SERVER_CLIENT_ID` Flutter berisi Web Client ID yang sama.
- [ ] Backend dijalankan sebelum tombol Google ditekan.

## Troubleshooting Google Sign-In

### `GOOGLE_SERVER_CLIENT_ID belum dikonfigurasi`

Launcher tidak menerima ID saat app dijalankan. Jalankan ulang dengan
`dart run.dart` dari folder utama project; hot reload tidak dapat mengubah
compile-time variable. Launcher mengambil nilainya dari `backend/.env`.

### `google_not_configured`

Backend dijalankan tanpa `GOOGLE_CLIENT_ID`. Set environment variable pada
terminal yang sama sebelum `go run`.

### `invalid_google_token`

Biasanya Web Client ID pada backend berbeda dengan ID yang digunakan Flutter.
Pastikan kedua nilainya sama persis dan tidak memakai Android Client ID.

### `clientConfigurationError` atau login dianggap canceled

Periksa tiga nilai berikut:

1. package name Android;
2. SHA-1 dari build variant yang sedang dijalankan;
3. Web Client ID pada `GOOGLE_SERVER_CLIENT_ID`.

Plugin Android dapat melaporkan beberapa kesalahan konfigurasi sebagai canceled
setelah pengguna memilih akun.

### Berfungsi pada debug tetapi gagal pada release

Debug dan release memakai signing certificate berbeda. Tambahkan OAuth Android
client untuk SHA-1 release atau SHA-1 Google Play App Signing.

## Guest auth gate

Routing tidak memaksa login ketika `/home` dibuka. Gate dijalankan ketika guest:

- memilih tab Belajar;
- memilih tab Ujian;
- memilih tab Profil;
- menekan kartu materi di Beranda;
- menekan JLPT level, exam, premium, atau action pembelajaran lain.

Setelah login dari pilihan tab, aplikasi otomatis membuka tab yang semula
dipilih. Tombol **Lanjut sebagai tamu** menutup login tanpa membuka fitur.

## Logout

Logout tersedia di tab Profil. Prosesnya:

1. konfirmasi dari pengguna;
2. `POST /v1/auth/logout` dengan bearer token;
3. backend menghapus token dari session store;
4. Flutter menghapus token lokal dan melakukan Google sign-out;
5. MainScaffold kembali ke tab Beranda dalam mode guest.

Local logout tetap selesai jika backend sedang offline agar pengguna tidak
terkunci di perangkat.

## Catatan production

- Session saat ini in-memory dan hilang saat backend restart.
- Pindahkan session ke Redis/PostgreSQL dengan expiry dan device metadata.
- Gunakan HTTPS; jangan pernah mengirim token melalui HTTP production.
- Tambahkan rate limit untuk endpoint login dan Google exchange.
- Tambahkan account registration, email verification, reset password, dan MFA.
- Pindahkan demo password dari seed source.
- Mobile session token saat ini disimpan pada secure storage/Keychain/Keystore.
- Rotasi session setelah perubahan password atau security event.
