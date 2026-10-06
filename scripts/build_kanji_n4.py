"""Build the reviewed N4 curriculum from authored cards and a dictionary snapshot.

Run from any directory: python scripts/build_kanji_n4.py
Dictionary attribution and update instructions: docs/kanji-n4.md.
"""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
GROUPS = [
    ('Keluarga', 'Orang tua dan saudara', '親族兄弟姉'),
    ('Keluarga', 'Diri dan peran', '妹私自主代'),
    ('Tubuh', 'Tubuh dan indra', '体心手口目'),
    ('Gerak', 'Berdiri dan berjalan', '立足歩走止'),
    ('Gerak', 'Bergerak dan tiba', '動起発着帰'),
    ('Transportasi', 'Perjalanan dan kendaraan', '転運通道駅'),
    ('Tempat', 'Rumah dan bangunan', '家屋室堂館'),
    ('Tempat', 'Kota dan lingkungan', '町地場建京'),
    ('Masyarakat', 'Dunia dan organisasi', '世界公社員'),
    ('Pekerjaan', 'Pekerjaan dan produksi', '仕事業工作'),
    ('Belajar', 'Belajar dan penelitian', '教習勉研究'),
    ('Belajar', 'Ujian dan pertanyaan', '試験答問質'),
    ('Bahasa', 'Huruf dan bahasa', '漢字文英言'),
    ('Seni', 'Musik dan gambar', '音歌楽画映'),
    ('Makanan', 'Makanan dan minuman', '茶飯肉牛飲'),
    ('Alam', 'Hewan dan tanaman', '鳥魚犬花野'),
    ('Alam', 'Laut dan langit', '海風空夕夜'),
    ('Waktu', 'Musim dan pagi', '春夏秋冬朝'),
    ('Waktu', 'Hari dan waktu', '昼週曜早急'),
    ('Pikiran', 'Berpikir dan memahami', '考思知理計'),
    ('Pikiran', 'Makna dan keadaan', '意味有不悪'),
    ('Sifat', 'Ketepatan dan perbedaan', '正真同別特'),
    ('Sifat', 'Jumlah dan ukuran', '少多度重力'),
    ('Warna', 'Warna dan terang', '色赤青黒明'),
    ('Kesehatan', 'Kesehatan dan keselamatan', '病医院死安'),
    ('Pemakaian sehari-hari', 'Kekuatan dan cara', '強元以方用'),
    ('Belanja', 'Jual beli dan biaya', '買売店品料'),
    ('Belanja', 'Memegang dan meminjam', '持待借貸使'),
    ('Benda', 'Kertas, pakaian, dan barang', '切写紙服物'),
    ('Tempat', 'Tempat tinggal dan jarak', '住広近田洋'),
    ('Perubahan', 'Lama, baru, mulai, dan selesai', '古新始終'),
    ('Kegiatan', 'Pertemuan dan perjalanan', '会集旅送'),
    ('Informasi', 'Gambar dan pembahasan', '台図注題'),
    ('Pemakaian sehari-hari', 'Pergi, membuka, uang, dan orang', '去開銀者'),
]

