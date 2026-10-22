# SSW / Tokutei Ginou

Section persiapan SSW tersedia dari halaman Belajar dan kategori SSW di halaman Ujian.
Route `/ssw` menampilkan daftar 19 bidang, dengan pencarian nama Indonesia, Jepang, dan kata kunci umum. Route `/ssw/{sectorId}` menampilkan halaman bidang dengan pilihan Materi dan Tes. Query `?tab=test` membuka area tes dari halaman Ujian.

Halaman daftar memakai ilustrasi lokal, ikon bidang, pencarian, serta filter ketersediaan. **Empat bidang aktif**:

- **Kaigo:** [12 modul / 36 lesson](kaigo-curriculum.md), latihan 87 soal, simulasi keterampilan 45 soal / 60 menit, dan bahasa Jepang 15 soal / 30 menit.
- **Pembersihan Gedung:** [10 modul / 30 lesson](ssw-assessments.md), latihan 60 soal, simulasi teori 20 soal / 20 menit + penilaian praktik 30 soal / 30 menit.

- **Pertanian:** 40 lesson, 160 latihan, simulasi tanaman dan peternakan.
- **Perikanan dan Budi Daya:** 30 lesson, 120 latihan, simulasi tangkap dan budi daya.

Lihat [cakupan bidang baru](ssw-agriculture-fishery.md).

Materi 15 bidang lain dan seluruh tesnya masih Coming Soon. Materi aktif gratis dan tersedia offline; progres lesson disimpan per akun pada perangkat yang sama. Tamu memakai progres sesi. Latihan dan simulasi SSW memerlukan login serta Lifetime Access, mengikuti aturan seluruh tes selain N5.

Latihan per topik tidak memakai timer, memberikan pembahasan setelah server memeriksa pilihan, dan tidak memberikan XP. Skor latihan hanya berlaku pada sesi tersebut. Simulasi memiliki pengacakan pilihan, timer server per bagian, tanda ragu, peta soal, hasil, pembahasan, riwayat, serta pemulihan sesi setelah restart backend. Data simulasi dan reward yang dinilai server disimpan di MariaDB; hasil dan reward selesai ditulis atomik pada satu record akun. Penghapusan akun menghapus progres perangkat dan sesi/riwayat server.

Katalog berada di `lib/features/ssw/domain/ssw_sector.dart`. Materi offline berada di `assets/content/ssw/`. Sumber penulisan: `scripts/build_kaigo_curriculum.py` dan `scripts/build_cleaning_curriculum.py`; bank server dibangun dengan `scripts/build_ssw_assessments.py`. Detail format, API, cakupan, dan keterbatasan ada di [dokumentasi latihan dan simulasi](ssw-assessments.md).

Daftar mencakup bidang SSW (i), diverifikasi pada **11 Oktober 2026** melalui [situs resmi Badan Pelayanan Imigrasi Jepang](https://www.ssw.go.jp/id/about/visa/). Termasuk tiga bidang tambahan: pengolahan linen, gudang logistik, serta daur ulang sumber daya. Daftar ini menampilkan cakupan bidang, bukan status penerimaan pekerja atau ketersediaan ujian resmi per bidang.
