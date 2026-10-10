# Backend Developer Guide

Panduan ini menjelaskan cara menyiapkan, menjalankan, menguji, dan
mengembangkan backend Nihongo Master pada mesin lokal.

## 1. Technology requirements

Backend menggunakan:

- Go standard library `net/http`
- Go module bernama `nihongo-master-backend`
- In-memory repository untuk phase-1
- JSON REST API
- OpenAPI 3.1 contract pada `backend/openapi.yaml`

### Versi Go

| Requirement | Version |
|---|---|
| Minimum | Go 1.22 |
| Tested in this repository | Go 1.26.5 windows/amd64 |
| Recommendation | Versi stable terbaru yang kompatibel dengan Go 1.22+ |

Go 1.22 dibutuhkan karena router memakai pola method-aware dari `http.ServeMux`,
misalnya `GET /v1/home` dan path value seperti `{lessonID}`.

Periksa versi yang aktif:

```powershell
go version
```

Contoh output yang valid:

```text
go version go1.26.5 windows/amd64
```

Jika versi lebih rendah dari 1.22, upgrade melalui halaman download resmi Go:
<https://go.dev/dl/>.

## 2. Instalasi Go

### Windows

1. Download installer Windows dari <https://go.dev/dl/>.
2. Jalankan installer dengan lokasi default.
3. Tutup dan buka kembali PowerShell atau terminal IDE.
4. Verifikasi:

```powershell
go version
go env GOROOT GOPATH
```

Tidak perlu menambahkan dependency database atau framework untuk phase-1.

Untuk menjelajahi kontrak endpoint, buka `backend/openapi.yaml` di Swagger
Editor. Kontrak ini mencakup autentikasi, Beranda, pembelajaran, ujian, paket,
profil, progress, dan achievement.

### macOS

Gunakan installer resmi atau Homebrew:

```bash
brew install go
go version
```

### Linux

Gunakan paket distribusi atau archive resmi dari <https://go.dev/dl/>, kemudian:

```bash
go version
```

Pastikan executable `go` tersedia pada `PATH` terminal dan IDE.

## 3. Menjalankan backend

Dari root repository:

```powershell
cd backend
go run ./cmd/api
```

Output normal:

```text
nihongo master API listening on http://localhost:8080
```

Server tetap berjalan selama terminal tersebut terbuka. Hentikan dengan
`Ctrl+C`.

### Mengubah port

PowerShell:

```powershell
$env:PORT = '9090'
go run ./cmd/api
```

Bash/zsh:

```bash
PORT=9090 go run ./cmd/api
```

Environment variable backend:

| Variable | Required | Default | Kegunaan |
|---|---|---|---|
| `PORT` | Tidak | `8080` | Port HTTP server |
| `GOOGLE_CLIENT_ID` | Hanya Google login | - | Web OAuth Client ID untuk validasi audience |

## 4. Memastikan backend aktif

Health endpoint tidak membutuhkan autentikasi.

PowerShell:

```powershell
Invoke-RestMethod http://localhost:8080/healthz
```

curl:

```bash
curl http://localhost:8080/healthz
```

Contoh response:

```json
{
  "status": "ok",
  "time": "2026-07-22T15:00:00Z"
}
```

## 5. Autentikasi

Login development account dan ambil bearer token:

```powershell
$session = Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/v1/auth/login `
  -ContentType application/json `
  -Body '{"email":"demo@nihongomaster.app","password":"Demo123!"}'

$authHeaders = @{
  Authorization = "Bearer $($session.data.token)"
}
```

Endpoint katalog bersifat public. Endpoint `/v1/me/*`, logout, dan memulai exam
membutuhkan bearer token. Google OAuth dijelaskan lengkap pada
[Authentication and Google OAuth Setup](authentication.md).

## 6. Integrasi dengan Flutter

Flutter membaca base URL dari compile-time variable `API_BASE_URL` di
`lib/core/network/dio_client.dart`.