# Character | Indonesian meaning | memory cue | word | reading | meaning | word | reading | meaning
# Memory cues are invented learning aids, not historical etymologies.
CARDS = """
親|orang tua; dekat|Bayangkan orang tua mengawasi anak dari dekat.|親|おや|orang tua|親切|しんせつ|baik hati
族|keluarga; kelompok|Bayangkan satu keluarga berkumpul di bawah bendera.|家族|かぞく|keluarga|民族|みんぞく|kelompok etnis
兄|kakak laki-laki|Bayangkan kakak berdiri dengan mulut besar untuk memanggil adiknya.|兄|あに|kakak laki-laki sendiri|兄弟|きょうだい|saudara kandung
弟|adik laki-laki|Bayangkan adik membawa busur mainan yang bertali panjang.|弟|おとうと|adik laki-laki sendiri|弟子|でし|murid; pengikut
姉|kakak perempuan|Ingat 女 di kiri sebagai petunjuk saudara perempuan.|姉|あね|kakak perempuan sendiri|姉妹|しまい|saudara perempuan
妹|adik perempuan|Bayangkan adik perempuan yang masa depannya belum terbuka.|妹|いもうと|adik perempuan sendiri|姉妹|しまい|saudara perempuan
私|saya; pribadi|Bayangkan hasil panen yang disimpan untuk diri sendiri.|私|わたし|saya|私立|しりつ|swasta, untuk lembaga pendidikan
自|diri sendiri|Bayangkan menunjuk hidung sendiri saat berkata saya.|自分|じぶん|diri sendiri|自転車|じてんしゃ|sepeda
主|utama; pemilik|Ingat satu tanda di puncak sebagai petunjuk orang utama.|主人|しゅじん|suami; pemilik|主に|おもに|terutama
代|menggantikan; generasi|Bayangkan seseorang mengambil giliran menggantikan orang lain.|時代|じだい|zaman|代わり|かわり|pengganti
体|tubuh|Bayangkan tubuh manusia berdiri di samping pohon.|体|からだ|tubuh|体育|たいいく|pendidikan jasmani
心|hati; pikiran|Bayangkan denyut hati yang menyebar menjadi tiga titik.|心|こころ|hati; batin|安心|あんしん|rasa tenang
手|tangan|Bayangkan jari-jari pada bentuk garis yang memanjang.|手|て|tangan|手紙|てがみ|surat
口|mulut; bukaan|Kotak kecil menyerupai mulut yang terbuka.|入口|いりぐち|pintu masuk|出口|でぐち|pintu keluar
目|mata|Bayangkan kotak dengan garis di dalam sebagai mata.|目的|もくてき|tujuan|一つ目|ひとつめ|yang pertama
立|berdiri|Bayangkan orang berdiri tegak di atas tanah.|立つ|たつ|berdiri|国立|こくりつ|milik negara, untuk lembaga
足|kaki; mencukupi|Bayangkan kaki yang menopang bagian tubuh di atasnya.|足|あし|kaki|足りる|たりる|cukup
歩|berjalan|Bayangkan langkah-langkah kecil yang bergerak maju.|歩く|あるく|berjalan|散歩|さんぽ|jalan-jalan
走|berlari|Bayangkan kaki di bagian bawah yang bergerak cepat.|走る|はしる|berlari|走行|そうこう|pergerakan kendaraan
止|berhenti|Bayangkan satu kaki menahan gerakan untuk berhenti.|止まる|とまる|berhenti|中止|ちゅうし|pembatalan; penghentian
動|bergerak|Ingat 力 sebagai tenaga untuk menggerakkan sesuatu yang berat.|動く|うごく|bergerak|運動|うんどう|olahraga; gerakan
起|bangkit; terjadi|Bayangkan seseorang bangkit lalu mulai berjalan.|起きる|おきる|bangun; terjadi|起こす|おこす|membangunkan; menyebabkan
発|berangkat; memunculkan|Bayangkan tenaga dilepaskan saat perjalanan dimulai.|出発|しゅっぱつ|keberangkatan|発音|はつおん|pelafalan
着|tiba; mengenakan|Bayangkan mata memeriksa baju yang sudah melekat di tubuh.|着る|きる|mengenakan, pada tubuh bagian atas|到着|とうちゃく|kedatangan
帰|pulang|Bayangkan membawa sapu kembali ke rumah setelah selesai bekerja.|帰る|かえる|pulang|帰国|きこく|pulang ke negara asal
転|berputar; berpindah|Ingat 車 sebagai roda yang berputar.|運転|うんてん|mengemudi|自転車|じてんしゃ|sepeda
運|mengangkut; nasib|Bayangkan kendaraan bergerak membawa barang di jalan.|運ぶ|はこぶ|mengangkut|運転|うんてん|mengemudi
通|melewati; berkomunikasi|Bayangkan jalan yang menghubungkan dua tempat.|通る|とおる|melewati|交通|こうつう|lalu lintas; transportasi
道|jalan|Bayangkan kepala mengarah ke jalan yang akan dilalui.|道|みち|jalan|道路|どうろ|jalan raya
駅|stasiun|Bayangkan tempat kuda berhenti yang kini menjadi stasiun.|駅|えき|stasiun|駅員|えきいん|petugas stasiun
家|rumah; keluarga|Bayangkan tempat berlindung dengan atap besar.|家|いえ|rumah|家族|かぞく|keluarga
屋|bangunan; toko|Bayangkan atap menaungi tempat sebuah usaha.|本屋|ほんや|toko buku|部屋|へや|kamar
室|ruangan|Bayangkan tempat di bawah atap yang dibagi menjadi ruangan.|教室|きょうしつ|ruang kelas|室内|しつない|di dalam ruangan
堂|aula; bangunan besar|Bayangkan aula besar tempat banyak orang berkumpul.|食堂|しょくどう|kantin; ruang makan|講堂|こうどう|aula
館|gedung; fasilitas|Bayangkan gedung dengan fasilitas untuk pengunjung.|図書館|としょかん|perpustakaan|旅館|りょかん|penginapan bergaya Jepang
町|kota kecil; wilayah|Bayangkan petak sawah di samping jalan sebuah kota kecil.|町|まち|kota kecil; lingkungan|町長|ちょうちょう|kepala pemerintahan kota kecil
地|tanah; tempat|Ingat 土 di kiri untuk tanah suatu tempat.|地下|ちか|bawah tanah|地図|ちず|peta
場|tempat; lokasi|Ingat 土 sebagai tanah tempat suatu kegiatan berlangsung.|場所|ばしょ|tempat|工場|こうじょう|pabrik
建|membangun|Bayangkan tangan menyusun bangunan langkah demi langkah.|建てる|たてる|membangun|建物|たてもの|bangunan
京|ibu kota|Bayangkan bangunan tinggi di pusat ibu kota.|東京|とうきょう|Tokyo|京都|きょうと|Kyoto
世|dunia; generasi|Bayangkan beberapa generasi terhubung dalam satu garis waktu.|世界|せかい|dunia|世話|せわ|bantuan; perawatan
界|batas; lingkup|Bayangkan petak-petak sawah yang dibatasi garis.|世界|せかい|dunia|業界|ぎょうかい|bidang industri
公|publik; umum|Bayangkan tangan membuka sesuatu agar semua orang dapat memakainya.|公園|こうえん|taman umum|公共|こうきょう|publik; untuk umum
社|perusahaan; kuil Shinto|Bayangkan tempat orang berkumpul sebagai organisasi.|会社|かいしゃ|perusahaan|神社|じんじゃ|kuil Shinto
員|anggota; petugas|Bayangkan orang-orang tercatat sebagai anggota sebuah lingkaran.|会社員|かいしゃいん|pegawai perusahaan|店員|てんいん|pegawai toko
仕|melayani; bekerja|Bayangkan orang di kiri menjalankan tugasnya.|仕事|しごと|pekerjaan|仕方|しかた|cara melakukan
事|hal; urusan|Bayangkan banyak urusan yang tersusun pada satu batang.|仕事|しごと|pekerjaan|食事|しょくじ|makan; kegiatan makan
業|pekerjaan; usaha|Bayangkan banyak cabang pekerjaan dalam satu usaha.|授業|じゅぎょう|pelajaran di kelas|卒業|そつぎょう|kelulusan
工|kerajinan; industri|Bentuknya seperti alat kerja untuk membuat barang.|工場|こうじょう|pabrik|工事|こうじ|pekerjaan konstruksi
作|membuat|Bayangkan orang di kiri membuat sesuatu dengan alat.|作る|つくる|membuat|作文|さくぶん|karangan
教|mengajar|Bayangkan orang tua menunjukkan cara kepada anak.|教える|おしえる|mengajarkan; memberi tahu|教室|きょうしつ|ruang kelas
習|mempelajari; berlatih|Bayangkan sayap berlatih terbang berulang kali.|習う|ならう|belajar dari seseorang|練習|れんしゅう|latihan
勉|berusaha; belajar tekun|Ingat 力 sebagai tenaga untuk belajar dengan tekun.|勉強|べんきょう|belajar|勉強会|べんきょうかい|pertemuan kelompok belajar
研|mengasah; meneliti|Bayangkan batu di kiri dipakai untuk mengasah pengetahuan.|研究|けんきゅう|penelitian|研修|けんしゅう|pelatihan
究|menyelidiki sampai tuntas|Bayangkan mencari hingga ke dalam lubang untuk mengetahui jawabannya.|研究|けんきゅう|penelitian|究める|きわめる|menyelidiki sampai tuntas
試|mencoba; menguji|Bayangkan kata-kata di kiri dipakai dalam sebuah tes.|試験|しけん|ujian|試す|ためす|mencoba; menguji
験|pengalaman; pengujian|Bayangkan perjalanan dengan kuda yang memberi banyak pengalaman.|試験|しけん|ujian|経験|けいけん|pengalaman
答|jawaban; menjawab|Bayangkan dua bagian bertemu untuk menghasilkan jawaban.|答える|こたえる|menjawab|答え|こたえ|jawaban
問|bertanya; persoalan|Bayangkan mulut di antara pintu sedang bertanya.|質問|しつもん|pertanyaan|問題|もんだい|masalah; soal
質|mutu; sifat|Bayangkan barang berharga diperiksa untuk mengetahui mutunya.|質問|しつもん|pertanyaan|品質|ひんしつ|mutu barang
漢|Tiongkok; aksara Han|Bayangkan karakter yang menghubungkan sejarah bahasa dengan Tiongkok.|漢字|かんじ|kanji|漢語|かんご|kata Jepang asal Tionghoa
字|huruf; karakter|Bayangkan anak di bawah atap sedang belajar huruf.|漢字|かんじ|kanji|文字|もじ|huruf; tulisan
文|tulisan; kalimat|Bayangkan garis-garis silang menjadi pola tulisan.|作文|さくぶん|karangan|文化|ぶんか|budaya
英|Inggris; unggul|Bayangkan bunga yang menonjol sebagai petunjuk keunggulan.|英語|えいご|bahasa Inggris|英国|えいこく|Britania Raya
言|berkata; ucapan|Bayangkan garis-garis kata keluar dari mulut.|言う|いう|mengatakan|言葉|ことば|kata; bahasa
音|bunyi|Bayangkan suara di bawah bangunan yang sedang berbunyi.|音|おと|bunyi|音楽|おんがく|musik
歌|lagu; menyanyi|Bayangkan mulut yang mengulang nada saat bernyanyi.|歌|うた|lagu|歌う|うたう|menyanyi
楽|menyenangkan; musik|Bayangkan irama yang membuat suasana menyenangkan.|楽しい|たのしい|menyenangkan|音楽|おんがく|musik
画|gambar; rencana|Bayangkan gambar dibingkai oleh garis di sekelilingnya.|映画|えいが|film|計画|けいかく|rencana
映|memantulkan; menayangkan|Bayangkan cahaya matahari di kiri memunculkan gambar di layar.|映画|えいが|film|映る|うつる|terpantul; tampil di layar
茶|teh|Bayangkan daun di atas manusia menjadi minuman teh.|お茶|おちゃ|teh|茶色|ちゃいろ|warna cokelat
飯|nasi; makanan|Ingat bagian makanan di kiri saat memikirkan nasi.|ご飯|ごはん|nasi; makanan|朝ご飯|あさごはん|sarapan
肉|daging|Bayangkan dua potongan di dalam kotak sebagai daging.|肉|にく|daging|牛肉|ぎゅうにく|daging sapi
牛|sapi|Bayangkan kepala sapi dengan tanduk di atasnya.|牛|うし|sapi|牛乳|ぎゅうにゅう|susu sapi
飲|minum|Bayangkan seseorang membuka mulut untuk minum.|飲む|のむ|minum|飲み物|のみもの|minuman
鳥|burung|Bayangkan bentuk kepala, sayap, dan kaki seekor burung.|鳥|とり|burung|小鳥|ことり|burung kecil
魚|ikan|Bayangkan kepala, badan berpetak, dan ekor ikan.|魚|さかな|ikan|魚屋|さかなや|toko ikan
犬|anjing|Bayangkan titik sebagai telinga pada gambar seekor anjing.|犬|いぬ|anjing|子犬|こいぬ|anak anjing
花|bunga|Ingat rumput di atas sebagai tanaman yang berbunga.|花|はな|bunga|花火|はなび|kembang api
野|padang; liar|Bayangkan tanah luas di luar permukiman.|野菜|やさい|sayuran|野原|のはら|padang
海|laut|Tiga titik air di kiri mengingatkan pada air laut.|海|うみ|laut|海外|かいがい|luar negeri
風|angin|Bayangkan bentuk luar memeluk udara yang berputar.|風|かぜ|angin|台風|たいふう|topan
空|langit; kosong|Bayangkan ruang terbuka luas di atas kepala.|空|そら|langit|空気|くうき|udara
夕|petang|Bayangkan bulan yang baru muncul saat petang.|夕方|ゆうがた|petang|夕食|ゆうしょく|makan malam
夜|malam|Bayangkan seseorang bergerak di bawah langit gelap.|夜|よる|malam|今夜|こんや|malam ini
春|musim semi|Bayangkan matahari muncul di antara tanaman baru.|春|はる|musim semi|春休み|はるやすみ|libur musim semi
夏|musim panas|Bayangkan orang berjalan di bawah panas matahari.|夏|なつ|musim panas|夏休み|なつやすみ|libur musim panas
秋|musim gugur|Bayangkan tanaman di kiri berubah warna seperti api.|秋|あき|musim gugur|秋分|しゅうぶん|ekuinoks musim gugur
冬|musim dingin|Bayangkan titik-titik di bawah sebagai es yang jatuh.|冬|ふゆ|musim dingin|冬休み|ふゆやすみ|libur musim dingin
朝|pagi|Bayangkan matahari dan bulan sama-sama terlihat saat pagi.|朝|あさ|pagi|毎朝|まいあさ|setiap pagi
昼|siang|Bayangkan matahari berada tinggi di tengah hari.|昼|ひる|siang|昼ご飯|ひるごはん|makan siang
週|minggu; pekan|Bayangkan putaran tujuh hari mengelilingi kalender.|今週|こんしゅう|minggu ini|毎週|まいしゅう|setiap minggu
曜|hari dalam pekan|Ingat 日 di kiri sebagai penanda hari.|曜日|ようび|hari dalam pekan|月曜日|げつようび|Senin
早|awal; cepat waktunya|Bayangkan matahari sudah muncul sejak awal.|早い|はやい|awal; cepat dalam hal waktu|早朝|そうちょう|pagi sekali
急|tergesa; mendesak|Bayangkan hati di bawah terasa gelisah karena terburu-buru.|急ぐ|いそぐ|bergegas|急行|きゅうこう|kereta ekspres
考|memikirkan|Bayangkan orang tua menunduk untuk berpikir.|考える|かんがえる|memikirkan|考え|かんがえ|pemikiran; pendapat
思|berpikir; merasa|Bayangkan pikiran dari kepala turun ke hati.|思う|おもう|berpikir; merasa|思い出|おもいで|kenangan
知|mengetahui|Bayangkan mulut mengungkapkan sesuatu yang telah diketahui.|知る|しる|mengetahui|知らない|しらない|tidak tahu
理|alasan; prinsip|Bayangkan susunan yang rapi menjelaskan sebuah prinsip.|理由|りゆう|alasan|料理|りょうり|masakan; memasak
計|menghitung; merencanakan|Bayangkan angka sepuluh di samping kata untuk menghitung.|時計|とけい|jam|計画|けいかく|rencana
意|niat; gagasan|Bayangkan bunyi di atas hati menjadi sebuah maksud.|意味|いみ|makna|意見|いけん|pendapat
味|rasa; makna|Bayangkan mulut di kiri sedang mencicipi rasa.|味|あじ|rasa|意味|いみ|makna
有|ada; memiliki|Bayangkan tangan memegang sesuatu yang dimiliki.|有名|ゆうめい|terkenal|有料|ゆうりょう|berbayar
不|tidak; negasi|Bayangkan garis yang menolak sebuah pernyataan.|不便|ふべん|tidak praktis|不足|ふそく|kekurangan
悪|buruk|Bayangkan hati di bawah sedang merasa tidak baik.|悪い|わるい|buruk|最悪|さいあく|paling buruk
正|benar; tepat|Bayangkan kaki berhenti tepat pada garis yang benar.|正しい|ただしい|benar|正月|しょうがつ|Tahun Baru Jepang
真|benar-benar; sejati|Bayangkan sesuatu berdiri lurus sebagai kebenaran.|写真|しゃしん|foto|真ん中|まんなか|tepat di tengah
同|sama|Bayangkan isi yang sama berada di dalam kotak.|同じ|おなじ|sama|同時|どうじ|pada saat yang sama
別|berbeda; berpisah|Bayangkan pisau di kanan memisahkan dua bagian.|別れる|わかれる|berpisah|特別|とくべつ|khusus; istimewa
特|khusus|Bayangkan sapi dengan tanda khusus yang membedakannya.|特別|とくべつ|khusus; istimewa|特に|とくに|terutama; khususnya
少|sedikit|Bayangkan hanya sedikit titik yang tersebar.|少ない|すくない|sedikit|少し|すこし|sedikit, sebagai keterangan
多|banyak|Bayangkan dua bentuk bertumpuk menjadi jumlah yang banyak.|多い|おおい|banyak|多分|たぶん|mungkin
度|tingkat; kali|Bayangkan ukuran yang diperiksa setiap kali.|一度|いちど|sekali|温度|おんど|suhu
重|berat; berlapis|Bayangkan banyak lapisan menumpuk hingga berat.|重い|おもい|berat|体重|たいじゅう|berat badan
力|tenaga; kemampuan|Bayangkan lengan menekuk saat mengeluarkan tenaga.|力|ちから|tenaga|体力|たいりょく|kekuatan fisik
色|warna|Bayangkan cat yang dituangkan untuk memberi warna.|色|いろ|warna|茶色|ちゃいろ|cokelat
赤|merah|Bayangkan tanah yang menyala merah karena panas.|赤い|あかい|merah|赤ちゃん|あかちゃん|bayi
青|biru; hijau tertentu|Bayangkan langit cerah di atas permukaan air.|青い|あおい|biru|青年|せいねん|pemuda
黒|hitam|Bayangkan api di bawah meninggalkan warna hitam.|黒い|くろい|hitam|黒板|こくばん|papan tulis
明|terang; jelas|Matahari dan bulan bersama memberi cahaya terang.|明るい|あかるい|terang; ceria|説明|せつめい|penjelasan
病|sakit; penyakit|Bayangkan seseorang berbaring di bawah atap karena sakit.|病気|びょうき|sakit; penyakit|病院|びょういん|rumah sakit
医|pengobatan; dokter|Bayangkan alat yang disimpan dokter dalam kotak.|医者|いしゃ|dokter|医学|いがく|ilmu kedokteran
院|lembaga; bangunan besar|Bayangkan pagar di kiri mengelilingi sebuah lembaga.|病院|びょういん|rumah sakit|大学院|だいがくいん|sekolah pascasarjana
死|mati; kematian|Bayangkan seseorang yang tidak lagi dapat berdiri.|死ぬ|しぬ|mati|死亡|しぼう|kematian
安|aman; murah|Bayangkan orang dapat duduk tenang di bawah atap.|安い|やすい|murah|安全|あんぜん|aman; keselamatan
強|kuat|Bayangkan busur yang ditarik dengan tenaga besar.|強い|つよい|kuat|勉強|べんきょう|belajar
元|asal; awal|Bayangkan seseorang berdiri di titik awal.|元気|げんき|sehat; bersemangat|地元|じもと|daerah setempat
以|dengan; sejak; batas perbandingan|Bayangkan garis yang menjadi patokan perbandingan.|以上|いじょう|sekurang-kurangnya; lebih dari atau sama dengan|以下|いか|paling banyak; kurang dari atau sama dengan
方|arah; cara|Bayangkan penunjuk yang mengarah ke satu sisi.|方|かた|orang, bentuk sopan|読み方|よみかた|cara membaca
用|menggunakan; keperluan|Bayangkan kotak alat yang siap digunakan.|用事|ようじ|urusan|利用|りよう|penggunaan
買|membeli|Bayangkan mata melihat uang sebelum membeli barang.|買う|かう|membeli|買い物|かいもの|belanja
売|menjual|Bayangkan pedagang berdiri di bawah atap lapak.|売る|うる|menjual|売り場|うりば|area penjualan
店|toko|Bayangkan atap melindungi tempat pedagang membuka toko.|店|みせ|toko|店員|てんいん|pegawai toko
品|barang; mutu|Tiga kotak mengingatkan pada barang yang ditata.|品物|しなもの|barang|商品|しょうひん|barang dagangan
料|bahan; biaya|Bayangkan beras diukur untuk menentukan bahan dan biaya.|料理|りょうり|masakan; memasak|料金|りょうきん|tarif; biaya
持|memegang; memiliki|Bayangkan tangan di kiri memegang sesuatu dengan mantap.|持つ|もつ|memegang; memiliki|気持ち|きもち|perasaan
待|menunggu|Bayangkan orang berhenti di jalan sambil menunggu.|待つ|まつ|menunggu|期待|きたい|harapan
借|meminjam|Bayangkan orang di kiri memakai barang untuk sementara.|借りる|かりる|meminjam|借金|しゃっきん|utang
貸|meminjamkan|Bayangkan barang berharga diberikan untuk dipakai sementara.|貸す|かす|meminjamkan|貸し出し|かしだし|peminjaman keluar
使|menggunakan; menyuruh|Bayangkan orang di kiri menggunakan suatu alat.|使う|つかう|menggunakan|大使館|たいしかん|kedutaan besar
切|memotong|Ingat pisau di kanan untuk memotong sesuatu.|切る|きる|memotong|切手|きって|prangko
写|menyalin; memotret|Bayangkan gambar disalin di bawah sebuah penutup.|写真|しゃしん|foto|写す|うつす|menyalin; memotret
紙|kertas|Bayangkan serat di kiri menjadi lembaran kertas.|紙|かみ|kertas|手紙|てがみ|surat
服|pakaian|Bayangkan pakaian melekat pada bagian tubuh.|服|ふく|pakaian|洋服|ようふく|pakaian bergaya Barat
物|benda|Bayangkan sesuatu yang dapat dilihat dan dipegang sebagai benda.|物|もの|benda|食べ物|たべもの|makanan
住|tinggal|Bayangkan orang di kiri menetap di satu tempat.|住む|すむ|tinggal|住所|じゅうしょ|alamat
広|luas|Bayangkan atap yang melebar menaungi ruang luas.|広い|ひろい|luas|広場|ひろば|lapangan terbuka
近|dekat|Bayangkan beberapa langkah di jalan menuju tempat dekat.|近い|ちかい|dekat|近所|きんじょ|lingkungan sekitar
田|sawah|Kotak berpetak menggambarkan sawah yang dibagi.|田んぼ|たんぼ|sawah|水田|すいでん|sawah berair
洋|samudra; Barat|Bayangkan air luas di kiri membentang ke samudra.|洋服|ようふく|pakaian bergaya Barat|西洋|せいよう|dunia Barat
古|lama; tua|Bayangkan cerita dari mulut yang sudah disampaikan sejak lama.|古い|ふるい|lama; tua|中古|ちゅうこ|bekas, untuk barang
新|baru|Bayangkan kayu dipotong untuk membuat barang baru.|新しい|あたらしい|baru|新聞|しんぶん|surat kabar
始|mulai|Bayangkan seseorang siap memulai kegiatan di panggung.|始まる|はじまる|mulai, untuk peristiwa|始める|はじめる|memulai sesuatu
終|selesai; akhir|Bayangkan benang di kiri mencapai ujungnya.|終わる|おわる|selesai|終電|しゅうでん|kereta terakhir
会|bertemu; perkumpulan|Bayangkan orang-orang bertemu di bawah satu atap.|会う|あう|bertemu|会議|かいぎ|rapat
集|berkumpul; mengumpulkan|Bayangkan burung-burung berkumpul di satu pohon.|集まる|あつまる|berkumpul|集める|あつめる|mengumpulkan
旅|perjalanan|Bayangkan bendera menunjukkan arah rombongan perjalanan.|旅行|りょこう|perjalanan; wisata|旅館|りょかん|penginapan bergaya Jepang
送|mengirim; mengantar|Bayangkan langkah membawa barang menuju penerimanya.|送る|おくる|mengirim; mengantar|放送|ほうそう|siaran
台|panggung; alas|Bayangkan mulut berbicara dari atas panggung.|台風|たいふう|topan|一台|いちだい|satu mesin atau kendaraan
図|gambar; diagram|Bayangkan sebuah diagram berada di dalam bingkai.|地図|ちず|peta|図書館|としょかん|perpustakaan
注|menuangkan; memperhatikan|Bayangkan air dituangkan dengan penuh perhatian.|注意|ちゅうい|perhatian; peringatan|注文|ちゅうもん|pesanan
題|topik; soal|Bayangkan kepala memikirkan satu topik pembahasan.|問題|もんだい|masalah; soal|宿題|しゅくだい|pekerjaan rumah
去|pergi; berlalu|Bayangkan meninggalkan tanah tempat semula berdiri.|去年|きょねん|tahun lalu|過去|かこ|masa lalu
開|membuka|Bayangkan dua daun pintu di kiri dan kanan terbuka.|開ける|あける|membuka sesuatu|開く|あく|terbuka
銀|perak|Ingat 金 di kiri sebagai petunjuk logam perak.|銀行|ぎんこう|bank|銀色|ぎんいろ|warna perak
者|orang; pelaku|Bayangkan orang yang dikenal melalui pekerjaan atau tindakannya.|医者|いしゃ|dokter|学者|がくしゃ|ilmuwan; akademisi
"""

