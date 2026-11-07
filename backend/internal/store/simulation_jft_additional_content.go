package store

// These are original A1–A2 daily-life practice questions. They use the public
// JFT-Basic section descriptions as a format guide and do not reproduce its
// sample or operational questions.
func simulationJFTAdditionalContent(paperNo int) []simulationItem {
	switch paperNo {
	case 2:
		return simulationJFTContent2()
	case 3:
		return simulationJFTContent3()
	case 4:
		return simulationJFTContent4()
	case 5:
		return simulationJFTContent5()
	default:
		// Papers 06 and later live in simulation_data/*.json (see simulation_json.go).
		return simulationJSONContent("JFT", paperNo)
	}
}

func jftOriginalItem(section, group, passage, prompt string, choices [4]string, explanation string) simulationItem {
	return simulationItem{section, group, "最もよい答えを一つ選んでください。", passage, prompt, choices, explanation}
}

func simulationJFTContent2() []simulationItem {
	q := jftOriginalItem
	return []simulationItem{
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "来週、会社の人と[[旅行]]に行きます。下線の言葉はどう読みますか。", [4]string{"りょこう", "りょうこう", "りょこ", "りょうこ"}, "旅行（りょこう） berarti perjalanan atau wisata."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "この[[住所]]を紙に書いてください。下線の言葉はどう読みますか。", [4]string{"じゅうしょ", "じゅしょ", "じょうしょ", "じゅうじょ"}, "住所 dibaca じゅうしょ, alamat."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "病院の[[受付]]で名前を言ってください。下線の言葉はどう読みますか。", [4]string{"うけつけ", "うけづけ", "うけとけ", "うけつき"}, "受付（うけつけ） adalah bagian penerimaan atau loket."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "電車やバスに乗る前に、駅の機械で切符を買います。切符は何ですか。", [4]string{"乗るための紙やカード", "荷物を入れるかばん", "食事をする店", "道を教える地図"}, "切符（きっぷ） adalah tiket untuk naik kendaraan."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "病院で先生に体の具合を話して、薬をもらいました。何をしましたか。", [4]string{"診察を受けました", "引っ越しました", "買い物をしました", "道に迷いました"}, "診察を受ける berarti diperiksa dokter."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "駅で電車の時間と行き先を見たいです。どこを見ますか。", [4]string{"時刻表", "申込書", "領収書", "案内状"}, "時刻表（じこくひょう） menunjukkan jadwal dan waktu keberangkatan."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "朝、熱を測ったら三十八度でした。会社に電話して、今日は（　）と伝えました。", [4]string{"休みます", "迎えます", "集めます", "返します"}, "Saat demam, konteksnya mengabari bahwa tidak masuk kerja: 休みます."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "この荷物は重いので、一人で持てません。少し（　）ください。", [4]string{"手伝って", "急いで", "磨いて", "乾いて"}, "手伝ってください meminta bantuan. 荷物が重いので sesuai dengan konteks."),
		q("vocabulary", "文字と語彙 · 漢字の意味と用法", "", "「右」という漢字を使う文はどれですか。", [4]string{"次の角を右に曲がります。", "今日は雨が右っています。", "右を二つ食べました。", "右で手紙を書きます。"}, "右（みぎ） dipakai untuk arah kanan, seperti 右に曲がる."),
		q("vocabulary", "文字と語彙 · 漢字の意味と用法", "", "「薬」の使い方が正しい文はどれですか。", [4]string{"食事のあとで薬を飲みます。", "駅で薬に乗ります。", "薬を着て会社へ行きます。", "薬を開けて道を渡ります。"}, "Obat diminum: 薬を飲む. Tiga pilihan lain memakai kata kerja yang tidak sesuai."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "買った物がこわれていたので、店に持って行って（　）もらいました。", [4]string{"交換して", "案内して", "出発して", "到着して"}, "交換する berarti menukar barang yang rusak dengan barang lain."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "駅の[[近く]]に郵便局があります。下線の言葉はどう読みますか。", [4]string{"ちかく", "とおく", "はやく", "おそく"}, "近く dibaca ちかく, dekat."),

		q("conversation", "会話と表現 · 文法", "", "A：明日は何時に出ますか。\nB：七時に家を（　）つもりです。", [4]string{"出る", "出て", "出た", "出ない"}, "Rencana dengan つもり menggunakan bentuk kamus: 出るつもりです."),
		q("conversation", "会話と表現 · 文法", "", "A：この薬はいつ飲みますか。\nB：朝ご飯を食べた（　）飲んでください。", [4]string{"あとで", "まえで", "ながら", "までに"}, "Makan lalu minum obat: 食べたあとで."),
		q("conversation", "会話と表現 · 文法", "", "A：駅までどうやって行きますか。\nB：このバス（　）行けます。", [4]string{"で", "を", "が", "も"}, "Kendaraan yang digunakan ditandai dengan で: バスで行く."),
		q("conversation", "会話と表現 · 文法", "", "A：週末、どこかへ行きましたか。\nB：雨が降っていた（　）、家にいました。", [4]string{"ので", "のに", "でも", "なら"}, "Alasan berada di rumah adalah hujan: ～ので berarti karena."),
		q("conversation", "会話と表現 · 文法", "", "A：窓を開けてもいいですか。\nB：すみません。寒いので、開けない（　）ください。", [4]string{"で", "て", "と", "に"}, "Larangan halus memakai ～ないでください: jangan dibuka."),
		q("conversation", "会話と表現 · 文法", "", "A：駅で山田さんに会いましたか。\nB：はい、改札の前で（　）。", [4]string{"会いました", "会います", "会いません", "会うでしょう"}, "Peristiwa yang sudah terjadi kemarin memakai bentuk lampau 会いました."),
		q("conversation", "会話と表現 · Ungkapan", "", "タクシーで、運転手に病院へ行ってほしいです。何と言いますか。", [4]string{"すみません、中央病院までお願いします。", "お会計は別々です。", "予約をキャンセルしました。", "この薬を飲んでください。"}, "Tujuan perjalanan disampaikan dengan 中央病院までお願いします."),
		q("conversation", "会話と表現 · Ungkapan", "", "郵便局で荷物を送りたいです。窓口で何と言いますか。", [4]string{"この荷物を送りたいです。", "この荷物を食べたいです。", "この荷物に乗りたいです。", "この荷物を着たいです。"}, "送りたいです menyampaikan tujuan mengirim paket di loket."),
		q("conversation", "会話と表現 · Ungkapan", "", "同僚に体調を心配されました。よくなったことを伝えるには、何と言いますか。", [4]string{"ありがとうございます。もう大丈夫です。", "お先にどうぞ。道が分かりません。", "ご注文をお聞きします。", "少々お待ちください。"}, "もう大丈夫です memberitahu bahwa kondisi sudah membaik."),
		q("conversation", "会話と表現 · Ungkapan", "", "窓口の説明が分かりません。もう一度説明してほしいとき、何と言いますか。", [4]string{"すみません、もう一度説明していただけますか。", "すみません、ここに座らないでください。", "すみません、早く帰りました。", "すみません、おなかがいっぱいです。"}, "Meminta penjelasan ulang dengan sopan: もう一度説明していただけますか."),
		q("conversation", "会話と表現 · Ungkapan", "", "友達が駅まで案内してくれました。別れるとき、何と言いますか。", [4]string{"案内してくれてありがとう。", "案内してくれていただきます。", "案内してくれて行ってきます。", "案内してくれておかえりなさい。"}, "Berterima kasih atas bantuan: 案内してくれてありがとう."),
		q("conversation", "会話と表現 · Ungkapan", "", "店で買った服のサイズを変えたいです。店員に何と言いますか。", [4]string{"すみません、別のサイズに交換できますか。", "すみません、別の駅に到着できますか。", "すみません、別の料理を洗えますか。", "すみません、別の時間を飲めますか。"}, "Menanyakan apakah ukuran bisa ditukar: 別のサイズに交換できますか."),

		q("reading", "読解 · 内容理解", "社員のみなさんへ\n来週の健康診断は、火曜日の午前九時からです。朝ご飯を食べないで、保険証を持って、二階の会議室へ来てください。仕事は診断が終わってから始めます。\n総務", "健康診断は何時からですか。", [4]string{"火曜日の午前九時", "火曜日の午後九時", "水曜日の午前九時", "来週の午前七時"}, "Jadwalnya Selasa pekan depan pukul 9 pagi."),
		q("reading", "読解 · 内容理解", "社員のみなさんへ\n来週の健康診断は、火曜日の午前九時からです。朝ご飯を食べないで、保険証を持って、二階の会議室へ来てください。仕事は診断が終わってから始めます。\n総務", "健康診断に持って行くものは何ですか。", [4]string{"保険証", "朝ご飯", "作業着", "会議の資料"}, "Pesan meminta membawa kartu asuransi kesehatan（保険証）."),
		q("reading", "読解 · 内容理解", "佐藤さん\nきのうは駅まで送ってくれてありがとう。忘れた傘は、今週の日曜日に持って行きます。日曜日は午後から雨の予報なので、駅前のカフェで会いませんか。二時ごろ着きます。\nミナ", "ミナさんは日曜日の何時ごろ来ますか。", [4]string{"午後二時ごろ", "午前二時ごろ", "午後一時ごろ", "午後四時ごろ"}, "Pesan mengatakan 二時ごろ dan konteksnya hari Minggu siang/sore."),
		q("reading", "読解 · 内容理解", "佐藤さん\nきのうは駅まで送ってくれてありがとう。忘れた傘は、今週の日曜日に持って行きます。日曜日は午後から雨の予報なので、駅前のカフェで会いませんか。二時ごろ着きます。\nミナ", "ミナさんが持って来るものは何ですか。", [4]string{"傘", "切符", "保険証", "仕事の資料"}, "ミナ akan mengembalikan payung yang tertinggal（忘れた傘）."),
		q("reading", "読解 · 情報検索", "さくらクリニック\n受付：月・水・金 8:30〜12:00、15:00〜18:00\n火・木・土 8:30〜12:00\n休み：日曜・祝日\n初診の方は、受付終了の30分前までに来てください。健康保険証をお持ちください。", "初めての人が金曜日の午後に行くとき、何時までに受付へ行きますか。", [4]string{"午後五時三十分", "午後六時", "午後三時", "午後五時"}, "受付 berakhir pukul 18.00; pasien baru harus datang 30 menit lebih awal, paling lambat 17.30."),
		q("reading", "読解 · 情報検索", "さくらクリニック\n受付：月・水・金 8:30〜12:00、15:00〜18:00\n火・木・土 8:30〜12:00\n休み：日曜・祝日\n初診の方は、受付終了の30分前までに来てください。健康保険証をお持ちください。", "土曜日に診察を受けたい人は、何時までに行けますか。", [4]string{"正午まで", "午後六時まで", "午後三時まで", "土曜日は休み"}, "Sabtu hanya buka pagi sampai pukul 12.00."),
		q("reading", "読解 · 情報検索", "まちのバス一日券\n大人：600円／子ども：300円\n市内のバスに何度でも乗れます。地下鉄では使えません。\n販売：駅の案内所、バス車内\n利用日だけ有効です。", "大人がこの券を買って、バスに三回乗ります。いくらですか。", [4]string{"600円", "300円", "1,200円", "1,800円"}, "Tiket harian dewasa berharga 600 yen dan dapat dipakai berkali-kali pada hari yang sama."),
		q("reading", "読解 · 情報検索", "まちのバス一日券\n大人：600円／子ども：300円\n市内のバスに何度でも乗れます。地下鉄では使えません。\n販売：駅の案内所、バス車内\n利用日だけ有効です。", "地下鉄に乗るとき、この券は使えますか。", [4]string{"いいえ、使えません。", "はい、何度でも使えます。", "子どもだけ使えます。", "駅の案内所でだけ使えます。"}, "Informasi secara langsung menyatakan 地下鉄では使えません."),
		q("reading", "読解 · 情報検索", "市民スポーツセンター\nプール：火〜金 9:00〜20:00／土・日 9:00〜17:00\n休館日：月曜日\n大人400円、学生200円。水泳帽が必要です。受付でロッカーの鍵を受け取ってください。", "日曜日にプールへ行くと、何時に閉まりますか。", [4]string{"午後五時", "午後八時", "午前九時", "午後四時"}, "Hari Minggu kolam buka sampai pukul 17.00."),
		q("reading", "読解 · 情報検索", "市民スポーツセンター\nプール：火〜金 9:00〜20:00／土・日 9:00〜17:00\n休館日：月曜日\n大人400円、学生200円。水泳帽が必要です。受付でロッカーの鍵を受け取ってください。", "プールに入るために、何が必要ですか。", [4]string{"水泳帽", "学生証だけ", "ロッカーの鍵を自分で買うこと", "月曜日の予約"}, "Kolam renang mewajibkan topi renang（水泳帽）."),
		q("reading", "読解 · 情報検索", "中央郵便局\n窓口：月〜金 9:00〜17:00／土 9:00〜12:00\n日曜・祝日は休み\nATM：毎日 8:00〜20:00\n荷物を送る方は、窓口で送り先の住所を書いてください。", "土曜日に荷物を送りたいです。窓口は何時まで開いていますか。", [4]string{"正午まで", "午後五時まで", "午後八時まで", "土曜日は休み"}, "Loket pengiriman hari Sabtu buka sampai pukul 12.00 (正午)."),
		q("reading", "読解 · 情報検索", "中央郵便局\n窓口：月〜金 9:00〜17:00／土 9:00〜12:00\n日曜・祝日は休み\nATM：毎日 8:00〜20:00\n荷物を送る方は、窓口で送り先の住所を書いてください。", "日曜日にATMを使えますか。", [4]string{"はい、使えます。", "いいえ、日曜日は郵便局もATMも休みです。", "いいえ、ATMは土曜日だけです。", "はい、午後八時から使えます。"}, "ATM は毎日8時から20時までなので、日曜日も使えます."),
	}
}

