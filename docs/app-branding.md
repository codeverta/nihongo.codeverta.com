# Ikon, splash, dan preview Play Store

## File siap pakai

Folder `output/play-store/` berisi:

- `app-icon-512.png`: ikon Play Store, PNG 512×512.
- `feature-graphic-1024x500.png`: feature graphic PNG RGB 1024×500.
- `previews/`: 7 gambar promosi 1080×1920 dari screenshot aplikasi asli.
- `screenshots/`: 7 screenshot Android asli 1080×1920 tanpa perubahan konten.
- `splash-default.png`: desain splash default 1080×1920, diekspor dari widget Flutter, tanpa status bar atau spinner. Ini aset branding, bukan screenshot store.
- `preview-overview.png`: lembar review, bukan untuk diunggah sebagai screenshot.
- `play-store-upload.zip`: ikon, feature graphic, dan 7 preview yang bisa diekstrak lalu diunggah.

Di Play Console, isi **App icon**, **Feature graphic**, dan **Phone screenshots**
dengan file yang sesuai. Pilih seluruh gambar dalam `previews/` secara berurutan;
gunakan `screenshots/` sebagai alternatif jika ingin tampilan tanpa caption.
Preview direkam dari emulator Android 15 dengan backend lokal dan akun demo.

Ukuran mengacu ke dokumentasi resmi Google:
https://support.google.com/googleplay/android-developer/answer/9866151?hl=en

## Mengedit branding di backend

Tambahkan ke `backend/.env`:

```dotenv
BRANDING_ADMIN_TOKEN=GANTI_DENGAN_TOKEN_ADMIN_PANJANG_DAN_ACAK
BRANDING_CONFIG_PATH=data/branding.json
```

Jalankan backend dari folder `backend`, lalu buka:

```text
http://localhost:8080/admin/branding
```

Masukkan admin token, nama aplikasi, tagline, URL ikon, URL splash, dan warna latar.
Panel ini menampilkan preview sebelum disimpan. Token tidak disimpan di browser.
Admin token berbeda dari token login pengguna; jangan memasukkannya ke Flutter.
Tanpa `BRANDING_ADMIN_TOKEN`, endpoint pengubahan branding ditutup.

Upload gambar PNG/JPG ke Tencent COS atau CDN terlebih dahulu, kemudian gunakan
URL **HTTPS publik**. Ikon disarankan persegi 1024×1024; splash disarankan portrait
1080×1920 dengan konten utama di tengah. Gunakan URL baru ketika mengganti gambar
agar cache CDN tidak menampilkan gambar lama. Kosongkan URL untuk kembali ke default.

Konfigurasi disimpan ke `BRANDING_CONFIG_PATH` dan tetap ada setelah restart.
Pastikan folder tersebut writable dan memakai persistent volume pada deployment
container. Konfigurasi file ini cocok untuk satu instance API; beberapa instance
harus memakai konfigurasi bersama atau storage terpusat.

API publik: `GET /v1/branding`. Pengubahan: `PUT /v1/admin/branding`, memakai
`Authorization: Bearer <BRANDING_ADMIN_TOKEN>` dan JSON lengkap:

```json
{
  "appName": "Nihongo Master",
  "tagline": "Selangkah lebih dekat dengan Jepang",
  "iconUrl": "",
  "splashUrl": "",
  "backgroundColor": "#FAFAFA"
}
```

Aplikasi mengambil pengaturan saat cold start, menyimpan konfigurasi terakhir,
dan memakai default jika koneksi/backend bermasalah. Gambar yang gagal dimuat
juga memakai default. Splash Flutter tampil selama sekitar 2,2 detik setelah
frame pertama; pengambilan konfigurasi dibatasi 1,8 detik. Gambar remote yang
lambat bisa baru terlihat pada pembukaan berikutnya. Cache konfigurasi tetap
tersedia offline; gambar remote tidak dipersistenkan untuk offline.

## Ikon launcher dan splash native

Ikon di layar utama Android/iOS dan splash OS pertama berasal dari paket aplikasi.
Backend mengubah ikon pada splash Flutter dan desain splash setelah engine siap.
Penggantian aset OS memerlukan build dan rilis baru. Ikon default sudah dipasang
untuk Android (termasuk adaptive icon), iOS, macOS, dan web. Splash native Android
dan iOS memakai logo default; build iOS belum diverifikasi di Windows.

Untuk memakai ikon backend dalam build berikutnya:

```powershell
python -m pip install Pillow
python scripts/prepare_branding.py --backend https://DOMAIN_API/v1
flutter build appbundle --release --dart-define=API_BASE_URL=https://DOMAIN_API/v1
```

Script mengambil `iconUrl` dan mengekspor ikon launcher serta logo splash native.
`splashUrl` adalah gambar splash Flutter, bukan gambar layar splash OS Android 12+.
Menjalankan script tanpa `--backend` mengekspor ulang ikon default yang dibundel.

## Membuat ulang aset

```powershell
python scripts/prepare_branding.py
flutter test scripts/export_splash_test.dart
python scripts/prepare_store_previews.py
```

Screenshot sumber harus diambil ulang saat UI berubah. Script preview mempertahankan
seluruh isi screenshot dan hanya mengubah ukurannya secara proporsional dalam layout.
Caption dapat diedit di `scripts/prepare_store_previews.py`.
