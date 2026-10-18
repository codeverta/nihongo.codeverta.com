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

- `GET /v1/dictionary/search?q=猫&mode=auto&page=1&jlpt=N5&common=true`: publik; envelope `data` berisi `query`, `page`, `hasMore`, `entries`, `source`.
- `GET /v1/dictionary/kanji/猫`: publik; `data` berisi objek kanjiapi.dev.
- `400`: query invalid, lebih dari 100 karakter, page di luar 1–100, filter invalid, atau path kanji bukan satu karakter Han.
- `404`: kanji tidak tersedia. `503`: sumber sibuk/offline/invalid JSON.

Flutter memanggil backend aplikasi; backend hanya mengakses host HTTPS tetap Jisho dan kanjiapi.dev. Token login tidak diteruskan ke sumber. Ukuran satu respons maksimal 1 MiB. Pencarian serupa yang sedang berjalan digabung. Maksimal dua request upstream aktif, diberi jarak minimal 500 ms; permintaan lain mendapat retry. HTTP upstream timeout delapan detik. Tidak ada scrape HTML maupun proxy URL arbitrer. Pastikan egress HTTPS server mengizinkan kedua host. Untuk deployment besar gunakan rate limit pada reverse proxy dan evaluasi hosting indeks JMdict sendiri sesuai lisensi sumber.

## Penyimpanan database bersama

Backend memakai **MariaDB**. Startup menginisialisasi tabel kamus pada database
yang sama dengan akun, billing, dan gamifikasi. Koneksi menggunakan `DB_*`
dari `backend/.env`, dengan port default 3306. Lihat [setup MariaDB](mariadb.md).

Alur pencarian: cache memori → MariaDB → API eksternal hanya jika identitas pencarian belum ada. Hasil berisi kata **dan hasil kosong** disimpan permanen, tanpa TTL 30 hari atau penghapusan otomatis setelah 5.000 respons. Memori dibatasi 128 respons sekaligus dan total 16 MiB. Setelah memori terbuang atau server restart, lookup berikutnya membaca MariaDB. Error jaringan, HTTP error, dan payload invalid tidak dianggap hasil kosong.

### Mode dan normalisasi

- `mode=auto` (default): kana/kanji memakai jalur Jepang; romaji dalam daftar padanan yang sudah ditinjau (contoh `konnichiwa`, `neko`, `gakkou`) juga memakai jalur Jepang. Kata Latin lainnya mempertahankan pencarian sumber, agar kata ambigu seperti `pan` tidak dipaksa menjadi Jepang. Pilih mode Jepang untuk romaji di luar daftar tersebut.
- `mode=japanese`: romaji ditransliterasi dengan tabel suku kata; mencakup Hepburn/Kunrei umum, konsonan ganda, `matcha`, apostrof setelah n, dan penanda vokal panjang. Huruf besar, komposisi Unicode, lebar karakter, hiragana dan katakana dinormalisasi sebelum fetch pertama. Ejaan greeting `konnichiwa`/`konnichiha`/`こんにちわ` menjadi `こんにちは`.
- `mode=english`: pencarian dikutip agar Jisho tidak otomatis mengubah huruf Latin menjadi romaji. Tanda kutip dan backslash dalam input mode Inggris ditolak. Operator pencarian Jisho pada Auto/Jepang dipertahankan sebagai ekspresi mentah dengan identitas terpisah.
- Auto dan Jepang berbagi cache jika menghasilkan query Jepang yang sama. Mode Inggris dan ekspresi mentah memiliki identitas berbeda. Halaman, filter JLPT, dan kata umum tetap bagian identitas. Identitas baru diberi versi `v2`.
- Kanji `橋` dan `箸` tetap berbeda. Tidak ada pemetaan seluruh bacaan setiap hasil ke respons pencarian: hasil luas tidak boleh menjadi jawaban salah untuk pencarian kata tertentu. Vokal panjang `こう`, `こお`, `こー` tidak disamakan secara umum. Macron disimpan sebagai penanda panjang `ー`; alias tempat yang ditinjau seperti `Tōkyō`/`toukyou` memiliki ejaan khusus `とうきょう`.

`dictionary_responses` menyimpan JSON tervalidasi, identitas hash, endpoint tetap, query kanonis, jenis pencarian efektif, halaman/filter, dan waktu fetch. Tidak ada ID pengguna. `dictionary_entries` tetap mengarsipkan entri unik `(source, source_entry_id)` lengkap dengan atribusi. Penulisan respons dan arsip berlangsung dalam satu transaksi. Kegagalan DB tidak menyebabkan bypass ke API.

Cache lama tetap dipertahankan. Saat query kanonis mempunyai URL upstream yang persis sama dengan cache lama, respons lama dapat diadopsi tanpa fetch. Baris lama yang hanya memiliki hash tidak bisa dikonversi menjadi alias secara massal; query yang URL kanonisnya berbeda dapat membutuhkan fetch awal baru. Baris tanpa metadata ditandai pada admin dan belum dapat di-refresh sampai query diketahui melalui lookup baru.

### Admin Cache Kamus

Buka `/admin/dictionary` dari menu admin. Kunci memakai `DICTIONARY_ADMIN_TOKEN`, atau fallback `BRANDING_ADMIN_TOKEN`. Kunci tetap di memori browser dan dapat dikunci kembali.

- `GET /v1/admin/dictionary?page=1&q=...`: bearer admin; daftar 50 respons per halaman, total, query/mode/filter/waktu fetch, dan statistik proses (`memoryHits`, `databaseHits`, `upstreamCalls`, `memoryBytes`, `memoryLimitBytes`). Daftar tidak memuat blob respons.
- `POST /v1/admin/dictionary/{id}/refresh`: bearer admin; fetch dan validasi dahulu, commit database, lalu mengganti cache memori. Refresh yang gagal mempertahankan data lama. Request refresh memakai throttle upstream yang sama dengan pencarian publik.
- Request dengan identitas kanonis sama digabung, termasuk variasi ejaan. Pembatalan satu pengguna tidak membatalkan fetch bersama. Pekerjaan bersama dibatasi 12 detik, dan timeout HTTP upstream tetap delapan detik. Refresh dan lookup biasa dikoordinasikan agar hasil lama tidak menimpa hasil refresh.

Backup MariaDB perlu mencakup tabel kamus. Penyimpanan tetap mengikuti lisensi sumber. Lihat hasil pengukuran dan batas cakupan di [laporan performa](dictionary-performance.md).

Route Flutter `/dictionary?q=...` juga bisa membuka pencarian tertentu secara langsung. UI tetap mengikuti tema Nihongo Master; tidak menyalin logo/branding Jisho.

## Pemeriksaan

Tes backend memakai HTTP transport fixture untuk memverifikasi encoding, filter, cache, pembatalan request yang menunggu, status upstream, validasi kanji, dan endpoint publik. Tes persistence memakai schema MariaDB nyata sementara: restart dan variasi romaji/kana tanpa fetch ulang, hasil lama/kosong permanen, retensi di atas 5.000 respons, refresh manual atomik, pemisahan mode/page/filter/homofon, penolakan payload invalid, rollback ketika penulisan arsip gagal, serta integrasi dengan koneksi DB yang sudah dipakai akun/billing/gamifikasi. Tes Flutter memverifikasi race pencarian, deduplikasi pagination, retry, persistensi/isolation favorit, layar ponsel, serta akses kamus sebagai tamu.

Smoke test sumber asli bersifat opt-in (memerlukan internet): jalankan `go test ./internal/dictionary -run TestLiveDictionarySources -v` dengan environment `DICTIONARY_LIVE_TEST=1`.
