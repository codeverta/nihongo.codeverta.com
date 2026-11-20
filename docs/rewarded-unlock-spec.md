# Monetisasi: buka konten dengan iklan rewarded atau Lifetime

Status: **fase 1 (backend inti untuk pelajaran) sudah ditulis dan lulus tes; belum di-deploy dan aplikasi belum diubah.** Keputusan produk sudah dikonfirmasi (lihat "Keputusan"). Dokumen ini berisi hasil fase 1, temuan kondisi sekarang, rencana teknis, risiko, dan urutan rilis.

## Hasil fase 1 (backend, pelajaran)

Sudah ada: paket `internal/unlock` (repositori memori dan MariaDB, verifier SSV, `GoogleKeys`, `Service`), endpoint `POST /v1/unlocks/intents`, `POST /v1/unlocks/claim`, `POST /v1/unlocks/fallback`, `GET /v1/unlocks`, dan `GET /v1/ads/ssv`, gerbang pada `authorizePremium` untuk pelajaran grammar, kanji, membaca, dan kosakata (termasuk attempt kuis dan tutor AI pelajaran itu), serta config `unlockRewarded`, `unlockDailyCap`, `unlockFallbackPerDay` di `/v1/ads-config`. Tabel baru dibuat otomatis saat start: `unlock_intents`, `content_unlocks`, `unlock_usage`.

Perbedaan dari rencana awal:
- **Katalog memakai `GET /v1/unlocks`, bukan field `locked`.** Katalog `/course` N3–N1 dibuka tanpa mengubah payload; aplikasi menghitung gembok dari level pelajaran dan daftar kunci `kind:itemId` di `/v1/unlocks`. Katalog hanya berisi judul dan progres, bukan isi pelajaran.
- **Header `X-Access: ad-unlock` belum dipasang.** Kebijakan cache offline ditangani di sisi aplikasi pada fase 4.
- Respons 402 kini membawa `unlockable: true` dan `unlock: {kind, itemId}` bila iklan bisa ditawarkan, supaya aplikasi tahu apa yang dibuka.
- **Simpan dari dashboard lama aman.** Dashboard iklan yang belum mengenal pengaturan unlock tidak lagi mereset `unlockRewarded`, `unlockDailyCap`, dan `unlockFallbackPerDay` saat menyimpan.
- Callback SSV yang sah tetapi tidak bisa diterapkan (nonce tidak dikenal, akun tidak cocok, transaksi dipakai ulang) dijawab 200, karena Google mengulang bila bukan 200 dan ulangan tidak mengubah hasil. Hanya tanda tangan salah (400) dan kegagalan penyimpanan (500) yang berupa error.

Temuan saat menguji di MariaDB nyata: klaim bersamaan sempat kena `Deadlock found (1213)`. Penanganan pertama (mengulang transaksi) hanya menutup gejala, jadi diganti dengan aturan penguncian yang membuat deadlock tidak bisa terbentuk:
1. Setiap transaksi tulis mengunci **baris pemakaian akun untuk hari itu lebih dulu** (lewat primary key). Semua penulisan untuk satu akun berjalan bergiliran dengan urutan kunci yang sama.
2. Isolasi **READ COMMITTED**: tidak ada gap lock, sehingga akun yang berbeda tidak saling menunggu.
3. Bila dua transaksi tetap menyimpan unlock yang sama (misalnya tepat di pergantian hari WIB), yang kalah di-rollback dan mengembalikan unlock yang sudah ada, tanpa dihitung dua kali dan tanpa error 500.

Pengulangan hanya tersisa sebagai cadangan terakhir. Tes MariaDB menjalankan **semua skenario dengan pengulangan dimatikan**, sehingga satu deadlock saja membuat tes gagal. Skenario stres mencakup 40 klaim bersamaan, 8 akun bersamaan, klaim dan cadangan bercampur, serta 12 permintaan bersamaan untuk pelajaran yang sama. Saat kunci baris pemakaian sengaja dilepas, empat skenario langsung gagal (batas terlampaui). Setelah dipasang kembali, tes lulus 5 kali berturut-turut.

