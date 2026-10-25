# Panduan mengirim push notification ke pengguna

Panduan ini untuk pemilik/operator Nihongo Master. Pengiriman sehari-hari dilakukan lewat panel admin backend, jadi tidak perlu mengubah kode atau membuat APK baru untuk setiap pengumuman.

**Fitur saat ini:** broadcast ke perangkat **Android** yang mengaktifkan pengumuman. Belum ada pilihan satu akun, segmen premium/gratis, penjadwalan, atau pengiriman iOS/web. Pengguna yang belum login juga dapat menerima jika perangkatnya berlangganan pengumuman.

## 1. Cara cepat mengirim pengumuman

Jika Firebase dan backend sudah dikonfigurasi, langsung gunakan langkah berikut.

1. Buka panel pada **domain backend**, bukan halaman aplikasi Flutter:

   ```text
   https://api.domain-anda.com/admin/notifications
   ```

   Ganti domain contoh dengan domain API Anda. URL ini **tidak memakai `/v1`**. Bisa juga buka `/admin/`, lalu pilih **Notifikasi pengumuman**.

2. Isi **Kunci admin backend**, lalu klik **Buka pengaturan**. Ini kunci `NOTIFICATIONS_ADMIN_TOKEN`, atau `BRANDING_ADMIN_TOKEN` jika token khusus notifikasi kosong. Bukan password akun pengguna aplikasi, bukan API key publik Firebase, dan bukan isi file service account.
3. Pada **Status pengiriman**, klik **Periksa koneksi Firebase**. Hasil yang diharapkan:

   ```text
   Konfigurasi Firebase valid. Tidak ada notifikasi yang dikirim.
   ```

   Tombol ini hanya memvalidasi konfigurasi, **tidak mengirim pesan tes ke HP**.

4. Isi **Judul notifikasi**, **Isi pesan**, dan **Halaman saat notifikasi diketuk**.
5. Klik **Simpan draft**. Pesan belum dikirim. Jika pesan diedit lagi, simpan perubahan terlebih dahulu supaya yang dikirim adalah isi terbaru.
6. Di bagian **Draft dan riwayat**, cari draft tersebut lalu klik **Pratinjau & kirim**.
7. Periksa judul, isi, halaman tujuan, dan penerimanya. **Kirim sekarang menyiarkan pesan ke semua perangkat yang berlangganan**, termasuk instalasi testing pada project Firebase yang sama.
8. Klik **Kirim sekarang** satu kali, lalu lihat status pada riwayat. Jika koneksi terputus, gunakan **Muat ulang** untuk memeriksa status sebelum mencoba pengiriman lain.

Membuka panel, menyimpan draft, dan memeriksa koneksi tidak menyiarkan pesan. Pesan yang sudah dikirim tidak dapat ditarik kembali melalui panel.

## 2. Contoh pesan yang bisa langsung digunakan

Gunakan hanya jika konten yang disebut memang sudah tersedia di aplikasi.

| Kebutuhan | Judul | Isi pesan | Pilihan halaman tujuan |
| --- | --- | --- | --- |
| Mengajak belajar | Yuk, lanjutkan belajar Jepang! | Luangkan beberapa menit untuk mengulang kosakata dan lanjutkan pelajaranmu hari ini. | Kosakata |
| Mengumumkan materi | Materi grammar siap dipelajari | Buka materi grammar dan pilih level yang sesuai dengan kemampuanmu. | Tata bahasa |
| Mengenalkan fitur | Ulangi kosakata dengan Anki | Pilih deck dan mulai sesi kartu untuk mengingat kembali kata yang sudah dipelajari. | Anki |
| Mengajak latihan | Sudah siap mencoba ujian N5? | Ukur kemampuanmu lewat latihan dan simulasi yang tersedia di halaman ujian. | Ujian N5 |
| Informasi aplikasi | Ada kabar dari Nihongo Master | Buka aplikasi untuk melihat materi dan fitur belajar yang tersedia. | Beranda |

Judul dibatasi **80 karakter**, isi **500 karakter**. Backend juga membatasi ukuran payload; jika ditolak karena terlalu besar, pendekkan pesan atau kurangi emoji. Tidak ada field gambar pada pengumuman versi ini.

## 3. Pilihan halaman tujuan

Saat penerima mengetuk notifikasi, aplikasi membuka salah satu halaman berikut:

| Label di panel | Path aplikasi |
| --- | --- |
| Beranda | `/home` |
| Kamus | `/dictionary` |
| Anki | `/anki` |
| SSW | `/ssw` |
| Kosakata | `/learn/vocabulary` |
| Tata bahasa | `/learn/grammar` |
| Kanji | `/learn/kanji` |
| Membaca | `/learn/reading` |
| Ujian N5 | `/exams/n5` |
| Lifetime access | `/premium` |

Panel belum mendukung URL website luar, detail lesson tertentu, atau detail lowongan. Notifikasi tidak memberi akses premium; pembatasan akun dan langganan tetap berlaku ketika halaman dibuka.

## 4. Supaya HP pengguna menerima pesan

