# APK release kecil

Gambar materi (58 PNG, 163.344.854 byte) dipindahkan dari Flutter assets ke
`backend/assets/images/reading` dan `backend/assets/images/vocabulary`.
Gambar onboarding (2.737.359 byte) juga dipindahkan ke `backend/assets/images`.
Path `imageAsset` di JSON tetap menjadi object key `assets/images/...` agar
kompatibel dengan COS. Aplikasi memakai `imageUrl`, atau menyelesaikan object
key terhadap origin `API_BASE_URL` untuk respons backend versi lama.
Ikon `app-icon-small.png` (76.078 byte) dibundel lokal, memakai versi
288 px yang sudah tersedia untuk native splash Android. Gambar master
`app-icon-default.png` tetap tersedia untuk script branding, tetapi tidak
masuk APK. Script branding juga memperbarui ikon kecil ketika branding berubah.
Folder WebP vocabulary kini juga dibundel sebagai cadangan untuk semua 20 tema
N5 (sekitar 2,6 MB sumber gambar). Daftar tema dan potongan ilustrasi kata
tetap tampil saat gambar jaringan sedang dimuat atau gagal, termasuk offline.

## Backend

Jalankan backend dari root repo atau folder `backend`. Gambar otomatis disajikan
di `/assets/images/...`. Sertakan folder `backend/assets` dalam deployment;
gunakan `ASSET_DIRECTORY` untuk lokasi lain.

Untuk COS/CDN, jalankan dari folder `backend`:

```powershell
go run ./cmd/assetsync -source assets -prefix assets
```

`ASSET_BASE_URL` (opsional) mengalahkan `COS_BUCKET_URL` untuk URL gambar.
Keduanya harus menunjuk origin publik, misalnya `https://cdn.example.com`;
backend menambahkan `/assets/images/...`. Tanpa keduanya, URL gambar relatif
diselesaikan aplikasi terhadap origin API. Jika API memakai reverse proxy,
teruskan `/assets/` ke backend juga. Jika deployment memakai CDN tanpa folder
lokal, endpoint aset backend mengarahkan permintaan ke CDN, termasuk onboarding.
Vocabulary N5 memakai ilustrasi bawaan selama menunggu atau gagal memuat gambar
jaringan. Pembaruan gambar dari backend tetap digunakan ketika berhasil dimuat.
Gambar tanpa cadangan bawaan menggunakan placeholder jika tidak dapat dimuat.

## Build

APK per arsitektur adalah default script agar pengguna hanya memasang binary
sesuai perangkatnya. Untuk Android modern biasanya gunakan ARM64.

```powershell
./scripts/build_release.ps1 -ApiBaseUrl https://api.example.com/v1
```

Output: `build/app/outputs/flutter-apk/app-armeabi-v7a-release.apk`,
`app-arm64-v8a-release.apk`, dan `app-x86_64-release.apk`.

Tambahkan `-Universal` untuk satu APK berisi semua arsitektur, atau `-AppBundle`
untuk Play Store. Jika login Google dipakai, berikan
`-GoogleServerClientId YOUR_WEB_CLIENT_ID`. Tanpa `-ApiBaseUrl`, build memakai
default emulator `http://10.0.2.2:8080/v1`; untuk distribusi pakai API HTTPS.

Android release mengaktifkan minification dan resource shrinking. Script juga
memisahkan simbol Dart ke `build/symbols`; simpan simbol tiap versi di luar APK
untuk diagnosis crash. Build Flutter tetap melakukan icon tree shaking.

Referensi: [Flutter Android deployment](https://docs.flutter.dev/deployment/android)
dan [Flutter app size](https://docs.flutter.dev/perf/app-size).

## Hasil build terverifikasi (6 Oktober 2026)

| APK | Byte | MB | MiB |
|---|---:|---:|---:|
| Universal sebelum optimasi | 222.865.218 | 222,87 | 212,54 |
| Universal sesudah optimasi | 52.834.034 | 52,83 | 50,39 |
| ARM 32-bit | 16.347.908 | 16,35 | 15,59 |
| ARM64 | 19.082.736 | 19,08 | 18,20 |
| x86_64 | 20.540.560 | 20,54 | 19,59 |

Universal berkurang 76,29%. Semua APK diverifikasi tanda tangannya dengan
`apksigner`, integritas ZIP, jumlah ABI, dan daftar aset. Satu-satunya aset
project di APK adalah `app-icon-small.png`. Seluruh 59 PNG yang dipindahkan
identik dengan file asli menurut Git object hash.

Build terverifikasi memakai API default emulator. Untuk distribusi, rebuild
dengan `-ApiBaseUrl` HTTPS dan deploy backend/aset. Karena unduhan library
native dari repository macet, build verifikasi memakai engine dari Flutter SDK
lokal versi sama melalui Maven repository sementara di `build/dependencies`.

`flutter analyze`, tes URL gambar/splash/onboarding/navigasi, tes penyajian PNG
dan redirect CDN, serta `go vet ./...` lolos. Suite penuh masih gagal pada
tes kurikulum grammar/kanji dan vocabulary (asumsi kurikulum, pilihan kuis,
serta setup preferences); kegagalan itu terpisah dari pemeriksaan aset di atas.