**Aman di-deploy sekarang.** Selama `NIHONGO_OPEN_ACCESS` tidak diatur `false`, semua N3–N1 tetap terbuka dan tidak ada perilaku yang berubah bagi pengguna. Yang bertambah hanya endpoint baru, tiga tabel, dan `/v1/ads/ssv`. Gerbang baru berlaku setelah langkah rilis (lihat "Risiko rilis").

Syarat operasional sebelum SSV bisa bekerja:
1. Server harus bisa **keluar** ke `https://www.gstatic.com` (kunci verifikasi Google), dan `/v1/ads/ssv` harus bisa dijangkau Google dari luar (jangan ditantang oleh aturan bot Cloudflare).
2. Unit AdMob `nm_rewarded_unlock` harus dibuat dengan callback `https://nihongo.codeverta.com/v1/ads/ssv`, lalu ID-nya diisi ke `unlockRewarded.unitId` (default sementara memakai unit uji Google).
3. Config default memakai unit uji Google dan jatah 20 per hari, 2 cadangan per hari.

## Konsep

Beberapa jenis konten menjadi **eksklusif**. Pengguna gratis punya dua jalan:

1. **Lifetime Access**: beli sekali, materi tanpa iklan, semua terbuka. Tanya AI 15 pesan per hari, lalu iklan rewarded opsional untuk 5 pesan lagi.
2. **Tonton satu iklan rewarded** untuk membuka satu item (Android saja).

| Konten | Satuan unlock |
|---|---|
| Pelajaran N3, N2, N1 (grammar, kanji, kosakata, membaca) beserta quiz-nya | per pelajaran, permanen |
| Deck Anki JLPT N3, N2, N1 | per deck, permanen |
| Simulasi N3 | per percobaan (setiap mulai simulasi = satu iklan) |
| Membaca satu artikel berita | per artikel, permanen |
| Melamar satu lowongan (mendapat tautan lamar) | per lowongan, permanen |

N5 dan N4 tetap gratis. Pengguna Premium/Lifetime tidak pernah melihat iklan. **iOS hanya Lifetime** (tidak ada iklan di iOS).

Rewarded harus **opsional dan jelas imbalannya** (syarat kebijakan AdMob): pengguna memilih sendiri "Tonton iklan untuk membuka ini", dan selalu ada tombol Lifetime serta Batal. Tidak ada iklan yang muncul otomatis.

## Keputusan (dikonfirmasi pemilik produk)

| # | Keputusan |
|---|---|
| 1 | Unlock **per pelajaran**, permanen untuk akun itu. |
| 2 | **Anki N3–N1 dan simulasi N3 ikut** digembok, bisa dibuka dengan iklan. |
| 3 | **iOS hanya Lifetime.** |
| 4 | **Langsung berlaku** untuk pengguna lama, tanpa masa transisi konten. (Urutan teknis rilis tetap wajib, lihat "Risiko rilis".) |
| 5 | Batas **20 unlock per akun per hari** dan **2 unlock cadangan per hari** (bila iklan tidak tersedia). |
| 6 | **Berita**: tiap membuka satu artikel, tawarkan iklan rewarded atau Lifetime. |
| 7 | **Melamar lowongan**: sekarang hanya Lifetime, ditambah opsi iklan rewarded. |

Usulan yang masih bisa diubah (tidak ditanyakan secara eksplisit): satuan simulasi per percobaan, deck Anki per deck permanen, dan satu unit AdMob untuk semua jenis konten.

## Hasil fase 2 (backend: Anki, simulasi, berita, lowongan)

