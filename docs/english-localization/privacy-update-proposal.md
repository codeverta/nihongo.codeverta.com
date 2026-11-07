# Usulan revisi privasi: Firebase Analytics

Status: **usulan, belum diterapkan** pada halaman live. Teks hukum perlu disetujui reviewer hukum/privasi sebelum dipakai. `[TANGGAL RILIS]` diisi saat build dengan analytics dirilis.

> **Pembaruan: aplikasi sekarang memuat SDK iklan (Google AdMob).** Semua kalimat "tanpa SDK iklan" / "does not include advertising SDKs" pada usulan di bawah **tidak boleh dipakai**. Gunakan bagian "Tambahan wajib jika iklan diaktifkan" di akhir dokumen ini. Rekomendasi teknis nomor 1 (mematikan ID iklan) juga **tidak berlaku** untuk Android, karena AdMob memakai ID iklan.

## Fakta yang menjadi dasar teks

- Aplikasi memakai Firebase Analytics (Google Analytics for Firebase), **hanya di Android**. iOS dan web tidak mengirim data analytics.
- Pengumpulan **aktif secara default** dan dapat dimatikan di Pengaturan ("Bantu tingkatkan aplikasi"). Saat dimatikan, pengumpulan berhenti dan data analytics lokal direset.
- Yang dicatat oleh aplikasi: layar dan fitur yang dibuka (mis. kosakata, kanji, berita), tab yang dipilih, serta dua properti: paket (free/premium/lifetime) dan bahasa aplikasi. Email, nama, ID akun, jawaban latihan, dan isi pencarian **tidak** dikirim ke analytics.
- SDK secara otomatis juga memproses ID instansi aplikasi, event dasar (pembukaan pertama, sesi, pembaruan aplikasi), versi aplikasi, model dan versi OS perangkat, bahasa perangkat, dan perkiraan lokasi tingkat kota/negara yang diturunkan dari alamat IP. Penyedianya adalah Google.
- Tidak ada SDK iklan dan data tidak dipakai untuk iklan. Lihat rekomendasi teknis di bawah untuk memastikan itu.

Karena itu frasa "tidak ... analytics perilaku" harus diganti di empat tempat.

## 1. `legal/privacy.html` (Indonesia)

Tambahkan butir pada daftar "Data yang diproses dan tujuannya" (sebelum "Data teknis" atau sesudah "Notifikasi opsional"):

```html
<li><strong>Analitik penggunaan (Android):</strong> aplikasi Android memakai Firebase Analytics dari Google untuk memahami fitur mana yang paling sering dipakai agar kami dapat memperbaiki aplikasi. Data mencakup layar atau fitur yang dibuka, tab yang dipilih, paket (gratis/premium/lifetime), bahasa aplikasi, ID instansi aplikasi, informasi perangkat dan versi aplikasi, serta perkiraan lokasi tingkat negara/kota dari alamat IP. Kami tidak mengirim email, nama, ID akun, jawaban latihan, atau kata pencarian ke analytics. Anda dapat mematikannya kapan saja di Pengaturan &gt; Bantu tingkatkan aplikasi.</li>
```

Ganti kalimat di "Pihak yang menerima data":

- Lama: `Kami tidak menjual data pengguna dan versi aplikasi saat ini tidak memasang SDK iklan atau analytics perilaku.`
- Baru: `Google juga menerima data analitik penggunaan dari aplikasi Android melalui Firebase Analytics. Kami tidak menjual data pengguna dan versi aplikasi saat ini tidak memasang SDK iklan serta tidak memakai data analitik untuk iklan.`

Ubah baris `class="meta"` di awal: `Pembaruan [TANGGAL RILIS]: ditambahkan penjelasan analitik penggunaan melalui Firebase Analytics (Android) dan cara mematikannya.`

## 2. `legal/en/privacy.html` (English)

Add to the "Data we process and why" list:

```html
<li><strong>Usage analytics (Android):</strong> the Android app uses Google's Firebase Analytics to understand which features are used most so we can improve the app. Data includes the screens or features you open, the tab you select, your plan (free/premium/lifetime), the app language, an app-instance ID, device and app-version information, and approximate country- or city-level location derived from your IP address. We do not send your email, name, account ID, practice answers, or search terms to analytics. You can turn it off at any time in Settings &gt; Help improve the app.</li>
```

Replace in "Who receives data":

- Old: `We do not sell user data, and the current app version does not include advertising SDKs or behavioral analytics.`
- New: `Google also receives usage-analytics data from the Android app through Firebase Analytics. We do not sell user data, and the current app version does not include advertising SDKs or use analytics data for advertising.`

Update the `class="meta"` line: `Updated [RELEASE DATE]: this version adds usage analytics through Firebase Analytics (Android) and explains how to turn it off.`

## 3. `legal/safety.html` (Indonesia) dan 4. `legal/en/safety.html` (English)

Tambahkan satu baris tabel:

```html
<tr><td>ID instansi aplikasi, layar/fitur yang dibuka, informasi perangkat, perkiraan lokasi dari IP (Android; dapat dimatikan)</td><td>Analitik penggunaan fitur untuk perbaikan aplikasi</td></tr>
```

```html
<tr><td>App-instance ID, screens/features opened, device information, approximate location from IP (Android; can be turned off)</td><td>Feature-usage analytics to improve the app</td></tr>
```

Ganti kalimat:

