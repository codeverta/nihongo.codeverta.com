"""Build the researched Indonesian JLPT N1 grammar curriculum."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "backend" / "data" / "grammar" / "n1.json"


def point(pattern, meaning, formation, explanation, mistake, examples):
    return {
        "pattern": pattern, "meaning": meaning, "formation": formation,
        "explanation": explanation, "commonMistake": mistake,
        "examples": [dict(zip(("japanese", "reading", "meaning"), e)) for e in examples],
    }


def p1(pattern, meaning, formation, explanation, mistake, j1, r1, m1, j2, r2, m2):
    return point(pattern, meaning, formation, explanation, mistake,
                 [(j1, r1, m1), (j2, r2, m2)])


LESSONS = [
    ("Waktu dan permulaan", "Kejadian yang langsung menyusul", [
        p1("～が早いか", "begitu … langsung …", "V辞書形 + が早いか",
           "Dua kejadian berlangsung hampir serentak; bagian kedua biasanya kejadian spontan yang diamati, bukan rencana pembicara.",
           "Gunakan bentuk kamus dan jangan menjadikannya janji tindakan yang disengaja.",
           "ベルが鳴るが早いか、学生たちは教室を飛び出した。", "べるが なるが はやいか、がくせいたちは きょうしつを とびだした。", "Begitu bel berbunyi, para siswa langsung berhamburan keluar kelas.",
           "彼は席に着くが早いか、資料を読み始めた。", "かれは せきに つくが はやいか、しりょうを よみはじめた。", "Begitu duduk, dia langsung mulai membaca dokumen."),
        p1("～や否や", "segera setelah …", "V辞書形 + や否や",
           "Bentuk formal atau tertulis yang menekankan bahwa kejadian kedua dimulai tepat setelah tindakan pertama.",
           "Klausa kedua umumnya menyatakan kejadian, bukan perintah atau kehendak pembicara.",
           "試験終了の合図があるや否や、受験者は答案を置いた。", "しけんしゅうりょうの あいずが あるやいなや、じゅけんしゃは とうあんを おいた。", "Begitu tanda ujian selesai diberikan, peserta meletakkan lembar jawabannya.",
           "窓を開けるや否や、冷たい風が吹き込んだ。", "まどを あけるやいなや、つめたい かぜが ふきこんだ。", "Begitu jendela dibuka, angin dingin menerpa masuk."),
    ]),
    ("Syarat dan ketergantungan", "Hasil yang bergantung pada keadaan", [
        p1("～いかんで（は）", "bergantung pada …", "N（の）+ いかんで（は）",
           "Menunjukkan bahwa keputusan atau hasil berubah menurut keadaan, hasil, atau isi hal yang disebutkan.",
           "Pola ini menyatakan hasil bergantung pada faktor itu; jangan samakan dengan ～にかかわらず yang berarti ‘tanpa memandang’.",
           "調査の結果いかんでは、計画を見直す必要がある。", "ちょうさの けっか いかんでは、けいかくを みなおす ひつようが ある。", "Bergantung pada hasil penyelidikan, rencana mungkin perlu ditinjau ulang.",
           "今後の対応いかんで、両社の関係は変わるだろう。", "こんごの たいおう いかんで、りょうしゃの かんけいは かわるだろう。", "Hubungan kedua perusahaan mungkin berubah bergantung pada penanganan selanjutnya."),
        p1("～いかんにかかわらず", "tanpa memandang …", "N（の）+ いかんにかかわらず",
           "Menyatakan bahwa keputusan atau tindakan tetap sama apa pun keadaan atau hasil yang disebutkan.",
           "Ini bukan ‘bergantung pada’; hasilnya tidak berubah oleh faktor tersebut.",
           "理由のいかんにかかわらず、無断欠席は認められません。", "りゆうの いかんに かかわらず、むだんけっせきは みとめられません。", "Apa pun alasannya, absen tanpa izin tidak diperbolehkan.",
           "経験の有無いかんにかかわらず、応募できます。", "けいけんの うむ いかんに かかわらず、おうぼできます。", "Anda boleh melamar terlepas dari ada atau tidaknya pengalaman."),
    ]),
    ("Syarat dan ketergantungan", "Fondasi yang mutlak diperlukan", [
        p1("～あっての", "ada … barulah … dapat terwujud", "N1 + あっての + N2",
           "Menekankan bahwa N2 hanya ada atau berhasil berkat N1; sering menyampaikan penghargaan atau pengakuan.",
           "Urutannya adalah fondasi lebih dahulu, lalu hal yang bergantung padanya.",
           "皆さんの支えあっての成功です。", "みなさんの ささえ あっての せいこうです。", "Keberhasilan ini terwujud berkat dukungan kalian.",
           "信頼あっての協力関係だと思います。", "しんらい あっての きょうりょくかんけいだと おもいます。", "Menurut saya, hubungan kerja sama ini hanya terwujud karena adanya kepercayaan."),
        p1("～なくして（は）", "tanpa …, … tidak mungkin", "N + なくして（は）",
           "Menegaskan bahwa hal yang disebutkan merupakan syarat penting bagi hasil berikut; lazim dalam tulisan dan pidato formal.",
           "Klausa sesudahnya biasanya menyatakan bahwa hasil tidak mungkin tercapai tanpanya.",
           "市民の協力なくして、この問題は解決できない。", "しみんの きょうりょく なくして、この もんだいは かいけつできない。", "Tanpa kerja sama warga, masalah ini tidak bisa diselesaikan.",
           "努力なくして、目標を達成することはできません。", "どりょく なくして、もくひょうを たっせいすることは できません。", "Tanpa usaha, tujuan tidak dapat dicapai."),
    ]),
    ("Sebab dan tujuan", "Alasan yang ditegaskan", [
        p1("～ばこそ", "justru karena …", "Vば／Aければ／N・Aなであれば + こそ",
           "Menegaskan sebab sebagai alasan kuat bagi tindakan atau hasil; sering bernada formal atau retoris.",
           "Bagian pertama harus berbentuk kondisional ば; jangan memakai bentuk kamus langsung.",
           "大切な友人であればこそ、率直に意見を伝えた。", "たいせつな ゆうじんで あればこそ、そっちょくに いけんを つたえた。", "Justru karena dia teman penting, saya menyampaikan pendapat dengan terus terang.",
           "何度も失敗すればこそ、成功の価値が分かる。", "なんども しっぱいすればこそ、せいこうの かちが わかる。", "Justru karena berulang kali gagal, kita memahami nilai keberhasilan."),
        p1("～とあって", "karena memang …, wajar jika …", "普通形 + とあって; N／Aな + とあって",
           "Menyatakan bahwa keadaan khusus menjadi alasan wajar bagi reaksi atau kejadian yang menyusul.",
           "Biasanya menghubungkan situasi khusus dengan respons yang dapat dipahami; bukan sebab netral untuk semua konteks.",
           "初めての海外公演とあって、会場は満席になった。", "はじめての かいがいこうえんと あって、かいじょうは まんせきに なった。", "Karena itu adalah pertunjukan luar negeri pertamanya, gedung dipenuhi penonton.",
           "有名な作家が来るとあって、書店の前に長い列ができた。", "ゆうめいな さっかが くると あって、しょてんの まえに ながい れつが できた。", "Karena penulis terkenal akan datang, antrean panjang terbentuk di depan toko buku."),
    ]),
    ("Tujuan dan kewajiban", "Tujuan formal dan keharusan", [
        p1("～べく", "untuk …; dengan tujuan …", "V辞書形（する → すべく）+ べく",
           "Bentuk formal untuk menyatakan tujuan atau tekad yang diikuti tindakan.",
           "する berubah menjadi すべく; jangan memakai するべく dalam bentuk standar pola ini.",
           "原因を究明すべく、専門家が現地に向かった。", "げんいんを きゅうめいすべく、せんもんかが げんちに むかった。", "Para ahli menuju lokasi untuk menyelidiki penyebabnya.",
           "合格すべく、毎日計画的に勉強している。", "ごうかくすべく、まいにち けいかくてきに べんきょうしている。", "Saya belajar terencana setiap hari demi lulus."),
        p1("～べからず", "dilarang …; jangan …", "V辞書形 + べからず（する → すべからず）",
           "Larangan keras dengan gaya tertulis, seperti pada peraturan, papan tanda, atau ungkapan tetap.",
           "Jangan tertukar dengan ～べきだ (‘seharusnya’); bentuk ini adalah larangan.",
           "関係者以外、立ち入るべからず。", "かんけいしゃ いがい、たちいるべからず。", "Selain pihak terkait, dilarang masuk.",
           "展示品に触れるべからず。", "てんじひんに ふれるべからず。", "Dilarang menyentuh barang pameran."),
    ]),
    ("Keharusan dan akibat", "Terpaksa dan tak terhindarkan", [
        p1("～を余儀なくされる", "terpaksa harus …", "N + を余儀なくされる",
           "Menyatakan bahwa keadaan memaksa seseorang atau organisasi melakukan sesuatu yang tidak diinginkan.",
           "Subjek yang dipaksa biasanya orang atau organisasi; nomina tindakan diikuti を.",
           "大雨のため、試合の中止を余儀なくされた。", "おおあめの ため、しあいの ちゅうしを よぎなくされた。", "Karena hujan lebat, pertandingan terpaksa dibatalkan.",
           "資金不足で計画の変更を余儀なくされた。", "しきんぶそくで けいかくの へんこうを よぎなくされた。", "Kekurangan dana membuat kami terpaksa mengubah rencana."),
        p1("～ずにはおかない", "pasti akan …; tidak mungkin tidak …", "Vない形（ない → ず）+ にはおかない",
           "Menekankan bahwa sesuatu pasti menimbulkan akibat atau reaksi tertentu, atau pelaku pasti akan bertindak demikian.",
           "Bentuk sebelum ず mengikuti perubahan negatif; する menjadi せず.",
           "彼の誠実な態度は、周囲の人々を感動させずにはおかない。", "かれの せいじつな たいどは、しゅういの ひとびとを かんどうさせずには おかない。", "Sikap tulusnya pasti akan menggerakkan orang-orang di sekitarnya.",
           "この映画は見る人に深い印象を与えずにはおかない。", "この えいがは みる ひとに ふかい いんしょうを あたえずには おかない。", "Film ini pasti meninggalkan kesan mendalam bagi penontonnya."),
    ]),
    ("Sikap dan perasaan", "Reaksi yang tak tertahankan", [
        p1("～を禁じ得ない", "tidak dapat menahan …", "N + を禁じ得ない",
           "Ungkapan formal untuk perasaan yang muncul kuat dan sulit ditahan.",
           "Diikuti nomina perasaan seperti 涙, 驚き, atau 同情; を menandai perasaan itu.",
           "被災地の映像を見て、涙を禁じ得なかった。", "ひさいちの えいぞうを みて、なみだを きんじえなかった。", "Melihat rekaman daerah bencana, saya tak kuasa menahan air mata.",
           "彼の無責任な発言に、怒りを禁じ得ない。", "かれの むせきにんな はつげんに、いかりを きんじえない。", "Saya tak dapat menahan amarah atas ucapannya yang tidak bertanggung jawab."),
        p1("～てやまない", "terus-menerus … dengan tulus", "Vて + やまない",
           "Menyatakan harapan, doa, atau perasaan yang terus dirasakan; sering dipakai dalam pidato dan tulisan formal.",
           "Umumnya dipakai dengan kata kerja perasaan atau harapan seperti 願う dan 尊敬する.",
           "皆様のご健康を願ってやみません。", "みなさまの ごけんこうを ねがってやみません。", "Saya sungguh terus mendoakan kesehatan Anda semua.",
           "彼女の努力を尊敬してやまない。", "かのじょの どりょくを そんけいしてやまない。", "Saya sangat terus menghormati usahanya."),
    ]),
    ("Derajat dan penekanan", "Derajat yang paling tinggi", [
        p1("～の極み", "puncak …; … yang tak terhingga", "N + の極み",
           "Menyatakan derajat tertinggi suatu perasaan atau keadaan; cenderung formal dan sering muncul dalam ungkapan tetap.",
           "Gunakan nomina sebelum 極み; jangan menambahkan な setelah kata sifat な.",
           "長年の夢がかなって、感激の極みです。", "ながねんの ゆめが かなって、かんげきの きわみです。", "Impian bertahun-tahun akhirnya terwujud; saya sangat terharu.",
           "このような賞をいただき、光栄の極みです。", "このような しょうを いただき、こうえいの きわみです。", "Menerima penghargaan seperti ini merupakan kehormatan yang luar biasa."),
        p1("～の至り", "sangat …; puncak …", "N + の至り",
           "Ungkapan sangat formal untuk derajat tertinggi, sering pada rasa syukur, malu, atau bangga.",
           "Biasanya melekat pada nomina seperti 光栄, 恐縮, atau 若気; nuansanya lebih resmi daripada percakapan sehari-hari.",
           "皆様にお会いできて、光栄の至りです。", "みなさまに おあいできて、こうえいの いたりです。", "Bisa bertemu Anda semua merupakan kehormatan besar bagi saya.",
           "私の説明不足でご迷惑をおかけし、恐縮の至りです。", "わたしの せつめいぶそくで ごめいわくを おかけし、きょうしゅくの いたりです。", "Saya sangat menyesal telah merepotkan Anda karena penjelasan saya kurang memadai."),
    ]),
    ("Derajat dan penilaian", "Perasaan yang sangat kuat", [
        p1("～限りだ", "sangat …; luar biasa …", "Aい + 限りだ; Aな語幹 + な限りだ",
           "Menyatakan perasaan pembicara yang sangat kuat, lazim dengan emosi seperti うれしい, 心強い, atau 残念.",
           "Gunakan bentuk kata sifat yang sesuai; pola ini menyampaikan perasaan subjektif, bukan batas kuantitas.",
           "家族が応援してくれるのは、心強い限りです。", "かぞくが おうえんしてくれるのは、こころづよい かぎりです。", "Dukungan keluarga sungguh membuat saya merasa tenang.",
           "長年の目標を達成できて、うれしい限りだ。", "ながねんの もくひょうを たっせいできて、うれしい かぎりだ。", "Saya sangat senang akhirnya mencapai tujuan yang lama saya cita-citakan."),
        p1("～に堪えない", "sangat …; tidak tertahankan", "N + に堪えない",
           "Pola formal yang memiliki dua penggunaan: perasaan sangat kuat (感謝に堪えない) atau sesuatu tak layak ditonton/didengar (見るに堪えない).",
           "Maknanya ditentukan oleh nomina: perasaan positif/negatif yang meluap atau penilaian bahwa sesuatu tidak pantas ditoleransi.",
           "長年のご支援に感謝の念に堪えません。", "ながねんの ごしえんに かんしゃの ねんに たえません。", "Saya sangat berterima kasih atas dukungan Anda selama ini.",
           "その映像はあまりにも残酷で、見るに堪えなかった。", "その えいぞうは あまりにも ざんこくで、みるに たえなかった。", "Rekaman itu begitu kejam sehingga tidak sanggup saya tonton."),
    ]),
    ("Kontras dan konsesi", "Mengakui fakta sambil membatasi", [
        p1("～といえども", "meskipun …; sekalipun …", "N／普通形 + といえども",
           "Mengakui status atau keadaan pertama, lalu menyatakan bahwa hal itu tidak mengubah batasan atau kenyataan berikutnya; bernada formal.",
           "Klausa setelahnya sering menyampaikan pengecualian atau batas; bukan sekadar penghubung ‘tetapi’ netral.",
           "専門家といえども、すべてを知っているわけではない。", "せんもんかと いえども、すべてを しっている わけではない。", "Meskipun seorang ahli, bukan berarti ia mengetahui segalanya.",
           "親といえども、子どもの考えを決めつけてはいけない。", "おやと いえども、こどもの かんがえを きめつけては いけない。", "Sekalipun orang tua, kita tidak boleh menghakimi pikiran anak."),
        p1("～はおろか", "jangankan …, bahkan …", "N + はおろか",
           "Menunjukkan bahwa hal yang lebih dasar atau diharapkan pun tidak tercapai, apalagi hal yang lebih jauh.",
           "Sering diikuti hasil negatif; pastikan hubungan tingkat antara dua hal jelas.",
           "忙しくて、昼食はおろか水を飲む時間もなかった。", "いそがしくて、ちゅうしょくは おろか みずを のむ じかんも なかった。", "Saya begitu sibuk sehingga jangankan makan siang, waktu untuk minum air pun tidak ada.",
           "彼は漢字はおろか、ひらがなもまだ読めない。", "かれは かんじは おろか、ひらがなも まだ よめない。", "Dia jangankan kanji, hiragana pun masih belum bisa dibaca."),
    ]),
    ("Perbandingan dan perubahan", "Kontras dan peningkatan", [
        p1("～にひきかえ", "berbeda sekali dengan …", "N／普通形 + にひきかえ",
           "Membandingkan dua hal yang kontras; sering menonjolkan perbedaan yang mencolok.",
           "Bagian kedua harus dapat dibandingkan dengan yang pertama; jangan gunakan untuk sebab-akibat.",
           "兄が社交的なのにひきかえ、弟は一人で過ごすのが好きだ。", "あにが しゃこうてきなのに ひきかえ、おとうとは ひとりで すごすのが すきだ。", "Berbeda dengan kakaknya yang supel, adiknya suka menghabiskan waktu sendirian.",
           "去年の好景気にひきかえ、今年は売上が伸び悩んでいる。", "きょねんの こうけいきに ひきかえ、ことしは うりあげが のびなやんでいる。", "Berbeda dengan ekonomi yang baik tahun lalu, penjualan tahun ini sulit meningkat."),
        p1("～にもまして", "lebih … daripada …; bahkan lebih …", "N + にもまして; 疑問詞 + にもまして",
           "Menekankan bahwa suatu hal lebih kuat atau penting daripada pembanding, termasuk keadaan sebelumnya.",
           "Sering menggunakan pembanding seperti 以前 atau 何にも; pastikan hal yang dibandingkan sejenis.",
           "今年は例年にもまして暑い日が続いている。", "ことしは れいねんにも まして あつい ひが つづいている。", "Tahun ini hari-hari panas berlangsung lebih banyak daripada tahun-tahun biasanya.",
           "何にもまして、健康が大切だ。", "なににも まして、けんこうが たいせつだ。", "Lebih dari apa pun, kesehatan itu penting."),
    ]),
    ("Lingkup dan urutan", "Dari awal hingga rentang luas", [
        p1("～を皮切りに", "dimulai dengan … lalu berlanjut", "N + を皮切りに（して）",
           "Menyatakan bahwa suatu rangkaian dimulai dari peristiwa atau tempat pertama, lalu berkembang ke yang lain.",
           "Biasanya ada kejadian lanjutan setelah titik awal; bukan sekadar menyebut satu peristiwa.",
           "東京公演を皮切りに、全国ツアーが始まった。", "とうきょうこうえんを かわきりに、ぜんこくつあーが はじまった。", "Tur nasional dimulai dengan pertunjukan di Tokyo.",
           "この店は京都を皮切りに、各地へ出店した。", "この みせは きょうとを かわきりに、かくちへ しゅってんした。", "Toko ini mulai membuka cabang di berbagai daerah, diawali dari Kyoto."),
        p1("～に至るまで", "hingga …; bahkan sampai …", "N + に至るまで",
           "Menunjukkan cakupan yang mencapai titik jauh atau terperinci; sering memakai pasangan から untuk awal rentang.",
           "Menekankan luasnya cakupan, bukan hanya batas akhir waktu.",
           "子どもから高齢者に至るまで、多くの人が参加した。", "こどもから こうれいしゃに いたるまで、おおくの ひとが さんかした。", "Banyak orang ikut serta, dari anak-anak hingga lansia.",
           "計画の細部に至るまで、丁寧に説明された。", "けいかくの さいぶに いたるまで、ていねいに せつめいされた。", "Segala hal dijelaskan dengan teliti sampai ke rincian rencana."),
    ]),
    ("Dasar dan sudut pandang", "Dalih dan pilihan yang tiada banding", [
        p1("～名のもとに", "atas nama …", "N + 名のもとに／名のもとで",
           "Menyebut nama, tujuan, atau prinsip yang digunakan sebagai dasar tindakan; kadang bernuansa kritis bila dasar itu hanya dalih.",
           "Bentuknya 名のもとに, dengan の wajib; konteks menentukan apakah netral atau kritis.",
           "安全確保の名のもとに、厳しい規則が導入された。", "あんぜんかくほの な の もとに、きびしい きそくが どうにゅうされた。", "Atas nama menjaga keamanan, peraturan ketat diberlakukan.",
           "彼らは改革の名のもとに、組織の制度を変えた。", "かれらは かいかくの な の もとに、そしきの せいどを かえた。", "Atas nama reformasi, mereka mengubah sistem organisasi."),
        p1("～をおいて（ほかに）", "selain … tidak ada yang …", "N + をおいて（ほかに）～ない",
           "Menegaskan bahwa hanya orang atau hal yang disebutkan yang cocok atau mampu melakukan sesuatu.",
           "Sering diikuti bentuk negatif seperti ほかにない atau いない.",
           "この難題を解決できる人は、彼をおいてほかにいない。", "この なんだいを かいけつできる ひとは、かれを おいて ほかに いない。", "Tidak ada orang selain dia yang dapat menyelesaikan masalah sulit ini.",
           "次の代表には、経験豊富な彼女をおいて適任者はいない。", "つぎの だいひょうには、けいけん ほうふな かのじょを おいて てきにんしゃは いない。", "Untuk menjadi wakil berikutnya, tidak ada orang yang lebih tepat daripada dia yang berpengalaman."),
    ]),
    ("Gaya tulis dan perumpamaan", "Bentuk klasik dalam tulisan", [
        p1("～ごとく", "seperti …; sebagaimana …", "Nの／V普通形 + ごとく; ごとき + N",
           "Bentuk bergaya klasik untuk membuat perumpamaan atau menyatakan cara; dalam percakapan modern sering digantikan ように.",
           "ごとく berfungsi seperti keterangan; sebelum nomina gunakan ごとき.",
           "彼は風のごとく、あっという間に走り去った。", "かれは かぜの ごとく、あっというまに はしりさった。", "Dia berlari pergi secepat angin.",
           "前に述べたごとく、調査は来月から始まる。", "まえに のべた ごとく、ちょうさは らいげつから はじまる。", "Sebagaimana telah disebutkan sebelumnya, penyelidikan dimulai bulan depan."),
        p1("～ごとき", "seperti …; orang/hal semacam …", "Nの／V普通形 + ごとき + N; N + ごとき",
           "Bentuk klasik sebelum kata benda untuk perumpamaan; bila merujuk diri sendiri atau hal kecil, dapat merendahkan diri atau mengecilkan.",
           "Nuansa sangat bergantung pada konteks; ごとき tidak selalu penghinaan.",
           "彼のごとき経験豊富な人に相談したい。", "かれの ごとき けいけん ほうふな ひとに そうだんしたい。", "Saya ingin berkonsultasi dengan orang berpengalaman seperti dia.",
           "私ごときに務まるか分かりませんが、精一杯努めます。", "わたしごときに つとまるか わかりませんが、せいいっぱい つとめます。", "Saya tidak tahu apakah orang seperti saya sanggup, tetapi saya akan berusaha sebaik mungkin."),
    ]),
    ("Tujuan dan keadaan", "Tindakan demi tujuan dan pengandaian khusus", [
        p1("～んがため（に）", "demi …; dengan tujuan untuk …", "Vない形（ない → ん）+ がため（に）; する → せんがため",
           "Ungkapan sangat formal atau bergaya sastra untuk tujuan yang kuat; biasanya menunjukkan tekad.",
           "Perubahan bentuknya khusus: 認められんがため, 成功せんがため; bukan bentuk kamus + ん.",
           "彼は家族を守らんがため、危険な任務を引き受けた。", "かれは かぞくを まもらんがため、きけんな にんむを ひきうけた。", "Demi melindungi keluarga, dia menerima tugas berbahaya.",
           "真実を明らかにせんがため、記者は取材を続けた。", "しんじつを あきらかに せんがため、きしゃは しゅざいを つづけた。", "Demi mengungkap kebenaran, wartawan terus melakukan peliputan."),
        p1("～とあれば", "kalau memang demi …; jika benar …", "N／普通形 + とあれば",
           "Menunjukkan bahwa jika keadaan atau alasan khusus itu berlaku, pembicara bersedia melakukan sesuatu yang biasanya tidak dilakukan.",
           "Sering diikuti tekad atau tindakan khusus; jangan samakan dengan としたら yang sekadar membayangkan keadaan.",
           "子どものためとあれば、多少の苦労はいとわない。", "こどもの ためと あれば、たしょうの くろうは いとわない。", "Kalau demi anak, saya tidak keberatan menghadapi sedikit kesulitan.",
           "必要とあれば、いつでもお手伝いします。", "ひつようと あれば、いつでも おてつだいします。", "Jika memang diperlukan, saya siap membantu kapan saja."),
    ]),
    ("Kebiasaan dan kesempatan", "Sambil melakukan hal lain dan sekaligus", [
        p1("～かたわら", "sambil …; selain … juga …", "V辞書形 + かたわら; Nの + かたわら",
           "Menggambarkan seseorang menjalankan kegiatan utama sekaligus kegiatan lain secara berkelanjutan.",
           "Kedua kegiatan biasanya berlangsung dalam jangka panjang; untuk tindakan sesaat gunakan pola lain.",
           "彼女は大学で教えるかたわら、研究書も執筆している。", "かのじょは だいがくで おしえる かたわら、けんきゅうしょも しっぴつしている。", "Selain mengajar di universitas, dia juga menulis buku penelitian.",
           "父は会社勤めのかたわら、地域の活動にも参加している。", "ちちは かいしゃづとめの かたわら、ちいきの かつどうにも さんかしている。", "Selain bekerja di perusahaan, ayah juga ikut kegiatan masyarakat."),
        p1("～かたがた", "sekalian …; sambil juga …", "N + かたがた",
           "Menyatakan dua tujuan dalam satu kunjungan atau tindakan, sering dengan tujuan sosial yang sopan seperti berkunjung atau menyampaikan terima kasih.",
           "Biasanya mengikuti nomina tujuan seperti お礼, 報告, 挨拶; gaya formal dan bukan untuk setiap kegiatan.",
           "お礼かたがた、先生のお宅を訪ねた。", "おれい かたがた、せんせいの おたくを たずねた。", "Saya mengunjungi rumah guru sekaligus untuk menyampaikan terima kasih.",
           "近況報告かたがた、旧友に手紙を書いた。", "きんきょうほうこく かたがた、きゅうゆうに てがみを かいた。", "Saya menulis surat kepada teman lama sekaligus untuk mengabarkan keadaan terbaru."),
    ]),
]


def build_questions(points, gloss_pool):
    all_examples = [(p["examples"][i], p) for p in points for i in range(2)]
    questions = []
    for idx, p in enumerate(points):
        other = points[1 - idx]
        # Distractors come from the paired point and fixed, semantically distinct glosses.
        distractors = [other["meaning"], *gloss_pool, "makna yang tidak berkaitan"]
        unique_distractors = []
        for gloss in distractors:
            if gloss != p["meaning"] and gloss not in unique_distractors:
                unique_distractors.append(gloss)
            if len(unique_distractors) == 3:
                break
        questions.append({"prompt": f"Apa makna pola berikut?\n{p['pattern']}",
                          "choices": [p["meaning"], *unique_distractors],
                          "correct": 0,
                          "explanation": f"{p['explanation']}\nPembentukan: {p['formation']}"})
    for idx, (example, p) in enumerate(all_examples):
        others = [e for e, owner in all_examples if owner is not p or e is not example]
        choices = [example["meaning"]] + [e["meaning"] for e in others[:3]]
        questions.append({"prompt": f"Pilih terjemahan yang sesuai:\n{example['japanese']}",
                          "choices": choices, "correct": 0,
                          "explanation": f"{p['pattern']}: {p['meaning']}\n{example['reading']}\n{example['meaning']}"})
        if idx == 1:
            break
    # Test sentence selection for each point, with one correct Japanese sentence and three alternatives.
    correct_example = points[0]["examples"][0]
    distractors = [e for e, _ in all_examples if e["japanese"] != correct_example["japanese"]]
    chosen = []
    for e in distractors:
        if e["japanese"] != correct_example["japanese"] and e["japanese"] not in [x["japanese"] for x in chosen]:
            chosen.append(e)
        if len(chosen) == 3:
            break
    questions.append({"prompt": f"Pilih kalimat yang berarti: {correct_example['meaning']}",
                      "choices": [correct_example["japanese"], *[e["japanese"] for e in chosen]],
                      "correct": 0,
                      "explanation": f"{points[0]['pattern']} — {correct_example['reading']}\n{correct_example['meaning']}"})
    return questions


def main():
    gloss_pool = [point["meaning"] for _, _, points in LESSONS for point in points]
    lessons = []
    for chapter, title, points in LESSONS:
        lessons.append({"chapter": chapter, "title": title, "points": points,
                        "questions": build_questions(points, gloss_pool)})
    OUT.write_text(json.dumps(lessons, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Wrote {len(lessons)} lessons, {sum(len(x['points']) for x in lessons)} patterns to {OUT}")


if __name__ == "__main__":
    main()