Sudah ada, dan seluruh suite backend lulus:
- **Anki N3–N1**: satu iklan membuka satu deck bawaan, permanen. Gerbang di `authorizePremium` dan di store (`AnkiDeckFor`, `AnkiStudyFor`, `AnkiReviewFor`) sama-sama menghormati unlock. Deck impor tidak terpengaruh.
- **Simulasi**: satu iklan = satu percobaan. Iklan memberi kredit `simulation:<examId>`; saat percobaan dimulai, kredit itu ditukar menjadi `simulation-attempt:<attemptId>` (atomik, satu kredit hanya bisa dipakai sekali). Melanjutkan percobaan yang sama gratis; percobaan berikutnya butuh iklan lagi; hasil percobaan lama tetap bisa dibuka. Halaman ringkasan paket dibuka agar aplikasi bisa menawarkan iklan.
- **Berita**: bila `contentGates.news` aktif, `GET /v1/news/{id}` butuh login (tamu mendapat 401 `login_required`) dan Lifetime atau unlock `news:<id>`. Daftar berita tetap publik.
- **Lowongan**: bila `contentGates.jobs` aktif, daftar dan detail publik tidak lagi berisi `applyUrl` maupun `sourceUrl` (diganti `applyLocked: true`). Tautan diambil lewat `POST /v1/jobs/{id}/apply-link` (butuh login), yang hanya memberikannya untuk Lifetime atau unlock `job:<id>`. Selama gerbang mati, endpoint itu memberikan tautan ke semua akun yang login.
- **Config baru**: `unlockKinds` (sakelar iklan per jenis: `lessons`, `anki`, `simulation`, `news`, `jobs`, default semua aktif) dan `contentGates` (`news`, `jobs`, default **mati**). Simpan dari dashboard lama tidak mereset keduanya.

Temuan yang mengubah cakupan:
- **Simulasi JLPT N3 paket 02–10 ternyata sudah gratis.** Yang terkunci hanya paket 01, jadi "simulasi N3 dengan iklan" berarti `jlpt-n3-simulation-1`. Paket 01 N2 dan N1, JFT, NAT, dan SSW tetap Lifetime. Menambahkan N2/N1 paket 01 cukup satu baris di `adUnlockableExams`.
- Deskripsi lowongan dari sumber bisa saja memuat tautan di dalam teksnya. Itu tidak disaring.

**Aman di-deploy sekarang.** Kedua gerbang konten default mati dan `NIHONGO_OPEN_ACCESS` belum diubah, jadi tidak ada perilaku pengguna yang berubah. Gerbang berita dan lowongan baru boleh dinyalakan setelah versi aplikasi yang bisa membuka keduanya beredar dan versi minimum dinaikkan. Klien lama yang membaca berita akan melihat error, dan tombol lamarnya akan nonaktif.

## Hasil fase 3 (admin)

Di dashboard admin, halaman **Iklan** (`/admin/ads`), kini ada:
- Placement **Rewarded buka konten** (aktif/nonaktif dan unit ID), dengan pengingat bahwa unit itu wajib memakai SSV ke `/v1/ads/ssv`.
- Kartu **Aturan buka konten**: batas harian, jatah cadangan, dan sakelar iklan per jenis.
- Kartu **Kunci konten gratis**: gerbang berita dan lowongan, dengan peringatan dan dialog konfirmasi setiap kali gerbang dinyalakan.
- Kartu **Statistik buka konten** (7, 14, atau 30 hari, per hari WIB): iklan yang diminta aplikasi, yang dikonfirmasi Google, item yang dibuka dengan iklan, dan yang dibuka dengan cadangan, per jenis dan per hari. Juga jumlah akun yang membuka sesuatu dan yang di antaranya kini Lifetime, serta jumlah callback Google (diterima, diabaikan, tanda tangan salah).

Endpoint: `GET /v1/admin/unlock-stats?days=14` (kunci admin iklan atau sesi admin).

Batasan statistik yang perlu diingat:
- Permintaan iklan (intent) dihapus setelah 7 hari, jadi kolom "Diminta" dan "Dikonfirmasi" kosong untuk hari yang lebih lama. Data unlock tersimpan permanen.
- Penghitung callback dihitung sejak server terakhir dinyalakan, dan reset setiap restart.
- "Kini Lifetime" hanya menunjukkan irisan akun, bukan bukti bahwa iklan mendorong pembelian. Mengukur konversi yang sebenarnya butuh pencatatan kapan pengguna menemui gerbang, yang belum dibuat.
- Simpan dari dashboard versi lama (yang belum mengenal kolom baru) tidak mereset pengaturan unlock.