### Android Emulator

Android Emulator memakai `10.0.2.2` untuk mengakses localhost komputer host.
Nilai ini sudah menjadi default aplikasi. Jika `API_BASE_URL` diisi dengan
`localhost` atau `127.0.0.1` pada Android, aplikasi juga menormalkannya otomatis
menjadi `10.0.2.2`:

```text
http://10.0.2.2:8080/v1
```

Jalankan backend terlebih dahulu, lalu:

```powershell
flutter run
```

### iOS Simulator, Windows, macOS, Linux, dan web

Gunakan localhost:

```powershell
flutter run --dart-define=API_BASE_URL=http://localhost:8080/v1
```

### Perangkat Android/iOS fisik

`localhost` dan `10.0.2.2` tidak menunjuk ke komputer developer dari perangkat
fisik. Gunakan alamat IPv4 LAN komputer, misalnya:

```powershell
flutter run `
  --dart-define=API_BASE_URL=http://192.168.1.20:8080/v1
```

Pastikan:

1. Komputer dan perangkat berada pada jaringan yang sama.
2. Windows Firewall atau firewall OS mengizinkan port backend.
3. Jaringan tidak mengaktifkan client isolation.

## 7. Struktur backend

```text
backend/
|- cmd/api/main.go                 Application entry point dan HTTP server
|- internal/api/server.go          Routes, middleware, request/response JSON
|- internal/api/server_test.go     HTTP contract tests
|- internal/auth/google.go         Official Google ID-token verification
|- internal/domain/auth.go         Session dan Google identity models
|- internal/domain/models.go       Domain dan API data structures
|- internal/store/auth.go          Account, bcrypt, dan revocable session
|- internal/store/memory.go        In-memory repository dan seed data
|- go.sum                          Dependency integrity checksums
|- go.mod                          Module dan minimum Go version
`- README.md                       Quick reference
```

Tanggung jawab masing-masing package:

- `cmd/api`: composition root. Membuat dependency dan menjalankan server.
- `internal/api`: transport HTTP. Tidak menyimpan business state secara global.
- `internal/auth`: verifikasi identitas dari provider eksternal.
- `internal/domain`: tipe domain tanpa ketergantungan pada HTTP.
- `internal/store`: penyimpanan dan aturan phase-1 seperti premium access.

## 8. API response format

Response berhasil memakai envelope `data`:

```json
{
  "data": {
    "id": "demo-user",
    "displayName": "Nihongo Learner"
  }
}
```

Response gagal memakai envelope `error`:

```json
{
  "error": {
    "code": "premium_required",
    "message": "premium entitlement required"
  }
}
```

Status penting:

| Status | Arti |
|---|---|
| 200 | Request berhasil |
| 201 | Resource, misalnya exam attempt, berhasil dibuat |
| 401 | Bearer token tidak ada, salah, atau sudah dicabut |
| 402 | Fitur membutuhkan entitlement premium |
| 404 | Resource tidak ditemukan |
| 422 | JSON valid tetapi data melanggar aturan domain |
| 500 | Error internal yang tidak diharapkan |

## 9. Endpoint dan contoh request

### Dashboard

```powershell
Invoke-RestMethod `
  -Uri http://localhost:8080/v1/home
```

### Filter learning track

```powershell
Invoke-RestMethod `
  -Uri 'http://localhost:8080/v1/learning/tracks?system=JLPT'
```

Nilai `system` saat ini: `JLPT`, `NAT`, atau `JFT`.

### Update profil

```powershell
Invoke-RestMethod `
  -Method Patch `
  -Uri http://localhost:8080/v1/me/profile `
  -Headers $authHeaders `
  -ContentType application/json `
  -Body '{"displayName":"Rabih","target":"JLPT N4"}'
