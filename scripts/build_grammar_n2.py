"""Original Indonesian N2 curriculum. Run from the repository root."""
import json
from pathlib import Path

# Each row: pattern | meaning | formation | usage | common mistake |
# Japanese | kana reading | Indonesian | Japanese | kana | Indonesian.
CHAPTERS = [
('Kesempatan dan urutan waktu', 'Kesempatan formal', '''
～際に|pada saat atau kesempatan …|V辞書形／Vた／Nの + 際に|Dipakai dalam petunjuk dan pengumuman formal untuk waktu melakukan tindakan.|N memerlukan の; bukan Nに際に.|お帰りの際に、鍵を返してください。|おかえりの さいに、かぎを かえしてください。|Saat pulang, tolong kembalikan kunci.|申し込む際に、住所を確認してください。|もうしこむ さいに、じゅうしょを かくにんしてください。|Saat mendaftar, tolong periksa alamat.
～に際して|ketika menghadapi kesempatan penting …|N／V辞書形 + に際して|Menandai kesempatan penting seperti pembukaan, keberangkatan, atau kontrak; gaya formal.|Bukan untuk setiap kebiasaan kecil; N langsung diikuti に際して.|開会に際して、会長が挨拶しました。|かいかいに さいして、かいちょうが あいさつしました。|Pada pembukaan acara, ketua memberi sambutan.|契約するに際して、条件を確認しました。|けいやくするに さいして、じょうけんを かくにんしました。|Saat menandatangani kontrak, saya memeriksa persyaratannya.
'''),
('Kesempatan dan urutan waktu', 'Persiapan dan segera setelah', '''
～にあたって|dalam rangka memulai …|N／V辞書形 + にあたって|Persiapan atau sikap menghadapi kegiatan penting yang akan dimulai.|Tidak cocok untuk kejadian buruk spontan seperti kecelakaan.|留学するにあたって、保険に入りました。|りゅうがくするに あたって、ほけんに はいりました。|Dalam persiapan studi di luar negeri, saya mengambil asuransi.|新年にあたって、目標を立てました。|しんねんに あたって、もくひょうを たてました。|Menjelang tahun baru, saya menetapkan tujuan.
～次第|segera setelah …|Vます語幹 + 次第|Tindakan berikut dilakukan segera begitu persiapan selesai; sering untuk janji atau instruksi.|Gunakan pangkal ます: 分かり次第, bukan 分かる次第.|日程が決まり次第、ご連絡します。|にっていが きまりしだい、ごれんらくします。|Segera setelah jadwal ditentukan, saya akan menghubungi Anda.|資料が届き次第、確認してください。|しりょうが とどきしだい、かくにんしてください。|Segera setelah dokumen tiba, tolong periksa.
'''),
('Kesempatan dan urutan waktu', 'Kejadian mendadak', '''
～たとたん（に）|tepat setelah … tiba-tiba …|Vた + とたん（に）|Kejadian kedua muncul spontan tepat setelah kejadian pertama, sering mengejutkan.|Hasil kedua biasanya bukan rencana atau perintah pembicara.|窓を開けたとたん、強い風が入ってきました。|まどを あけたとたん、つよい かぜが はいってきました。|Begitu jendela dibuka, angin kencang masuk.|立ち上がったとたん、足が痛くなりました。|たちあがったとたん、あしが いたくなりました。|Begitu berdiri, kaki saya mendadak sakit.
～かと思うと／～かと思ったら|baru saja … langsung …|Vた + かと思うと／かと思ったら|Menggambarkan perubahan cepat yang diamati pembicara.|Biasanya pengamatan terhadap kejadian, bukan janji tindakan diri sendiri.|雨がやんだかと思うと、また降り始めました。|あめが やんだかと おもうと、また ふりはじめました。|Hujan baru saja reda lalu langsung turun lagi.|子供は帰ったかと思ったら、外へ走っていきました。|こどもは かえったかと おもったら、そとへ はしっていきました。|Anak itu baru saja pulang lalu langsung berlari keluar.
'''),
('Kesempatan dan urutan waktu', 'Dua kejadian nyaris bersamaan', '''
～か～ないかのうちに|baru saja … atau bahkan belum … sudah …|V辞書形 + か + Vない形 + かのうちに|Mengulang kata kerja untuk dua batas waktu yang sangat berdekatan.|Kata kerja harus sama: 終わるか終わらないか, bukan dua tindakan berbeda.|授業が終わるか終わらないかのうちに、学生が席を立ちました。|じゅぎょうが おわるか おわらないかの うちに、がくせいが せきを たちました。|Pelajaran nyaris belum selesai ketika mahasiswa sudah berdiri.|横になるかならないかのうちに、眠ってしまいました。|よこに なるか ならないかの うちに、ねむってしまいました。|Saya baru saja merebahkan diri dan langsung tertidur.
～ところに／～ところへ|tepat ketika … datang atau terjadi …|V辞書形／Vている／Vた + ところに／ところへ|Kejadian luar masuk pada tahap tindakan yang sedang dibicarakan.|Bedakan ところで sebagai penghubung pergantian topik.|出かけようとしているところに、電話が来ました。|でかけようと している ところに、でんわが きました。|Tepat ketika hendak keluar, telepon masuk.|夕食を作っているところへ、友達が来ました。|ゆうしょくを つくっている ところへ、ともだちが きました。|Saat sedang membuat makan malam, teman datang.
'''),
('Kesempatan dan urutan waktu', 'Sejak suatu titik dan prasyarat waktu', '''
～て以来|sejak … terus …|Vて + 以来; N + 以来|Keadaan berlangsung dari suatu titik masa lalu sampai sekarang.|Tidak dipakai untuk satu kejadian sesaat setelah tindakan pertama.|日本に来て以来、毎日日記を書いています。|にほんに きていらい、まいにち にっきを かいています。|Sejak datang ke Jepang, saya menulis jurnal setiap hari.|卒業以来、彼とは会っていません。|そつぎょういらい、かれとは あっていません。|Sejak lulus, saya belum bertemu dia.
～てからでないと／～てからでなければ|kalau belum … tidak bisa …|Vて + からでないと／からでなければ + hasil negatif|Tindakan pertama wajib selesai sebelum tindakan berikut dapat dilakukan.|Akhir klausa lazim negatif; bukan sekadar urutan bebas.|予約してからでないと、入れません。|よやくしてからでないと、はいれません。|Kalau belum memesan, Anda tidak boleh masuk.|実物を見てからでなければ、決められません。|じつぶつを みてからでなければ、きめられません。|Saya tidak bisa memutuskan sebelum melihat barangnya langsung.
'''),
('Proses dan hasil', 'Prosedur dan hasil pembahasan', '''
～た上で|setelah … dengan sengaja …|Vた／Nの + 上で|Tindakan awal merupakan persiapan yang perlu bagi keputusan berikut.|Untuk urutan sederhana cukup てから; 上で menekankan dasar persiapan.|説明を読んだ上で、署名してください。|せつめいを よんだうえで、しょめいしてください。|Setelah membaca penjelasan, silakan tanda tangan.|相談の上で、日程を決めます。|そうだんの うえで、にっていを きめます。|Setelah berkonsultasi, kami akan menentukan jadwal.
～末に|akhirnya setelah proses panjang …|Vた／Nの + 末に|Hasil tercapai setelah proses, usaha, atau kebimbangan yang cukup lama.|Bukan untuk setiap tindakan singkat; proses sebelumnya penting.|悩んだ末に、留学を決めました。|なやんだ すえに、りゅうがくを きめました。|Setelah lama bimbang, saya memutuskan studi ke luar negeri.|長い議論の末に、合意に達しました。|ながい ぎろんの すえに、ごういに たっしました。|Setelah diskusi panjang, kami mencapai kesepakatan.
'''),
('Proses dan hasil', 'Hasil buruk dan tindakan tanpa jeda', '''
～あげく（に）|setelah akhirnya … malah …|Vた／Nの + あげく（に）|Proses yang panjang berakhir dengan hasil buruk atau mengecewakan.|Berbeda dari 末に yang dapat berakhir baik; あげく biasanya bernada negatif.|迷ったあげく、何も買いませんでした。|まよった あげく、なにも かいませんでした。|Setelah lama bingung, akhirnya saya tidak membeli apa pun.|口論のあげく、二人は別れました。|こうろんの あげく、ふたりは わかれました。|Setelah bertengkar, keduanya akhirnya berpisah.
～つつ|sambil …|Vます語幹 + つつ|Dua tindakan dilakukan oleh pelaku yang sama; padanan formal dari ながら.|Jangan sambungkan ke bentuk kamus; 読みつつ benar.|資料を読みつつ、メモを取りました。|しりょうを よみつつ、メモを とりました。|Saya mencatat sambil membaca dokumen.|将来を考えつつ、仕事を探しています。|しょうらいを かんがえつつ、しごとを さがしています。|Saya mencari pekerjaan sambil memikirkan masa depan.
'''),
('Proses dan hasil', 'Perubahan yang terus berlangsung', '''
～つつある|sedang berangsur …|Vます語幹 + つつある|Perubahan bergerak menuju keadaan baru; gaya laporan atau tulisan.|Bukan pengganti semua Vている; gunakan untuk proses perubahan.|町の人口は減りつつあります。|まちの じんこうは へりつつあります。|Penduduk kota sedang berangsur berkurang.|環境への意識が高まりつつあります。|かんきょうへの いしきが たかまりつつあります。|Kesadaran terhadap lingkungan sedang meningkat.
～一方だ|terus semakin …|V辞書形 + 一方だ|Perubahan menuju satu arah terus terjadi, sering sesuatu yang mengkhawatirkan.|Tidak berarti dua sisi seperti 一方で; sebelum 一方だ gunakan bentuk kamus.|物価は上がる一方です。|ぶっかは あがる いっぽうです。|Harga barang terus meningkat.|仕事は増える一方です。|しごとは ふえる いっぽうです。|Pekerjaan terus bertambah.
'''),
('Sebab dan tanggung jawab', 'Alasan dengan penyesalan', '''
～ものだから|karena … sebagai penjelasan atau alasan|V普通形／Aい + ものだから; Aな／N + なものだから|Alasan subjektif, sering menjelaskan keadaan yang membuat pembicara gagal atau terlambat.|N memakai な: 初めてなものだから; tidak harus selalu meminta maaf.|道が混んでいたものだから、遅れてしまいました。|みちが こんでいた ものだから、おくれてしまいました。|Karena jalan macet, saya jadi terlambat.|初めてなものだから、緊張しました。|はじめてな ものだから、きんちょうしました。|Karena ini pertama kalinya, saya gugup.
～ばかりに|hanya karena … akibat buruk …|V普通形／Aい + ばかりに; Aな + なばかりに; N + であるばかりに|Satu sebab memicu akibat buruk yang disesali; fokus pada sebab tersebut.|Bukan makna hanya melakukan tindakan seperti Vてばかりいる.|一言多く言ったばかりに、友達を怒らせました。|ひとこと おおく いった ばかりに、ともだちを おこらせました。|Hanya karena mengatakan satu hal berlebihan, saya membuat teman marah.|安いばかりに、品質を確認せずに買いました。|やすい ばかりに、ひんしつを かくにんせずに かいました。|Hanya karena murah, saya membeli tanpa memeriksa kualitas.
'''),
('Sebab dan tanggung jawab', 'Dasar kesimpulan dan alasan khusus', '''
～ことから|dari fakta bahwa …|V普通形／Aい + ことから; Aな + なことから; N + であることから|Fakta menjadi dasar penamaan, penemuan, atau kesimpulan yang relatif objektif.|Tidak sama dengan ことだから yang menyimpulkan berdasarkan sifat orang.|形が星に似ていることから、星のパンと呼ばれています。|かたちが ほしに にている ことから、ほしの パンと よばれています。|Karena bentuknya menyerupai bintang, roti itu disebut roti bintang.|足跡があることから、誰かが入ったと分かりました。|あしあとが ある ことから、だれかが はいったと わかりました。|Dari adanya jejak kaki, diketahui seseorang telah masuk.
～からこそ|justru karena …|V普通形／Aい + からこそ; Aな／N + だからこそ|Menegaskan sebab yang dianggap paling penting atau justru sering diabaikan.|こそ mengikuti から, bukan mengganti から; nominal memerlukan だ.|大切な友達だからこそ、正直に話します。|たいせつな ともだちだからこそ、しょうじきに はなします。|Justru karena dia teman penting, saya berbicara jujur.|難しいからこそ、挑戦する価値があります。|むずかしいからこそ、ちょうせんする かちが あります。|Justru karena sulit, ada nilai dalam mencobanya.
'''),
('Sebab dan tanggung jawab', 'Tanggung jawab setelah keputusan', '''
～以上（は）|karena sudah … maka wajib atau wajar …|V普通形／Aい + 以上（は）; Aな／N + である以上（は）|Fakta yang diterima menimbulkan kewajiban, tekad, atau kesimpulan.|Bukan sekadar sebab netral; akibat biasanya tanggung jawab yang logis.|約束した以上、守らなければなりません。|やくそくした いじょう、まもらなければ なりません。|Karena sudah berjanji, saya harus menepatinya.|責任者である以上、説明する必要があります。|せきにんしゃである いじょう、せつめいする ひつようが あります。|Sebagai penanggung jawab, saya perlu memberi penjelasan.
～からには|karena sudah memilih atau berjanji … maka …|V普通形／Aい + からには; Aな／N + であるからには|Menegaskan tekad atau kewajiban yang mengikuti keputusan.|Klausa berikut jangan berupa kejadian kebetulan tanpa hubungan kewajiban.|参加するからには、最後まで頑張ります。|さんかするからには、さいごまで がんばります。|Karena sudah memutuskan ikut, saya akan berusaha sampai akhir.|引き受けたからには、責任を持ちます。|ひきうけたからには、せきにんを もちます。|Karena sudah menerima tugas, saya akan bertanggung jawab.
'''),
('Sebab dan tanggung jawab', 'Sebab yang membuat hasil terasa wajar', '''
～だけに|karena … terlebih lagi …|V普通形／Aい + だけに; Aな + なだけに; N + だけに|Hasil terasa lebih kuat atau dapat dimengerti karena keadaan sebelumnya.|Jangan selalu menerjemahkan だけ sebagai hanya; di sini menekankan sebab.|期待していただけに、残念でした。|きたいしていただけに、ざんねんでした。|Karena sangat berharap, saya merasa semakin kecewa.|人気の店だけに、予約が取りにくいです。|にんきの みせだけに、よやくが とりにくいです。|Karena restoran populer, sulit mendapatkan reservasi.
～だけあって|sesuai reputasi atau kemampuan … memang …|V普通形／Aい + だけあって; Aな + なだけあって; N + だけあって|Hasil baik sesuai kemampuan, pengalaman, harga, atau reputasi yang diharapkan.|Biasanya evaluasi positif; berbeda dari だけに yang juga cocok untuk kekecewaan.|経験が長いだけあって、説明が上手です。|けいけんが ながいだけあって、せつめいが じょうずです。|Sesuai pengalamannya yang panjang, dia memang pandai menjelaskan.|専門家だけあって、知識が豊富です。|せんもんかだけあって、ちしきが ほうふです。|Sebagai ahli, pengetahuannya memang luas.
'''),
('Kontras dan konsesi', 'Fakta yang bertentangan dengan dugaan', '''
～ものの|meskipun … kenyataannya …|V普通形／Aい + ものの; Aな + なものの; N + であるものの|Mengakui fakta pertama lalu menyampaikan keadaan yang tidak sesuai harapan; banyak dipakai tertulis.|Tidak selalu cocok untuk perintah atau ajakan langsung.|買ったものの、まだ使っていません。|かった ものの、まだ つかっていません。|Meskipun sudah membeli, saya belum menggunakannya.|便利なものの、値段が高いです。|べんりな ものの、ねだんが たかいです。|Meskipun praktis, harganya mahal.
～にもかかわらず|walaupun … tetap …|V普通形／Aい + にもかかわらず; Aな／N + であるにもかかわらず; N + にもかかわらず|Hasil melawan fakta atau dugaan, dengan nada formal atau penilaian pembicara.|N dapat langsung diikuti にもかかわらず; jangan gunakan な untuk sambungan sifat な.|雨にもかかわらず、多くの人が集まりました。|あめにも かかわらず、おおくの ひとが あつまりました。|Walaupun hujan, banyak orang berkumpul.|何度も確認したにもかかわらず、間違えました。|なんども かくにんしたにも かかわらず、まちがえました。|Walaupun sudah memeriksa berkali-kali, saya tetap salah.
'''),
('Kontras dan konsesi', 'Kontradiksi pada pelaku yang sama', '''
～ながら（も）：逆接|meskipun …|Vます語幹／Vている／Aい／N + ながら（も）; Aな語幹 + ながら（も）|Mengontraskan keadaan atau pengetahuan dengan perilaku, bukan dua tindakan bersamaan.|Arti ditentukan konteks; 知りながら biasanya meskipun mengetahui.|悪いと知りながら、嘘をつきました。|わるいと しりながら、うそを つきました。|Meskipun tahu itu buruk, saya berbohong.|狭いながらも、快適な部屋です。|せまい ながらも、かいてきな へやです。|Meskipun kecil, kamar ini nyaman.
～つつ（も）：逆接|meski menyadari … tetap …|Vます語幹 + つつ（も）|Kontradiksi antara pikiran atau niat dengan tindakan pelaku; bentuk tulisan.|Bedakan つつ simultan; kontras tampak dari isi kedua klausa.|休もうと思いつつ、仕事を続けました。|やすもうと おもいつつ、しごとを つづけました。|Meskipun berniat beristirahat, saya terus bekerja.|必要だと分かりつつも、申し込みを忘れました。|ひつようだと わかりつつも、もうしこみを わすれました。|Walaupun tahu itu perlu, saya lupa mendaftar.
'''),
('Kontras dan konsesi', 'Dua sisi dan pembatasan pernyataan', '''
～反面|di sisi lain memiliki sifat berlawanan …|V普通形／Aい + 反面; Aな + な反面; N + である反面|Dua sisi berlawanan dari objek atau situasi yang sama.|Jangan menghubungkan dua subjek tak terkait seolah dua sisi benda yang sama.|この仕事は大変な反面、やりがいがあります。|この しごとは たいへんな はんめん、やりがいが あります。|Pekerjaan ini berat, tetapi di sisi lain memuaskan.|一人暮らしは自由な反面、寂しいです。|ひとりぐらしは じゆうな はんめん、さびしいです。|Tinggal sendiri memberi kebebasan, tetapi juga kesepian.
～とはいえ|walaupun benar … tetapi …|V普通形／Aい／Aな語幹／N + とはいえ|Menerima pernyataan pertama, lalu memberi batasan atau koreksi yang perlu.|とはいえ juga dapat mengawali kalimat; bukan menyangkal seluruh fakta sebelumnya.|春とはいえ、朝はまだ寒いです。|はるとはいえ、あさは まだ さむいです。|Walaupun sudah musim semi, pagi masih dingin.|安いとはいえ、必要のない物は買いません。|やすいとはいえ、ひつようの ない ものは かいません。|Walaupun murah, saya tidak membeli barang yang tidak diperlukan.
'''),
('Kontras dan konsesi', 'Melawan prediksi dan pertentangan keras', '''
～に反して|bertentangan dengan …|N + に反して; N + に反する + N|Hasil atau tindakan berbeda dari harapan, prediksi, aturan, atau kehendak.|Penerang kata benda memakai に反する, bukan に反して langsung.|予想に反して、試験は簡単でした。|よそうに はんして、しけんは かんたんでした。|Bertentangan dengan perkiraan, ujian itu mudah.|親の希望に反して、彼は海外で働いています。|おやの きぼうに はんして、かれは かいがいで はたらいています。|Bertentangan dengan harapan orang tua, dia bekerja di luar negeri.
～どころか|jangankan … justru …|V普通形／Aい／Aな語幹／N + どころか|Menyangkal dugaan dengan kenyataan yang jauh berbeda atau berlawanan.|Bukan selalu lebih sedikit; arah perbedaan ditentukan konteks.|涼しいどころか、暑くて眠れません。|すずしいどころか、あつくて ねむれません。|Jangankan sejuk, panasnya membuat saya tidak bisa tidur.|貯金するどころか、借金が増えました。|ちょきんするどころか、しゃっきんが ふえました。|Jangankan menabung, utang saya justru bertambah.
'''),
('Syarat dan hipotesis', 'Syarat minimal dan syarat mutlak', '''
～さえ～ば|asalkan … saja …|N + さえ + Vば; Vます語幹 + さえすれば; Aい→くさえあれば; Aな／N + でさえあれば|Menjadikan satu syarat sebagai syarat minimum yang cukup, pada konteks N2 yang lebih kompleks.|Harus ada bentuk bersyarat: Nさえ saja dapat berarti bahkan.|住所さえ分かれば、届けられます。|じゅうしょさえ わかれば、とどけられます。|Asalkan tahu alamatnya, barang bisa diantar.|静かでさえあれば、どこでも勉強できます。|しずかでさえ あれば、どこでも べんきょうできます。|Asalkan tenang, saya bisa belajar di mana saja.
～ないことには|tanpa … tidak dapat …|Vない + ことには + hasil negatif|Tindakan awal merupakan prasyarat mutlak bagi penilaian atau hasil; penguatan materi sebelumnya.|Akhirnya lazim negatif; jangan menghilangkan ない.|本人に聞かないことには、理由は分かりません。|ほんにんに きかないことには、りゆうは わかりません。|Tanpa bertanya langsung kepada orangnya, alasan itu tidak diketahui.|資金を集めないことには、始められません。|しきんを あつめないことには、はじめられません。|Tanpa mengumpulkan dana, kami tidak bisa memulai.
'''),
('Syarat dan hipotesis', 'Pengandaian netral', '''
～としたら／～とすれば|jika diasumsikan …|V普通形／Aい + としたら／とすれば; Aな／N + だとしたら／だとすれば|Menerima situasi hipotetis sebagai dasar tindakan atau kesimpulan.|Kalimat nominal memerlukan だ sebelum と; bedakan kutipan biasa.|引っ越すとしたら、駅の近くがいいです。|ひっこすとしたら、えきの ちかくが いいです。|Jika pindah rumah, saya ingin dekat stasiun.|それが事実だとすれば、調査が必要です。|それが じじつだとすれば、ちょうさが ひつようです。|Jika itu benar, diperlukan penyelidikan.
～としても|sekalipun diasumsikan …|V普通形／Aい + としても; Aな／N + だとしても|Hasil atau pendirian tidak berubah walau hipotesis pertama diterima.|Bukan syarat yang membuat hasil berubah seperti としたら.|失敗するとしても、挑戦したいです。|しっぱいするとしても、ちょうせんしたいです。|Sekalipun mungkin gagal, saya ingin mencoba.|参加者が少ないとしても、開催します。|さんかしゃが すくないとしても、かいさいします。|Sekalipun pesertanya sedikit, acara akan diselenggarakan.
'''),
('Syarat dan hipotesis', 'Hipotesis sulit dan akibat berbahaya', '''
～ものなら|kalau bisa … yang sulit diwujudkan …|V可能形辞書形 + ものなら|Keinginan pada kondisi yang dianggap sulit atau hampir mustahil.|Untuk hipotesis tindakan berakibat buruk, gunakan ようものなら.|戻れるものなら、学生時代に戻りたいです。|もどれる ものなら、がくせいじだいに もどりたいです。|Kalau bisa kembali, saya ingin kembali ke masa kuliah.|休めるものなら、今日は休みたいです。|やすめる ものなら、きょうは やすみたいです。|Kalau bisa mengambil libur, saya ingin libur hari ini.
～ようものなら|kalau sampai berani … akibat buruk …|V意向形 + ものなら|Pengandaian kuat dengan akibat buruk atau reaksi keras.|Gunakan bentuk kehendak: 言おうものなら, bukan 言うものなら untuk pola ini.|秘密を話そうものなら、大変なことになります。|ひみつを はなそうものなら、たいへんな ことに なります。|Kalau sampai membocorkan rahasia, akibatnya akan gawat.|連絡せずに休もうものなら、上司に叱られます。|れんらくせずに やすもうものなら、じょうしに しかられます。|Kalau sampai tidak masuk tanpa memberi kabar, saya akan dimarahi atasan.
'''),
('Batas dan pengecualian', 'Selama kondisi dan lingkup pengetahuan', '''
～限り（は）|selama … berlaku …|V辞書形／Vている／Vない／Aい + 限り; Aな + な限り; N + である限り|Keadaan pertama membatasi masa berlaku hasil atau prinsip berikut.|Bukan jaminan selamanya; hasil berlaku dalam kondisi yang disebut.|ここに住んでいる限り、規則を守ってください。|ここに すんでいる かぎり、きそくを まもってください。|Selama tinggal di sini, tolong patuhi peraturan.|体が丈夫な限り、働きたいです。|からだが じょうぶな かぎり、はたらきたいです。|Selama tubuh sehat, saya ingin bekerja.
～限りでは|sejauh informasi atau pengamatan …|V辞書形／Vた／Vている + 限りでは; Nの + 限りでは|Membatasi kepastian penilaian pada apa yang diketahui atau diamati.|Bukan makna selama; kesimpulan dapat berubah jika informasi bertambah.|私が知る限りでは、彼はまだ東京にいます。|わたしが しる かぎりでは、かれは まだ とうきょうに います。|Sejauh yang saya tahu, dia masih di Tokyo.|調べた限りでは、問題はありません。|しらべた かぎりでは、もんだいは ありません。|Sejauh hasil pemeriksaan, tidak ada masalah.
'''),
('Batas dan pengecualian', 'Batas layanan dan kejadian khusus', '''
～に限り|khusus terbatas pada …|N + に限り|Pengecualian atau layanan khusus yang berlaku hanya pada golongan atau waktu tertentu; formal.|Jangan campur dengan に限って yang sering menyatakan kebetulan buruk atau keyakinan.|本日に限り、入場は無料です。|ほんじつに かぎり、にゅうじょうは むりょうです。|Khusus hari ini saja, masuk gratis.|会員に限り、この部屋を利用できます。|かいいんに かぎり、この へやを りようできます。|Hanya anggota yang dapat memakai ruangan ini.
～に限って|justru pada …; khusus orang itu tidak mungkin …|N + に限って|Konteks menentukan kebetulan buruk pada saat tertentu atau keyakinan kuat tentang seseorang.|Jangan menyamakan semua に限って dengan batas administratif に限り.|急いでいる日に限って、電車が遅れます。|いそいでいる ひに かぎって、でんしゃが おくれます。|Justru pada hari saya terburu-buru, kereta terlambat.|彼に限って、約束を忘れるはずがありません。|かれに かぎって、やくそくを わすれる はずが ありません。|Khusus dia, tidak mungkin lupa janji.
'''),
('Batas dan pengecualian', 'Melampaui batas dan bergantung pada faktor', '''
～に限らず|tidak terbatas pada …|N + に限らず|Memperluas cakupan dari contoh pertama ke kelompok lain.|Sering diikuti も; jangan menafsirkan sebagai hanya.|この映画は若者に限らず、年配の人にも人気です。|この えいがは わかものに かぎらず、ねんぱいの ひとにも にんきです。|Film ini populer bukan hanya pada pemuda, tetapi juga orang yang lebih tua.|平日に限らず、週末も営業しています。|へいじつに かぎらず、しゅうまつも えいぎょうしています。|Kami buka bukan hanya pada hari kerja, tetapi juga akhir pekan.
～次第で／～次第だ|bergantung pada …|N + 次第で／次第だ|Hasil berubah menurut faktor yang disebut, seperti usaha atau keputusan.|Berbeda dari Vます語幹次第 yang berarti segera setelah.|結果は努力次第です。|けっかは どりょくしだいです。|Hasil bergantung pada usaha.|天気次第で、予定を変えます。|てんきしだいで、よていを かえます。|Kami mengubah rencana bergantung pada cuaca.
'''),
('Perubahan dan hubungan', 'Penyesuaian dan perubahan bertahap', '''
～に応じて|sesuai dengan …|N + に応じて; N + に応じた + N|Menyesuaikan tindakan atau besaran dengan kebutuhan, kemampuan, atau keadaan.|Penerang N memakai に応じた, bukan に応じて langsung.|能力に応じて、仕事を分けます。|のうりょくに おうじて、しごとを わけます。|Kami membagi tugas sesuai kemampuan.|必要に応じて、説明を追加します。|ひつように おうじて、せつめいを ついかします。|Kami menambahkan penjelasan sesuai kebutuhan.
～につれて|seiring perubahan …|V辞書形／N + につれて|Perubahan pertama diikuti perubahan lain secara bertahap, biasanya alami.|Hindari hasil berupa perintah; dua bagian biasanya sama-sama berubah.|年を取るにつれて、考え方が変わりました。|としを とるにつれて、かんがえかたが かわりました。|Seiring bertambah usia, cara berpikir saya berubah.|気温の上昇につれて、客が増えました。|きおんの じょうしょうにつれて、きゃくが ふえました。|Seiring kenaikan suhu, jumlah pelanggan bertambah.
'''),
('Perubahan dan hubungan', 'Mengikuti proses dan perubahan yang menyertai', '''
～に従って|seiring …; mengikuti …|V辞書形／N + に従って|Dapat menandai perubahan bertahap atau mengikuti instruksi; perhatikan kata sebelum pola.|Makna mengikuti aturan berbeda dari dua perubahan yang bergerak bersama.|説明書に従って、組み立ててください。|せつめいしょに したがって、くみたててください。|Tolong rakit mengikuti petunjuk.|高度が上がるに従って、気温が下がります。|こうどが あがるに したがって、きおんが さがります。|Seiring ketinggian bertambah, suhu turun.
～に伴って|seiring atau bersamaan dengan …|N／V辞書形 + に伴って|Perubahan besar disertai perubahan lain; umum dalam laporan sosial dan bisnis.|伴う dibaca ともなう, bukan ばんう; bukan selalu hubungan sebab langsung.|人口の増加に伴って、住宅が増えました。|じんこうの ぞうかに ともなって、じゅうたくが ふえました。|Seiring pertambahan penduduk, perumahan bertambah.|制度が変わるに伴って、手続きも変わります。|せいどが かわるに ともなって、てつづきも かわります。|Seiring perubahan sistem, prosedur juga berubah.
'''),
('Perubahan dan hubungan', 'Bersama dan rentang', '''
～とともに|bersama …; seiring …|N／V辞書形 + とともに|Bisa menyatakan kebersamaan, dua hal sekaligus, atau perubahan paralel.|Pilih arti berdasarkan konteks, jangan selalu menerjemahkan dengan teman.|家族とともに、新しい生活を始めました。|かぞくと ともに、あたらしい せいかつを はじめました。|Bersama keluarga, saya memulai kehidupan baru.|時間がたつとともに、不安が減りました。|じかんが たつと ともに、ふあんが へりました。|Seiring waktu berlalu, kecemasan berkurang.
～にわたって|sepanjang rentang …|N durasi／wilayah + にわたって; N + にわたる + N|Cakupan menyeluruh suatu masa atau wilayah, sering rentang yang luas.|Untuk penerang N gunakan にわたる: 三日間にわたる会議.|会議は三日間にわたって行われました。|かいぎは みっかかんに わたって おこなわれました。|Rapat berlangsung selama tiga hari.|被害は広い地域にわたって報告されました。|ひがいは ひろい ちいきに わたって ほうこくされました。|Kerusakan dilaporkan di wilayah yang luas.
'''),
('Sarana dan topik formal', 'Perantara dan sepanjang masa', '''
～を通じて／～を通して：媒介|melalui perantara …|N + を通じて／を通して|Sarana, pengalaman, atau orang menghubungkan pelaku dengan hasil.|通じて dibaca つうじて; 通して dibaca とおして.|友人を通じて、その会社を知りました。|ゆうじんを つうじて、その かいしゃを しりました。|Melalui teman, saya mengetahui perusahaan itu.|交流を通して、文化を学びました。|こうりゅうを とおして、ぶんかを まなびました。|Melalui pertukaran, saya belajar budaya.
～を通じて／～を通して：期間|sepanjang …|N periode + を通じて／を通して|Keadaan berlaku sepanjang satu periode; bukan perantara orang.|Kata waktu menentukan makna periode, berbeda dari 友人を通じて.|この町は一年を通じて暖かいです。|この まちは いちねんを つうじて あたたかいです。|Kota ini hangat sepanjang tahun.|夏休みを通して、図書館で勉強しました。|なつやすみを とおして、としょかんで べんきょうしました。|Sepanjang libur musim panas, saya belajar di perpustakaan.
'''),
('Sarana dan topik formal', 'Lingkup formal dan isu yang diperdebatkan', '''
～において／～における|di atau dalam lingkup …|N + において; N + における + N|Bentuk formal untuk lokasi, masa, bidang, atau situasi suatu kejadian.|における menerangkan N; jangan memakai において langsung sebelum N.|会議は本社において行われます。|かいぎは ほんしゃに おいて おこなわれます。|Rapat diadakan di kantor pusat.|教育における課題を話し合いました。|きょういくに おける かだいを はなしあいました。|Kami membahas tantangan dalam pendidikan.
～をめぐって|seputar persoalan …|N + をめぐって; N + をめぐる + N|Topik menimbulkan diskusi, pertentangan, atau persaingan di antara pihak-pihak.|Bukan untuk semua topik netral; ada kegiatan beberapa pihak di sekeliling isu.|予算をめぐって、議論が続いています。|よさんを めぐって、ぎろんが つづいています。|Perdebatan seputar anggaran terus berlangsung.|土地をめぐる争いが起きました。|とちを めぐる あらそいが おきました。|Terjadi sengketa mengenai tanah.
'''),
('Dasar dan arah tindakan', 'Dasar keputusan dan bahan pembentukan', '''
～に基づいて|berdasarkan dasar atau aturan …|N + に基づいて; N + に基づく + N|Fakta, hukum, atau data menjadi landasan penilaian dan pelaksanaan.|基づく dibaca もとづく; bukan を基づいて.|データに基づいて、判断します。|データに もとづいて、はんだんします。|Kami menilai berdasarkan data.|法律に基づく手続きが必要です。|ほうりつに もとづく てつづきが ひつようです。|Diperlukan prosedur berdasarkan hukum.
～をもとに（して）|dengan bahan atau sumber …|N + をもとに（して）; N + をもとにした + N|Bahan awal diolah menjadi karya, rencana, atau gagasan baru.|Lebih menekankan sumber bahan daripada kewajiban mengikuti aturan.|実話をもとに、小説を書きました。|じつわを もとに、しょうせつを かきました。|Saya menulis novel berdasarkan kisah nyata.|意見をもとにした計画です。|いけんを もとにした けいかくです。|Ini rencana yang dibuat dari masukan.
'''),
('Dasar dan arah tindakan', 'Mengikuti arah dan menjawab harapan', '''
～に沿って|mengikuti arah atau pedoman …|N + に沿って; N + に沿った + N|Tindakan sesuai kebijakan atau jalur; juga dapat berarti sepanjang rute fisik.|に沿った menerangkan N; jangan mengganti に dengan を.|方針に沿って、計画を立てました。|ほうしんに そって、けいかくを たてました。|Kami menyusun rencana sesuai kebijakan.|川に沿って、歩きました。|かわに そって、あるきました。|Saya berjalan menyusuri sungai.
～にこたえて|menjawab harapan atau permintaan …|N + にこたえて|Usaha atau tindakan memenuhi harapan, permintaan, atau kepercayaan.|Bukan sekadar menjawab pertanyaan; 質問に答える memakai makna jawaban biasa.|要望にこたえて、営業時間を延ばしました。|ようぼうに こたえて、えいぎょうじかんを のばしました。|Menanggapi permintaan, kami memperpanjang jam buka.|期待にこたえて、良い結果を出したいです。|きたいに こたえて、よい けっかを だしたいです。|Saya ingin meraih hasil baik untuk memenuhi harapan.
'''),
('Dasar dan arah tindakan', 'Tujuan dan cakupan yang terkait', '''
～に向けて|menuju target …|N + に向けて; N + に向けた + N|Persiapan atau gerak diarahkan ke target, waktu, atau kelompok penerima.|Bukan sekadar lokasi tujuan fisik; ada arah usaha atau sasaran pesan.|大会に向けて、練習しています。|たいかいに むけて、れんしゅうしています。|Saya berlatih menuju kejuaraan.|学生に向けた説明会を開きます。|がくせいに むけた せつめいかいを ひらきます。|Kami mengadakan sesi penjelasan untuk mahasiswa.
～に関わる|berkaitan erat atau menyangkut …|N + に関わる + N; N + に関わって|Menandai hal yang berhubungan penting dengan kehidupan, pekerjaan, atau masalah tertentu.|Bedakan にかかわらず yang berarti tanpa memandang; bukan makna kebalikan sederhana.|命に関わる問題です。|いのちに かかわる もんだいです。|Ini masalah yang menyangkut nyawa.|教育に関わる仕事をしています。|きょういくに かかわる しごとを しています。|Saya bekerja di bidang yang berkaitan dengan pendidikan.
'''),
('Penilaian dan kepastian', 'Kesimpulan yang dapat dijelaskan', '''
～わけだ|jadi wajar atau berarti …|V普通形／Aい + わけだ; Aな／N + なわけだ|Menarik kesimpulan logis setelah mengetahui alasan atau fakta.|Tidak selalu berarti alasan fisik; sering menunjukkan pemahaman baru.|毎日練習しているので、上手なわけです。|まいにち れんしゅうしているので、じょうずな わけです。|Karena berlatih setiap hari, pantas dia mahir.|駅が近いから、便利なわけです。|えきが ちかいから、べんりな わけです。|Karena dekat stasiun, pantas tempatnya praktis.
～わけではない|bukan berarti selalu atau sepenuhnya …|V普通形／Aい + わけではない; Aな／N + なわけではない|Menyangkal kesimpulan berlebihan atau membatasi pernyataan, bukan menolak seluruh fakta.|安いわけではない bukan persis tidak murah; dapat berarti tidak bisa disebut murah.|嫌いなわけではありませんが、今は食べたくありません。|きらいな わけでは ありませんが、いまは たべたくありません。|Bukannya saya tidak suka, tetapi sekarang tidak ingin makan.|全員が賛成しているわけではありません。|ぜんいんが さんせいしている わけでは ありません。|Bukan berarti semua orang setuju.
'''),
('Penilaian dan kepastian', 'Ketidakmungkinan dan hambatan sosial', '''
～わけがない|mustahil …|V普通形／Aい + わけがない; Aな／N + なわけがない|Penolakan kuat berdasarkan alasan yang diyakini pembicara.|Bukan larangan sosial; itu memakai わけにはいかない.|一日で全部覚えられるわけがありません。|いちにちで ぜんぶ おぼえられる わけが ありません。|Mustahil menghafal semuanya dalam satu hari.|彼がそんな嘘をつくわけがありません。|かれが そんな うそを つく わけが ありません。|Mustahil dia mengatakan kebohongan seperti itu.
～わけにはいかない|tidak bisa karena tanggung jawab …|V辞書形 + わけにはいかない|Secara fisik mungkin, tetapi norma, janji, atau keadaan menghalangi.|Jangan menyamakan dengan kemampuan negatif: 泳げない ialah tidak mampu berenang.|明日は試験なので、遊ぶわけにはいきません。|あしたは しけんなので、あそぶ わけには いきません。|Karena besok ujian, saya tidak bisa bersantai bermain.|秘密なので、話すわけにはいきません。|ひみつなので、はなす わけには いきません。|Karena itu rahasia, saya tidak boleh membicarakannya.
'''),
('Penilaian dan kepastian', 'Kewajiban dan penolakan dugaan', '''
～ないわけにはいかない|tidak bisa tidak …; harus …|Vない + わけにはいかない|Kewajiban sosial atau moral membuat pembicara harus melakukan tindakan.|Negatif ganda berarti wajib melakukan, bukan tidak boleh melakukan.|招待されたので、行かないわけにはいきません。|しょうたいされたので、いかない わけには いきません。|Karena diundang, saya harus hadir.|借りた物は返さないわけにはいきません。|かりた ものは かえさない わけには いきません。|Barang pinjaman harus dikembalikan.
～はずがない|tidak mungkin menurut penalaran …|V普通形／Aい + はずがない; Aな + なはずがない; N + のはずがない|Menolak kemungkinan berdasarkan informasi atau dugaan yang kuat.|N memakai の, berbeda dari Nなわけがない.|まだ朝六時なので、店が開いているはずがありません。|まだ あさ ろくじなので、みせが あいている はずが ありません。|Karena baru pukul enam pagi, toko tidak mungkin sudah buka.|確認したので、間違いのはずがありません。|かくにんしたので、まちがいの はずが ありません。|Karena sudah saya periksa, itu semestinya bukan kesalahan.
'''),
('Penilaian dan kepastian', 'Keyakinan dan kepastian formal', '''
～に違いない|pasti … berdasarkan dugaan kuat|V普通形／Aい／Aな語幹／N + に違いない|Pembicara menarik dugaan yang sangat kuat dari petunjuk yang ada.|Tetap dugaan pembicara, bukan bukti objektif mutlak.|電気がついているので、誰かいるに違いありません。|でんきが ついているので、だれか いるに ちがいありません。|Karena lampu menyala, pasti ada seseorang.|彼はこの話を知っているに違いありません。|かれは この はなしを しっているに ちがいありません。|Dia pasti mengetahui cerita ini.
～に相違ない|tidak diragukan bahwa …|V普通形／Aい／Aな語幹／N + に相違ない|Keyakinan kuat dengan gaya lebih resmi daripada に違いない; banyak dipakai dalam pernyataan tertulis.|相違 dibaca そうい; jangan menambah だ setelah N pada sambungan ini.|これは本人の署名に相違ありません。|これは ほんにんの しょめいに そういありません。|Ini tidak diragukan merupakan tanda tangan orang itu sendiri.|記録は正確に相違ありません。|きろくは せいかくに そういありません。|Catatan itu tidak diragukan akurat.
'''),
('Penilaian dan kepastian', 'Batas kepastian dan kemungkinan', '''
～とは限らない|belum tentu …|V普通形／Aい + とは限らない; Aな／N + だとは限らない|Menyangkal bahwa suatu kesimpulan berlaku pada semua kasus.|Bukan selalu berarti kesimpulan itu salah; ada pengecualian.|高い物が良いとは限りません。|たかい ものが よいとは かぎりません。|Barang mahal belum tentu bagus.|努力すれば必ず成功するとは限りません。|どりょくすれば かならず せいこうするとは かぎりません。|Berusaha tidak selalu menjamin keberhasilan.
～得る／～得ない|mungkin terjadi／tidak mungkin terjadi|Vます語幹 + 得る（える／うる）; Vます語幹 + 得ない（えない）|Kemungkinan suatu kejadian secara teori, bukan kemampuan keterampilan seseorang.|得ない dibaca えない; うない salah. Bentuk ます ialah 得ます（えます）.|誰にでも起こり得る問題です。|だれにでも おこりうる もんだいです。|Ini masalah yang bisa terjadi pada siapa pun.|その説明はあり得ません。|その せつめいは ありえません。|Penjelasan itu tidak mungkin benar.
'''),
('Terpaksa dan kemampuan', 'Sulit memenuhi dan risiko', '''
～かねる|sulit atau tidak dapat memenuhi …|Vます語幹 + かねる|Penolakan sopan atau kesulitan mengambil tindakan karena posisi dan pertimbangan.|Bukan ketidakmampuan fisik umum; できません lebih netral.|その質問には答えかねます。|その しつもんには こたえかねます。|Saya sulit menjawab pertanyaan itu.|すぐには判断しかねます。|すぐには はんだんしかねます。|Saya belum dapat menilai segera.
～かねない|berisiko atau bisa saja … buruk|Vます語幹 + かねない|Kemungkinan negatif yang dikhawatirkan pembicara.|Negatif bentuknya tidak berarti mustahil; justru mungkin terjadi.|このままでは、事故が起こりかねません。|このままでは、じこが おこりかねません。|Jika begini terus, kecelakaan bisa terjadi.|無理をすると、体を壊しかねません。|むりを すると、からだを こわしかねません。|Jika memaksakan diri, Anda bisa jatuh sakit.
'''),
('Terpaksa dan kemampuan', 'Terpaksa memilih dan tidak sepenuhnya mustahil', '''
～ざるを得ない|terpaksa harus …|Vない tanpa ない + ざるを得ない; する→せざるを得ない|Tidak ada pilihan lain karena keadaan luar; gaya formal.|する menjadi せざる, bukan しざる; 得ない dibaca えない.|電車が止まったので、歩かざるを得ませんでした。|でんしゃが とまったので、あるかざるを えませんでした。|Karena kereta berhenti beroperasi, saya terpaksa berjalan.|予算が足りず、計画を変更せざるを得ません。|よさんが たりず、けいかくを へんこうせざるを えません。|Karena anggaran kurang, kami terpaksa mengubah rencana.
～ないことはない／～ないこともない|bukan tidak bisa; mungkin dengan syarat …|Vない／Aい→くない／Aな・N→ではない + ことはない／こともない|Mengakui kemungkinan terbatas dengan nada ragu atau syarat tersirat.|Negatif ganda tidak otomatis menjamin kemauan atau kemampuan penuh.|読めないことはありませんが、時間がかかります。|よめない ことは ありませんが、じかんが かかります。|Bukan tidak bisa membaca, tetapi butuh waktu.|安くないこともありませんが、品質が心配です。|やすくない ことも ありませんが、ひんしつが しんぱいです。|Memang bisa dibilang murah, tetapi kualitasnya mengkhawatirkan.
'''),
('Perasaan dan dorongan', 'Dorongan yang tidak tertahan', '''
～ずにはいられない|tidak bisa menahan diri untuk tidak …|Vない tanpa ない + ずにはいられない; する→せずにはいられない|Dorongan atau reaksi emosional membuat tindakan tidak tertahan.|する memakai せず; bedakan kewajiban luar ざるを得ない.|その話を聞いて、笑わずにはいられませんでした。|その はなしを きいて、わらわずには いられませんでした。|Mendengar cerita itu, saya tidak bisa menahan tawa.|困っている人を見ると、助けずにはいられません。|こまっている ひとを みると、たすけずには いられません。|Melihat orang kesulitan, saya tidak bisa menahan diri untuk membantu.
～ないではいられない|tidak tahan untuk tidak …|Vない + ではいられない|Dorongan spontan dari dalam; lebih dekat bahasa sehari-hari daripada ずにはいられない.|Negatif ganda di sini bermakna akhirnya melakukan tindakan.|結果が気になって、確認しないではいられません。|けっかが きになって、かくにんしないでは いられません。|Saya penasaran pada hasilnya dan tidak tahan untuk tidak memeriksa.|この景色を見ると、写真を撮らないではいられません。|この けしきを みると、しゃしんを とらないでは いられません。|Melihat pemandangan ini, saya tidak tahan untuk tidak memotret.
'''),
('Perasaan dan dorongan', 'Perasaan sangat kuat', '''
～てたまらない|sangat … sampai tak tertahankan|Vて／Aい→くて／Aな→で + たまらない|Sensasi, keinginan, atau perasaan sangat kuat; biasanya pengalaman pembicara.|Untuk perasaan orang lain perlu konteks atau penanda pengamatan seperti ようだ.|喉が渇いてたまりません。|のどが かわいて たまりません。|Saya sangat haus sampai tidak tahan.|合格できて、うれしくてたまりません。|ごうかくできて、うれしくて たまりません。|Saya sangat gembira karena berhasil lulus.
～てならない|sangat atau terus terasa …|Vて／Aい→くて／Aな→で + ならない|Perasaan atau pikiran timbul alami dan sulit dihilangkan; umum untuk 心配, 残念, 気になる.|Tidak cocok untuk setiap keadaan fisik seperti kaki sangat panjang.|家族のことが心配でなりません。|かぞくの ことが しんぱいで なりません。|Saya terus merasa sangat cemas tentang keluarga.|結果が気になってなりません。|けっかが きになって なりません。|Saya terus-menerus penasaran tentang hasilnya.
'''),
('Derajat dan penegasan', 'Sekadar dan penyebab utama', '''
～にすぎない|hanya sebatas …|V普通形／Aい／Aな語幹／N + にすぎない|Mengecilkan jumlah, status, atau pentingnya sesuatu menjadi batas tertentu.|Bukan selalu kritik; dapat membatasi klaim agar lebih tepat.|これは一つの例にすぎません。|これは ひとつの れいに すぎません。|Ini hanyalah satu contoh.|私は手伝ったにすぎません。|わたしは てつだったに すぎません。|Saya hanya membantu saja.
～にほかならない|tidak lain adalah …|N + にほかならない; alasan + からにほかならない|Penegasan identitas atau sebab yang diyakini sebagai inti penjelasan; formal.|N langsung diikuti に; sebab berbentuk klausa memerlukan から.|成功は努力の結果にほかなりません。|せいこうは どりょくの けっかに ほかなりません。|Keberhasilan tidak lain adalah hasil usaha.|続けられたのは、仲間がいたからにほかなりません。|つづけられたのは、なかまが いたからに ほかなりません。|Saya dapat bertahan tidak lain karena ada rekan.
'''),
('Derajat dan penegasan', 'Pilihan terbaik dan tidak perlu', '''
～に越したことはない|tidak ada yang lebih baik daripada …|V辞書形／Vない／Aい／Aな語幹／N + に越したことはない|Saran umum: sesuatu lebih baik jika memungkinkan, tanpa menyatakan kewajiban mutlak.|Bukan larangan atau syarat wajib; あれば lebih aman adalah pilihan ideal.|準備は早いに越したことはありません。|じゅんびは はやいに こした ことは ありません。|Persiapan lebih awal tentu lebih baik.|荷物は少ないに越したことはありません。|にもつは すくないに こした ことは ありません。|Barang bawaan lebih sedikit tentu lebih baik.
～までもない|tidak perlu sampai …|V辞書形 + までもない|Tindakan dianggap berlebihan atau tidak diperlukan karena hal sudah jelas.|Bukan tidak mampu; 言うまでもない berarti tak perlu dikatakan.|近いので、タクシーに乗るまでもありません。|ちかいので、タクシーに のるまでも ありません。|Karena dekat, tidak perlu sampai naik taksi.|その大切さは言うまでもありません。|その たいせつさは いうまでも ありません。|Pentingnya hal itu tidak perlu dikatakan lagi.
'''),
('Derajat dan penegasan', 'Lebih dari satu unsur', '''
～ばかりか|bukan hanya … bahkan …|V普通形／Aい + ばかりか; Aな + なばかりか; N + ばかりか|Menambahkan unsur kedua yang lebih kuat atau tak terduga; sering diikuti も.|Berbeda dari Vたばかり yang berarti baru saja.|彼は英語ばかりか、中国語も話せます。|かれは えいごばかりか、ちゅうごくごも はなせます。|Dia bukan hanya bisa berbahasa Inggris, tetapi juga Mandarin.|遅刻したばかりか、資料も忘れました。|ちこくしたばかりか、しりょうも わすれました。|Saya bukan hanya terlambat, tetapi juga lupa membawa dokumen.
～のみならず|tidak hanya … tetapi juga …|V普通形／Aい + のみならず; Aな／N + であるのみならず; N + のみならず|Penambahan cakupan dalam tulisan dan pidato formal, sering dipasangkan dengan も.|Gaya lebih formal daripada だけでなく; bukan sekadar hanya.|国内のみならず、海外でも販売しています。|こくないのみならず、かいがいでも はんばいしています。|Kami menjual bukan hanya di dalam negeri, tetapi juga di luar negeri.|価格が安いのみならず、品質も良いです。|かかくが やすいのみならず、ひんしつも よいです。|Bukan hanya harganya murah, kualitasnya juga bagus.
'''),
('Derajat dan penegasan', 'Hal yang sudah tentu dan bahkan', '''
～はもとより|tentu … dan juga …|N + はもとより|Unsur pertama sudah dianggap jelas, lalu menambahkan unsur lain; gaya formal.|Jangan menafsirkan もと sebagai asal lokasi dalam pola ini.|社員はもとより、家族も参加できます。|しゃいんは もとより、かぞくも さんかできます。|Pegawai tentu boleh ikut, dan keluarga juga boleh.|文法はもとより、発音も大切です。|ぶんぽうは もとより、はつおんも たいせつです。|Tata bahasa tentu penting, dan pelafalan juga penting.
～さえ／～すら|bahkan …|N + さえ／すら; beberapa partikel + さえ／すら|Mengangkat contoh ekstrem yang tidak terduga, sering dalam klausa negatif.|Bedakan dari さえ～ば yang mensyaratkan cukup satu unsur.|忙しくて、昼食を食べる時間さえありません。|いそがしくて、ちゅうしょくを たべる じかんさえ ありません。|Karena sibuk, bahkan waktu makan siang pun tidak ada.|彼は自分の名前すら書けませんでした。|かれは じぶんの なまえすら かけませんでした。|Dia bahkan tidak dapat menulis namanya sendiri.
'''),
('Evaluasi dan harapan', 'Standar dugaan', '''
～にしては|untuk ukuran … ternyata …|V普通形／N + にしては|Hasil berbeda dari standar yang dibayangkan untuk status atau keadaan tersebut.|N tidak memakai の; untuk perbandingan angka kabur sering わりに lebih sesuai.|初めてにしては、上手にできました。|はじめてに しては、じょうずに できました。|Untuk percobaan pertama, hasilnya bagus.|冬にしては、今日は暖かいです。|ふゆに しては、きょうは あたたかいです。|Untuk musim dingin, hari ini hangat.
～わりに（は）|dibanding yang diperkirakan …|V普通形／Aい + わりに; Aな + なわりに; N + のわりに|Perbandingan hasil dengan harapan yang didasarkan pada keadaan atau derajat.|N memakai の, berbeda dari Nにしては.|値段のわりに、品質が良いです。|ねだんの わりに、ひんしつが よいです。|Dibanding harganya, kualitasnya bagus.|長く勉強したわりに、点数が低かったです。|ながく べんきょうした わりに、てんすうが ひくかったです。|Walaupun belajar lama, nilainya lebih rendah dari yang diharapkan.
'''),
('Evaluasi dan harapan', 'Cenderung dan bernuansa', '''
～がち|cenderung sering …|Vます語幹／N + がち; がちな + N|Kecenderungan berulang, sering tidak diinginkan, seperti lupa atau absen.|Bukan untuk satu kejadian; berbeda dari 気味 yang menandai keadaan agak terasa.|忙しいと、食事を抜きがちです。|いそがしいと、しょくじを ぬきがちです。|Saat sibuk, saya cenderung melewatkan makan.|冬は曇りがちです。|ふゆは くもりがちです。|Pada musim dingin, cuaca cenderung mendung.
～気味|agak terasa …|Vます語幹／N + 気味（ぎみ）|Keadaan ringan yang terasa saat ini, sering kelelahan atau gejala kurang baik.|Tidak sama dengan kebiasaan berulang がち; 気味 dibaca ぎみ.|最近、疲れ気味です。|さいきん、つかれぎみです。|Akhir-akhir ini saya agak kelelahan.|風邪気味なので、早く寝ます。|かぜぎみなので、はやく ねます。|Karena terasa agak masuk angin, saya akan tidur lebih awal.
'''),
('Evaluasi dan harapan', 'Kesan sifat dan penuh sesuatu', '''
～っぽい|terkesan …; mudah …|N／sebagian Vます語幹 + っぽい; konjugasi seperti Aい|Sifat mirip sesuatu atau kecenderungan mudah melakukan; sering bernada kritik.|Bukan produktif untuk semua kata; gunakan gabungan lazim seperti 忘れっぽい.|彼の話し方は子供っぽいです。|かれの はなしかたは こどもっぽいです。|Cara bicaranya terkesan kekanak-kanakan.|最近、忘れっぽくなりました。|さいきん、わすれっぽく なりました。|Belakangan ini saya menjadi mudah lupa.
～だらけ|penuh dengan … yang kurang baik|N + だらけ; Nだらけの + N|Banyak unsur yang tidak diinginkan tersebar pada sesuatu.|Jangan selalu pakai untuk benda positif; 花がいっぱい lebih netral daripada 花だらけ.|答案は間違いだらけでした。|とうあんは まちがいだらけでした。|Lembar jawaban itu penuh kesalahan.|泥だらけの靴を洗いました。|どろだらけの くつを あらいました。|Saya mencuci sepatu yang penuh lumpur.
'''),
('Ungkapan dan penjelasan', 'Penilaian umum dan pembatasannya', '''
～というものだ|itulah yang disebut …; memang …|V普通形／Aい／Aな語幹／N + というものだ|Pembicara menilai sesuatu sebagai wajar, masuk akal, atau sesuai hakikatnya.|Bukan definisi kamus semata; sering mengandung penilaian pribadi.|一人で全部するのは無理というものです。|ひとりで ぜんぶ するのは むりという ものです。|Mengerjakan semuanya sendiri memang tidak masuk akal.|互いに支え合うのが友情というものです。|たがいに ささえあうのが ゆうじょうという ものです。|Saling mendukung itulah yang disebut persahabatan.
～というものではない|tidak sesederhana bahwa …|V普通形／Aい／Aな語幹／N + というものではない|Menolak generalisasi atau anggapan bahwa satu faktor cukup menjamin hasil.|Bukan penyangkalan benda fisik; fokusnya pemikiran yang terlalu sederhana.|お金があれば幸せになれるというものではありません。|おかねが あれば しあわせに なれるという ものでは ありません。|Memiliki uang tidak otomatis membuat orang bahagia.|謝れば済むというものではありません。|あやまれば すむという ものでは ありません。|Masalah ini tidak otomatis selesai hanya dengan meminta maaf.
'''),
('Ungkapan dan penjelasan', 'Merevisi ungkapan dan membedakan bentuk', '''
～というより（も）|lebih tepat disebut … daripada …|V普通形／Aい／Aな語幹／N + というより|Mengganti deskripsi awal dengan ungkapan yang lebih tepat.|Bukan selalu membandingkan tingkat; sering memperbaiki pilihan kata.|彼は友達というより、兄のような存在です。|かれは ともだちというより、あにの ような そんざいです。|Dia lebih seperti kakak daripada sekadar teman.|疲れたというより、眠いです。|つかれたというより、ねむいです。|Lebih tepat saya mengantuk daripada lelah.
～ことは～が|memang … tetapi …|V／A dengan predikat diulang + ことは + predikat + が|Mengakui hal tertentu lalu membatasi kesan atau hasilnya.|Ulangi predikat yang sama dengan waktu yang sesuai; bukan ことはない.|読んだことは読みましたが、よく覚えていません。|よんだ ことは よみましたが、よく おぼえていません。|Memang sudah saya baca, tetapi saya tidak terlalu ingat.|便利なことは便利ですが、値段が高いです。|べんりな ことは べんりですが、ねだんが たかいです。|Memang praktis, tetapi harganya mahal.
'''),
('Tanpa dan pengecualian', 'Tanpa melakukan dan tanpa unsur', '''
～ことなく|tanpa melakukan …|V辞書形 + ことなく|Cara formal menyatakan tindakan lain dilakukan tanpa tindakan pertama.|Gunakan bentuk kamus, bukan Vないことなく yang mengubah makna.|彼は休むことなく、働き続けました。|かれは やすむ ことなく、はたらきつづけました。|Dia terus bekerja tanpa beristirahat.|迷うことなく、答えを選びました。|まよう ことなく、こたえを えらびました。|Saya memilih jawaban tanpa ragu.
～抜きで／～抜きにして|tanpa menyertakan …|N + 抜きで／抜きにして|Menghilangkan suatu unsur yang biasanya ada atau mengesampingkan topik tertentu.|抜き dibaca ぬき; bukan hanya tanpa tindakan, melainkan tanpa unsur.|冗談抜きで、話してください。|じょうだんぬきで、はなしてください。|Tolong berbicara tanpa bercanda.|朝食抜きで、出かけました。|ちょうしょくぬきで、でかけました。|Saya keluar tanpa sarapan.
'''),
('Tanpa dan pengecualian', 'Tidak memandang dan kecuali', '''
～にかかわらず|tanpa memandang …|N + にかかわらず; pasangan V辞書形かVないか + にかかわらず|Hasil berlaku terlepas dari perbedaan kondisi seperti usia atau keikutsertaan.|Bedakan に関わる yang berarti berkaitan, dan にもかかわらず yang berarti meskipun.|年齢にかかわらず、参加できます。|ねんれいに かかわらず、さんかできます。|Siapa pun dapat ikut tanpa memandang usia.|参加するかしないかにかかわらず、返事をください。|さんかするか しないかに かかわらず、へんじを ください。|Terlepas dari ikut atau tidak, tolong beri jawaban.
～は別として|mengesampingkan …; kecuali …|N + は別として; klausa + かどうかは別として|Menunda pertimbangan satu hal untuk membicarakan aspek lain.|Bukan berarti hal itu tidak penting; hanya dikesampingkan dalam pembahasan ini.|値段は別として、良い製品です。|ねだんは べつとして、よい せいひんです。|Jika harga dikesampingkan, ini produk yang bagus.|成功するかどうかは別として、やってみたいです。|せいこうするかどうかは べつとして、やってみたいです。|Terlepas dari berhasil atau tidak, saya ingin mencoba.
'''),
('Kesempatan dan kelanjutan', 'Kesempatan yang terlewat dan sejak terakhir', '''
～ところだった|hampir saja … tetapi tidak jadi|V辞書形 + ところだった|Tindakan atau kejadian nyaris terjadi, tetapi akhirnya tidak; sering akibat buruk.|Berbeda dari Vたところだ yang berarti baru saja selesai.|もう少しで、電車に乗り遅れるところでした。|もうすこしで、でんしゃに のりおくれる ところでした。|Saya hampir saja ketinggalan kereta.|危うく、財布を忘れるところでした。|あやうく、さいふを わすれる ところでした。|Saya hampir saja lupa membawa dompet.
～たきり|sejak … tidak lagi …|Vた + きり + kelanjutan negatif atau tidak berubah|Sesudah kejadian terakhir, tindakan yang diharapkan tidak terjadi lagi.|Berbeda dari 二人きり yang berarti hanya berdua; contoh di sini fokus kelanjutan.|彼は出かけたきり、戻ってきません。|かれは でかけたきり、もどってきません。|Sejak keluar, dia belum kembali.|去年会ったきり、連絡を取っていません。|きょねん あったきり、れんらくを とっていません。|Sejak bertemu tahun lalu, kami tidak berhubungan lagi.
'''),
('Sebab dan tanggung jawab', 'Tekad dan sifat yang dikenal', '''
～上は|karena sudah … wajib …|V辞書形／Vた + 上は|Kewajiban atau tekad kuat setelah suatu pilihan; formal, dekat dengan 以上は.|Bukan 上で yang menandai prosedur; akhir biasanya tanggung jawab.|決めた上は、最後までやります。|きめた うえは、さいごまで やります。|Karena sudah memutuskan, saya akan melakukannya sampai akhir.|参加する上は、規則を守ります。|さんかする うえは、きそくを まもります。|Karena akan ikut, saya akan mematuhi aturan.
～ことだから|mengingat sifat … tentu …|N orang + のことだから|Menduga tindakan berdasarkan sifat orang yang sudah dikenal.|Berbeda dari ことから yang memakai fakta sebagai dasar kesimpulan.|慎重な彼のことだから、何度も確認したでしょう。|しんちょうな かれの ことだから、なんども かくにんしたでしょう。|Mengingat dia berhati-hati, pasti dia memeriksa berkali-kali.|親切な田中さんのことだから、助けてくれると思います。|しんせつな たなかさんの ことだから、たすけてくれると おもいます。|Mengingat Pak Tanaka baik hati, saya pikir dia akan membantu.
'''),
('Sebab dan penambahan', 'Pengumuman sebab dan informasi tambahan', '''
～につき：理由|karena … dalam pengumuman|N + につき|Alasan ringkas dan resmi dalam pemberitahuan, misalnya penutupan toko.|につき juga berarti per unit; lesson ini khusus penggunaan sebab.|工事中につき、通行できません。|こうじちゅうに つき、つうこうできません。|Karena sedang ada pekerjaan konstruksi, jalan tidak dapat dilalui.|満席につき、入場をお断りします。|まんせきに つき、にゅうじょうを おことわりします。|Karena kursi penuh, kami tidak menerima pengunjung lagi.
～上に|selain … juga …|V普通形／Aい + 上に; Aな + な上に; N + の上に／である上に|Menambahkan dua sifat atau fakta yang biasanya searah, baik sama-sama baik maupun buruk.|Bukan urutan seperti Vた上で; informasi kedua menambah bobot yang pertama.|この店は安い上に、おいしいです。|この みせは やすい うえに、おいしいです。|Restoran ini selain murah juga enak.|彼は親切な上に、説明も上手です。|かれは しんせつな うえに、せつめいも じょうずです。|Dia selain baik hati juga pandai menjelaskan.
'''),
('Kontras dan kesempatan', 'Dua sisi dan interupsi yang sopan', '''
～一方（で）|di satu sisi … di sisi lain …|V普通形／Aい + 一方で; Aな + な一方で; N + である一方で|Membandingkan dua sisi atau kelompok, atau menambahkan kegiatan paralel.|Berbeda dari V辞書形一方だ yang menyatakan perubahan satu arah terus-menerus.|都会は便利な一方で、生活費が高いです。|とかいは べんりな いっぽうで、せいかつひが たかいです。|Kota besar praktis, tetapi biaya hidupnya tinggi.|兄は料理が好きな一方で、弟は外食が好きです。|あには りょうりが すきな いっぽうで、おとうとは がいしょくが すきです。|Kakak suka memasak, sedangkan adik suka makan di luar.
～ところを|saat sedang … sebagai pengantar sopan|V普通形／Aい + ところを; Aな + なところを; Nの + ところを|Mengakui situasi lawan bicara ketika mengucapkan terima kasih atau meminta maaf atas interupsi.|ところを juga dapat menjadi objek penangkapan; konteks ini berfokus pada kesopanan.|お忙しいところを、ありがとうございます。|おいそがしい ところを、ありがとうございます。|Terima kasih telah meluangkan waktu di tengah kesibukan Anda.|お休みのところを、失礼しました。|おやすみの ところを、しつれいしました。|Maaf telah mengganggu saat Anda beristirahat.
'''),
('Penilaian dan risiko', 'Tidak ada kesempatan dan kekhawatiran', '''
～どころではない|sama sekali tidak sempat atau mampu …|V辞書形／N + どころではない|Keadaan terlalu sibuk atau serius sehingga kegiatan yang disebut tidak mungkin dipikirkan.|Berbeda dari どころか yang mengoreksi dugaan dengan fakta lain.|忙しくて、旅行どころではありません。|いそがしくて、りょこうどころでは ありません。|Saya terlalu sibuk untuk memikirkan bepergian.|熱があって、勉強するどころではありません。|ねつが あって、べんきょうするどころでは ありません。|Karena demam, saya sama sekali tidak sanggup belajar.
～おそれがある|ada risiko bahwa …|V辞書形／Vない + おそれがある; Nの + おそれがある|Kemungkinan buruk atau bahaya yang disampaikan secara objektif dan formal.|Bukan peluang baik; untuk hasil baik gunakan 可能性がある tanpa nada kekhawatiran.|大雨で、川が氾濫するおそれがあります。|おおあめで、かわが はんらんする おそれが あります。|Akibat hujan lebat, sungai berisiko meluap.|この方法では、間に合わないおそれがあります。|この ほうほうでは、まにあわない おそれが あります。|Dengan cara ini, ada risiko tidak selesai tepat waktu.
'''),
('Perasaan dan alasan', 'Intensitas sehari-hari dan alasan informal', '''
～てしょうがない／～てしかたがない|sangat … tidak bisa dihindari|Vて／Aい→くて／Aな→で + しょうがない／しかたがない|Sensasi atau perasaan yang kuat; しょうがない lebih percakapan.|Makna intensitas ini berbeda dari 仕方がない yang berdiri sendiri sebagai tidak ada pilihan.|眠くてしょうがありません。|ねむくて しょうがありません。|Saya sangat mengantuk.|故郷が懐かしくてしかたがありません。|ふるさとが なつかしくて しかたがありません。|Saya sangat merindukan kampung halaman.
～もの／～もん：理由|soalnya …|V普通形／Aい + もの／もん; Aな／N + だもの／だもん|Memberi alasan subjektif secara akrab, kadang bernada membela diri atau manja.|Bukan bentuk sopan untuk surat bisnis; もん lebih informal.|今日は行きたくないです。疲れているもの。|きょうは いきたくないです。つかれているもの。|Hari ini saya tidak ingin pergi. Soalnya lelah.|傘を持っていくよ。雨が降りそうだもん。|かさを もっていくよ。あめが ふりそうだもん。|Saya akan membawa payung. Soalnya kelihatannya akan hujan.
'''),
('Penolakan kuat', 'Penolakan emosional dan ketidakmungkinan lisan', '''
～ものか／～もんか|mana mungkin …; tidak akan …|V普通形／Aい + ものか; Aな／N + なものか|Penolakan emosional yang keras terhadap dugaan atau kemauan; informal.|Bukan pertanyaan netral; hindari kepada atasan dalam percakapan sopan.|二度とあの店へ行くものか。|にどと あの みせへ いくものか。|Saya tidak akan pergi ke restoran itu lagi.|こんなに難しい問題が分かるもんか。|こんなに むずかしい もんだいが わかるもんか。|Mana mungkin saya memahami soal sesulit ini.
～っこない|sama sekali tidak mungkin …|Vます語幹 + っこない|Penolakan kemungkinan dengan gaya percakapan yang tegas.|Gunakan pangkal ます: できっこない, bukan できるっこない.|こんな量は、一人で食べきれっこない。|こんな りょうは、ひとりで たべきれっこない。|Jumlah sebanyak ini tidak mungkin dihabiskan sendiri.|一度聞いただけでは、覚えっこない。|いちど きいただけでは、おぼえっこない。|Hanya mendengar sekali, tidak mungkin langsung hafal.
'''),
('Penilaian dan nasihat', 'Kepastian subjektif dan larangan umum', '''
～に決まっている|sudah pasti … menurut keyakinan pembicara|V普通形／Aい／Aな語幹／N + に決まっている|Keyakinan tegas dan subjektif, sering dalam percakapan; dapat tanpa bukti cukup.|Bukan selalu keputusan resmi; 決まっている di sini ialah ungkapan keyakinan.|そんな話は嘘に決まっています。|そんな はなしは うそに きまっています。|Cerita seperti itu pasti bohong.|毎日練習すれば、上達するに決まっています。|まいにち れんしゅうすれば、じょうたつするに きまっています。|Kalau berlatih setiap hari, tentu akan meningkat.
～ものではない|sebaiknya tidak … menurut norma umum|V辞書形 + ものではない|Nasihat atau larangan berdasarkan kepantasan dan kebiasaan umum.|Bukan larangan satu kali saja; untuk instruksi konkret gunakan ないでください.|人の失敗を笑うものではありません。|ひとの しっぱいを わらう ものでは ありません。|Tidak pantas menertawakan kegagalan orang.|約束を簡単に破るものではありません。|やくそくを かんたんに やぶる ものでは ありません。|Tidak semestinya begitu saja mengingkari janji.
'''),
('Perumpamaan dan pendahuluan', 'Seolah-olah dan sebelum tahap penting', '''
～かのようだ|seolah-olah … padahal belum tentu|V普通形／Aい + かのようだ; Aな／N + であるかのようだ|Membandingkan kesan dengan keadaan yang tidak benar-benar terjadi atau belum pasti.|Nominal formal memakai である; bukan hanya Nのかのようだ.|彼は何も知らないかのように話しました。|かれは なにも しらないかのように はなしました。|Dia berbicara seolah-olah tidak mengetahui apa pun.|まるで夢であるかのようでした。|まるで ゆめであるかのようでした。|Rasanya seolah-olah itu mimpi.
～に先立って|sebelum … sebagai tahap pendahuluan|N／V辞書形 + に先立って; N + に先立つ + N|Tindakan persiapan sebelum acara atau langkah penting; formal.|Bukan untuk setiap urutan kecil sehari-hari; 先立つ dibaca さきだつ.|試合に先立って、選手を紹介しました。|しあいに さきだって、せんしゅを しょうかいしました。|Sebelum pertandingan, para pemain diperkenalkan.|事業を始めるに先立って、調査を行いました。|じぎょうを はじめるに さきだって、ちょうさを おこないました。|Sebelum memulai usaha, kami melakukan riset.
'''),
('Cakupan dan prioritas', 'Tanpa batas kategori dan menunda pertimbangan', '''
～を問わず|tanpa membedakan …|N + を問わず|Kategori seperti usia, jenis, atau pengalaman tidak membatasi kelayakan; formal.|問わず dibaca とわず; bukan Nに問わず.|経験を問わず、応募できます。|けいけんを とわず、おうぼできます。|Anda dapat melamar tanpa memandang pengalaman.|昼夜を問わず、問い合わせを受け付けます。|ちゅうやを とわず、といあわせを うけつけます。|Kami menerima pertanyaan siang maupun malam.
～はともかく（として）|mengesampingkan dulu …|N + はともかく（として）; klausa + かはともかく|Memprioritaskan aspek lain dengan menunda pembahasan suatu hal.|Bukan selalu berarti hal pertama buruk; hanya bukan fokus sekarang.|結果はともかく、努力は評価したいです。|けっかは ともかく、どりょくは ひょうかしたいです。|Hasilnya dikesampingkan dulu, saya ingin menghargai usahanya.|値段はともかく、使いやすさを確認しましょう。|ねだんは ともかく、つかいやすさを かくにんしましょう。|Harga dikesampingkan dulu, mari periksa kemudahan pemakaiannya.
'''),
]