- Lama (id): `Versi ini tidak menggunakan SDK iklan atau analytics perilaku dan tidak meminta izin lokasi, kontak, atau mikrofon.`
- Baru (id): `Versi ini tidak menggunakan SDK iklan, memakai Firebase Analytics hanya untuk analitik penggunaan fitur (Android, dapat dimatikan di Pengaturan), dan tidak meminta izin lokasi, kontak, atau mikrofon.`
- Old (en): `This version does not use advertising SDKs or behavioral analytics and does not request location, contacts, or microphone permissions.`
- New (en): `This version does not use advertising SDKs, uses Firebase Analytics only for feature-usage analytics (Android; can be turned off in Settings), and does not request location, contacts, or microphone permissions.`

Ubah juga kalimat pembuka kedua halaman ("...versi ... kebijakan tanggal 17 Oktober 2026") ke tanggal baru.

## Google Play: formulir Data safety (saran jawaban)

Jawaban akhir tanggung jawab pemilik akun Play Console; cocokkan dengan build yang dirilis.

| Pertanyaan | Saran |
| --- | --- |
| Apakah aplikasi mengumpulkan data? | Ya (sudah ya karena akun, pembelian, dll.) |
| Tipe data baru: **App activity → App interactions** | Dikumpulkan; tujuan: **Analytics**; tidak dibagikan ke pihak ketiga untuk iklan; opsional (pengguna dapat menolak di Pengaturan) |
| Tipe data baru: **Device or other IDs** | Dikumpulkan (ID instansi aplikasi); tujuan: **Analytics**; opsional |
| Tipe data: **App info and performance / Diagnostics** | Sesuaikan dengan pengiriman crash/diagnostik yang sudah ada (`diagnostics/errors`) |
| Apakah data dienkripsi saat transit? | Ya |
| Apakah pengguna dapat meminta penghapusan? | Ya (sudah ada: hapus akun); catat bahwa data analytics terkait ID instansi, bukan akun |

## Rekomendasi teknis (belum diterapkan)

1. **Matikan ID iklan dan personalisasi iklan** di `android/app/src/main/AndroidManifest.xml`, di dalam `<application>`, agar sejalan dengan "tanpa SDK iklan":

   ```xml
   <meta-data android:name="google_analytics_adid_collection_enabled" android:value="false" />
   <meta-data android:name="google_analytics_default_allow_ad_personalization_signals" android:value="false" />
   ```

2. Di Firebase Console → Analytics → Settings: atur **Data retention** ke 2 bulan, matikan **Google signals**, dan jangan tautkan ke produk iklan.
3. Pastikan Pengaturan → "Bantu tingkatkan aplikasi" disebut di panduan atau halaman dukungan bila diperlukan.
4. Putuskan apakah default sebaiknya **mati** untuk wilayah dengan aturan persetujuan ketat (mis. EEA/Inggris, karena aplikasi akan tersedia bagi audiens berbahasa Inggris). Jika ya, ubah default di `AnalyticsService.enabled` menjadi `false` dan tampilkan permintaan persetujuan sekali saat pertama dibuka. Mintalah saran reviewer hukum.

## Tambahan wajib jika iklan diaktifkan (AdMob)

Fakta: iklan banner (Beranda, Belajar, Kamus) dan native (daftar Berita dan Lowongan) hanya ditampilkan kepada pengguna gratis dan tamu di Android. Premium dan Lifetime tidak pernah melihat iklan. Iklan dikendalikan saklar server (`/admin/ads`); saat saklar mati, SDK tidak diinisialisasi dan tidak ada permintaan iklan. Persetujuan dikumpulkan dengan Google UMP sebelum SDK dimulai.

Teks privasi (id): `Pengguna paket gratis di Android melihat iklan dari Google AdMob. Google dan mitranya dapat memakai ID iklan perangkat dan data perangkat untuk menayangkan dan mengukur iklan, sesuai pilihan persetujuan Anda. Pengguna Premium dan Lifetime tidak melihat iklan. Anda dapat mengubah pilihan di Pengaturan > Pilihan privasi iklan (bila tersedia di wilayah Anda) dan mengatur ID iklan di pengaturan perangkat Android.`

Privacy text (en): `Free-plan users on Android see ads served by Google AdMob. Google and its partners may use the device advertising ID and device data to serve and measure ads, according to your consent choices. Premium and Lifetime users do not see ads. You can change your choice in Settings > Ad privacy choices (where available in your region) and manage the advertising ID in Android device settings.`

Checklist Play Console dan teknis:

1. App content → **Ads: Yes, my app contains ads**.
2. Data safety: tambahkan **Device or other IDs** (ID iklan), tujuan **Advertising or marketing** dan **Analytics**; data **dibagikan** dengan Google (iklan). Tambahkan **App activity** untuk tujuan iklan bila AdMob dibiarkan memakai interaksi.
3. Deklarasi **Advertising ID** (izin `com.google.android.gms.permission.AD_ID` sudah ditambahkan di manifest): pilih tujuan Advertising.
4. AdMob: buat akun, aplikasi Android, dan unit banner serta native; ganti App ID uji di `AndroidManifest.xml` (`com.google.android.gms.ads.APPLICATION_ID`) dan unit uji di `/admin/ads`. Di AdMob > Privacy & messaging, buat pesan persetujuan GDPR (UMP) dan opsi privasi AS bila diperlukan. Tanpa pesan ini, pengguna EEA/Inggris tidak akan melihat iklan.
5. Untuk build rilis, jangan sampai ID uji Google masih aktif: ID uji tidak menghasilkan pendapatan. Build debug otomatis memakai ID uji.
6. Pertimbangkan pembatasan konten iklan di AdMob (kategori sensitif) karena audiens belajar bahasa.
7. Perbarui juga empat halaman legal dan tinjau hukum sebelum merilis; sesuaikan kebijakan Google Play Families bila aplikasi menyasar anak di bawah 13 tahun (jangan tampilkan iklan berbasis minat kepada anak).
