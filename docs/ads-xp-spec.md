# Penempatan iklan, XP, dan double XP

Status: **disetujui pemilik produk (sesi grilling 20 November 2026). Tahap A selesai di kode (belum di-commit dan belum dicoba di perangkat); tahap B dan C belum dikerjakan.** Melengkapi [rewarded-unlock-spec.md](rewarded-unlock-spec.md).

## Hasil tahap A

- `InterstitialController` tidak lagi dipicu oleh pindah rute atau tab. Ada dua titik baru: `prepareAfterCompletion` (memuat iklan saat layar hasil tampil) dan `showAfterCompletion` (menampilkan iklan saat layar hasil ditinggalkan), dengan sakelar per jenis dan daftar penyelesaian yang sudah "dibayar" dengan rewarded.
- `CompletionAdBoundary` (`lib/core/ads/completion_ads.dart`) dipasang di:
  - hasil kuis grammar, kanji, membaca, dan kosakata;
  - hasil simulasi (termasuk JFT, NAT, dan SSW) dan diagnostik;
  - akhir latihan SSW;
  - lesson Kaigo setelah semua jawaban tepat;
  - sesi Anki, setelah ambang kartu tercapai atau semua kartu jatuh tempo habis.
- Sheet buka konten me-reset jam jeda setiap kali rewarded tampil, dan menandai item yang dibukanya supaya tidak mendapat iklan layar penuh di akhir.
- Penempatan baru: banner tab Ujian, banner akhir Anki/SSW/Kaigo, dan native di akhir artikel berita (tidak untuk artikel yang dibuka dengan iklan).
- Config baru (backend dan aplikasi): `interstitialKinds`, `examTabBanner`, `practiceResultBanner`, `newsArticleNative`, `ankiSessionMinCards`. Simpan dari dashboard lama tidak mereset nilai-nilai ini.
- Dashboard Iklan:
  - kartu baru;
  - panduan "Cara membuat unit iklan ini" di setiap kartu (format, nama unit, langkah AdMob, langkah SSV untuk buka konten, unit yang boleh dibagi);
  - pilihan "Pakai ID yang sama dengan …" untuk format yang sama (tidak untuk rewarded);
  - kartu "Kapan layar penuh boleh muncul" berisi empat sakelar dan ambang sesi Anki.
  Diverifikasi di server pratinjau lokal.
- Ditemukan dan diperbaiki: layar hasil yang ikut dibuang bersama aplikasi (keluar akun, restart) sempat mencoba menampilkan iklan setelah provider-nya dibuang, sehingga muncul `StateError`. Tes regresinya sudah ada.
- Tes: tes iklan layar penuh ditulis ulang untuk perilaku baru, ditambah tes `CompletionAdBoundary`. Suite Flutter penuh menyisakan 20 kegagalan yang sama persis dengan HEAD; tidak ada kegagalan baru. Backend lulus semua.

## Masalah yang ditemukan

- **Iklan layar penuh muncul saat membuka lesson.** Pemicunya setiap pindah rute atau tab (`app_router.dart` memanggil `onNaturalBreak` pada setiap perubahan rute; `main_scaffold.dart` pada setiap ketukan tab). Pengecualiannya hanya `/`, onboarding, login, dan premium. Jadi iklan bisa menutupi lesson, sesi Anki, latihan SSW, bahkan ujian bertimer yang waktunya sudah berjalan di server.
- Di akhir lesson justru tidak ada iklan layar penuh, karena hasil kuis hanya berganti tampilan di layar yang sama. Akhir sesi Anki, latihan SSW, dan lesson Kaigo tidak punya iklan sama sekali.
- **Iklan layar penuh bisa muncul tepat setelah rewarded pembuka konten**, karena alur unlock tidak me-reset jam jeda (hanya Tanya AI yang melakukannya).
- **XP hanya diberikan sekali seumur hidup per lesson (50) atau ujian (100), dan hanya bila nilai ≥80%.** Mengulang atau gagal memberi 0 XP.
- Latihan SSW, lesson Kaigo (hanya tercatat di HP), dan Anki tidak memberi XP maupun streak.
- **Bug streak**: lesson yang disinkronkan offline dicatat pada hari sinkronisasi, bukan hari dikerjakan.
- Tidak ada tampilan "+N XP" di layar hasil mana pun.

## Keputusan

### Iklan pasif

1. Tidak ada iklan pasif selama belajar, sejak membuka lesson sampai kuis selesai.
2. **Iklan layar penuh hanya muncul saat pengguna keluar dari layar hasil**, setelah nilainya sempat terlihat. Pemicu "pindah rute/tab" dihapus total.
3. Titik selesai yang dihitung:
   - lesson JLPT (grammar, kanji, membaca, kosakata) sampai layar hasil kuis;
   - simulasi dan diagnostik JLPT, JFT, NAT, dan ujian SSW sampai layar hasil;
   - satu sesi latihan SSW atau lesson Kaigo;
   - satu sesi Anki: semua kartu jatuh tempo habis, atau sesi ditutup setelah minimal N kartu (N = 10, bisa diatur di admin).