func simulationJFTContent3() []simulationItem {
	q := jftOriginalItem
	return []simulationItem{
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "工場の[[近く]]にコンビニがあります。下線の言葉はどう読みますか。", [4]string{"ちかく", "とおく", "ひろく", "あかるく"}, "近く（ちかく） berarti dekat."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "仕事の前に、今日の[[予定]]を確認します。下線の言葉はどう読みますか。", [4]string{"よてい", "ようてい", "よて", "よたい"}, "予定（よてい） adalah rencana atau jadwal."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "作業の前に、この説明を[[読んで]]ください。下線の言葉はどう読みますか。", [4]string{"よんで", "やんで", "おんで", "こんで"}, "読んで dibaca よんで, bentuk te dari 読む."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "仕事で使う時間や休みの日を書いた紙を、会社からもらいました。何ですか。", [4]string{"勤務表", "買い物袋", "乗車券", "料理本"}, "勤務表（きんむひょう） memuat jadwal kerja dan hari libur."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "仕事が終わったあと、会社の人と一緒に食事をしました。これは何ですか。", [4]string{"食事会", "健康診断", "説明書", "引っ越し"}, "Makan bersama setelah kerja adalah 食事会."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "仕事を始める前に、リーダーから安全について話を聞きました。何を聞きましたか。", [4]string{"注意事項", "注文番号", "天気予報", "住所変更"}, "注意事項 adalah hal-hal yang harus diperhatikan, cocok dengan pengarahan keselamatan."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "機械の使い方が分からないので、先輩に（　）しました。", [4]string{"質問", "運転", "出発", "到着"}, "質問する berarti bertanya saat tidak memahami cara menggunakan mesin."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "仕事が終わったら、使った道具を元の場所に（　）ください。", [4]string{"戻して", "乗って", "払って", "浴びて"}, "道具を元の場所に戻す berarti mengembalikan alat ke tempat semula."),
		q("vocabulary", "文字と語彙 · 漢字の意味と用法", "", "「休」という漢字を使う文はどれですか。", [4]string{"昼休みに食堂でご飯を食べます。", "昼休みに電車を着ます。", "昼休みに手紙を泳ぎます。", "昼休みに住所を乗ります。"}, "休（やす）む berkaitan dengan istirahat: 昼休み adalah istirahat siang."),
		q("vocabulary", "文字と語彙 · 漢字の意味と用法", "", "「開ける」の使い方が正しい文はどれですか。", [4]string{"暑いので、窓を開けます。", "暑いので、窓を飲みます。", "暑いので、窓を着ます。", "暑いので、窓を返します。"}, "窓を開ける berarti membuka jendela untuk ventilasi."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "来月の休みを決めたので、会社へ休暇を（　）しました。", [4]string{"申請", "出勤", "掃除", "洗濯"}, "休暇を申請する berarti mengajukan cuti."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "作業が終わったら、機械の電源を[[切って]]ください。下線の言葉はどう読みますか。", [4]string{"きって", "きいて", "けって", "こって"}, "切って（きって） adalah bentuk te dari 切る, mematikan atau memutus."),

		q("conversation", "会話と表現 · 文法", "", "A：日本で働いてどのくらいですか。\nB：去年の四月から働いているので、もう一年（　）です。", [4]string{"になります", "にします", "をします", "でいます"}, "Durasi yang akan dicapai: 一年になります = akan genap satu tahun."),
		q("conversation", "会話と表現 · 文法", "", "A：この箱はどこに置きますか。\nB：あの棚の上に（　）ください。", [4]string{"置いて", "置くと", "置いた", "置きながら"}, "Permintaan menggunakan kata kerja bentuk te: 置いてください."),
		q("conversation", "会話と表現 · 文法", "", "A：今日は残業しますか。\nB：いいえ、用事があるので、六時に（　）予定です。", [4]string{"帰る", "帰って", "帰った", "帰らないで"}, "Kata benda 予定 ditemani bentuk kamus: 帰る予定です."),
		q("conversation", "会話と表現 · 文法", "", "A：この紙に名前を書きましたか。\nB：はい、住所も電話番号（　）書きました。", [4]string{"も", "を", "が", "で"}, "も menambahkan informasi setara: selain nama, alamat dan nomor telepon juga ditulis."),
		q("conversation", "会話と表現 · 文法", "", "A：どうして手袋をしていますか。\nB：手が汚れない（　）です。", [4]string{"ように", "ことに", "そうに", "みたいに"}, "Tujuan/agar tangan tidak kotor dinyatakan dengan ～ないように."),
		q("conversation", "会話と表現 · 文法", "", "A：この荷物を一人で運べますか。\nB：重すぎて、一人では（　）。", [4]string{"運べません", "運びます", "運びました", "運ぶでしょう"}, "Beban terlalu berat sehingga tidak bisa diangkut sendiri: 運べません."),
		q("conversation", "会話と表現 · Ungkapan", "", "仕事で失敗したことをリーダーに報告します。最初に何と言いますか。", [4]string{"すみません、報告したいことがあります。", "おめでとうございます、失敗しました。", "いってらっしゃい、報告しました。", "ごちそうさまでした、お願いします。"}, "Laporan masalah dapat dimulai dengan meminta perhatian secara sopan: 報告したいことがあります."),
		q("conversation", "会話と表現 · Ungkapan", "", "同僚から仕事の道具を借りました。返すとき、何と言いますか。", [4]string{"ありがとうございました。助かりました。", "ご注文は何になさいますか。", "いただきます。", "おかえりなさい。"}, "Mengucapkan terima kasih dan mengatakan bantuan itu berguna: 助かりました."),
		q("conversation", "会話と表現 · Ungkapan", "", "朝、会社に着いたとき、同僚に何と言いますか。", [4]string{"おはようございます。", "おやすみなさい。", "ごちそうさまでした。", "いってらっしゃい。"}, "Salam saat tiba di tempat kerja pada pagi hari adalah おはようございます."),
		q("conversation", "会話と表現 · Ungkapan", "", "作業の説明を聞いたあと、内容が分かったことを伝えます。何と言いますか。", [4]string{"分かりました。気をつけます。", "まだ注文していません。", "お大事にしてください。", "何名様ですか。"}, "分かりました。気をつけます menyatakan pemahaman dan kesediaan berhati-hati."),
		q("conversation", "会話と表現 · Ungkapan", "", "仕事に遅れそうです。到着する時間も伝えて、電話で何と言いますか。", [4]string{"電車が遅れていて、九時十分ごろ着きます。", "電車に乗って、昨日帰りました。", "電車を洗って、九時に食べます。", "電車が好きなので、休みました。"}, "Kalimat pertama menjelaskan keterlambatan kereta sekaligus perkiraan tiba pukul 9.10."),
		q("conversation", "会話と表現 · Ungkapan", "", "休みの日に同僚から出勤を頼まれましたが、病院の予約があります。何と言いますか。", [4]string{"すみません、その日は病院の予約があるので出勤できません。", "はい、その日は病院へ行きません。", "お先に失礼します。予約を食べます。", "ぜひ、予約を洗ってください。"}, "Menolak permintaan dengan sopan dan memberi alasan: ada janji di rumah sakit."),

		q("reading", "読解 · 内容理解", "みなさんへ\n新しい作業着を金曜日に配ります。自分の名前を書いた袋を持って、午後三時に休憩室へ来てください。サイズを変えたい人は、今週の水曜日までに事務所へ連絡してください。\n山田", "作業着はいつ配りますか。", [4]string{"金曜日", "水曜日", "今週の日曜日", "毎週の月曜日"}, "Pakaian kerja baru dibagikan pada Jumat."),
		q("reading", "読解 · 内容理解", "みなさんへ\n新しい作業着を金曜日に配ります。自分の名前を書いた袋を持って、午後三時に休憩室へ来てください。サイズを変えたい人は、今週の水曜日までに事務所へ連絡してください。\n山田", "サイズを変えたい人は、いつまでに連絡しますか。", [4]string{"水曜日まで", "金曜日まで", "来月まで", "午後三時まで"}, "Perubahan ukuran harus dilaporkan paling lambat Rabu minggu ini."),
		q("reading", "読解 · 内容理解", "お母さんへ\n今日は仕事のあと、日本語教室へ行きます。帰るのは九時ごろです。冷蔵庫にカレーがありますから、先に食べてください。明日の朝は早いので、洗濯は週末にします。\nアリ", "アリさんは今日、仕事のあとどこへ行きますか。", [4]string{"日本語教室", "病院", "駅の事務所", "スポーツセンター"}, "Setelah bekerja, Ari pergi ke kelas bahasa Jepang."),
		q("reading", "読解 · 内容理解", "お母さんへ\n今日は仕事のあと、日本語教室へ行きます。帰るのは九時ごろです。冷蔵庫にカレーがありますから、先に食べてください。明日の朝は早いので、洗濯は週末にします。\nアリ", "洗濯はいつしますか。", [4]string{"週末", "今日の午後", "明日の朝", "日本語教室の前"}, "Ia akan mencuci pakaian pada akhir pekan karena besok pagi harus bangun awal."),
		q("reading", "読解 · 情報検索", "にじいろ食堂の営業時間\n月〜金：11:30〜14:00、17:00〜21:00\n土：11:30〜21:00\n日曜休み\nランチセットは平日の14:00まで。現金と交通系ICカードが使えます。予約は電話でお願いします。", "土曜日の午後三時に食事をしたいです。店は開いていますか。", [4]string{"はい、開いています。", "いいえ、午後二時に閉まります。", "いいえ、土曜日は休みです。", "はい、ランチセットだけです。"}, "Sabtu buka terus pukul 11.30 sampai 21.00; pukul 15.00 masih buka."),
		q("reading", "読解 · 情報検索", "にじいろ食堂の営業時間\n月〜金：11:30〜14:00、17:00〜21:00\n土：11:30〜21:00\n日曜休み\nランチセットは平日の14:00まで。現金と交通系ICカードが使えます。予約は電話でお願いします。", "平日のランチセットを注文できるのは、何時までですか。", [4]string{"午後二時まで", "午後五時まで", "午後九時まで", "午前十一時まで"}, "Lunch set tersedia pada hari kerja sampai pukul 14.00."),
		q("reading", "読解 · 情報検索", "社内ミーティング\n場所：三階の会議室\n時間：木曜日 16:00〜16:30\n持ち物：今月の勤務表\n欠席する人は、前日までにチームリーダーへ連絡してください。", "ミーティングに必要なものは何ですか。", [4]string{"今月の勤務表", "健康保険証", "水泳帽", "新しい作業着"}, "Peserta harus membawa jadwal kerja bulan ini（勤務表）."),
		q("reading", "読解 · 情報検索", "社内ミーティング\n場所：三階の会議室\n時間：木曜日 16:00〜16:30\n持ち物：今月の勤務表\n欠席する人は、前日までにチームリーダーへ連絡してください。", "木曜日のミーティングに出られない人は、どうしますか。", [4]string{"前日までにリーダーへ連絡します。", "当日の午後に会議室へ行きます。", "勤務表を家に置きます。", "次の月曜日まで何もしません。"}, "Ketidakhadiran diberitahukan kepada ketua tim paling lambat sehari sebelumnya."),
		q("reading", "読解 · 情報検索", "市役所のごみ相談\n電話相談：平日 8:30〜17:00\n粗大ごみの申込み：電話または市のウェブサイト\n申込みのとき、住所とごみの大きさを教えてください。収集日は申込みのあとに決まります。", "大きな家具を捨てたいです。どのように申し込みますか。", [4]string{"電話か市のウェブサイトで申し込みます。", "駅の切符売り場で申し込みます。", "毎週日曜日に市役所へ持って行きます。", "住所を書かずにごみ置き場へ出します。"}, "Sampah besar diajukan lewat telepon atau situs kota."),
		q("reading", "読解 · 情報検索", "市役所のごみ相談\n電話相談：平日 8:30〜17:00\n粗大ごみの申込み：電話または市のウェブサイト\n申込みのとき、住所とごみの大きさを教えてください。収集日は申込みのあとに決まります。", "電話で相談できるのは、何曜日の何時ですか。", [4]string{"平日の午前八時半から午後五時まで", "日曜日の午前九時から午後五時まで", "毎日の午後五時から", "土曜日の午前八時まで"}, "Konsultasi telepon tersedia hari kerja pukul 8.30–17.00."),
		q("reading", "読解 · 内容理解", "となり町の田中さんへ\n日曜日に公園の掃除をします。朝九時に公園の入口に集まってください。ごみ袋と手袋はこちらで用意します。雨が降ったら、来週の日曜日にします。参加できるかどうか、金曜日までに教えてください。\n自治会の伊藤", "日曜日に雨が降ったら、掃除はいつしますか。", [4]string{"来週の日曜日", "今週の金曜日", "月曜日の朝", "雨の日の午後"}, "Jika hujan pada hari Minggu, kegiatan dipindah ke Minggu berikutnya."),
		q("reading", "読解 · 内容理解", "となり町の田中さんへ\n日曜日に公園の掃除をします。朝九時に公園の入口に集まってください。ごみ袋と手袋はこちらで用意します。雨が降ったら、来週の日曜日にします。参加できるかどうか、金曜日までに教えてください。\n自治会の伊藤", "参加できるかどうか、いつまでに伝えますか。", [4]string{"金曜日まで", "日曜日の朝", "来週の日曜日", "掃除が終わったあと"}, "Pesan meminta konfirmasi partisipasi paling lambat hari Jumat."),
	}
}

