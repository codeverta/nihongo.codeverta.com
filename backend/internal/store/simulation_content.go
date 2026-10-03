package store

type simulationItem struct {
	section, group, instructions, passage, prompt string
	choices                                       [4]string // Correct answer first; randomized per attempt on the server.
	explanation                                   string
}

func simulationContent() []simulationItem {
	const reading = "＿＿の ことばは どう よみますか。１・２・３・４から いちばん よい ものを ひとつ えらんで ください。"
	const spelling = "＿＿の ことばを かんじや カタカナで かくとき、どれが ただしいですか。ひとつ えらんで ください。"
	const context = "（　）に はいる いちばん よい ことばを ひとつ えらんで ください。"
	const paraphrase = "もとの ぶんと いみが いちばん ちかい ぶんを ひとつ えらんで ください。"
	const grammar = "（　）に はいる いちばん よい ものを ひとつ えらんで ください。"
	const star = "４つの ことばを ならべて ぶんを つくって ください。★に はいる ことばを えらんで ください。"
	const text = "つぎの ぶんしょうを よんで、（　）に はいる いちばん よい ものを えらんで ください。"
	const passage = "わたしは アンナです。せんしゅう、あたらしい まちに 【35】。いま、えきの ちかくに すんでいます。まいあさ、バス【36】 がっこうへ いきます。がっこうで にほんごを べんきょうしています。クラスの ともだちは とても 【37】。きょうは どようびです。がっこうは やすみです。【38】、ともだちと こうえんへ いきます。"
	return []simulationItem{
		{"vocabulary", "文字・語彙 問題1 · 漢字の読み", reading, "", "きょうは [[水曜日]]です。", [4]string{"すいようび", "もくようび", "げつようび", "きんようび"}, "水曜日 dibaca すいようび (Rabu). 木曜日 adalah Kamis; 月曜日 Senin; 金曜日 Jumat."},
		{"vocabulary", "文字・語彙 問題1 · 漢字の読み", reading, "", "あねは まいにち [[電車]]で がっこうへ いきます。", [4]string{"でんしゃ", "てんしゃ", "でんじゃ", "でんしや"}, "電車 dibaca でんしゃ, kereta listrik. Huruf kecil ゃ membentuk bunyi しゃ; しや bukan bacaan yang sama."},
		{"vocabulary", "文字・語彙 問題1 · 漢字の読み", reading, "", "この かばんは [[安い]]です。", [4]string{"やすい", "たかい", "ひくい", "あかい"}, "安い dibaca やすい (murah). 高い = mahal/tinggi, 低い = rendah, 赤い = merah."},
		{"vocabulary", "文字・語彙 問題1 · 漢字の読み", reading, "", "わたしは [[毎朝]] パンを たべます。", [4]string{"まいあさ", "まいばん", "まいにち", "まいしゅう"}, "毎朝 dibaca まいあさ, setiap pagi. 毎晩 = setiap malam; 毎日 = setiap hari; 毎週 = setiap minggu."},
		{"vocabulary", "文字・語彙 問題1 · 漢字の読み", reading, "", "えきの まえで ともだちに [[会いました]]。", [4]string{"あいました", "かいました", "いいました", "おいました"}, "会う (あう) berarti bertemu. Bentuk lampau sopannya 会いました (あいました). 買いました = membeli; 言いました = mengatakan."},
		{"vocabulary", "文字・語彙 問題1 · 漢字の読み", reading, "", "この みせは [[午前]] 九じに あきます。", [4]string{"ごぜん", "ごご", "ごせん", "こぜん"}, "午前 dibaca ごぜん, sebelum tengah hari (AM). 午後 dibaca ごご (PM)."},
		{"vocabulary", "文字・語彙 問題1 · 漢字の読み", reading, "", "わたしの うちは [[東]]に あります。", [4]string{"ひがし", "にし", "みなみ", "きた"}, "東 dibaca ひがし (timur). 西 = にし (barat); 南 = みなみ (selatan); 北 = きた (utara)."},
		{"vocabulary", "文字・語彙 問題2 · 表記", spelling, "", "きのう、あたらしい [[ほん]]を かいました。", [4]string{"本", "木", "休", "体"}, "ほん (buku) ditulis 本. 木 = pohon; 休 berkaitan dengan istirahat; 体 = tubuh."},
		{"vocabulary", "文字・語彙 問題2 · 表記", spelling, "", "にわに [[しろい]] ねこが います。", [4]string{"白い", "百い", "自い", "目い"}, "しろい (putih) ditulis 白い. 百 berarti seratus, 自 berarti diri sendiri, 目 berarti mata."},
		{"vocabulary", "文字・語彙 問題2 · 表記", spelling, "", "つぎの かどを [[みぎ]]に まがって ください。", [4]string{"右", "左", "石", "名"}, "みぎ (kanan) ditulis 右. 左 adalah kiri. Perhatikan perbedaan bentuk 右 dan 石."},
		{"vocabulary", "文字・語彙 問題2 · 表記", spelling, "", "あした、[[がっこう]]へ いきます。", [4]string{"学校", "学交", "字校", "字交"}, "がっこう (sekolah) ditulis 学校. 学 menyatakan belajar dan 校 digunakan dalam kata sekolah."},
		{"vocabulary", "文字・語彙 問題2 · 表記", spelling, "", "あさ、[[こーひー]]を のみます。", [4]string{"コーヒー", "コーヒイ", "コヒー", "コーヒ"}, "こーひー ditulis コーヒー. Kedua vokal panjang ditandai dengan ー pada ejaan katakana yang benar."},
		{"vocabulary", "文字・語彙 問題3 · 文脈", context, "", "あめが ふっていますから、（　）を もって いきます。", [4]string{"かさ", "さいふ", "はし", "さら"}, "Karena hujan turun, benda yang diperlukan adalah かさ (payung). さいふ = dompet; はし = sumpit; さら = piring."},
		{"vocabulary", "文字・語彙 問題3 · 文脈", context, "", "この へやは くらいです。でんきを（　）ください。", [4]string{"つけて", "けして", "しめて", "ぬいで"}, "Ruangan gelap, maka nyalakan lampu: でんきをつけてください. けす berarti mematikan; しめる menutup; ぬぐ melepas pakaian."},
		{"vocabulary", "文字・語彙 問題3 · 文脈", context, "", "ここでは くつを（　）から、へやに はいって ください。", [4]string{"ぬいで", "きて", "かぶって", "はいて"}, "ぬいでから berarti setelah melepas. Dalam konteks masuk kamar, sepatu dilepas terlebih dahulu. はく = mengenakan sepatu/celana."},
		{"vocabulary", "文字・語彙 問題3 · 文脈", context, "", "はじめて あう ひとに『（　）。たなかです』と いいます。", [4]string{"はじめまして", "いただきます", "おかえりなさい", "ごちそうさまでした"}, "はじめまして digunakan saat berkenalan pertama kali. いただきます dan ごちそうさまでした terkait makan; おかえりなさい menyambut orang pulang."},
		{"vocabulary", "文字・語彙 問題3 · 文脈", context, "", "この みちは くるまが おおいです。わたる とき、（　）して ください。", [4]string{"ちゅうい", "さんぽ", "せんたく", "りょうり"}, "ちゅういする berarti berhati-hati/memperhatikan. Jalan ramai mobil memerlukan kehati-hatian ketika menyeberang."},
		{"vocabulary", "文字・語彙 問題3 · 文脈", context, "", "わたしは まいあさ はを（　）。", [4]string{"みがきます", "あらいます", "きります", "かきます"}, "Pasangan kata yang tepat adalah はをみがく, menyikat gigi. あらう = mencuci; きる = memotong; かく = menulis."},
		{"vocabulary", "文字・語彙 問題4 · 言い換え", paraphrase, "", "[[あの ひとは わたしの あにです。]]", [4]string{"あの ひとは わたしの おとうとでは ありません。わたしより としが うえです。", "あの ひとは わたしの おとうとです。", "あの ひとは わたしの ちちです。", "あの ひとは わたしの おじです。"}, "あに adalah kakak laki-laki sendiri: saudara yang lebih tua, bukan おとうと (adik laki-laki), ちち (ayah), atau おじ (paman)."},
		{"vocabulary", "文字・語彙 問題4 · 言い換え", paraphrase, "", "[[この かばんは あまり おもく ありません。]]", [4]string{"この かばんは かるいです。", "この かばんは とても おもいです。", "この かばんは おおきいです。", "この かばんは あたらしいです。"}, "あまりおもくありません berarti tidak begitu berat; padanan terdekat adalah かるい (ringan). Pilihan lain menyatakan berat, besar, atau baru."},
		{"vocabulary", "文字・語彙 問題4 · 言い換え", paraphrase, "", "[[きのうは しごとを しませんでした。]]", [4]string{"きのうは しごとを やすみました。", "きのうは しごとを はじめました。", "きのうは しごとが いそがしかったです。", "きのうは しごとに おくれました。"}, "Tidak bekerja kemarin paling dekat dengan しごとをやすみました (libur dari pekerjaan). はじめる = mulai; いそがしい = sibuk; おくれる = terlambat."},
		{"grammar", "文法 問題1 · 文の文法", grammar, "", "わたしは まいにち 七じ（　）おきます。", [4]string{"に", "で", "を", "が"}, "Partikel に menandai waktu tertentu: 七じにおきます, bangun pukul tujuh."},
		{"grammar", "文法 問題1 · 文の文法", grammar, "", "きのう、ともだち（　）えいがを みました。", [4]string{"と", "を", "が", "の"}, "と menunjukkan orang yang diajak melakukan kegiatan: ともだちとえいがをみました, menonton bersama teman."},
		{"grammar", "文法 問題1 · 文の文法", grammar, "", "この へやは あまり（　）。", [4]string{"ひろく ありません", "ひろい ありません", "ひろく です", "ひろい では ありません"}, "Negatif kata sifat い dibuat dengan mengganti い menjadi く: ひろくありません. あまり biasanya diikuti bentuk negatif."},
		{"grammar", "文法 問題1 · 文の文法", grammar, "", "きのうは とても（　）です。", [4]string{"さむかった", "さむい", "さむく", "さむくて"}, "きのう mengacu pada masa lampau. Kata sifat さむい menjadi さむかったです, dulu dingin."},
		{"grammar", "文法 問題1 · 文の文法", grammar, "", "ここで しゃしんを（　）も いいですか。", [4]string{"とって", "とる", "とった", "とり"}, "Izin menggunakan bentuk て + もいいですか. とる berubah menjadi とって: boleh mengambil foto di sini?"},
		{"grammar", "文法 問題1 · 文の文法", grammar, "", "きょうしつに がくせいが 三にん（　）。", [4]string{"います", "あります", "です", "します"}, "Keberadaan manusia memakai います; benda tidak hidup memakai あります. がくせいが三にんいます = ada tiga siswa."},
		{"grammar", "文法 問題1 · 文の文法", grammar, "", "あした、ほんを（　）に いきます。", [4]string{"かい", "かう", "かって", "かった"}, "Tujuan gerakan memakai batang ます + にいく. かいます → かいにいきます, pergi untuk membeli."},
		{"grammar", "文法 問題1 · 文の文法", grammar, "", "Ａ『いっしょに おちゃを のみませんか。』Ｂ『ええ、（　）。』", [4]string{"のみましょう", "のみませんでした", "のまないで ください", "のんで いません"}, "Jawaban menerima ajakan のみませんか adalah のみましょう, mari minum. Pilihan lain tidak menerima ajakan tersebut."},
		{"grammar", "文法 問題1 · 文の文法", grammar, "", "ごはんを たべる（　）、てを あらいます。", [4]string{"まえに", "あとで", "ながら", "まで"}, "Mencuci tangan sebelum makan: bentuk kamus + まえに. Untuk 'setelah makan' dibutuhkan たべたあとで, bukan たべるあとで."},
		{"grammar", "文法 問題2 · 文の組み立て", star, "", "これは ＿＿ ＿＿ ★ ＿＿ です。", [4]string{"かいた", "わたしが", "きのう", "てがみ"}, "Urutan: これは わたしが きのう かいた てがみです (ini surat yang saya tulis kemarin). ★ adalah slot ketiga, かいた."},
		{"grammar", "文法 問題2 · 文の組み立て", star, "", "わたしは ＿＿ ★ ＿＿ ＿＿ すんでいます。", [4]string{"ちかくに", "えきの", "ある", "アパートに"}, "Urutan: わたしは えきの ちかくに ある アパートに すんでいます. ★ adalah slot kedua, ちかくに. Artinya tinggal di apartemen yang ada di dekat stasiun."},
		{"grammar", "文法 問題2 · 文の組み立て", star, "", "あの ひとは ＿＿ ＿＿ ★ ＿＿ です。", [4]string{"いる", "にほんごを", "おしえて", "せんせい"}, "Urutan: あのひとは にほんごを おしえて いる せんせいです (orang itu adalah guru yang mengajar bahasa Jepang). ★ adalah slot ketiga, いる. おしえている menerangkan せんせい."},
		{"grammar", "文法 問題2 · 文の組み立て", star, "", "この へやは あついですから、まどを ＿＿ ★ ＿＿ ＿＿。", [4]string{"も", "あけて", "いい", "ですか"}, "Urutan: まどを あけて も いい ですか (bolehkah membuka jendela?). ★ adalah slot kedua, も. Pola bentuk て + もいいですか meminta izin."},
		{"grammar", "文法 問題3 · 文章の文法", text, passage, "【35】に はいる ものは どれですか。", [4]string{"きました", "きます", "きません", "きて ください"}, "せんしゅう (minggu lalu) membutuhkan bentuk lampau きました, telah datang ke kota baru."},
		{"grammar", "文法 問題3 · 文章の文法", text, passage, "【36】に はいる ものは どれですか。", [4]string{"で", "に", "を", "が"}, "バスで menandai sarana transportasi (dengan bus). がっこうへ menunjukkan arah tujuan."},
		{"grammar", "文法 問題3 · 文章の文法", text, passage, "【37】に はいる ものは どれですか。", [4]string{"しんせつです", "しんせつなです", "しんせつくです", "しんせつでです"}, "しんせつ adalah kata sifat な. Pada akhir pernyataan sopan dipakai しんせつです; な hanya dipakai sebelum kata benda."},
		{"grammar", "文法 問題3 · 文章の文法", text, passage, "【38】に はいる ものは どれですか。", [4]string{"ですから", "でも", "それから", "ところで"}, "Hubungan sebab-akibat: sekolah libur, karena itu pergi ke taman. ですから tepat; でも menyatakan pertentangan, それから urutan tambahan, ところで pergantian topik."},
	}
}
