# Lifetime Access dan pembayaran store

Produk sekali bayar, bukan subscription berulang. Harga tampilan awal Rp99.000 dan dapat diubah melalui halaman Admin → Harga Lifetime Access. Perubahan ini hanya mengatur harga yang ditampilkan aplikasi; nominal transaksi tetap mengikuti Google Play atau App Store dan perlu diperbarui manual di console store. Semua materi belajar (N5–N1), kuis, simulasi JLPT, JFT-Basic, NAT, latihan SSW, dan deck Anki JLPT terbuka untuk setiap akun yang sudah login (tamu tidak bisa mengaksesnya); saklar backend `NIHONGO_OPEN_ACCESS=false` mengembalikan gerbang Lifetime Access lama. Lifetime Access tetap untuk manga, tautan lamar kerja, unduhan offline, dan belajar tanpa iklan. Tanya AI untuk Lifetime: 15 pesan per hari, lalu iklan rewarded opsional yang menambah 5 pesan (maks 4 iklan per hari, diatur di admin Tanya AI). Produk tidak menjanjikan konten N1 yang belum tersedia atau materi offline.

Pembelian dan Restore memerlukan akun Google pribadi. Akun demo memakai kredensial bersama sehingga backend menolak pembelian/Restore melalui akun tersebut agar akses dan transaksi tidak dimiliki bersama seluruh pengguna demo.

Package Android sekarang `com.codeverta.nihongomaster`. Product ID default kedua platform `nihongo_master_lifetime`, dapat diganti melalui `LIFETIME_PRODUCT_ID` sebelum membuat produk. iOS masih memakai identifier template proyek; tentukan bundle ID di Xcode dan `APPLE_BUNDLE_ID` sebelum distribusi. Minimum iOS 15 untuk StoreKit 2. Harga tombol beli berasal dari store, bukan konversi harga backend; harga Indonesia perlu diatur di masing-masing console. Jangan memakai tipe consumable, rental, atau auto-renewing subscription.

## Google Play

1. Buat aplikasi dengan package tersebut. Atur signing rilis/Play App Signing dan buat ulang client OAuth Android dengan package serta SHA-1 yang tepat; package sebelumnya telah berubah. Jangan publish dengan debug signing.
2. Buat dan aktifkan one-time product `nihongo_master_lifetime`, opsi **Buy**, lalu atur harga Indonesia agar cocok dengan harga yang disimpan lewat Admin. Jangan consume pembelian lifetime.
3. Aktifkan Android Publisher API dan beri service account akses aplikasi, informasi pembelian, serta pengelolaan pesanan yang diperlukan untuk acknowledge. Mount JSON key di server dan isi `GOOGLE_APPLICATION_CREDENTIALS`; alternatif gunakan ADC dengan `GOOGLE_PLAY_USE_ADC=true`. Kredensial tidak masuk APK.
4. Buat Pub/Sub topic untuk Real-time Developer Notifications, berikan publisher Google Play hak publish, dan arahkan authenticated push ke `https://API/v1/billing/notifications/google`. Isi `GOOGLE_PLAY_PUSH_AUDIENCE` dengan audience JWT push dan `GOOGLE_PLAY_PUSH_EMAIL` dengan email service account push yang diizinkan. Jangan memakai push tanpa autentikasi.
5. Upload signed build ke internal testing, aktifkan license testers, pasang dari Play Store, dan uji pembelian approved, pending/deferred, declined/canceled, already-owned, restore, uninstall/reinstall, refund, serta login dengan akun aplikasi lain.

