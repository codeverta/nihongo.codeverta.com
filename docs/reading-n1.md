# Reading N1

Buka **Belajar → N1 → Reading** setelah login. Tersedia 12 lesson orisinal,
36 bagian bacaan, 72 kosakata kontekstual, dan 60 soal dengan pembahasan
bahasa Indonesia. Ini materi latihan buatan aplikasi, bukan soal resmi JLPT.

Setiap lesson berisi sekitar **861–939 karakter Jepang** dalam tiga bagian,
dengan bantuan kana, terjemahan ringkas bahasa Indonesia, dua kosakata, dan
lima soal. Materi melatih pembacaan argumen abstrak, batas bukti, maksud
penulis, perbandingan pendapat, bacaan terpadu, serta pencarian syarat dan
informasi dalam pengumuman. Lulus dengan empat jawaban benar dari lima.

| Lesson | Fokus |
| --- | --- |
| 01 Layanan publik dan keadilan akses | Indikator, pengalaman yang terlewat, dan pemerataan |
| 02 Ingatan dan sejarah lokal | Arsip, kesaksian, dan batas ingatan |
| 03 Tanggung jawab ahli | Fakta, tafsiran, ketidakpastian, dan dialog |
| 04 Mengukur mutu kerja | Indikator kinerja dan dampak pada perilaku |
| 05 Wisata dan konservasi | Pembagian manfaat, dampak, dan biaya lingkungan |
| 06 Koreksi informasi dan kepercayaan | Koreksi, penyebaran, dan kredibilitas |
| 07 Kebisingan kota | Pengukuran dan pengalaman warga |
| 08 Pilihan dalam pendidikan | Pembelajaran individual, ruang bersama, dan data |
| 09 Menerjemahkan budaya | Pilihan penerjemah dan nuansa yang tersisa |
| 10 Informasi bencana dan tindakan | Peringatan, hambatan, dan kesiapan bertindak |
| 11 Dua usulan tentang transportasi kota | Bacaan terpadu: titik temu dan perbedaan dua pendapat |
| 12 Membaca pengumuman program riset | Tenggat, syarat, biaya, dan aturan |

## Dasar penyusunan

Deskripsi resmi JLPT menyebut kemampuan N1 membaca tulisan yang kompleks secara
logis atau abstrak, memahami struktur dan isi, serta mengikuti narasi dan
maksud penulis secara menyeluruh. Spesifikasi resmi juga mencantumkan bacaan
panjang, pemahaman terpadu, pemahaman tematik, dan pencarian informasi sebagai
jenis soal Reading N1. Bacaan dan pertanyaan di kursus ini ditulis baru untuk
latihan keterampilan tersebut; tidak menyalin soal contoh JLPT.

- [Ringkasan kemampuan bahasa yang diperlukan untuk setiap level JLPT](https://www.jlpt.jp/e/about/levelsummary.html)
- [Susunan bagian dan jenis soal JLPT](https://www.jlpt.jp/e/guideline/testsections.html)
- [Contoh soal resmi JLPT untuk peserta](https://www.jlpt.jp/e/samples/forlearners.html)

## Integrasi developer

- Konten: `backend/data/reading/n1.json`, di-embed saat backend Go dibangun.
- Katalog: `GET /v1/reading/course?level=N1` (memerlukan bearer token).
- Detail lesson: `GET /v1/reading/lessons/reading-n1-01` sampai `reading-n1-12`.
- Percobaan, jawaban, dan penyelesaian memakai endpoint reading yang sama
  dengan level lain.
- Restart backend setelah mengubah JSON karena konten tertanam saat build.
- Ilustrasi menggunakan aset reading N3 yang telah tersedia.