# The blank in each lesson tests the second pattern in an original sentence.
CLOZE = '''に際して
決まり次第
かと思うと
ところに
てからでないと
末に
読みつつ
上がる一方
ばかりに
だからこそ
からには
だけあって
にもかかわらず
思いつつ
とはいえ
どころか
ないことには
としても
話そうものなら
限りでは
日に限って
努力次第
につれて
に伴って
にわたって
を通じて
をめぐって
をもとに
にこたえて
に関わる
なわけではありません
わけにはいきません
はずがありません
に相違ありません
起こり得る
起こりかねません
読めないことはありません
確認しないではいられません
心配でなりません
にほかなりません
乗るまでもありません
のみならず
時間さえ
のわりに
疲れ気味
間違いだらけ
というものではありません
ことは読みましたが
冗談抜きで
は別として
たきり
のことだから
安い上に
ところを
おそれがあります
もの
食べきれっこない
ものではありません
に先立って
はともかく'''.splitlines()

def build():
    lessons = []
    for chapter, title, raw in CHAPTERS:
        points = []
        for row in raw.strip().splitlines():
            fields = row.split('|')
            assert len(fields) == 11, (title, len(fields))
            pattern, meaning, formation, explanation, mistake, *examples = fields
            points.append(dict(pattern=pattern, meaning=meaning, formation=formation,
                explanation=explanation, commonMistake=mistake,
                examples=[dict(japanese=examples[i], reading=examples[i+1], meaning=examples[i+2]) for i in (0, 3)]))
        assert len(points) == 2
        lessons.append(dict(chapter=chapter, title=title, points=points))
    pool = [p for l in lessons for p in l['points']]
    for index, lesson in enumerate(lessons):
        questions = []
        for p in lesson['points']:
            # Meaning recognition plus sentence comprehension and production.
            wrong = list(dict.fromkeys(x['meaning'] for x in pool if x['meaning'] != p['meaning']))
            questions.append(dict(prompt='Makna pola berikut dalam penggunaan yang diajarkan adalah apa?\n'+p['pattern'], choices=[p['meaning']]+wrong[(index*3)% (len(wrong)-3):][:3], correct=0, explanation=p['explanation']+'\nPembentukan: '+p['formation']))
        examples = [e for p in lesson['points'] for e in p['examples']]
        for p in lesson['points']:
            e = p['examples'][0]
            questions.append(dict(prompt='Pilih terjemahan yang paling sesuai:\n'+e['japanese'], choices=[e['meaning']]+[x['meaning'] for x in examples if x != e], correct=0, explanation=p['pattern']+': '+p['meaning']+'.\n'+e['reading']+'\n'+e['meaning']))
        p = lesson['points'][1]
        source = p['examples'][0]
        answer = CLOZE[index]
        assert answer in source['japanese'], (index+1, answer, source['japanese'])
        distractors = [x for x in ['に限らず', 'をもとに', 'わけがない', 'に際して'] if x != answer][:3]
        questions[1] = dict(prompt='Lengkapi kalimat agar sesuai maksud berikut:\n'+source['japanese'].replace(answer, '（　）', 1)+'\nMaksud: '+source['meaning'], choices=[answer]+distractors, correct=0, explanation=p['explanation']+'\nPembentukan: '+p['formation']+'\n'+source['reading'])
        e = p['examples'][1]
        questions.append(dict(prompt='Pilih kalimat yang sesuai maksud berikut:\n'+e['meaning'], choices=[e['japanese']]+[x['japanese'] for x in examples if x != e], correct=0, explanation=p['explanation']+'\n'+e['reading']+'\n'+e['meaning']))
        for qi, q in enumerate(questions):
            assert len(q['choices']) == len(set(q['choices'])) == 4
            offset = (index+qi)%4
            q['choices'] = q['choices'][offset:]+q['choices'][:offset]
            q['correct'] = (-offset)%4
        lesson['questions'] = questions
    assert len({p['pattern'] for p in pool}) == len(pool)
    Path('backend/data/grammar/n2.json').write_text(json.dumps(lessons, ensure_ascii=False, indent=2)+'\n', encoding='utf-8')
    rows = ['# Grammar N2', '', f'{len(lessons)} lesson, {len(pool)} unit pola, {len(pool)*2} contoh, {len(lessons)*5} soal.', '', 'Setiap pola berisi arti, pembentukan, penjelasan bahasa Indonesia, kesalahan umum, dan dua contoh dengan bacaan kana serta terjemahan. Kuis menguji arti pola, isian pola dalam kalimat, pemahaman kalimat, dan pemilihan kalimat sesuai maksud. Nilai minimal 80% menyelesaikan lesson; progress dan XP memakai layanan yang sama dengan level lain.', '', 'Kurikulum ini adalah materi pembelajaran asli yang disusun untuk N2, termasuk penguatan beberapa pola prasyarat. Cakupan mengikuti kebutuhan membaca beragam konteks pada [ringkasan kompetensi JLPT](https://www.jlpt.jp/e/about/levelsummary.html), bukan daftar pola resmi atau salinan soal resmi.', '', 'Jalankan `python scripts/build_grammar_n2.py` dari root untuk menghasilkan ulang konten deterministik.', '', '| Lesson | Tema | Judul | Pola |', '| --- | --- | --- | --- |']
    rows += [f"| {i:02d} | {l['chapter']} | {l['title']} | {'; '.join(p['pattern'] for p in l['points'])} |" for i,l in enumerate(lessons,1)]
    Path('docs/grammar-n2.md').write_text('\n'.join(rows)+'\n', encoding='utf-8')
    print(f'{len(lessons)} lessons, {len(pool)} patterns, {len(pool)*2} examples, {len(lessons)*5} questions')

if __name__ == '__main__':
    build()