Diverifikasi di server pratinjau lokal: tampilan semua kartu, dialog konfirmasi saat mengunci berita, nilai tersimpan di `/v1/ads-config`, validasi batas harian (nilai 500 ditolak), dan pergantian rentang hari.

## Hasil fase 4 (aplikasi Flutter)

Sudah ada:
- **Inti** di `lib/core/unlock/`:
  - `UnlockRepository`: intent, klaim, cadangan, dan status.
  - `UnlockFlow`: intent, lalu iklan dengan data SSV (`userId`, `customData`), lalu klaim dengan polling.
  - `showUnlockSheet` / `UnlockSheet`: "Tonton iklan (gratis)", "Beli Lifetime", "Batal", plus sisa jatah hari ini. Opsi "Buka tanpa iklan" muncul setelah dua iklan gagal dimuat.
  - `LockedContentException`, `LockedContentView`, dan `UnlockLockIcon`.
- **Config iklan** (`AdsConfig`): `unlockRewarded` dan `unlockKinds`. `RewardedAdShower.show` kini menerima data SSV.
- **Pelajaran** grammar, kanji, reading, kosakata: repositori melempar `LockedContentException` saat 402. Layar pelajaran menampilkan `LockedContentView` dan memuat ulang setelah terbuka. Daftar pelajaran menampilkan gembok pada N3–N1 hanya bila server melaporkan gerbang aktif (`gates.lessons`), item belum dibuka, dan akun bukan Lifetime.
- **Anki**: tombol mulai pada deck bawaan yang terkunci membuka sheet unlock. Deck impor tetap ke Lifetime.
- **Simulasi**: memulai paket yang terkunci menawarkan "Buka satu percobaan simulasi", lalu langsung memulai setelah terbuka.
- **Berita**: artikel terkunci menampilkan `LockedContentView`. Tamu melihat ajakan masuk, lalu artikel dimuat ulang.
- **Lowongan**: bila `applyLocked`, tombol lamar meminta tautan lewat `POST /v1/jobs/{id}/apply-link`, menawarkan sheet unlock saat 402, lalu membuka tautan. Tamu diarahkan login. Selama gerbang mati, perilaku lama (Lifetime saja) tetap berlaku.
- **Backend**: `GET /v1/unlocks` kini juga mengirim `gates` (lessons, news, jobs) dan `offered` per jenis, supaya aplikasi tidak menggambar gembok palsu selama konten masih terbuka.
- Halaman Lifetime kini bisa kembali ke `/news/...` dan `/anki/...` setelah pembelian.
- Terjemahan Inggris ada di `lib/core/localization/ui_english_unlock.dart`.

Perilaku penting:
- **Build debug** memakai unit uji Google, yang tidak pernah memanggil backend ini. Karena itu iklan uji yang ditonton dicatat lewat jalur cadangan, yang terbatas 2 per hari. Untuk pengujian lokal, naikkan jatah cadangan di admin. SSV yang sebenarnya hanya bisa diuji dengan build rilis dan unit asli (Play internal testing).
- **iOS** tidak menampilkan opsi iklan, hanya Lifetime.
- Cache offline: respons premium yang didapat akun gratis tersimpan dengan tanda premium dan hanya bisa dibaca dengan lease Lifetime, jadi pelajaran hasil iklan praktis hanya bisa dibuka online. Ini belum diuji di perangkat.

Tes: `test/core/unlock/unlock_test.dart` (11 tes: parsing 402, alur SSV, pending, iklan ditutup atau gagal, cadangan, sheet, sakelar per jenis, ikon gembok) dan `test/features/jobs/job_apply_unlock_test.dart` (3 tes). Seluruh suite Flutter menyisakan 20 kegagalan, dan **20 kegagalan yang sama terjadi di HEAD yang bersih** (login, SSW, exam hub, coverage string `OK`/`Premium`/`Lesson`, dan lainnya). Artinya tidak ada regresi dari fase ini, tetapi kegagalan lama itu perlu dibereskan terpisah.

