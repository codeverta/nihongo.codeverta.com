# Register coverage soal English

Register ini memuat seluruh soal dan materi belajar yang ditemukan di sumber aplikasi. Satu baris di `questions.csv` mewakili satu soal unik; setiap pilihan jawaban memiliki satu baris di `options.csv`. `materials.csv` mencatat isi pelajaran seperti arti kosakata, contoh dan catatan, penjelasan pola grammar, mnemonic dan contoh kanji, isi reading, serta isi modul SSW. `shared_text.csv` mencatat teks bersama yang memberi konteks pada soal, seperti nama bab, judul bagian, instruksi, atau deskripsi pelajaran, dan menautkannya ke `part_id` terkait. `parts.csv` membagi soal berdasarkan jenis materi, level, sektor (untuk SSW), atau form (untuk simulasi) menjadi bagian maksimal 200 soal. Pemisahan menjaga batas pelajaran bila memungkinkan. Bagian berstatus `Mapped` hanya jika semua teks soal, pilihan, teks bersama, dan materi yang ditautkan sudah tercakup.

Judul referensi yang tampil pada daftar sumber SSW juga dicatat sebagai baris `sourceTitle` di `materials.csv` dan ditautkan ke seluruh soal sektor terkait. Paket JSON Inggris untuk aplikasi dibuat dari teks sumber dan padanan manual dalam register; URL, Japanese, serta readings tetap sama.

Kolom `*_source` mempertahankan teks sumber. Kolom `*_english` hanya diisi bila ada padanan persis yang ditulis manual untuk sumber tersebut. Status `Preserve Japanese` berarti teks Jepang dipertahankan; `Already English` berarti sumber memang sudah berbahasa Inggris; `Needs translation` berarti belum ada padanan manual. Gloss pilihan N1/N2 Kanji yang memang tersimpan dalam bahasa Inggris ditandai `Already English`; pilihan bacaan Jepangnya tetap `Preserve Japanese`. `kind` biasanya metadata internal, tetapi pada kuis Vocabulary berisi label latihan yang ditampilkan kepada pelajar dan ikut diperiksa. Deskripsi gambar kuis Vocabulary ikut direkam di tiga kolom terakhir `questions.csv`. `correct_index` dan `option_index` memakai indeks mulai dari nol. `Mapped` pada tingkat soal/bagian hanya mengukur kelengkapan pemetaan, bukan persetujuan reviewer atau kesiapan rilis.

Untuk memperbarui register setelah katalog berubah, jalankan dari folder `backend` di PowerShell:

```powershell
$env:COVERAGE_EXPORT_PATH = "../../../temp/english-localization-coverage.json"
go test ./internal/store -run '^TestExportLocalizationCoverage$' -count=1

$env:ENGLISH_COVERAGE_BUNDLE_PATH = "../../../temp/english-localization-coverage.json"
$env:ENGLISH_COVERAGE_REPORT_DIR = "../../../docs/english-localization"
go test ./internal/api -run '^TestExportEnglishQuestionCoverage$' -count=1

go run ./cmd/export_ssw_english -source ../assets/content/ssw -coverage ../docs/english-localization -out ../assets/content/ssw/en
```

Buka CSV melalui Excel atau impor ke Google Sheets. Mulai dari `parts.csv` untuk memilih bagian, lalu filter `questions.csv`, `options.csv`, `shared_text.csv`, dan `materials.csv` berdasarkan `part_id` atau konteks materi. Jangan menyalakan gate English publik sebelum seluruh cakupan diterjemahkan dan ditinjau; flag rilis tetap nonaktif.