4. Frekuensi mengikuti aturan jeda yang sudah ada: minimal N menit sejak iklan layar penuh terakhir (default 10, bisa diatur di admin).
5. Tidak ada iklan layar penuh saat masuk ujian bertimer, saat kembali dari halaman Lifetime, atau setelah pengguna baru saja menonton rewarded (pembuka konten, double XP, atau Tanya AI). **Setiap rewarded me-reset jam jeda.**
6. Lesson yang dibuka dengan iklan pembuka tidak mendapat iklan layar penuh di akhir.
7. Banner di layar hasil kuis dan ujian tetap ada. Ditambah:
   - banner di akhir sesi Anki dan akhir latihan SSW/Kaigo;
   - banner di tab Ujian (tab Profil tetap tanpa iklan);
   - satu native di akhir artikel berita, untuk pengguna gratis (tidak untuk artikel yang dibuka dengan iklan pembuka).
8. Tab Kamus tetap banner dan native.

### Iklan pembuka konten

Tetap di awal: rewarded opsional untuk N3–N1, Anki, simulasi N3, berita, dan lowongan (lihat rewarded-unlock-spec.md).

### Admin

1. Setiap penempatan punya kartu sendiri dengan sakelar on/off.
2. Setiap kartu punya panduan yang bisa dibuka, "Cara membuat unit iklan ini":
   - di mana iklan muncul dan kapan;
   - format AdMob (Banner, Native advanced, Interstitial, Rewarded);
   - nama unit yang disarankan;
   - langkah membuat unit;
   - catatan khusus (SSV wajib beserta URL callback-nya, jangan centang partner bidding);
   - unit mana yang boleh dipakai bersama.
3. Tombol "Pakai ID yang sama dengan …" untuk penempatan yang boleh berbagi unit.
4. Iklan layar penuh: satu unit, satu aturan jeda, empat sakelar (Lesson JLPT, Ujian, Latihan SSW & Kaigo, Sesi Anki).
5. Penempatan baru memakai ulang `nm_banner` dan `nm_native`, dengan sakelar dan kartu sendiri.
6. Bisa diatur di admin: batas XP pengulangan per hari, batas double XP per hari, ambang sesi Anki, semua sakelar iklan, dan aturan jeda iklan layar penuh. Angka XP dasar tetap di kode.

### XP dan streak

| Aktivitas | XP | Streak |
|---|---|---|
| Lesson JLPT selesai | 10 setiap selesai, nilai berapa pun, +40 bonus saat pertama kali ≥80% (total lulus pertama tetap 50) | Ya |
| Ujian selesai (simulasi, diagnostik, JFT, NAT, SSW) | 20 setiap selesai, +80 bonus saat pertama kali ≥80% (total tetap 100) | Ya |
| Latihan SSW selesai (semua soal dijawab) | 10 | Ya |
| Lesson Kaigo selesai | 5, sekali per lesson | Ya |
| Anki | 10 sekali per deck per hari, setelah 10 kartu direview hari itu | Ya |

- XP pengulangan (semua XP di luar bonus lulus pertama dan Kaigo) dibatasi **200 XP per hari**, bisa diatur di admin.
- Tidak berlaku mundur.
- Lesson offline dicatat pada hari dikerjakan menurut waktu HP, tidak boleh di masa depan dan maksimal 7 hari ke belakang.
- Setiap layar hasil menampilkan "+N XP" dan streak saat ini. Server mengembalikan XP yang diperoleh di respons penyelesaian.

### Double XP

1. Tombol opsional di layar hasil, "Tonton iklan untuk 2× XP (+N XP)". Tidak pernah muncul otomatis, dan hanya tampil bila penyelesaian itu memberi XP lebih dari 0.
2. Yang digandakan adalah seluruh XP dari penyelesaian itu, termasuk bonus lulus pertama.
3. **Diverifikasi server lewat SSV**, sekali per penyelesaian, dan bisa diklaim maksimal 30 menit setelah selesai.
4. Unit baru `nm_rewarded_xp` (Rewarded, SSV ke `https://nihongo.codeverta.com/v1/ads/ssv`).
5. Maksimal 10 per hari, bisa diatur di admin. Batas ini terpisah dari batas buka konten. XP penggandaan tidak masuk batas 200 XP pengulangan.
6. Tidak ada jalur cadangan: bila iklan gagal dimuat, pengguna bisa mencoba lagi.
7. Pengguna Lifetime bisa menekan "Klaim 2× XP" tanpa iklan, dengan batas harian yang sama.
8. Setelah double XP diklaim, iklan layar penuh saat keluar dari layar hasil itu dilewati, dan jam jeda di-reset.
9. Kartu sendiri di admin, dengan sakelar per jenis aktivitas dan panduan.

