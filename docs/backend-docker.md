# Deploy backend dengan Docker Compose

Compose ini hanya menjalankan API dan tidak membuat container database. Pada
Linux, `network_mode: host` mengizinkan container mengakses service host melalui
`127.0.0.1`. Port default API adalah `18084`; Compose tidak menerbitkan port
karena backend mendengarkan langsung pada network host. Port tersebut harus
tersedia di host dan diizinkan aturan firewall yang berlaku.

## Jalankan

Di server Linux, clone atau salin repository beserta `backend/assets`, lalu
buat file konfigurasi lokal:

```sh
cp backend/.env.example backend/.env
```

Isi nilai OAuth dan konfigurasi lain yang diperlukan. `backend/.env` dibaca
oleh API di dalam container dan tidak dimasukkan ke Docker image. Jangan commit
file ini. Pastikan port `18084` tersedia, kemudian build dan jalankan:

```sh
docker compose -f docker-compose.backend.production.yml up -d --build
docker compose -f docker-compose.backend.production.yml ps
curl http://127.0.0.1:18084/healthz
```

Log dapat dilihat dengan:

```sh
docker compose -f docker-compose.backend.production.yml logs -f backend
```

Atur port lain dengan variabel Compose, misalnya
`BACKEND_DOCKER_PORT=18085 docker compose -f docker-compose.backend.production.yml up -d --build`.
Pada host tanpa port remapping atau reverse proxy, aplikasi klien harus memakai
URL HTTPS production yang diarahkan ke port API tersebut.

Gunakan image hasil build dan Compose pada host Linux. Docker Desktop memerlukan
fitur host networking diaktifkan terlebih dahulu; jika tidak, pilih bridge
network dan arahkan koneksi database ke hostname gateway host sesuai platform.

## Catatan database

Backend memakai SQLite untuk akun, pembayaran, gamifikasi/progres lesson, serta
cache dan arsip hasil kamus. File ditentukan oleh `GAMIFICATION_DB_PATH` dan pada
Compose berada di `/app/data/gamification.db` pada volume data. Data tersebut
tetap tersedia setelah restart; session login dan attempt aktif masih di RAM.
Tidak ada adapter MySQL/MariaDB atau koneksi ke database host yang diaktifkan
oleh mode network host. Jalankan satu proses API per file SQLite dan backup
volume data lengkap. Lihat [penyimpanan kamus](dictionary.md).

Volume bernama `backend_branding` menjaga konfigurasi branding yang diedit
melalui endpoint admin. Materi gambar reading, vocabulary, dan onboarding
dibundel di image backend. `.dockerignore` mengecualikan source Flutter dan
hasil build Android agar tidak ikut sebagai Docker build context.
