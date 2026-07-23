# Nihongo Master Backend

Backend Go untuk guest catalog, email login, Google OAuth, revocable session,
learning progress, exam entitlement, plans, profile, dan achievement.

Kontrak seluruh endpoint tersedia di [`openapi.yaml`](./openapi.yaml) dan dapat
dibuka dengan Swagger Editor atau dipakai untuk menghasilkan API client.

Dokumentasi lengkap:

- [Backend Developer Guide](../docs/backend-development.md)
- [Authentication and Google OAuth Setup](../docs/authentication.md)

## Requirements

- Minimum: **Go 1.22**
- Tested: **Go 1.26.5 windows/amd64**
- Google OAuth: Web OAuth Client ID dari Google Cloud

## Environment

Backend otomatis membaca `backend/.env` ketika dijalankan dari root project atau
folder backend. File tersebut diabaikan Git. Gunakan `.env.example` sebagai
template untuk developer lain.

## Quick start

```powershell
cd backend
go test ./...
go run ./cmd/api
```

Server default: `http://localhost:8080`

Demo account:

```text
Email: demo@nihongomaster.app
Password: Demo123!
```

Login dan gunakan bearer token:

```powershell
$session = Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/v1/auth/login `
  -ContentType application/json `
  -Body '{"email":"demo@nihongomaster.app","password":"Demo123!"}'

$headers = @{
  Authorization = "Bearer $($session.data.token)"
}

Invoke-RestMethod `
  -Uri http://localhost:8080/v1/auth/me `
  -Headers $headers
```

Google OAuth aktif otomatis ketika `.env` berisi `GOOGLE_CLIENT_ID`. Nilai
tersebut harus sama dengan `GOOGLE_SERVER_CLIENT_ID` yang diberikan ke Flutter.
Lihat [panduan OAuth](../docs/authentication.md).

## Endpoint

| Access | Method | Path | Kegunaan |
|---|---|---|---|
| Public | GET | `/healthz` | Health check |
| Public | POST | `/v1/auth/login` | Email/password login |
| Public | POST | `/v1/auth/google` | Exchange verified Google ID token |
| Auth | POST | `/v1/auth/logout` | Revoke session |
| Auth | GET | `/v1/auth/me` | Restore session/user |
| Public/optional auth | GET | `/v1/home` | Dashboard guest atau user |
| Public | GET | `/v1/learning/categories` | Kategori pembelajaran |
| Public | GET | `/v1/learning/tracks` | Jalur JLPT, NAT, JFT |
| Public | GET | `/v1/exams` | Katalog exam |
| Auth | POST | `/v1/exam-attempts` | Mulai exam dan cek premium |
| Public | GET | `/v1/plans` | Free, monthly, yearly, lifetime |
| Auth | GET/PATCH | `/v1/me/profile` | Profil pengguna |
| Auth | GET | `/v1/me/progress` | Progress pengguna |
| Auth | PUT | `/v1/me/progress/{lessonID}` | Simpan progress |
| Auth | GET | `/v1/me/achievements` | Achievement pengguna |
| Auth | GET | `/v1/vocabulary/levels` | Level N5-N1, lesson, dan progress vocabulary |
| Auth | GET | `/v1/vocabulary/lessons/{lessonID}` | Konten pengajaran vocabulary |
| Auth | POST | `/v1/vocabulary/lessons/{lessonID}/attempts` | Mulai/retry kuis lesson |
| Auth | POST | `/v1/vocabulary/attempts/{attemptID}/answers` | Validasi jawaban dan feedback |
| Auth | POST | `/v1/vocabulary/attempts/{attemptID}/complete` | Nilai dan simpan progress |

## Quality checks

```powershell
go fmt ./...
go vet ./...
go test ./...
```
