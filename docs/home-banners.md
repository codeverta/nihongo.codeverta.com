# Banner beranda

Beranda memakai carousel dari `GET /v1/home` melalui field `data.banners`.
Tiga banner awal mengarah ke **Materi Belajar**, **SSW / Tokutei Ginou**, dan
**Lowongan Kerja**. Gambar awal menggunakan ilustrasi yang sudah ada di proyek.
Carousel berganti setiap lima detik; pengguna bisa menggeser, memilih indikator,
atau menjeda. Menggeser atau memilih banner menghentikan putaran sampai pengguna
menekan tombol putar. Putaran juga berhenti ketika tab beranda tersembunyi,
halaman lain dibuka, aplikasi masuk latar belakang, atau preferensi perangkat
meminta pengurangan animasi.

Kartu **Kerja di Jepang** kini hanya ada di Beranda. Kartu **SSW / Tokutei Ginou**
ada di Beranda serta Belajar. Keduanya tetap tampil ketika carousel dinonaktifkan.

## Mengubah isi dan gambar

1. Buka `/admin/` pada origin backend dan pilih **Banner Beranda**, atau buka
   `/admin/home-banners` langsung.
2. Masukkan kunci admin backend yang digunakan pada branding.
3. Ubah judul, deskripsi, teks tombol, tujuan, dan gambar. Gunakan tombol panah
   untuk mengatur urutan, sakelar untuk aktif/nonaktif, atau tambah/hapus banner.
4. Gambar bisa memakai URL HTTPS, atau unggah PNG/JPEG maksimal **5 MB** dan
   **12 juta piksel**. Setelah unggah, klik **Simpan semua banner**.
5. Tarik beranda aplikasi ke bawah untuk memuat ulang, atau buka ulang aplikasi.

Tujuan dibatasi pada `learn`, `ssw`, dan `jobs`, sehingga tombol memakai navigasi
aplikasi yang tersedia. Maksimal 12 banner. Menonaktifkan semua banner atau
menyimpan daftar kosong menyembunyikan carousel. Saat gambar gagal dimuat,
aplikasi menampilkan ilustrasi bawaan sesuai tujuan banner.

## Konfigurasi server dan penyimpanan

```dotenv
HOME_BANNERS_CONFIG_PATH=data/home-banners.json
# Opsional; kosong berarti menggunakan BRANDING_ADMIN_TOKEN.
HOME_BANNERS_ADMIN_TOKEN=
```

Konfigurasi disimpan secara atomik di `HOME_BANNERS_CONFIG_PATH`. Gambar unggahan
disimpan pada subfolder `home-banner-images/` di sebelah file konfigurasi,
dengan nama hash isi gambar. Nama berkas dari pengguna tidak dipakai.
Keduanya bertahan setelah restart. Backup direktori tersebut bersama file JSON.
Docker Compose memakai `/app/data` yang sudah dipasang sebagai volume permanen.
Kunci admin tidak pernah dikirim lewat API publik maupun disimpan oleh halaman
admin ke penyimpanan browser. Jika kunci admin kosong, API perubahan dinonaktifkan.

Gambar awal berada di `backend/assets/images/exams/n5.png`,
`backend/assets/images/ssw/ssw-careers.png`, dan
`backend/assets/images/jobs/job-category-atlas.png`. Ketiga jalur ini boleh
diakses melalui handler aset publik. Gambar unggahan disajikan lewat
`/home-banner-images/{hash}.png` atau `.jpg`, dengan tipe konten gambar dan cache
immutable. JSON pengaturan dan berkas lain tidak disajikan lewat jalur itu.

## API admin

Semua endpoint admin memerlukan `Authorization: Bearer <kunci-admin>`.

| Metode | Endpoint | Fungsi |
|---|---|---|
| GET | `/v1/admin/home-banners` | Membaca seluruh banner, termasuk yang nonaktif |
| PUT | `/v1/admin/home-banners` | Menyimpan `{ "banners": [...] }` |
| POST | `/v1/admin/home-banners/images` | Mengunggah multipart field `image`; hasilnya `data.imageUrl` |

Elemen banner memuat `id`, `title`, `description`, `imageUrl`, `action`,
`buttonLabel`, `order`, dan `enabled`. `/v1/home` hanya mengirim banner aktif
dengan urutan naik. Perubahan diterapkan pada permintaan berikutnya tanpa build
ulang aplikasi.

## Target harian

Target harian sebelumnya sudah terhubung ke data gamifikasi: 20 menit per hari,
berdasarkan durasi sesi belajar/ujian yang diselesaikan dan tanggal Asia/Jakarta.
Itu tidak menghitung seluruh waktu membuka atau membaca layar. Tampilan target
di beranda diganti sesuai keputusan produk; data `dailyGoal` tetap tersedia
untuk kompatibilitas klien dan fitur gamifikasi.
