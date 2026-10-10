# Kamus Jepang

Tab **Kamus** tersedia gratis untuk tamu dan pengguna login, termasuk pencarian kosakata semua level. Pembatasan premium pada lesson N3–N1 dan ujian tetap berlaku pada materi belajar; kamus merupakan alat pencarian terpisah.

## Perilaku

- Cari kanji, kana, romaji, atau kata Inggris. Definisi tetap dalam bahasa Inggris sebagaimana diberikan Jisho, tanpa layanan terjemahan berbayar.
- Filter N5–N1 dan kata umum memakai tag pencarian Jisho. Hasil mempertahankan setiap kelompok makna, jenis kata, bacaan/bentuk lain, catatan, kata terkait, dan antonim.
- Pagination 20 hasil per halaman mengikuti Jisho; halaman berikutnya dapat kosong jika halaman sebelumnya berisi tepat 20 hasil. Entri duplikat tidak ditambahkan ulang. Respons pencarian lama tidak menggantikan pencarian baru.
- Ketuk kata untuk membaca detail, salin, simpan favorit, dan lihat kanji yang membentuk kata tersebut. Detail kanji berasal dari kanjiapi.dev: arti, jumlah goresan, onyomi, kunyomi, bacaan nama, grade sekolah, frekuensi koran.
- Label JLPT pada **kanji** memakai sistem lama KANJIDIC. Aplikasi memberi label `JLPT lama`, tidak menyamakannya dengan tag N1–N5 pada kata Jisho.
- Favorit menyimpan entri asli untuk dibaca offline: maksimal 60 kata, terpisah berdasarkan ID akun/tamu di perangkat. Riwayat maksimal 20 query; dapat dihapus. Penghapusan akun melalui aplikasi menghapus library lokal akun tersebut. Favorit belum disinkronkan antardevice. Detail kanji dan pencarian baru memerlukan koneksi.
- API kata Jisho tidak menyediakan seluruh fitur situs. Contoh kalimat lengkap, pencarian radikal, dan diagram goresan tersedia lewat tombol yang membuka halaman terkait di Jisho. Tidak ada contoh/diagram dummy atau janji fitur handwriting yang belum tersedia.

## Sumber gratis

