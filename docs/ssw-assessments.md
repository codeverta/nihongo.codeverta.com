# SSW Kaigo dan Pembersihan Gedung

Ditinjau 17 Oktober 2026. Buka **Belajar / Ujian → SSW → Kaigo atau Pembersihan Gedung → Materi / Tes**.

## Konten aktif

| Bidang | Materi | Latihan per topik | Simulasi |
| --- | --- | --- | --- |
| Kaigo | 12 modul, 36 lesson yang sudah tersedia | 72 soal keterampilan + 15 bahasa Jepang, 13 topik | Keterampilan 45 soal / 60 menit; bahasa Jepang 15 soal / 30 menit |
| Pembersihan Gedung | 10 modul, 30 lesson, 60 cek pemahaman, 120 entri kosakata, 60 ungkapan | 60 soal, 10 topik | Teori 20 soal / 20 menit + keputusan praktik 30 soal / 30 menit |

Soal latihan dan simulasi memakai bank materi yang sama; ini sarana pembelajaran, bukan pengukuran independen yang diawasi. Simulasi Kaigo membagi 45 soal menjadi dasar 10, tubuh/mental 6, komunikasi 4, dukungan kehidupan 20, dan keputusan praktik 5. Bahasa Jepang membagi kosakata, percakapan, serta dokumen masing-masing 5. Kunci tidak dikirim di respons simulasi aktif; pilihan diacak lagi per attempt oleh server.

Bahasa simulasi keterampilan Kaigo dan Pembersihan Gedung adalah Indonesia dengan istilah Jepang. Bahasa Jepang Kaigo memakai bahasa Jepang dengan bacaan kana dalam tanda kurung, bukan ruby tipografis. Pembersihan Gedung resmi menggunakan bahasa Jepang dengan furigana. Bagian praktik aplikasi menggunakan kasus teks dan pilihan tunggal; tidak menyertakan semua foto, ilustrasi, tipe benar-salah, drag/susun urutan, atau demonstrasi alat. Nilai ditampilkan sebagai persentase benar dengan rincian topik, tanpa klaim kelulusan resmi.

