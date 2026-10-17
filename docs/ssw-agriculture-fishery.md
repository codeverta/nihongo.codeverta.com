# SSW Pertanian, Perikanan Tangkap, dan Budi Daya

Ditinjau 17 Oktober 2026. Di halaman Belajar, seluruh materi JLPT tampil lebih dahulu, diikuti Anki lalu SSW.

| Bidang | Materi | Latihan | Simulasi |
| --- | --- | --- | --- |
| Pertanian | 10 modul, 40 lesson; tanaman 20, peternakan 20 | 160 soal; 16 per modul | Tanaman dan peternakan masing-masing 70 soal / 60 menit |
| Perikanan dan Budi Daya | 10 modul, 30 lesson; tangkap 15, budi daya 15 | 120 soal; 12 per modul | Tangkap 55 soal / 70 menit; budi daya 50 soal / 70 menit |

Setiap lesson berisi tiga tujuan, konsep, penerapan, empat langkah, perhatian keselamatan, kasus dengan respons, empat kosakata Jepang/kana/arti, dua ungkapan, serta empat cek berpenjelasan. Estimasi 25 menit per lesson: Pertanian 16 jam 40 menit, Perikanan 12 jam 30 menit termasuk cek dan pengulangan. Soal konsep, kasus, kosakata, dan prioritas langkah ditulis orisinal dalam bahasa Indonesia.

## Cakupan materi

Pertanian memiliki lima modul tanaman: dasar/pertumbuhan/tanah, pupuk/air/fasilitas, benih/pembentukan tanaman, perlindungan/keselamatan/padi, dan buah/panen/hasil. Lima modul peternakan: dasar/biosekuriti/pakan/air, lingkungan/pengamatan/reproduksi, sapi/kelahiran/pedet/susu, sapi potong/babi/unggas, dan telur/limbah/darurat/laporan.

Perikanan memiliki lima modul tangkap: dasar/sumber daya/cuaca, pelampung/darurat, tali/mesin/jaring, metode pukat/trawl/pancing, dan kesegaran/hasil/bahasa. Lima modul budi daya: sistem/lokasi/air, benih/aklimatisasi/kepadatan, pakan/biomassa/kesehatan, keramba/kerang/rumput laut, dan panen/higiene/laporan.

Contoh hitung dosis rencana, luas, pakan, biomassa, jumlah hidup, dan berat netto adalah latihan numerasi, bukan rekomendasi dosis nyata. Mesin, kapal, kerja di air, ruang tertutup, kelahiran, pestisida, serta obat memerlukan pelatihan dan petugas berwenang sesuai SOP.

## Format dan batas simulasi

- **Tanaman dan Peternakan:** masing-masing satu bagian 70 soal tertulis dengan timer 60 menit. Ujian resmi memuat sekitar 70 soal termasuk mendengar instruksi Jepang. **Audio belum tersedia** di aplikasi; `omittedSections` mencatat `listening` dan `visual_practical`, dan hasil tidak mengukur listening.
- **Tangkap:** teori 30 soal / 50 menit, lalu praktik 25 soal / 20 menit. Jumlah ini mengikuti rincian CBT Prometric Oktober–November 2026 dan berbeda dari budi daya.
- **Budi daya:** teori 40 soal / 50 menit, lalu praktik 10 soal / 20 menit.

Bahasa aplikasi Indonesia dengan istilah Jepang; perikanan resmi menggunakan bahasa Jepang (hiragana menurut penyelenggara). Praktik aplikasi berupa skenario teks dan prioritas langkah, tanpa keseluruhan foto/ilustrasi resmi atau demonstrasi alat. Bank lesson, latihan, dan simulasi saling terkait sehingga ditujukan untuk belajar, bukan asesmen independen terawasi. Skor adalah persentase benar dengan rincian topik, bukan keputusan kelulusan resmi. Distribusi jenis soal pertanian aplikasi adalah susunan orisinal.

## Akses dan data

Materi adalah asset offline gratis. Semua cek harus benar sebelum progres lesson disimpan pada perangkat per akun. Key `ssw_agriculture_completed_v1_{userId}` dan `ssw_fishery_completed_v1_{userId}` terpisah dari Kaigo/Pembersihan Gedung dan akun lain; keduanya dihapus saat account deletion berhasil. Tamu memakai progres sesi.

Latihan dan simulasi membutuhkan login serta Lifetime Access, termasuk URL/API langsung dan attempt berjalan. Latihan memberi feedback setelah server memeriksa pilihan, tanpa timer/XP; hitungan latihan hanya sesi. Simulasi memakai engine SSW: opsi acak, key tersembunyi saat aktif, konfirmasi simpan server, tanda ragu, peta soal, timer absolut per bagian, hasil/pembahasan, riwayat, dan pemulihan restart MariaDB. Bagian terkunci setelah dikirim/deadline. Penghapusan akun menghapus sesi server. Reward mengikuti aturan ujian aplikasi, bukan ambang resmi SSW.

API latihan `GET /v1/ssw/{agriculture|fishery}/practice?module=...`, lalu `POST .../practice/check` dengan `questionId` dan indeks `answer`. Exam ID untuk `POST /v1/exam-attempts`:

- `ssw-agriculture-crop-simulation-1`
- `ssw-agriculture-livestock-simulation-1`
- `ssw-fishery-capture-simulation-1`
- `ssw-fishery-aquaculture-simulation-1`

Overview memakai `/v1/simulation/ssw-agriculture-crop` dan level lain tanpa suffix `-simulation-1`. Session/answer/submit mengikuti kontrak simulasi yang sudah ada. Snapshot pertanyaan, key, deadline, dan bagian memastikan deploy konten baru tidak mengubah attempt lama.

## Resource dan validasi

```text
python scripts/build_agriculture_curriculum.py
python scripts/build_fishery_curriculum.py
python scripts/build_ssw_assessments.py
```

Helper `scripts/ssw_course_authoring.py` memvalidasi empat opsi unik dan membuat schema lesson. Asset: `assets/content/ssw/agriculture.json`, `fishery.json`. Resource Go: `backend/internal/store/ssw_resources/assessments.json`. Tidak ada API eksternal saat runtime.

Tes mencakup dua jalur tiap bidang, kelengkapan konten, keterlacakan seluruh key, opsi/ID unik, jumlah/timer per bagian, login/premium/revocation, bagian terkunci, restart MariaDB tiap jalur, hapus akun, progres terpisah, urutan JLPT–Anki–SSW, dan ponsel 320px dengan teks 140%.

## Referensi resmi

- [NCA: cakupan ASAT 1](https://asat-nca.jp/asat1/overview/cover), [pedoman ASAT 1](https://www.nca.or.jp/upload/asat_guidelines.pdf)
- [MAFF: keselamatan pertanian](https://www.maff.go.jp/j/seisan/sien/sizai/s_kikaika/anzen/)
- [Dainippon Suisankai: SSW dan teks belajar](https://suisankai.or.jp/skill/)
- [Prometric: perikanan tangkap](https://www.prometric-jp.com/ssw/test_list/archives/9), [budi daya](https://www.prometric-jp.com/ssw/test_list/archives/8bmgetj8jgci)
- [Japan Fisheries Agency: keselamatan](https://www.jfa.maff.go.jp/j/kikaku/anzen.html)
