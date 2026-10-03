# Panduan Menjalankan Nihongo Master

Project ini terdiri dari dua proses yang harus aktif bersamaan saat memakai
aplikasi Android:

1. **Backend Go** menyediakan API untuk Beranda, katalog, akun, dan progress.
2. **Flutter app** berjalan di Android Emulator atau perangkat Android.

Kalau backend belum aktif, aplikasi tetap bisa terbuka, tetapi Beranda akan
menampilkan pesan bahwa data tidak dapat dimuat.

## Kebutuhan awal

- Flutter SDK terpasang dan bisa dipanggil dengan `flutter --version`.
- Go 1.22 atau lebih baru dan bisa dipanggil dengan `go version`.
- Android Studio dengan plugin Flutter dan Dart.
- Android SDK serta satu Android Emulator (AVD), atau perangkat Android.

## Cara menjalankan dari Android Studio

### 1. Buka project Flutter

Di Android Studio pilih **File > Open**, lalu buka folder utama repository ini,
yaitu folder yang berisi `pubspec.yaml`. Jangan buka folder `android` sebagai
project terpisah untuk menjalankan aplikasi Flutter.

Pastikan plugin Flutter dan Dart aktif. Pilih Android Emulator dari daftar
perangkat di bagian atas, lalu jalankan emulator dari **Device Manager** jika
belum menyala.

### 2. Jalankan backend

Buka terminal Android Studio dengan **View > Tool Windows > Terminal** (atau
`Alt+F12`). Dari folder utama project, jalankan:

```powershell
cd backend
go run ./cmd/api
```

Biarkan terminal ini tetap terbuka. Backend normalnya menampilkan:

```text
nihongo master API listening on http://localhost:8080
```

Periksa backend dari terminal kedua:

```powershell
Invoke-RestMethod http://localhost:8080/healthz
Invoke-RestMethod http://localhost:8080/v1/home
```

Health check harus menampilkan status `ok`, dan `/v1/home` harus mengembalikan
data Beranda.

### 3. Jalankan Flutter di emulator

Buka terminal Android Studio kedua pada folder utama repository, lalu jalankan:

```powershell
flutter pub get
flutter run
```

Jika ada lebih dari satu perangkat, pilih emulator secara eksplisit:

```powershell
flutter devices
flutter run -d emulator-5554
```

`emulator-5554` hanya contoh; gunakan ID yang tampil dari `flutter devices`.
Alternatifnya, pilih konfigurasi Flutter `lib/main.dart` dan tekan **Run** di
Android Studio. Pastikan target device yang dipilih adalah emulator Android.

Alamat API default aplikasi sudah disiapkan untuk Android Emulator:

```text
http://10.0.2.2:8080/v1
```

Alamat `10.0.2.2` adalah jalur khusus dari emulator ke `localhost` komputer.
Jadi backend berjalan di komputer, sementara app berjalan di emulator.

## Menjalankan di perangkat Android fisik

Hubungkan komputer dan ponsel ke jaringan Wi-Fi yang sama, aktifkan USB
debugging, lalu cek perangkat dengan `flutter devices`. Cari alamat IPv4 komputer
dengan `ipconfig`, misalnya `192.168.1.20`, kemudian jalankan:

```powershell
flutter run -d <DEVICE_ID> --dart-define=API_BASE_URL=http://192.168.1.20:8080/v1
```

Ganti alamat contoh dengan IPv4 komputer sendiri. Izinkan koneksi Go pada port
8080 di Windows Firewall jika diminta. Pada perangkat fisik, `localhost` berarti
ponsel itu sendiri, bukan komputer.

## Memperbaiki pesan “Beranda belum dapat dimuat”

1. Pastikan terminal backend masih terbuka dan menampilkan `API listening`.
2. Di terminal komputer, jalankan `Invoke-RestMethod
   http://localhost:8080/healthz`. Jika gagal, backend belum aktif atau portnya
   berbeda.
3. Pastikan app sedang berjalan di Android Emulator, bukan browser atau
   perangkat fisik. Emulator harus memakai `10.0.2.2`, bukan `localhost`.
4. Jika backend memakai port selain 8080, samakan port pada `PORT` backend dan
   `API_BASE_URL` Flutter.
5. Di app, coba muat ulang Beranda. Jika masih gagal, stop Flutter dengan
   `q`/`Ctrl+C`, lalu jalankan lagi `flutter run` setelah backend aktif.
6. Periksa error jaringan di terminal Flutter. `Connection refused` biasanya
   berarti server mati atau port salah; timeout pada perangkat fisik biasanya
   terkait alamat IP, Wi-Fi, atau firewall.

Jika port 8080 sudah dipakai, jalankan backend pada port lain:

```powershell
$env:PORT = '9090'
go run ./cmd/api
```

Lalu jalankan app emulator dengan URL yang sama portnya:

```powershell
flutter run -d emulator-5554 --dart-define=API_BASE_URL=http://10.0.2.2:9090/v1
```

## Login demo

Backend lokal menyediakan akun development berikut:

```text
Email: demo@nihongomaster.app
Password: Demo123!
```

Backend fase saat ini memakai penyimpanan in-memory. Data progress dan sesi
akan kembali ke seed awal ketika proses backend dihentikan.

## Membuka Reading N5 dan N4

Setelah backend dan aplikasi berjalan, login lalu buka **Belajar → N5/N4 →
Reading**. Masing-masing level berisi 12 cerita bergambar. Materi berasal
dari backend; ilustrasinya dibundel di aplikasi.

Jika baru menambahkan materi atau gambar, restart backend dan jalankan ulang
Flutter agar konten serta asset terbaru dimuat. Progres Reading dihitung
terpisah per level, tetapi masih reset ketika backend direstart.
Detail: [Reading N5](reading-n5.md) dan [Reading N4](reading-n4.md).

## Menghentikan aplikasi

- Flutter: tekan `q` di terminal `flutter run`, atau tombol **Stop** Android
  Studio.
- Backend: tekan `Ctrl+C` di terminal backend.

## Jika build Android gagal

Jalankan pemeriksaan berikut dari terminal:

```powershell
flutter doctor -v
flutter doctor --android-licenses
flutter clean
flutter pub get
flutter run -d emulator-5554
```

Terima lisensi Android yang diminta oleh perintah tersebut. Ganti device ID
sesuai output `flutter devices`.
