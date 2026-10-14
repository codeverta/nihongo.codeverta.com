# SSW / Tokutei Ginou

Section persiapan SSW tersedia dari halaman Belajar dan kategori SSW di halaman Ujian.
Route `/ssw` menampilkan daftar 19 bidang, dengan pencarian nama Indonesia, Jepang, dan kata kunci umum. Route `/ssw/{sectorId}` menampilkan halaman bidang dengan pilihan Materi dan Tes. Query `?tab=test` membuka area tes dari halaman Ujian.

Halaman daftar memakai ilustrasi lokal, ikon tiap bidang, pencarian, filter ketersediaan, dan pintu masuk unggulan Kaigo. **Materi Kaigo aktif** dengan [kurikulum 12 modul / 36 lesson](kaigo-curriculum.md); bidang lainnya memiliki status **Coming Soon**:

- Materi bidang.
- Kosakata bidang.
- Latihan soal.
- Simulasi ujian.

Semua latihan soal dan simulasi ujian SSW masih Coming Soon. Materi Kaigo menyediakan cek pemahaman dalam lesson dan progress belajar lokal per akun, tanpa timer, attempt ujian, XP, atau achievement. Materi Kaigo dapat dibaca tanpa pembelian; entitlement JLPT dan ujian aktif tidak diubah. Pintu masuk dari Belajar dan Ujian mengikuti autentikasi aplikasi.

Katalog berada di `lib/features/ssw/domain/ssw_sector.dart`, UI di `lib/features/ssw/presentation/ssw_screens.dart`. Konten offline Kaigo berada di `assets/content/ssw/kaigo.json`, dapat dibangun ulang dari `scripts/build_kaigo_curriculum.py`. Progress perangkat dijelaskan dalam dokumentasi kurikulum. Tidak ada perubahan database backend atau pembayaran.

Daftar mencakup bidang SSW (i), diverifikasi pada **11 Oktober 2026** melalui [situs resmi Badan Pelayanan Imigrasi Jepang](https://www.ssw.go.jp/id/about/visa/). Termasuk tiga bidang tambahan: pasokan linen, gudang logistik, serta daur ulang sumber daya. Daftar ini menampilkan cakupan bidang, bukan status penerimaan pekerja atau ketersediaan ujian resmi per bidang.
