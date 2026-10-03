# JLPT N5 Diagnostic

Buka **Ujian → Buka diagnostic N5 → Mulai diagnostic** setelah login.
Ini latihan original Nihongo Master untuk menentukan fokus belajar, bukan
simulasi lengkap atau penilaian kelulusan resmi JLPT.

## Alur dan tampilan

- Kartu pembuka indigo, ringkasan waktu/jumlah soal, serta panduan pengerjaan.
- 25 soal pilihan ganda: 7 kosakata, 5 kanji, 8 grammar, 5 reading.
- Durasi 30 menit; deadline ditentukan backend. Timer tetap berjalan ketika
  aplikasi di latar belakang atau pengguna meninggalkan sesi.
- Satu jawaban per soal; dapat diubah sebelum dikirim atau deadline berlalu.
- Setiap pilihan disimpan ke backend. Saat penyimpanan gagal, UI tidak
  menandainya sebagai tersimpan dan pengguna diminta memilih ulang.
- Peta nomor soal menunjukkan jawaban tersimpan, posisi aktif, dan tanda
  tinjau. Tanda tinjau hanya berlaku selama halaman sesi masih terbuka.
- Tombol lanjut/mundur dan konfirmasi kirim dengan jumlah soal kosong.
- Saat waktu habis, aplikasi mengirim penyelesaian; jika koneksi gagal,
  tombol **Lihat hasil** dapat dicoba lagi. Backend menolak jawaban baru
  setelah deadline dan tetap menilai jawaban sebelumnya.
- Keluar dapat dilanjutkan lewat kartu sesi aktif. Sesi yang sudah lewat
  deadline akan dinilai saat dibuka kembali, tanpa memperpanjang waktu.

## Hasil

Persentase = jumlah benar / 25 × 100; soal kosong salah, tanpa pengurangan
nilai. Tersedia hasil per bagian, durasi, jumlah jawaban, rekomendasi dua
bagian dengan persentase terendah, tautan ke materi N5, dan pembahasan semua
soal. Filter **Perlu ditinjau** menampilkan jawaban salah/kosong.

Riwayat menampilkan sampai sepuluh diagnostic selesai untuk akun pengguna.
Angka kesiapan dan riwayat contoh dari UI lama sudah diganti data sebenarnya.
Tidak ada prediksi kelulusan resmi atau skor JLPT berskala dalam modul ini.
Bank soal saat ini tetap sama ketika mengulang; pilihan jawaban dirotasi
antarsoal. Kunci jawaban hanya dikirim setelah sesi diselesaikan.

## Implementasi

- Bank soal dan penjelasan: `backend/internal/store/diagnostic_content.go`.
- Sesi, jawaban, deadline, hasil, dan riwayat disimpan di MemoryStore;
  seluruhnya reset saat backend direstart, seperti modul lainnya.
- `POST /v1/exam-attempts` dengan `examId: jlpt-n5-diagnostic` membuat sesi
  atau mengembalikan sesi belum selesai milik akun yang sama. Retry start
  tidak membuat sesi ganda. Pengguna dapat mulai baru setelah penyelesaian.
- `GET /v1/diagnostic/n5`: metadata, jumlah per bagian, sesi aktif, riwayat.
- `GET /v1/exam-attempts/{attemptID}`: soal, jawaban tersimpan, waktu server,
  deadline, dan hasil jika sudah diselesaikan.
- `PUT /v1/exam-attempts/{attemptID}/answers`: `questionId`, `optionId`.
- `POST /v1/exam-attempts/{attemptID}/complete`: hasil; retry mengembalikan
  hasil yang sama tanpa menghitung ulang atau membuat riwayat ganda.
- Semua endpoint sesi memerlukan bearer token dan memeriksa pemilik akun.
- Skor, kunci, dan pengecekan opsi dihitung oleh backend.
- Modul ini tidak memberi XP dan tidak mengubah progres lesson.

Route Flutter: `/exams/n5`, `/exams/n5/attempt/:attemptId`.
Jalankan ulang backend setelah mengubah bank soal.
[Panduan menjalankan project](running-project.md).
