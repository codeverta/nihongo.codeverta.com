"""Authored N3 cards with pinned KANJIDIC metadata. Data: CC BY-SA 4.0.

Run python scripts/build_kanji_n3.py from any directory. No network required.
"""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
# Character | meaning | word | kana | word meaning | sentence | kana | translation
# Sentences and Indonesian teaching text are authored for this application.
CARDS = """
才|bakat; usia|才能|さいのう|bakat|彼には音楽の才能があります。|かれには おんがくの さいのうが あります。|Dia mempunyai bakat musik.
王|raja|王様|おうさま|raja|物語には優しい王様が出てきます。|ものがたりには やさしい おうさまが でてきます。|Dalam cerita muncul seorang raja yang baik hati.
石|batu|石|いし|batu|川で丸い石を見つけました。|かわで まるい いしを みつけました。|Saya menemukan batu bulat di sungai.
内|bagian dalam|以内|いない|dalam batas|三日以内に返事をください。|みっか いないに へんじを ください。|Tolong beri jawaban dalam tiga hari.
太|gemuk; tebal|太い|ふとい|tebal; besar diameternya|太いペンで名前を書きました。|ふとい ペンで なまえを かきました。|Saya menulis nama dengan pena tebal.
引|menarik|引く|ひく|menarik|このドアは引いて開けます。|この ドアは ひいて あけます。|Pintu ini dibuka dengan ditarik.
市|kota; pasar|市民|しみん|warga kota|市民が公園に集まりました。|しみんが こうえんに あつまりました。|Warga kota berkumpul di taman.
他|lain|他人|たにん|orang lain|他人の意見も聞きましょう。|たにんの いけんも ききましょう。|Mari dengarkan juga pendapat orang lain.
号|nomor; sebutan|番号|ばんごう|nomor|受付で番号を呼ばれました。|うけつけで ばんごうを よばれました。|Nomor saya dipanggil di meja penerimaan.
平|datar; damai|平和|へいわ|perdamaian|平和な世界を願っています。|へいわな せかいを ねがっています。|Saya mengharapkan dunia yang damai.
打|memukul|打つ|うつ|memukul|ボールを強く打ちました。|ボールを つよく うちました。|Saya memukul bola dengan kuat.
申|menyampaikan secara merendah|申し込む|もうしこむ|mendaftar|来週の講座に申し込みました。|らいしゅうの こうざに もうしこみました。|Saya mendaftar kelas minggu depan.
礼|kesopanan; terima kasih|お礼|おれい|ucapan terima kasih|手伝ってくれた友達にお礼を言いました。|てつだってくれた ともだちに おれいを いいました。|Saya mengucapkan terima kasih kepada teman yang membantu.
耳|telinga|耳|みみ|telinga|小さな音に耳を澄ませました。|ちいさな おとに みみを すませました。|Saya memasang telinga untuk mendengar suara kecil.
交|berinteraksi; berpotongan|交流|こうりゅう|pertukaran; interaksi|留学生との交流を楽しみました。|りゅうがくせいとの こうりゅうを たのしみました。|Saya menikmati interaksi dengan pelajar asing.
光|cahaya|光|ひかり|cahaya|窓から朝の光が入ります。|まどから あさの ひかりが はいります。|Cahaya pagi masuk melalui jendela.
回|putaran; kali|今回|こんかい|kali ini|今回は一人で旅行します。|こんかいは ひとりで りょこうします。|Kali ini saya bepergian sendiri.
当|mengenai; tepat|当たる|あたる|kena; tepat|ボールが壁に当たりました。|ボールが かべに あたりました。|Bola mengenai dinding.
米|beras|米|こめ|beras|米を洗ってから炊きます。|こめを あらってから たきます。|Saya mencuci beras sebelum memasaknya.
声|suara manusia|声|こえ|suara|大きな声で名前を呼びました。|おおきな こえで なまえを よびました。|Saya memanggil namanya dengan suara keras.
形|bentuk|形|かたち|bentuk|この雲は猫の形に見えます。|この くもは ねこの かたちに みえます。|Awan ini terlihat berbentuk kucing.
草|rumput|草|くさ|rumput|庭の草を抜きました。|にわの くさを ぬきました。|Saya mencabut rumput di halaman.
化|berubah menjadi|変化|へんか|perubahan|町の変化に驚きました。|まちの へんかに おどろきました。|Saya terkejut melihat perubahan kota.
両|kedua; dua sisi|両親|りょうしん|kedua orang tua|両親に旅行の写真を送りました。|りょうしんに りょこうの しゃしんを おくりました。|Saya mengirim foto perjalanan kepada kedua orang tua.
全|seluruh|全部|ぜんぶ|semuanya|荷物を全部運びました。|にもつを ぜんぶ はこびました。|Saya mengangkut semua barang bawaan.
向|arah; menghadap|向かう|むかう|menuju|今、駅に向かっています。|いま、えきに むかっています。|Sekarang saya sedang menuju stasiun.
曲|melengkung; lagu|曲がる|まがる|berbelok|次の角を右に曲がってください。|つぎの かどを みぎに まがってください。|Tolong belok kanan di sudut berikutnya.
次|berikutnya|次回|じかい|lain kali; pertemuan berikutnya|次回は別の方法を試します。|じかいは べつの ほうほうを ためします。|Lain kali saya akan mencoba cara lain.
直|lurus; memperbaiki|直す|なおす|memperbaiki|作文の間違いを直しました。|さくぶんの まちがいを なおしました。|Saya memperbaiki kesalahan dalam karangan.
活|hidup; aktif|生活|せいかつ|kehidupan|日本での生活に慣れました。|にほんでの せいかつに なれました。|Saya sudah terbiasa dengan kehidupan di Jepang.
点|titik; nilai|点数|てんすう|nilai; skor|前より点数が上がりました。|まえより てんすうが あがりました。|Nilai saya meningkat dibanding sebelumnya.
科|bidang; bagian ilmu|科学|かがく|sains|科学の本を読んでいます。|かがくの ほんを よんでいます。|Saya sedang membaca buku sains.
首|leher|首|くび|leher|長く座ると首が疲れます。|ながく すわると くびが つかれます。|Leher saya lelah jika duduk lama.
欠|kurang; absen|欠席|けっせき|ketidakhadiran|体調が悪くて授業を欠席しました。|たいちょうが わるくて じゅぎょうを けっせきしました。|Saya tidak masuk kelas karena kondisi badan kurang baik.
由|alasan; asal|理由|りゆう|alasan|遅れた理由を説明しました。|おくれた りゆうを せつめいしました。|Saya menjelaskan alasan keterlambatan.
民|rakyat; penduduk|住民|じゅうみん|penduduk|住民が町の地図を作りました。|じゅうみんが まちの ちずを つくりました。|Penduduk membuat peta kota.
付|melekat; menambahkan|付ける|つける|memasang; menambahkan|かばんに名前を付けました。|かばんに なまえを つけました。|Saya memasang nama pada tas.
失|kehilangan; gagal|失敗|しっぱい|kegagalan|失敗から多くのことを学びました。|しっぱいから おおくの ことを まなびました。|Saya belajar banyak hal dari kegagalan.
必|pasti; harus|必要|ひつよう|perlu|申し込みには写真が必要です。|もうしこみには しゃしんが ひつようです。|Foto diperlukan untuk pendaftaran.
未|belum|未来|みらい|masa depan|未来の仕事について考えています。|みらいの しごとについて かんがえています。|Saya memikirkan pekerjaan di masa depan.
末|akhir|週末|しゅうまつ|akhir pekan|週末は図書館に行きます。|しゅうまつは としょかんに いきます。|Saya akan pergi ke perpustakaan pada akhir pekan.
記|mencatat|記録|きろく|catatan; rekaman|毎日の運動を記録しています。|まいにちの うんどうを きろくしています。|Saya mencatat olahraga setiap hari.
組|kelompok; menyusun|番組|ばんぐみ|acara siaran|日本語の番組を見ました。|にほんごの ばんぐみを みました。|Saya menonton acara berbahasa Jepang.
船|kapal|船|ふね|kapal|船で島へ渡りました。|ふねで しまへ わたりました。|Saya menyeberang ke pulau dengan kapal.
雪|salju|雪|ゆき|salju|朝から雪が降っています。|あさから ゆきが ふっています。|Salju turun sejak pagi.
支|menopang; mendukung|支える|ささえる|menopang; mendukung|友達が私を支えてくれました。|ともだちが わたしを ささえてくれました。|Teman mendukung saya.
助|membantu|助ける|たすける|membantu|困っている人を助けたいです。|こまっている ひとを たすけたいです。|Saya ingin membantu orang yang kesulitan.
君|kamu; sapaan akrab|君|きみ|kamu, akrab|君の意見も聞きたいです。|きみの いけんも ききたいです。|Saya juga ingin mendengar pendapatmu.
対|menghadapi; pasangan|反対|はんたい|menentang; berlawanan|私はその案に反対です。|わたしは その あんに はんたいです。|Saya menentang usulan itu.
局|kantor; biro|郵便局|ゆうびんきょく|kantor pos|郵便局で荷物を送りました。|ゆうびんきょくで にもつを おくりました。|Saya mengirim paket di kantor pos.
役|peran; tugas|役に立つ|やくにたつ|berguna|この地図は役に立ちます。|この ちずは やくに たちます。|Peta ini berguna.
投|melempar|投げる|なげる|melempar|ボールを友達に投げました。|ボールを ともだちに なげました。|Saya melempar bola kepada teman.
決|memutuskan|決める|きめる|memutuskan|旅行の日を決めました。|りょこうの ひを きめました。|Saya menentukan tanggal perjalanan.
馬|kuda|馬|うま|kuda|牧場で馬を見ました。|ぼくじょうで うまを みました。|Saya melihat kuda di peternakan.
番|giliran; nomor urut|順番|じゅんばん|urutan; giliran|順番を守って並びましょう。|じゅんばんを まもって ならびましょう。|Mari antre dengan mengikuti urutan.
絵|gambar; lukisan|絵|え|gambar; lukisan|妹は花の絵を描きました。|いもうとは はなの えを かきました。|Adik perempuan menggambar bunga.
数|jumlah; menghitung|数える|かぞえる|menghitung|参加者の数を数えました。|さんかしゃの かずを かぞえました。|Saya menghitung jumlah peserta.
所|tempat|住所|じゅうしょ|alamat|ここに住所を書いてください。|ここに じゅうしょを かいてください。|Tolong tulis alamat di sini.
具|alat; kelengkapan|道具|どうぐ|alat|使った道具を片付けました。|つかった どうぐを かたづけました。|Saya merapikan alat yang sudah dipakai.
受|menerima|受ける|うける|menerima; mengikuti|来月、試験を受けます。|らいげつ、しけんを うけます。|Saya akan mengikuti ujian bulan depan.
和|harmoni; gaya Jepang|和食|わしょく|masakan Jepang|和食の作り方を習いました。|わしょくの つくりかたを ならいました。|Saya belajar cara membuat masakan Jepang.
定|menetapkan; tetap|予定|よてい|rencana; jadwal|明日の予定を確認しました。|あしたの よていを かくにんしました。|Saya memeriksa jadwal besok.
実|kenyataan; buah|実際|じっさい|sebenarnya; secara nyata|実際に使ってみると便利でした。|じっさいに つかってみると べんりでした。|Setelah mencoba memakainya langsung, ternyata praktis.
泳|berenang|泳ぐ|およぐ|berenang|夏は海で泳ぎます。|なつは うみで およぎます。|Pada musim panas saya berenang di laut.
苦|pahit; sulit; menderita|苦手|にがて|kurang mahir; tidak suka|私は早起きが苦手です。|わたしは はやおきが にがてです。|Saya sulit bangun pagi.
表|permukaan; menunjukkan|発表|はっぴょう|presentasi; pengumuman|研究の結果を発表しました。|けんきゅうの けっかを はっぴょうしました。|Saya mempresentasikan hasil penelitian.
部|bagian; departemen|部分|ぶぶん|bagian|分からない部分に印を付けました。|わからない ぶぶんに しるしを つけました。|Saya memberi tanda pada bagian yang tidak dipahami.
乗|naik kendaraan|乗る|のる|naik kendaraan|駅でバスに乗りました。|えきで バスに のりました。|Saya naik bus di stasiun.
客|tamu; pelanggan|お客さん|おきゃくさん|pelanggan; tamu|お客さんに席を案内しました。|おきゃくさんに せきを あんないしました。|Saya mengantar pelanggan ke tempat duduk.
相|saling; pihak lain|相談|そうだん|konsultasi|先生に進学の相談をしました。|せんせいに しんがくの そうだんを しました。|Saya berkonsultasi kepada guru tentang melanjutkan pendidikan.
美|indah|美しい|うつくしい|indah|山から美しい景色が見えました。|やまから うつくしい けしきが みえました。|Dari gunung terlihat pemandangan indah.
負|kalah; menanggung|負ける|まける|kalah|試合に負けても練習を続けます。|しあいに まけても れんしゅうを つづけます。|Walaupun kalah pertandingan, saya tetap berlatih.
談|berbicara; membahas|相談|そうだん|konsultasi|困った時は友達に相談します。|こまった ときは ともだちに そうだんします。|Saat kesulitan, saya berkonsultasi dengan teman.
要|perlu; pokok|重要|じゅうよう|penting|これは重要な知らせです。|これは じゅうような しらせです。|Ini pemberitahuan penting.
勝|menang|勝つ|かつ|menang|最後の試合で勝ちました。|さいごの しあいで かちました。|Saya menang dalam pertandingan terakhir.
速|cepat dalam gerak|速度|そくど|kecepatan|車の速度を落としました。|くるまの そくどを おとしました。|Saya menurunkan kecepatan mobil.
配|membagikan; mengatur|心配|しんぱい|khawatir|家族のことを心配しています。|かぞくの ことを しんぱいしています。|Saya mengkhawatirkan keluarga.
酒|minuman beralkohol|お酒|おさけ|minuman beralkohol|この料理にはお酒を使います。|この りょうりには おさけを つかいます。|Masakan ini menggunakan minuman beralkohol.
進|maju|進む|すすむ|maju; berkembang|計画が順調に進んでいます。|けいかくが じゅんちょうに すすんでいます。|Rencana berkembang dengan lancar.
落|jatuh; menjatuhkan|落とす|おとす|menjatuhkan|駅で財布を落としてしまいました。|えきで さいふを おとしてしまいました。|Saya tanpa sengaja menjatuhkan dompet di stasiun.
葉|daun|葉|は|daun|秋になると葉の色が変わります。|あきに なると はの いろが かわります。|Saat musim gugur, warna daun berubah.
路|jalan; jalur|道路|どうろ|jalan raya|この道路は夜も明るいです。|この どうろは よるも あかるいです。|Jalan ini terang bahkan pada malam hari.
鳴|berbunyi; bersuara hewan|鳴く|なく|berkicau; bersuara hewan|朝、鳥が鳴いていました。|あさ、とりが ないていました。|Burung berkicau pada pagi hari.
横|samping; mendatar|横|よこ|samping|机の横にかばんを置きました。|つくえの よこに かばんを おきました。|Saya meletakkan tas di samping meja.
調|menyelidiki; kondisi|調べる|しらべる|memeriksa; mencari tahu|電車の時間を調べました。|でんしゃの じかんを しらべました。|Saya mencari jadwal kereta.
頭|kepala|頭|あたま|kepala|難しい問題で頭を使いました。|むずかしい もんだいで あたまを つかいました。|Saya berpikir keras untuk soal yang sulit.
顔|wajah|顔|かお|wajah|久しぶりに友達の顔を見ました。|ひさしぶりに ともだちの かおを みました。|Saya melihat wajah teman setelah lama tidak bertemu.
最|paling|最近|さいきん|akhir-akhir ini|最近、日本語の本をよく読みます。|さいきん、にほんごの ほんを よく よみます。|Akhir-akhir ini saya sering membaca buku bahasa Jepang.
争|bersaing; berselisih|争う|あらそう|bersaing; berebut|二つのチームが優勝を争いました。|ふたつの チームが ゆうしょうを あらそいました。|Dua tim bersaing memperebutkan juara.
伝|menyampaikan|伝える|つたえる|menyampaikan|先生に欠席の理由を伝えました。|せんせいに けっせきの りゆうを つたえました。|Saya menyampaikan alasan tidak hadir kepada guru.
共|bersama|共通|きょうつう|kesamaan; bersama|二人には共通の趣味があります。|ふたりには きょうつうの しゅみが あります。|Keduanya memiliki hobi yang sama.
好|suka|好み|このみ|selera; kesukaan|人によって好みが違います。|ひとによって このみが ちがいます。|Selera berbeda menurut orangnya.
成|menjadi; mencapai|成功|せいこう|keberhasilan|みんなの協力で成功しました。|みんなの きょうりょくで せいこうしました。|Kami berhasil berkat kerja sama semua orang.
老|tua|老人|ろうじん|orang lanjut usia|老人に席を譲りました。|ろうじんに せきを ゆずりました。|Saya memberikan tempat duduk kepada orang lanjut usia.
位|posisi; peringkat|一位|いちい|peringkat pertama|大会で一位になりました。|たいかいで いちいに なりました。|Saya meraih peringkat pertama dalam kompetisi.
初|pertama; permulaan|初めて|はじめて|untuk pertama kali|初めて一人で飛行機に乗りました。|はじめて ひとりで ひこうきに のりました。|Saya naik pesawat sendiri untuk pertama kali.
利|manfaat; keuntungan|便利|べんり|praktis|このアプリは旅行に便利です。|この アプリは りょこうに べんりです。|Aplikasi ini praktis untuk perjalanan.
努|berusaha|努力|どりょく|usaha|毎日の努力が大切です。|まいにちの どりょくが たいせつです。|Usaha setiap hari itu penting.
労|jerih payah; bekerja|苦労|くろう|kesulitan; jerih payah|部屋を探すのに苦労しました。|へやを さがすのに くろうしました。|Saya mengalami kesulitan mencari kamar.
命|nyawa; perintah|命|いのち|nyawa|命を大切にしてください。|いのちを たいせつに してください。|Tolong hargai kehidupan.
放|melepaskan; menyiarkan|放送|ほうそう|siaran|駅の放送を聞きました。|えきの ほうそうを ききました。|Saya mendengarkan pengumuman siaran stasiun.
昔|dahulu|昔|むかし|dahulu|昔、この道によく来ました。|むかし、この みちに よく きました。|Dahulu saya sering datang ke jalan ini.
育|membesarkan; tumbuh|育てる|そだてる|membesarkan; merawat tumbuh|ベランダで花を育てています。|ベランダで はなを そだてています。|Saya merawat bunga di balkon.
指|jari; menunjuk|指|ゆび|jari|指で地図の場所を示しました。|ゆびで ちずの ばしょを しめしました。|Saya menunjukkan lokasi di peta dengan jari.
神|dewa|神社|じんじゃ|kuil Shinto|旅行で古い神社を訪ねました。|りょこうで ふるい じんじゃを たずねました。|Saya mengunjungi kuil Shinto tua saat bepergian.
追|mengejar|追う|おう|mengejar|犬がボールを追いました。|いぬが ボールを おいました。|Anjing mengejar bola.
戦|pertempuran; pertandingan|挑戦|ちょうせん|tantangan; mencoba tantangan|新しい仕事に挑戦したいです。|あたらしい しごとに ちょうせんしたいです。|Saya ingin mencoba tantangan pekerjaan baru.
良|baik|良い|よい|baik|良い方法を見つけました。|よい ほうほうを みつけました。|Saya menemukan cara yang baik.
便|kemudahan; kiriman|不便|ふべん|tidak praktis|駅が遠くて不便です。|えきが とおくて ふべんです。|Stasiunnya jauh sehingga tidak praktis.
働|bekerja|働く|はたらく|bekerja|週末は店で働いています。|しゅうまつは みせで はたらいています。|Saya bekerja di toko pada akhir pekan.
庭|halaman; taman rumah|庭|にわ|halaman|庭に小さな木を植えました。|にわに ちいさな きを うえました。|Saya menanam pohon kecil di halaman.
息|napas|息|いき|napas|階段を上って息が切れました。|かいだんを のぼって いきが きれました。|Saya kehabisan napas setelah menaiki tangga.
流|mengalir|流れる|ながれる|mengalir|川の水がゆっくり流れています。|かわの みずが ゆっくり ながれています。|Air sungai mengalir perlahan.
消|menghilangkan; padam|消す|けす|mematikan; menghapus|部屋を出る前に電気を消しました。|へやを でる まえに でんきを けしました。|Saya mematikan lampu sebelum meninggalkan kamar.
都|ibu kota; metropolitan|都市|とし|kota|大きな都市では電車が便利です。|おおきな としでは でんしゃが べんりです。|Di kota besar kereta itu praktis.
商|perdagangan|商品|しょうひん|barang dagangan|新しい商品が店に並びました。|あたらしい しょうひんが みせに ならびました。|Barang dagangan baru dipajang di toko.
深|dalam|深い|ふかい|dalam|この川は深いです。|この かわは ふかいです。|Sungai ini dalam.
球|bola; bulatan|地球|ちきゅう|bumi|地球の環境について学びました。|ちきゅうの かんきょうについて まなびました。|Saya belajar tentang lingkungan bumi.
陽|sinar matahari; positif|太陽|たいよう|matahari|雲の間から太陽が見えました。|くもの あいだから たいようが みえました。|Matahari terlihat di antara awan.
寒|dingin cuaca|寒い|さむい|dingin, cuaca|今朝はとても寒かったです。|けさは とても さむかったです。|Pagi ini sangat dingin.
悲|sedih|悲しい|かなしい|sedih|その話を聞いて悲しくなりました。|その はなしを きいて かなしく なりました。|Saya menjadi sedih setelah mendengar cerita itu.
期|periode; waktu yang ditentukan|期間|きかん|periode; jangka waktu|申し込みの期間を確認しました。|もうしこみの きかんを かくにんしました。|Saya memeriksa periode pendaftaran.
歯|gigi|歯|は|gigi|寝る前に歯を磨きます。|ねる まえに はを みがきます。|Saya menyikat gigi sebelum tidur.
港|pelabuhan|港|みなと|pelabuhan|港に大きな船が着きました。|みなとに おおきな ふねが つきました。|Kapal besar tiba di pelabuhan.
登|mendaki; naik|登る|のぼる|mendaki|友達と山に登りました。|ともだちと やまに のぼりました。|Saya mendaki gunung dengan teman.
亡|meninggal; lenyap|亡くなる|なくなる|meninggal dunia|祖父は去年亡くなりました。|そふは きょねん なくなりました。|Kakek saya meninggal tahun lalu.
合|cocok; bergabung|合う|あう|cocok|この靴は足に合います。|この くつは あしに あいます。|Sepatu ini pas di kaki.
予|sebelumnya; perkiraan|予約|よやく|reservasi|ホテルを予約しました。|ホテルを よやくしました。|Saya memesan hotel.
反|berlawanan; menentang|反応|はんのう|reaksi|新しい案への反応を聞きました。|あたらしい あんへの はんのうを ききました。|Saya menanyakan reaksi terhadap usulan baru.
返|mengembalikan|返す|かえす|mengembalikan|借りた本を返しました。|かりた ほんを かえしました。|Saya mengembalikan buku yang dipinjam.
宿|penginapan; bermalam|宿題|しゅくだい|pekerjaan rumah|宿題を終えてから出かけます。|しゅくだいを おえてから でかけます。|Saya pergi setelah menyelesaikan pekerjaan rumah.
想|membayangkan; pikiran|感想|かんそう|kesan; tanggapan|映画の感想を書きました。|えいがの かんそうを かきました。|Saya menulis kesan tentang film.
感|perasaan; merasakan|感動|かんどう|tersentuh; terharu|友達の言葉に感動しました。|ともだちの ことばに かんどうしました。|Saya tersentuh oleh kata-kata teman.
暗|gelap|暗い|くらい|gelap|この道は夜になると暗いです。|この みちは よるに なると くらいです。|Jalan ini gelap ketika malam.
様|keadaan; sapaan hormat|様子|ようす|keadaan; situasi|窓から外の様子を見ました。|まどから そとの ようすを みました。|Saya melihat keadaan di luar melalui jendela.
福|keberuntungan; kebahagiaan|幸福|こうふく|kebahagiaan|家族と過ごす時間に幸福を感じます。|かぞくと すごす じかんに こうふくを かんじます。|Saya merasakan kebahagiaan saat bersama keluarga.
殺|membunuh|殺す|ころす|membunuh|その小説では誰も殺されません。|その しょうせつでは だれも ころされません。|Dalam novel itu tidak ada orang yang dibunuh.
然|demikian; keadaan alami|自然|しぜん|alam; alami|自然の中を歩きました。|しぜんの なかを あるきました。|Saya berjalan di alam.
熱|panas; antusias|熱|ねつ|panas; demam|今日は熱があるので休みます。|きょうは ねつが あるので やすみます。|Hari ini saya beristirahat karena demam.
選|memilih|選ぶ|えらぶ|memilih|自分に合う本を選びました。|じぶんに あう ほんを えらびました。|Saya memilih buku yang cocok untuk saya.
願|berharap; memohon|願う|ねがう|mengharapkan|みんなの成功を願っています。|みんなの せいこうを ねがっています。|Saya mengharapkan keberhasilan semua orang.
情|perasaan; informasi keadaan|情報|じょうほう|informasi|旅行の情報を集めました。|りょこうの じょうほうを あつめました。|Saya mengumpulkan informasi perjalanan.
疑|meragukan|疑問|ぎもん|pertanyaan; keraguan|説明を聞いて疑問が残りました。|せつめいを きいて ぎもんが のこりました。|Masih ada pertanyaan setelah mendengar penjelasan.
皆|semua orang|皆さん|みなさん|semua orang, sopan|皆さんの意見を聞きたいです。|みなさんの いけんを ききたいです。|Saya ingin mendengar pendapat Anda semua.
例|contoh|例えば|たとえば|misalnya|例えば、こんな方法があります。|たとえば、こんな ほうほうが あります。|Misalnya, ada cara seperti ini.
参|berpartisipasi; datang merendah|参加|さんか|partisipasi|町の活動に参加しました。|まちの かつどうに さんかしました。|Saya mengikuti kegiatan kota.
完|selesai sepenuhnya|完成|かんせい|penyelesaian|新しい図書館が完成しました。|あたらしい としょかんが かんせいしました。|Perpustakaan baru selesai dibangun.
念|pikiran; perhatian|残念|ざんねん|disayangkan; kecewa|会えなくて残念です。|あえなくて ざんねんです。|Saya kecewa karena tidak bisa bertemu.
折|melipat; patah|折る|おる|melipat; mematahkan|紙を半分に折りました。|かみを はんぶんに おりました。|Saya melipat kertas menjadi dua.
望|mengharapkan; memandang jauh|希望|きぼう|harapan|希望する時間を教えてください。|きぼうする じかんを おしえてください。|Tolong beri tahu waktu yang Anda inginkan.
束|ikatan; janji gabungan|約束|やくそく|janji|友達との約束を守りました。|ともだちとの やくそくを まもりました。|Saya menepati janji kepada teman.
残|tersisa; meninggalkan|残る|のこる|tersisa|まだ時間が少し残っています。|まだ じかんが すこし のこっています。|Masih tersisa sedikit waktu.
求|meminta; mencari|求める|もとめる|meminta; mencari|店は新しいスタッフを求めています。|みせは あたらしい スタッフを もとめています。|Toko mencari staf baru.
約|janji; kira-kira|約|やく|kira-kira|駅まで約十分かかります。|えきまで やく じゅっぷん かかります。|Perjalanan ke stasiun memakan waktu sekitar sepuluh menit.
性|sifat; jenis|性格|せいかく|kepribadian|彼は明るい性格です。|かれは あかるい せいかくです。|Dia berkepribadian ceria.
格|karakter; standar|合格|ごうかく|lulus|試験に合格しました。|しけんに ごうかくしました。|Saya lulus ujian.
能|kemampuan|能力|のうりょく|kemampuan|自分の能力を伸ばしたいです。|じぶんの のうりょくを のばしたいです。|Saya ingin mengembangkan kemampuan saya.
術|teknik; keterampilan|技術|ぎじゅつ|teknologi; teknik|新しい技術を学んでいます。|あたらしい ぎじゅつを まなんでいます。|Saya sedang mempelajari teknologi baru.
晴|cerah|晴れる|はれる|menjadi cerah|午後には晴れるそうです。|ごごには はれるそうです。|Kabarnya cuaca akan cerah pada sore hari.
列|barisan|列|れつ|barisan; antrean|店の前に長い列ができました。|みせの まえに ながい れつが できました。|Antrean panjang terbentuk di depan toko.
式|upacara; bentuk|式|しき|upacara|卒業の式に参加しました。|そつぎょうの しきに さんかしました。|Saya mengikuti upacara kelulusan.
信|percaya; pesan|信じる|しんじる|percaya|友達の言葉を信じています。|ともだちの ことばを しんじています。|Saya mempercayai kata-kata teman.
単|satu; sederhana|簡単|かんたん|mudah; sederhana|簡単な言葉で説明しました。|かんたんな ことばで せつめいしました。|Saya menjelaskan dengan kata-kata sederhana.
変|berubah; aneh|変わる|かわる|berubah|予定が急に変わりました。|よていが きゅうに かわりました。|Rencana tiba-tiba berubah.
夫|suami|夫|おっと|suami sendiri|夫と夕食を作りました。|おっとと ゆうしょくを つくりました。|Saya membuat makan malam bersama suami.
昨|kemarin; sebelumnya|昨年|さくねん|tahun lalu, formal|昨年、この町に引っ越しました。|さくねん、この まちに ひっこしました。|Saya pindah ke kota ini tahun lalu.
法|cara; hukum|方法|ほうほう|cara; metode|別の方法で試してみましょう。|べつの ほうほうで ためしてみましょう。|Mari coba dengan cara lain.
晩|malam|今晩|こんばん|malam ini|今晩は家で映画を見ます。|こんばんは いえで えいがを みます。|Malam ini saya menonton film di rumah.
猫|kucing|猫|ねこ|kucing|猫が窓のそばで寝ています。|ねこが まどの そばで ねています。|Kucing tidur di dekat jendela.
園|taman|公園|こうえん|taman umum|公園で昼ご飯を食べました。|こうえんで ひるごはんを たべました。|Saya makan siang di taman.
遠|jauh|遠い|とおい|jauh|学校は家から遠いです。|がっこうは いえから とおいです。|Sekolah jauh dari rumah.
係|penanggung jawab; hubungan|係|かかり|petugas; penanggung jawab|受付の係に聞きました。|うけつけの かかりに ききました。|Saya bertanya kepada petugas penerimaan.
取|mengambil|取る|とる|mengambil|棚から本を取りました。|たなから ほんを とりました。|Saya mengambil buku dari rak.
守|menjaga; mematuhi|守る|まもる|menjaga; mematuhi|交通のルールを守りましょう。|こうつうの ルールを まもりましょう。|Mari patuhi aturan lalu lintas.
幸|bahagia; keberuntungan|幸せ|しあわせ|kebahagiaan; bahagia|家族と一緒にいると幸せです。|かぞくと いっしょに いると しあわせです。|Saya bahagia ketika bersama keluarga.
箱|kotak|箱|はこ|kotak|箱に本を入れました。|はこに ほんを いれました。|Saya memasukkan buku ke kotak.
面|permukaan; sisi|画面|がめん|layar|画面の文字が小さいです。|がめんの もじが ちいさいです。|Huruf pada layar kecil.
喜|gembira|喜ぶ|よろこぶ|bergembira|母はプレゼントを喜びました。|ははは プレゼントを よろこびました。|Ibu senang menerima hadiah.
治|mengobati; mengatur|治る|なおる|sembuh|風邪が治りました。|かぜが なおりました。|Saya sudah sembuh dari pilek.
笑|tertawa; tersenyum|笑う|わらう|tertawa; tersenyum|みんなで楽しく笑いました。|みんなで たのしく わらいました。|Kami semua tertawa dengan gembira.
辞|mengundurkan diri; kata|辞書|じしょ|kamus|分からない言葉を辞書で調べました。|わからない ことばを じしょで しらべました。|Saya mencari kata yang tidak diketahui di kamus.
関|berhubungan; penghalang|関係|かんけい|hubungan|二つの出来事には関係があります。|ふたつの できごとには かんけいが あります。|Ada hubungan antara kedua kejadian.
政|pemerintahan|政治|せいじ|politik|新聞で政治のニュースを読みました。|しんぶんで せいじの ニュースを よみました。|Saya membaca berita politik di koran.
留|tetap tinggal; menahan|留学|りゅうがく|belajar di luar negeri|来年、日本に留学します。|らいねん、にほんに りゅうがくします。|Tahun depan saya belajar di Jepang.
険|berbahaya; terjal|危険|きけん|bahaya; berbahaya|この道は夜に歩くと危険です。|この みちは よるに あるくと きけんです。|Jalan ini berbahaya jika dilalui berjalan malam hari.
危|bahaya|危ない|あぶない|berbahaya|ここで遊ぶのは危ないです。|ここで あそぶのは あぶないです。|Bermain di sini berbahaya.
存|ada; bertahan|存在|そんざい|keberadaan|その島の存在を知りませんでした。|その しまの そんざいを しりませんでした。|Saya tidak mengetahui keberadaan pulau itu.
原|asal; dataran|原因|げんいん|penyebab|遅れた原因を調べています。|おくれた げんいんを しらべています。|Saya sedang mencari penyebab keterlambatan.
薬|obat|薬|くすり|obat|薬の説明を読みました。|くすりの せつめいを よみました。|Saya membaca penjelasan obat.
側|sisi|両側|りょうがわ|kedua sisi|道の両側に店があります。|みちの りょうがわに みせが あります。|Ada toko di kedua sisi jalan.
席|tempat duduk|席|せき|tempat duduk|窓の近くの席に座りました。|まどの ちかくの せきに すわりました。|Saya duduk di kursi dekat jendela.
敗|kalah; gagal|失敗|しっぱい|kegagalan|失敗した理由を考えました。|しっぱいした りゆうを かんがえました。|Saya memikirkan alasan kegagalan.
果|hasil; buah|結果|けっか|hasil|試験の結果を待っています。|しけんの けっかを まっています。|Saya sedang menunggu hasil ujian.
因|sebab|原因|げんいん|penyebab|問題の原因が分かりました。|もんだいの げんいんが わかりました。|Penyebab masalah telah diketahui.
常|biasa; selalu|日常|にちじょう|keseharian|日常の会話を練習しています。|にちじょうの かいわを れんしゅうしています。|Saya berlatih percakapan sehari-hari.
識|pengetahuan; mengenali|知識|ちしき|pengetahuan|日本の文化の知識が増えました。|にほんの ぶんかの ちしきが ふえました。|Pengetahuan saya tentang budaya Jepang bertambah.
非|bukan; kesalahan|非常に|ひじょうに|sangat|非常に分かりやすい説明でした。|ひじょうに わかりやすい せつめいでした。|Penjelasannya sangat mudah dipahami.
官|pejabat; pemerintahan|警察官|けいさつかん|petugas polisi|警察官に道を聞きました。|けいさつかんに みちを ききました。|Saya bertanya arah kepada petugas polisi.
察|mengamati; menyelidiki|警察|けいさつ|polisi|落とした財布を警察に届けました。|おとした さいふを けいさつに とどけました。|Saya menyerahkan dompet yang terjatuh kepada polisi.
愛|cinta|愛情|あいじょう|kasih sayang|手紙から家族の愛情を感じました。|てがみから かぞくの あいじょうを かんじました。|Saya merasakan kasih sayang keluarga melalui surat.
警|memperingatkan; waspada|警告|けいこく|peringatan|画面に警告が出ました。|がめんに けいこくが でました。|Peringatan muncul di layar.
覚|mengingat; sadar|覚える|おぼえる|menghafal; mengingat|毎日五つの単語を覚えます。|まいにち いつつの たんごを おぼえます。|Saya menghafal lima kosakata setiap hari.
説|menjelaskan; teori|説明|せつめい|penjelasan|使い方の説明を聞きました。|つかいかたの せつめいを ききました。|Saya mendengarkan penjelasan cara penggunaan.
告|memberi tahu|報告|ほうこく|laporan|仕事の結果を報告しました。|しごとの けっかを ほうこくしました。|Saya melaporkan hasil pekerjaan.
種|jenis; benih|種類|しゅるい|jenis|この店には多くの種類のパンがあります。|この みせには おおくの しゅるいの パンが あります。|Toko ini memiliki banyak jenis roti.
達|mencapai; penanda kelompok|友達|ともだち|teman|友達と駅で待ち合わせました。|ともだちと えきで まちあわせました。|Saya janjian bertemu teman di stasiun.
類|jenis; golongan|書類|しょるい|dokumen|必要な書類を準備しました。|ひつような しょるいを じゅんびしました。|Saya menyiapkan dokumen yang diperlukan.
報|melaporkan; kabar|情報|じょうほう|informasi|正しい情報を確認しましょう。|ただしい じょうほうを かくにんしましょう。|Mari periksa informasi yang benar.
等|setara; dan sebagainya|平等|びょうどう|kesetaraan|みんなに平等にチャンスがあります。|みんなに びょうどうに チャンスが あります。|Semua orang memiliki kesempatan yang setara.
座|duduk; tempat duduk|座る|すわる|duduk|空いている席に座りました。|あいている せきに すわりました。|Saya duduk di kursi kosong.
忘|lupa|忘れる|わすれる|lupa|傘を電車に忘れました。|かさを でんしゃに わすれました。|Saya lupa membawa payung dari kereta.
洗|mencuci|洗う|あらう|mencuci|料理の前に手を洗います。|りょうりの まえに てを あらいます。|Saya mencuci tangan sebelum memasak.
許|mengizinkan; memaafkan|許可|きょか|izin|写真を撮る許可をもらいました。|しゃしんを とる きょかを もらいました。|Saya mendapat izin mengambil foto.
静|tenang; sunyi|静か|しずか|tenang; sunyi|図書館では静かに話しましょう。|としょかんでは しずかに はなしましょう。|Mari berbicara pelan di perpustakaan.
煙|asap|煙|けむり|asap|遠くに煙が見えました。|とおくに けむりが みえました。|Asap terlihat di kejauhan.
加|menambahkan|加える|くわえる|menambahkan|説明に写真を加えました。|せつめいに しゃしんを くわえました。|Saya menambahkan foto pada penjelasan.
容|isi; menerima|内容|ないよう|isi|本の内容を簡単に説明しました。|ほんの ないようを かんたんに せつめいしました。|Saya menjelaskan isi buku secara sederhana.
易|mudah|容易|ようい|mudah, formal|この問題は容易ではありません。|この もんだいは よういでは ありません。|Masalah ini tidak mudah.
財|harta; kekayaan|財布|さいふ|dompet|財布の中を確認しました。|さいふの なかを かくにんしました。|Saya memeriksa isi dompet.
若|muda|若い|わかい|muda|若い人の意見も聞きました。|わかい ひとの いけんも ききました。|Saya juga mendengar pendapat orang muda.
忙|sibuk|忙しい|いそがしい|sibuk|今日は仕事が忙しいです。|きょうは しごとが いそがしいです。|Pekerjaan saya sibuk hari ini.
徒|murid; pengikut|生徒|せいと|murid|生徒が質問しました。|せいとが しつもんしました。|Murid mengajukan pertanyaan.
得|mendapat; keuntungan|得意|とくい|mahir|彼は料理が得意です。|かれは りょうりが とくいです。|Dia mahir memasak.
続|melanjutkan|続ける|つづける|melanjutkan|毎日練習を続けています。|まいにち れんしゅうを つづけています。|Saya terus berlatih setiap hari.
連|menghubungkan; beriringan|連絡|れんらく|menghubungi; kontak|着いたら連絡してください。|ついたら れんらくしてください。|Tolong hubungi saya setelah tiba.
困|kesulitan|困る|こまる|mengalami kesulitan|道が分からなくて困りました。|みちが わからなくて こまりました。|Saya kesulitan karena tidak tahu jalan.
機|mesin; kesempatan|機会|きかい|kesempatan|日本語を話す機会が増えました。|にほんごを はなす きかいが ふえました。|Kesempatan berbicara bahasa Jepang bertambah.
飛|terbang|飛ぶ|とぶ|terbang|空を鳥が飛んでいます。|そらを とりが とんでいます。|Burung sedang terbang di langit.
害|kerugian; gangguan|被害|ひがい|kerusakan; dampak merugikan|台風の被害についてニュースで知りました。|たいふうの ひがいについて ニュースで しりました。|Saya mengetahui kerusakan akibat topan melalui berita.
余|tersisa; lebih|余る|あまる|tersisa|時間が少し余りました。|じかんが すこし あまりました。|Ada sedikit waktu tersisa.
難|sulit|難しい|むずかしい|sulit|この漢字は書くのが難しいです。|この かんじは かくのが むずかしいです。|Kanji ini sulit ditulis.
確|pasti; memastikan|確認|かくにん|pemeriksaan; memastikan|予約の日付を確認しました。|よやくの ひづけを かくにんしました。|Saya memeriksa tanggal reservasi.
在|berada; ada|現在|げんざい|saat ini|現在、東京に住んでいます。|げんざい、とうきょうに すんでいます。|Saat ini saya tinggal di Tokyo.
夢|mimpi|夢|ゆめ|mimpi|子供のころの夢を思い出しました。|こどもの ころの ゆめを おもいだしました。|Saya teringat mimpi saat kecil.
産|menghasilkan; melahirkan|生産|せいさん|produksi|この工場で部品を生産しています。|この こうじょうで ぶひんを せいさんしています。|Pabrik ini memproduksi komponen.
倒|roboh; menjatuhkan|倒れる|たおれる|roboh; jatuh|風で木が倒れました。|かぜで きが たおれました。|Pohon tumbang akibat angin.
妻|istri|妻|つま|istri sendiri|妻と旅行の予定を立てました。|つまと りょこうの よていを たてました。|Saya merencanakan perjalanan bersama istri.
議|membahas; musyawarah|会議|かいぎ|rapat|午後は会議があります。|ごごは かいぎが あります。|Ada rapat pada sore hari.
犯|melanggar; pelaku kejahatan|犯人|はんにん|pelaku kejahatan|小説の最後に犯人が分かりました。|しょうせつの さいごに はんにんが わかりました。|Pelaku kejahatan terungkap di akhir novel.
罪|dosa; kejahatan|犯罪|はんざい|kejahatan|犯罪についての記事を読みました。|はんざいについての きじを よみました。|Saya membaca artikel tentang kejahatan.
論|teori; pembahasan|議論|ぎろん|diskusi; perdebatan|新しい案について議論しました。|あたらしい あんについて ぎろんしました。|Kami mendiskusikan usulan baru.
経|melalui; mengelola|経験|けいけん|pengalaman|この仕事で多くの経験をしました。|この しごとで おおくの けいけんを しました。|Saya mendapat banyak pengalaman dari pekerjaan ini.
済|selesai; menyelesaikan|済む|すむ|selesai; cukup|電話するだけで済みました。|でんわするだけで すみました。|Urusannya selesai cukup dengan menelepon.
判|menilai; memutuskan|判断|はんだん|penilaian; keputusan|情報を集めてから判断します。|じょうほうを あつめてから はんだんします。|Saya menilai setelah mengumpulkan informasi.
制|sistem; mengendalikan|制度|せいど|sistem; aturan kelembagaan|新しい制度について説明を聞きました。|あたらしい せいどについて せつめいを ききました。|Saya mendengarkan penjelasan tentang sistem baru.
務|tugas; kewajiban|事務|じむ|administrasi kantor|事務の仕事をしています。|じむの しごとを しています。|Saya mengerjakan administrasi kantor.
資|sumber daya; modal|資料|しりょう|bahan; dokumen rujukan|会議の資料を読みました。|かいぎの しりょうを よみました。|Saya membaca bahan rapat.
権|hak; wewenang|権利|けんり|hak|自分の権利について学びました。|じぶんの けんりについて まなびました。|Saya belajar tentang hak saya.
件|perkara; penghitung kasus|条件|じょうけん|syarat|参加の条件を確認しました。|さんかの じょうけんを かくにんしました。|Saya memeriksa syarat partisipasi.
断|memutus; menolak|断る|ことわる|menolak|忙しかったので誘いを断りました。|いそがしかったので さそいを ことわりました。|Saya menolak ajakan karena sibuk.
任|tanggung jawab; menyerahkan|任せる|まかせる|mempercayakan|この仕事は彼に任せます。|この しごとは かれに まかせます。|Saya mempercayakan pekerjaan ini kepadanya.
責|tanggung jawab; menyalahkan|責任|せきにん|tanggung jawab|自分の仕事に責任を持ちます。|じぶんの しごとに せきにんを もちます。|Saya bertanggung jawab atas pekerjaan saya.
増|bertambah; menambah|増える|ふえる|bertambah|店の客が増えました。|みせの きゃくが ふえました。|Pelanggan toko bertambah.
解|memecahkan; memahami|理解|りかい|pemahaman|説明を聞いて理解できました。|せつめいを きいて りかいできました。|Saya bisa memahami setelah mendengar penjelasan.
際|kesempatan; batas|実際|じっさい|sebenarnya; secara nyata|実際の建物の写真を見ました。|じっさいの たてものの しゃしんを みました。|Saya melihat foto bangunan yang sebenarnya.
認|mengakui; mengenali|認める|みとめる|mengakui|自分の間違いを認めました。|じぶんの まちがいを みとめました。|Saya mengakui kesalahan saya.
過|melampaui; berlalu|過ごす|すごす|menghabiskan waktu|休みを家で過ごしました。|やすみを いえで すごしました。|Saya menghabiskan liburan di rumah.
寝|tidur|寝る|ねる|tidur|昨日は早く寝ました。|きのうは はやく ねました。|Kemarin saya tidur lebih awal.
置|meletakkan|置く|おく|meletakkan|鍵を机の上に置きました。|かぎを つくえの うえに おきました。|Saya meletakkan kunci di atas meja.
費|biaya; menghabiskan|費用|ひよう|biaya|旅行の費用を計算しました。|りょこうの ひようを けいさんしました。|Saya menghitung biaya perjalanan.
示|menunjukkan|示す|しめす|menunjukkan|先生が例を示しました。|せんせいが れいを しめしました。|Guru memberikan contoh.
観|melihat; pandangan|観光|かんこう|wisata|京都で観光を楽しみました。|きょうとで かんこうを たのしみました。|Saya menikmati wisata di Kyoto.
値|nilai; harga|値段|ねだん|harga|買う前に値段を確認しました。|かう まえに ねだんを かくにんしました。|Saya memeriksa harga sebelum membeli.
吸|mengisap; menghirup|吸う|すう|mengisap; menghirup|外で新鮮な空気を吸いました。|そとで しんせんな くうきを すいました。|Saya menghirup udara segar di luar.
状|keadaan; bentuk|状況|じょうきょう|situasi|今の状況を説明しました。|いまの じょうきょうを せつめいしました。|Saya menjelaskan situasi saat ini.
収|mengumpulkan; memperoleh|収入|しゅうにゅう|penghasilan|新しい仕事で収入が増えました。|あたらしい しごとで しゅうにゅうが ふえました。|Penghasilan meningkat berkat pekerjaan baru.
職|pekerjaan; jabatan|就職|しゅうしょく|mendapat pekerjaan|卒業してから就職しました。|そつぎょうしてから しゅうしょくしました。|Saya mendapat pekerjaan setelah lulus.
規|aturan; standar|規則|きそく|peraturan|学校の規則を守ります。|がっこうの きそくを まもります。|Saya mematuhi peraturan sekolah.
割|membagi; pecahan|割合|わりあい|proporsi; persentase|参加者の割合を調べました。|さんかしゃの わりあいを しらべました。|Saya memeriksa proporsi peserta.
演|menampilkan; pertunjukan|演奏|えんそう|pertunjukan musik|友達の演奏を聞きました。|ともだちの えんそうを ききました。|Saya mendengarkan penampilan musik teman.
師|guru; ahli|教師|きょうし|guru|教師になるのが夢です。|きょうしに なるのが ゆめです。|Mimpi saya menjadi guru.
備|menyiapkan|準備|じゅんび|persiapan|明日の授業の準備をしました。|あしたの じゅぎょうの じゅんびを しました。|Saya menyiapkan pelajaran besok.
優|unggul; lembut|優しい|やさしい|baik hati; lembut|先生はいつも優しいです。|せんせいは いつも やさしいです。|Guru selalu baik hati.
宅|rumah; kediaman|自宅|じたく|rumah sendiri|今日は自宅で仕事をします。|きょうは じたくで しごとを します。|Hari ini saya bekerja di rumah.
現|muncul; saat ini|現れる|あらわれる|muncul|雲の間に月が現れました。|くもの あいだに つきが あらわれました。|Bulan muncul di antara awan.
呼|memanggil|呼ぶ|よぶ|memanggil|受付で名前を呼ばれました。|うけつけで なまえを よばれました。|Nama saya dipanggil di meja penerimaan.
違|berbeda; salah|違う|ちがう|berbeda; salah|二人の意見は違います。|ふたりの いけんは ちがいます。|Pendapat keduanya berbeda.
差|perbedaan; selisih|差|さ|perbedaan; selisih|二つの商品の値段に差があります。|ふたつの しょうひんの ねだんに さが あります。|Ada selisih harga antara dua barang.
供|menyediakan; menawarkan|子供|こども|anak|子供たちが公園で遊んでいます。|こどもたちが こうえんで あそんでいます。|Anak-anak sedang bermain di taman.
限|batas|限り|かぎり|batas; sejauh|できる限り手伝います。|できる かぎり てつだいます。|Saya membantu semampu saya.
与|memberikan|与える|あたえる|memberikan|この経験は私に自信を与えました。|この けいけんは わたしに じしんを あたえました。|Pengalaman ini memberi saya rasa percaya diri.
渡|menyeberang; menyerahkan|渡る|わたる|menyeberang|橋を渡って駅に行きます。|はしを わたって えきに いきます。|Saya menyeberangi jembatan untuk pergi ke stasiun.
景|pemandangan|景色|けしき|pemandangan|窓から景色を楽しみました。|まどから けしきを たのしみました。|Saya menikmati pemandangan dari jendela.
抜|mencabut; melewati|抜く|ぬく|mencabut|庭の草を抜きました。|にわの くさを ぬきました。|Saya mencabut rumput di halaman.
候|keadaan musim; menunggu|候補|こうほ|calon; kandidat|旅行先の候補を三つ選びました。|りょこうさきの こうほを みっつ えらびました。|Saya memilih tiga calon tujuan perjalanan.
構|susunan; membangun|構成|こうせい|susunan; komposisi|文章の構成を考えました。|ぶんしょうの こうせいを かんがえました。|Saya memikirkan susunan tulisan.
捕|menangkap|捕まえる|つかまえる|menangkap|庭で虫を捕まえました。|にわで むしを つかまえました。|Saya menangkap serangga di halaman.
慣|terbiasa|慣れる|なれる|terbiasa|新しい仕事に慣れました。|あたらしい しごとに なれました。|Saya terbiasa dengan pekerjaan baru.
満|penuh; puas|満足|まんぞく|kepuasan; puas|旅行の内容に満足しました。|りょこうの ないように まんぞくしました。|Saya puas dengan isi perjalanan.
掛|menggantung; mengenakan|掛ける|かける|menggantung; mengenakan|壁に時計を掛けました。|かべに とけいを かけました。|Saya menggantung jam di dinding.
居|berada; tempat tinggal|居間|いま|ruang keluarga|居間で家族と話しました。|いまで かぞくと はなしました。|Saya berbicara dengan keluarga di ruang keluarga.
突|menonjok; tiba-tiba|突然|とつぜん|tiba-tiba|突然、雨が降り始めました。|とつぜん、あめが ふりはじめました。|Hujan tiba-tiba mulai turun.
招|mengundang|招待|しょうたい|undangan|友達を家に招待しました。|ともだちを いえに しょうたいしました。|Saya mengundang teman ke rumah.
段|tingkat; tangga|階段|かいだん|tangga|階段で二階に上がりました。|かいだんで にかいに あがりました。|Saya naik ke lantai dua melalui tangga.
腹|perut|腹|はら|perut|朝から腹が減っています。|あさから はらが へっています。|Saya merasa lapar sejak pagi.
痛|sakit; nyeri|痛い|いたい|sakit; nyeri|足が痛くて走れません。|あしが いたくて はしれません。|Kaki saya sakit sehingga tidak bisa berlari.
退|mundur; keluar dari posisi|退院|たいいん|keluar dari rumah sakit|祖母は昨日退院しました。|そぼは きのう たいいんしました。|Nenek keluar dari rumah sakit kemarin.
迷|bingung; tersesat|迷う|まよう|bingung; tersesat|駅への道で迷いました。|えきへの みちで まよいました。|Saya tersesat di jalan menuju stasiun.
訪|mengunjungi|訪ねる|たずねる|mengunjungi|京都の友達を訪ねました。|きょうとの ともだちを たずねました。|Saya mengunjungi teman di Kyoto.
怒|marah|怒る|おこる|marah|父は約束を守らなかった私に怒りました。|ちちは やくそくを まもらなかった わたしに おこりました。|Ayah marah kepada saya karena tidak menepati janji.
眠|tidur; mengantuk|眠い|ねむい|mengantuk|昨夜は遅く寝たので眠いです。|さくやは おそく ねたので ねむいです。|Saya mengantuk karena tidur larut tadi malam.
靴|sepatu|靴|くつ|sepatu|歩きやすい靴を選びました。|あるきやすい くつを えらびました。|Saya memilih sepatu yang nyaman untuk berjalan.
途|jalan; dalam perjalanan|途中|とちゅう|di tengah; dalam perjalanan|帰る途中でパンを買いました。|かえる とちゅうで パンを かいました。|Saya membeli roti dalam perjalanan pulang.
給|memberikan; gaji|給料|きゅうりょう|gaji|給料の一部を貯金しています。|きゅうりょうの いちぶを ちょきんしています。|Saya menabung sebagian gaji.
冷|dingin benda; mendinginkan|冷たい|つめたい|dingin, sentuhan|冷たい水を飲みました。|つめたい みずを のみました。|Saya minum air dingin.
処|menangani; tempat|処理|しょり|penanganan; pemrosesan|データを処理しています。|データを しょりしています。|Saya sedang memproses data.
婦|perempuan menikah|夫婦|ふうふ|pasangan suami istri|夫婦で店を開きました。|ふうふで みせを ひらきました。|Pasangan suami istri itu membuka toko bersama.
程|tingkat; rentang|程度|ていど|tingkat; sekitar|三十分程度かかります。|さんじゅっぷん ていど かかります。|Dibutuhkan waktu sekitar tiga puluh menit.
精|teliti; energi batin|精神|せいしん|mental; jiwa|精神的に疲れていました。|せいしんてきに つかれていました。|Saya merasa lelah secara mental.
絶|terputus; mutlak|絶対|ぜったい|mutlak; pasti|絶対に約束を忘れません。|ぜったいに やくそくを わすれません。|Saya pasti tidak akan melupakan janji.
杯|cangkir; penghitung minuman|一杯|いっぱい|satu cangkir atau gelas|水を一杯飲みました。|みずを いっぱい のみました。|Saya minum segelas air.
積|menumpuk|積む|つむ|menumpuk; memuat|車に荷物を積みました。|くるまに にもつを つみました。|Saya memuat barang ke mobil.
寄|mendekat; mampir|寄る|よる|mampir; mendekat|帰りに本屋に寄りました。|かえりに ほんやに よりました。|Saya mampir ke toko buku dalam perjalanan pulang.
娘|anak perempuan|娘|むすめ|anak perempuan|娘は絵を描くのが好きです。|むすめは えを かくのが すきです。|Anak perempuan saya suka menggambar.
怖|takut|怖い|こわい|menakutkan|この映画は少し怖いです。|この えいがは すこし こわいです。|Film ini agak menakutkan.
恐|takut; khawatir|恐れる|おそれる|takut akan|失敗を恐れずに挑戦しましょう。|しっぱいを おそれずに ちょうせんしましょう。|Mari mencoba tantangan tanpa takut gagal.
婚|pernikahan|結婚|けっこん|pernikahan; menikah|姉は去年結婚しました。|あねは きょねん けっこんしました。|Kakak perempuan saya menikah tahun lalu.
遊|bermain|遊ぶ|あそぶ|bermain|子供たちが外で遊んでいます。|こどもたちが そとで あそんでいます。|Anak-anak sedang bermain di luar.
貧|miskin|貧しい|まずしい|miskin|その物語は貧しい家族の話です。|その ものがたりは まずしい かぞくの はなしです。|Cerita itu mengenai keluarga miskin.
適|sesuai; tepat|適切|てきせつ|tepat; sesuai|適切な言葉を選びました。|てきせつな ことばを えらびました。|Saya memilih kata yang tepat.
押|mendorong; menekan|押す|おす|mendorong; menekan|このボタンを押してください。|この ボタンを おしてください。|Tolong tekan tombol ini.
更|lebih lanjut; memperbarui|変更|へんこう|perubahan|予定を変更しました。|よていを へんこうしました。|Saya mengubah rencana.
浮|mengapung|浮かぶ|うかぶ|mengapung; muncul dalam pikiran|いい考えが浮かびました。|いい かんがえが うかびました。|Ide bagus muncul di pikiran saya.
越|melampaui; menyeberangi|引っ越す|ひっこす|pindah tempat tinggal|来月、新しい部屋に引っ越します。|らいげつ、あたらしい へやに ひっこします。|Bulan depan saya pindah ke kamar baru.
背|punggung; tinggi badan|背|せ|tinggi badan; punggung|弟は私より背が高いです。|おとうとは わたしより せが たかいです。|Adik laki-laki saya lebih tinggi dari saya.
盗|mencuri|盗む|ぬすむ|mencuri|かばんを盗まれて困りました。|かばんを ぬすまれて こまりました。|Saya kesulitan karena tas saya dicuri.
除|menghilangkan; mengecualikan|掃除|そうじ|pembersihan|朝、部屋を掃除しました。|あさ、へやを そうじしました。|Saya membersihkan kamar pada pagi hari.
幾|berapa; beberapa|幾つ|いくつ|berapa banyak; berapa usia|りんごは幾つありますか。|りんごは いくつ ありますか。|Ada berapa apel?
散|menyebar|散歩|さんぽ|jalan-jalan|夕方、公園を散歩しました。|ゆうがた、こうえんを さんぽしました。|Saya berjalan-jalan di taman pada sore hari.
似|mirip|似る|にる|mirip|妹は母に似ています。|いもうとは ははに にています。|Adik perempuan mirip ibu.
富|kaya; kekayaan|豊富|ほうふ|melimpah; kaya akan|この店は種類が豊富です。|この みせは しゅるいが ほうふです。|Toko ini memiliki jenis barang yang beragam.
探|mencari|探す|さがす|mencari|図書館で本を探しました。|としょかんで ほんを さがしました。|Saya mencari buku di perpustakaan.
迎|menyambut; menjemput|迎える|むかえる|menyambut; menjemput|駅で友達を迎えました。|えきで ともだちを むかえました。|Saya menjemput teman di stasiun.
祖|leluhur|祖母|そぼ|nenek sendiri|祖母から手紙が届きました。|そぼから てがみが とどきました。|Surat dari nenek tiba.
雑|beragam; tidak teratur|雑誌|ざっし|majalah|旅行の雑誌を買いました。|りょこうの ざっしを かいました。|Saya membeli majalah perjalanan.
賛|menyetujui; memuji|賛成|さんせい|persetujuan|私はその計画に賛成です。|わたしは その けいかくに さんせいです。|Saya menyetujui rencana itu.
込|masuk ke dalam; padat|申し込む|もうしこむ|mendaftar|日本語の講座に申し込みました。|にほんごの こうざに もうしこみました。|Saya mendaftar kelas bahasa Jepang.
欲|menginginkan|欲しい|ほしい|ingin memiliki|新しい辞書が欲しいです。|あたらしい じしょが ほしいです。|Saya ingin kamus baru.
閉|menutup|閉める|しめる|menutup|寝る前に窓を閉めました。|ねる まえに まどを しめました。|Saya menutup jendela sebelum tidur.
窓|jendela|窓|まど|jendela|窓を開けて空気を入れました。|まどを あけて くうきを いれました。|Saya membuka jendela untuk memasukkan udara.
否|menyangkal; tidak|否定|ひてい|penyangkalan|彼はうわさを否定しました。|かれは うわさを ひていしました。|Dia menyangkal rumor itu.
暮|menjalani hidup; senja|暮らす|くらす|tinggal; menjalani hidup|静かな町で暮らしています。|しずかな まちで くらしています。|Saya tinggal di kota yang tenang.
誤|salah|誤解|ごかい|kesalahpahaman|説明して誤解を解きました。|せつめいして ごかいを ときました。|Saya menjelaskan untuk meluruskan kesalahpahaman.
降|turun|降りる|おりる|turun kendaraan|次の駅で降ります。|つぎの えきで おります。|Saya turun di stasiun berikutnya.
勤|bekerja; tekun|勤務|きんむ|tugas kerja|今日は朝から勤務しています。|きょうは あさから きんむしています。|Hari ini saya bekerja sejak pagi.
刻|mengukir; waktu|時刻|じこく|waktu pada jam|出発の時刻を確認しました。|しゅっぱつの じこくを かくにんしました。|Saya memeriksa waktu keberangkatan.
遅|terlambat; lambat|遅れる|おくれる|terlambat|バスが十分遅れました。|バスが じゅっぷん おくれました。|Bus terlambat sepuluh menit.
破|merusak; sobek|破れる|やぶれる|sobek|袋が破れてしまいました。|ふくろが やぶれてしまいました。|Kantongnya terlanjur sobek.
互|saling|互い|たがい|satu sama lain|互いの意見を聞きました。|たがいの いけんを ききました。|Kami mendengarkan pendapat satu sama lain.
彼|dia laki-laki|彼|かれ|dia laki-laki|彼は私の同級生です。|かれは わたしの どうきゅうせいです。|Dia teman sekelas saya.
恥|malu|恥ずかしい|はずかしい|malu|名前を間違えて恥ずかしかったです。|なまえを まちがえて はずかしかったです。|Saya malu karena salah menyebut nama.
払|membayar; menyingkirkan|払う|はらう|membayar|受付で料金を払いました。|うけつけで りょうきんを はらいました。|Saya membayar biaya di meja penerimaan.
舞|menari|舞台|ぶたい|panggung|舞台で歌いました。|ぶたいで うたいました。|Saya bernyanyi di panggung.
頼|mengandalkan; meminta|頼む|たのむ|meminta|友達に手伝いを頼みました。|ともだちに てつだいを たのみました。|Saya meminta bantuan kepada teman.
戻|kembali; mengembalikan|戻る|もどる|kembali|忘れ物を取りに家に戻りました。|わすれものを とりに いえに もどりました。|Saya kembali ke rumah untuk mengambil barang tertinggal.
抱|memeluk; memegang dalam lengan|抱く|だく|memeluk; menggendong|母は赤ちゃんを抱いています。|ははは あかちゃんを だいています。|Ibu menggendong bayi.
緒|awal; bersama dalam gabungan|一緒|いっしょ|bersama|友達と一緒に勉強しました。|ともだちと いっしょに べんきょうしました。|Saya belajar bersama teman.
逃|melarikan diri|逃げる|にげる|melarikan diri|猫が庭から逃げました。|ねこが にわから にげました。|Kucing melarikan diri dari halaman.
御|awalan hormat|御礼|おれい|ucapan terima kasih, formal|先生に御礼の手紙を書きました。|せんせいに おれいの てがみを かきました。|Saya menulis surat terima kasih kepada guru.
吹|meniup|吹く|ふく|meniup; bertiup|強い風が吹いています。|つよい かぜが ふいています。|Angin kuat sedang bertiup.
到|tiba|到着|とうちゃく|kedatangan|電車は予定通りに到着しました。|でんしゃは よていどおりに とうちゃくしました。|Kereta tiba sesuai jadwal.
髪|rambut kepala|髪|かみ|rambut|昨日、髪を切りました。|きのう、かみを きりました。|Kemarin saya memotong rambut.
疲|lelah|疲れる|つかれる|lelah|長く歩いて疲れました。|ながく あるいて つかれました。|Saya lelah setelah berjalan lama.
歳|usia; tahun|二十歳|はたち|dua puluh tahun|妹は来月二十歳になります。|いもうとは らいげつ はたちに なります。|Adik perempuan saya berusia dua puluh tahun bulan depan.
偶|kebetulan; pasangan|偶然|ぐうぜん|kebetulan|駅で偶然友達に会いました。|えきで ぐうぜん ともだちに あいました。|Saya kebetulan bertemu teman di stasiun.
偉|hebat|偉い|えらい|hebat; terpuji|毎日続けているのは偉いです。|まいにち つづけているのは えらいです。|Terpuji bahwa Anda terus melakukannya setiap hari.
頂|puncak; menerima merendah|頂く|いただく|menerima atau makan secara merendah|先生から本を頂きました。|せんせいから ほんを いただきました。|Saya menerima buku dari guru dengan rasa hormat.
"""

