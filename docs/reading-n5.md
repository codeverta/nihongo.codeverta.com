# Reading N5: cerita bergambar

Buka **Belajar → N5 → Reading** setelah login. Modul gratis ini berisi
**12 cerita, 36 adegan bergambar, 72 entri kosakata bantu, dan 60 soal**.
Setiap cerita mempunyai tiga halaman dan lima soal pemahaman.

## Pengalaman membaca

- Setiap adegan memiliki ilustrasi tersendiri sesuai isi paragraf.
- Teks Jepang ditampilkan dengan ukuran besar dan jarak baris yang lapang.
- Tombol **Bantuan kana** membuka cara baca seluruh paragraf.
- Tombol **Terjemahan** membuka arti bahasa Indonesia. Bantuan dapat
  disembunyikan lagi saat berlatih membaca tanpa terjemahan.
- Dua kosakata bantu tersedia pada setiap halaman. Entri dapat berulang
  antarcerita; jumlah 72 bukan jumlah kosakata unik.
- Navigasi maju/mundur berada di bagian bawah dan tetap terlihat.
- Katalog menampilkan sampul bergambar, judul Jepang, tema, estimasi
  waktu belajar, dan skor terbaik. Pencarian menerima judul Indonesia/Jepang.
- Saat kuis, pengguna dapat membuka **Baca cerita lagi** untuk meninjau
  seluruh teks Jepang. Jawaban didasarkan pada teks, bukan jumlah benda
  atau detail kecil yang mungkin berbeda dalam ilustrasi.

Bantuan kana berupa paragraf terpisah, bukan furigana di atas tiap kanji.
Estimasi menit pada katalog adalah perkiraan waktu belajar, bukan timer.

## Cerita

| Lesson | Judul | Fokus |
|---|---|---|
| 1 | Pagi bersama Hana | Rutinitas, waktu, sarapan, perjalanan sekolah |
| 2 | Sehari di sekolah | Kelas, pelajaran, guru, makan siang |
| 3 | Belanja bersama ibu | Buah, jumlah benda, urutan kegiatan |
| 4 | Janji di kafe | Janji, letak tempat, makanan dan minuman |
| 5 | Piknik di taman | Keluarga, posisi, hewan, kegiatan |
| 6 | Naik kereta bersama Kenta | Tiket, transportasi, durasi, tujuan |
| 7 | Hari hujan dan payung merah | Cuaca, alasan, kegiatan di rumah |
| 8 | Buku untuk dibaca | Perpustakaan, penghitung buku, keinginan |
| 9 | Ulang tahun Yui | Usia, hadiah, memberi, waktu |
| 10 | Kucing kecil bernama Momo | Sifat, nama, urutan, perasaan |
| 11 | Perjalanan ke pantai | Liburan, keluarga, alasan, perjalanan |
| 12 | Akhir pekan yang menyenangkan | Urutan, membantu keluarga, rencana |

Cerita, terjemahan, dan soal ditulis untuk aplikasi ini. Materi memakai
pola dasar N5 dan menyediakan bantuan untuk kosakata tambahan. Ini adalah
kurikulum latihan aplikasi, bukan salinan atau checklist resmi ujian JLPT.

## Ilustrasi AI

Ilustrasi dibuat dengan **tool imagegen bawaan**, bukan CLI. Seluruh hasil
akhir disalin ke `backend/assets/images/reading/story-01.png` sampai `story-12.png`.
Setiap berkas merupakan tiga panel persegi berjajar; widget
`ReadingIllustration` menampilkan panel sesuai halaman tanpa membuat salinan
gambar. Totalnya 12 lembar ilustrasi dengan **36 adegan berbeda**.

Gaya: watercolor/gouache hangat, lingkungan keseharian Jepang, karakter
berulang Hana dan Yui, tanpa judul atau balon percakapan di dalam gambar.
Prompt lengkap, pemetaan berkas, dan mode pembuatan dicatat dalam
[manifest ilustrasi](reading-n5-art.json). Ilustrasi telah diperiksa untuk
kesesuaian adegan, tetapi detail kecil hasil AI dapat bervariasi.

Gambar dibundel agar materi visual tampil saat layanan gambar jarak jauh
tidak tersedia. Backend mengirim `imageAsset`, `imagePanel`, dan `imageUrl`.
Jika `COS_BUCKET_URL` dikonfigurasi, URL dibentuk dari asset key yang sama;
aplikasi tetap memakai gambar bundel ketika gambar remote belum diunggah
atau gagal dimuat. Implementasi ini tidak otomatis mengunggah gambar ke COS.

## Backend dan progres

Teks, bantuan baca, kosakata, metadata gambar, soal, dan kunci jawaban berada
di `backend/data/reading/n5.json`, di-embed oleh Go saat build. Kunci jawaban
tidak disertakan dalam respons awal kuis. Flutter tidak menyimpan naskah
cerita atau kunci jawaban di source code.

- Lulus minimal 80% (empat dari lima soal).
- Lulus pertama kali memberi 50 XP; pengulangan tidak menggandakan XP.
- Skor terbaik dan status selesai dipertahankan saat percobaan berikutnya gagal.
- Retry jaringan pada jawaban atau penyelesaian tidak menggandakan penilaian.
- Endpoint progres umum menolak ID lesson Reading; backend menentukan hasil.
- Progres masih disimpan di memori dan reset ketika backend direstart.
- Perubahan naskah memerlukan restart backend; perubahan asset memerlukan
  build ulang Flutter agar asset masuk ke APK.

Semua endpoint memakai bearer token:

| Method | Endpoint |
|---|---|
| GET | `/v1/reading/course` |
| GET | `/v1/reading/lessons/{lessonID}` |
| POST | `/v1/reading/lessons/{lessonID}/attempts` |
| POST | `/v1/reading/attempts/{attemptID}/answers` |
| POST | `/v1/reading/attempts/{attemptID}/complete` |

Endpoint jawaban menerima `questionId` dan `optionId`. Kontrak lengkap ada
di `backend/openapi.yaml`. Listening dan Speaking AI tetap disembunyikan.
Panduan menjalankan project: [Running project](running-project.md).