func simulationJFTContent4() []simulationItem {
	q := jftOriginalItem
	return []simulationItem{
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "この部屋は[[家賃]]が安いです。下線の言葉はどう読みますか。", [4]string{"やちん", "いえちん", "やきん", "かちん"}, "家賃（やちん） adalah uang sewa rumah."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "買い物のあとで[[袋]]を一つください。下線の言葉はどう読みますか。", [4]string{"ふくろ", "ふくろう", "ふくる", "ふくろく"}, "袋 dibaca ふくろ, tas atau kantong."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "銀行でお金を[[下ろし]]ます。下線の言葉はどう読みますか。", [4]string{"おろし", "おとし", "くだし", "さげし"}, "下ろす（おろす） berarti menarik uang dari bank atau ATM."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "アパートの部屋を借りる前に、不動産会社で部屋を見ました。何をしましたか。", [4]string{"見学", "出勤", "運転", "洗濯"}, "部屋を見学する berarti melihat-lihat kamar sebelum memutuskan."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "引っ越しのとき、住所が変わったことを市役所に知らせます。何をしますか。", [4]string{"住所の変更手続き", "商品の予約", "病気の診察", "電車の乗り換え"}, "住所が変わったら市役所で住所変更の手続きをします."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "お金を入れたり、口座の残高を調べたりする機械は何ですか。", [4]string{"ATM", "冷蔵庫", "改札", "自動販売機"}, "ATM digunakan untuk transaksi rekening seperti menarik uang dan memeriksa saldo."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "この服は少し大きいので、小さいサイズに（　）たいです。", [4]string{"交換し", "出発し", "説明し", "到着し"}, "Pakaian yang ukurannya tidak cocok ingin ditukar: 交換したいです."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "料理の前に、野菜をよく（　）ください。", [4]string{"洗って", "渡って", "払って", "曲がって"}, "Sebelum memasak, sayuran dicuci: 野菜を洗う."),
		q("vocabulary", "文字と語彙 · 漢字の意味と用法", "", "「貸す」の使い方が正しい文はどれですか。", [4]string{"友達に傘を貸しました。", "友達に傘を食べました。", "友達に傘を泳ぎました。", "友達に傘を着ました。"}, "貸す berarti meminjamkan sesuatu kepada orang lain: 傘を貸す."),
		q("vocabulary", "文字と語彙 · 漢字の意味と用法", "", "「返す」の使い方が正しい文はどれですか。", [4]string{"図書館に本を返します。", "図書館に本を歩きます。", "図書館に本を乗ります。", "図書館に本を洗います。"}, "図書館に本を返す berarti mengembalikan buku ke perpustakaan."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "店で買った物の値段が分かる紙を、レジでもらいました。これを（　）と言います。", [4]string{"レシート", "切符", "時刻表", "申込書"}, "レシート adalah struk yang menunjukkan barang dan harga."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "この建物の[[入口]]はあちらです。下線の言葉はどう読みますか。", [4]string{"いりぐち", "でぐち", "いりくち", "にゅうこう"}, "入口 dibaca いりぐち, pintu masuk."),

		q("conversation", "会話と表現 · 文法", "", "A：この部屋に住んでどのくらいですか。\nB：去年から住んでいるので、もう一年（　）。", [4]string{"になります", "にします", "でいます", "をします"}, "Lama waktu yang akan genap diungkapkan dengan 一年になります."),
		q("conversation", "会話と表現 · 文法", "", "A：このシャツを着てもいいですか。\nB：はい、試着室で着て（　）ください。", [4]string{"みて", "おいて", "しまって", "あって"}, "試着して確認する: 着てみてください (silakan coba kenakan)."),
		q("conversation", "会話と表現 · 文法", "", "A：銀行はどこですか。\nB：この道をまっすぐ行く（　）、右にあります。", [4]string{"と", "から", "まで", "でも"}, "Pola arah ～と、... digunakan untuk menjelaskan hasil saat mengikuti rute: lurus lalu ada di kanan."),
		q("conversation", "会話と表現 · 文法", "", "A：部屋の窓が閉まっていますね。\nB：はい。寒い（　）、閉めました。", [4]string{"ので", "のに", "でも", "ながら"}, "Alasan menutup jendela adalah dingin, maka gunakan ので."),
		q("conversation", "会話と表現 · 文法", "", "A：新しい住所を市役所に知らせましたか。\nB：はい、先週、変更の手続きを（　）。", [4]string{"しました", "します", "しません", "するでしょう"}, "Tindakan sudah dilakukan minggu lalu: 手続きをしました."),
		q("conversation", "会話と表現 · 文法", "", "A：このいす、使ってもいいですか。\nB：どうぞ。そこに（　）あります。", [4]string{"置いて", "置く", "置いた", "置きながら"}, "置いてあります menjelaskan kursi sudah diletakkan di sana."),
		q("conversation", "会話と表現 · Ungkapan", "", "アパートを見たあと、部屋を借りたいと不動産会社に伝えます。何と言いますか。", [4]string{"この部屋を借りたいです。", "この部屋を返したいです。", "この部屋を買って食べたいです。", "この部屋で切符を作りたいです。"}, "Menyampaikan keinginan menyewa: この部屋を借りたいです."),
		q("conversation", "会話と表現 · Ungkapan", "", "銀行のATMの使い方が分かりません。近くの人にどう頼みますか。", [4]string{"すみません、使い方を教えていただけませんか。", "すみません、口座を食べませんか。", "すみません、住所を歩きませんか。", "すみません、駅まで泳ぎませんか。"}, "Permintaan bantuan yang sopan: 使い方を教えていただけませんか."),
		q("conversation", "会話と表現 · Ungkapan", "", "店員に、袋がいらないことを伝えます。何と言いますか。", [4]string{"袋は大丈夫です。", "袋を着ています。", "袋に乗ります。", "袋が食べたいです。"}, "袋は大丈夫です dapat dipakai untuk menyatakan tidak perlu kantong."),
		q("conversation", "会話と表現 · Ungkapan", "", "大家さんに、水が出ないことを知らせます。何と言いますか。", [4]string{"台所の水が出ないので、確認していただけますか。", "台所の水を食べたいので、駅へ行けますか。", "台所の水を着たので、遅れます。", "台所の水は日曜日に休みます。"}, "Menjelaskan masalah keran dan meminta pemilik rumah memeriksa."),
		q("conversation", "会話と表現 · Ungkapan", "", "友達の家に招待されました。食事を始めるとき、何と言いますか。", [4]string{"いただきます。", "おやすみなさい。", "お先に失礼します。", "いってらっしゃい。"}, "Sebelum mulai makan, orang Jepang mengucapkan いただきます."),
		q("conversation", "会話と表現 · Ungkapan", "", "買った商品に問題があり、店員に相談します。最初に何と言いますか。", [4]string{"すみません、この商品について相談したいです。", "おめでとうございます、この商品を飲みます。", "いただきます、この商品を洗います。", "おかえりなさい、この商品を寝ます。"}, "Memulai konsultasi dengan sopan: 商品について相談したいです."),

		q("reading", "読解 · 内容理解", "山本さん\n新しい住所の連絡をありがとう。市役所の手続きは、引っ越した日から十四日以内にする必要があるそうです。私は土曜日に手伝いに行きます。午前中に荷物が届くので、午後一時に駅で会いましょう。\nミカ", "ミカさんは土曜日の何時に会いますか。", [4]string{"午後一時", "午前九時", "正午", "午後二時"}, "Pesan menyebut janji bertemu pukul 13.00."),
		q("reading", "読解 · 内容理解", "山本さん\n新しい住所の連絡をありがとう。市役所の手続きは、引っ越した日から十四日以内にする必要があるそうです。私は土曜日に手伝いに行きます。午前中に荷物が届くので、午後一時に駅で会いましょう。\nミカ", "市役所の手続きは、いつまでにしますか。", [4]string{"引っ越した日から十四日以内", "荷物が届く前", "土曜日の午後一時", "新しい住所を決める前"}, "Pemberitahuan menyatakan prosedur dilakukan dalam 14 hari setelah pindah."),
		q("reading", "読解 · 内容理解", "お客様へ\n本日、エレベーターの点検をしています。午前十時から正午まで、エレベーターは使えません。階段をご利用ください。大きな荷物を運ぶ方は、管理人室へご相談ください。\nマンション管理室", "エレベーターが使えないのは、何時までですか。", [4]string{"正午まで", "午前十時まで", "午後二時まで", "午後五時まで"}, "Lift tidak dapat dipakai dari jam 10 sampai siang: 正午まで."),
		q("reading", "読解 · 内容理解", "お客様へ\n本日、エレベーターの点検をしています。午前十時から正午まで、エレベーターは使えません。階段をご利用ください。大きな荷物を運ぶ方は、管理人室へご相談ください。\nマンション管理室", "点検中、どうやって上の階へ行きますか。", [4]string{"階段を使います。", "エレベーターを使います。", "管理人室の中で待ちます。", "大きな荷物を置いて帰ります。"}, "Pengumuman menyuruh penghuni menggunakan tangga selama pemeriksaan."),
		q("reading", "読解 · 情報検索", "あおぞらスーパー\n営業時間：毎日 9:00〜21:00\n水曜日はポイント二倍\nレジ袋：一枚5円\n配達サービス：2,000円以上のお買い物で無料。配達の申込みはサービスカウンターへ。", "水曜日に買い物をすると、何が二倍になりますか。", [4]string{"ポイント", "レジ袋", "営業時間", "配達の料金"}, "Promo Rabu menggandakan poin belanja."),
		q("reading", "読解 · 情報検索", "あおぞらスーパー\n営業時間：毎日 9:00〜21:00\n水曜日はポイント二倍\nレジ袋：一枚5円\n配達サービス：2,000円以上のお買い物で無料。配達の申込みはサービスカウンターへ。", "2,500円買った物を家に届けてほしいです。どこで頼みますか。", [4]string{"サービスカウンター", "レジ袋の棚", "店の入口", "水曜日のポイント売り場"}, "Pengiriman gratis untuk belanja minimal 2.000 yen dan dipesan di service counter."),
		q("reading", "読解 · 情報検索", "市民ホールの料理教室\n日時：6月8日（土）10:00〜12:00\n参加費：800円（材料費を含む）\n定員：20人\n申込み：6月1日までに電話で。エプロンを持って来てください。", "料理教室に申し込む方法は何ですか。", [4]string{"6月1日までに電話します。", "6月8日に市役所へ行きます。", "材料を買って郵送します。", "当日にメールを送ります。"}, "Pendaftaran melalui telepon paling lambat 1 Juni."),
		q("reading", "読解 · 情報検索", "市民ホールの料理教室\n日時：6月8日（土）10:00〜12:00\n参加費：800円（材料費を含む）\n定員：20人\n申込み：6月1日までに電話で。エプロンを持って来てください。", "参加するとき、自分で持って行くものは何ですか。", [4]string{"エプロン", "材料費", "電話", "二十人分の料理"}, "Peserta diminta membawa celemek (エプロン). Biaya bahan sudah termasuk."),
		q("reading", "読解 · 情報検索", "さくら銀行 ATM\n利用時間：平日 7:00〜22:00／土・日 8:00〜20:00\n手数料：平日 8:45〜18:00は無料。それ以外は110円。\n通帳の記帳は平日だけできます。", "火曜日の午後一時にお金を下ろすと、手数料はいくらですか。", [4]string{"無料", "110円", "220円", "分かりません"}, "Selasa pukul 13.00 termasuk hari kerja dan rentang 8.45–18.00, sehingga gratis."),
		q("reading", "読解 · 情報検索", "さくら銀行 ATM\n利用時間：平日 7:00〜22:00／土・日 8:00〜20:00\n手数料：平日 8:45〜18:00は無料。それ以外は110円。\n通帳の記帳は平日だけできます。", "日曜日に通帳を記帳できますか。", [4]string{"いいえ、平日だけできます。", "はい、午前八時からできます。", "はい、手数料が無料ならできます。", "いいえ、銀行の通帳はありません。"}, "Passbook update is available only on weekdays（平日だけ）."),
		q("reading", "読解 · 情報検索", "あさひ不動産 内覧予約\n電話受付：月〜土 9:30〜18:00\n内覧希望日の前日までに予約してください。\n当日は、駅の北口で担当者と会います。本人確認書類を持参してください。", "日曜日に部屋を見たいです。いつまでに予約しますか。", [4]string{"土曜日まで", "日曜日の午後", "月曜日の朝", "内覧のあと"}, "Kunjungan dijadwalkan paling lambat sehari sebelumnya; untuk Minggu berarti Sabtu."),
		q("reading", "読解 · 情報検索", "あさひ不動産 内覧予約\n電話受付：月〜土 9:30〜18:00\n内覧希望日の前日までに予約してください。\n当日は、駅の北口で担当者と会います。本人確認書類を持参してください。", "内覧当日、何を持って行きますか。", [4]string{"本人確認書類", "家賃の請求書", "駅の切符", "料理教室のエプロン"}, "Pengunjung diminta membawa dokumen identitas (本人確認書類)."),
	}
}