Belum dicoba di perangkat atau emulator.

## Kondisi sekarang (hasil pembacaan kode)

| Temuan | Rujukan | Dampak |
|---|---|---|
| **Semua N3–N1 terbuka untuk semua akun yang login.** `openLearningAccess` aktif secara default; gerbang Lifetime baru berlaku bila `NIHONGO_OPEN_ACCESS=false`. | `backend/internal/store/billing.go` | Fitur ini mengaktifkan kembali pembatasan. Pengguna yang sekarang belajar gratis akan terkena. |
| Gerbang pelajaran ada di satu titik: `authorizePremium` (dipanggil dari `requireAuth`) memakai `PremiumRequest`. | `billing.go`, `server.go` | Titik tunggal untuk menambah "atau sudah dibuka lewat iklan". |
| Katalog level (`/course`) N3–N1 ikut digembok. | `PremiumRequest` | Pengguna tidak bisa melihat daftar yang mau dibuka. Perlu diubah. |
| **Gerbang apply lowongan hanya di klien.** `GET /v1/jobs/{id}` publik, di-cache publik 60 detik, dan mengirim `applyUrl` dan `sourceUrl` ke siapa pun. | `api/jobs.go`, `job_apply_button.dart` | Gerbang bisa dilewati dengan memanggil API. Harus dipindah ke server. |
| `sourceUrl` dan `applyUrl` sering sama (banyak sumber mengisi keduanya dengan URL yang sama). | `jobfetch/*.go` | Menyembunyikan `applyUrl` saja tidak cukup; `sourceUrl` pun harus ikut. |
| **Berita publik tanpa login.** `GET /v1/news` dan `/v1/news/{id}` tidak memakai autentikasi. | `api/news.go` | Gerbang per artikel butuh login. |
| Router Flutter tidak lagi mengalihkan ke `/premium` untuk materi belajar. | `app_router.dart` | Klien yang beredar tidak tahu cara menangani 402 pada pelajaran. |
| Rewarded Tanya AI mempercayai klaim klien, dibatasi kuota harian. | `docs/ai-chat-spec.md`, `aichat/quota.go` | Tidak cukup untuk konten pengganti produk berbayar. Perlu SSV. |
| Iklan hanya di Android (`adsSupportedPlatform`). | `ads_controller.dart` | iOS tidak punya jalur iklan. |
| Cache offline membaca konten premium hanya dengan lease premium yang valid. | `offline_access.dart` | Konten hasil iklan tidak boleh tersimpan offline. |

## Alur

1. Pengguna membuka item yang terkunci. App menampilkan sheet: **Tonton iklan (gratis)**, **Beli Lifetime**, **Batal**. Di iOS dan saat placement nonaktif, hanya Lifetime.
2. Pengguna memilih iklan. App memanggil `POST /v1/unlocks/intents {kind, itemId}`. Server memeriksa batas harian, membuat `nonce` sekali pakai (berlaku 10 menit) yang terikat ke akun dan item.
3. App memuat rewarded ad dengan `ServerSideVerificationOptions(userId: <ID akun buram>, customData: <nonce>)`.
4. Setelah iklan selesai, Google memanggil `GET /v1/ads/ssv?...` di backend. Server memverifikasi tanda tangan ECDSA (kunci publik Google, di-cache), menolak `transaction_id` yang sudah pernah dipakai, mencocokkan `custom_data` dengan nonce, lalu menandai nonce **terverifikasi**.
5. App memanggil `POST /v1/unlocks/claim {nonce}` dan polling singkat (maksimal sekitar 10 detik). Begitu nonce terverifikasi, server membuat catatan unlock, mengembalikan 200, dan app membuka item.
6. Selanjutnya server mengizinkan item itu karena ada catatan unlock atau akun Lifetime.

