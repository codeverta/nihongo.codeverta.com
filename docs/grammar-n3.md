# Grammar JLPT N3

Buka **Belajar → N3 → Grammar** setelah login, atau pilih chip N3 pada
katalog Grammar. Tersedia **36 lesson, 72 unit pola, 144 contoh kalimat, dan
180 soal**. Satu lesson terdiri dari dua unit pola dan lima soal.

Setiap unit memuat arti, rumus, penjelasan Indonesia, kesalahan umum, dan
dua contoh Jepang dengan bacaan kana serta terjemahan. Bacaan adalah baris
terpisah, bukan furigana per kanji. Pencarian pola/judul dan filter bab memakai
UI Grammar yang sama dengan N5/N4.

## Fokus pembelajaran

- Hubungan waktu, kesempatan, prasyarat, tahap kegiatan, dan hasil tindakan.
- Keputusan, usaha, kewajiban, serta pembeda “tidak boleh” dengan “harus”.
- Sebab positif/negatif, dugaan, ekspektasi, dan batas suatu generalisasi.
- Pertentangan, penekanan, sudut pandang, sumber informasi, dan perbandingan.
- Kecenderungan, keadaan ringan, tindakan belum selesai, definisi, dan laporan.
- Kausatif pasif, penggunaan izin secara sopan, sifat umum, dan nostalgia.

Beberapa pola dasar N4 diperdalam untuk membaca hubungan dan nuansa dalam
kalimat N3, misalnya 間／間に, らしい, について, ほど～ない, dan ても.
Unit pola bisa memuat beberapa variasi yang berkaitan. Semua contoh dan soal
ditulis untuk aplikasi, bukan salinan soal ujian.

