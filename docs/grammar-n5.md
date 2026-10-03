# Modul Grammar JLPT N5

Modul tersedia melalui **Belajar → N5 → Grammar** setelah login.
Kurikulum aplikasi berisi **40 lesson, 80 unit pola, 160 contoh kalimat,
dan 200 soal latihan**. Setiap lesson memiliki dua pola dan lima soal.
Semua penjelasan dan soal ditulis untuk aplikasi ini dalam bahasa Indonesia;
contoh dilengkapi kalimat Jepang, cara baca kana, dan terjemahan.

## Cakupan kurikulum

| Lesson | Materi |
|---|---|
| 1–4 | Identitas, negatif dan lampau kata benda, は, か, の, も |
| 5–7 | これ・それ・あれ・どれ, この・その・あの・どの, tempat dan arah sopan |
| 8–9 | ある・いる, lokasi keberadaan, posisi relatif |
| 10–15 | を, で, へ, に, waktu dan durasi, から・まで, と・や, sarana, が |
| 16–19 | ます・ません・ました・ませんでした, kelompok kata kerja, bentuk kamus, ない, た |
| 20–23 | Kata sifat い dan な: positif, negatif, lampau, dan menerangkan benda |
| 24–27 | Bentuk て, urutan tindakan, permintaan, larangan, izin, ている, てから |
| 28–30 | Keinginan たい・ほしい, ajakan, tujuan perjalanan, 前に・後で |
| 31–33 | とき, alasan から, 好き・嫌い・上手・下手, kewajiban dan tidak perlu |
| 34–37 | Perbandingan, いちばん, あまり・ぜんぜん, もう・まだ, penghitung, frekuensi, だけ・くらい |
| 38–40 | Kontras, gabungan sifat, frasa penerang benda, nominalisasi の, ね・よ, でしょう |

Unit pola mengelompokkan beberapa variasi terkait; jumlah unit tidak sama
dengan jumlah setiap bentuk tata bahasa yang mungkin digunakan.

JLPT tidak lagi menerbitkan daftar tertutup kosakata/kanji/grammar setelah
revisi 2010. Karena itu, cakupan di atas adalah kurikulum persiapan N5 milik
aplikasi, bukan checklist resmi seluruh soal yang mungkin muncul di ujian.
Penjelasan bahasa disusun dengan mengacu pada tata bahasa dasar; contoh dan
soal latihan bukan salinan soal JLPT atau teks buku.

Referensi pengajar:

- [JLPT — FAQ tentang Test Content Specifications](https://www.jlpt.jp/e/faq/).
- [JLPT — ringkasan kemampuan N1–N5](https://www.jlpt.jp/e/about/levelsummary.html).
- [Japan Foundation — Irodori Grammar Notes](https://www.irodori.jpf.go.jp/assets/data/Grammar_all.pdf).

## Alur pengguna dan penilaian

1. Cari judul lesson atau pilih bab. Tombol lanjut memilih lesson pertama yang
   belum selesai.
2. Baca dua pola, rumus, penjelasan, contoh, dan kesalahan umum.
3. Jawab lima soal. Tiga soal menguji bentuk/pemakaian, dua soal menguji makna
   kalimat dari contoh. Setiap soal memiliki empat pilihan.
4. Backend memberi jawaban benar dan penjelasan setelah jawaban dikirim.
5. Lulus dengan minimal empat dari lima jawaban benar (80%). Nilai terbaik dan
   status selesai disimpan per pengguna. Nilai yang lebih rendah pada retry
   tidak menurunkan nilai terbaik atau membatalkan lesson yang sudah selesai.
6. Kelulusan pertama memberi 50 XP; mengulang kuis tidak memberi XP tambahan.

Progress kursus adalah persentase lesson yang sudah lulus, bukan rata-rata
nilai kuis. Jika koneksi gagal saat mengirim jawaban atau menyimpan hasil,
screen tetap menyimpan posisi dan pilihan sehingga pengguna dapat mencoba
lagi. Backend menolak perubahan jawaban setelah feedback dan menghitung ulang
hasil yang sama secara idempotent jika request penyelesaian dikirim ulang.

Backend saat ini masih **in-memory**, mengikuti arsitektur project: sesi,
nilai, progress, dan XP hilang ketika proses backend dimulai ulang.

## Menjalankan perubahan

Jalankan backend yang sudah diperbarui dari folder `backend`:

```powershell
go run ./cmd/api
```

Jika server lama masih berjalan, hentikan dengan Ctrl+C terlebih dahulu.
Lalu restart Flutter dan login kembali jika sesi backend sudah berakhir:

```powershell
flutter run -d emulator-5554
```

Lihat juga [panduan menjalankan project](running-project.md).

## Implementasi dan API

- Materi: `backend/internal/store/grammar_content.go`.
- Penilaian/progress: `backend/internal/store/grammar.go`.
- Model API: `backend/internal/domain/grammar.go`.
- Flutter: `lib/features/learn/grammar/`.

Semua endpoint membutuhkan bearer token:

| Method | Endpoint | Hasil |
|---|---|---|
| GET | `/v1/grammar/course` | Kurikulum, ringkasan lesson, dan progress pengguna |
| GET | `/v1/grammar/lessons/{lessonId}` | Dua pola beserta contoh; tanpa kunci kuis |
| POST | `/v1/grammar/lessons/{lessonId}/attempts` | Attempt baru dengan lima soal |
| POST | `/v1/grammar/attempts/{attemptId}/answers` | Kirim `questionId`, `optionId`; terima feedback |
| POST | `/v1/grammar/attempts/{attemptId}/complete` | Hitung nilai, simpan terbaik, progress dan XP |

Progress Grammar tidak dapat ditulis lewat endpoint progress umum; nilai
hanya disimpan oleh hasil kuis server. Kontrak lengkap ada di
`backend/openapi.yaml`.

## Pemeriksaan pengembangan

```powershell
flutter test test/features/learn/grammar
flutter analyze lib/features/learn/grammar test/features/learn/grammar
cd backend
go test ./...
```

Tes backend memeriksa integritas seluruh materi, seluruh course dapat
diselesaikan, pemisahan data pengguna, retry, nilai terbaik, aturan 80%,
idempotensi hasil, serta kontrak autentikasi dan JSON. Tes Flutter memeriksa
pencarian, layout ponsel kecil, alur membaca dan kuis, retry jaringan, serta
respons asinkron setelah screen ditutup.