# Themes are our teaching order, not categories defined by the dictionary.
THEMES = [
    ('Orang dan hubungan', '王他礼君客相共老夫妻婦娘祖彼御皆交両婚'),
    ('Tubuh dan perasaan', '耳声首頭顔命息歯感情念愛喜笑夢腹痛怒眠怖恐恥髪疲欲悲背抱優'),
    ('Alam dan lingkungan', '石光草葉雪馬猫庭流陽球晴寒熱煙然船港神景園冷'),
    ('Rumah dan keseharian', '内太引米形曲全具和宿暗箱面座洗静財宅掛居窓閉置寝靴杯吸酒暮掃薬'.replace('掃','')),
    ('Perjalanan dan gerakan', '市向直所泳乗速進落路鳴横登都遠飛倒渡抜捕突退迷訪途押浮越逃吹到戻舞遊打段消'),
    ('Belajar dan komunikasi', '才号回次活点科由未末記組絵数表部伝初例参完想望能術期式信単法係辞覚説告種達類報識昨晩昔忘調'),
    ('Pekerjaan dan kegiatan', '支助役投決受定努労働商育指追放取得続連忙徒機産議経済制務資件任責収職規割演師備給処精積寄供招勤刻払頼頂申守許'),
    ('Masyarakat dan aturan', '民対局戦政関留険危存原側席察警罪犯権否官平等害'),
    ('Pilihan dan penilaian', '欠付失必苦美負談要勝好成位利良便深合予反返様福殺選願疑折束残求約性格列変幸治果因常非加容容易若余難確在論判断増解際認費示観値差限与候構慣満程絶貧適更盗除幾散似富探迎雑賛込誤降遅破互緒歳偶偉当化最争亡敗困過状現呼違'),
]

