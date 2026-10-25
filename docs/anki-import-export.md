# Import dan export deck Anki

## Pemakaian di aplikasi

1. Masuk ke akun, buka **Anki**, lalu tekan ikon **Import deck** di kanan atas.
2. Pilih file `.apkg`, `.csv`, atau `.json`. Periksa jumlah kartu/media serta contoh kartu, ubah nama jika perlu, lalu konfirmasi impor.
3. Deck muncul sebagai deck pribadi di katalog akun. Deck bawaan tetap tersedia. Belajar dan penilaian kartu memakai SRS Nihongo Master dengan jadwal baru.
4. Untuk backup, buka detail deck dan tekan **Export deck**. Pilih format, lalu simpan atau bagikan file melalui dialog perangkat.
5. Deck pribadi dapat dihapus dari detailnya setelah konfirmasi. Ekspor backup dahulu bila diperlukan.

Impor/ekspor dan pemuatan media pribadi memerlukan koneksi internet. Jangan menutup aplikasi saat konfirmasi impor berlangsung. Jika respons terputus, tekan konfirmasi kembali pada halaman yang sama: ID permintaan tetap sama sehingga deck tidak dibuat dua kali. Memilih file kembali memulai impor baru.

## Format dan batas

| Format | Kartu | Media | Catatan |
| --- | --- | --- | --- |
| APKG | Pertanyaan/jawaban dasar, arti | Gambar/audio yang dibundel | Ekspor memakai koleksi SQLite legacy agar kompatibel. Impor juga membaca kemasan modern zstd. |
| JSON | Semua jawaban dan metadata | Base64 dalam file | Format backup Nihongo Master versi 1, bukan format JSON arbitrer dari aplikasi lain. |
| CSV UTF-8 | `front,answers,meaning,imageUrl,audio` | Hanya URL eksternal | `back` diterima sebagai alias jawaban. Referensi media lokal dikosongkan saat ekspor. |

Ekspor CSV menyertakan baris komentar `#nihongo-master:` berisi metadata base64 JSON sebelum header. Baris ini dibaca kembali oleh Nihongo Master; hapus baris tersebut bila aplikasi spreadsheet lain menganggapnya data. `answers` bisa berupa teks biasa atau array JSON; CSV mengutip koma, tanda kutip, dan baris baru sesuai aturan CSV.

Contoh CSV sederhana yang dapat diimpor:

```csv
front,back,meaning
猫,ねこ,kucing
犬,いぬ,anjing
```

Contoh JSON minimal:

```json
{"version":1,"title":"Kosakata pribadi","cards":[{"front":"猫","answers":["ねこ"],"meaning":"kucing"}]}
```

Batas file impor maupun hasil ekspor 20 MiB, maksimal 10.000 kartu/deck, 50 deck/akun, dan 100 MiB total kartu+media/akun. File media individual maksimal 10 MiB; total media/deck 50 MiB; total koleksi dan media setelah dekompresi maksimal 60 MiB. Gambar maksimal 20 megapiksel. Format media: JPEG, PNG, GIF, WebP, MP3, Ogg, WAV, MP4/M4A yang dikenali dari isi file. SVG dan HTML ditolak sebagai media.

Ekspor gagal dengan pesan bila seluruh deck melampaui batas; kartu tidak dipotong diam-diam. Coba APKG untuk ukuran lebih ringkas atau pisahkan deck di aplikasi sumber.

## Kompatibilitas Anki

