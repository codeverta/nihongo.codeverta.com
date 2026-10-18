# Pengukuran kamus — 18 Oktober 2026

Implementasi: cache bersama permanen, normalisasi query sebelum fetch, mode Auto/Jepang/Inggris, refresh admin atomik, single-flight dengan pembatalan independen, dan cache memori maksimal 128 respons / 16 MiB.

## Hasil di lingkungan development

Windows amd64, AMD Ryzen 5 4600G; Go benchmark dan MariaDB test schema terisolasi. Pengukuran tidak mencakup jaringan Android ke backend, waktu render layar, atau beban produksi.

| Jalur | Metode | Hasil |
|---|---|---|
| Hit memori | `BenchmarkSearchMemory`, 2 detik, 441.482 iterasi; mencakup normalisasi katakana dan parse hasil | 7.156 ns/op ≈ 7,16 µs; 1.408 B/op; 23 allocs/op |
| Hit MariaDB | 120 pencarian, memori dikosongkan sebelum setiap request; query katakana memakai respons kanonis tersimpan | p50 44,18 ms; p95 47,47 ms; 0 panggilan API tambahan |
| Cold fixture | HTTP transport fixture dengan delay buatan 20 ms, ditambah lookup miss dan transaksi MariaDB | 191,51 ms; bukan latensi Jisho asli |
| Pengguna bersamaan | 24 pemanggil memakai variasi kana, setelah pemanggil awal romaji membatalkan | Hanya 1 request upstream; seluruh pemanggil yang masih menunggu berhasil |

Run DB sebelumnya menghasilkan p95 65,74 ms; hasil bergantung pada kondisi DB dan mesin. Nilai per-request memori dari `time.Now()` terbaca nol pada timer Windows ini, sehingga angka memori dilaporkan dari benchmark agregat, bukan mengklaim latensi nol.

## Verifikasi sumber asli

`TestLiveCanonicalDatabaseReuse` menggunakan Jisho asli dan schema MariaDB test baru:

1. Klien pertama mencari `konnichiwa`; backend mengirim query kanonis `こんにちは` dan mendapat 3 entri.
2. Klien kedua memakai koneksi/client baru, tanpa cache memori klien pertama. Transport HTTP eksternalnya dinonaktifkan oleh tes.
3. Query `こんにちは`, `コンニチハ`, dan `ｺﾝﾆﾁﾊ` mengembalikan entri identik. Klien kedua mencatat 1 hit DB, lalu hit memori untuk variasi berikutnya, serta 0 panggilan API.

Tidak ada bulk crawl. Jumlah entri dapat berubah pada sumber; tes membandingkan kesamaan hasil, bukan mengunci jumlahnya menjadi tiga.

## Cakupan pemeriksaan

- Normalisasi huruf besar/lebar/Unicode/kana, greeting, romaji umum, `matcha`, `Tōkyō`, dan macron; pemisahan homofon, vokal panjang, mode, halaman, dan filter.
- Hasil lama dan hasil kosong tetap permanen setelah cache memori dihapus; respons di atas 5.000 tetap dipertahankan.
- Refresh gagal karena API maupun transaksi DB mempertahankan hasil sebelumnya; refresh sukses mengganti DB dan memori.
- Request bersama tidak dibatalkan oleh pemanggil pertama; byte budget memori diperiksa.
- Token admin, pagination/validasi, metadata daftar, dan refresh lewat endpoint admin.
- Integrasi koneksi MariaDB aplikasi dan migrasi UUID yang sudah ada.
- 9 tes Flutter kamus dan 5 widget test aplikasi lulus; analyzer tidak menemukan masalah. Tes ponsel memeriksa mode, teks input terkini, hasil pencarian, favorit, dan detail.
- Sintaks JavaScript halaman admin lolos `node --check`.

Go race detector tidak dijalankan: lingkungan Windows ini memakai `CGO_ENABLED=0` dan tidak memiliki GCC. Tes konkurensi deterministik tetap dijalankan.

## Menjalankan ulang

Dari `backend/`, dengan MariaDB test DSN seperti setup proyek:

```powershell
go test ./internal/dictionary -count=1 -v
go test ./internal/api -run TestDictionary -count=1 -v
go test ./internal/store -run TestDictionary -count=1 -v
go test ./internal/dictionary -run '^$' -bench BenchmarkSearchMemory -benchmem -benchtime=2s
$env:DICTIONARY_LIVE_TEST = '1'
go test ./internal/dictionary -run TestLiveCanonicalDatabaseReuse -count=1 -v
```

Dari root proyek:

```powershell
flutter test --no-pub test/features/dictionary/dictionary_test.dart test/widget_test.dart
flutter analyze --no-pub lib/features/dictionary test/features/dictionary/dictionary_test.dart
```

Untuk romaji di luar daftar Auto yang sudah ditinjau, pilih **Jepang**. Auto sengaja mempertahankan kata Latin ambigu; ia tidak menebak semua kata Latin sebagai bahasa Jepang. Statistik admin berlaku untuk proses server saat ini dan direset ketika server restart.
