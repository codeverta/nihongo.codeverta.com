# Modul Kanji N5

Buka **Belajar → N5 → Kanji** setelah login. Listening dan Speaking AI
disembunyikan dari menu Belajar, Beranda, dan kategori API untuk sementara.

## Isi materi

Kurikulum persiapan N5 aplikasi terdiri dari **20 lesson, 100 kanji unik,
200 contoh kata/frasa, dan 200 pertanyaan**. Setiap lesson memiliki lima
kanji dan sepuluh soal: lima soal arti, lima soal cara membaca kata.

Setiap kartu menampilkan karakter, arti bahasa Indonesia, bacaan on’yomi
dan kun’yomi yang umum, jumlah goresan, petunjuk mengingat, dan dua contoh
dengan cara baca serta terjemahan. Petunjuk mengingat adalah bantuan belajar,
bukan klaim asal-usul historis karakter.

On’yomi biasanya digunakan dalam gabungan kanji, sedangkan kun’yomi sering
digunakan pada kata asli Jepang. Ini bukan aturan mutlak: pelajari bacaan
melalui contoh kata, termasuk perubahan bunyi dan bacaan khusus.
Titik pada kun’yomi memisahkan bagian kanji dari okurigana; tanda `—`
menunjukkan tidak ada bacaan umum yang diajarkan pada kartu tersebut.
Modul menampilkan pilihan bacaan umum, bukan seluruh isi kamus kanji.

Area **Coba tulis ulang kanji** menyediakan karakter samar dan garis bantu.
Pengguna dapat menyalin dengan jari dan menghapus tulisan. Area ini tidak
menilai bentuk atau urutan goresan; jumlah goresan pada kartu bukan animasi
urutan penulisan. Tulisan latihan bersifat sementara.

## Kurikulum

| Lesson | Kanji |
|---|---|
| 1 | 一 二 三 四 五 |
| 2 | 六 七 八 九 十 |
| 3 | 日 月 火 水 木 |
| 4 | 金 土 曜 年 今 |
| 5 | 時 分 半 午 前 |
| 6 | 後 毎 何 人 名 |
| 7 | 父 母 子 男 女 |
| 8 | 学 校 先 生 友 |
| 9 | 口 目 耳 手 足 |
| 10 | 上 下 中 外 右 |
| 11 | 左 東 西 南 北 |
| 12 | 大 小 多 少 長 |
| 13 | 山 川 田 天 雨 |
| 14 | 店 社 魚 犬 気 |
| 15 | 行 来 帰 入 出 |
| 16 | 見 聞 読 書 話 |
| 17 | 食 飲 買 休 会 |
| 18 | 国 車 電 駅 道 |
| 19 | 本 語 字 百 千 |
| 20 | 万 円 白 赤 空 |

JLPT tidak menerbitkan daftar tertutup kanji untuk ujian setelah revisi
2010. Daftar ini adalah kurikulum dasar aplikasi untuk persiapan N5,
bukan jaminan seluruh kanji yang dapat muncul pada ujian.
[FAQ resmi JLPT](https://www.jlpt.jp/e/faq/).

## Kuis, progres, dan XP

- Nilai minimal 80% (8 dari 10 soal) menyelesaikan lesson.
- Lulus pertama kali memberi 50 XP; mengulang tidak menggandakan XP.
- Skor terbaik dan status selesai tetap tersimpan ketika percobaan berikutnya gagal.
- Kunci jawaban tidak dikirim dalam respons awal kuis.
- Jawaban pertama untuk suatu pertanyaan tetap, sehingga retry jaringan aman.
- Backend menghitung skor; endpoint progres umum tidak dapat menyelesaikan lesson Kanji.
- Kegagalan jaringan menampilkan pesan dan memungkinkan tindakan diulang.
- Progres dan akun masih memakai store memori; restart backend mereset progres.

## Backend

Materi berada di `backend/data/kanji/n5.json`, di-embed oleh Go saat build.
Flutter mengambil materi melalui API; kunci jawaban tetap di backend.
Perubahan JSON memerlukan restart backend. Struktur kurikulum divalidasi
saat startup agar kanji duplikat atau soal tidak valid tidak terpublikasi.

Semua endpoint berikut membutuhkan bearer token:

| Method | Endpoint |
|---|---|
| GET | `/v1/kanji/course` |
| GET | `/v1/kanji/lessons/{lessonID}` |
| POST | `/v1/kanji/lessons/{lessonID}/attempts` |
| POST | `/v1/kanji/attempts/{attemptID}/answers` |
| POST | `/v1/kanji/attempts/{attemptID}/complete` |

Endpoint jawaban menerima `questionId` dan `optionId`. Kontrak lengkap
ada di `backend/openapi.yaml`. Panduan menjalankan backend dan emulator:
[Running project](running-project.md).
