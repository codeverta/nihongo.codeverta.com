# Belajar dengan koneksi terbatas

## Perilaku aplikasi

- Respons materi JLPT, detail deck dan sesi Anki yang telah disiapkan, dashboard,
  dan paket latihan disimpan di perangkat. Materi SSW berasal dari resource
  aplikasi; paket latihan SSW disiapkan ketika membuka kurikulum.
- Cache tampil terlebih dahulu tanpa menunggu jaringan. Setelah 24 jam, aplikasi
  memperbaruinya di latar belakang dengan maksimal dua permintaan bersamaan.
  Cache yang lebih lama tetap dapat dibaca ketika offline.
- Membuka course menyiapkan teks lesson beserta paket latihannya. Lesson aktif
  didahulukan. Tombol **Simpan teks** mempertahankan materi dari pembersihan
  otomatis. Anki menyiapkan sesi kartu saat ini, bukan seluruh isi deck.
- Materi yang belum pernah diunduh tetap membutuhkan koneksi pertama. Foto yang
  tidak tersedia ditampilkan sebagai placeholder; audio tidak diunduh massal.
- Kegagalan jaringan, timeout, dan HTTP 5xx tidak menghapus sesi login. Pencabutan
  sesi yang dikonfirmasi melalui HTTP 401 mengakhiri sesi lokal.

## Ruang penyimpanan dan jaringan

Cache materi otomatis dibatasi **100 MiB** (100 × 1024 × 1024 byte).
Teks diprioritaskan; gambar otomatis dibuang lebih dahulu. Materi yang disimpan
secara eksplisit tidak dibuang untuk menerima data baru. Jika ruang yang tersisa
tidak cukup, unduhan baru ditolak dan pengguna dapat membersihkan cache melalui
**Profil → Penyimpanan offline**.

Ilustrasi kecil yang sedang dibuka dapat diunduh melalui data seluler. Media
lebih besar menggunakan Wi-Fi/Ethernet, kecuali pengguna mengaktifkan izin data
seluler pada pengaturan offline. Unduhan otomatis berhenti ketika koneksi hilang.
Manga, avatar, dan konten media besar tidak dimasukkan ke cache ilustrasi publik.

Cache materi terenkripsi pada Android/iOS/desktop. Web menggunakan cache memori,
sehingga materi tidak dijanjikan tetap tersedia setelah browser ditutup.
Antrean latihan dan riwayat privat menggunakan penyimpanan terpisah yang kecil
dan dibatasi; antrean review Anki berada dalam penyimpanan preferensi privat
aplikasi. Batas cache materi bukan batas seluruh ukuran data aplikasi.

## Premium dan pergantian akun

Konten premium yang tersimpan tetap memerlukan akun yang berhak. Bukti offline
ditandatangani backend dan berlaku maksimal 30 hari sejak verifikasi pembelian
store terakhir, bukan 30 hari sejak membaca cache. Perpanjangan memerlukan
verifikasi online. Akun gratis tidak mendapat akses dari cache akun premium.
Refund/pencabutan yang terkonfirmasi ketika online mengunci akses kembali.

Logout menghapus identitas, bukti premium, antrean jawaban/review, riwayat, dan
cache privat akun tersebut. Materi publik yang telah dibersihkan dari progres
akun dapat tetap tersimpan. Jawaban yang belum tersinkronisasi akan hilang saat
logout; dialog logout menjelaskan hal ini.

Konfigurasi kunci dan prosedur build dijelaskan di [offline-access.md](offline-access.md).
Kunci lokal tidak menggantikan konfigurasi deployment: backend produksi harus
menggunakan private key stabil dan build aplikasi harus membawa public pin
pasangannya. CI Android/iOS memerlukan variable `OFFLINE_ACCESS_PUBLIC_KEY`.

## Latihan dan sinkronisasi

Latihan lesson JLPT dan latihan biasa SSW dapat menggunakan paket soal yang sudah
tersimpan. Umpan balik jawaban diberikan lokal. Hasil offline ditandai sementara;
kelulusan, XP, achievement, dan progres resmi baru berubah setelah backend
memverifikasi paket, akun, entitlement, serta jawaban. Retry menggunakan ID
percobaan yang sama agar tidak menggandakan hadiah. Latihan SSW mempertahankan
perilaku tanpa XP dari fitur latihan biasa.

Review Anki menggunakan request ID stabil dan versi kartu. Backend tetap
menentukan jadwal SRS. Konflik versi tidak menimpa jadwal yang lebih baru dari
perangkat lain. Posisi membaca menggunakan timestamp terbaru dan hasil terbaik
tetap dipertahankan.

Simulasi ujian penuh JLPT, JFT, NAT, dan SSW tetap membutuhkan internet. Aplikasi
menjelaskan kebutuhan koneksi tanpa mengarahkan pengguna offline ke login.

## Validasi pengembang

```powershell
flutter analyze --no-pub lib test
flutter test --no-pub
```

Jalankan `go test ./...` dari `backend`. Tes persistensi memerlukan MariaDB
terisolasi melalui `MARIADB_TEST_DSN`; helper tes membuat dan menghapus schema
khusus tes. Jangan arahkan pengujian persistensi ke database produksi.

Perubahan ini menambahkan plugin native untuk penyimpanan aman, direktori data,
dan konektivitas. Instal build baru; hot reload saja tidak memasang plugin native.
