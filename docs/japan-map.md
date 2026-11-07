# Peta Jepang

Status: **kode backend, admin, dan aplikasi sudah ditulis; belum dijalankan di lingkungan lengkap.** Lihat bagian "Implementasi" di bawah untuk langkah verifikasi dan deploy.

## Konsep

Peta Jepang untuk belajar budaya dan bahasa lewat tempat, sekaligus melihat
lowongan kerja per daerah. Gratis. Tempat wisata terkenal tampil dari sisi
budaya dan bahasa, bukan sebagai panduan wisata murni.

## Keputusan

| Area | Keputusan |
|---|---|
| Peta | Ilustrasi 47 prefektur, offline, zoom dua jari, Okinawa dalam inset. Geometri dari data batas wilayah MLIT (国土数値情報, lihat "Implementasi"). Tidak ada tile atau API key. |
| Lapisan | Chip **Tempat & budaya** dan **Lowongan kerja**, keduanya aktif secara default. Filter kategori: Budaya, Kuil & Shrine, Alam, Kuliner, Kota & modern, Sejarah. |
| Interaksi | Ketuk prefektur → zoom ke sana dan bottom sheet berisi spot dan ringkasan job di prefektur itu. Ketuk pin spot → halaman detail. |
| Job | Satu pin per prefektur di ibu kotanya, dengan angka jumlah job aktif. Hanya prefektur yang punya job. Bottom sheet → "Lihat semua" membuka `/jobs` dengan filter prefektur; aturan Lifetime Access tetap berlaku. |
| Masuk | Kartu "Peta Jepang" di explore Beranda, tombol "Lihat di peta" di layar Jobs. Route `/map`. |
| Konten spot | Admin memasukkan judul en.wikipedia atau ID Wikidata. Backend mengambil ringkasan (Wikipedia id/en), tips berkunjung (Wikivoyage), koordinat, nama Jepang, dan foto (Wikidata/Commons). Gratis, tanpa API key. |
| Penyimpanan | Diambil sekali saat spot ditambahkan, disimpan permanen di MariaDB, ada tombol Refresh di admin. Aplikasi tidak pernah memanggil Wikimedia langsung. |
| Lisensi | Atribusi dan tautan sumber di setiap spot (teks CC BY-SA 4.0). Foto hanya dari Wikimedia Commons (berlisensi bebas) dengan nama pembuat dan lisensinya; admin bisa mengunggah foto sendiri. |
| Belajar | Nama Jepang dari Wikidata, tautan ke kamus. Kosakata dan frasa manual per spot (opsional). |
| Bahasa | Teks manual admin → Wikipedia bahasa itu → en.wikipedia dengan label "Teks asli dalam bahasa Inggris". Admin menandai spot tanpa teks Indonesia. Frasa: arti Indonesia wajib, arti Inggris opsional (frasa disembunyikan di mode Inggris kalau kosong). UI lewat `app_localizations`. |
| Detail spot | Foto → nama, nama Jepang, kategori, prefektur → ringkasan → tips berkunjung → kosakata & frasa → "Buka di Google Maps" dan "Lowongan di prefektur ini (N)" → sumber & lisensi. |
| Offline | Daftar spot dan teks disimpan di HP; foto disimpan setelah dilihat; jumlah job memakai data terakhir. Daftar job tetap butuh koneksi. |
| Harga | Gratis. |

## Rencana teknis

### Backend (Go)

- **Paket `internal/places`**, tabel `map_spots`: ID, ID Wikidata (unik), kategori, prefektur, lat/lng, nama (id/en/ja), bacaan, ringkasan per bahasa beserta bahasa sumbernya, tips Wikivoyage per bahasa, URL artikel per bahasa, foto (URL lokal, pembuat, lisensi, URL lisensi, halaman Commons), teks manual (ringkasan/tips id dan en), frasa (JSON: jepang, bacaan, arti id, arti en), status terbit, waktu ambil, revisi.
- **Pengambil data** (sekali per spot, saat diimpor atau di-refresh):
  - Wikidata `wbgetentities`: sitelinks (idwiki, enwiki, jawiki, enwikivoyage, idwikivoyage), koordinat (P625), foto (P18), label Jepang.
  - Wikipedia REST `page/summary` untuk id dan en; Wikivoyage `page/summary` untuk tips.
  - Commons `imageinfo` + `extmetadata` untuk pembuat dan lisensi foto. Thumbnail diunduh sekali dan disimpan lewat penyimpanan aset yang sudah ada.
  - Prefektur dihitung dari koordinat dengan point-in-polygon memakai geometri prefektur yang sama dengan peta (tidak perlu daftar ID Wikidata prefektur). Admin bisa mengubahnya.
  - Bacaan nama Jepang dibuat dengan tokenizer kagome yang sudah dipakai backend, bisa diedit admin.
  - User-Agent jelas sesuai kebijakan Wikimedia (nama aplikasi + URL halaman bantuan, tanpa email), satu request berjalan sekaligus, timeout dan validasi ukuran respons seperti kamus.
