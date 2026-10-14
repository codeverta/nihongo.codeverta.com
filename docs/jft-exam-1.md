# JFT Exam 1

Latihan tertulis orisinal Nihongo Master, tersedia pada Ujian → JFT Test → JFT Exam 1 untuk pemilik Lifetime Access.

| Bagian | Soal |
| --- | ---: |
| 文字と語彙: bacaan kanji, makna, penggunaan kata | 12 |
| 会話と表現: grammar dan ungkapan situasional | 12 |
| 読解: pesan/pengumuman dan pencarian informasi | 12 |
| Total | 36 |

Waktu latihan 45 menit dengan satu tenggat server untuk seluruh bagian. Waktu tidak direset saat pindah bagian. Bagian yang dikirim dikunci; pengguna dapat memeriksa jawaban dan tanda ragu pada bagian aktif. Kunci dan pembahasan bahasa Indonesia terbuka setelah seluruh bagian selesai. Jawaban kosong dihitung salah, hasil berupa persentase benar, dan penyelesaian mengikuti ledger XP/achievement yang sudah ada.

Lingkup: **listening belum termasuk**. [JFT-Basic resmi](https://www.jpf.go.jp/jft-basic/e/about/index.html) terdiri dari empat bagian termasuk listening, sekitar 50 soal dalam 60 menit. Paket ini merupakan latihan tertulis dengan durasi yang disesuaikan, bukan tes resmi atau prediksi kelulusan berskala.

## Integrasi

- Exam ID: `jft-a2-full-1` (menggantikan katalog placeholder dengan paket nyata).
- Paper ID: `jft-written-01`.
- Overview: `GET /v1/simulation/jft`, `sharedTimer: true`.
- Mulai: `POST /v1/exam-attempts` dengan exam ID di atas.
- Baca, simpan jawaban, dan kirim bagian memakai endpoint `/v1/simulation/attempts/{attemptID}` yang sama dengan JLPT.
- Section ID: `vocabulary`, `conversation`, `reading`. Overview memberi durasi total pada section pertama, nol pada section berikutnya karena waktu dibagi bersama.
- Route: `/exams/jft/exam-1` dan `/exams/jft/exam-1/{attemptID}`.
- Entitlement diperiksa server untuk overview, start, dan operasi attempt. Free diarahkan ke premium; ownership tetap diperiksa.
- Attempt dan riwayat simulasi tetap berada di MemoryStore seperti paket JLPT, sehingga hilang saat restart backend. Ledger gamifikasi memakai MariaDB.

## Thumbnail

Asset bundel: `assets/images/exams/jft.png`. Dibuat dengan built-in image generation; tidak memerlukan fetch gambar eksternal.

Prompt: Square thumbnail for JFT-Basic Japanese daily-life exam in a learning mobile app. A friendly Japanese neighborhood shopping street, a small grocery store and bus stop, an open study notebook on a cafe table in foreground. Refined editorial gouache illustration, soft paper texture, teal and warm cream with coral accents, inviting calm daylight, centered readable shapes suitable for a 96 pixel thumbnail, edge-safe full bleed. No words, letters, numbers, logos or watermark. One coherent scene, no collage. Opaque background.
