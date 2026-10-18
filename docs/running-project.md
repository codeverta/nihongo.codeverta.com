# Panduan Menjalankan Nihongo Master

Project ini terdiri dari backend Go dan aplikasi Flutter. Satu launcher
menjalankan keduanya supaya konfigurasi Google Sign-In ikut masuk ke Flutter.

## Kebutuhan

- Flutter SDK dan Dart dari Flutter SDK ada di `PATH`.
- Go 1.22 atau lebih baru ada di `PATH`.
- Android SDK dan Android Emulator, atau perangkat Android dengan USB debugging.
- File `backend/.env` berisi `GOOGLE_CLIENT_ID` (Web OAuth Client ID).
- MariaDB aktif di port 3306 dan konfigurasi `DB_*` sudah ada di `backend/.env`.
  Lihat [setup MariaDB](mariadb.md).

## Jalankan semuanya

Buka terminal di folder utama project—folder yang memiliki `pubspec.yaml`—dan
jalankan perintah yang sama di Windows, Linux, maupun macOS:

```text
dart run.dart
```

Launcher membaca `backend/.env`, menyalakan backend, menunggu endpoint health
siap, lalu menjalankan Flutter dengan `API_BASE_URL` dan
`GOOGLE_SERVER_CLIENT_ID` yang benar. Output backend dan Flutter tampil di
terminal yang sama. Jika tidak ada perangkat Android aktif, launcher menyalakan
AVD pertama dan menunggu sampai Flutter mendeteksinya. Tekan `q` untuk
menghentikan Flutter; backend ikut berhenti.

Pilih emulator tertentu bila ada beberapa AVD atau perangkat:

```text
dart run.dart -d Pixel_8_API_35
```

Gunakan ID AVD dari `flutter emulators`. Untuk perangkat yang sudah aktif,
gunakan ID dari `flutter devices`. Alamat API default
`http://10.0.2.2:8080/v1` adalah alamat khusus Android Emulator untuk mengakses
backend di komputer.

## Menjalankan ke perangkat Android fisik

Hubungkan ponsel dan komputer ke jaringan yang sama, aktifkan USB debugging,
lalu arahkan app ke alamat IPv4 komputer:

```text
dart run.dart --api-base-url http://192.168.1.20:8080/v1 -d <DEVICE_ID>
```

Ganti alamat contoh dan `<DEVICE_ID>` dengan alamat komputer dan ID perangkat
sendiri. Ponsel harus dapat mengakses port backend `8080` melalui firewall.

## Opsi launcher

- `dart run.dart --help` menampilkan bantuan.
- `dart run.dart --backend-only` hanya menyalakan backend.
- `dart run.dart --no-backend` menjalankan Flutter jika backend sudah aktif di
  terminal lain.
- `dart run.dart --api-base-url URL` mengganti alamat API yang dipakai app.

Kalau backend memakai port selain `8080`, set `PORT` pada `backend/.env` atau
environment terminal, lalu berikan port yang sama melalui `--api-base-url`.

## Troubleshooting

### Emulator tidak mau menyala atau statusnya `offline`

Android Studio dan Android Emulator sudah terpasang, tetapi emulator masih bisa
macet saat memulihkan snapshot boot sebelumnya. Pilih **Cold Boot Now** dari
menu ⋮ AVD di Android Studio > Device Manager. Ini memulai ulang Android tanpa
memakai snapshot yang tersimpan dan tidak menghapus data aplikasi.

Launcher mencoba cold boot otomatis. Untuk cold boot manual, gunakan nama AVD
dari `flutter emulators`:

```text
adb -s emulator-5554 emu kill
emulator -avd Pixel_8_API_35_12GB -no-snapshot-load -gpu swiftshader_indirect
```

Tunggu sampai layar utama Android muncul, lalu cek dengan `flutter devices`.
Jika perintah `adb` atau `emulator` tidak ditemukan, jalankan dengan path SDK
lengkap, di Windows biasanya `$env:ANDROID_SDK_ROOT\platform-tools\adb.exe` dan
`$env:ANDROID_SDK_ROOT\emulator\emulator.exe`. Di macOS dan Linux biasanya
`$ANDROID_SDK_ROOT/platform-tools/adb` dan `$ANDROID_SDK_ROOT/emulator/emulator`.
Pastikan juga akselerasi hardware aktif lewat `emulator -accel-check`.

Jika emulator tetap tidak terdeteksi, jalankan `adb kill-server` lalu
`adb start-server`, kemudian cek lagi `flutter devices`. Hindari **Wipe Data**
kecuali cold boot gagal dan Anda siap menghapus data AVD.

Jika build Android melaporkan lisensi SDK belum diterima, jalankan
`flutter doctor --android-licenses` dan terima lisensi yang diminta.

### `dart`, `flutter`, atau `go` tidak ditemukan

Pastikan Flutter SDK (yang menyertakan Dart) dan Go sudah dipasang dan folder
`bin` masing-masing sudah ditambahkan ke `PATH`. Tutup dan buka ulang terminal
setelah mengubah `PATH`.

### `GOOGLE_SERVER_CLIENT_ID belum dikonfigurasi`

Jalankan app dengan `dart run.dart` dari folder utama project. Jangan memakai
`flutter run` langsung atau tombol Run Android Studio untuk alur ini, karena
keduanya melewati launcher yang meneruskan nilai dari `backend/.env`.

Pastikan `GOOGLE_CLIENT_ID` di `backend/.env` berupa Web OAuth Client ID yang
berakhiran `.apps.googleusercontent.com`. Jika `GOOGLE_SERVER_CLIENT_ID` juga
diisi, nilainya harus sama.

### Beranda gagal memuat data

Pastikan output terminal menunjukkan backend aktif. Emulator Android memakai
`10.0.2.2`; perangkat fisik memakai alamat IPv4 komputer. Jika `PORT` bukan
`8080`, pastikan port pada konfigurasi backend dan URL API app sama.

## Akun demo

```text
Email: demo@nihongomaster.app
Password: Demo123!
```

Akun ini hanya untuk development. Backend fase saat ini memakai penyimpanan
in-memory untuk sebagian data; state tersebut kembali ke seed awal saat proses
backend dihentikan.
