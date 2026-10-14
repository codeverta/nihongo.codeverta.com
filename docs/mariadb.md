# Database MariaDB

Backend menggunakan MariaDB untuk akun, pembayaran, token billing, XP,
progress lesson, achievements, cache respons kamus, dan arsip kata/kanji.
Backend memakai GORM dengan driver MySQL/MariaDB. Startup menjalankan
`AutoMigrate` dari model GORM untuk menyiapkan schema. Koneksi yang gagal
menghentikan startup.

## Konfigurasi lokal

Database lokal sudah dibuat pada server MariaDB yang berjalan di port 3306
(container `fitapp-mariadb-1`): database dan user aplikasi bernama
`nihongo_master`. Password acak tersimpan di `backend/.env` yang diabaikan Git.
Backend tidak menggunakan akun root. Data database mengikuti volume server
MariaDB tersebut; volume Compose backend hanya menyimpan branding.

Konfigurasi backend:

```dotenv
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=nihongo_master
DB_USER=nihongo_master
DB_PASSWORD=PASSWORD_DATABASE
DB_TLS=false
```

Jalankan aplikasi dan backend dari root project:

```sh
dart run.dart
```

Atau jalankan backend saja dari folder `backend`:

```sh
go run ./cmd/api
```

Di server lain, buat database/user menggunakan akun administrator MariaDB:

```sql
CREATE DATABASE nihongo_master CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
CREATE USER 'nihongo_master'@'%' IDENTIFIED BY 'GANTI_DENGAN_PASSWORD_KUAT';
GRANT ALL PRIVILEGES ON nihongo_master.* TO 'nihongo_master'@'%';
```

Isi `.env` dengan host, user, dan password server tersebut. Compose production
membaca `backend/.env`. Pada Linux dengan host networking, `127.0.0.1:3306`
menunjuk ke MariaDB host; pada bridge networking pakai alamat server database
yang dapat dijangkau container.

## Tabel dan proses API

Tujuh tabel: `accounts`, `billing_purchases`, `billing_tokens`, `deleted_demo`,
`gamification_users`, `dictionary_responses`, dan `dictionary_entries`.
Schema memakai InnoDB, `utf8mb4`, dan transaksi agar teks Jepang dan perubahan
account/billing tetap tersimpan dengan benar. Payload JSON lama disimpan dalam
LONGBLOB tanpa mengubah bentuk data aplikasi.

Semua tabel memiliki primary key `id` UUID dengan kolom `CHAR(36)`. GORM
membuat UUID v4 melalui hook `BeforeCreate`. ID akun sekaligus menjadi user ID;
relasi `user_id` pada progress, billing token, purchase, dan marker penghapusan
juga memakai UUID. Akun demo memakai UUID tetap agar tetap dikenali saat restart.

Email, cache key, dan pasangan sumber/entri tetap memiliki unique index untuk
mencegah duplikasi. Identitas transaksi toko ada di `purchase_key`, dan slug/ID
sumber kamus ada di `source_entry_id`. Keduanya dibutuhkan oleh layanan sumber.

Saat startup, migrasi UUID berjalan di bawah writer lock sebelum AutoMigrate.
ID user lama dipetakan konsisten, termasuk ID pada payload akun dan pemilik
pembelian. Migrasi dapat dijalankan ulang tanpa mengganti UUID record yang sudah
dibuat. Karena DDL MariaDB melakukan commit sendiri, simpan backup sebelum
upgrade schema. Backup lokal sebelum perubahan UUID ada di
`output/mariadb-migration/before-uuid-*.sql` dan diabaikan Git.

Model ketujuh tabel ada di `backend/internal/database/models.go`, dan koneksi
GORM berada di `backend/internal/database/gorm.go`. Operasi akun, billing,
gamifikasi, serta cache/arsip kamus menggunakan model GORM; upsert memakai
`clause.OnConflict`, dan penghapusan akun serta importer memakai transaksi
GORM. Nama tabel dan encoding payload sesuai data MariaDB yang sudah ada.
SQL khusus MariaDB tetap digunakan untuk writer lock dan pemangkasan cache.

Backend masih memegang cache akun/progress di memori. Jalankan satu proses API
per database: named writer lock MariaDB menolak proses kedua agar cache tidak
menimpa progress. Koneksi writer dijaga aktif; jika koneksi itu terputus, restart
API untuk memuat ulang data dan memperoleh lock. Session login dan attempt aktif
masih di memori, sehingga pengguna perlu login ulang setelah restart.

## Pemindahan data lama

Perpindahan lokal sudah dijalankan. File `backend/data/gamification.db` ditemukan
memiliki 0 record pada ketujuh tabel; database MariaDB sudah diinisialisasi.
File lama tetap disimpan untuk backup dan tidak dibaca backend.

Untuk database lama lain, hentikan API lalu jalankan dari root project:

```sh
python scripts/export_legacy_sqlite.py backend/data/gamification.db output/mariadb-migration/legacy-export.json
cd backend
go run ./cmd/import-legacy --input ../output/mariadb-migration/legacy-export.json
```

Exporter membaca SQLite lama dalam mode read-only menggunakan Python; backend
Go hanya memiliki driver MariaDB. Importer memindahkan ketujuh tabel dalam satu
transaksi, dan menolak database tujuan yang sudah berisi data. File export berisi
data privat dan diabaikan Git; simpan bersama backup bila diperlukan. Untuk
mengulang export, gunakan nama file output baru.

## Backup

Jalankan dengan client MariaDB; password akan diminta:

```sh
mariadb-dump --host=127.0.0.1 --port=3306 --user=nihongo_master --password --single-transaction nihongo_master > backup.sql
```

## Pengujian

Test persistence memakai schema MariaDB sementara dengan prefix
`nihongo_test_`, lalu menghapus schema tersebut setelah selesai. Set
`MARIADB_TEST_DSN` menggunakan akun yang dapat membuat dan menghapus schema
test itu, lalu jalankan `go test ./...` dari folder backend. Test juga membaca
nilai tersebut dari `.env.test` lokal. Jika kedua pengaturan tidak tersedia,
test MariaDB dilewati dan test unit tetap berjalan.

Pada setup lokal tersedia akun test terpisah dengan hak hanya untuk schema
`nihongo_test_*`. DSN-nya tersimpan di `backend/.env.test`; file ini tidak
di-commit. Jangan menggunakan database aplikasi sebagai schema test.

Referensi: [driver Go MySQL/MariaDB](https://github.com/go-sql-driver/mysql)
dan [koneksi GORM](https://gorm.io/docs/connecting_to_the_database.html)
dan [pengaturan image MariaDB](https://mariadb.com/docs/server/server-management/automated-mariadb-deployment-and-administration/docker-and-mariadb/mariadb-server-docker-official-image-environment-variables).
