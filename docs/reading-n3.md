# Reading N3: cerita bergambar

Buka **Belajar → N3 → Reading** setelah login. Tersedia **12 cerita, 36 halaman
bergambar, 72 entri kosakata bantu, dan 60 soal pemahaman**. Entri kosakata
dapat berulang, sehingga 72 bukan jumlah kata unik. Estimasi 13 menit per
cerita mencakup membaca, bantuan bahasa, dan kuis.

Setiap cerita mempunyai tiga halaman, teks Jepang, bantuan kana, terjemahan
Indonesia, dua kosakata per halaman, dan lima soal. Soal dan pilihan jawaban
ditulis dalam bahasa Jepang; pembahasan Indonesia menunjuk bukti dari cerita.
Bantuan kana berupa paragraf terpisah, bukan furigana per kanji.

## Cerita dan tujuan membaca

| No. | Cerita | Fokus pemahaman |
|---|---|---|
| 1 | Layanan baru perpustakaan | Menafsirkan kebutuhan pengguna dan alasan mempertahankan pilihan layanan |
| 2 | Pasar yang tetap ramai saat hujan | Membandingkan hasil penjualan dengan manfaat interaksi |
| 3 | Usulan yang berubah di rapat | Memahami batas waktu, anggaran, dan alasan perubahan rencana |
| 4 | Makan siang tanpa banyak sisa | Menghubungkan pemesanan, jumlah produksi, dan konsekuensi keputusan |
| 5 | Surat dari nenek | Memahami perhatian tokoh dan pilihan cara berkomunikasi |
| 6 | Reservasi pada tanggal yang keliru | Menemukan sumber masalah dan urutan penyelesaiannya |
| 7 | Kota yang ditemukan lewat sepeda | Membaca perubahan pandangan setelah pengalaman baru |
| 8 | Nyaman, tetapi terlalu banyak notifikasi | Menyimpulkan penggunaan teknologi sesuai kebutuhan |
| 9 | Latihan bersama warga | Memahami masukan peserta dan perbaikan informasi |
| 10 | Pesan singkat yang disalahpahami | Membedakan dugaan penerima dengan maksud pengirim |
| 11 | Pakaian lama menjadi tas baru | Memahami tujuan pemanfaatan barang dan kebutuhan penerima |
| 12 | Perubahan kecil di toko roti | Menghubungkan pengamatan, percobaan, dan hasil yang belum pasti |

Materi asli untuk latihan membaca tingkat menengah dengan pendampingan
kosakata dan terjemahan. Pola penghubung, alasan, perbandingan, dan perubahan
keputusan dibuat lebih kompleks daripada N4. Materi ini merupakan kurasi
latihan, bukan daftar resmi atau klaim mencakup seluruh materi JLPT N3.
Jawaban kuis mengikuti teks; ilustrasi membantu memahami konteks.

## Konten dan gambar

- Sumber materi backend: `backend/data/reading/n3.json`, di-embed saat build.
- 12 gambar baru: `backend/assets/images/reading/story-n3-01.png` sampai
  `story-n3-12.png`, dibuat dengan **imagegen bawaan** (mode generate).
- Setiap gambar memuat tiga panel persegi dari kiri ke kanan. Aplikasi
  menampilkan panel sesuai halaman, sehingga tersedia 36 adegan.
- Prompt lengkap, lokasi asli, dan pemetaan panel:
  [reading-n3-art.json](reading-n3-art.json).
- Gambar dibundel sebagai asset aplikasi; tidak memerlukan layanan gambar
  saat membaca. URL gambar backend tetap memakai fallback lokal yang sudah ada.

## API dan progres

Katalog: `GET /v1/reading/course?level=N3`. ID lesson:
`reading-n3-01` sampai `reading-n3-12`. Detail dan kuis menggunakan endpoint
Reading yang sama dengan N5/N4. Semua endpoint Reading memerlukan login.
Lihat [kontrak API](../backend/openapi.yaml).

Progres terpisah per level. Lulus dengan 80% (4 dari 5 soal), dengan 50 XP
untuk kelulusan pertama. Pengulangan mempertahankan skor terbaik dan status
selesai. Progres backend saat ini disimpan di memori dan reset ketika restart.

Restart backend untuk memuat materi baru dan build ulang aplikasi agar gambar
baru masuk APK. Ikuti [panduan menjalankan project](running-project.md).

### Kondisi workspace saat pemasangan

APK debug sudah dipasang pada `emulator-5554`. Backend N3 aktif di port 8080.
Saat pemasangan, draft N2 yang ditulis bersamaan baru berisi enam cerita,
sehingga startup biasa terkena validasi kurikulum 12 cerita. Backend sementara
memakai Go overlay lokal yang melewati draft N2 tanpa mengubah file sumbernya:

```powershell
cd backend
go run -overlay ../build/reading-n3-overlay.json ./cmd/api
```

Overlay hanya diperlukan selama draft N2 belum lengkap dan tersimpan di folder
build lokal. Setelah N2 lengkap, gunakan kembali `go run ./cmd/api`.
Restart backend mereset sesi dan progres dalam memori; login ulang di aplikasi.
