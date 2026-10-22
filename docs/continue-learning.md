# Lanjutkan belajar

Beranda menampilkan satu materi terakhir dan hingga tiga materi sebelumnya, setelah banner dan sebelum Mulai belajar. Cakupan: vocabulary, grammar, kanji, reading. Materi baru dicatat setelah berhasil dimuat; kunjungan yang ditinggalkan saat loading tidak menggantikan riwayat.

Posisi berupa indeks materi (kata/pola/karakter/halaman), terpisah dari LessonProgress.LastPosition yang dipakai untuk durasi kuis. Membuka materi tidak memberi XP, mengubah skor, atau menyelesaikan lesson. Jawaban kuis aktif tidak dipulihkan; tombol membuka kembali bagian belajar. Materi yang sudah lulus tetap tampil sebagai Selesai / Pelajari lagi dan dibuka dari awal.

## Penyimpanan dan sinkronisasi

- GET /v1/me/learning-history dan PUT /v1/me/learning-history/{kind}/{lessonId} memerlukan bearer session akun.
- Metadata, jumlah materi, kelulusan, dan pembatasan premium ditentukan server. Request PUT hanya membawa position dan accessedAt.
- Riwayat tersimpan dalam state akun di tabel gamification_users MariaDB, mengikuti mekanisme persistence, single-writer lock, dan penghapusan akun yang sudah ada. Tidak diperlukan tabel atau migrasi terpisah. Server produksi tetap harus mengaktifkan persistence seperti biasa.
- Flutter memakai SharedPreferences dengan key learning_history_v1_<userId>. Perpindahan materi langsung disimpan lokal; sinkron dikelompokkan dengan jeda 600 ms. Salinan pending bertahan setelah aplikasi ditutup dan dicoba ulang saat beranda di-refresh atau materi dibuka.
- Update dengan waktu akses lebih lama tidak menimpa posisi lebih baru. Balasan server yang terlambat tidak menimpa perpindahan lokal yang lebih baru. Akun lain tidak menerima cache akun sebelumnya.
- Kartu riwayat tetap ditampilkan dari cache saat dashboard gagal dimuat. Ini tidak membuat konten JLPT menjadi offline; konten tetap memerlukan koneksi.
- Jika materi dihapus, riwayat ditandai tidak tersedia dan tombol membuka kursus terkait. Jika entitlement berakhir, tampilkan status terkunci dan arahkan melalui halaman akses premium.

## Validasi

Tes mencakup urutan dan deduplikasi riwayat, isolasi akun, persistensi MariaDB setelah restart, penghapusan akun, validasi posisi/waktu/premium, tidak adanya reward dari navigasi, cache offline, konflik update, pemulihan posisi materi, dan navigasi panel di layar kecil dengan teks diperbesar.
