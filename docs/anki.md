# Anki / Flashcards

Buka **Belajar → Anki · Flashcards**. Daftar deck memiliki pencarian nama/topik dan filter kategori. Setiap deck membuka halaman detail terlebih dahulu: pengantar bahasa Indonesia, jenis latihan, jumlah kartu, progress, contoh pertanyaan/jawaban, serta atribusi sumber. Contoh: Wakayama dijelaskan sebagai prefektur Jepang dengan deck bacaan nama tempat, sedangkan `stations_wakayama` khusus nama stasiun.

## Resource backend

Seluruh **191 JSON** dalam `mistval/kotoba/resources/quiz_data` disalin dari commit `a0f8f984bf5dd8014527c6608d36b479ab622b07`. Dataset sumber berisi **1.461.909 entri**. Salinan asli dipertahankan byte demi byte dalam gzip; manifest mencatat Git blob, SHA-256, dan ukuran tiap file. Pengujian memverifikasi checksum seluruh file.

Lokasi: `backend/internal/anki/resources/`.

- `source/*.json.gz`: seluruh JSON asli, termasuk metadata, atribusi, entri kosong, dan konfigurasi strategi Kotoba.
- `cards/{deckId}/*.json.gz`: pasangan pertanyaan/jawaban yang valid, 256 kartu per chunk.
- `catalog.json`: 191 deck dengan judul, penjelasan, instruksi belajar, kategori, preview, jumlah entri dan status ketersediaan.
- `manifest.json`: provenance dan checksum sumber.
- `KOTOBA-LICENSE.txt`: lisensi repository Kotoba, disertakan juga dalam API credits.

Ada **187 deck aktif / 1.416.940 kartu belajar**. Tiga deck `esyn_easy`, `esyn_med`, `esyn_hard` memiliki jawaban kosong dan membutuhkan thesaurus eksternal pada mesin asal. `strokes_kanken` berisi placeholder `0`, yang baru diganti jawaban sebenarnya setelah mesin asal memilih dan merender goresan. Keempat deck tetap disimpan serta ditampilkan dengan alasan ketersediaan; tombol belajar tidak ditampilkan.

Sebanyak 2.061 entri lain yang tidak dapat membentuk pasangan valid (`n2m`: 247, `Namae`: 1, `places_full`: 1.813) dipertahankan dalam arsip asli dan dikecualikan dari sesi. Tidak ada jawaban rekaan atau layanan sinonim berbayar. Anagram diacak secara deterministik saat normalisasi. Nomor sumber asli tetap tersedia pada `sourceIndex`.

Data di-embed ke binary Go, sehingga tidak perlu request GitHub saat pengguna mencari atau belajar. Backend hanya membaca chunk yang diperlukan dengan cache maksimum 16 chunk; aplikasi menerima hingga 20 kartu per sesi. Dataset besar tidak dibundel ke APK atau diunduh sekaligus. Dua deck bergambar tetap menggunakan URL gambar asli; UI memiliki fallback ketika media eksternal tidak dapat dimuat. Arti/definisi mengikuti bahasa sumber, umumnya Inggris; pengantar dan UI ditulis dalam bahasa Indonesia.

## Jadwal review dan persistensi

Pengguna mengingat jawaban sebelum membalik kartu, kemudian memilih **Belum hafal**, **Sulit**, **Ingat**, atau **Mudah**. Jawaban alternatif ditampilkan lengkap; daftar yang sangat panjang dilipat agar mudah dibaca. Ini merupakan sistem flashcard dengan pengulangan terjadwal dalam Nihongo Master, bukan sinkronisasi AnkiWeb, import `.apkg`, atau implementasi FSRS.

Backend mendahulukan kartu jatuh tempo, kemudian kartu baru. Penjadwalan awal: 1 menit / 6 jam / 1 hari / 4 hari. Review berikutnya menyesuaikan interval dan faktor kemudahan; interval dibatasi maksimum 365 hari. Jam server menentukan jatuh tempo, bukan jam perangkat.

