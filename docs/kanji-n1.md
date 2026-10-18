# Kanji JLPT N1

Kurikulum N1 berisi **1.232 kanji dalam 247 lesson**. Setiap kartu mencantumkan
gloss kanji, jumlah goresan, pilihan bacaan on/kun, serta dua contoh kata dengan
bacaannya. Kuis berisi **2.464 soal**: satu soal gloss dan satu soal bacaan kata
untuk setiap kanji. Contohnya berupa kosakata yang dapat ditelusuri ke kamus;
materi ini tidak membuat kalimat contoh sintetis.

## Catatan tentang daftar N1

JLPT tidak menerbitkan daftar resmi kanji, kosakata, atau tata bahasa per level
sejak revisi 2010. Karena itu, N1 di sini adalah **pemetaan untuk belajar**,
bukan klaim daftar resmi dan lengkap dari penyelenggara JLPT. Karakter dipilih
dari entri dengan `jlpt_new == 1` di dataset [kanji-data oleh David Gouveia](https://github.com/davidluzgouveia/kanji-data).
README dataset tersebut menyebut KANJIDIC sebagai sumber data kanji dan
[sumber JLPT Jonathan Waller](https://www.tanos.co.uk/jlpt/) untuk level.
Rujukan resmi menjelaskan alasan spesifikasi berisi daftar kanji, kosakata, dan
grammar tidak lagi dipublikasikan di [FAQ JLPT](https://www.jlpt.jp/e/faq/).

## Cara menggunakan materi

- Buka **Belajar → N1 → Kanji**.
- Setiap lesson memuat empat atau lima kanji, gloss Inggris dari KANJIDIC, dan
  dua contoh kata yang dipilih dari JMdict.
- Bacaan on ditampilkan dalam katakana; kun memakai ejaan dictionary termasuk
  titik pemisah okurigana. Daftar pada kartu adalah pilihan bacaan sumber, bukan
  seluruh bacaan yang mungkin.
- Gloss JMdict pada contoh kata juga dipertahankan dalam bahasa Inggris karena
  sumber kamus ini menyediakan komponen Jepang–Inggris. Arti tersebut tidak
  diterjemahkan otomatis ke bahasa Indonesia.
- Dua belas karakter tidak memiliki headword JMdict yang cocok di snapshot yang
  dipakai. Untuk karakter tersebut contoh satu-kanji bersumber dari entri
  KANJIDIC dan ditandai dalam `docs/kanji-n1-reference.json`:
  栞・瑛・舜・朕・彪・麿・侑・崚・晟・滉・燿・赳.

Kuis menguji gloss pada kartu dan bacaan contoh pertama. Lesson menggunakan
format empat atau lima kanji agar jumlah pilihan kuis konsisten. Pengelompokan
adalah urutan belajar aplikasi, bukan kategori resmi JLPT.

## Sumber, tanggal, dan lisensi

- **Kanji, bacaan, gloss, dan goresan:** KANJIDIC/KANJIDIC2 oleh James William
  Breen dan Electronic Dictionary Research and Development Group (EDRDG),
  melalui [kanji-data](https://github.com/davidluzgouveia/kanji-data). Level
  `jlpt_new` mengikuti rujukan Jonathan Waller yang dicantumkan oleh dataset.
- **Kosakata contoh, bacaan, dan gloss:** JMdict oleh James William Breen dan
  EDRDG, [berkas unduhan resmi](https://www.edrdg.org/pub/Nihongo/JMdict.gz).
- EDRDG menyediakan KANJIDIC dan JMdict di bawah
  [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/), dengan
  ketentuan pada [pernyataan lisensi EDRDG](https://www.edrdg.org/edrdg/licence.html).
  Konversi `kanji-data` memiliki [lisensi MIT](licenses/kanji-data-MIT.txt).
- Snapshot sumber diambil 18 Oktober 2026. URL dan SHA-256 untuk tiap sumber,
  serta metadata dan contoh yang dipakai, tersimpan di
  [kanji-n1-reference.json](kanji-n1-reference.json). Data adaptasi kurikulum
  dirilis di bawah CC BY-SA 4.0.

## Pembuatan ulang dan integrasi

- Snapshot: `docs/kanji-n1-reference.json`.
- Generator offline: `scripts/build_kanji_n1.py`.
- Data embed backend: `backend/data/kanji/n1.json` melalui `N1` di
  `backend/data/kanji/n1_content.go`.
- Seed helper: `backend/internal/store/kanji_n1.go` menyediakan
  `seedKanjiN1()` (247 lesson, 1.232 kanji). Pemanggilan bersama level lain
  berada pada registrasi `seedKanji()`.
- Daftar lesson: [kanji-n1-curriculum.md](kanji-n1-curriculum.md).

Jalankan `python scripts/build_kanji_n1.py` untuk membuat ulang JSON kurikulum
dan daftar lesson tanpa akses jaringan.
