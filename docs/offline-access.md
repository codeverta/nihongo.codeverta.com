# Pemulihan akun dan akses premium offline

Aplikasi menyimpan profil akun terakhir supaya pengguna dapat kembali ke akun
setelah perangkat kehilangan jaringan atau server sedang tidak tersedia. Sesi
tetap tersimpan pada kegagalan jaringan dan HTTP 5xx. HTTP 401 yang menandakan
sesi dicabut menghapus sesi lokal. Logout menghapus token, profil cache, bukti
offline, dan data lokal privat per akun; katalog dan konten publik tetap tersedia.

## Lease premium

Premium offline memakai JWT ber-algoritma Ed25519 yang ditandatangani backend.
Claims memuat issuer dan audience Nihongo Master, ID akun, tier premium,
`isLifetime: true`, waktu verifikasi terakhir, expiry, serta ID purchase.
Expiry tetap 30 hari setelah waktu verifikasi store yang otoritatif. Meminta
`GET /v1/billing` berulang kali tidak mengubah waktu itu. Jika batas verifikasi
store lima menit terlewati, backend memeriksa ulang Google Play atau App Store
sebelum menerbitkan expiry yang lebih baru.

Purchase pending, refund/revoke yang terkonfirmasi, free entitlement, proof
invalid, atau signature yang tidak dikenal tidak memberi akses. Billing outage
menjawab 503 tanpa mengganti lease lokal yang masih berlaku. Setelah `/billing`
berhasil menyatakan akun free atau refund, klien segera menghapus lease lokal.
Backend tetap memeriksa entitlement server untuk endpoint premium dan ongoing
attempts.

## Konfigurasi signing

Dari folder `backend`, buat pasangan key stabil dan secret HMAC terpisah untuk
paket practice offline:

```powershell
go run ./cmd/offline-keys
```

Helper menyimpan private Ed25519 seed dan HMAC secret di
`backend/secrets/`, dengan permission terbatas dan tanpa pernah mencetak nilai
rahasia. Helper mencetak public pin beserta dua path environment variable.
Salin path tersebut dan nilai public ke `backend/.env`:

```dotenv
OFFLINE_ACCESS_PRIVATE_KEY_FILE=secrets/offline-access-private-key
OFFLINE_PRACTICE_SIGNING_KEY_FILE=secrets/offline-practice-signing-key
OFFLINE_ACCESS_PUBLIC_KEY=<public-key-yang-dicetak-helper>
```

`dart run.dart` membaca public pin dari environment atau `backend/.env` dan
meneruskannya ke Flutter. Build release Windows menerima public pin melalui
parameter `-OfflineAccessPublicKey` atau environment variable
`OFFLINE_ACCESS_PUBLIC_KEY`. Android/iOS release workflow juga memerlukan
public pin sebagai CI variable. Untuk build Flutter manual:

```powershell
flutter build apk --dart-define=OFFLINE_ACCESS_PUBLIC_KEY=<base64-public-key>
```

Gunakan key yang sama stabilnya pada server dan public pin pasangannya pada
build aplikasi. Helper memakai key yang sudah ada dan tidak merotasinya. Jika
pin kosong atau tidak cocok, offline premium ditolak sementara billing online
tetap berjalan. Jangan rotasi signing key tanpa merilis build aplikasi yang
mempercayai public key baru.

Server menerima 32-byte seed atau 64-byte Ed25519 private key dalam base64.
Public key dapat disalin dari log startup saat server signer aktif. Jangan
aktifkan request logging yang merekam purchase proof atau JWT lease.