Jalur cadangan (iklan tidak tersedia setelah dua kali gagal memuat): `POST /v1/unlocks/fallback`. Server menghitung jatah cadangan harian dan membuat unlock tanpa SSV.

## Rencana teknis

### Backend (Go)

**Data**
- Tabel `content_unlocks`: `user_id`, `content_key`, `source` (`rewarded` | `fallback`), `created_at`; unik pada `(user_id, content_key)`. Contoh kunci: `grammar:<id>`, `kanji:<id>`, `reading:<id>`, `vocab:<id>`, `anki:<deckID>`, `sim:<attemptID>`, `news:<id>`, `job:<id>`.
- Tabel `unlock_intents`: `nonce`, `user_id`, `content_key`, `expires_at`, `verified_at`, `transaction_id` (indeks unik untuk mencegah pemakaian ulang).
- Job berkala menghapus intent kedaluwarsa.

**SSV dan endpoint**
- Verifier SSV: ambil kunci dari `https://www.gstatic.com/admob/reward/verifier-keys.json`, cache dengan masa berlaku, verifikasi tanda tangan atas query string sesuai spesifikasi Google.
- `POST /v1/unlocks/intents`, `POST /v1/unlocks/claim`, `POST /v1/unlocks/fallback`, `GET /v1/ads/ssv` (publik, dilindungi tanda tangan), `GET /v1/unlocks` (daftar yang sudah dibuka akun ini).

**Pelajaran, Anki, simulasi**
- Ubah `authorizePremium`: bila `PremiumRequest` mengatakan "butuh premium" dan akun bukan Lifetime, periksa `content_unlocks` untuk kunci item yang diminta. Kunci diturunkan dari ID pelajaran, atau dari ID attempt lewat pencarian attempt→pelajaran yang sudah ada.
- Anki: kunci dari `deckID` untuk rute `study` dan `review` (pengecekan level deck sudah ada di `authorizePremium`).
- Simulasi N3: setiap mulai simulasi butuh unlock `sim:<attemptID>`. Unlock dibuat sebelum attempt dimulai (intent memakai ID ujian; attempt yang dimulai mengonsumsi satu unlock yang belum terpakai). Perlu keputusan implementasi saat fase 1.
- Respons yang tidak diizinkan tetap 402 `premium_required`, kini dengan `unlockable: true` bila syarat iklan terpenuhi.
- Katalog (`/course`, daftar deck, daftar ujian) N3–N1 terbuka untuk akun login, dengan field `locked` per item. Pastikan respons katalog tidak membawa isi pelajaran.

**Berita**
- Daftar berita tetap publik (judul, ringkasan, gambar). `GET /v1/news/{id}` hanya mengirim isi artikel bila akun login dan punya unlock `news:<id>` atau Lifetime. Selain itu 402 `premium_required` dengan `unlockable: true`. Tamu mendapat 401 dan diarahkan login.
- Pertahankan `Cache-Control: no-store` (sudah ada) agar isi tidak masuk cache CDN.

**Lowongan**
- Hapus `applyUrl` dan `sourceUrl` dari respons publik daftar dan detail (`GET /v1/jobs`, `GET /v1/jobs/{id}`), dan sertakan field `applyLocked: true`. Tampilan lain (judul, deskripsi, lokasi, gaji) tetap publik.
- Endpoint baru `POST /v1/jobs/{id}/apply-link` (butuh login): mengembalikan `applyUrl` bila akun Lifetime atau ada unlock `job:<id>`. Respons ini `no-store`.
- Catatan: halaman "sumber lowongan" (`/v1/jobs/sources`) tetap publik dan tidak berisi tautan per lowongan.

**Config**
- Perluas `AdsConfig`: `UnlockRewarded` (aktif/nonaktif dan unit ID), `UnlockDailyCap` (default 20), `UnlockFallbackPerDay` (default 2), dan **sakelar per jenis** (`lessons`, `anki`, `simulation`, `news`, `jobs`) supaya tiap jenis bisa dinyalakan atau dimatikan terpisah dari admin tanpa rilis.
- Validasi sama seperti placement lain.