- **Endpoint publik**: `GET /v1/map/spots` (daftar ringkas yang sudah terbit, dengan revisi untuk cache), `GET /v1/map/spots/{id}` (detail), `GET /v1/jobs/prefecture-counts` (jumlah job aktif per prefektur).
- **Admin**: `/admin/map` (HTML seperti admin kamus/banner) dengan token admin; impor dari judul atau ID Wikidata, edit teks manual dan frasa, unggah foto, terbitkan atau sembunyikan, refresh; penanda "belum ada teks Indonesia". Endpoint di bawah `/v1/admin/map/...`.
- **Egress server**: izinkan HTTPS ke `www.wikidata.org`, `id.wikipedia.org`, `en.wikipedia.org`, `en.wikivoyage.org`, `id.wikivoyage.org`, `commons.wikimedia.org`, `upload.wikimedia.org`.
- Tes memakai transport palsu seperti tes kamus; tes MariaDB lewat `testdb`.

### Geometri peta

- Skrip sekali jalan di `scripts/` mengubah batas prefektur Natural Earth (admin-1, public domain) menjadi `assets/map/japan_prefectures.json`: poligon yang disederhanakan dan diproyeksikan, kode prefektur `01`–`47`, titik ibu kota, dan kotak inset Okinawa. Berkas yang sama dipakai backend untuk point-in-polygon.

### Aplikasi (Flutter)

- `lib/features/map/`: `CustomPainter` untuk prefektur (hit-test dengan `Path.contains`), `InteractiveViewer` untuk zoom, pin spot dan pin job sebagai overlay yang posisinya dihitung dari proyeksi yang sama.
- Bottom sheet prefektur, halaman detail spot, chip lapisan dan kategori.
- "Buka di Google Maps" memakai `url_launcher` ke `https://www.google.com/maps/search/?api=1&query=<lat>,<lng>` (tanpa API key).
- Cache offline memakai mekanisme offline yang sudah ada; route `/map` dan `/map/spot/:id`; kartu explore di Beranda dan aksi di layar Jobs; string di `app_localizations`.
- Tes widget untuk peta, bottom sheet, detail, dan fallback bahasa.

### Urutan kerja

1. Review daftar spot awal (di bawah).
2. Backend: tabel, pengambil data, endpoint publik, jumlah job per prefektur, admin, tes.
3. Geometri peta.
4. Layar aplikasi, titik masuk, offline, lokalisasi, tes.
5. Impor daftar spot awal lewat admin, lengkapi teks Indonesia dan frasa untuk spot populer.

## Daftar spot awal (untuk direview)

Kolom **Judul** adalah judul en.wikipedia yang akan dimasukkan di admin
(redirect diikuti otomatis). Prefektur di sini hanya untuk review; nilai akhirnya
dihitung dari koordinat. Hapus, ganti, atau tambahkan baris sesuai keinginan.

