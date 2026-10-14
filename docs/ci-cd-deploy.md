# CI/CD deploy API Nihongo dan Fitapp

Kedua API dibangun oleh GitHub Actions sebagai binary Linux amd64. Setelah perubahan backend masuk ke `main`, workflow menjalankan test memakai MariaDB sementara khusus CI, mengirim binary ke server lewat SSH, mengganti binary secara atomik, me-restart hanya service API terkait, lalu memeriksa health endpoint. Build dan test tidak memakai database produksi. Workflow tidak membangun Android/Flutter, tidak menjalankan Docker di server, dan tidak mengubah Cloudflare Tunnel.

## Satu kali konfigurasi

Lakukan konfigurasi berikut untuk kedua repository: `codeverta/nihongo.codeverta.com` dan `codeverta/fitapp.codeverta.com`.

1. Salin `ops/deploy-api-from-ci` ke `/home/rabihutomo/apps/codeverta-api/deploy-api-from-ci` di server dan beri izin eksekusi. Skrip ini hanya menerima perintah deploy/rollback untuk API Nihongo atau Fitapp. Contoh setup satu kali dari komputer yang sudah punya SSH alias `serverku`:

   ```sh
   scp ops/deploy-api-from-ci serverku:/home/rabihutomo/apps/codeverta-api/deploy-api-from-ci
   ssh serverku 'chmod 700 /home/rabihutomo/apps/codeverta-api/deploy-api-from-ci && install -d -m 700 /home/rabihutomo/.ssh'
   ```

2. Buat SSH key khusus GitHub Actions. Tambahkan **public key** ke `/home/rabihutomo/.ssh/authorized_keys` dengan pembatasan forced command berikut. Jangan menimpa baris key yang sudah ada:

   ```text
   command="/home/rabihutomo/apps/codeverta-api/deploy-api-from-ci",restrict ssh-ed25519 PUBLIC_KEY_DARI_KEY_KHUSUS
   ```

   Key khusus dapat dibuat sekali dengan `ssh-keygen -t ed25519 -f ~/.ssh/codeverta-api-ci -C github-actions-api-deploy -N ""`. Append baris terbatas tanpa mengganti key lama dengan:

   ```sh
   printf 'command="/home/rabihutomo/apps/codeverta-api/deploy-api-from-ci",restrict %s\n' "$(cat ~/.ssh/codeverta-api-ci.pub)" | ssh serverku 'cat >> /home/rabihutomo/.ssh/authorized_keys && chmod 600 /home/rabihutomo/.ssh/authorized_keys'
   ```

   `restrict` menonaktifkan shell interaktif, port forwarding, agent forwarding, dan X11 forwarding. Key hanya bisa mengirim binary untuk dua API ini atau meminta rollback ke rilis sebelumnya.
3. Di **Settings → Secrets and variables → Actions → New repository secret**, buat secrets berikut:

   | Secret | Nilai |
   |---|---|
   | `DEPLOY_HOST` | IP/host SSH server |
   | `DEPLOY_PORT` | Port SSH server |
   | `DEPLOY_USER` | `rabihutomo` |
   | `DEPLOY_KNOWN_HOSTS` | Baris host key server yang sudah diverifikasi, format `[host]:port ssh-ed25519 ...` |

4. Buat repository secret `DEPLOY_SSH_KEY` dengan **private key khusus** yang pasangannya dibatasi pada langkah 2. Jangan gunakan password MariaDB atau private key SSH pribadi yang aksesnya luas. Isi secret dengan konten file `~/.ssh/codeverta-api-ci`; private key jangan ditempel di issue, commit, atau chat.
5. Ulangi secrets pada repository kedua. Jika repository berada di organisasi yang sama, konfigurasi dapat dipusatkan sebagai organization secrets dengan akses hanya ke kedua repository ini.

Host key harus berasal dari `known_hosts` server yang sudah dipercaya, bukan hasil `ssh-keyscan` yang belum diverifikasi. Workflow menolak koneksi jika host key tidak cocok.

## Pemakaian

- Push perubahan API ke branch `main`: setelah test dan build lulus, API terkait deploy otomatis.
- Pull request: hanya test dan build; tidak mendapat deploy key dan tidak menyentuh server.
- Deploy ulang atau rollback: buka **Actions → Nihongo API CI/CD** atau **Fitapp API CI/CD → Run workflow**, pilih `deploy` atau `rollback`. Rollback beralih ke binary sebelumnya dan tetap mengecek health.

Sebelum mengganti binary atau melakukan rollback, helper membuat dump gzip database target memakai akun aplikasi dari `.env`/`DATABASE_DSN`. Backup disimpan di `/home/rabihutomo/apps/codeverta-api/backups/` dengan nama rilis dan waktu; backup lama tidak dibersihkan otomatis. Health endpoint Nihongo `http://127.0.0.1:18100/healthz`; Fitapp `http://127.0.0.1:18101/healthz`. Jika deployment baru gagal health check, helper mengembalikan symlink ke binary sebelumnya dan mencoba restart service tersebut lagi.

## Batas perubahan dan database

Nihongo hanya me-restart `nihongo-api.service`; Fitapp hanya me-restart `fitapp-api.service`. Deploy tidak me-restart MariaDB, Cloudflare Tunnel, Nginx, Malabar, atau NscBantuPerizinan. File `.env`, folder data, database, backup, dan rilis lama tidak dihapus atau ditimpa. Database tidak di-restore saat rollback. Backup bertambah pada setiap rilis/rollback dan perlu dipantau kapasitas disk; tidak ada pembersihan otomatis. Aplikasi tetap menjalankan migrasi schema normal saat startup; perubahan schema tetap perlu ditinjau sebelum merge karena rollback binary tidak membalikkan migrasi database.
