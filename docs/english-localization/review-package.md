# Paket review bahasa Inggris

Dokumen ini untuk reviewer yang akan menyetujui pengalaman English sebelum dibuka ke publik. Status teknis ("Mapped" di register) hanya berarti padanan sudah ditulis; **belum ada review manusia**.

## Yang perlu direview dan oleh siapa

| Area | Isi | Reviewer | Berkas |
| --- | --- | --- | --- |
| Materi dan soal belajar | 9.518 soal, 50.710 pilihan jawaban, 54.013 baris materi, 2.519 teks bersama (JLPT N5–N1, diagnostik N5, simulasi JLPT/NAT/JFT, SSW 19 sektor) | Penutur fasih Inggris dengan pemahaman bahasa Jepang; untuk SSW sebaiknya yang paham istilah sektor | `review-sample-content.csv` (sampel 875 baris) dan register lengkap (`questions.csv`, `options.csv`, `materials.csv`, `shared_text.csv`) |
| Antarmuka aplikasi | 176 teks UI baru dari Wave 1 (+ ±1.470 entri katalog lama di `app_localizations.dart`) | Penutur fasih Inggris / UX writer | `review-ui-strings.csv` |
| Respons backend | Pesan error (86 kode, 135 pesan persis), achievement, fitur paket, label lowongan, deck Anki bawaan | Penutur fasih Inggris | `backend/internal/api/localization_errors.go`, `localization_account.go`, `english_*.go` |
| Halaman publik dan legal | Marketing page, privasi, ketentuan, refund, hapus akun, dukungan, data safety | **Reviewer hukum/privasi** + penutur fasih | `backend/internal/api/site/index.en.html`, `backend/internal/api/legal/en/*.html` |
| Listing store | Google Play dan App Store | Penutur fasih + pemilik produk | `docs/public-english/google-play-listing.md` |

Panel admin sengaja tetap Indonesia dan tidak perlu direview.

## Hasil pemeriksaan otomatis (5 Nov 2026)

Dijalankan terhadap register; hasilnya tidak menggantikan review manusia.

- Semua 103 bagian berstatus `Mapped`; tidak ada baris "Needs translation".
- Tidak ada kata-kata khas Indonesia (yang, dan, tidak, untuk, dengan, dst.) di kolom English pada 50.710 pilihan, 54.013 materi, 2.519 teks bersama, dan 33.272 sel soal.
- Ada 284 pilihan, 40 materi, dan 5 teks bersama yang persis sama dengan sumber. Contoh: `bank`, `supermarket`, `panda`, `steak`, `OpenJLPT`, label seperti `文法 問題1 · Grammar & Reading`. Ini kata serapan atau nama; reviewer cukup memastikan tidak ada yang seharusnya diterjemahkan.
- Pemeriksaan ini tidak menangkap terjemahan yang salah makna, terlalu harfiah, atau tidak wajar. Itu tugas reviewer.

## Cara mengerjakan review

1. Buka CSV di Excel atau Google Sheets. Isi kolom `verdict` dengan `ok`, `fix`, atau `reject`; untuk `fix`/`reject` tulis usulan English di kolom komentar.
2. Sampel konten diambil acak tetap dari setiap bagian register (3 soal per bagian, 2 pilihan per bagian, 2 materi per kategori/level/field). Jika temuan di sampel berulang (misalnya satu istilah diterjemahkan keliru), minta pencarian di register lengkap untuk istilah itu.
3. Kriteria: (a) makna sama dengan sumber Indonesia dan, untuk soal, kunci jawaban tetap benar; (b) teks Jepang, furigana, dan istilah yang dipelajari tidak ikut diterjemahkan; (c) nada ramah dan konsisten (Lesson, Premium, Anki ditulis sama); (d) tidak ada klaim produk, harga, atau hukum yang keliru.
4. Untuk perbaikan, ubah kolom `*_english` di register lalu jalankan ulang ekspor seperti di `README.md` folder ini; jangan sunting JSON hasil ekspor langsung.

Ambang saran: setujui area bila ≥ 98% baris sampel `ok` dan tidak ada `reject` pada soal yang kunci jawabannya berubah makna.

## Wajib dicek reviewer hukum/privasi

- Kebijakan privasi dan ringkasan data safety (Indonesia dan English) saat ini menyatakan versi ini **tidak memakai behavioral analytics**. Pernyataan itu tidak lagi benar karena Firebase Analytics sekarang ditambahkan (lihat `release-checklist.md`, blocker 1).
- Gambar berita: pengambilan gambar kini mencakup gambar tanpa lisensi eksplisit dengan kredit "Foto dari <domain>" dan lisensi "Hak cipta pemilik sumber". Mintalah pendapat hukum sebelum rilis publik; halaman Attributions dan ketentuan perlu konsisten dengan keputusan itu.
- Harga: harga di app berasal dari produk store; halaman web tidak boleh menyiratkan harga IDR berlaku di semua pasar.
- Penghapusan akun dan refund: pastikan klaim sama dengan perilaku build yang dirilis.

## Tabel persetujuan

| Area | Reviewer | Tanggal | Hasil (setuju / perlu revisi) | Catatan |
| --- | --- | --- | --- | --- |
| Materi JLPT | | | | |
| Materi SSW | | | | |
| Simulasi (JLPT/NAT/JFT) | | | | |
| UI aplikasi | | | | |
| Respons backend | | | | |
| Halaman publik | | | | |
| Legal / privasi | | | | |
| Listing store | | | | |