### Aplikasi (Flutter)

- **Sheet unlock** (satu komponen untuk semua jenis): dipanggil saat item terkunci dan pada respons 402 `unlockable`. Tombol iklan disembunyikan di iOS dan saat sakelar jenis itu nonaktif.
- **Gateway rewarded**: perluas `RewardedAds` (sudah ada) agar menerima `ServerSideVerificationOptions`. Tambah `UnlockRepository` (intent, claim, fallback, daftar unlock).
- **Pelajaran, Anki, simulasi**: ikon gembok di daftar, teks "Buka dengan iklan atau Lifetime".
- **Berita**: tap artikel membuka sheet unlock bila terkunci. Tamu diarahkan login lebih dulu.
- **Lowongan**: `JobApplyButton` memanggil `apply-link` setelah unlock, bukan membaca `applyUrl` dari model. Tampilkan "Buka tautan lamar" setelah terbuka.
- **Cache**: respons dengan header `X-Access: ad-unlock` tidak ditulis ke `LearningCache`. Unduhan offline dan PDF tetap meminta Lifetime.
- **Kegagalan**: iklan gagal dimuat → coba lagi → setelah dua kali tawarkan jalur cadangan. Klaim belum terverifikasi → "Sedang memverifikasi" dengan polling terbatas.
- **String**: tambahkan di `app_localizations` (ID dan EN).

### Admin

- Placement **Rewarded buka konten** di `/admin/ads`: aktif/nonaktif, unit ID, batas harian, jatah cadangan, dan sakelar per jenis.
- Statistik sederhana: jumlah unlock per hari per jenis, rasio iklan→unlock, klaim ditolak, dan **konversi ke Lifetime setelah menemui gerbang**.

### AdMob

- Buat unit **`nm_rewarded_unlock`** (Rewarded). Di **Advanced settings → Server-side verification**, isi callback `https://nihongo.codeverta.com/v1/ads/ssv`. Pengaturan ini tidak bisa dipakai pada unit dengan partner bidding.
- Jangan memakai ulang `nm_rewarded_ai`: Tanya AI memakai alur klaim klien dan statistiknya harus terpisah.
- Pastikan Cloudflare tidak memblokir request Google ke `/v1/ads/ssv` (jangan ada tantangan bot pada path itu).

## Keamanan dan penyalahgunaan

- **SSV wajib**, nonce sekali pakai, `transaction_id` unik, nonce terikat ke akun dan item.
- `user_id` yang dikirim ke Google adalah ID akun buram, **bukan email**.
- Rate limit pada `intents`, `claim`, `fallback` per akun dan per IP.
- Batas harian (20) berlaku **bersama** untuk semua jenis. Berita dan pelajaran akan berbagi batas yang sama. Bila terasa terlalu ketat, pecah per jenis lewat config.
- Jalur cadangan adalah satu-satunya jalur tanpa SSV, dan dibatasi 2 per hari. Server tidak bisa membedakan platform dengan pasti, jadi klien iOS yang dimodifikasi bisa memakainya, tetapi hanya sampai jatah harian.
- Akun Lifetime yang di-refund kehilangan akses seperti biasa. Unlock iklan tidak terpengaruh.
- Gerbang lowongan dipindah ke server (lihat di atas); tanpa itu, klien mana pun masih bisa mengambil tautan lamar dari API publik.

## Risiko

**Teknis dan rilis**
1. **Klien lama akan rusak.** Versi yang beredar tidak tahu cara menangani 402 pada pelajaran dan berita, dan menampilkan layar error. Pada lowongan, tombol lamar akan nonaktif karena `applyUrl` tidak lagi dikirim. **Urutan wajib**: rilis app dengan sheet unlock → tunggu adopsi → naikkan versi minimum lewat `app_update_gate` (lihat `docs/app-updates.md`) → baru aktifkan gerbang di server. "Langsung berlaku" berarti tanpa masa transisi konten setelah langkah ini, bukan melewatinya.
2. **SSV hanya bisa diuji penuh dengan unit asli.** Build debug memakai unit uji Google. Rencanakan uji internal di Play dengan akun tester.
3. **Perubahan payload lowongan** memengaruhi klien lama (lihat di atas) dan semua tempat yang membaca `applyUrl` atau `sourceUrl`. Cari semua pemakai sebelum menghapus field.

