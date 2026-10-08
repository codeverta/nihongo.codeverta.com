# Simulasi JLPT N4 — Paket 01

Dapat diakses lewat **Ujian → Buka simulasi N4**. Paket gratis ini memuat 57 soal orisinal, bukan salinan soal JLPT.

| Bagian | Jumlah | Waktu |
| --- | ---: | ---: |
| Kosakata dan kanji | 28 | 25 menit |
| Grammar dan reading | 29 | 55 menit |

Komposisi: 7 bacaan kanji, 5 penulisan, 8 konteks, 4 padanan makna, 4 penggunaan kata; 13 pilihan grammar, 4 susunan kalimat ★, 4 grammar teks; 3 pertanyaan bacaan pendek, 3 bacaan sedang, 2 pencarian informasi. Listening belum tersedia.

Timer ditentukan server. Kosakata terkunci setelah dikirim atau waktunya habis, lalu timer grammar dan reading langsung berjalan. Saat kembali setelah offline, batas waktu dihitung dari deadline awal. Sesi N4, riwayat terakhir sepuluh hasil, dan sesi N5 terpisah. Pilihan jawaban diacak setiap percobaan; jawaban benar dan pembahasan baru tersedia setelah selesai.

Nilai adalah persentase jawaban benar, dibulatkan ke bawah. Jawaban kosong dihitung salah. Hasil menampilkan nilai per jenis soal serta pembahasan bahasa Indonesia; bukan skor berskala atau keputusan kelulusan resmi JLPT. Penyimpanan mengikuti backend memory store saat ini: bertahan selama proses server aktif, belum persisten lintas restart.

API: `GET /v1/simulation/n4`; mulai melalui `POST /v1/exam-attempts` dengan `examId: jlpt-n4-simulation-1`. Endpoint session, answers, dan submit-section digunakan bersama dengan N5.

Referensi struktur: [waktu bagian resmi JLPT](https://www.jlpt.jp/guideline/testsections.html), [revisi 2020](https://www.jlpt.jp/topics/202011051597200631.html), [perkiraan jumlah soal N4](https://www.jlpt.jp/userdata/image/N4question.jpg), dan [penjelasan skor resmi](https://www.jlpt.jp/e/guideline/results.html). Jumlah soal resmi berupa perkiraan, dapat bervariasi pada ujian sesungguhnya.
