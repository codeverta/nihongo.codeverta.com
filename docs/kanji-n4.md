# Kanji N4

Buka **Belajar → N4 → Kanji** setelah login. Kurikulum berisi **34 lesson,
166 kanji unik, 332 contoh kata/frasa, dan 332 soal**. Sebanyak 30 lesson
berisi lima kanji dan sepuluh soal; empat lesson berisi empat kanji dan
delapan soal. Ada satu soal arti dan satu soal bacaan kata untuk setiap kanji.

Setiap kartu menyediakan arti Indonesia, pilihan on’yomi dan kun’yomi umum,
jumlah goresan, petunjuk mengingat, catatan penggunaan, serta dua contoh dengan
kana dan arti. Petunjuk mengingat dibuat untuk latihan, bukan penjelasan
asal-usul historis kanji. Titik pada kun’yomi menunjukkan batas okurigana;
`—` berarti tidak ada bacaan umum yang diajarkan pada kartu tersebut.

Materi mencakup semua 166 karakter N4 pada rujukan `kanji-data` yang digunakan.
Sebanyak **21 karakter juga ada di kurikulum N5 aplikasi** dan diulang dalam
contoh N4. N5 dan N4 bersama-sama mencakup 245 karakter unik. Pelajari N5
sebelum atau bersamaan dengan N4. Daftar ini adalah kurikulum persiapan,
bukan daftar resmi tertutup: [FAQ JLPT](https://www.jlpt.jp/e/faq/)
menjelaskan bahwa spesifikasi kosakata, kanji, dan grammar tidak dipublikasikan
untuk ujian yang direvisi sejak 2010.

## Kurikulum

| Lesson | Tema | Kanji |
|---|---|---|
| 01 | Orang tua dan saudara | 親 族 兄 弟 姉 |
| 02 | Diri dan peran | 妹 私 自 主 代 |
| 03 | Tubuh dan indra | 体 心 手 口 目 |
| 04 | Berdiri dan berjalan | 立 足 歩 走 止 |
| 05 | Bergerak dan tiba | 動 起 発 着 帰 |
| 06 | Perjalanan dan kendaraan | 転 運 通 道 駅 |
| 07 | Rumah dan bangunan | 家 屋 室 堂 館 |
| 08 | Kota dan lingkungan | 町 地 場 建 京 |
| 09 | Dunia dan organisasi | 世 界 公 社 員 |
| 10 | Pekerjaan dan produksi | 仕 事 業 工 作 |
| 11 | Belajar dan penelitian | 教 習 勉 研 究 |
| 12 | Ujian dan pertanyaan | 試 験 答 問 質 |
| 13 | Huruf dan bahasa | 漢 字 文 英 言 |
| 14 | Musik dan gambar | 音 歌 楽 画 映 |
| 15 | Makanan dan minuman | 茶 飯 肉 牛 飲 |
| 16 | Hewan dan tanaman | 鳥 魚 犬 花 野 |
| 17 | Laut dan langit | 海 風 空 夕 夜 |
| 18 | Musim dan pagi | 春 夏 秋 冬 朝 |
| 19 | Hari dan waktu | 昼 週 曜 早 急 |
| 20 | Berpikir dan memahami | 考 思 知 理 計 |
| 21 | Makna dan keadaan | 意 味 有 不 悪 |
| 22 | Ketepatan dan perbedaan | 正 真 同 別 特 |
| 23 | Jumlah dan ukuran | 少 多 度 重 力 |
| 24 | Warna dan terang | 色 赤 青 黒 明 |
| 25 | Kesehatan dan keselamatan | 病 医 院 死 安 |
| 26 | Kekuatan dan cara | 強 元 以 方 用 |
| 27 | Jual beli dan biaya | 買 売 店 品 料 |
| 28 | Memegang dan meminjam | 持 待 借 貸 使 |
| 29 | Kertas, pakaian, dan barang | 切 写 紙 服 物 |
| 30 | Tempat tinggal dan jarak | 住 広 近 田 洋 |
| 31 | Lama, baru, mulai, dan selesai | 古 新 始 終 |
| 32 | Pertemuan dan perjalanan | 会 集 旅 送 |
| 33 | Gambar dan pembahasan | 台 図 注 題 |
| 34 | Pergi, membuka, uang, dan orang | 去 開 銀 者 |

## Latihan dan progres

Katalog mendukung pencarian judul dan filter bab, lanjut belajar, dan skor
terbaik. Area menulis ulang yang sama dengan N5 tersedia pada tiap kartu;
ini latihan menyalin dengan jari, tanpa penilaian bentuk atau urutan goresan.

Ambang lulus 80%: **8/10** pada lesson lima kanji, atau **7/8** pada lesson
empat kanji. Kelulusan pertama memberi 50 XP. Retry tidak menggandakan XP,
menurunkan skor terbaik, atau membatalkan status selesai. Backend menghitung
nilai dan menjaga kunci jawaban. Progres N4 terpisah dari N5 dan pengguna lain.
Progres masih memakai store memori dan direset ketika backend direstart.

## Sumber dan lisensi data

Jumlah goresan dan data bacaan menggunakan **KANJIDIC**, karya James William
Breen dan Electronic Dictionary Research and Development Group (EDRDG),
melalui [kanji-data oleh David Gouveia](https://github.com/davidluzgouveia/kanji-data).
Penetapan level N4 mengikuti `jlpt_new`, yang dalam dataset tersebut berasal
dari [Jonathan Waller's JLPT Resources](https://www.tanos.co.uk/jlpt/).
Data WaniKani dan arti bahasa Inggris dalam dataset tidak digunakan.

Data telah diadaptasi: bacaan dipilih untuk pembelajar, on’yomi ditulis dengan
katakana, dan arti, contoh, mnemonic, serta kuis Indonesia ditambahkan.
`backend/data/kanji/n4.json` dan snapshot rujukannya merupakan materi data
adaptasi di bawah [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/).
Lisensi ini berlaku untuk data tersebut dan tidak mengubah lisensi kode aplikasi.

- [Pernyataan lisensi kamus EDRDG](https://www.edrdg.org/edrdg/licence.html).
- Salinan lokal: [EDRDG](licenses/kanjidic-edrdg-license.html) dan
  [lisensi MIT konversi kanji-data](licenses/kanji-data-MIT.txt).
- Pengakuan sumber juga tersedia pada **Tentang materi dan sumber** di katalog N4.

## Integrasi dan pembaruan

- Konten backend: `backend/data/kanji/n4.json`, di-embed saat build Go.
- Snapshot bacaan/goresan: `docs/kanji-n4-reference.json`, diambil 3 Oktober 2026.
- Sumber konten yang dapat diedit: `scripts/build_kanji_n4.py`.
- Jalankan `python scripts/build_kanji_n4.py` untuk membuat ulang JSON secara
  deterministik. Tidak diperlukan paket Python tambahan.
- Untuk memperbarui rujukan, ambil ulang dataset sumber dan pilih semua entri
  `jlpt_new == 4`, simpan `strokes`, `readings_on`, dan `readings_kun`, serta
  perbarui metadata sumber. Tinjau perbedaan dan konten sebelum regenerasi;
  generator menolak karakter yang hilang atau tidak tercakup. Periksa sumber
  saat setiap pembaruan kurikulum sebelum rilis.
- `GET /v1/kanji/course?level=N4` mengambil katalog N4; tanpa parameter tetap N5.
  Level selain N5/N4 ditolak dengan HTTP 400. Seluruh endpoint kanji butuh login.
- ID N4 adalah `kanji-n4-01` sampai `kanji-n4-34`. Endpoint detail, attempt,
  answer, dan complete sama seperti N5. Lihat `backend/openapi.yaml`.
- Restart backend dan reload Flutter setelah perubahan agar katalog baru tampil.

Pemeriksaan:

```powershell
cd backend
go test ./internal/store ./internal/api -run Kanji -count=1
```

```powershell
flutter test test/features/learn/kanji/kanji_n4_test.dart
```
