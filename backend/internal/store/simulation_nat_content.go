package store

// Original NAT 5Q written practice: 21 vocabulary + 17 grammar + 5 reading.
func simulationNATContent() []simulationItem {
	type question struct {
		prompt      string
		choices     [4]string
		explanation string
	}
	items := make([]simulationItem, 0, 43)
	add := func(section, group, instructions, passage string, questions []question) {
		for _, q := range questions {
			items = append(items, simulationItem{section, group, instructions, passage, q.prompt, q.choices, q.explanation})
		}
	}
	add("vocabulary", "文字・語彙 問題1 · Bacaan kanji", "＿＿の言葉はどう読みますか。いちばんよいものを一つ選んでください。", "", []question{
		{"駅で[[友達]]に会います。", [4]string{"ともだち", "とまだち", "ゆうだち", "ともたち"}, "友達（ともだち） berarti teman. 会います = bertemu."},
		{"この[[電車]]は東京へ行きます。", [4]string{"でんしゃ", "てんしゃ", "でんじゃ", "でんしや"}, "電車（でんしゃ） adalah kereta. しゃ ditulis dengan ゃ kecil."},
		{"[[毎朝]]、パンを食べます。", [4]string{"まいあさ", "まいひる", "まいばん", "まいにち"}, "毎朝（まいあさ） = setiap pagi. 毎晩 = setiap malam, 毎日 = setiap hari."},
		{"[[右]]に小さい店があります。", [4]string{"みぎ", "ひだり", "うえ", "した"}, "右（みぎ） = kanan. 左 = kiri, 上 = atas, 下 = bawah."},
		{"この本は[[高い]]です。", [4]string{"たかい", "やすい", "ながい", "ひろい"}, "高い（たかい） dalam konteks harga berarti mahal. 安い = murah."},
		{"教室に学生が[[十人]]います。", [4]string{"じゅうにん", "じゅうひと", "とおひと", "じゅにん"}, "十人 dibaca じゅうにん, sepuluh orang. Bunyi じゅう panjang."},
		{"今日は[[火曜日]]です。", [4]string{"かようび", "すいようび", "もくようび", "きんようび"}, "火曜日（かようび） = Selasa. 水曜日 Rabu, 木曜日 Kamis, 金曜日 Jumat."},
	})
	add("vocabulary", "文字・語彙 問題2 · Penulisan", "＿＿の言葉を漢字で書くとき、いちばんよいものを一つ選んでください。", "", []question{
		{"ここに[[なまえ]]を書いてください。", [4]string{"名前", "名剪", "各前", "各剪"}, "名前（なまえ） berarti nama. Kanji yang tepat adalah 名 dan 前."},
		{"[[みず]]を飲みます。", [4]string{"水", "木", "火", "氷"}, "水（みず） = air. 木 = pohon, 火 = api, 氷 = es."},
		{"学校は駅の[[きた]]にあります。", [4]string{"北", "南", "東", "西"}, "北（きた） = utara. 南 selatan, 東 timur, 西 barat."},
		{"[[あたらしい]]くつを買いました。", [4]string{"新しい", "親しい", "薪しい", "折しい"}, "新しい（あたらしい） = baru. 親 memakai kanji yang berbeda."},
		{"きのう、手紙を[[かきました]]。", [4]string{"書きました", "聞きました", "読きました", "言きました"}, "書く（かく） = menulis. 書きました adalah bentuk lampau sopan."},
	})
	add("vocabulary", "文字・語彙 問題3 · Konteks", "（　）に入るいちばんよい言葉を一つ選んでください。", "", []question{
		{"この部屋は（　）です。電気をつけましょう。", [4]string{"暗い", "甘い", "丸い", "若い"}, "暗い（くらい） = gelap. Menyalakan lampu sesuai keadaan ruangan gelap."},
		{"朝ご飯を食べる前に、顔を（　）。", [4]string{"洗います", "読みます", "借ります", "乗ります"}, "顔を洗う = mencuci muka. 読む membaca, 借りる meminjam, 乗る naik kendaraan."},
		{"雨が降っているので、（　）を持っていきます。", [4]string{"かさ", "さら", "まくら", "はし"}, "かさ = payung untuk hujan. さら piring, まくら bantal, はし sumpit."},
		{"このかばんは（　）です。中に何も入っていません。", [4]string{"軽い", "重い", "辛い", "寒い"}, "軽い（かるい） = ringan. Tas kosong mendukung pilihan ini; 重い berarti berat."},
		{"牛乳は（　）に入れてください。冷たくしておきたいです。", [4]string{"冷蔵庫", "本棚", "机の引き出し", "くつ箱"}, "冷蔵庫（れいぞうこ） = kulkas untuk menjaga susu tetap dingin."},
		{"きのうの夜はたくさん寝ました。今日はあまり（　）ありません。", [4]string{"眠く", "赤く", "広く", "細く"}, "眠い（ねむい） = mengantuk. Bentuk negatif sopan 眠くありません."},
	})
	add("vocabulary", "文字・語彙 問題4 · Padanan makna", "＿＿の言葉と意味が近いものを一つ選んでください。", "", []question{
		{"日曜日は[[仕事をしません]]。", [4]string{"仕事を休みます", "仕事を始めます", "仕事を教えます", "仕事を探します"}, "仕事をしない dan 仕事を休む sama-sama menyatakan tidak bekerja pada hari tersebut."},
		{"駅はここから[[近い]]です。", [4]string{"駅まであまり時間がかかりません", "駅までとても時間がかかります", "駅は今日休みです", "駅に電車がありません"}, "近い = dekat; perjalanan ke stasiun tidak memakan banyak waktu."},
		{"この店には[[いろいろな]]パンがあります。", [4]string{"たくさんの種類の", "一つの種類の", "昨日と同じだけの", "とても高い"}, "いろいろな = berbagai macam; menekankan banyak jenis, bukan harga atau jumlah saja."},
	})
	add("grammar", "文法 問題1 · Pilihan grammar", "（　）に入るいちばんよいものを一つ選んでください。", "", []question{
		{"日曜日（　）友達と映画を見ます。", [4]string{"に", "を", "で", "が"}, "に menandai waktu tertentu: 日曜日に = pada hari Minggu."},
		{"この部屋（　）勉強します。", [4]string{"で", "へ", "を", "の"}, "で menandai tempat melakukan kegiatan, yaitu belajar di ruangan ini."},
		{"私は兄（　）二人で住んでいます。", [4]string{"と", "から", "まで", "を"}, "と = bersama. 兄と二人で berarti tinggal berdua dengan kakak laki-laki."},
		{"昨日は雨（　）。", [4]string{"でした", "です", "でありますか", "ではありませんか"}, "昨日 = kemarin, sehingga kalimat pernyataan memakai bentuk lampau 雨でした."},
		{"この映画はあまり（　）。", [4]string{"おもしろくなかったです", "おもしろかったです", "おもしろいでした", "おもしろくでした"}, "あまり berpasangan dengan bentuk negatif: tidak terlalu menarik. Adjektiva い negatif lampau menjadi ～くなかったです."},
		{"A：もう宿題をしましたか。\nB：いいえ、まだ（　）。", [4]string{"していません", "しました", "しますか", "しませんでしたか"}, "まだしていません = belum mengerjakan. Sesuai jawaban いいえ terhadap pertanyaan もう."},
		{"朝ご飯を（　）から、学校へ行きます。", [4]string{"食べて", "食べたらて", "食べます", "食べるて"}, "Vてから = setelah melakukan. 食べてから berarti setelah makan."},
		{"ここでたばこを（　）いけません。", [4]string{"吸っては", "吸っても", "吸うとき", "吸いたくて"}, "Vてはいけません menyatakan larangan. 吸ってはいけません = dilarang merokok."},
		{"A：りんごはいくつありますか。\nB：三つ（　）。", [4]string{"あります", "います", "ですか", "でしたか"}, "Benda mati memakai ある. 三つあります = ada tiga buah."},
	})
	add("grammar", "文法 問題2 · Susunan kalimat", "＿＿に言葉を入れて文を完成させてください。★に入るものを一つ選んでください。", "", []question{
		{"この ＿＿ ＿＿ ★ ＿＿ です。", [4]string{"駅の", "店", "近く", "は"}, "Urutan: 店／は／駅の／近く. この店は駅の近くです = toko ini dekat stasiun. ★ slot ketiga adalah 駅の."},
		{"私は ＿＿ ★ ＿＿ ＿＿ あります。", [4]string{"へ", "日本", "行った", "ことが"}, "Urutan: 日本／へ／行った／ことが. 日本へ行ったことがあります = pernah pergi ke Jepang. ★ slot kedua adalah へ."},
		{"寝る ＿＿ ＿＿ ★ ＿＿ ください。", [4]string{"を", "前に", "電気", "消して"}, "Urutan: 前に／電気／を／消して. Sebelum tidur, matikan lampu. ★ slot ketiga adalah を."},
		{"これは ＿＿ ＿＿ ＿＿ ★ かばんです。", [4]string{"新しい", "昨日", "買った", "私が"}, "Urutan: 私が／昨日／買った／新しい. Ini tas baru yang saya beli kemarin. ★ slot keempat adalah 新しい."},
	})
	text := "私は毎日、自転車で学校へ行きます。でも、今日は朝から雨が【1】。それで、バスで行くことにしました。バス停まで歩いて【2】分かかります。\nバスの中には人がたくさんいました。私は座らないで、立っていました。学校に着いたとき、雨は【3】降っていました。かさを持ってきてよかったです。\n授業は九時からです。今日は八時五十分に着いたので、授業の前に友達と少し【4】。"
	add("grammar", "文法 問題3 · Grammar teks", "文章を読んで、【1】から【4】に入るいちばんよいものを選んでください。", text, []question{
		{"【1】に入るものはどれですか。", [4]string{"降っています", "降ってください", "降りましょう", "降りたいです"}, "雨が降っています = hujan sedang turun. Ini keadaan cuaca, bukan permintaan atau keinginan."},
		{"【2】に入るものはどれですか。", [4]string{"五", "五時", "五人", "五つ"}, "五分 = lima menit. Angka langsung digabung dengan satuan 分; 五時分 bukan durasi yang benar."},
		{"【3】に入るものはどれですか。", [4]string{"まだ", "もうすぐ", "いつもは", "初めての"}, "まだ降っていました = masih turun saat tiba. Hujan berlangsung sejak pagi."},
		{"【4】に入るものはどれですか。", [4]string{"話しました", "話してください", "話しませんか", "話したいでした"}, "Penulis menceritakan kejadian pagi tadi: 話しました = sempat berbicara. Pilihan lain berupa permintaan atau ajakan."},
	})
	add("grammar", "読解 問題4 · Bacaan pendek", "次の文章を読んで、質問に答えてください。", "山田さんへ\n明日は一緒に図書館へ行きましょう。図書館は十時に開きます。九時半に駅の前で会いませんか。私は赤いかばんを持っていきます。\nリン", []question{
		{"山田さんとリンさんは、何時にどこで会いますか。", [4]string{"九時半に駅の前で会います。", "十時に図書館の中で会います。", "九時に駅の中で会います。", "十時半に店の前で会います。"}, "Ajakan bertemu berbunyi 九時半に駅の前. Pukul sepuluh adalah waktu perpustakaan buka."},
	})
	add("grammar", "読解 問題4 · Bacaan pendek", "次の文章を読んで、質問に答えてください。", "私は小さいアパートに住んでいます。学校からは遠いですが、駅から歩いて三分です。近くに安いスーパーもあります。部屋は少し狭いですが、生活は便利です。", []question{
		{"このアパートについて、正しいものはどれですか。", [4]string{"駅が近くて、近くにスーパーもあります。", "学校からとても近いです。", "部屋はとても広いです。", "駅まで歩いて三十分です。"}, "駅徒歩三分 dan dekat supermarket membuat lokasi praktis. Sekolah jauh dan ruangan sempit."},
	})
	add("grammar", "読解 問題5 · Bacaan sedang", "次の文章を読んで、質問に答えてください。", "先週の日曜日、ミナさんと公園へ行きました。朝は晴れていました。私たちは公園で写真を撮って、お弁当を食べました。午後、雨が降り始めました。私はかさを持っていませんでしたが、ミナさんは大きいかさを持っていました。二人でそのかさを使って、駅まで歩きました。その後、電車で家へ帰りました。", []question{
		{"雨が降り始める前に、何をしましたか。", [4]string{"写真を撮って、お弁当を食べました。", "電車で家へ帰りました。", "駅でかさを買いました。", "公園へバスで行きました。"}, "Foto dan makan bekal terjadi saat pagi cerah, sebelum hujan sore hari."},
		{"駅まで、どうやって行きましたか。", [4]string{"ミナさんのかさを二人で使って歩きました。", "一人で走りました。", "新しいかさを買ってバスに乗りました。", "電車で行きました。"}, "Penulis dan Mina berjalan berbagi payung besar milik Mina. Kereta baru dipakai setelah sampai stasiun."},
	})
	add("grammar", "読解 問題6 · Pencarian informasi", "次のお知らせを読んで、質問に答えてください。", "みどりスポーツセンター\nプール：月曜〜土曜 10:00〜18:00／一回400円\n体育館：月曜〜土曜 9:00〜20:00／一回300円\n日曜日は全部休みです。プールでは、水泳の帽子を使ってください。帽子は受付で100円で借りられます。", []question{
		{"土曜日にプールへ行きます。帽子を持っていないので借ります。全部でいくら払いますか。", [4]string{"500円", "400円", "300円", "700円"}, "Biaya kolam 400円 ditambah sewa topi 100円 = 500円. Biaya gedung olahraga tidak ikut dihitung."},
	})
	return items
}
