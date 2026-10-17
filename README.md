# Nihongo Master

Nihongo Master adalah aplikasi pembelajaran bahasa Jepang berbasis Flutter
dengan backend Go.

## Menjalankan aplikasi

Pastikan Flutter SDK, Go, dan Android SDK sudah terpasang, lalu dari folder
utama project jalankan perintah yang sama di Windows, Linux, atau macOS:

```text
dart run.dart
```

Launcher akan membaca Google Web Client ID dari `backend/.env`, menjalankan
backend, menunggu API siap, lalu menjalankan Flutter. Backend dan app berhenti
bersama saat Flutter dihentikan.

Pilih emulator tertentu dengan `dart run.dart -d emulator-5554`. Untuk ponsel
fisik, gunakan alamat IP komputer: `dart run.dart --api-base-url
http://192.168.1.20:8080/v1 -d <DEVICE_ID>`.

## Struktur proyek

```text
nihongo_master/
|- android, ios, web, ...  Flutter platform projects
|- assets/                 Gambar, ikon, dan vector
|- lib/                    Source code Flutter
|- test/                   Flutter widget tests
|- backend/                Backend HTTP API berbasis Go
|- docs/                   Dokumentasi developer
`- run.dart                Launcher Windows, Linux, dan macOS
```

## Dokumentasi

- [Panduan menjalankan aplikasi dan backend](docs/running-project.md)
- [Authentication dan Google OAuth](docs/authentication.md)
- [Build Android dan rilis Google Play](docs/android-play-store-deployment.md)
- [Build iOS dan rilis App Store](docs/ios-app-store-deployment.md)
- [Setup MariaDB dan migrasi database](docs/mariadb.md)
- [Backend Developer Guide](docs/backend-development.md)
- [Backend API quick reference](backend/README.md)
- [Ikon, splash, branding backend, dan preview Play Store](docs/app-branding.md)
- [JLPT N5 Simulation](docs/jlpt-n5-simulation.md)
- [JLPT N3 Simulation dan thumbnail ujian](docs/jlpt-n3-simulation.md)
- [JFT Exam 1: latihan tertulis](docs/jft-exam-1.md)
- [NAT 5Q Exam 1: latihan tertulis](docs/nat-exam-1.md)
- [SSW / Tokutei Ginou: Kaigo dan Pembersihan Gedung](docs/ssw-section.md)
- [Materi, latihan soal, dan simulasi SSW](docs/ssw-assessments.md)
- [Anki / Flashcards: resource Kotoba, deck, dan jadwal review](docs/anki.md)
- [Modul Grammar JLPT N5](docs/grammar-n5.md)
- [Grammar JLPT N2: 60 lesson dan 120 unit pola](docs/grammar-n2.md)
- [Lifetime Access Rp99.000 dan konfigurasi pembayaran store](docs/lifetime-billing.md)
- [Kamus Jepang gratis: Jisho dan detail kanji](docs/dictionary.md)
- [Kanji N4](docs/kanji-n4.md)