- Kata: `https://jisho.org/api/v1/search/words?keyword=...&page=...`. Tidak memerlukan API key atau langganan berbayar. [Penjelasan API oleh pengelola Jisho](https://jisho.org/forum/54fefc1f6e73340b1f160000-is-there-any-kind-of-search-api), [petunjuk pencarian](https://jisho.org/docs).
- Kanji: `https://kanjiapi.dev/v1/kanji/{character}`. [Dokumentasi dan sumber proyek](https://github.com/onlyskin/kanjiapi.dev).
- Atribusi ditampilkan pada layar kamus/detail/bantuan: Jisho.org, JMdict/JMnedict/KANJIDIC2 milik EDRDG, dan DBpedia jika definisi bersumber dari ensiklopedia. [Lisensi EDRDG](https://www.edrdg.org/edrdg/licence.html), [sumber Jisho](https://jisho.org/about). Data tidak menjadi milik aplikasi dan tidak dilisensikan ulang sebagai konten proprietary.

Kedua sumber publik tidak menjamin SLA. Tidak ada biaya API maupun secret tambahan; hosting backend dan jaringan tetap memakai infrastruktur aplikasi yang ada. Jika upstream error/timeout/rate limit, aplikasi menampilkan retry dan tautan Jisho, tanpa mengarang hasil atau menganggap error sebagai hasil kosong.

## Backend

- `GET /v1/dictionary/search?q=猫&page=1&jlpt=N5&common=true`: publik; envelope `data` berisi `query`, `page`, `hasMore`, `entries`, `source`.
- `GET /v1/dictionary/kanji/猫`: publik; `data` berisi objek kanjiapi.dev.
- `400`: query invalid, lebih dari 100 karakter, page di luar 1–100, filter invalid, atau path kanji bukan satu karakter Han.
- `404`: kanji tidak tersedia. `503`: sumber sibuk/offline/invalid JSON.

Flutter memanggil backend aplikasi; backend hanya mengakses host HTTPS tetap Jisho dan kanjiapi.dev. Token login tidak diteruskan ke sumber. Ukuran satu respons maksimal 1 MiB. Pencarian serupa yang sedang berjalan digabung. Maksimal dua request upstream aktif, diberi jarak minimal 500 ms; permintaan lain mendapat retry. HTTP upstream timeout delapan detik. Tidak ada scrape HTML maupun proxy URL arbitrer. Pastikan egress HTTPS server mengizinkan kedua host. Untuk deployment besar gunakan rate limit pada reverse proxy dan evaluasi hosting indeks JMdict sendiri sesuai lisensi sumber.

## Penyimpanan database bersama

Backend saat ini menggunakan **SQLite**, bukan MySQL/MariaDB. Startup menginisialisasi tabel kamus pada koneksi yang sama dengan akun, billing, dan gamifikasi, di lokasi `GAMIFICATION_DB_PATH` (default `backend/data/gamification.db` ketika menjalankan API dari folder backend; Docker `/app/data/gamification.db`). Tidak memerlukan API key atau konfigurasi DB tambahan.

Alur pencarian: cache memori (maksimal 128 respons) → cache SQLite → API eksternal bila belum ada atau kedaluwarsa. Cache produksi berlaku 30 hari dan tetap dapat dipakai setelah restart. Berlaku bersama untuk seluruh pengguna/tamu; tidak diduplikasi per akun. Query, halaman, filter JLPT, dan filter kata umum menjadi identitas cache berbeda agar hasil tidak tercampur. Client test tanpa persistence tetap memakai cache memori enam jam.

- `dictionary_responses`: respons JSON asli yang tervalidasi, sumber, waktu fetch, dan SHA-256 identitas request. Tidak menyimpan ID user maupun teks query mentah. Maksimal 5.000 kombinasi request; yang paling lama dihapus jika kapasitas terlampaui. Hasil kosong yang sah juga disimpan.
- `dictionary_entries`: arsip setiap entri kata Jisho dan detail kanji, unik berdasarkan `(source, entry_id)`. JSON lengkap mempertahankan makna, bacaan, tag dan atribusi sumber. Query berbeda yang menghasilkan kata yang sama memperbarui satu entri. Arsip ini tetap disimpan saat response cache kedaluwarsa/dihapus dan saat akun pengguna dihapus.

Respons HTTP error, JSON invalid, atau payload yang tidak sesuai schema tidak dimasukkan database. Respons dan arsip entri ditulis dalam satu transaksi: kegagalan salah satu membatalkan semuanya. Jika DB read/write gagal, endpoint mengembalikan `503`; kegagalan baca tidak menyebabkan backend terus memanggil sumber eksternal. Setelah 30 hari, satu permintaan melakukan refresh ke sumber; request serupa yang sedang berjalan menunggu hasil yang sama. Jika refresh gagal, respons lama dan arsip tetap berada di DB, tetapi respons yang kedaluwarsa belum disajikan sebagai data baru.

Cache menyalin hasil berdasarkan query yang benar-benar dicari, bukan melakukan bulk crawling. Arsip data tetap mengikuti atribusi/lisensi EDRDG/DBpedia; penyimpanan tidak memindahkan kepemilikan data ke aplikasi. Backup SQLite/volume harus mencakup tabel kamus; pantau ukuran arsip dan disk karena entri dikumpulkan secara bertahap. Setup memakai satu proses API per file DB dan tidak memiliki adapter MySQL/MariaDB.

Route Flutter `/dictionary?q=...` juga bisa membuka pencarian tertentu secara langsung. UI tetap mengikuti tema Nihongo Master; tidak menyalin logo/branding Jisho.

## Pemeriksaan

Tes backend memakai HTTP transport fixture untuk memverifikasi encoding, filter, cache, pembatalan request yang menunggu, status upstream, validasi kanji, dan endpoint publik. Tes persistence memakai DB SQLite nyata sementara: restart tanpa fetch ulang, deduplikasi arsip kata/kanji, refresh setelah TTL, pemisahan page/filter, penolakan payload invalid, rollback ketika penulisan arsip gagal, serta integrasi dengan koneksi DB yang sudah dipakai akun/billing/gamifikasi. Tes Flutter memverifikasi race pencarian, deduplikasi pagination, retry, persistensi/isolation favorit, layar ponsel, serta akses kamus sebagai tamu.

Smoke test sumber asli bersifat opt-in (memerlukan internet): jalankan `go test ./internal/dictionary -run TestLiveDictionarySources -v` dengan environment `DICTIONARY_LIVE_TEST=1`.
