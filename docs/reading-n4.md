# Reading N4: cerita bergambar

Buka **Belajar → N4 → Reading** setelah login. Modul berisi **12 cerita,
36 adegan bergambar, 72 entri kosakata bantu, dan 60 soal pemahaman**.
Setiap cerita mempunyai tiga halaman dengan paragraf empat kalimat dan
lima soal. Entri kosakata dapat berulang; angka 72 bukan jumlah kata unik.

Format sama dengan [Reading N5](reading-n5.md): ilustrasi per halaman,
teks Jepang, bantuan kana yang dapat dibuka, terjemahan Indonesia, kosakata,
navigasi halaman, kuis dengan penjelasan, skor terbaik, dan lanjut belajar.
Bantuan kana berupa paragraf terpisah, bukan furigana per kanji.
Pencarian dan filter tema tersedia di katalog.

## Cerita dan fokus bahasa

| No. | Cerita | Fokus latihan |
|---|---|---|
| 1 | Rumah baru, tetangga baru | ～たばかり, ～てくれる, ～ようになる |
| 2 | Kerja pertama di kafe | ～なければならない, potensial, ～ておく |
| 3 | Kereta terlewat, rencana berubah | ～てしまう, ～たら, ～ことにする |
| 4 | Malam festival musim panas | ～てみる, ～そう, ～前に |
| 5 | Hari hujan di museum | ～ので, ～ことがある, ～かもしれない |
| 6 | Memasak kari bersama | ～ておく, ～すぎる, ～ながら |
| 7 | Berani mencoba badminton | potensial, ～ても, ～ようになる |
| 8 | Hujan di jalur pendakian | ～かもしれない, ～ておく, ～まで |
| 9 | Membantu membersihkan sungai | ～ために, ～ずつ, ～ようになる |
| 10 | Dompet yang tertinggal | ～かどうか, ～てあげる, ～つもり |
| 11 | Menginap di ryokan | ～てある, ～前に, ～たら |
| 12 | Kafe di festival sekolah | ～ておく, ～てしまう, alasan dan kerja sama |

Teks, terjemahan, dan soal ditulis khusus untuk aplikasi. Ini adalah
latihan membaca dengan pola N4 dan bantuan untuk kosakata tambahan,
bukan salinan soal atau daftar resmi seluruh materi JLPT.
Jawaban kuis mengikuti teks; gambar membantu konteks cerita.

## Konten dan ilustrasi

- Konten: `backend/data/reading/n4.json`, di-embed ketika backend dibuild.
- Gambar baru: `backend/assets/images/reading/story-n4-01.png` sampai
  `story-n4-12.png`, dibuat dengan **imagegen bawaan**, tanpa CLI.
- Setiap gambar terdiri dari tiga panel persegi berdampingan. Widget yang
  sama dengan N5 menampilkan panel yang sesuai pada tiap halaman.
- Prompt dan pemetaan gambar: [reading-n4-art.json](reading-n4-art.json).
- Gambar dibundel dalam APK. `imageUrl` dari backend dapat memakai COS;
  jika gambar remote gagal, aplikasi memakai asset lokal. Tidak ada unggahan
  otomatis ke COS dalam perubahan ini.

## API dan progres

`GET /v1/reading/course?level=N4` mengambil katalog N4. Tanpa parameter,
endpoint tetap mengambil N5. Nilai selain N5/N4 ditolak dengan HTTP 400.
ID lesson N4 adalah `reading-n4-01` sampai `reading-n4-12`. Endpoint detail,
attempt, jawaban, dan penyelesaian sama seperti N5 dan memerlukan login.
Lihat [kontrak API](../backend/openapi.yaml).

Progres dihitung terpisah per level. Ambang lulus 80% (4 dari 5 soal),
50 XP hanya untuk kelulusan pertama. Skor terbaik dan status selesai
dipertahankan ketika mengulang. Progres masih disimpan di memori dan
reset ketika backend direstart.

Restart backend setelah mengubah JSON; build ulang Flutter setelah
menambah gambar agar asset masuk APK. [Panduan menjalankan project](running-project.md).