Pada perangkat Android penerima:

1. Pasang versi Nihongo Master yang sudah memiliki fitur push notification dan buka aplikasinya.
2. Ketuk ikon **lonceng di Beranda**, atau buka **Profil → tombol pengaturan notifikasi**.
3. Aktifkan **Terima pengumuman**.
4. Jika Android menampilkan permintaan izin notifikasi, pilih **Izinkan**. Android 13 ke atas memakai izin runtime `POST_NOTIFICATIONS`; lihat [panduan Android FCM](https://firebase.google.com/docs/cloud-messaging/android/get-started).
5. Pastikan status di aplikasi menjadi **Aktif pada perangkat ini**. Jika masih tertunda, sambungkan internet dan gunakan **Coba sambungkan lagi** bila tombol itu muncul.
6. Untuk mengecek tampilan notifikasi, pindahkan aplikasi ke background dengan tombol Home, lalu kirim pengumuman yang memang ingin disiarkan.

Jika izin pernah ditolak, buka **Pengaturan notifikasi Android** dari layar yang sama dan aktifkan izin/channel pengumumannya. Perangkat perlu Google Play services; untuk emulator, gunakan image yang menyertakannya. Aplikasi yang dihentikan lewat **Force stop** perlu dibuka kembali.

Keikutsertaan disimpan per perangkat. Satu pengguna dengan dua HP yang sama-sama mengaktifkan pengumuman dapat menerima pesan pada kedua HP.

**Panel tidak memiliki tombol tes untuk satu HP.** Untuk uji coba terbatas sebelum broadcast produksi, gunakan lingkungan aplikasi/backend dan project Firebase pengujian yang hanya dipakai perangkat tes. Tombol **Periksa koneksi Firebase** tetap hanya validasi, bukan bukti pesan sudah diterima perangkat.

## 5. Setup Firebase dan backend — cukup sekali

Lewati bagian ini jika **Periksa koneksi Firebase** sudah berhasil.

### A. Siapkan project dan kredensial server

Project yang tercatat pada konfigurasi Android repository adalah `firstproject-ad87e`, dengan package `com.codeverta.nihongomaster`. Pastikan build yang dipasang dan backend pengirim memakai project Firebase yang sama.

1. Buka [Firebase Console](https://console.firebase.google.com/) dan pilih project aplikasi.
2. Pastikan **Firebase Cloud Messaging API (HTTP v1)** aktif. Jika perlu, buka bagian Cloud Messaging pada pengaturan project dan ikuti tautan pengelolaan API.
3. Siapkan service account pengirim di project tersebut. Di Google Cloud Console, buka **IAM & Admin → IAM**, pilih project yang sama, lalu berikan peran **Firebase Cloud Messaging API Admin** (`roles/firebasecloudmessaging.admin`) kepada email service account pengirim.
4. Untuk mendapatkan file JSON lewat Firebase Console, buka **Project settings → Service accounts → Generate New Private Key**, lalu konfirmasi **Generate Key**. Simpan file hasil unduhan secara privat. Pastikan service account pada key itulah yang mendapat izin pengiriman.
5. File harus bertipe `service_account`, dan `project_id` di dalamnya harus sama dengan `FCM_PROJECT_ID`; backend memeriksa keduanya. Implementasi ini membaca file JSON dari path yang Anda tetapkan, bukan otomatis dari kredensial login browser.

Rujukan autentikasi dan izin pengirim: [Firebase: mengirim pesan lewat HTTP v1](https://firebase.google.com/docs/cloud-messaging/send/v1-api). Jangan mencari atau memakai legacy server key untuk endpoint backend ini.

`android/app/google-services.json` adalah konfigurasi klien; **file itu bukan kredensial pengirim backend**. Simpan private key service account hanya di server, di luar Git dan image aplikasi. Jangan menaruhnya di APK, browser, atau chat.

### B. Isi environment backend

Contoh `.env` privat pada server Linux; sesuaikan path dengan deployment:

```dotenv
FCM_PROJECT_ID=firstproject-ad87e
FCM_CREDENTIALS_FILE=/run/secrets/firebase-service-account.json
NOTIFICATIONS_CONFIG_PATH=data/notifications.json
NOTIFICATIONS_ADMIN_TOKEN=<KUNCI_ADMIN_RAHASIA_ANDA>
```

Ganti placeholder token dengan kunci admin yang panjang dan acak, lalu gunakan kunci itu saat membuka panel. Jangan menyalin placeholder secara literal. Jika `NOTIFICATIONS_ADMIN_TOKEN` kosong, backend menggunakan `BRANDING_ADMIN_TOKEN`.

File pada `FCM_CREDENTIALS_FILE` harus bisa dibaca proses backend. Jika memakai Docker, mount file key sebagai file read-only di dalam container; path file di host tidak otomatis tersedia di container. Pada Linux, batasi akses key ke pemilik proses yang memerlukannya, misalnya permission `600`.

Simpan `NOTIFICATIONS_CONFIG_PATH` pada lokasi persisten yang dapat ditulis backend. Draft dan riwayat pengumuman fitur ini disimpan pada file JSON tersebut, **bukan MariaDB**. Backup file dan pertahankan volume saat mengganti container. Jalankan satu instance pengirim untuk file riwayat ini; jangan mengandalkan file bersama untuk koordinasi beberapa replica.

### C. Terapkan dan validasi

1. Restart layanan backend setelah mengubah environment. Jika memakai Docker Compose dan environment didefinisikan dari `.env`/`env_file`, buat ulang container layanan agar environment baru diterapkan; `restart` container saja tidak membaca ulang environment.
2. Buka panel `/admin/notifications` melalui HTTPS.
3. Masukkan token admin dan klik **Periksa koneksi Firebase**.
4. Setelah validasi berhasil, buat draft dan ikuti langkah pengiriman di bagian 1.

Jika salah satu variabel FCM kosong, pengirim tidak aktif tetapi draft masih bisa disimpan. Kredensial yang diisi tetapi file/format/project-nya salah dapat membuat backend gagal startup; periksa log deployment tanpa membagikan isi private key.

## 6. Arti status di riwayat

| Status panel | Artinya | Tindakan |
| --- | --- | --- |
| Draft | Pesan tersimpan, belum dikirim | Boleh diedit, disimpan, lalu dipratinjau |
| Sedang dikirim | Backend sudah mencatat proses pengiriman | Tunggu dan muat ulang riwayat |
| Diterima Firebase | FCM menerima permintaan pengiriman | Pengiriman server selesai; penerimaan tiap HP belum dibuktikan |
| Ditolak Firebase | Pengiriman ditolak | Perbaiki penyebabnya; gunakan **Buka sebagai draft**, simpan, lalu kirim secara eksplisit |
| Status belum pasti | Timeout/koneksi putus/server restart membuat hasil tidak pasti | Jangan langsung membuat kiriman pengganti; pesan sebelumnya mungkin sudah diterima FCM |

HTTP `200` pada endpoint kirim belum tentu berarti terkirim: periksa field `data.status`. Backend juga mengembalikan status gagal/tidak pasti sebagai data agar panel menampilkan riwayat yang tersimpan.

Panel tidak menampilkan jumlah pelanggan topik, jumlah HP yang menerima, atau read receipt. Pesan menggunakan prioritas normal dan TTL 24 jam; jangan mengandalkannya untuk informasi darurat atau pengiriman tepat pada detik tertentu.

## 7. Kalau pesan tidak masuk

| Gejala | Yang perlu diperiksa |
| --- | --- |
| Panel 404/tidak ditemukan | Gunakan domain **backend**, path `/admin/notifications`, tanpa `/v1`. Pastikan versi backend memuat fitur ini dan reverse proxy meneruskan `/admin/` |
| Kunci admin salah | Cocokkan `NOTIFICATIONS_ADMIN_TOKEN`; jika kosong, gunakan `BRANDING_ADMIN_TOKEN`. Jangan memakai token login pengguna |
| Pengirim belum diaktifkan | Periksa kedua variabel FCM, path key di lingkungan proses/container, lalu terapkan ulang konfigurasi |
| Firebase menolak akses/401/403 | Periksa izin service account, kecocokan project, validitas key, dan API HTTP v1 yang aktif |
| Pesan terlalu besar/422 | Pendekkan judul/isi, kurangi emoji, dan pilih halaman dari daftar |
| Sudah diproses/409 | Pesan bukan draft lagi. Muat ulang riwayat; jangan mengulang request kirim tanpa memeriksa hasil |
| Riwayat mencapai 500 pesan | Minta operator backup dan arsipkan riwayat secara terkontrol; jangan mengosongkan file saat layanan sedang mengirim |
| Status diterima FCM, HP belum menerima | Pastikan app terbaru, **Terima pengumuman** aktif, izin dan channel Android aktif, koneksi tersedia, Google Play services berfungsi, serta app tidak di-Force stop |
| Hanya sebagian HP menerima | Periksa izin, versi app, koneksi, dan langganan tiap perangkat. Panel belum punya laporan delivery per perangkat |
| Pesan terlambat | Prioritas normal, pembatasan baterai, dan koneksi dapat menunda penerimaan; TTL yang dipakai adalah 24 jam |
| Ketukan membuka Beranda | Tujuan yang tidak dikenali diarahkan ke Beranda; gunakan pilihan tujuan yang disediakan panel |
| iPhone tidak menerima | Implementasi iOS belum diaktifkan. Memasang key FCM pada backend saja tidak mengaktifkan push iOS |

## 8. Referensi teknis

- [Dokumentasi implementasi push notification](push-notifications.md).
- [FCM HTTP v1: autentikasi dan pengiriman](https://firebase.google.com/docs/cloud-messaging/send/v1-api).
- [FCM Android: client, izin, dan konfigurasi](https://firebase.google.com/docs/cloud-messaging/android/get-started).

Panduan ini disusun dari implementasi panel, endpoint backend, dan penerima Android di repository. Menulis panduan ini tidak mengirim broadcast atau memvalidasi konfigurasi server produksi.