NOTES = {
    '未': '未 berarti belum, misalnya 未来 (masa depan). Bedakan bentuknya dari 末 yang berarti akhir; garis atas 未 lebih pendek.',
    '末': '末 berarti akhir, misalnya 週末 (akhir pekan). Garis atas 末 lebih panjang daripada garis di bawah; bandingkan 未.',
    '寒': '寒い dipakai untuk dinginnya cuaca atau lingkungan. 冷たい untuk sesuatu yang terasa dingin saat disentuh.',
    '冷': '冷たい dibaca つめたい untuk dinginnya benda atau sentuhan. Cuaca dingin umumnya memakai 寒い.',
    '熱': '熱 (ねつ) dapat berarti panas atau demam. 熱い (あつい) untuk benda panas; 暑い untuk cuaca panas.',
    '太': '太い (ふとい) untuk ketebalan atau diameter, bukan semua ukuran besar. 大きい berarti besar secara umum.',
    '速': '速い menyatakan kecepatan gerak; 早い menyatakan waktu yang awal. Keduanya dibaca はやい.',
    '神': '神社 dibaca じんじゃ. Bacaan gabungan tidak selalu sama dengan bacaan kanji saat berdiri sendiri: 神 (かみ).',
    '歳': '二十歳 lazim dibaca はたち pada contoh ini. 一歳 (いっさい) dan 八歳 (はっさい) memakai bunyi っ.',
    '給': '給料 (きゅうりょう) memiliki vokal panjang pada kedua bagian. Jangan memendekkan bunyi menjadi きゅりょ.',
    '婦': '夫婦 dibaca ふうふ. 夫 sendiri dibaca おっと untuk suami sendiri, 妻 (つま) untuk istri sendiri.',
    '夫': '夫 (おっと) untuk suami sendiri. Suami orang lain lazim disebut ご主人 (ごしゅじん); pilih sapaan sesuai hubungan.',
    '妻': '妻 (つま) untuk istri sendiri. Istri orang lain lazim disebut 奥さん (おくさん).',
    '達': '友達 dibaca ともだち dengan perubahan た→だ. ～たち juga menandai kelompok, seperti 子供たち.',
    '約': '約 (やく) sebelum angka berarti kira-kira. Dalam 約束 (やくそく), maknanya berkaitan dengan janji.',
    '杯': '一杯 dibaca いっぱい. Bedakan 一杯 (satu gelas/cangkir) dengan いっぱい yang dapat berarti penuh menurut konteks.',
    '頂': '頂く dibaca いただく dalam bahasa merendah; sering ditulis いただく dengan kana. 頂上 (ちょうじょう) berarti puncak.',
    '御': '御 dalam awalan hormat dapat dibaca お atau ご menurut kata. 御礼 sering ditulis お礼.',
    '腹': '腹 (はら) berarti perut; お腹 (おなか) lebih umum dalam percakapan sopan. 腹が減る adalah ungkapan kasual merasa lapar.',
    '君': '君 (きみ) adalah sapaan akrab dan bisa terdengar merendahkan kepada orang yang tidak dekat. ～君 sebagai akhiran nama dibaca くん.',
    '老': '老人 (ろうじん) berarti orang lanjut usia; お年寄り (おとしより) sering terdengar lebih halus dalam percakapan.',
    '亡': '亡くなる (なくなる) adalah ungkapan lebih halus untuk meninggal. 無くなる dengan 無 berarti menjadi tidak ada.',
    '治': '治る (なおる) untuk kondisi yang sembuh; 直る (なおる) untuk benda atau kesalahan yang diperbaiki.',
    '直': '直す (なおす) untuk memperbaiki benda atau kesalahan. 治す (なおす) untuk mengobati kondisi kesehatan.',
    '努': '努力 dibaca どりょく. Bedakan 努める (つとめる: berusaha) dan 勤める (つとめる: bekerja untuk suatu tempat).',
    '勤': '勤務 (きんむ) berarti pekerjaan/tugas kerja. 勤める (つとめる) untuk bekerja pada instansi; 努める untuk berusaha.',
    '折': '折る (おる) bisa melipat kertas atau mematahkan benda. 折れる (おれる) adalah bentuk keadaan patah.',
    '降': '降りる (おりる) untuk turun kendaraan; 降る (ふる) untuk hujan atau salju yang turun.',
    '申': '申し込む (もうしこむ) berarti mendaftar. 申す (もうす) juga dipakai sebagai bentuk merendah dari 言う.',
    '込': '～込む muncul pada kata kerja gabungan. Pada 申し込む artinya mendaftar; makna tiap gabungan perlu dipelajari sebagai kata.',
    '苦': '苦手 dibaca にがて (kurang mahir/tidak suka); 苦しい dibaca くるしい dan 苦い dibaca にがい.',
    '美': '美しい dibaca うつくしい. 美術 (びじゅつ) memakai on’yomi ビ.',
    '幾': '幾つ (いくつ) menanyakan jumlah atau usia, sering ditulis いくつ. Untuk usia lebih sopan gunakan おいくつですか。',
    '際': '実際 dibaca じっさい dengan bunyi っ. ～際（さい）juga berarti pada kesempatan ketika sesuatu terjadi.',
    '当': '当たる berarti mengenai atau tepat/menang menurut konteks. 当てる adalah tindakan membuat sesuatu mengenai sasaran.',
    '乗': 'Naik kendaraan memakai kendaraan＋に乗る, bukan を乗る. 乗り換える (のりかえる) berarti berganti kendaraan.',
    '返': '返す (かえす) berarti mengembalikan sesuatu. 帰る (かえる) berarti pulang; 戻る (もどる) berarti kembali.',
    '戻': '戻る (もどる) berarti kembali; 戻す (もどす) mengembalikan sesuatu ke keadaan/tempat sebelumnya.',
    '非': '非常に (ひじょうに) berarti sangat. 非常 pada 非常口 (ひじょうぐち) mengacu keadaan darurat, yakni pintu darurat.',
}