Materi ini adalah kurikulum persiapan tingkat menengah milik aplikasi.
[Ringkasan resmi JLPT](https://www.jlpt.jp/e/about/levelsummary.html)
menempatkan N3 sebagai tingkat penghubung yang memahami bahasa Jepang sehari-hari
sampai taraf tertentu. [FAQ JLPT](https://www.jlpt.jp/e/faq/index.html)
menjelaskan bahwa spesifikasi daftar isi tes tidak lagi diterbitkan; karena itu
kurikulum ini tidak diklaim sebagai daftar tertutup seluruh pola resmi N3.

## Belajar dan kuis

1. Cari pola atau pilih bab, lalu buka lesson.
2. Pelajari dua pola, rumus, contoh, dan catatan kesalahan umum.
3. Kerjakan dua soal bentuk/pemakaian, dua soal pemahaman makna, serta satu
   soal memilih kalimat Jepang untuk situasi Indonesia.
4. Baca pembahasan setelah menjawab. Pilihan jawaban dirotasi oleh backend.
5. Lulus dengan nilai minimal 80% (4/5). Kelulusan pertama memberi 50 XP.
   Mengulang menjaga skor terbaik dan status selesai serta tidak menggandakan XP.

Progres terpisah untuk N5, N4, dan N3. Backend masih menyimpan akun/sesi/progres
di memori; restart menghapus sesi dan progres, sehingga perlu login ulang.

## Konten dan integrasi

- Sumber: `backend/data/grammar/n3.json`, di-embed saat backend dibuild.
- Script penulisan ulang konten: `python scripts/build_grammar_n3.py` dari root.
- Katalog: `GET /v1/grammar/course?level=N3`; tanpa level tetap N5.
- ID: `grammar-n3-01` sampai `grammar-n3-36`.
- Detail, attempt, jawaban, dan penyelesaian menggunakan endpoint Grammar
  yang sudah ada dan memerlukan login. Jawaban benar tetap disimpan di backend.
- Kontrak: [OpenAPI](../backend/openapi.yaml).

Setelah mengubah materi, restart backend dengan `go run ./cmd/api` dari
folder backend. Untuk pilihan level baru, build/jalankan ulang Flutter.
Lihat [panduan menjalankan project](running-project.md).

## Daftar lesson

| No. | Bab | Lesson | Unit pola |
|---|---|---|---|
| 01 | Waktu dan kesempatan | Selagi keadaan masih berlangsung | ～うちに; ～ているうちに |
| 02 | Waktu dan kesempatan | Rentang waktu dan kejadian di dalamnya | ～間／～間に; ～最中に |
| 03 | Waktu dan kesempatan | Syarat urutan dan waktu mendadak | ～てからでないと; ～たとたん（に） |
| 04 | Waktu dan kesempatan | Setiap kejadian dan sambil sekalian | ～たびに; ～ついでに |
| 05 | Tahap dan hasil kegiatan | Hampir terjadi dan temuan setelah mencoba | ～ところだった; ～たところ |
| 06 | Tahap dan hasil kegiatan | Sejak suatu kejadian dan kegiatan terus-menerus | ～て以来; ～てばかりいる |
| 07 | Keputusan dan usaha | Kebiasaan yang diputuskan dan instruksi tidak langsung | ～ことにしている; ～ように言う |
| 08 | Keputusan dan usaha | Mencoba bertindak dan enggan bertindak | ～ようとする; ～ようとしない |
| 09 | Saran dan kewajiban | Tindakan yang seharusnya dilakukan | ～べきだ; ～べきではない |
| 10 | Saran dan kewajiban | Tidak perlu dan tidak punya pilihan | ～ことはない; ～しかない |
| 11 | Saran dan kewajiban | Tidak bisa menghindari tanggung jawab | ～わけにはいかない; ～ないわけにはいかない |
| 12 | Alasan dan akibat | Penyebab yang menguntungkan dan merugikan | ～おかげで; ～せいで／～せいだ |
| 13 | Alasan dan akibat | Penyebab formal dan emosi berlebihan | ～ため（に）：原因; ～あまり（に） |
| 14 | Penilaian dan penyangkalan | Tidak berarti demikian dan mustahil demikian | ～わけではない; ～わけがない |
| 15 | Dugaan dan kepastian | Ekspektasi yang gagal dan kepastian negatif | ～はずだった; ～はずがない |
| 16 | Dugaan dan kepastian | Keyakinan kuat dan kesimpulan yang tidak mutlak | ～に違いない; ～とは限らない |
| 17 | Dugaan dan informasi | Informasi belum pasti dan penampilan khas | ～らしい：推量・伝聞; Nらしい |
| 18 | Pertentangan dan penilaian | Tidak sesuai harapan dan kritik tajam | ～わりに（は）; ～くせに |
| 19 | Pertentangan dan penilaian | Standar suatu kategori dan dua sisi | ～にしては; ～反面 |
| 20 | Pertentangan dan penilaian | Meluruskan sebutan dan menolak kesimpulan | ～というより; ～からといって |
| 21 | Penambahan dan penekanan | Bukan hanya dan kejutan bahkan | ～ばかりでなく; ～さえ |
| 22 | Penambahan dan penekanan | Justru yang ditekankan dan cukup satu syarat | ～こそ; ～さえ～ば |
| 23 | Topik dan arah sikap | Mengenai topik dan terhadap sasaran | ～について; ～に対して |
| 24 | Sudut pandang dan peran | Bagi seseorang dan sebagai suatu peran | ～にとって; ～として |
| 25 | Sarana dan sumber informasi | Menurut keadaan dan menurut sumber | ～によって; ～によると／～によれば |
| 26 | Perbandingan dan derajat | Membandingkan dan tidak mencapai tingkat | ～に比べて; ～ほど～ない |
| 27 | Perbandingan dan derajat | Perubahan seiring tingkat dan gambaran intensitas | ～ば～ほど; ～くらい／～ほど：程度 |
| 28 | Pengandaian | Walaupun hipotesis dan prasyarat yang wajib | ～たとえ～ても; ～ないことには |
| 29 | Penggantian dan pertukaran | Sebagai pengganti dan mewakili orang | ～代わりに; ～に代わって |
| 30 | Kecenderungan dan keadaan | Sering terjadi dan gejala ringan | ～がち; ～気味 |
| 31 | Kecenderungan dan keadaan | Terkesan seperti dan penuh sesuatu | ～っぽい; ～だらけ |
| 32 | Hasil dan beban tindakan | Cukup dengan dan bebas dari keharusan | ～て済む; ～ずに済む／～ないで済む |
| 33 | Kemungkinan dan tahap belum selesai | Tidak ada cara dan baru sebagian | ～ようがない; ～かける／～かけの |
| 34 | Menjelaskan informasi | Melaporkan isi dan mendefinisikan istilah | ～ということだ; ～というのは |
| 35 | Kausatif dan kesopanan | Terpaksa melakukan dan izin yang dihargai | 使役受身：～させられる; ～させていただく |
| 36 | Kebiasaan dan kenangan | Sifat umum dan kebiasaan masa lalu | ～ものだ：一般的な性質; ～たものだ |