### Ditunda

- Menampilkan target harian. Saat ini dihitung server tetapi tidak ditampilkan, dan menit dari lesson offline tercatat 0.
- Iklan rewarded untuk mempertahankan streak yang hampir putus (bisa memakai mekanisme SSV yang sama).

## Rencana teknis

### Tahap A: penempatan iklan

**Aplikasi**
- `InterstitialController`:
  - hapus pemanggilan dari listener router dan ketukan tab;
  - tambah `prepareAfterCompletion(kind)`, dipanggil saat layar hasil tampil, untuk memuat iklan lebih dulu bila jedanya hampir habis;
  - tambah `showAfterCompletion(kind)`, dipanggil saat pengguna keluar dari layar hasil, yang memeriksa sakelar jenis, jam jeda, dan penanda "baru menonton rewarded".
- Jenis titik selesai: `lessons`, `exams`, `practice` (SSW/Kaigo), `anki`.
- `UnlockFlow` (dan nanti double XP) memanggil `markFullScreenAdShown()` setiap kali rewarded ditampilkan.
- Lesson yang dibuka dengan iklan diberi penanda di sesi, sehingga `showAfterCompletion` melewatinya.
- Layar hasil yang memanggil `showAfterCompletion` saat keluar:
  - lesson JLPT (tombol kembali dan "Kembali ke daftar lesson");
  - hasil simulasi dan diagnostik;
  - akhir latihan SSW;
  - lesson Kaigo selesai;
  - akhir sesi Anki (hanya bila ≥ ambang kartu atau semua kartu jatuh tempo habis).
- Banner baru di tab Ujian, layar akhir Anki, latihan SSW, dan Kaigo, serta native di akhir artikel berita.

**Backend dan admin**
- `AdsConfig` bertambah:
  - `interstitialKinds {lessons, exams, practice, anki}`;
  - `examTabBanner`;
  - `practiceResultBanner` (akhir Anki/SSW/Kaigo);
  - `newsArticleNative`;
  - `ankiSessionMinCards` (default 10).
  Simpan dari dashboard lama tidak mereset nilai-nilai ini.
- Dashboard: kartu baru, panduan per kartu (`<details>`), dan tombol "Pakai ID yang sama dengan …".

### Tahap B: XP dan streak

**Backend** (`store/gamification.go` dan pemanggilnya)
- `recordStudyLocked`: XP dasar per penyelesaian, bonus lulus pertama, dan batas harian XP pengulangan dihitung per hari studi.
- Jenis event baru:
  - `ssw-practice` kini memberi XP dan menghitung streak;
  - `kaigo` lewat endpoint baru `POST /v1/me/kaigo-lessons/{lessonId}/complete`;
  - `anki` dicatat di `AnkiReview` saat review deck hari itu mencapai ambang.
- Sinkron offline memakai waktu pengerjaan dari klien, dalam batas yang aman.
- Respons penyelesaian (lesson, sinkron offline, simulasi, diagnostik, SSW, Kaigo, review Anki yang melewati ambang) menyertakan `xp: {awarded, total, streak, completionId, doubleable}`.

**Aplikasi**
- Widget bersama "+N XP · Streak N hari" di semua layar hasil.
- Kaigo memanggil endpoint baru saat lesson selesai.
- Perubahan XP me-refresh dashboard dan gamifikasi.

### Tahap C: double XP

**Backend**
- Jenis unlock baru `xp-boost` dengan `itemId` = `completionId`.
- Klaim menambah bonus XP sebesar XP penyelesaian itu, sekali saja, dalam 30 menit, dan tunduk pada batas harian sendiri.
- Endpoint `POST /v1/xp-boosts/{completionId}` untuk Lifetime tanpa iklan, dengan batas yang sama.
- Statistik admin mendapat kelompok "Double XP".

**Aplikasi**
- Tombol double XP di widget hasil memakai `UnlockFlow` dengan unit `xpBoostRewarded`.
- Lifetime melihat tombol "Klaim 2× XP" tanpa iklan.

**Admin**
- Kartu "Rewarded 2× XP": sakelar, unit, sakelar per jenis aktivitas, batas harian, dan panduan pembuatan `nm_rewarded_xp` dengan SSV.

## Risiko

- **Pendapatan iklan layar penuh bisa turun** karena pemicunya kini jauh lebih jarang (hanya di titik selesai). Tuasnya: menurunkan menit jeda di admin. Pantau statistik AdMob setelah rilis.
- **Aplikasi versi lama** tetap memakai pemicu "pindah rute" sampai pengguna memperbarui. Perubahan XP berlaku di server untuk semua versi, tetapi "+N XP" hanya tampil di versi baru.
- **XP Kaigo tidak bisa diverifikasi**: nilainya kecil dan hanya sekali per lesson.
- **Batas XP harian** bisa terasa membatasi bagi pengguna yang sangat aktif. Batasnya bisa dinaikkan di admin.