SPECIAL_NOTES = {
    '親': '親 dapat berarti orang tua; 親切 dibaca しんせつ dan berarti baik hati.',
    '兄': 'あに untuk kakak laki-laki sendiri. 兄弟 dibaca きょうだい; jangan membaca tiap kanji secara terpisah.',
    '弟': '弟子 dibaca でし, bukan おとうとこ. Kata ini berarti murid atau pengikut.',
    '私': '私 dibaca わたし dalam percakapan biasa; わたくし lebih formal.',
    '体': '体 dibaca からだ saat berdiri sendiri; 体育 memakai bacaan on たい.',
    '目': 'Pada 一つ目, 目 dibaca め dan menandai urutan. 目的 memakai bacaan もく.',
    '発': '発音 dibaca はつおん; 出発 dibaca しゅっぱつ dengan bunyi っ.',
    '着': '着る adalah mengenakan pakaian bagian atas; 到着 memakai bacaan ちゃく.',
    '通': '通る（とおる）berarti melewati; 通う（かよう）berarti pergi secara rutin.',
    '場': '工場 dibaca こうじょう; 場所 memakai bacaan ば.',
    '社': '神社 dibaca じんじゃ. Bunyi しゃ berubah menjadi じゃ dalam kata ini.',
    '工': '工場 dibaca こうじょう, sedangkan 工事 dibaca こうじ.',
    '勉': '勉 biasanya dipelajari lewat gabungan 勉強. Tidak ada kun’yomi umum yang diajarkan di kartu ini.',
    '究': '研究 dibaca けんきゅう, dengan vokal panjang きゅう.',
    '質': '質問 dibaca しつもん; 質 di sini tidak dibaca ひん.',
    '漢': '漢字 dibaca かんじ. Tidak ada kun’yomi umum yang diajarkan di kartu ini.',
    '字': '文字 memiliki bacaan khusus もじ; pelajari kata ini sebagai satu kesatuan.',
    '楽': '音楽 memakai ガク. 楽しい memakai たの.しい; 楽（らく）berarti mudah atau tidak berat.',
    '画': '映画 memakai ガ. 計画 memakai カク. Keduanya contoh on’yomi yang berbeda.',
    '飯': 'ご飯 dibaca ごはん, sedangkan 朝ご飯 dibaca あさごはん.',
    '牛': '牛乳 dibaca ぎゅうにゅう, dengan dua vokal panjang.',
    '魚': '魚 umumnya dibaca さかな; ギョ dipakai dalam gabungan seperti 金魚（きんぎょ）.',
    '青': '青 biasanya biru, tetapi beberapa ungkapan memakai 青 untuk sesuatu yang tampak hijau.',
    '早': '早い dipakai untuk waktu yang awal; kecepatan gerak umumnya ditulis 速い.',
    '曜': '曜 sering muncul pada nama hari. Tidak ada kun’yomi umum yang diajarkan di kartu ini.',
    '思': '思い出（おもいで）berarti kenangan; bagian 出 dibaca で dalam kata ini.',
    '不': '不 adalah unsur negasi pada gabungan kata; 不便 dibaca ふべん.',
    '正': '正しい memakai ただ.しい, sedangkan 正月 dibaca しょうがつ.',
    '真': '真ん中 dibaca まんなか. 写真 memakai bacaan シン.',
    '同': '同じ adalah おなじ, sedangkan 同時 memakai ドウ.',
    '度': '一度 dibaca いちど; 温度 dibaca おんど. 度 dapat menandai kali atau tingkat.',
    '元': '元気 memakai ゲン, sedangkan 地元 memakai もと.',
    '以': '以上 dan 以下 mencakup batas yang disebut. 10以上 berarti 10 atau lebih; 10以下 berarti 10 atau kurang.',
    '方': '方（かた）dapat berarti orang secara sopan. Pada 読み方, artinya cara membaca.',
    '品': '品物 memakai しな, sedangkan 商品 memakai ヒン.',
    '借': '借りる berarti menerima pinjaman. 借金 dibaca しゃっきん dengan bunyi っ.',
    '貸': '貸す berarti memberikan pinjaman; bedakan dari 借りる yang menerima pinjaman.',
    '切': '切手 dibaca きって dengan bunyi っ; 切る dibaca きる.',
    '始': '始まる adalah peristiwa mulai; 始める berarti seseorang memulai sesuatu.',
    '集': '集まる berarti berkumpul; 集める berarti mengumpulkan orang atau benda.',
    '台': '台 adalah penghitung mesin atau kendaraan. 一台 dibaca いちだい.',
    '去': '去年 dibaca きょねん, sedangkan 過去 dibaca かこ.',
    '開': '開ける（あける）berarti membuka; 開く（あく）berarti terbuka. 開く（ひらく）juga ada, misalnya membuka buku atau mengadakan rapat.',
}

