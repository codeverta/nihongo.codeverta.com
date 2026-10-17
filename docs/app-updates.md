# Pembaruan wajib Android dan iOS

Aplikasi membaca `GET /v1/app-update-policy` sebelum membuka konten, saat kembali dari background/store, dan setiap lima menit selama berada di foreground. Endpoint dapat diakses sebelum login dan tidak boleh di-cache. Informasi build terpasang dan OS dibaca langsung dari Android/iOS melalui platform channel, bukan nomor versi yang ditulis manual di Dart.

## Pengaturan di server

Buka `https://nihongo.codeverta.com/admin/app-updates`. Gunakan admin token branding yang sudah ada, atau atur `APP_UPDATE_ADMIN_TOKEN` terpisah di `.env` server. Token tidak dimasukkan ke aplikasi, repository, atau URL. `PUT /v1/admin/app-update-policy` membutuhkan Bearer admin token. Halaman pengaturan hanya menampilkan nilai yang memang publik; akses untuk menyimpan tetap membutuhkan token.

Perubahan tersimpan secara atomik di `APP_UPDATE_CONFIG_PATH` (default `data/app-updates.json`) dan langsung berlaku tanpa restart. File berada di server dan tetap ada setelah deployment binary. Menurunkan build minimum ke `0` atau mematikan pemeriksaan platform membatalkan kewajiban update pada pengecekan berikutnya.

Default Android: aktif, build minimum `0` (belum memaksa update), URL untuk `com.codeverta.nihongomaster`. Default iOS: nonaktif sampai Apple ID angka dan URL App Store resminya tersedia. Jangan mengganti URL iOS dengan URL pencarian atau ID aplikasi lain.

## Urutan menerbitkan rilis

1. Bangun aplikasi dengan nomor build yang lebih tinggi. Contoh `1.0.3+4`: nama versi `1.0.3`, nomor build `4`.
2. Publikasikan ke store dan pastikan **semua pengguna yang akan diwajibkan update** sudah dapat menerima rilis itu, termasuk track testing, negara, akun, rollout, dan perangkat yang dituju.
3. Isi build terbaru/nama versi di halaman admin, lalu naikkan build minimum yang masih didukung. Contoh minimum `4`: build di bawah `4` menampilkan dialog wajib. Build `4` atau lebih baru tetap bisa digunakan.
4. Isi persyaratan OS sesuai **rilis minimum yang diwajibkan**: Android memakai API level (`26` = Android 8, `28` = Android 9); iOS memakai versi seperti `15.0` atau `16.4.1`. Android `0`/iOS kosong berarti tidak ada batas OS tambahan dari kebijakan ini.

Build pertama yang menyertakan fitur ini perlu dipasang pengguna terlebih dahulu. Instalasi lama yang belum memiliki pemeriksaan versi tidak bisa menampilkan dialog ini secara retroaktif. Build minimum dari backend adalah pengaturan UX; bukan pengganti autentikasi/otorisasi API dan bisa dilewati oleh client yang dimodifikasi.

## Perangkat yang tidak kompatibel

Jika perangkat memenuhi persyaratan OS tetapi build aplikasinya terlalu lama, dialog tidak dapat ditutup dengan tombol Kembali atau mengetuk di luar. Tombol Update membuka aplikasi Play Store (fallback ke halaman HTTPS) atau App Store. Saat pengguna kembali, aplikasi memeriksa kebijakan dan build lagi.

Jika OS perangkat lebih lama daripada persyaratan rilis yang diwajibkan, aplikasi menjelaskan cara memperbarui OS, bukan mengirim pengguna berulang kali ke store untuk update yang tidak bisa dipasang. Pilih kebijakan:

- **Izinkan versi lama (`allow`, default)**: pengguna mendapat pemberitahuan lalu dapat memilih “Lanjutkan versi ini”. Cocok selama versi lama masih didukung backend. Pemberitahuan yang sama hanya perlu ditutup sekali per sesi; kebijakan yang berubah diperiksa lagi.
- **Hentikan akses (`block`)**: dialog tidak dapat ditutup. Pengguna perlu memperbarui OS atau pindah perangkat, dan mendapat alamat dukungan serta tombol “Periksa lagi”. Gunakan setelah benar-benar menghentikan dukungan versi lama.

Build terpasang yang sudah memenuhi build minimum tetap bisa digunakan walaupun rilis opsional yang lebih baru membutuhkan OS lebih tinggi. Persyaratan OS dalam kebijakan ini tidak mengubah `minSdk` Android maupun deployment target iOS; ubah konfigurasi native pada rilis yang memang membutuhkan perubahan itu, lalu samakan kebijakannya di server.

Pemeriksaan lokal mencakup versi OS/API. Filter store lain seperti arsitektur CPU, hardware, negara, akun, kapasitas penyimpanan, atau rollout tidak dapat dipastikan hanya dengan membuka URL store. Dialog memberi petunjuk dan kontak dukungan jika store tidak menawarkan update. Pastikan cakupan rollout sebelum memaksa update; gunakan rollback kebijakan jika pengguna yang semestinya didukung terblokir.

## Koneksi dan rollback

Saat pertama membuka aplikasi, kegagalan jaringan, data versi native yang tidak dapat dibaca, atau JSON kebijakan yang tidak valid tidak mengunci pengguna. Ada batas waktu pengecekan dan aplikasi mencoba lagi saat kembali ke foreground atau setelah lima menit. Kebijakan wajib tidak disimpan permanen di HP, sehingga gangguan server saat cold start tidak mengunci aplikasi tanpa batas.

Jika sesi aplikasi sudah menerima kewajiban update yang valid, kegagalan koneksi sesaat saat memeriksa ulang mempertahankan dialog tersebut. Respons kebijakan valid yang mengizinkan versi terpasang, menurunkan minimum, atau menonaktifkan platform akan menutupnya. Tombol “Periksa lagi” tetap tersedia.

## Validasi

`go test ./internal/api -run AppUpdate` memeriksa admin auth, validasi nomor build/URL, persistensi setelah restart, iOS nonaktif, dan rollback. `flutter test test/core/app_update` memeriksa OS lama, versi iOS numerik, koneksi gagal, perubahan kebijakan, tombol Kembali, dan modal yang memblokir interaksi aplikasi. CI juga mengompilasi host iOS simulator tanpa signing; build rilis Android mengompilasi platform channel Android.