| # | Judul en.wikipedia | Kategori | Prefektur |
|---|---|---|---|
| 1 | Sapporo Snow Festival | Budaya | 01 Hokkaido |
| 2 | Otaru Canal | Sejarah | 01 Hokkaido |
| 3 | Shiretoko National Park | Alam | 01 Hokkaido |
| 4 | Aomori Nebuta Matsuri | Budaya | 02 Aomori |
| 5 | Chūson-ji | Kuil & Shrine | 03 Iwate |
| 6 | Matsushima | Alam | 04 Miyagi |
| 7 | Risshaku-ji | Kuil & Shrine | 06 Yamagata |
| 8 | Ōuchi-juku | Sejarah | 07 Fukushima |
| 9 | Nikkō Tōshō-gū | Kuil & Shrine | 09 Tochigi |
| 10 | Kusatsu Onsen | Alam | 10 Gunma |
| 11 | Kawagoe, Saitama | Sejarah | 11 Saitama |
| 12 | Shibuya Crossing | Kota & modern | 13 Tokyo |
| 13 | Hachikō | Budaya | 13 Tokyo |
| 14 | Sensō-ji | Kuil & Shrine | 13 Tokyo |
| 15 | Meiji Shrine | Kuil & Shrine | 13 Tokyo |
| 16 | Akihabara | Kota & modern | 13 Tokyo |
| 17 | Harajuku | Budaya | 13 Tokyo |
| 18 | Tokyo Skytree | Kota & modern | 13 Tokyo |
| 19 | Tsukiji fish market | Kuliner | 13 Tokyo |
| 20 | Kōtoku-in | Kuil & Shrine | 14 Kanagawa |
| 21 | Yokohama Chinatown | Kuliner | 14 Kanagawa |
| 22 | Sado Island | Alam | 15 Niigata |
| 23 | Tateyama Kurobe Alpine Route | Alam | 16 Toyama |
| 24 | Kenroku-en | Alam | 17 Ishikawa |
| 25 | Arakurayama Sengen Park | Kuil & Shrine | 19 Yamanashi |
| 26 | Jigokudani Monkey Park | Alam | 20 Nagano |
| 27 | Matsumoto Castle | Sejarah | 20 Nagano |
| 28 | Zenkō-ji | Kuil & Shrine | 20 Nagano |
| 29 | Shirakawa-gō | Sejarah | 21 Gifu |
| 30 | Takayama, Gifu | Sejarah | 21 Gifu |
| 31 | Mount Fuji | Alam | 22 Shizuoka |
| 32 | Nagoya Castle | Sejarah | 23 Aichi |
| 33 | Toyota Commemorative Museum of Industry and Technology | Kota & modern | 23 Aichi |
| 34 | Ise Grand Shrine | Kuil & Shrine | 24 Mie |
| 35 | Lake Biwa | Alam | 25 Shiga |
| 36 | Fushimi Inari-taisha | Kuil & Shrine | 26 Kyoto |
| 37 | Kinkaku-ji | Kuil & Shrine | 26 Kyoto |
| 38 | Kiyomizu-dera | Kuil & Shrine | 26 Kyoto |
| 39 | Arashiyama | Alam | 26 Kyoto |
| 40 | Gion | Budaya | 26 Kyoto |
| 41 | Nishiki Market | Kuliner | 26 Kyoto |
| 42 | Dōtonbori | Kuliner | 27 Osaka |
| 43 | Osaka Castle | Sejarah | 27 Osaka |
| 44 | Shinsekai | Kota & modern | 27 Osaka |
| 45 | Himeji Castle | Sejarah | 28 Hyogo |
| 46 | Arima Onsen | Alam | 28 Hyogo |
| 47 | Tōdai-ji | Kuil & Shrine | 29 Nara |
| 48 | Nara Park | Alam | 29 Nara |
| 49 | Kumano Kodō | Sejarah | 30 Wakayama |
| 50 | Mount Kōya | Kuil & Shrine | 30 Wakayama |
| 51 | Tottori Sand Dunes | Alam | 31 Tottori |
| 52 | Izumo-taisha | Kuil & Shrine | 32 Shimane |
| 53 | Kōraku-en | Alam | 33 Okayama |
| 54 | Hiroshima Peace Memorial | Sejarah | 34 Hiroshima |
| 55 | Itsukushima Shrine | Kuil & Shrine | 34 Hiroshima |
| 56 | Kintai Bridge | Sejarah | 35 Yamaguchi |
| 57 | Awa Odori | Budaya | 36 Tokushima |
| 58 | Sanuki udon | Kuliner | 37 Kagawa |
| 59 | Dōgo Onsen | Budaya | 38 Ehime |
| 60 | Shimanto River | Alam | 39 Kochi |
| 61 | Dazaifu Tenman-gū | Kuil & Shrine | 40 Fukuoka |
| 62 | Tonkotsu ramen | Kuliner | 40 Fukuoka |
| 63 | Arita ware | Budaya | 41 Saga |
| 64 | Hashima Island | Sejarah | 42 Nagasaki |
| 65 | Kumamoto Castle | Sejarah | 43 Kumamoto |
| 66 | Mount Aso | Alam | 43 Kumamoto |
| 67 | Hells of Beppu | Alam | 44 Oita |
| 68 | Takachiho Gorge | Alam | 45 Miyazaki |
| 69 | Sakurajima | Alam | 46 Kagoshima |
| 70 | Yakushima | Alam | 46 Kagoshima |
| 71 | Shuri Castle | Sejarah | 47 Okinawa |
| 72 | Okinawa Churaumi Aquarium | Alam | 47 Okinawa |