- Satu kartu dasar per note: field Front/Question/Word menjadi pertanyaan; Back/Answer/Reading menjadi jawaban; field lainnya digabung menjadi arti. Jika nama field berbeda, dua field pertama dipakai. Pratinjau wajib diperiksa sebelum konfirmasi.
- HTML diubah menjadi teks. Skrip, CSS, dan template interaktif tidak dijalankan. Satu gambar dan satu audio per kartu didukung; kartu dengan beberapa media berbeda ditolak agar tidak kehilangan media tanpa pemberitahuan.
- Cloze, reverse/secondary cards, dan note dengan beberapa template ditolak. Gunakan ekspor deck dasar dari Anki untuk bentuk tersebut.
- Progres, jadwal Anki, XP, serta riwayat review tidak dipindahkan. Impor membuat jadwal SRS baru.
- URL gambar/audio eksternal tetap menjadi URL; backend tidak mengunduhnya. File media dalam paket disimpan di MariaDB dan disertakan kembali dalam APKG/JSON.
- Metadata sumber dan lisensi deck bawaan disertakan dalam hasil ekspor. Atribusi paket pribadi mengikuti metadata yang disediakan pemiliknya.
- Struktur paket diperiksa terhadap [protokol resmi Anki](https://github.com/ankitects/anki/blob/main/proto/anki/import_export.proto).

## Backend dan keamanan akses

Semua endpoint transfer wajib bearer token:

| Endpoint relatif `/v1` | Kegunaan |
| --- | --- |
| `POST /anki/import/preview` | Multipart `file`; mengembalikan metadata, tiga contoh, SHA-256. Tidak menyimpan deck. |
| `POST /anki/import` | Multipart `file`, `title` opsional, `sha256`, UUID `requestId`. Retry UUID yang sama mengembalikan deck semula. |
| `GET /anki/decks/{id}/export?format=apkg` | Hasil biner APKG/CSV/JSON seluruh deck. |
| `DELETE /anki/decks/{id}` | Hapus deck pribadi, media, dan jadwal review secara transaksional. |
| `GET /anki/decks/{id}/media/{name}` | Media pribadi melalui pemeriksaan pemilik akun. |

`GET /anki/decks` menambahkan deck pribadi hanya untuk akun yang sedang masuk. Akun lain mendapatkan 404 untuk deck/media pribadi. Deck premium tetap diperiksa di backend pada belajar, review, export, dan media; mengimpor ulang sumber premium yang dikenali tidak menurunkan tingkat aksesnya.

Migrasi GORM otomatis membuat `user_anki_decks` dan `user_anki_deck_media` saat persistence backend aktif. Kartu disimpan sebagai JSON LONGTEXT, media sebagai LONGBLOB terpisah. Startup hanya membaca metadata ukuran media; blob diambil saat dibutuhkan. Penghapusan akun menghapus kedua tabel untuk akun tersebut dalam transaksi penghapusan akun.

Validasi mencakup nama/path arsip, MIME dari isi file, ukuran, batas dekompresi, referensi media, koleksi SQLite read-only, penolakan view/trigger/virtual table, dan timeout query. Maksimal dua pemrosesan transfer berat bersamaan per proses server. Cache materi akun dibatalkan setelah import/delete. Respons transfer/media bersifat privat dan tidak disimpan pada cache HTTP bersama.

## Menjalankan perubahan

Backend harus dijalankan dari versi kode ini agar route dan migrasi tersedia. Aplikasi menambahkan `file_selector`, `share_plus`, dan `audioplayers`: lakukan build lalu instal ulang APK dengan mempertahankan data. Hot reload saja tidak mendaftarkan plugin native baru. iOS perlu build dari macOS/Xcode; belum dapat diverifikasi dari Windows.

Pengujian:

```text
cd backend
go test ./internal/ankitransfer ./internal/store ./internal/api -run 'TestTransfer|TestCSV|TestAPKG|TestPrivateAnki|TestAnki' -count=1
cd ..
flutter test test/features/anki test/core/offline/learning_cache_test.dart
flutter build apk --debug
```

Tes persistence memakai database terisolasi melalui `MARIADB_TEST_DSN` atau `backend/.env.test`. Jangan arahkan kredensial pengujian ke schema produksi.

## Catatan validasi implementasi

- 38 tes Flutter Anki/cache lolos, termasuk pembatalan cache setelah import/delete, isolasi akun, dan penghapusan deck pribadi setelah akses premium berakhir.
- Analisis statis area Anki, router, lokalisasi, dan tes Anki bersih.
- Seluruh tes package store dan codec backend lolos, termasuk persistence MariaDB, retry setelah restart, dan penghapusan akun.
- Tes API Anki menguji preview, digest berubah, retry impor, ketiga format ekspor, ekspor katalog lengkap, premium gate, serta penghapusan.
- Build APK debug Android berhasil. Build iOS belum dijalankan dari Windows.
- Percobaan suite API penuh masih mendeteksi kegagalan pemetaan bahasa Inggris materi SSW di area lokalisasi yang sudah ada. Hasil pengujian Anki di atas tidak berarti seluruh suite API lolos.
