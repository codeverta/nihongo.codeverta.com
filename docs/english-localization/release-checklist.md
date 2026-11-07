# Checklist rilis English

Gerbang English saat ini **mati** untuk publik. Jangan menyalakannya sebelum semua blocker di bawah selesai dan paket review disetujui.

## Gerbang

| Lapisan | QA internal | Rilis publik |
| --- | --- | --- |
| Flutter | `--dart-define=ENGLISH_QA_PREVIEW=true` pada build **debug** (di build release tidak berlaku) | `--dart-define=ENGLISH_PUBLIC_RELEASE=true` |
| Backend (konten belajar, akun, berita) | `ENGLISH_QA_PREVIEW_ENABLED=true` di server QA privat | `PUBLIC_ENGLISH_CONTENT_ENABLED=true` |
| Halaman publik `/en/...` | ikut flag backend publik | ikut flag backend publik |

Aplikasi mengirim `Accept-Language` sesuai bahasa pilihan; tanpa flag di backend, permintaan English dijawab Indonesia.

## Blocker sebelum rilis publik

1. **Privasi dan analytics.** `legal/privacy.html`, `legal/en/privacy.html`, `legal/safety.html`, `legal/en/safety.html` menyatakan tidak ada behavioral analytics. Perbarui (Firebase Analytics, hanya Android, data penggunaan anonim, bisa dimatikan di Pengaturan) dan isi ulang formulir Data safety di Google Play. Usulan teks dan jawaban ada di `privacy-update-proposal.md`.
2. **Kode belum dikompilasi dan dites.** Perubahan Wave 1 dan 2, analytics, dan gambar berita ditulis tanpa Flutter/Go di lingkungan kerja. Jalankan `flutter pub get`, `flutter analyze`, `flutter test`, lalu `go build ./... && go test ./...` di `backend`. Perbaiki error kompilasi atau test yang gagal sebelum lanjut.
3. **Review manusia** untuk semua area di `review-package.md` selesai dan tabel persetujuan terisi.
4. **Keputusan gambar berita.** Putuskan dan dokumentasikan kebijakan gambar tanpa lisensi (lihat paket review).
5. **Notifikasi push.** Judul dan isi ditulis admin dalam satu bahasa. Putuskan: kirim hanya ke pengguna Indonesia, kirim dwibahasa, atau pisahkan topik per bahasa. Nama dan deskripsi channel notifikasi Android masih Indonesia.
6. **iOS.** Push dan analytics belum dikonfigurasi (tidak ada `GoogleService-Info.plist`). Jika English dirilis di iOS, pastikan tidak ada klaim yang menyiratkan fitur itu aktif.

## Pengujian teknis

- `flutter test` termasuk `test/core/localization/ui_english_coverage_test.dart` (gagal bila ada teks `.localized` tanpa padanan English).
- `go test ./internal/api/...` termasuk `localization_account_middleware_test.go`, `localization_errors_test.go`, `localization_account_test.go`, dan tes English yang sudah ada.
- Dengan `ENGLISH_QA_PREVIEW_ENABLED=true`, uji dengan `curl -H "Accept-Language: en"` pada: `/v1/me/achievements`, `/v1/plans`, `/v1/jobs/meta`, `/v1/anki/decks`, `/v1/app-update-policy`, serta satu respons error (mis. `/v1/me/profile` tanpa token). Hasilnya harus English; dengan `Accept-Language: id` harus tetap Indonesia.
- Unduhan (ekspor deck Anki, gambar) tidak boleh berubah dengan `Accept-Language: en`.

## Uji manual (build debug + server QA)

Gunakan perangkat Android dengan bahasa sistem English **dan** satu dengan Indonesia.

1. Instalasi baru, bahasa sistem English → app mulai dalam English; bahasa sistem Indonesia → mulai Indonesia.
2. Pembaruan dari versi lama: instalasi lama tetap Indonesia sampai pengguna memilih sendiri.
3. Pengaturan → Tampilan → ganti bahasa; semua layar berubah tanpa restart dan tidak ada campuran Indonesia di layar yang sama.
4. Ganti bahasa lalu buka ulang materi yang pernah dibuka: konten tidak boleh tetap dalam bahasa lama (cache dan unduhan offline).
5. Telusuri: onboarding, beranda, Belajar (kosakata, kanji, grammar, reading N5–N1), Ujian (diagnostik, simulasi JLPT/NAT/JFT), SSW (beberapa sektor termasuk Kaigo), kamus, Anki, berita, lowongan, manga, profil, Pengaturan, Premium, notifikasi, update paksa.
6. Pemicu error: matikan jaringan, login salah, kirim ujian dengan soal kosong; pesan harus English.
7. Premium: harga tampil dari produk store sesuai wilayah; alur beli dan restore dalam English.
8. Teks Jepang dan furigana tetap utuh di mode English.
9. Latensi dari luar Indonesia (backend masih di Indonesia): catat waktu muat beranda dan satu pelajaran.
10. Pengaturan → "Bantu tingkatkan aplikasi" mati → tidak ada event masuk di Firebase DebugView.

## Urutan rilis

1. Deploy backend (flag tetap mati) dan jalankan `news-image-migrate` serta `news-sync -force` bila perlu.
2. Rilis internal/closed testing dengan build debug QA atau flag publik di build uji terpisah; selesaikan uji manual.
3. Dapatkan persetujuan reviewer dan reviewer hukum; perbarui halaman legal dan Data safety.
   - Iklan: bila ingin menyalakan iklan, selesaikan bagian "Tambahan wajib jika iklan diaktifkan" di `privacy-update-proposal.md` (akun AdMob, ID asli, UMP, deklarasi Play) sebelum menyalakan saklar di `/admin/ads`.
4. Nyalakan `PUBLIC_ENGLISH_CONTENT_ENABLED=true` di backend.
5. Rilis build aplikasi dengan `ENGLISH_PUBLIC_RELEASE=true` (rollout bertahap di Play Console).
6. Pantau 7 hari: crash, error API 4xx/5xx menurut `Accept-Language`, latensi, dan metrik `app_language` di analytics.

## Rollback

- Matikan `PUBLIC_ENGLISH_CONTENT_ENABLED` di backend: permintaan English kembali dijawab Indonesia.
- Hentikan rollout di Play Console; pengguna yang sudah memilih English tetap dapat memakai app karena teks UI tersimpan di klien. Rilis build dengan gerbang mati bila perlu mengembalikan semua ke Indonesia.

## Setelah rilis

Jalankan ulang ekspor register setiap kali katalog berubah (`README.md` folder ini) dan tinjau ulang pernyataan fakta di halaman publik bila perilaku app, penyedia, izin, produk pembelian, atau retensi data berubah.
