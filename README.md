# Nihongo Master

## Dokumentasi

- [Deploy dan upload Android ke Google Play Store](docs/android-play-store-deployment.md)
- [Backend Developer Guide](docs/backend-development.md)
- [Authentication dan Google OAuth](docs/authentication.md)

Nihongo Master adalah aplikasi pembelajaran bahasa Jepang berbasis Flutter
dengan backend Go. Produk mencakup jalur JLPT N5-N1, NAT-TEST, JFT-Basic,
Vocabulary, Grammar, Kanji, Listening, Reading, exam, premium/lifetime access,
profil, progress, dan achievement.

## Struktur proyek

```text
nihongo_master/
|- android, ios, web, ...  Flutter platform projects
|- assets/                 Gambar, ikon, dan vector
|- lib/                    Source code Flutter
|- test/                   Flutter widget tests
|- backend/                Backend HTTP API berbasis Go
`- docs/                   Dokumentasi developer
```

## Quick start Flutter

```powershell
flutter pub get
flutter run
```

Konfigurasi default Android Emulator memakai backend lokal pada
`http://10.0.2.2:8080/v1`. Host lain dapat diberikan melalui `dart-define`:

```powershell
flutter run --dart-define=API_BASE_URL=http://localhost:8080/v1
```

## Quick start backend

Backend membutuhkan **Go 1.22 atau lebih baru**.

```powershell
cd backend
go test ./...
go run ./cmd/api
```

API akan tersedia di `http://localhost:8080`.

## Dokumentasi

- [Backend developer guide](docs/backend-development.md)
- [Authentication and Google OAuth setup](docs/authentication.md)
- [Backend API quick reference](backend/README.md)
