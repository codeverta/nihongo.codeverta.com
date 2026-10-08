# Kanji N3

Buka **Belajar → N3 → Kanji** setelah login. Kurikulum berisi **77 lesson,
367 kanji unik, 367 contoh kosakata, 367 contoh kalimat, dan 734 soal**.
Sebanyak 59 lesson memuat lima kanji dan sepuluh soal, sedangkan 18 lesson
memuat empat kanji dan delapan soal. Setiap kanji mempunyai satu soal arti
dan satu soal bacaan kosakata.

## Materi pada kartu

- Arti Indonesia dan pilihan on’yomi/kun’yomi dari kamus.
- Jumlah goresan, pengingat melalui adegan, dan catatan penggunaan.
- Satu contoh kosakata serta satu kalimat Jepang dengan bacaan dan terjemahan.
- Catatan khusus seperti 未／末, 寒い／冷たい, 二十歳, sapaan keluarga,
  bahasa merendah, dan perubahan bunyi dalam kata gabungan.
- Area latihan menyalin dengan jari yang sama dengan N5/N4.

Pengingat merupakan bantuan belajar, bukan penjelasan etimologi historis.
Jumlah goresan tidak menggantikan diagram urutan goresan. Area menulis tidak
menilai bentuk atau urutan goresan. Bacaan kana kalimat adalah baris terpisah,
bukan furigana per kanji. Titik pada kun’yomi memisahkan okurigana; `—` berarti
bacaan itu tidak diajarkan pada kartu tersebut, bukan klaim kanji tidak mempunyai
bacaan jenis tersebut dalam seluruh kamus.