Referensi: [one-time purchase lifecycle](https://developer.android.com/google/play/billing/lifecycle/one-time), [ProductPurchaseV2](https://developers.google.com/android-publisher/api-ref/rest/v3/purchases.productsv2), [acknowledgement](https://developers.google.com/android-publisher/api-ref/rest/v3/purchases.products/acknowledge).

## App Store

1. Daftarkan bundle ID dan aplikasi di App Store Connect, isi signing team di Xcode, serta aktifkan capability In-App Purchase. Pastikan agreements, tax, dan banking aktif. Build iOS membutuhkan macOS/Xcode.
2. Buat **Non-Consumable** `nihongo_master_lifetime`; atur harga Indonesia agar cocok dengan harga yang disimpan lewat Admin dan periksa harga final. Jangan aktifkan Family Sharing: akses di implementasi ini terikat satu akun aplikasi, hanya menerima ownership `PURCHASED`.
3. Buat In-App Purchase API key; isi `APPLE_ISSUER_ID`, `APPLE_KEY_ID`, `APPLE_BUNDLE_ID`, dan mount `.p8` pada `APPLE_PRIVATE_KEY_PATH` di server.
4. URL App Store Server Notifications V2: `https://API/v1/billing/notifications/apple`. Server memverifikasi JWS dan membaca status terbaru melalui Get Transaction Info sebelum mengubah entitlement.
5. Gunakan server terpisah `APPLE_ENVIRONMENT=Sandbox` untuk sandbox/TestFlight, dan `Production` untuk pembelian produksi. Jangan mengaktifkan fallback sandbox di deployment produksi. Local StoreKit configuration/Xcode unsigned purchases sengaja tidak diterima backend produksi.
6. Uji sandbox pembelian, cancel/deferred, restore, finish transaction yang gagal, reinstall, refund/revocation, serta account deletion. Review login sosial dan kebutuhan Sign in with Apple sebelum submit; integrasi login yang ada memakai Google dan email demo.

Referensi: [Get Transaction Info](https://developer.apple.com/documentation/appstoreserverapi/get-transaction-info), [account deletion](https://developer.apple.com/help/app-review/guideline-reference/5-1-1-account-deletion), [Flutter IAP](https://pub.dev/packages/in_app_purchase).

## Lifecycle dan batas perilaku

| Kasus | Perilaku |
| --- | --- |
| Cancel checkout atau payment error | Tidak memberi entitlement; transaksi gagal ditutup jika store meminta. |
| Pending / Ask to Buy | Tidak membuka premium sampai store melaporkan purchased dan backend memverifikasi. |
| Callback ganda / response hilang | Satu record per platform dan transaction identity; retry tidak membuat kepemilikan ganda. |
| Backend offline / verification gagal | Tidak finish transaksi sukses; simpan retry dalam sesi, gunakan ulang queue store atau Restore setelah restart. |
| Acknowledge Google / finish Apple gagal | Kepemilikan server tersimpan dulu; retry atau reconciliation menyelesaikan acknowledge tanpa membeli ulang. |
| Restore / pindah perangkat | Store proof harus valid; satu pembelian hanya dapat diklaim satu akun aplikasi aktif. |
| Login pada platform lain | Entitlement berlaku di Android dan iOS selama akun aplikasi sama; pembelian cukup sekali. |
| Refund / revoke | Notification memicu verifikasi ulang. Reconciliation setiap lima menit menjadi fallback; akses dicabut termasuk kuis premium yang sedang berlangsung. Refund pada satu store tidak mencabut pembelian aktif lain. |
| Store API outage | Tidak menganggap purchase valid baru. Record aktif kedaluwarsa untuk otorisasi setelah 15 menit tanpa verifikasi; premium dapat sementara terkunci sampai store pulih. |
| Cancel subscription | Tidak ada renewal untuk produk ini. Paket bulanan/tahunan dummy sudah dihapus; tidak ada subscription store lama yang dibuat oleh proyek ini. |
| Logout saat pembayaran | Hasil tidak membuka akun pengguna berikutnya; masuk akun pembeli dan Restore. |
| Hapus akun | Hapus profil, login, progress, achievement, dan attempts; transaksi dilepas dari user. Store tidak otomatis refund. Setelah membuat akun baru, Restore dengan akun store pembeli dapat mengklaim transaksi yang telah dilepas. |

Bukti transaksi minimum (store identity/proof, opaque account token, product, status, dan waktu pemeriksaan) tetap disimpan setelah penghapusan untuk restore, rekonsiliasi, dan mencegah replay; email/nama/user ID akun lama dilepas. Jelaskan retensi tersebut pada privacy policy. Simpan database pada volume terlindungi dengan backup dan kredensial hanya di server. Database MariaDB memakai satu proses API per database dengan named writer lock. Tabel billing/account memakai koneksi `DB_*` yang sama, sehingga backup harus mencakup semuanya. Lihat [panduan MariaDB](mariadb.md).

Saat memakai Docker Compose, mount kredensial melalui override Compose yang read-only ke lokasi sesuai `.env` (contoh `/run/secrets/play.json`, `/run/secrets/apple.p8`); file host tidak otomatis tersedia di container. Jangan menyalin kunci ke image. Gunakan HTTPS publik untuk API, notifications, dan release `API_BASE_URL`.

## API

- `GET /v1/billing`: product ID, UUID account token, platform yang sudah dikonfigurasi, dan entitlement. Token ini dipakai sebagai Google obfuscated account ID / Apple appAccountToken.
- `POST /v1/billing/verify`: `{platform, productId, verificationData, transactionId}`; Google menggunakan purchase token, Apple menggunakan StoreKit 2 signed JWS. `202` untuk pending, `422` untuk proof invalid, `402` untuk konflik pemilik, `503` jika verifier tidak tersedia.
- `DELETE /v1/me/account`: bearer login dan `{confirmation: "DELETE"}`; `204` setelah penghapusan durable berhasil. Hapus akun tersedia di Profil. Sediakan juga URL permintaan penghapusan akun di listing Google Play sebelum rilis.
- `POST /v1/billing/notifications/google`: authenticated Pub/Sub push. `POST /v1/billing/notifications/apple`: signed App Store V2 notification. Notification diproses ulang berdasarkan status store terkini agar replay tidak menghidupkan refund.

Tidak ada endpoint client untuk mengatur `tier`, `isLifetime`, atau status pembayaran. Semua endpoint lesson dan ongoing attempt diperiksa backend. Katalog vocabulary boleh menampilkan judul lesson premium; isi lesson dan jawaban tidak diberikan kepada pengguna free.

## Validasi rilis

Unit/integration test mencakup ownership, forged proof/JWS, wrong product/app, pending, refund, ack retry, restore setelah deletion/restart, isolasi account, dan penguncian ongoing attempt. Tes native store dengan license tester dan sandbox **tetap wajib**; tanpa produk, akun store, dan kredensial resmi, pembayaran uang nyata tidak dapat diaktifkan hanya dari kode ini.

## Akun Play yang sama, akun aplikasi berbeda, dan kode promo

- Pembelian Google Play dimiliki akun Play di perangkat, sedangkan Lifetime milik satu akun aplikasi. Bila akun Play sudah memiliki Lifetime untuk akun aplikasi lain, `POST /v1/billing/verify` menjawab `409 purchase_owned` dengan `owner` berisi email pemilik yang disamarkan (`ab***@gmail.com`). Aplikasi mengganti tombol beli dengan pesan "sudah dipakai oleh akun …" dan tombol "Masuk dengan akun lain". Saat aplikasi kembali aktif (misalnya setelah ganti akun di Play Store), kepemilikan diperiksa ulang.
- Restore berjalan otomatis saat aplikasi dibuka atau login, jadi kepemilikan sudah diketahui sebelum pengguna menekan beli.
- Bila Google Play menolak checkout karena item sudah dimiliki ("You already own this item"), aplikasi langsung menjalankan restore. Pembelian milik akun yang sedang masuk (atau belum terikat) langsung aktif; pembelian milik akun lain menampilkan pesan pemilik. Teks error mentah dari Play tidak ditampilkan.
- Kode promo yang ditukar di Play Store tidak membawa `obfuscatedExternalAccountId`. Server menerimanya dan mengikatnya ke akun aplikasi pertama yang memverifikasinya, lalu mengonfirmasi (acknowledge) ke Google. Setelah itu berlaku aturan biasa: tidak berpindah antar akun hidup. Pembelian iOS tetap wajib membawa token akun.
- Google mengembalikan dana pembelian yang tidak dikonfirmasi dalam sekitar 3 hari. Kode promo yang ditukar sebelum versi ini dan sudah dibatalkan Google perlu diganti kode baru secara manual (periksa Play Console → Order management).

## Tawaran Lifetime setelah login dan berkala

- Setelah setiap login berhasil (email atau Google, termasuk akun Google baru), akun yang belum Lifetime membuka halaman Lifetime Access dalam mode tawaran (`/premium?offer=after_login&from=<tujuan>`). Mode ini punya tombol X dan "Lanjut gratis"; tombol kembali Android sama dengan lewati. Lewati membawa ke tujuan login semula, atau ke Beranda.
- Saat aplikasi dibuka biasa dan Beranda tampil, tawaran muncul lagi (`/premium?offer=periodic`) bila sudah lewat `offerIntervalDays` hari sejak terakhir tampil di HP itu untuk akun itu (SharedPreferences `lifetime_offer_last_shown_v1_<userId>`). Pemeriksaan ini hanya sekali per proses aplikasi dan menunggu sebentar; bila aplikasi dibuka dari notifikasi dan pindah dari Beranda, tawaran tidak muncul. Login selalu menampilkan tawaran dan memulai hitungan dari awal.
- Tidak ditampilkan ke akun Lifetime/premium, tamu, serta akun demo dan reviewer Play (`canPurchase: false` di data user dari `/auth/me` dan login). Bila Lifetime aktif saat tawaran terbuka (beli, restore, klaim promo), halaman tertutup sendiri dengan pesan "Lifetime Access aktif".
- Pengaturan di Admin → Harga Lifetime Access, disimpan bersama harga (`GET /v1/pricing`): `offerAfterLogin` (bawaan menyala; mematikan juga menghentikan tampilan berkala) dan `offerIntervalDays` (bawaan 10, 0–90; 0 = hanya setelah login). Bila server tidak bisa dihubungi, aplikasi memakai bawaan.
- Analitik: `lifetime_offer_view`, `lifetime_offer_skip`, `lifetime_offer_buy_tap`, `lifetime_offer_purchase`, masing-masing dengan `source` = `after_login` atau `periodic`.