State per kartu menyimpan version, due, interval, ease, repetitions, lapses, request ID terakhir dan rating. State per deck menyimpan cursor kartu baru dan revision dataset. State tersimpan bersama data pengguna di `gamification_users.State.Anki` melalui koneksi writer MariaDB yang sudah digunakan aplikasi. Write harus berhasil sebelum cache diubah atau klien berpindah kartu. Seluruh state ikut dibersihkan oleh penghapusan akun yang sudah ada. Revision berbeda memulai jadwal baru agar progress tidak salah diterapkan pada dataset berbeda.

Request ulang memakai ID dan rating yang sama; backend mengembalikan hasil sebelumnya. Version lama, review sebelum waktunya, dan kartu baru di luar antrean menghasilkan konflik. Klien menawarkan muat ulang untuk merekonsiliasi konflik. Pada timeout, klien mempertahankan operasi yang belum terkonfirmasi dan menawarkan penyimpanan ulang. Penilaian diri tidak memberi XP yang seolah berasal dari kuis dengan grading server.

Daftar serta detail dapat dilihat tanpa login. Belajar dan menyimpan review memerlukan login. Deck JLPT N3–N1, termasuk variasi arti dan penggunaan kanji, mengikuti Lifetime Access di UI dan backend. Verifikasi entitlement toko tetap dilakukan melalui jalur billing yang ada sebelum akses belajar/review premium.

## API

Semua respons menggunakan envelope `data` / `error` aplikasi.

| Endpoint | Akses | Keterangan |
| --- | --- | --- |
| `GET /v1/anki/decks?q=&category=` | Publik | Katalog tanpa payload kartu; bisa difilter |
| `GET /v1/anki/decks/{deckId}` | Publik / akun opsional | Detail, preview dan progress pemilik sesi; preview premium disembunyikan bila terkunci |
| `GET /v1/anki/credits` | Publik | Repository, revision, lisensi repository dan catatan atribusi |
| `GET /v1/anki/decks/{deckId}/study?limit=20` | Login | 1–20 kartu; due dahulu, lalu baru; version dan estimasi interval tiap rating |
| `POST /v1/anki/decks/{deckId}/review` | Login | Body `index`, `version`, `rating` (`again/hard/good/easy`), `requestId` UUID |

Respons review berisi `saved`, `due`, `version`. Error: 401 login, 402 premium, 404 deck tidak ditemukan, 409 jadwal/version konflik, 422 input atau deck belum dapat digunakan. Respons yang memuat progress memakai `Cache-Control: private, no-store`.

## Membangun ulang resource

Jalankan dari root repository:

```sh
python scripts/import_kotoba.py
python scripts/build_anki_resources.py
```

Importer menggunakan revision tetap, memverifikasi Git blob, mempertahankan cache gzip, dan mengambil file yang belum tersedia. Builder tidak melakukan request jaringan. Setelah hanya mengubah judul/deskripsi/instruksi, jalankan `python scripts/build_anki_resources.py --metadata-only` agar chunk tidak perlu ditulis ulang. Perubahan revision memerlukan audit ulang jumlah/format dan pembaruan ekspektasi tes.

## Validasi

- Checksum arsip seluruh 191 sumber dan validitas seluruh 1.416.940 kartu.
- Scheduling, due-first, isolasi akun, premium, idempotency, stale version dan penolakan placeholder.
- Persistensi MariaDB setelah restart dan kegagalan write tanpa perubahan cache.
- Tes Flutter dari pencarian katalog nyata → detail Wakayama → tampilkan jawaban → rating → selesai, pada layar 320 px dengan teks diperbesar; retry timeout mempertahankan identitas operasi.

## Sumber dan atribusi

[Kotoba / mistval](https://github.com/mistval/kotoba/tree/a0f8f984bf5dd8014527c6608d36b479ab622b07/resources/quiz_data). Lisensi MIT repository disalin utuh. Data turunan tetap menyertakan asalnya dalam JSON sumber dan tautan ke file asal pada detail deck; lisensi repository tidak menggantikan ketentuan sumber pihak ketiga. Contoh sumber yang disebut Kotoba: EDRDG JMdict/JMnedict/ENAMDICT, Japanese WordNet, Wiktionary, JPDB, dan arsip trivia. Media eksternal mengikuti sumber masing-masing.