func simulationJFTContent5() []simulationItem {
	q := jftOriginalItem
	return []simulationItem{
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "来月、会社の[[面接]]を受けます。下線の言葉はどう読みますか。", [4]string{"めんせつ", "めんぜつ", "めんさつ", "めんせち"}, "面接（めんせつ） adalah wawancara kerja."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "日本で働くために、在留[[資格]]を確認します。下線の言葉はどう読みますか。", [4]string{"しかく", "じかく", "しがく", "じがく"}, "資格 dibaca しかく, kualifikasi atau status yang memenuhi syarat."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "大雨のときは、近くの[[避難所]]へ行ってください。下線の言葉はどう読みますか。", [4]string{"ひなんじょ", "ひなんしょ", "びなんじょ", "ひなじょ"}, "避難所（ひなんじょ） adalah tempat pengungsian."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "仕事の内容や給料、働く時間が書いてある紙を読みました。何という紙ですか。", [4]string{"求人票", "時刻表", "診察券", "領収書"}, "求人票（きゅうじんひょう） memuat informasi lowongan pekerjaan."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "地震が起きたときに安全に逃げる場所を、町の地図で確認します。何という場所ですか。", [4]string{"避難場所", "駐車場", "休憩室", "受付"}, "避難場所 adalah titik aman untuk berlindung saat bencana."),
		q("vocabulary", "文字と語彙 · 語の意味", "", "仕事を始める前に、会社の人が職場のルールを説明しました。これは何ですか。", [4]string{"オリエンテーション", "お見舞い", "おつり", "乗り換え"}, "オリエンテーション adalah penjelasan atau pengenalan awal di tempat kerja."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "台風が近づいているので、窓の近くの物を中に（　）ください。", [4]string{"入れて", "出発して", "運転して", "申し込んで"}, "Benda di dekat jendela dimasukkan ke dalam: 中に入れる."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "面接の前に、履歴書の住所と電話番号を（　）しました。", [4]string{"確認", "洗濯", "出発", "配達"}, "Sebelum wawancara, data di resume diperiksa: 確認しました."),
		q("vocabulary", "文字と語彙 · 漢字の意味と用法", "", "「送る」の使い方が正しい文はどれですか。", [4]string{"母に写真をメールで送ります。", "母に写真を飲みます。", "母に写真を乗ります。", "母に写真を寝ます。"}, "送る berarti mengirim, misalnya mengirim foto lewat email."),
		q("vocabulary", "文字と語彙 · 漢字の意味と用法", "", "「危ない」の使い方が正しい文はどれですか。", [4]string{"車が来るので、この道を渡るのは危ないです。", "車が来るので、この道を食べるのは危ないです。", "車が来るので、この道を着るのは危ないです。", "車が来るので、この道を借りるのは危ないです。"}, "渡るのは危ない describes danger when crossing a road with traffic."),
		q("vocabulary", "文字と語彙 · 語の用法", "", "仕事に応募したいので、会社へ履歴書を（　）します。", [4]string{"提出", "交換", "運転", "到着"}, "履歴書を提出する berarti menyerahkan resume untuk melamar."),
		q("vocabulary", "文字と語彙 · 漢字の読み", "", "面接の[[時間]]を間違えないでください。下線の言葉はどう読みますか。", [4]string{"じかん", "じけん", "しけん", "じっかん"}, "時間 dibaca じかん, waktu atau durasi."),

		q("conversation", "会話と表現 · 文法", "", "A：面接は何時からですか。\nB：十時からです。九時半までに会社に（　）ください。", [4]string{"来て", "来る", "来た", "来ない"}, "Batas waktu sebelum sesi menggunakan までに, lalu permintaan 来てください."),
		q("conversation", "会話と表現 · 文法", "", "A：この紙はいつまでに出しますか。\nB：金曜日までに（　）ください。", [4]string{"出して", "出る", "出た", "出ない"}, "Permintaan untuk menyerahkan dokumen memakai 出してください."),
		q("conversation", "会話と表現 · 文法", "", "A：大雨になったら、どうしますか。\nB：外へ出ないで、安全な場所に（　）つもりです。", [4]string{"います", "いる", "いて", "いた"}, "つもりです memakai bentuk kamus: いるつもりです."),
		q("conversation", "会話と表現 · 文法", "", "A：この漢字が読めますか。\nB：いいえ、読み方が分からない（　）、教えてください。", [4]string{"ので", "のに", "まで", "しか"}, "Alasan meminta bantuan: karena tidak tahu bacaannya, gunakan ので."),
		q("conversation", "会話と表現 · 文法", "", "A：履歴書をメールで送りましたか。\nB：はい、昨日、会社に（　）。", [4]string{"送りました", "送ります", "送るでしょう", "送ってください"}, "Tindakan kemarin sudah selesai: 送りました."),
		q("conversation", "会話と表現 · 文法", "", "A：避難するとき、何を持って行きますか。\nB：水や食べ物など、必要な物（　）持って行きます。", [4]string{"を", "に", "で", "へ"}, "Objek yang dibawa ditandai を: 必要な物を持って行く."),
		q("conversation", "会話と表現 · Ungkapan", "", "会社に応募するとき、面接の時間を確認したいです。何と言いますか。", [4]string{"面接は何時からでしょうか。", "面接は何を食べますか。", "面接に何回乗りますか。", "面接の住所を着ますか。"}, "面接は何時からでしょうか sopan untuk memastikan jam wawancara."),
		q("conversation", "会話と表現 · Ungkapan", "", "同僚が避難場所を教えてくれました。何と言いますか。", [4]string{"教えてくれてありがとうございます。", "教えてくれていただきます。", "教えてくれておやすみなさい。", "教えてくれてご注文は。"}, "Mengucapkan terima kasih atas informasi: 教えてくれてありがとうございます."),
		q("conversation", "会話と表現 · Ungkapan", "", "避難訓練で、係の人に「こちらへ来てください」と言われました。どう答えますか。", [4]string{"はい、分かりました。", "いいえ、注文します。", "いただきます。", "お会計をお願いします。"}, "はい、分かりました merupakan tanggapan wajar atas instruksi."),
		q("conversation", "会話と表現 · Ungkapan", "", "仕事を休むため、朝、会社へ電話します。何と言いますか。", [4]string{"熱があるので、今日は休ませていただけますか。", "熱があるので、今日は旅行に来てください。", "熱があるので、会議の資料を食べます。", "熱があるので、窓を開けましたか。"}, "Meminta izin tidak masuk kerja karena demam dengan ungkapan sopan."),
		q("conversation", "会話と表現 · Ungkapan", "", "仕事の面接の最後に、相手へお礼を言います。何と言いますか。", [4]string{"本日はありがとうございました。", "いってらっしゃい。", "おかえりなさい。", "おかわりをください。"}, "Ungkapan terima kasih yang sopan di akhir wawancara: 本日はありがとうございました."),
		q("conversation", "会話と表現 · Ungkapan", "", "地震のあと、友達がけがをしていないか聞きます。何と言いますか。", [4]string{"けがはありませんか。大丈夫ですか。", "何名様ですか。お決まりですか。", "どちらまでですか。何時ですか。", "何を食べますか。いくらですか。"}, "Menanyakan kondisi orang setelah gempa: apakah terluka dan baik-baik saja."),

		q("reading", "読解 · 内容理解", "リーさん\n面接の日が決まりました。来週の木曜日、午前十時に会社へ来てください。履歴書と在留カードを持って来てください。会社の入口で「採用担当の田中さんに会いに来ました」と言ってください。\n人事部", "面接はいつですか。", [4]string{"来週の木曜日の午前十時", "今週の木曜日の午後十時", "来週の金曜日の午前十時", "今日の午前九時"}, "Wawancara dijadwalkan Kamis depan pukul 10 pagi."),
		q("reading", "読解 · 内容理解", "リーさん\n面接の日が決まりました。来週の木曜日、午前十時に会社へ来てください。履歴書と在留カードを持って来てください。会社の入口で「採用担当の田中さんに会いに来ました」と言ってください。\n人事部", "面接の日に持って行く物は何ですか。", [4]string{"履歴書と在留カード", "保険証と水泳帽", "切符と勤務表", "薬と住所の地図"}, "Pesan meminta membawa resume dan kartu izin tinggal."),
		q("reading", "読解 · 内容理解", "防災のお知らせ\n明日、町の防災訓練をします。午前九時に市民公園へ集まってください。雨の場合は、市民センターで行います。参加する人は、飲み物とタオルを持って来てください。\n市役所", "防災訓練は何時に始まりますか。", [4]string{"午前九時", "午前七時", "正午", "午後九時"}, "Peserta diminta berkumpul pukul 9 pagi."),
		q("reading", "読解 · 内容理解", "防災のお知らせ\n明日、町の防災訓練をします。午前九時に市民公園へ集まってください。雨の場合は、市民センターで行います。参加する人は、飲み物とタオルを持って来てください。\n市役所", "雨が降ったら、どこで訓練をしますか。", [4]string{"市民センター", "市民公園", "駅の北口", "会社の休憩室"}, "Jika hujan, latihan dipindahkan ke pusat warga（市民センター）."),
		q("reading", "読解 · 情報検索", "市役所の窓口\n住民票：月〜金 8:30〜17:00\n市民税：月〜金 9:00〜16:30\nパスポート：月・水・金 9:00〜16:00\n土・日・祝日は休み。番号札を取ってお待ちください。", "水曜日の午後四時に住民票を取りたいです。窓口は開いていますか。", [4]string{"はい、午後五時まで開いています。", "いいえ、午後四時に閉まります。", "いいえ、水曜日は休みです。", "はい、午後六時まで開いています。"}, "住民票窓口 tutup pukul 17.00 pada hari kerja; Rabu pukul 16.00 masih buka."),
		q("reading", "読解 · 情報検索", "市役所の窓口\n住民票：月〜金 8:30〜17:00\n市民税：月〜金 9:00〜16:30\nパスポート：月・水・金 9:00〜16:00\n土・日・祝日は休み。番号札を取ってお待ちください。", "パスポートの手続きをしたいです。窓口が開いている曜日はいつですか。", [4]string{"月曜・水曜・金曜", "火曜・木曜", "土曜・日曜", "毎日"}, "Pengurusan paspor tersedia Senin, Rabu, dan Jumat."),
		q("reading", "読解 · 情報検索", "みどり日本語教室\n入門クラス：火曜日 18:30〜20:00\n初級クラス：土曜日 10:00〜11:30\n参加費：一回500円\n初めての人は、授業の15分前に来て、申込書を書いてください。", "初めて土曜日の初級クラスに参加します。何時に来ますか。", [4]string{"午前九時四十五分", "午前十時十五分", "午後六時十五分", "午前十時"}, "Kelas mulai pukul 10.00, peserta baru diminta tiba 15 menit lebih awal: 9.45."),
		q("reading", "読解 · 情報検索", "みどり日本語教室\n入門クラス：火曜日 18:30〜20:00\n初級クラス：土曜日 10:00〜11:30\n参加費：一回500円\n初めての人は、授業の15分前に来て、申込書を書いてください。", "初級クラスに三回参加すると、いくらですか。", [4]string{"1,500円", "500円", "1,000円", "2,000円"}, "Biaya satu kali 500 yen; tiga kali adalah 1.500 yen."),
		q("reading", "読解 · 情報検索", "さくら駅前ホテル\nチェックイン：15:00から／チェックアウト：10:00まで\n朝食：6:30〜9:00（一階レストラン）\n荷物預かり：チェックイン前も無料\n全館禁煙です。喫煙所は入口の外にあります。", "朝ご飯を食べたいです。何時までに一階へ行きますか。", [4]string{"午前九時まで", "午前十時まで", "午後三時まで", "午前六時まで"}, "Sarapan tersedia sampai pukul 9 pagi."),
		q("reading", "読解 · 情報検索", "さくら駅前ホテル\nチェックイン：15:00から／チェックアウト：10:00まで\n朝食：6:30〜9:00（一階レストラン）\n荷物預かり：チェックイン前も無料\n全館禁煙です。喫煙所は入口の外にあります。", "チェックインの前に荷物を預けたいです。料金はいくらですか。", [4]string{"無料です。", "500円です。", "1,000円です。", "預けられません。"}, "Informasi menyatakan penitipan sebelum check-in gratis."),
		q("reading", "読解 · 情報検索", "国際交流センター 日本語相談\n火曜日 13:00〜16:00／土曜日 10:00〜12:00\n相談無料・予約不要\n住所や仕事の書類を相談する人は、関係する書類を持って来てください。電話相談はありません。", "土曜日に相談したいです。何時に行けばいいですか。", [4]string{"午前十時から正午まで", "午後一時から四時まで", "午前九時から十時まで", "土曜日は相談できません"}, "Sesi Sabtu berlangsung pukul 10.00 sampai 12.00."),
		q("reading", "読解 · 情報検索", "国際交流センター 日本語相談\n火曜日 13:00〜16:00／土曜日 10:00〜12:00\n相談無料・予約不要\n住所や仕事の書類を相談する人は、関係する書類を持って来てください。電話相談はありません。", "仕事の書類について相談するとき、何を持って行きますか。", [4]string{"関係する書類", "予約の電話", "ホテルの鍵", "日本語の宿題だけ"}, "Pemohon disarankan membawa dokumen yang terkait dengan pertanyaan mereka."),
	}
}
