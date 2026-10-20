# Berita Jepang

News tersedia gratis termasuk untuk guest. Beranda menampilkan tiga artikel
terbaru; `/news` membuka daftar dan `/news/:id` membuka reader. Teks penerbit
dipertahankan, bukan terjemahan atau penyederhanaan otomatis.

## Sumber dan atribusi

| Sumber | Jenis | Endpoint | Ketentuan |
|---|---|---|---|
| Global Voices 日本語 | Berita dunia, bahasa reguler | https://jp.globalvoices.org/feed/ | [CC BY 3.0 dan aturan penayangan ulang](https://jp.globalvoices.org/about/記事の転載について/): kredit penerbit, penulis, penerjemah, tautan artikel dan lisensi. |
| Osaka City | Info kehidupan Jepang, kategori easy | https://www.city.osaka.lg.jp/shisei/category/3054-1-2-21-3-4-0-0-0-0.html | [Kebijakan Osaka](https://www.city.osaka.lg.jp/main/site_policy/0000000124.html): teks HTML dapat dipakai dengan atribusi dan pengungkapan perubahan; gambar/lampiran tidak otomatis tercakup. |

Global Voices asli dengan kredit penulis/penerjemah diambil dari RSS resmi.
Artikel yang terdeteksi mencantumkan perjanjian penayangan ulang pihak ketiga
atau kredit partner tertentu dilewati dan dihitung pada status sumber. Ini
penyaringan konservatif berbasis metadata/penanda penerbit, bukan pemeriksaan
lisensi otomatis yang menjamin semua pengecualian terdeteksi. Periksa status
dan ketentuan sumber saat menambahkan sumber atau memperbarui parser.

Parser memisahkan caption/media dari teks bacaan. Saat sinkronisasi, gambar
artikel dalam caption dengan lisensi individual CC BY, CC BY-SA, CC0 atau public
domain diunduh ke backend. Lisensi NC/ND, fair use, dan izin khusus penerbit tidak
diimpor otomatis. Host gambar dibatasi ke penyimpanan penerbit yang dikenal,
ukuran maksimal 5 MiB/4096 piksel; PNG/JPEG/WebP dikonversi ke JPEG maksimal 1280
piksel, metadata dibuang. Kredit caption, tautan sumber, dan lisensi disimpan di
DB serta ditampilkan di reader. Gagal mengunduh gambar tidak menggagalkan teks.
Gambar editorial yang sudah disimpan tetap dipertahankan saat sinkronisasi.
Kartu tanpa gambar berizin tampil sebagai kartu teks, tanpa ikon placeholder.
Gambar PNG/JPEG milik sendiri atau berizin tetap bisa diunggah melalui admin.
Semua gambar dilayani dari volume `data/news-images`, tanpa hotlink penerbit.
Untuk mengisi gambar pada artikel lama jalankan `go run ./cmd/news-sync -force`
dari direktori backend, atau fetch manual melalui admin.

NHK Easy, Slow Communication, dan Nishinippon tidak diaktifkan: izin menyimpan
dan menampilkan ulang teks penuhnya belum terkonfirmasi. Osaka adalah informasi
publik dengan pembaruan tidak teratur; Global Voices bukan easy Japanese
bawaan. Jadwal pemeriksaan tidak menjamin ada artikel baru setiap pemeriksaan.

## Penyimpanan dan scheduler

- Tabel MariaDB/GORM `news_articles` dan `news_sources` dibuat saat startup.
- Global Voices: Selasa/Jumat 18.00 WIB; Osaka: Minggu/Kamis 18.00 WIB.
  Hari/jam/zona IANA/aktif dapat diubah di admin; maksimal dua hari per minggu.
- Worker memeriksa deadline pada startup dan setiap menit. Deadline berikutnya
  dihitung sebagai waktu kalender pada zona sumber, lalu disimpan UTC. Banyak
  jadwal terlewat saat server mati menghasilkan satu pemeriksaan saat aktif lagi.
- Lease DB 10 menit dan transaksi row lock mencegah dua replica mengimpor sumber
  yang sama bersamaan. Batas waktu pekerjaan delapan menit. Deadline ditulis
  sebelum jaringan; kegagalan tetap menyimpan jadwal berikutnya dan error.
  Crash ketika request sudah dimulai tidak diulang tanpa batas saat restart:
  tunggu jadwal berikutnya atau jalankan fetch manual setelah lease kedaluwarsa.
- Pengisian pertama dan setiap batch dibatasi maksimal 20 artikel yang terbit
  dalam enam bulan terakhir. Feed hanya dapat menyediakan jendela artikelnya;
  batas 20 bukan jaminan tepat 20 artikel. Tidak ada konten dummy atau tanggal baru
  buatan ketika penerbit belum menambah artikel.
- Request Global Voices menggunakan ETag/Last-Modified. Fetch manual melewati
  validator. Osaka memeriksa ulang halaman HTML yang masih ada di jendela indeks
  karena koreksi isi belum tentu mengubah validator indeks.
- Respons sumber maksimal 8 MiB, artikel maksimal 60.000 karakter/400 paragraf.
  Batch di-commit atomik; fetch atau validasi gagal mempertahankan data sebelumnya.
