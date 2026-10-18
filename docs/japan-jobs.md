# Lowongan Jepang

Fitur Lowongan ada di halaman Belajar, setelah materi JLPT, Anki, dan SSW.
Katalog, detail, pencarian, pagination, dan filter membaca MariaDB. Pencarian
pengguna tidak memanggil situs eksternal. Setiap kartu dan detail menampilkan
sumber serta waktu pemeriksaan; tombol melamar membuka situs sumber.

## Data dan jadwal

Backend membuat tabel `job_listings`, `job_observations`, `job_sources`,
`job_sync_domains`, dan `job_sync_runs`. Data akun dan pembayaran tetap memakai
tabel yang sudah ada. Tidak ada lowongan contoh yang dimasukkan ke production.

```dotenv
JOBS_SYNC_ENABLED=true
JOBS_SYNC_INTERVAL_HOURS=168
```

`168` berarti sekali setiap tujuh hari. `84` berarti dua kali setiap tujuh hari.
Nilai lain ditolak. Worker berjalan bersama API, memeriksa deadline database
setiap 30 menit. Deadline berikutnya ditulis **sebelum** akses jaringan.
Lease per domain mencegah dua worker atau restart memicu fetch ganda. Bila
backend mati saat deadline tiba, sinkronisasi dilakukan pada startup berikutnya.
Worker harus berjalan di server agar jadwal tetap berjalan; tidak bergantung
pada aplikasi pengguna atau automation Codex.

Setiap domain dibatasi empat request per batch. Feed perusahaan memakai satu
request per sumber. Adaptor HTML memakai satu robots.txt, satu halaman daftar,
dan maksimal dua detail dengan jeda tiga detik. Tidak ada retry langsung,
pagination tak terbatas, browser automation, login, atau penyelesaian captcha.
Redirect dan pembatasan HTTP menghentikan akses; Retry-After memperpanjang
deadline jika melebihi jadwal normal. Halaman yang tidak dapat diurai dilaporkan
sebagai tidak didukung, bukan sinkronisasi kosong yang dianggap berhasil.

## Cakupan sumber

- Feed resmi perusahaan: Mujin dan Woven by Toyota melalui Lever; Accor Jepang
  melalui SmartRecruiters. Initial capture berisi 167 lowongan nyata.
  Setelah sinkronisasi pertama adaptor board, katalog lokal berisi 177
  lowongan (18 Oktober 2026). Angka ini akan mengikuti hasil sumber berikutnya.
- Adaptor publik terbatas: TokyoDev, Japan Dev, Jobs in Japan, Findy, LAPRAS,
  HirePlanner, dan Agri Navi. Hanya `schema.org/JobPosting` yang valid diambil.
  Keberadaan adaptor tidak menjamin setiap situs menyediakan data yang cocok;
  hasil dan pembatasannya terlihat pada halaman Sumber & pembaruan.
- Direktori memuat 57 situs dari daftar pengguna, termasuk portal pekerjaan,
  agen, layanan matching, dan pusat dukungan pemerintah. Entri direktori bukan
  klaim bahwa semua situs sedang disinkronkan atau mempunyai lowongan aktif.
  LinkedIn, Indeed, CareerCross, serta situs yang melarang crawler atau
  memerlukan login tetap berupa tautan sumber.
- TOKUTEI JOB MATCHING menampilkan tautan lowongan, tetapi satu detail yang
  diperiksa mengalihkan akses. Data privatnya tidak diambil. Layanan matching
  seperti Nakayoshi juga tetap menjadi direktori.

Audit per situs: [IT dan pekerjaan umum](job-board-it-audit.md),
[SSW dan industri](job-board-ssw-audit.md), serta
[riset sumber](job-sources-research.md). Direktori dibundel di
`backend/internal/jobs/resources/sources.json`.

## Integritas katalog

ID eksternal per sumber dan URL lamaran kanonis mencegah duplikasi link yang
sama. Parameter pelacakan dibuang. Nilai gaji hanya dibandingkan pada mata uang
JPY dan periode yang sama; nominal tidak dikonversi atau diannualisasi secara
asumsi. Visa, SSW, bahasa Jepang, gaji, serta kualifikasi yang tidak disebutkan
tetap tidak diketahui. Kategori industri tidak membuktikan kelayakan visa SSW.

Hasil lengkap yang berhasil dapat menutup lowongan yang hilang. Hasil parsial,
error, dan sampel HTML tidak menutup lowongan lain. Lowongan yang kedaluwarsa
atau tidak terlihat selama 30 hari tidak ditampilkan sebagai aktif. Detail lama
tetap menjelaskan statusnya dan tombol melamar dinonaktifkan.

## Operasional

Dari folder `backend`, dengan konfigurasi MariaDB yang sama dengan API:

```sh
go run ./cmd/jobs-sync --status
go run ./cmd/jobs-sync
```

`--status` hanya membaca database. Perintah sinkronisasi tetap mematuhi deadline
mingguan; tidak ada opsi force jaringan. Worker API dan CLI berbagi lease yang
sama. Log menyimpan hitungan dan status, tidak menyimpan kredensial database.

`--snapshot path.json` dapat mengimpor ulang capture nyata yang berumur maksimal
24 jam tanpa HTTP. Waktu observasi asli dipertahankan dan deadline tidak
dimajukan lebih cepat. Capture yang lebih tua dari data tersimpan hanya mengisi
metadata kategori/lokasi yang sebelumnya tidak diketahui pada record identik;
tidak mengubah deskripsi atau status lowongan terbaru.

Endpoint publik:

| Endpoint | Hasil |
|---|---|
| `GET /v1/jobs/meta` | Kategori, 47 prefektur, jenis kerja, opsi bahasa/visa |
| `GET /v1/jobs` | Katalog dengan cursor dan filter |
| `GET /v1/jobs/{id}` | Detail dan atribusi sumber |
| `GET /v1/jobs/sources` | Direktori, status, waktu sukses dan jadwal berikutnya |

Filter: `q`, `category`, `sector`, `prefecture` (kode `01`–`47`),
`employmentType`, `japaneseLevel`, `visa`, `remote`, `salaryMin`,
`salaryPeriod`, `cursor`, dan `limit` (maksimal 30).