COMMON_KUN = {
    '化':'ば.ける、ば.かす', '全':'すべ.て、まった.く', '由':'よ.る',
    '支':'ささ.える', '定':'さだ.める、さだ.まる', '実':'み、みの.る',
    '美':'うつく.しい', '要':'い.る、かなめ', '成':'な.る、な.す',
    '位':'くらい', '利':'き.く', '労':'いたわ.る、ねぎら.う',
    '育':'そだ.つ、そだ.てる', '神':'かみ', '様':'さま',
    '望':'のぞ.む', '係':'かかり、かか.わる', '辞':'や.める',
    '留':'と.まる、と.める', '常':'つね', '種':'たね', '許':'ゆる.す',
    '得':'え.る、う.る', '経':'へ.る', '済':'す.む、す.ます',
    '際':'きわ', '認':'みと.める', '費':'つい.やす', '観':'み.る',
    '割':'わ.る、わり', '備':'そな.える、そな.わる',
    '限':'かぎ.る、かぎ.り', '与':'あた.える', '居':'い.る',
    '寄':'よ.る、よ.せる', '更':'さら.に、ふ.ける',
    '幾':'いく.つ、いく.ら', '似':'に.る', '互':'たが.い',
    '御':'お', '歳':'とし', '頂':'いただ.く、いただき',
    '君':'きみ', '産':'う.む、う.まれる', '顔':'かお',
}
NO_COMMON_KUN = set('才王科対局役部陽期福然念性格能術列式信単法識非官察容財徒機議論判制資権件状職規師供候途給処婦精婚適祖賛偶両号点番報')