Materi mencakup semua 367 karakter dengan `jlpt_new == 3` pada snapshot rujukan
yang digunakan. Ada satu pengulangan dari N5 aplikasi, yaitu 耳. Materi ini
merupakan daftar persiapan milik aplikasi, bukan daftar resmi tertutup.
[FAQ JLPT](https://www.jlpt.jp/e/faq/) menjelaskan bahwa spesifikasi daftar
kosakata, kanji, dan grammar untuk ujian yang direvisi tidak lagi dipublikasikan.

## Kuis dan progres

Pencarian judul/kanji dan filter tema, lanjut belajar, nilai terbaik, serta retry
mengikuti alur Kanji sebelumnya. Lulus dengan minimal 80%: 8/10 untuk lesson
lima kanji atau 7/8 untuk lesson empat kanji. Progres dihitung terpisah per level.
Kelulusan pertama memberi 50 XP; pengulangan mempertahankan status selesai dan
nilai terbaik serta tidak menggandakan XP.

Sesi dan progres backend masih disimpan di memori dan reset saat restart.
Login ulang setelah backend dijalankan ulang.

## Sumber dan lisensi

Jumlah goresan dan data bacaan memakai **KANJIDIC**, karya James William Breen
dan Electronic Dictionary Research and Development Group (EDRDG), melalui
[kanji-data oleh David Gouveia](https://github.com/davidluzgouveia/kanji-data).
Penetapan level menggunakan `jlpt_new` yang dalam dataset berasal dari
[Jonathan Waller JLPT Resources](https://www.tanos.co.uk/jlpt/).
Arti Inggris dan data WaniKani tidak digunakan dalam snapshot.

Data diadaptasi dengan pilihan bacaan, konversi on’yomi ke katakana, serta arti,
contoh kata, kalimat, pengingat, catatan, dan kuis Indonesia yang ditulis untuk
aplikasi ini. `backend/data/kanji/n3.json` dan snapshot kamus adalah materi data
adaptasi di bawah [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/).
Lisensi data ini tidak mengubah lisensi kode aplikasi.

- [Lisensi EDRDG](https://www.edrdg.org/edrdg/licence.html).
- Salinan lokal: [EDRDG](licenses/kanjidic-edrdg-license.html) dan
  [MIT konversi kanji-data](licenses/kanji-data-MIT.txt).
- Pengakuan sumber juga tampil pada **Tentang materi dan sumber** di katalog N3.

## Integrasi dan pembaruan

- Konten: `backend/data/kanji/n3.json`, di-embed saat backend dibuild.
- Metadata kamus: `docs/kanji-n3-reference.json`, diambil 6 Oktober 2026,
  dengan URL sumber dan SHA-256 file yang diunduh.
- Sumber materi: `scripts/build_kanji_n3.py`. Jalankan
  `python scripts/build_kanji_n3.py` untuk membuat ulang JSON secara deterministik.
  Tidak memerlukan jaringan atau paket Python tambahan saat regenerasi.
- Untuk memperbarui metadata, unduh ulang dataset sumber dan pilih entri
  `jlpt_new == 3`, lalu simpan hanya `strokes`, `readings_on`, dan `readings_kun`.
  Perbarui tanggal/hash, tinjau perubahan daftar, bacaan, serta materi sebelum
  regenerasi. Generator menolak kartu yang hilang atau contoh tanpa kanji target.
- Katalog: `GET /v1/kanji/course?level=N3`. Tanpa level tetap N5; N5/N4/N3
  tersedia. Level lain ditolak dengan HTTP 400.
- ID lesson: `kanji-n3-01` sampai `kanji-n3-77`. Endpoint detail, attempt,
  jawaban, dan penyelesaian sama dengan N5/N4; seluruh endpoint membutuhkan login.
- Kontrak: [OpenAPI](../backend/openapi.yaml).

Restart backend dan jalankan ulang Flutter untuk memuat materi serta pilihan
level baru. [Panduan menjalankan project](running-project.md).

## Kurikulum

| Lesson | Tema | Kanji |
|---|---|---|
| 01 | Orang dan hubungan 1 · 王他礼君客 | 王 他 礼 君 客 |
| 02 | Orang dan hubungan 2 · 相共老夫妻 | 相 共 老 夫 妻 |
| 03 | Orang dan hubungan 3 · 婦娘祖彼御 | 婦 娘 祖 彼 御 |
| 04 | Orang dan hubungan 4 · 皆交両婚 | 皆 交 両 婚 |
| 05 | Tubuh dan perasaan 1 · 耳声首頭顔 | 耳 声 首 頭 顔 |
| 06 | Tubuh dan perasaan 2 · 命息歯感情 | 命 息 歯 感 情 |
| 07 | Tubuh dan perasaan 3 · 念愛喜笑夢 | 念 愛 喜 笑 夢 |
| 08 | Tubuh dan perasaan 4 · 腹痛怒眠怖 | 腹 痛 怒 眠 怖 |
| 09 | Tubuh dan perasaan 5 · 恐恥髪疲欲 | 恐 恥 髪 疲 欲 |
| 10 | Tubuh dan perasaan 6 · 悲背抱優 | 悲 背 抱 優 |
| 11 | Alam dan lingkungan 1 · 石光草葉雪 | 石 光 草 葉 雪 |
| 12 | Alam dan lingkungan 2 · 馬猫庭流陽 | 馬 猫 庭 流 陽 |
| 13 | Alam dan lingkungan 3 · 球晴寒熱 | 球 晴 寒 熱 |
| 14 | Alam dan lingkungan 4 · 煙然船港 | 煙 然 船 港 |
| 15 | Alam dan lingkungan 5 · 神景園冷 | 神 景 園 冷 |
| 16 | Rumah dan keseharian 1 · 内太引米形 | 内 太 引 米 形 |
| 17 | Rumah dan keseharian 2 · 曲全具和宿 | 曲 全 具 和 宿 |
| 18 | Rumah dan keseharian 3 · 暗箱面座洗 | 暗 箱 面 座 洗 |
| 19 | Rumah dan keseharian 4 · 静財宅掛居 | 静 財 宅 掛 居 |
| 20 | Rumah dan keseharian 5 · 窓閉置寝靴 | 窓 閉 置 寝 靴 |
| 21 | Rumah dan keseharian 6 · 杯吸酒暮薬 | 杯 吸 酒 暮 薬 |
| 22 | Perjalanan dan gerakan 1 · 市向直所泳 | 市 向 直 所 泳 |
| 23 | Perjalanan dan gerakan 2 · 乗速進落路 | 乗 速 進 落 路 |
| 24 | Perjalanan dan gerakan 3 · 鳴横登都遠 | 鳴 横 登 都 遠 |
| 25 | Perjalanan dan gerakan 4 · 飛倒渡抜捕 | 飛 倒 渡 抜 捕 |
| 26 | Perjalanan dan gerakan 5 · 突退迷訪途 | 突 退 迷 訪 途 |
| 27 | Perjalanan dan gerakan 6 · 押浮越逃 | 押 浮 越 逃 |
| 28 | Perjalanan dan gerakan 7 · 吹到戻舞 | 吹 到 戻 舞 |
| 29 | Perjalanan dan gerakan 8 · 遊打段消 | 遊 打 段 消 |
| 30 | Belajar dan komunikasi 1 · 才号回次活 | 才 号 回 次 活 |
| 31 | Belajar dan komunikasi 2 · 点科由未末 | 点 科 由 未 末 |
| 32 | Belajar dan komunikasi 3 · 記組絵数表 | 記 組 絵 数 表 |
| 33 | Belajar dan komunikasi 4 · 部伝初例参 | 部 伝 初 例 参 |
| 34 | Belajar dan komunikasi 5 · 完想望能術 | 完 想 望 能 術 |
| 35 | Belajar dan komunikasi 6 · 期式信単法 | 期 式 信 単 法 |
| 36 | Belajar dan komunikasi 7 · 係辞覚説 | 係 辞 覚 説 |
| 37 | Belajar dan komunikasi 8 · 告種達類 | 告 種 達 類 |
| 38 | Belajar dan komunikasi 9 · 報識昨晩 | 報 識 昨 晩 |
| 39 | Belajar dan komunikasi 10 · 昔忘調番 | 昔 忘 調 番 |
| 40 | Pekerjaan dan kegiatan 1 · 支助役投決 | 支 助 役 投 決 |
| 41 | Pekerjaan dan kegiatan 2 · 受定努労働 | 受 定 努 労 働 |
| 42 | Pekerjaan dan kegiatan 3 · 商育指追放 | 商 育 指 追 放 |
| 43 | Pekerjaan dan kegiatan 4 · 取得続連忙 | 取 得 続 連 忙 |
| 44 | Pekerjaan dan kegiatan 5 · 徒機産議経 | 徒 機 産 議 経 |
| 45 | Pekerjaan dan kegiatan 6 · 済制務資件 | 済 制 務 資 件 |
| 46 | Pekerjaan dan kegiatan 7 · 任責収職規 | 任 責 収 職 規 |
| 47 | Pekerjaan dan kegiatan 8 · 割演師備給 | 割 演 師 備 給 |
| 48 | Pekerjaan dan kegiatan 9 · 処精積寄供 | 処 精 積 寄 供 |
| 49 | Pekerjaan dan kegiatan 10 · 招勤刻払頼 | 招 勤 刻 払 頼 |
| 50 | Pekerjaan dan kegiatan 11 · 頂申守許配 | 頂 申 守 許 配 |
| 51 | Masyarakat dan aturan 1 · 民対局戦政 | 民 対 局 戦 政 |
| 52 | Masyarakat dan aturan 2 · 関留険危存 | 関 留 険 危 存 |
| 53 | Masyarakat dan aturan 3 · 原側席察警 | 原 側 席 察 警 |
| 54 | Masyarakat dan aturan 4 · 罪犯権否 | 罪 犯 権 否 |
| 55 | Masyarakat dan aturan 5 · 官平等害 | 官 平 等 害 |
| 56 | Pilihan dan penilaian 1 · 欠付失必苦 | 欠 付 失 必 苦 |
| 57 | Pilihan dan penilaian 2 · 美負談要勝 | 美 負 談 要 勝 |
| 58 | Pilihan dan penilaian 3 · 好成位利良 | 好 成 位 利 良 |
| 59 | Pilihan dan penilaian 4 · 便深合予反 | 便 深 合 予 反 |
| 60 | Pilihan dan penilaian 5 · 返様福殺選 | 返 様 福 殺 選 |
| 61 | Pilihan dan penilaian 6 · 願疑折束残 | 願 疑 折 束 残 |
| 62 | Pilihan dan penilaian 7 · 求約性格列 | 求 約 性 格 列 |
| 63 | Pilihan dan penilaian 8 · 変幸治果因 | 変 幸 治 果 因 |
| 64 | Pilihan dan penilaian 9 · 常非加容易 | 常 非 加 容 易 |
| 65 | Pilihan dan penilaian 10 · 若余難確在 | 若 余 難 確 在 |
| 66 | Pilihan dan penilaian 11 · 論判断増解 | 論 判 断 増 解 |
| 67 | Pilihan dan penilaian 12 · 際認費示観 | 際 認 費 示 観 |
| 68 | Pilihan dan penilaian 13 · 値差限与候 | 値 差 限 与 候 |
| 69 | Pilihan dan penilaian 14 · 構慣満程絶 | 構 慣 満 程 絶 |
| 70 | Pilihan dan penilaian 15 · 貧適更盗除 | 貧 適 更 盗 除 |
| 71 | Pilihan dan penilaian 16 · 幾散似富探 | 幾 散 似 富 探 |
| 72 | Pilihan dan penilaian 17 · 迎雑賛込誤 | 迎 雑 賛 込 誤 |
| 73 | Pilihan dan penilaian 18 · 降遅破互緒 | 降 遅 破 互 緒 |
| 74 | Pilihan dan penilaian 19 · 歳偶偉当 | 歳 偶 偉 当 |
| 75 | Pilihan dan penilaian 20 · 化最争亡 | 化 最 争 亡 |
| 76 | Pilihan dan penilaian 21 · 敗困過状 | 敗 困 過 状 |
| 77 | Pilihan dan penilaian 22 · 現呼違実 | 現 呼 違 実 |