# Select modern, useful readings instead of exposing every historical dictionary form.
COMMON_ON = {'兄': 'ケイ、キョウ', '弟': 'テイ、ダイ、デ', '京': 'キョウ、ケイ', '地': 'チ、ジ',
             '世': 'セイ、セ', '工': 'コウ、ク', '楽': 'ガク、ラク',
             '画': 'ガ、カク', '不': 'フ、ブ', '正': 'セイ、ショウ',
             '図': 'ズ、ト', '台': 'ダイ、タイ', '去': 'キョ、コ'}
COMMON_KUN = {
    '親': 'おや、した.しい、した.しむ', '自': 'みずか.ら',
    '主': 'ぬし、おも、あるじ', '代': 'か.わる、か.える', '体': 'からだ',
    '手': 'て', '立': 'た.つ、た.てる', '足': 'あし、た.りる、た.す',
    '止': 'と.まる、と.める', '動': 'うご.く、うご.かす',
    '起': 'お.きる、お.こる、お.こす', '着': 'き.る、つ.く',
    '帰': 'かえ.る、かえ.す', '転': 'ころ.ぶ、ころ.がる、ころ.がす',
    '通': 'とお.る、とお.す、かよ.う', '道': 'みち', '家': 'いえ、や、うち',
    '建': 'た.てる、た.つ', '作': 'つく.る', '教': 'おし.える、おそ.わる',
    '試': 'ため.す、こころ.みる', '答': 'こた.える、こた.え', '問': 'と.う',
    '言': 'い.う、こと',
    '楽': 'たの.しい、たの.しむ', '映': 'うつ.る、うつ.す、は.える',
    '魚': 'さかな、うお', '犬': 'いぬ', '野': 'の', '風': 'かぜ',
    '空': 'そら、あ.く、から', '夜': 'よる、よ', '早': 'はや.い、はや.まる、はや.める',
    '急': 'いそ.ぐ', '思': 'おも.う', '知': 'し.る、し.らせる',
    '味': 'あじ、あじ.わう', '悪': 'わる.い', '正': 'ただ.しい、ただ.す',
    '真': 'ま、まこと', '少': 'すく.ない、すこ.し', '多': 'おお.い',
    '重': 'おも.い、かさ.ねる、かさ.なる', '赤': 'あか、あか.い',
    '青': 'あお、あお.い', '黒': 'くろ、くろ.い', '明': 'あか.るい、あ.ける、あ.かり',
    '病': 'やまい、や.む', '死': 'し.ぬ', '安': 'やす.い',
    '強': 'つよ.い、つよ.める、つよ.まる', '売': 'う.る、う.れる',
    '店': 'みせ', '持': 'も.つ', '貸': 'か.す', '使': 'つか.う',
    '切': 'き.る、き.れる', '写': 'うつ.す、うつ.る', '物': 'もの',
    '住': 'す.む', '広': 'ひろ.い、ひろ.がる、ひろ.げる',
    '古': 'ふる.い', '新': 'あたら.しい、あら.た',
    '始': 'はじ.まる、はじ.める', '終': 'お.わる、お.える',
    '会': 'あ.う', '集': 'あつ.まる、あつ.める',
    '図': 'はか.る', '注': 'そそ.ぐ', '開': 'あ.く、あ.ける、ひら.く',
}
NO_COMMON_KUN = {'勉', '漢', '曜', '員', '英', '料', '院', '発', '室',
                 '験', '質', '字', '画', '肉', '医', '以', '台', '館'}

