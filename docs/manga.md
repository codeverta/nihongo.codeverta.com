# Manga Jepang: katalog, reader, MariaDB, dan COS

App membuka **Beranda → Manga Jepang**. Katalog berisi judul terkurasi; detail
menampilkan cover, penulis, genre, sinopsis, daftar chapter, dan tautan sumber
Rawkuma. Reader menyediakan halaman kanan-ke-kiri dengan zoom dan mode scroll.
Tulisan pada gambar dipertahankan dalam bahasa Jepang, tanpa menerjemahkan panel.

## Sumber dan seleksi

Adapter pertama menggunakan `https://rawkuma.net/`. Judul awal: Ruri Dragon,
Ao no Hako, dan Medalist. Sampel halaman dari ketiganya diperiksa sebagai bahasa
Jepang. Metadata `inLanguage` dari theme Rawkuma tidak digunakan karena nilai
`en_US` juga muncul pada chapter Jepang.

Jangan memasukkan homepage/rekomendasi sumber ke katalog secara otomatis.
Admin menambahkan URL detail judul yang sudah ditinjau di `/admin/manga`,
mencentang verifikasi bahasa Jepang dan peninjauan konten. Gunakan konten yang
diizinkan untuk diimpor. Tag kosong ditolak; Adult, Ecchi, Hentai, Smut, Mature,
Harem, dan variasi tag seksual/18+ diblokir saat impor. Daftar terkurasi dan tag
bekerja bersama; filter tag bukan pengganti peninjauan isi chapter. Hilangkan
centang aktif untuk menyembunyikan judul dan semua chapternya seketika dari API.

Gambar reader diekstrak hanya dari CDN sumber untuk slug dan nomor chapter yang
sesuai, dalam urutan DOM. Iklan, tracking pixel, script, iframe, cover lain,
dan rekomendasi sumber tidak masuk reader.

## Alur cache

1. Worker mengklaim lease global MariaDB agar beberapa replika/CLI tidak mengimpor
   sumber bersamaan. Worker berjalan setiap menit, dengan jeda batch lima menit.
2. Deadline detail judul dinaikkan **tujuh hari sebelum request**. Restart,
   kegagalan parsing, atau pemanggilan CLI tidak melewati deadline tersebut.
3. Detail dan daftar chapter disimpan di `manga_series` dan `manga_chapters`.
   Cover diunduh sekali, diperiksa sebagai gambar, lalu disimpan di COS.
4. Chapter awal yang tersedia diantrekan setelah impor judul. Chapter lain hanya
   diantrekan saat pembaca menekan **Siapkan chapter**. POST tersebut hanya
   mengubah antrean DB, tidak melakukan fetch sinkron.
5. Worker memproses maksimal delapan detail judul dan dua chapter per batch.
   Request upstream diberi jarak minimal dua detik. Kegagalan mengambil detail
   chapter dan respons 403/429 diberi cooldown tujuh hari, tanpa proxy/browser
   challenge. Manifest URL halaman disimpan sebelum transfer gambar. Gangguan
   jaringan/COS setelah manifest tersimpan dilanjutkan hanya untuk aset yang
   kurang setelah 5 lalu 30 menit, tanpa mengambil detail chapter kembali;
   setelah tiga kegagalan, tunggu tujuh hari sebelum percobaan berikutnya.
6. Semua gambar chapter diunggah ke COS sebelum chapter berstatus `ready`.
   Chapter parsial disembunyikan. Aset yang sudah berhasil disimpan digunakan
   kembali saat pemulihan, sehingga tidak diunduh ulang.
7. Chapter `ready` tidak di-fetch ulang ketika detail judul diperiksa mingguan.
   Semua GET aplikasi memakai MariaDB; URL halaman dan cover menunjuk ke COS/CDN
   milik aplikasi. Pengguna kedua dan seterusnya memakai cache yang sama.

Cache gambar memakai pemetaan hash URL sumber pada `manga_assets`, serta object
COS berdasarkan hash isi di `manga/pages/`. Batas: 12 MiB/gambar, 120 halaman dan
160 MiB/chapter. JPEG/PNG/WebP diperiksa dimensi dan formatnya. URL sumber memakai
allowlist HTTPS, redirect dibatasi, dan koneksi ke IP privat/loopback diblokir.
`robots.txt` diperiksa per batch dan origin; redirect khusus CDN dari robots
yang tidak ada menuju homepage lama tidak diikuti sebagai dokumen robots.

## Konfigurasi dan operasi

Gunakan konfigurasi COS yang sudah dipakai aplikasi:

```dotenv
COS_BUCKET_URL=https://YOUR_BUCKET.cos.YOUR_REGION.myqcloud.com
COS_SECRET_ID=...
COS_SECRET_KEY=...
ASSET_BASE_URL=https://YOUR_PUBLIC_CDN
MANGA_SYNC_ENABLED=true
MANGA_ADMIN_TOKEN=...
```

`ASSET_BASE_URL` boleh kosong untuk memakai origin bucket. Gambar memakai ACL
public-read seperti aset aplikasi lain. Kredensial hanya berada di backend.
Tanpa COS, impor dan prepare ditolak, sementara katalog/chapter yang sudah
tersimpan tetap bisa dibaca. `MANGA_ADMIN_TOKEN` kosong memakai
`BRANDING_ADMIN_TOKEN`. Worker otomatis memakai lifecycle backend dan berhenti
saat shutdown.

Untuk worker terpisah, matikan `MANGA_SYNC_ENABLED` pada API lalu jalankan dari
folder backend secara periodik; deadline DB dan lease tetap berlaku:

```powershell
go run ./cmd/manga-sync
```

Tidak tersedia flag force yang melewati jatah mingguan. Antrean dibatasi 64
chapter. Reader memeriksa status DB setiap 15 detik selama maksimal lima menit,
berhenti saat layar ditutup/background, dan menyediakan muat ulang manual.

API utama sudah mendaftarkan endpoint Manga. Sebagai opsi, jalankan domain ini
secara terpisah tanpa mengambil lock writer akun:

```powershell
go run ./cmd/manga-api
flutter run
```

Port service terpisah memakai `MANGA_PORT` (default 8081). Debug build dengan
API lokal HTTP port 8080 (10.0.2.2, localhost, atau loopback) otomatis memakai
port 8081 untuk Manga, sehingga `flutter run` biasa tidak meminta route Manga
ke binary API utama yang lama. Jalankan kedua service ketika development.
`MANGA_API_BASE_URL` selalu dapat mengganti alamat tersebut, termasuk mengarah
ke port 8080 jika memakai API utama terbaru. Build release dan backend remote
tetap memakai API utama secara default. Worker kedua service tetap dilindungi
lease yang sama. Untuk produksi dengan service
terpisah, gunakan URL HTTPS publik pada Dart define atau reverse proxy route
Manga ke service itu. API terpisah hanya membuka Manga, admin Manga, dan health;
tidak membuka login, akun, maupun billing.

## API

Semua respons menggunakan envelope `data` standar aplikasi.

- `GET /v1/manga?page=1&limit=20&q=&genre=` — daftar tersimpan, maksimal 40 item.
- `GET /v1/manga/{id}` — metadata dan daftar chapter berurutan.
- `GET /v1/manga/chapters/{id}` — status dan halaman COS jika siap.
- `POST /v1/manga/chapters/{id}/prepare` — antrekan idempotent; 202 bila menunggu,
  200 bila sudah siap/gagal dalam cooldown, 429 bila antrean penuh.
- `GET /v1/admin/manga` — status/jadwal dan kesiapan konfigurasi COS.
- `POST /v1/admin/manga/series` — URL detail Rawkuma, `enabled`,
  `japaneseVerified`, `nonAdultReviewed`; membutuhkan bearer admin.

## Verifikasi

```powershell
go test ./internal/manga ./internal/api
flutter test test/features/manga/manga_test.dart
```

Test MariaDB menggunakan schema terisolasi dari `MARIADB_TEST_DSN`/
`backend/.env.test`; tidak memakai tabel aplikasi produksi. Cakupan meliputi
cooldown lintas restart, antrean paralel, lease beberapa worker, pemulihan
chapter parsial, penyembunyian judul, penolakan konten dewasa, serta pembacaan
berulang tanpa request upstream. Test Flutter memeriksa pencarian/filter,
navigasi detail, prepare eksplisit, reader kanan-ke-kiri, dan penghentian polling.

## Lifetime Access dan atribusi

Seluruh route Manga di aplikasi dan keempat endpoint API memerlukan sesi login
serta entitlement `tier=premium` dan `isLifetime=true`. Akun gratis, pembelian
pending/refund/revoked, dan sesi akun yang dihapus ditolak. API utama menyegarkan
status billing sebelum memberikan akses. Service Manga terpisah memeriksa
`GET /v1/billing` pada `MANGA_AUTH_API_BASE_URL` (default
`http://127.0.0.1:8080/v1`) dengan bearer token pengguna, tanpa menyimpan grant
akses. Verifikasi gagal berarti akses ditolak. Respons API bersifat private/no-store.

Beranda memuat pintasan Manga premium dan slide carousel, Anki, SSW, dan
lowongan. Belajar berfokus pada JLPT. Profil → Sumber data & atribusi berisi
sumber-sumber aplikasi; tautan sumber chapter tetap ditampilkan kecil.