Catatan untuk review:

- Judul belum diverifikasi ke Wikipedia (Wikipedia tidak bisa diakses dari
  lingkungan penyusunan). Admin akan menolak judul yang tidak ditemukan, jadi
  salah ketik langsung ketahuan saat impor.
- Spot makanan atau festival (Sanuki udon, Tonkotsu ramen, Awa Odori, Sapporo Snow
  Festival) tidak punya satu lokasi pasti di Wikidata. Koordinat atau
  prefekturnya mungkin kosong, sehingga admin perlu mengisinya manual, atau
  spot tersebut diganti tempat yang punya koordinat (misalnya restoran atau
  lokasi festival tertentu).
- Prefektur yang belum terwakili: Akita, Ibaraki, Chiba, Fukui. Bisa ditambahkan
  bila perlu.

## Implementasi

### Perbedaan dari rencana

- **Geometri prefektur** memakai data batas wilayah MLIT (国土数値情報 行政区域データ) lewat paket npm `open-data-jp-prefectures-geojson` (MIT), bukan Natural Earth (tidak bisa diunduh saat pengerjaan). Datanya lebih akurat, tapi wajib atribusi "出典：国土数値情報（行政区域データ）国土交通省" (dicantumkan di field `source` pada JSON). Periksa ketentuan pemakaian MLIT sebelum rilis.
- **Satu endpoint publik** `GET /v1/map/spots` berisi semua data spot yang sudah terbit (teks sudah dipilih per bahasa, dengan ETag), tanpa endpoint detail terpisah. Aplikasi menyimpan seluruh respons sehingga peta dan detail bisa dibuka offline.
- Pulau-pulau Izu/Ogasawara di selatan Tokyo tidak digambar, supaya peta tidak memanjang ke Pasifik.

### Berkas

| Bagian | Berkas |
|---|---|
| Geometri | `scripts/build_japan_map.py` → `assets/map/japan_prefectures.json` dan `backend/internal/places/japan_prefectures.json` |
| Backend | `backend/internal/places/` (model, pengambil Wikimedia, service, repository, penentu prefektur), `backend/internal/api/map.go`, `map_admin.html`, `jobs.Repository.PrefectureCounts`, `news.Reading`, `media` (prefix COS `map/images/`) |
| Aplikasi | `lib/features/map/` (model, repository + cache, layar peta, detail spot, kartu explore), rute `/map` dan `/map/spot/:id`, `/jobs?prefecture=NN`, tombol peta di layar Jobs, kartu di Beranda, `assets/images/explore/japan-map.webp` |
| Tes | `backend/internal/places/places_test.go`, `backend/internal/api/map_test.go`, `backend/internal/jobs/prefecture_counts_test.go`, `test/features/map/map_test.dart` |

### Konfigurasi server

- `MAP_ADMIN_TOKEN` (fallback `BRANDING_ADMIN_TOKEN`) untuk `/admin/map`.
- `MAP_SOURCE_USER_AGENT` (opsional). Default `NihongoMaster/1.0 (+https://nihongo.codeverta.com/support)`, sesuai kebijakan User-Agent Wikimedia.
- Foto disimpan di COS (prefix `map/images/`) bila COS terkonfigurasi; jika tidak, di folder `map-images` di samping konfigurasi branding dan disajikan dari `/map-images/`.
- Izinkan egress HTTPS ke `www.wikidata.org`, `id.wikipedia.org`, `en.wikipedia.org`, `en.wikivoyage.org`, `id.wikivoyage.org`, `commons.wikimedia.org`, `upload.wikimedia.org`.
- Tabel `map_spots` dibuat otomatis saat startup (GORM migrate).

### Verifikasi sebelum rilis

1. Backend: `go build ./...`, `go vet ./...`, `go test ./internal/places ./internal/api ./internal/jobs` (tes jobs memerlukan `MARIADB_TEST_DSN`).
2. Aplikasi: `flutter pub get`, `flutter analyze`, `flutter test test/features/map test/core/localization`.
3. Buka `/admin/map`, impor beberapa spot dari daftar di atas, lengkapi teks Indonesia dan frasa, lalu terbitkan.
4. Coba di HP: kartu Peta Jepang di Beranda, zoom, ketuk prefektur dan pin job, halaman detail, mode offline (matikan internet setelah peta pernah dibuka), dan bahasa Inggris.