**Produk (monetisasi agresif)**
4. **Retensi.** Berita dan lowongan adalah pintu masuk harian yang ringan; menggembok setiap artikel dan membuat tamu login dulu hampir pasti menurunkan pembaca harian dan jumlah pengguna baru. Jangan menyalakan semuanya sekaligus. Gunakan sakelar per jenis di admin dan nyalakan bertahap sambil memantau retensi hari ke-1 dan hari ke-7.
5. **iOS terkunci total untuk berita.** Dengan keputusan "iOS hanya Lifetime", pengguna iOS tidak bisa membaca berita atau melamar tanpa membeli. Itu konsisten dengan keputusan, tetapi paling mungkin menurunkan pengguna iOS. Ini layak ditinjau ulang.
6. **Pendapatan iklan kecil per tayangan di Indonesia.** Nilai utamanya adalah mendorong Lifetime. Pantau konversi pembelian, bukan hanya pendapatan iklan.
7. **Kebijakan Play.** Rewarded harus ditawarkan, bukan dipaksakan; selalu tampilkan Batal dan Lifetime, jangan menyembunyikan tombol keluar. Perbarui kebijakan privasi (SSV mengirim ID akun buram ke Google) dan tinjau Data safety.
8. **Ulasan negatif.** Mengubah konten gratis menjadi berbayar secara mendadak kerap memicu ulasan bintang satu. Siapkan catatan rilis yang jelas dan jalur dukungan.

## Urutan kerja

| Fase | Isi | Perkiraan |
|---|---|---|
| 0 | Buat unit AdMob `nm_rewarded_unlock` dengan callback SSV | 0,5 hari |
| 1 | Backend inti: tabel, verifier SSV, endpoint unlock, `authorizePremium`, katalog `locked`, tes (tanda tangan, nonce, replay, batas harian) | 4–5 hari |
| 2 | Backend per jenis: gerbang berita, pemindahan `applyUrl` ke `apply-link`, Anki, simulasi, tes | 3–4 hari |
| 3 | Admin: placement, sakelar per jenis, statistik | 1–2 hari |
| 4 | Aplikasi: sheet unlock, alur rewarded + SSV, gembok di daftar, berita, lowongan, kebijakan cache, string ID/EN, tes widget | 5–6 hari |
| 5 | Uji internal Play dengan unit asli, verifikasi callback SSV nyata | 1–2 hari |
| 6 | Rilis app → versi minimum → aktifkan `NIHONGO_OPEN_ACCESS=false` → nyalakan jenis demi jenis di admin, pantau retensi dan konversi | bertahap |

Total perkiraan: sekitar 15–20 hari kerja sebelum rilis bertahap.

## Tes

- **Backend**: verifier SSV dengan kunci uji, tanda tangan salah, `transaction_id` ganda, nonce kedaluwarsa, nonce milik akun lain, batas harian bersama, jatah cadangan, `authorizePremium` untuk GET pelajaran, attempt, submit, Anki, simulasi, katalog `locked` tanpa kebocoran isi, berita (tamu, login tanpa unlock, dengan unlock, Lifetime), lowongan (payload publik tanpa `applyUrl` dan `sourceUrl`, `apply-link` dengan dan tanpa unlock).
- **Aplikasi**: sheet unlock (Android vs iOS, sakelar jenis nonaktif), alur sukses, iklan gagal, klaim belum terverifikasi, 402 pada tiap jenis, tidak ada cache untuk `X-Access: ad-unlock`, akun Lifetime tidak melihat sheet, tamu diarahkan login untuk berita.
- **Manual**: satu putaran penuh di Play internal testing dengan unit asli, termasuk memastikan callback SSV benar-benar sampai ke backend produksi.