Format dipetakan dari [MHLW Kaigo](https://www.mhlw.go.jp/stf/newpage_000117702.html) dan [JBMA SSW 1号](https://www.j-bma.or.jp/qualification-training/zairyu-1go). Referensi keselamatan tambahan: [MHLW workplace safety](https://anzeninfo.mhlw.go.jp/) dan [CDC cleaning/disinfecting principles](https://www.cdc.gov/hygiene/about/when-and-how-to-clean-and-disinfect-your-home.html). Konten ditulis mandiri dalam bahasa Indonesia; bukan reproduksi buku/soal resmi. Petunjuk produk, SOP fasilitas, dan pelatihan lapangan tetap diperlukan.

## Kurikulum Pembersihan Gedung

Tiga lesson per modul, estimasi 25 menit per lesson (12,5 jam termasuk cek dan pengulangan):

1. **Dasar profesi dan mutu layanan:** lingkup tugas, harian/berkala/inspeksi, etika dan privasi.
2. **Keselamatan dan kesehatan kerja:** slip/trip dan rambu, APD/ergonomi/beban, insiden/tumpahan/darurat.
3. **Kotoran, permukaan, dan bahan kimia:** material dan noda, label/SDS/penyimpanan, pengenceran/waktu kontak.
4. **Peralatan manual dan mesin:** kain/mop/zona, vacuum/debu, mesin lantai/listrik/isolasi alat rusak.
5. **Lantai keras dan karpet:** debu/pengepelan, pencucian/lapisan pelindung, noda tekstil/pengeringan.
6. **Toilet dan sanitasi:** persiapan/privasi/urutan, wastafel/kloset/titik sentuh, stok/fungsi/pemeriksaan akhir.
7. **Kaca, logam, dan area umum:** squeegee/akses aman, finishing/panel elektronik, tangga/lift/jalur ramai.
8. **Sampah dan kontaminasi:** pemilahan/pengangkutan, benda tajam/cairan tubuh, tangan/pembersihan/disinfeksi.
9. **Perencanaan dan mutu:** membaca instruksi/urutan, inspeksi/kerusakan/keluhan, efisiensi/serah terima.
10. **Bahasa Jepang dan integrasi:** rambu/larangan, waktu/satuan/laporan, kasus terpadu/format CBT.

Setiap lesson berisi tujuan, dua bagian penjelasan, empat langkah, perhatian keselamatan, kasus dengan respons, empat kosakata, dua ungkapan, dan dua cek berpenjelasan. Semua cek harus benar sebelum progres lesson disimpan. Progres bersifat lokal per akun (`ssw_building-cleaning_completed_v1_{userId}`); akun lain terpisah dan penghapusan akun membersihkan key. Materi Kaigo memakai key lamanya. Cek lesson dan latihan per topik tidak memberi XP.

## API dan persistensi

- `GET /v1/ssw/{sector}/practice?module={moduleId}` → `{questions:[...]}` tanpa kunci/pembahasan. Sector `nursing-care` atau `building-cleaning`.
- `POST /v1/ssw/{sector}/practice/check` → `{questionId,answer:<index>}`; respons `{correct,answer,explanation,lessonId}`. Tanpa attempt/reward; feedback memang dapat diulang. Pilihan wajib integer dengan indeks valid.
- `GET /v1/simulation/{level}` memakai `ssw-kaigo`, `ssw-kaigo-jp`, atau `ssw-cleaning` untuk overview dan riwayat.
- `POST /v1/exam-attempts` dengan `examId` masing-masing `ssw-kaigo-simulation-1`, `ssw-kaigo-jp-simulation-1`, `ssw-cleaning-simulation-1`.
- Session/save/submit memakai API simulasi yang sudah ada. Pengguna hanya boleh mengakses attempt miliknya. Kunci dan pembahasan muncul setelah seluruh bagian selesai. Pilihan kosong dinilai salah.

Semua endpoint latihan dan simulasi SSW membutuhkan autentikasi dan entitlement Lifetime yang masih valid, termasuk akses langsung serta attempt berjalan. Refund/revocation ditangani middleware billing yang sudah ada. Materi offline tidak memerlukan pembelian. Latihan menampilkan kegagalan jaringan dan dapat diulang tanpa menambah skor sebelum respons berhasil. Simulasi menerbitkan pilihan di UI setelah server mengonfirmasi simpan.

SSW attempt disimpan pada `gamification_users.state.SSWAttempts` menggunakan writer MariaDB yang sudah ada: waktu absolut, bagian aktif, jawaban, tanda ragu, pertanyaan/opsi/key per attempt, dan hasil. Saat startup, sesi dipulihkan dengan urutan waktu. Batas waktu tetap berlaku meskipun aplikasi/backend ditutup. Simpan gagal tidak memutasi sesi aktif. Penyelesaian simulasi dan event gamifikasi ditulis atomik dalam record yang sama; submit ulang tidak memberi reward ganda. Reward mengikuti aturan ujian aplikasi (100 XP sekali per paket jika skor ≥80 dan ada jawaban), bukan ambang kelulusan resmi SSW. Account deletion menghapus row akun dan attempt dari cache.

Latihan per topik hanya menyimpan hitungan selama layar/sesi tersebut; tidak diklaim tersinkron. Riwayat simulasi ditampilkan hingga 10 hasil terakhir per paket. Proses API produksi harus memakai persistence MariaDB yang sudah diwajibkan konfigurasi proyek.

## Resource dan validasi

Jalankan berurutan ketika mengubah konten:

```text
python scripts/build_cleaning_curriculum.py
python scripts/build_ssw_assessments.py
```

Materi adalah asset Flutter `assets/content/ssw/building-cleaning.json`. Resource backend `backend/internal/store/ssw_resources/assessments.json` di-embed oleh Go; tidak membutuhkan API eksternal. Perubahan key/pertanyaan saat deployment tidak memutasi snapshot attempt berjalan.

Tes meliputi kelengkapan konten, jumlah/bagian/waktu soal, key tidak bocor, opsi valid, login/premium/revocation, isolasi akun, submit idempoten, penguncian bagian, timeout, penghapusan akun, restart MariaDB, kegagalan write, alur reader ponsel 320 px dengan teks 140%, feedback latihan/network retry, serta timer Kaigo 60 menit.
