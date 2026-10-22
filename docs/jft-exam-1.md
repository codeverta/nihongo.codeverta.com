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

## Exam 2–5

Ujian → JFT Test menyediakan lima paket latihan tertulis yang dapat dipilih secara terpisah. Exam 1 tetap memakai ID, route, dan riwayat lama agar sesi yang sudah ada dapat dilanjutkan. Exam 2–5 masing-masing memiliki 36 soal orisinal, satu timer 45 menit, riwayat dan hasil sendiri, serta persyaratan Lifetime Access yang sama.

| Paket | Exam ID | Overview API | Route |
| --- | --- | --- | --- |
| Exam 1 | `jft-a2-full-1` | `/v1/simulation/jft` | `/exams/jft/exam-1` |
| Exam 2 | `jft-a2-full-2` | `/v1/simulation/jft-2` | `/exams/jft/exam-2` |
| Exam 3 | `jft-a2-full-3` | `/v1/simulation/jft-3` | `/exams/jft/exam-3` |
| Exam 4 | `jft-a2-full-4` | `/v1/simulation/jft-4` | `/exams/jft/exam-4` |
| Exam 5 | `jft-a2-full-5` | `/v1/simulation/jft-5` | `/exams/jft/exam-5` |

Semua paket berisi tiga bagian tertulis: 文字と語彙, 会話と表現, dan 読解. Listening belum tersedia. JFT-Basic resmi saat ini menjelaskan empat bagian, sekitar 50 soal, dan waktu 60 menit; latihan aplikasi ini bukan replika ujian resmi, tidak menggunakan soal resmi, dan skornya bukan hasil berskala atau keputusan kelulusan.

Panduan format dan topik publik: [struktur resmi JFT-Basic](https://www.jpf.go.jp/jft-basic/e/about/index.html), [contoh soal resmi](https://www.jpf.go.jp/jft-basic/sample/q01.html), dan [petunjuk belajar JFT-Basic](https://www.jpf.go.jp/jft-basic/e/support/index.html). [FAQ resmi Irodori](https://www.irodori.jpf.go.jp/en/faq.html) menyebut materi Irodori dapat membantu persiapan JFT-Basic; materi Irodori hanya dipakai sebagai rujukan umum untuk ranah komunikasi sehari-hari dan kerja. Teks, contoh, jawaban, audio, dan ilustrasi seluruh soal di aplikasi ditulis terpisah dan orisinal.

## Thumbnail

Asset bundel: `assets/images/exams/jft.webp`. Dibuat dengan built-in image generation; tidak memerlukan fetch gambar eksternal.

Prompt: Square thumbnail for JFT-Basic Japanese daily-life exam in a learning mobile app. A friendly Japanese neighborhood shopping street, a small grocery store and bus stop, an open study notebook on a cafe table in foreground. Refined editorial gouache illustration, soft paper texture, teal and warm cream with coral accents, inviting calm daylight, centered readable shapes suitable for a 96 pixel thumbnail, edge-safe full bleed. No words, letters, numbers, logos or watermark. One coherent scene, no collage. Opaque background.
