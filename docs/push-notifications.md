# Push notification pengumuman

Versi pertama untuk Android: pengguna memilih ikon lonceng di Beranda atau tombol pengaturan di Profil, lalu mengaktifkan **Terima pengumuman** dan memberikan izin Android. Firebase baru diinisialisasi setelah pilihan aktif, atau untuk memulihkan langganan yang pernah aktif. Firebase Analytics tidak dipasang. iOS, web, dan desktop belum diaktifkan untuk push.

## Pengaturan Android / Firebase

- Project: `firstproject-ad87e`, nomor `798568705597`.
- Package: `com.codeverta.nihongomaster`.
- File konfigurasi klien: `android/app/google-services.json`. File ini berisi identitas/config publik aplikasi; bukan private service-account key.
- Plugin Google Services dideklarasikan di `android/settings.gradle.kts` sesuai struktur Flutter. Plugin diterapkan di `android/app/build.gradle.kts`.
- `firebase_core` dan `firebase_messaging` mengelola Firebase Android BoM dan library native yang kompatibel; tidak menambahkan Firebase Analytics atau BoM native kedua secara manual.
- SDK Flutter minimum mengikuti plugin: Flutter 3.38.1 / Dart 3.10; CI memakai Flutter 3.44.6.
- Android memerlukan Google Play services. Tidak ada perubahan akun Google Sign-In atau tanda tangan aplikasi.
- Channel Android: `nihongo_announcements`; ikon putih `ic_notification` dipertahankan oleh resource shrinker.
- Semua install yang berlangganan topik `nihongo_announcements` menerima pengumuman, termasuk install dari track testing. Ini bukan pesan privat, bukan bukti pemberian akses premium, dan bukan notifikasi berdasarkan akun.

## Pengirim backend

Isi `.env` privat backend production:

```dotenv
FCM_PROJECT_ID=firstproject-ad87e
FCM_CREDENTIALS_FILE=/home/rabihutomo/apps/codeverta-api/secrets/nihongo-play.json
NOTIFICATIONS_CONFIG_PATH=data/notifications.json
```

Service account pada file tersebut memerlukan `roles/firebasecloudmessaging.admin` pada project dan Firebase Cloud Messaging API aktif. File key hanya di server dengan izin 600; tidak boleh dikirim ke aplikasi, browser, chat, atau GitHub. Backend meminta OAuth scope `firebase.messaging` dan memakai FCM HTTP v1. Saat kedua variabel FCM kosong, pengiriman nonaktif dan draft tetap bisa disimpan.

Password admin memakai `NOTIFICATIONS_ADMIN_TOKEN`, atau `BRANDING_ADMIN_TOKEN` jika kosong. Password hanya ada di memori halaman saat dibuka, bukan localStorage/sessionStorage. Gunakan HTTPS. File draft/riwayat `data/notifications.json` harus disimpan pada direktori kerja persisten dan dibackup; tidak masuk repository.

## Admin

Buka `/admin/` → **Notifikasi pengumuman** (`/admin/notifications`).

1. Masukkan kunci admin backend.
2. Klik **Periksa koneksi Firebase**. Ini `validate_only`: tidak mengirim ke perangkat.
3. Tulis judul, isi, dan pilih halaman tujuan; simpan draft.
4. Pilih **Pratinjau & kirim**, periksa penerima/pesan, lalu **Kirim sekarang**.

Tidak ada pesan yang dikirim otomatis saat deploy, memuat admin, membuat draft, atau memeriksa koneksi. Tidak ada penjadwalan pada versi ini. Judul maksimal 80 karakter dan isi maksimal 500; batas JSON/UTF-8 2048 byte untuk topik juga diperiksa (emoji atau escape JSON dapat membuat ukuran lebih besar).

Status **Diterima Firebase** berarti FCM menerima permintaan, bukan bukti setiap HP menerima pesan. Pengiriman tetap bergantung pada koneksi, izin/channel Android, Play services, dan batasan baterai; pesan kedaluwarsa setelah 24 jam. Aplikasi yang dihentikan paksa perlu dibuka lagi. Halaman tidak menghitung pelanggan topik atau delivery per perangkat.

Saat HTTP timeout/koneksi putus atau server crash ketika mengirim, riwayat menjadi **Status belum pasti**. Pesan tersebut tidak dapat dikirim ulang otomatis karena FCM mungkin sudah menerimanya. Periksa dahulu sebelum membuat pesan pengganti. Pesan yang ditolak secara pasti dapat dibuka kembali, disimpan menjadi draft, dan dikirim secara eksplisit. Reservasi `sending` disimpan sebelum menghubungi FCM, sehingga permintaan ganda/refresh tidak menyiarkan ulang. Riwayat dibatasi 500 pesan; operator dapat mengarsipkan file dengan menghentikan layanan, membackup seluruh riwayat, dan mempertahankan item draft/sending/uncertain ketika memang diperlukan.

Frontend menampilkan semua isi pesan sebagai teks, bukan HTML. Tujuan ketukan memakai daftar path internal yang disetujui; URL eksternal atau tujuan tak dikenal kembali ke Beranda. Login dan akses premium tetap diperiksa halaman aplikasi.

## Kontrol pengguna dan batas versi

Install lama tidak punya penerima push. Pengguna harus memasang rilis yang berisi fitur ini dan mengaktifkan notifikasi sekali. Izin Android tidak diminta saat pertama membuka aplikasi.

Pilihan aktif/nonaktif disimpan per perangkat. Saat offline, perubahan langganan dapat tertunda; aplikasi menjelaskan keadaan itu dan mencoba lagi saat kembali aktif. Mematikan izin notifikasi di pengaturan Android langsung menghentikan tampilan notifikasi. Notifikasi foreground ditampilkan lokal satu kali, sedangkan payload notifikasi background ditangani Android untuk menghindari duplikat. Ketukan notifikasi pada cold start menunggu onboarding selesai sebelum membuka tujuan.

## iOS dan deklarasi privasi

iOS belum menginisialisasi Firebase/FCM. Tahap berikutnya membutuhkan registrasi app iOS, GoogleService-Info.plist, APNs key, push capability, dan signing Apple. Firebase auto-init dan analytics collection tetap false di Info.plist.

Kebijakan privasi dan ringkasan data di server menjelaskan Firebase Cloud Messaging serta identitas instalasi/perangkat opsional. Saat merilis SDK baru ke Google Play, sesuaikan deklarasi Data safety untuk identitas perangkat/instalasi yang diproses Firebase untuk fungsi notifikasi opsional. Tidak ada SDK iklan/analytics perilaku pada implementasi ini.

## Pemeriksaan

- Go: autentikasi admin, validasi draft, draft tidak mengirim, `validate_only`, persistensi, request ganda/paralel, crash recovery, gagal menulis sebelum kirim, serta penolakan FCM vs hasil tidak pasti.
- Flutter: pembatasan path, menunggu onboarding saat ketukan, iOS nonaktif, pemeriksaan update wajib, dan navigasi aplikasi yang sudah ada.
- Build Android release lewat Actions dan compile simulator iOS untuk menjaga kompatibilitas native plugin.
- Penerimaan sebenarnya di HP belum bisa dibuktikan tanpa perangkat dengan rilis baru serta izin notifikasi. Pemeriksaan server menggunakan `validate_only` agar tidak mengirim broadcast tanpa tindakan admin.