- URL unik mencegah duplikat. Data permanen tanpa TTL/pruning otomatis.
  Koreksi sumber diperbarui ketika artikel masih tersedia dalam jendela fetch.
  Status hidden/gambar editorial tetap dipertahankan; koreksi furigana untuk
  judul/paragraf yang sama tetap dipakai. Paragraf berubah dianalisis ulang.
  Koreksi token memakai revision artikel untuk menolak posisi token yang sudah
  berubah selama editor terbuka; admin perlu memuat ulang sebelum menyimpan.
  Hapus membuat tombstone URL dan mengosongkan teks/token agar tidak terbit ulang.
- `GET /v1/news` dan `GET /v1/news/{id}` hanya membaca DB; tidak ada fallback
  ke situs sumber. Pull-to-refresh aplikasi juga hanya membaca backend.

`NEWS_SYNC_ENABLED=false` menghentikan worker otomatis. Endpoint baca DB dan
fetch manual admin tetap tersedia. Kunci admin: `NEWS_ADMIN_TOKEN`, lalu
`BRANDING_ADMIN_TOKEN`, lalu `DICTIONARY_ADMIN_TOKEN`. Kunci kosong menolak
semua tindakan admin. Kunci tidak ditanam di aplikasi atau disimpan di browser.

Untuk menjalankan satu kali pemeriksaan yang sudah jatuh tempo:

```powershell
cd backend
go run ./cmd/news-sync
```

API production harus tetap berjalan dan volume `/app/data` harus persisten
untuk gambar. Pengaturan, artikel, status, serta lease persisten di MariaDB.
Scheduler tidak menggunakan automation Codex/chat atau cron perangkat pengguna.

## Furigana dan kamus

Kagome v2.9.11 + IPADIC berjalan lokal di Go, tanpa API AI berbayar atau CGO.
Dictionary dimuat satu kali; seluruh analyzer hanya berada di backend. Notice
Kagome, wrapper dictionary, dan data IPADIC berada di `backend/third_party/news`.

Ruby asli sumber diutamakan. Token menyimpan teks, bentuk dasar, posisi UTF-16,
bacaan hiragana, dan `readingSource` (`source`, `automatic`, `editorial`,
`unknown`). Offset selalu mengacu teks asli, termasuk surrogate pairs seperti
emoji/kanji tambahan. Bacaan otomatis, terutama nama, tidak dijamin tepat.
Token tidak dikenal tidak ditebak; admin dapat memperbaiki atau mengosongkan
bacaan. Furigana default aktif dan preferensi toggle disimpan di perangkat.

Flutter mempertahankan `SelectableText.rich` berisi teks asli. Ruby digambar
di atas selection box native, bukan disisipkan sebagai WidgetSpan atau karakter
teks. Seleksi dan salin menggunakan teks asli. Pilih teks lalu **Cari arti**:
satu kata langsung mencari bentuk dasarnya; kalimat menampilkan daftar kata
unik untuk dipilih satu per satu. Hanya kata yang dipilih dikirim ke repository
Jisho existing dengan `mode=japanese`, menggunakan cache bersama permanen yang
sudah tersedia. Jisho tidak dipakai sebagai penerjemah kalimat. Fitur ini berlaku
untuk News; materi Reading JLPT yang ada tidak memakai renderer baru ini.

## API dan admin

Halaman `/admin/news` memakai key backend. Menu mencakup jadwal/status,
fetch manual, daftar semua artikel, koreksi token judul/paragraf, hidden/delete,
serta unggah gambar. Fetch manual mengembalikan 202 dan berjalan di background;
muat ulang status untuk melihat hasil. Request bersamaan dilindungi lease,
dan fetch manual sumber yang sama memiliki cooldown satu menit.

- `GET /v1/news?page=1&limit=20&source=global-voices|osaka`
- `GET /v1/news/{id}`
- `GET /v1/admin/news`, `GET/PUT/DELETE /v1/admin/news/{id}`
- `PUT /v1/admin/news/sources/{id}`
- `POST /v1/admin/news/sources/{id}/fetch`
- `POST /v1/admin/news/images` (multipart `file`, PNG/JPEG, maksimal 5 MiB)
- `GET /news-images/{hash}.png|jpg`

List diurutkan tanggal terbit asli, lalu ID; maksimum 50 item per halaman.
`originalDate` mempertahankan tanggal kalender Jepang; tanggal sumber Osaka
tidak digeser ke hari sebelumnya saat timestamp tengah malam JST dilihat di WIB.
Osaka dilabeli “Tanggal sumber” karena indeks tidak membedakan terbit/koreksi.
List tidak mengirim isi paragraf, detail mengirim teks dan token. Guest tidak
dapat melihat artikel hidden/deleted atau pesan error internal sumber.

## Verifikasi

```powershell
cd backend
go test ./internal/news ./internal/api -count=1
# Opsional, menghubungi kedua penerbit sungguhan:
$env:NEWS_LIVE_TEST='1'
go test ./internal/news -run TestLiveSources -count=1 -v
Remove-Item Env:NEWS_LIVE_TEST
```

```powershell
flutter test --no-pub test/features/news/news_test.dart
flutter analyze --no-pub lib/features/news
```

Tes DB memakai schema MariaDB sementara dari `internal/testdb`, bukan tabel
akun/artikel aplikasi. Tes menguji jadwal WIB, offset UTF-16, ruby sumber,
penyaringan media/partner, restart/dedup/lease, batch atomik, koreksi editorial,
tombstone, endpoint guest/admin, dan unggah gambar. Tes Flutter memeriksa
seleksi/salin lintas baris dengan skala font besar, bentuk dasar Jisho, seleksi
kalimat tanpa request massal, preferensi furigana, dan navigasi guest.
