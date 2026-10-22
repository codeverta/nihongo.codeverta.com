# JLPT N3 Simulation

Paket `jlpt-n3-simulation-1` tersedia di halaman Ujian untuk Lifetime Access.
Soal orisinal Nihongo Master, empat pilihan yang diacak oleh server, dengan pembahasan bahasa Indonesia setelah kedua sesi selesai.

| Sesi | Soal | Waktu |
| --- | ---: | ---: |
| Kosakata dan kanji | 35 | 30 menit |
| Grammar dan reading | 39 | 70 menit |
| Total tertulis | 74 | 100 menit |

Kosakata: 8 bacaan kanji, 6 penulisan, 11 konteks, 5 padanan makna, 5 penggunaan kata.
Grammar: 13 pilihan pola, 5 susunan kalimat dengan ★, 5 grammar teks.
Reading: 4 soal teks pendek, 6 sedang, 4 panjang, 2 pencarian informasi.

Waktu sesi merujuk [struktur JLPT](https://www.jlpt.jp/e/guideline/testsections.html).
Jenis soal merujuk [contoh jenis soal N3](https://www.jlpt.jp/e/guideline/pdf/n3_e.pdf).
Distribusi jumlah merupakan rancangan latihan; listening belum termasuk. Skor adalah persentase benar, bukan nilai berskala atau keputusan kelulusan resmi JLPT.

## Perilaku

- Pengguna free diarahkan ke halaman premium. API memeriksa entitlement untuk overview, mulai, dan akses attempt.
- Attempt aktif dapat dilanjutkan; riwayat N3 terpisah dari N5 dan N4.
- Timer berjalan berdasarkan tenggat server, termasuk saat aplikasi ditutup. Sesi yang selesai dikunci.
- Jawaban kosong dihitung salah. Kunci dan penjelasan hanya dikirim setelah selesai.
- Penyelesaian memakai alur XP/achievement yang sudah ada; pengiriman ulang tidak menggandakan reward.
- Attempt simulasi masih berada di MemoryStore seperti N5/N4: restart backend menghapus attempt dan riwayat simulasi. Ledger gamifikasi menggunakan MariaDB.

## Thumbnail

Empat ilustrasi asli dibuat melalui built-in image generation, satu gambar per ujian. File dibundel bersama Flutter sehingga tidak perlu fetch gambar eksternal:

- `assets/images/exams/n5.webp`: stasiun musim semi.
- `assets/images/exams/n4.webp`: toko buku dan sepeda.
- `assets/images/exams/n3.webp`: perpustakaan dan skyline.
- `assets/images/exams/diagnostic-n5.webp`: meja belajar dengan kaca pembesar.

Prompt bersama: square thumbnail for a Japanese learning mobile app exam card; refined editorial illustration, soft gouache paper texture, clean friendly shapes, tasteful depth; one coherent scene readable at 96 pixels, centered subject, edge-safe full bleed crop; calm and inviting; no words, letters, numbers, logos, watermark, collage or panels; opaque background.

Prompt adegan:

1. A quiet Japanese train station in spring, a study notebook and pencil in the foreground, gentle cherry blossoms, mint and warm cream palette.
2. A Japanese neighborhood bookshop and a street bicycle, open study notes in the foreground, soft amber and warm cream palette.
3. A Japanese city library with tall bookshelves and a view of a modern skyline through the window, an open study notebook in the foreground, indigo and warm cream palette.
4. A cozy Japanese study desk with a notebook, a magnifying glass resting on its blank page and a small leafy plant, coral red and warm cream palette.

Label level ditampilkan sebagai teks Flutter di atas thumbnail agar tetap jelas dan dapat diakses. Decode gambar dibatasi ukurannya untuk daftar mobile.
