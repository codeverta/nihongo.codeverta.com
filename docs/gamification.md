# Gamifikasi dan achievements

Profil dan halaman **Achievements** memakai `GET /v1/me/gamification`, bukan angka contoh. Home, profil, dan achievements membaca progres yang sama. Setelah kuis atau ujian selesai, aplikasi memperbarui data tersebut. Halaman achievements menyediakan filter, tanggal diperoleh, refresh, serta retry ketika permintaan gagal. Data akun sebelumnya tidak dipakai setelah pergantian akun.

## Aturan server

- Lesson vocabulary, grammar, kanji, dan reading mendapat **50 XP** hanya saat pertama lulus (nilai ≥80%).
- Diagnostic atau simulasi mendapat **100 XP** hanya saat pertama memperoleh ≥80% pada setiap ID ujian. Mengulang paket yang sama tidak menggandakan XP.
- Level dimulai dari 1 dan naik setiap 500 XP.
- Kuis yang selesai atau ujian yang selesai dengan setidaknya satu jawaban mencatat hari aktif, termasuk saat nilainya belum lulus. Ujian kosong tidak menghasilkan hari aktif, XP, atau badge.
- Streak menggunakan hari kalender **Asia/Jakarta (UTC+7)**. Streak kemarin tetap terlihat sampai hari ini selesai; setelah melewatkan satu hari penuh, streak menjadi 0. Streak terbaik dan badge yang diperoleh tetap tersimpan.
- Dua belas achievements mencakup lesson pertama, 10/50 lesson, kosakata, grammar, kanji, reading, empat bidang, streak 7/30 hari, 500 XP, dan skor ujian 80%.
- API progress manual tidak memberikan XP atau badge. Completion lesson yang tersedia harus berasal dari hasil kuis yang dinilai server, termasuk vocabulary.

`minutesToday` menghitung durasi sesi kuis/ujian yang selesai berdasarkan waktu server, dibatasi maksimal dua jam per sesi dan satu hari per tanggal. Ini bukan pelacakan waktu aktif membaca. `examsCompleted` menghitung jumlah ID jenis ujian berbeda yang pernah diselesaikan dengan jawaban.

## Penyimpanan dan operasi

Startup API wajib membuka MariaDB melalui konfigurasi `DB_*` pada
`backend/.env` (port default 3306). Kegagalan membuka atau membaca database
menghentikan startup. Setup dan pemindahan data ada di [panduan MariaDB](mariadb.md).

XP, event idempotency, hak reward per lesson/ujian, hari aktif, timestamp unlock, serta best score/completion lesson disimpan atomik sebelum respons completion berhasil. Jika penulisan gagal, API mengembalikan 500 tanpa memberi hasil sukses atau reward parsial; request dapat dicoba lagi. Saat restart, progres lesson yang terverifikasi ikut dipulihkan, sehingga mengulang lesson tidak menghasilkan reward baru.

Gunakan **satu proses API per database**. Named writer lock MariaDB menolak
proses API kedua untuk melindungi cache progress. Backup menggunakan
`mariadb-dump`, bukan salinan file lokal. Test persistence memakai schema
MariaDB sementara yang terpisah dari database aplikasi.

Lingkup persistensi ini adalah gamifikasi dan progres lesson terverifikasi. Sesi login, attempt yang sedang berjalan, serta detail riwayat ujian masih mengikuti memory store aplikasi. Setelah restart, pengguna perlu masuk kembali; progres dan rewards tetap melekat pada ID akun yang sama. Ini belum merupakan migrasi seluruh backend ke penyimpanan permanen.

Referensi implementasi database: [driver Go MySQL/MariaDB](https://github.com/go-sql-driver/mysql).

## Verifikasi

Backend menguji empat alur lesson nyata, skor ujian, retry, request bersamaan, pergantian hari Jakarta, streak terputus, kepemilikan data, restore setelah restart, dan kegagalan database. Flutter menguji API/model, statistik nyata, navigasi, filter, layout ponsel, retry, akses guest, dan refresh setelah pergantian akun.