```

### Simpan lesson progress

`completion` harus berada antara 0 dan 100. Endpoint legacy ini tidak memberikan
XP atau achievement. Lesson vocabulary, grammar, kanji, dan reading harus
diselesaikan melalui kuis yang dinilai server; lihat [aturan gamifikasi](gamification.md).

```powershell
Invoke-RestMethod `
  -Method Put `
  -Uri http://localhost:8080/v1/me/progress/n5-grammar-001 `
  -Headers $authHeaders `
  -ContentType application/json `
  -Body '{"completion":100,"lastPosition":420}'
```

### Memulai exam gratis

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/v1/exam-attempts `
  -Headers $authHeaders `
  -ContentType application/json `
  -Body '{"examId":"jlpt-n5-diagnostic"}'
```

Exam premium akan menghasilkan HTTP `402` untuk pengguna free.

## 10. Development workflow

Sebelum membuka pull request, jalankan dari folder `backend`:

```powershell
go fmt ./...
go vet ./...
go test ./...
```

Menjalankan satu package:

```powershell
go test ./internal/api -v
```

Menjalankan satu test:

```powershell
go test ./internal/api -run TestFreeUserCannotStartPremiumExam -v
```

`go mod download` atau `go mod tidy` akan mengunduh official Google API client
untuk validasi ID token dan `x/crypto` untuk bcrypt password hashing.

### Menambah endpoint baru

1. Tambahkan atau ubah model di `internal/domain`.
2. Tambahkan operasi penyimpanan di `internal/store`.
3. Daftarkan route dan handler di `internal/api/server.go`.
4. Gunakan `writeData`, `writeError`, dan status HTTP yang sesuai.
5. Tambahkan contract test dengan `httptest`.
6. Jalankan formatter, vet, dan test.

## 11. Batasan phase-1

- Semua data hilang ketika server restart.
- Tidak ada database atau migration.
- Tidak ada password, JWT, atau refresh token.
- Upgrade premium/lifetime belum terhubung payment gateway.
- Question bank dan scoring exam belum persisten.
- Achievement masih berupa fondasi katalog dan progress awal.
- File audio, video, dan offline package belum memakai object storage.

## 12. Troubleshooting

### `go` is not recognized

Go belum terpasang atau terminal belum membaca PATH baru.

1. Jalankan `where.exe go` pada Windows atau `which go` pada macOS/Linux.
2. Install/upgrade Go dari <https://go.dev/dl/>.
3. Tutup dan buka kembali terminal serta IDE.
4. Jalankan `go version`.

### Port 8080 sudah digunakan

Gunakan port lain:

```powershell
$env:PORT = '9090'
go run ./cmd/api
```

Jangan lupa menyesuaikan Flutter:

```powershell
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:9090/v1
```

### Android Emulator tidak dapat terhubung

- Utamakan `10.0.2.2`; aplikasi akan menormalkan `localhost` untuk Android
  Emulator, tetapi URL eksplisit lebih mudah diverifikasi dari log.
- Pastikan backend masih berjalan.
- Coba health check dari komputer terlebih dahulu.
- Periksa firewall dan Android network security apabila memakai HTTP lokal.

### Response `401 unauthorized`

Login kembali dan kirim token yang baru:

```text
Authorization: Bearer <session-token>
```

### Data hilang setelah restart

Ini perilaku yang diharapkan karena repository masih in-memory. Persistence ke
PostgreSQL merupakan pekerjaan backend phase berikutnya.

## 13. Roadmap backend production

1. PostgreSQL schema, migration, dan transaction-safe repository.
2. Persistent session dengan expiry, refresh rotation, dan device management.
3. Payment provider webhook dengan signature validation dan idempotency key.
4. Premium/lifetime entitlement audit log.
5. Exam question bank, scoring engine, timer, dan attempt recovery.
6. Object storage serta signed URL untuk audio dan materi offline.
7. Background jobs untuk streak, reminder, leaderboard, dan achievement.
8. Observability: structured logs, tracing, metrics, dan error reporting.
