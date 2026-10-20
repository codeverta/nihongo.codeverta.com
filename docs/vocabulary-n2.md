# Kosakata JLPT N2

Materi ini berisi 5 lesson bertema dan 40 entri kosakata N2, masing-masing dengan bacaan, arti bahasa Indonesia, catatan pemakaian, dan contoh kalimat. Tema meliputi bisnis dan pekerjaan, ekonomi dan organisasi, lingkungan dan masyarakat, masalah dan penyelesaian, serta tindakan dan perubahan.

JLPT tidak menerbitkan daftar kosakata resmi yang wajib untuk tiap level. Karena itu, materi ini adalah kurikulum belajar pilihan, bukan klaim mencakup seluruh kata N2. Kosakata disusun dengan merujuk daftar publik N2 dari [Tanos](https://www.tanos.co.uk/jlpt/jlpt2/); contoh kalimat dan terjemahan Indonesia ditulis untuk aplikasi ini. Bacaan per kata tersedia di `reading`; contoh menyertakan bacaan kalimat penuh. Aplikasi saat ini menampilkan bacaan kata sebagai satu baris kana, belum sebagai furigana per kanji.

Materi disimpan di `backend/data/vocabulary/n2.json` dan dimuat lewat Go embed pada seed N2 backend. Daftar ini mencakup pilihan kosakata formal, profesional, sosial, dan isu publik, bukan seluruh inventaris kosakata yang mungkin muncul pada JLPT N2. Bacaan kata berkanji tampil di atas katanya sebagai furigana; pengguna dapat menyembunyikannya lewat tombol di layar belajar.
