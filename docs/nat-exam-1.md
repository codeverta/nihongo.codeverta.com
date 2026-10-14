# NAT 5Q Exam 1

Latihan tertulis orisinal untuk level dasar, tersedia di Ujian → NAT Test → NAT 5Q Exam 1. Paket memerlukan Lifetime Access seperti ujian selain JLPT N5.

| Bagian | Soal | Waktu |
| --- | ---: | ---: |
| Kosakata dan kanji | 21 | 20 menit |
| Grammar dan reading | 22 | 40 menit |
| Total tertulis | 43 | 60 menit |

Kosakata: 7 bacaan kanji, 5 penulisan, 6 konteks, 3 padanan makna.
Grammar: 9 pilihan pola, 4 susunan kalimat ★, 4 grammar teks.
Reading: 2 soal bacaan pendek, 2 sedang, 1 pencarian informasi.

Waktu dan jumlah soal tertulis mengikuti [format NAT-TEST 5Q](https://nat-test.jp/en/level.html). Soal dibuat khusus untuk Nihongo Master. **Listening belum termasuk**, sehingga ini bukan simulasi seluruh ujian resmi. Skor berupa persentase benar, bukan skor berskala atau keputusan kelulusan resmi NAT-TEST.

## Perilaku dan integrasi

- Exam ID `nat-5q-full-1` menggantikan placeholder katalog dengan paket nyata; paper ID `nat-5q-written-01`.
- Overview `GET /v1/simulation/nat`; start melalui `POST /v1/exam-attempts` dengan exam ID tersebut.
- Sesi: `vocabulary`, lalu `grammar`. Setiap bagian memiliki timer server sendiri. Timer tetap berjalan saat aplikasi ditutup; reconnect mengikuti tenggat awal.
- Jawaban dapat diubah dan ditandai ragu pada bagian aktif. Sesi sebelumnya terkunci setelah dikirim atau waktunya habis.
- Jawaban kosong dihitung salah. Kunci dan pembahasan Indonesia tersedia setelah kedua bagian selesai.
- Resume, riwayat, retake, penilaian dan ledger XP/achievement memakai alur simulasi yang sudah ada; riwayat NAT terpisah dari JLPT dan JFT.
- Route `/exams/nat/simulation` dan `/exams/nat/simulation/{attemptID}`. Entitlement dan ownership diperiksa server untuk operasi attempt.
- Seperti paket sebelumnya, attempt dan riwayat simulasi berada di MemoryStore dan hilang saat backend restart. Ledger gamifikasi memakai MariaDB.

## Thumbnail

File bundel `assets/images/exams/nat.png`, dibuat melalui built-in image generation. Gambar tidak membutuhkan fetch eksternal.

Prompt: Square thumbnail for a Japanese beginner NAT language exam in a learning mobile app. A peaceful Japanese school courtyard with a school building, a maple tree and an open Japanese study notebook with a pencil on a wooden bench in the foreground. Friendly refined editorial gouache illustration with soft paper texture, deep blue, warm cream and soft gold palette. Calm focused learning mood. Large clear shapes readable at 96 pixels, centered coherent scene, edge safe full bleed square, opaque background. No writing, letters, numbers, logos, watermarks or collage.
