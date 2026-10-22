# SSW / Tokutei Ginou

Section persiapan SSW tersedia dari halaman Belajar dan kategori SSW di halaman Ujian. Route `/ssw` menampilkan 19 bidang dengan pencarian nama Indonesia, Jepang, dan kata kunci umum. Route `/ssw/{sectorId}` menampilkan materi dan tes bidang terkait. Query `?tab=test` membuka area tes dari halaman Ujian.

Semua **19 bidang memiliki materi dan latihan**:

- **Kaigo:** [12 modul / 36 lesson](kaigo-curriculum.md), latihan 87 soal, simulasi keterampilan 45 soal / 60 menit dan bahasa Jepang 15 soal / 30 menit.
- **Pembersihan Gedung:** [10 modul / 30 lesson](ssw-assessments.md), latihan 60 soal dan simulasi teori 20 soal / 20 menit + keputusan praktik 30 soal / 30 menit.
- **Pertanian:** 40 lesson, 160 latihan, simulasi tanaman dan peternakan.
- **Perikanan dan Budi Daya:** 30 lesson, 120 latihan, simulasi tangkap dan budi daya.
- **Pengolahan Linen:** 4 modul / 12 lesson, 48 cek pemahaman dan simulasi belajar orisinal 24 soal / 35 menit.
- **Manufaktur Produk Industri:** 4 modul / 12 lesson, 36 latihan dan simulasi belajar orisinal 24 soal / 35 menit.
- **Konstruksi:** 4 modul / 12 lesson, 36 latihan dan simulasi belajar orisinal 24 soal / 35 menit.
- **Pembuatan Kapal dan Mesin Kapal:** 4 modul / 12 lesson, 36 latihan dan simulasi belajar orisinal 24 soal / 35 menit.
- **Perawatan dan Perbaikan Mobil, Penerbangan, Perhotelan, Transportasi Kendaraan Bermotor, Perkeretaapian, Gudang Logistik, Produksi Makanan dan Minuman, Layanan Makanan/Restoran, Kehutanan, Industri Kayu, dan Daur Ulang Sumber Daya:** masing-masing 4 modul / 12 lesson, 36 latihan dan simulasi belajar orisinal 24 soal / 35 menit.

Materi baru disusun dari sumber resmi dan ditulis dengan penjelasan Indonesia serta kosakata/dialog Jepang yang memiliki bacaan kana. Lihat sumber untuk [Pengolahan Linen, Manufaktur Produk Industri, Konstruksi, dan Pembuatan Kapal](ssw-four-sectors-sources.md), [mobil, penerbangan, dan perhotelan](ssw-sources-auto-aviation-accommodation.md), [transportasi darat, kereta api, dan gudang](ssw-sources-transport-rail-warehouse.md), serta [pangan, kehutanan, industri kayu, dan daur ulang](ssw-sources-food-forestry-recycling.md). Cakupan [Pertanian serta Perikanan dan Budi Daya](ssw-agriculture-fishery.md) juga didokumentasikan.

Materi tersedia offline dan gratis; progres lesson disimpan per akun pada perangkat yang sama. Tamu memakai progres sesi. Latihan dan simulasi SSW memerlukan login serta Lifetime Access, mengikuti aturan seluruh tes selain N5. Materi dan ujian dalam aplikasi adalah bantuan belajar orisinal, bukan soal resmi atau bukti kelayakan visa. Menurut pengumuman resmi yang ditinjau, penerimaan baru/ujian SSW (i) Layanan Makanan sedang dihentikan sementara; rincian format dan tanggal mulai ujian Gudang Logistik serta Daur Ulang Sumber Daya belum diumumkan pada halaman resmi yang ditinjau. Periksa pengumuman resmi terbaru sebelum membuat rencana kerja atau ujian.

Latihan per topik tidak memakai timer, memberikan pembahasan setelah server memeriksa pilihan, dan tidak memberikan XP. Simulasi memiliki pengacakan pilihan, timer server per bagian, tanda ragu, peta soal, hasil, pembahasan, riwayat, serta pemulihan sesi setelah restart backend. Data simulasi dan reward yang dinilai server disimpan di MariaDB; hasil dan reward selesai ditulis atomik pada satu record akun. Penghapusan akun menghapus progres perangkat dan sesi/riwayat server.

Katalog berada di `lib/features/ssw/domain/ssw_sector.dart`. Materi offline berada di `assets/content/ssw/`. Bank server dibangun dengan `scripts/build_ssw_assessments.py`; detail API, cakupan, dan keterbatasan ada di [dokumentasi latihan dan simulasi](ssw-assessments.md).

Daftar bidang SSW (i) diverifikasi pada **22 Oktober 2026** melalui [situs resmi Badan Pelayanan Imigrasi Jepang](https://www.ssw.go.jp/id/about/visa/). Ketersediaan materi di aplikasi tidak menjamin penerimaan pekerja, kelayakan visa, atau jadwal ujian resmi suatu bidang.