def katakana(text):
    return ''.join(chr(ord(c) + 0x60) if '\u3041' <= c <= '\u3096' else c for c in text)

def build():
    reference = json.loads((ROOT / 'docs/kanji-n4-reference.json').read_text(encoding='utf-8'))
    dictionary = reference['characters']
    rows = {}
    for line in CARDS.strip().splitlines():
        parts = line.split('|')
        assert len(parts) == 9, line
        character, meaning, mnemonic, *example_data = parts
        assert character not in rows, character
        info = dictionary[character]
        on = info['readings_on'][:1]
        kun = [r for r in info['readings_kun'] if not r.startswith('-')][:1]
        if character in NO_COMMON_KUN:
            kun = []
        examples = [dict(zip(('japanese', 'reading', 'meaning'), example_data[i:i+3])) for i in (0, 3)]
        assert all(character in example['japanese'] for example in examples), character
        rows[character] = dict(character=character, meaning=meaning, strokes=info['strokes'],
            onReading=COMMON_ON.get(character, '、'.join(map(katakana, on)) or '—'),
            kunReading=COMMON_KUN.get(character, '、'.join(kun) or '—'),
            mnemonic=mnemonic, note=SPECIAL_NOTES.get(character,
                'Pelajari bacaan melalui contoh kata. Titik memisahkan bagian kanji dari okurigana; tanda — berarti tidak ada bacaan umum yang diajarkan. Bacaan pada kartu merupakan pilihan, bukan seluruh bacaan kamus.'),
            examples=examples)
    order = ''.join(group[2] for group in GROUPS)
    assert len(order) == len(set(order)) == len(rows) == 166
    assert set(order) == set(rows) == set(dictionary), 'N4 reference coverage is incomplete'
    lessons = []
    for index, (chapter, title, characters) in enumerate(GROUPS, 1):
        cards = [rows[c] for c in characters]
        questions = []
        for i, card in enumerate(cards):
            choices = [card['meaning']] + [other['meaning'] for other in cards if other != card][:3]
            assert len(set(choices)) == 4
            correct = (index + i) % 4
            choices.insert(correct, choices.pop(0))
            questions.append(dict(Prompt=f"Apa arti utama kanji {card['character']} pada kartu ini?",
                Choices=choices, Correct=correct,
                Explanation=f"{card['character']} berarti {card['meaning']}. Contoh: {card['examples'][0]['japanese']}（{card['examples'][0]['reading']}）berarti {card['examples'][0]['meaning']}."))
        for i, card in enumerate(cards):
            example = card['examples'][i % 2]
            distractors = []
            for other in cards:
                for candidate in other['examples']:
                    reading = candidate['reading']
                    if reading != example['reading'] and reading not in distractors:
                        distractors.append(reading)
            choices = [example['reading']] + distractors[:3]
            assert len(set(choices)) == 4
            correct = (index + i + 1) % 4
            choices.insert(correct, choices.pop(0))
            questions.append(dict(Prompt=f"Bagaimana membaca「{example['japanese']}」?", Choices=choices,
                Correct=correct, Explanation=f"{example['japanese']} dibaca {example['reading']}, berarti {example['meaning']}. " + card['note']))
        lessons.append(dict(ID=f'kanji-n4-{index:02d}', Chapter=chapter, Title=title,
                            Characters=cards, Questions=questions))
    target = ROOT / 'backend/data/kanji/n4.json'
    target.write_text(json.dumps(lessons, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    print(f'{len(lessons)} lessons, {len(rows)} characters, {len(rows)*2} examples and questions')

if __name__ == '__main__':
    build()