def katakana(value):
    return ''.join(chr(ord(c)+0x60) if '\u3041' <= c <= '\u3096' else c for c in value)

def chunks(chars):
    for count in range((len(chars)+4)//5, len(chars)//4+1):
        fives = len(chars)-4*count
        if 0 <= fives <= count:
            sizes = [5]*fives + [4]*(count-fives)
            result, start = [], 0
            for size in sizes:
                result.append(chars[start:start+size]); start += size
            return result
    raise ValueError(f'Theme cannot be partitioned: {len(chars)}')

def build():
    reference = json.loads((ROOT/'docs/kanji-n3-reference.json').read_text(encoding='utf-8'))
    dictionary = reference['characters']
    rows = {}
    for line in CARDS.strip().splitlines():
        c, meaning, word, reading, word_meaning, jp, kana, translation = line.split('|')
        if c in rows or c not in dictionary or c not in word or c not in jp:
            raise ValueError('Invalid authored card: '+c)
        info = dictionary[c]
        kun = [r for r in info['readings_kun'] if not r.startswith('-')][:2]
        on = [r for r in info['readings_on'] if not r.startswith('-')][:2]
        note = NOTES.get(c, f'{word} dibaca {reading}, berarti {word_meaning}. Pelajari bacaan dalam kata, bukan dengan menyambung semua bacaan kanji.')
        rows[c] = dict(character=c, meaning=meaning, strokes=info['strokes'],
            onReading='、'.join(map(katakana,on)) or '—',
            kunReading=COMMON_KUN.get(c, '—' if c in NO_COMMON_KUN else ('、'.join(kun) or '—')),
            mnemonic=f'Hubungkan {c} dengan {word} ({word_meaning}). Bayangkan adegan: {translation}',
            note=note+' Titik pada kun’yomi memisahkan okurigana. Bacaan kartu adalah pilihan, bukan seluruh bacaan kamus; — berarti bacaan itu tidak diajarkan pada kartu ini.',
            examples=[dict(japanese=word, reading=reading, meaning=word_meaning),
                      dict(japanese=jp, reading=kana, meaning=translation)])
    missing = set(dictionary)-set(rows)
    if missing:
        raise ValueError('Missing cards: '+''.join(sorted(missing)))
    ordered, seen = [], set()
    themes = [(theme, chars + {'Belajar dan komunikasi':'番', 'Pekerjaan dan kegiatan':'配', 'Pilihan dan penilaian':'実'}.get(theme,'')) for theme,chars in THEMES]
    for theme, chars in themes:
        selected = list(dict.fromkeys(c for c in chars if c in rows and c not in seen))
        seen.update(selected)
        for number, group in enumerate(chunks(selected), 1):
            ordered.append((theme, f'{theme} {number}', group))
    remaining = [c for c in rows if c not in seen]
    if remaining:
        for number, group in enumerate(chunks(remaining),1):
            ordered.append(('Pemakaian tambahan', f'Pemakaian tambahan {number}', group))
    all_chars = [c for _,_,chars in ordered for c in chars]
    if len(all_chars) != len(set(all_chars)) or set(all_chars) != set(dictionary):
        raise ValueError('Curriculum must cover every reference character exactly once')
    lessons = []
    all_meanings = list(dict.fromkeys(r['meaning'] for r in rows.values()))
    all_readings = list(dict.fromkeys(r['examples'][0]['reading'] for r in rows.values()))
    for index,(theme,title,chars) in enumerate(ordered,1):
        cards = [rows[c] for c in chars]
        questions = []
        for i,card in enumerate(cards):
            preferred = [r['meaning'] for r in cards if r['meaning'] != card['meaning']]
            wrong = list(dict.fromkeys(preferred+[m for m in all_meanings if m != card['meaning']]))[:3]
            choices = [card['meaning']]+wrong
            correct = (index+i)%4; choices.insert(correct,choices.pop(0))
            questions.append(dict(Prompt=f'Apa arti utama kanji {card["character"]} pada kartu ini?', Choices=choices,Correct=correct,
                Explanation=f'{card["character"]} berarti {card["meaning"]}. Contoh: {card["examples"][0]["japanese"]}（{card["examples"][0]["reading"]}）berarti {card["examples"][0]["meaning"]}.'))
        for i,card in enumerate(cards):
            ex=card['examples'][0]
            preferred = [r['examples'][0]['reading'] for r in cards if r['examples'][0]['reading'] != ex['reading']]
            wrong = list(dict.fromkeys(preferred+[r for r in all_readings if r != ex['reading']]))[:3]
            choices = [ex['reading']]+wrong
            correct=(index+i+1)%4; choices.insert(correct,choices.pop(0))
            questions.append(dict(Prompt=f'Bagaimana membaca「{ex["japanese"]}」?',Choices=choices,Correct=correct,
                Explanation=f'{ex["japanese"]} dibaca {ex["reading"]}, berarti {ex["meaning"]}. '+card['note']))
        lessons.append(dict(ID=f'kanji-n3-{index:02d}',Chapter=theme,Title=f'{title} · {"".join(chars)}',Characters=cards,Questions=questions))
    (ROOT/'backend/data/kanji/n3.json').write_text(json.dumps(lessons,ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    table=['| Lesson | Tema | Kanji |','|---|---|---|']+[f'| {i:02d} | {l["Title"]} | {" ".join(c["character"] for c in l["Characters"])} |' for i,l in enumerate(lessons,1)]
    (ROOT/'build/kanji-n3-curriculum.md').write_text('\n'.join(table)+'\n',encoding='utf-8')
    print(f'{len(lessons)} lessons, {len(rows)} kanji, {len(rows)} words, {len(rows)} sentences, {len(rows)*2} questions')

if __name__ == '__main__':
    build()
