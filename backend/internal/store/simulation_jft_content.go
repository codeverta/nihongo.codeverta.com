package store

// Original written JFT-Basic practice. Listening is explicitly excluded.
func simulationJFTContent() []simulationItem {
	type question struct {
		prompt      string
		choices     [4]string
		explanation string
	}
	items := make([]simulationItem, 0, 36)
	add := func(section, group, passage string, questions []question) {
		for _, q := range questions {
			items = append(items, simulationItem{section, group, "最もよい答えを一つ選んでください。", passage, q.prompt, q.choices, q.explanation})
		}
	}
	add("vocabulary", "文字と語彙 · Bacaan kanji", "", []question{
		{"明日は[[休み]]です。下線の言葉はどう読みますか。", [4]string{"やすみ", "たのみ", "すみ", "のみ"}, "休み（やすみ） berarti libur atau istirahat. 明日は休みです = besok libur."},
		{"この[[薬]]は食事の後に飲んでください。下線の言葉はどう読みますか。", [4]string{"くすり", "くつ", "こめ", "かさ"}, "薬 dibaca くすり, obat. Petunjuknya adalah minum sesudah makan."},
		{"駅の[[出口]]で待っています。下線の言葉はどう読みますか。", [4]string{"でぐち", "いりぐち", "でくち", "しゅつこう"}, "出口（でぐち） = pintu keluar. 入口（いりぐち） = pintu masuk."},
		{"[[住所]]をここに書いてください。下線の言葉はどう読みますか。", [4]string{"じゅうしょ", "じゅしょ", "じゅうそう", "じしょ"}, "住所（じゅうしょ） adalah alamat tempat tinggal. Bunyi じゅう panjang."},
	})
	add("vocabulary", "文字と語彙 · Makna", "", []question{
		{"雨が降っています。外へ行くとき、何を持っていきますか。", [4]string{"傘", "枕", "皿", "鍵盤"}, "傘（かさ） = payung untuk hujan. 枕 = bantal, 皿 = piring, 鍵盤 = tuts."},
		{"店でお金を払った後、買った物と値段が書いてある紙をもらいました。この紙は何ですか。", [4]string{"レシート", "メニュー", "カレンダー", "パスポート"}, "レシート adalah struk pembayaran. メニュー berisi pilihan makanan, bukan bukti transaksi."},
		{"明日の午後三時に病院へ行きたいので、先に電話で時間を決めました。何をしましたか。", [4]string{"予約", "掃除", "洗濯", "引っ越し"}, "予約（よやく） = membuat janji atau reservasi sebelum datang."},
		{"電車の中で、かばんを座席に置いたまま降りてしまいました。何をしましたか。", [4]string{"忘れ物", "買い物", "散歩", "料理"}, "忘れ物（わすれもの） = barang tertinggal. Tas tertinggal di kereta."},
	})
	add("vocabulary", "文字と語彙 · Penggunaan", "", []question{
		{"この料理は味が薄いです。もう少し塩を（　）ください。", [4]string{"入れて", "閉めて", "消して", "脱いで"}, "塩を入れる berarti menambahkan garam. Bentuk permintaan: 入れてください."},
		{"部屋が暗いです。電気を（　）ください。", [4]string{"つけて", "止めて", "洗って", "返して"}, "電気をつける = menyalakan lampu. 電気を消す = mematikan lampu."},
		{"バスを待っていますが、まだ（　）。", [4]string{"来ません", "食べません", "着ません", "読みません"}, "Kendaraan datang: バスが来る. まだ来ません = belum datang."},
		{"紙とプラスチックは別々に（　）ください。", [4]string{"捨てて", "泳いで", "借りて", "住んで"}, "捨てる = membuang. 別々に berarti secara terpisah, sesuai petunjuk pemilahan sampah."},
	})
	add("conversation", "会話と表現 · Grammar", "", []question{
		{"A：ここで写真を撮って（　）ですか。\nB：はい、大丈夫です。", [4]string{"もいい", "はいけない", "いるつもり", "おきたい"}, "Vて＋もいいですか digunakan untuk meminta izin: boleh mengambil foto di sini?"},
		{"A：新しい仕事はどうですか。\nB：まだ慣れていませんが、毎日少しずつ日本語（　）話しています。", [4]string{"で", "へ", "から", "まで"}, "Bahasa yang digunakan ditandai で: 日本語で話す = berbicara dalam bahasa Jepang."},
		{"A：お昼はもう食べましたか。\nB：いいえ、まだ（　）。", [4]string{"食べていません", "食べました", "食べたことがあります", "食べすぎました"}, "まだ＋Vていません = belum melakukan. Pertanyaan menanyakan makan siang hari ini."},
		{"A：明日、一緒に買い物に行きませんか。\nB：すみません。明日は病院へ行く（　）があります。", [4]string{"予定", "よう", "ながら", "だけ"}, "V辞書形＋予定がある = ada rencana. Penutur memiliki janji ke rumah sakit."},
		{"A：この機械の使い方を教えてください。\nB：まず、このボタンを（　）ください。", [4]string{"押して", "押すて", "押したて", "押していると"}, "押す menjadi 押して dalam bentuk て. 押してください = silakan tekan."},
		{"A：昨日、どうして早く帰りましたか。\nB：頭が痛かった（　）です。", [4]string{"から", "まで", "より", "だけ"}, "Alasan dijelaskan dengan ～からです: karena kepala saya sakit."},
	})
	add("conversation", "会話と表現 · Ungkapan", "", []question{
		{"店で、商品の値段が分かりません。店員に何と言いますか。", [4]string{"すみません、これはいくらですか。", "お先に失礼します。", "いただきます。", "お帰りなさい。"}, "これはいくらですか menanyakan harga. Ungkapan lain tidak meminta informasi harga."},
		{"職場に五分遅れそうです。電話で何と言いますか。", [4]string{"すみません、五分ほど遅れます。", "五分早く着きました。", "今日はもう帰りました。", "明日は休んでもいいですか。"}, "Beritahukan keterlambatan yang akan terjadi: 五分ほど遅れます = terlambat sekitar lima menit."},
		{"相手の話すスピードが速くて、よく分かりません。何と言いますか。", [4]string{"もう少しゆっくり話していただけますか。", "もっと速く話してください。", "ここに名前を書いてください。", "お会計をお願いします。"}, "ゆっくり話していただけますか = bisakah berbicara lebih pelan? Permintaan ini sesuai kendala kecepatan."},
		{"友達に食事へ誘われましたが、その日は用事があります。どう答えますか。", [4]string{"すみません、その日は用事があるので、また今度お願いします。", "ぜひ行きます。その日は何もありません。", "おかわりをお願いします。", "こちらこそ、初めまして。"}, "Menolak undangan dengan sopan: jelaskan ada keperluan dan tawarkan kesempatan lain."},
		{"家の水道から水が出ません。管理会社に何と言いますか。", [4]string{"水が出ないので、見に来ていただけますか。", "水をもう一杯ください。", "今日は水曜日ですか。", "この水は冷たくておいしいです。"}, "Jelaskan kerusakan dan minta pemeriksaan: 水が出ない = air tidak keluar."},
		{"同僚が仕事を手伝ってくれました。何と言いますか。", [4]string{"手伝ってくれて、ありがとうございます。", "ご注文はお決まりですか。", "いってらっしゃい。", "お大事に。"}, "Mengucapkan terima kasih atas bantuan: 手伝ってくれてありがとうございます. お大事に ditujukan pada orang sakit."},
	})
	add("reading", "読解 · Pesan dan pengumuman", "田中さんへ\n明日の朝、八時半に工場の入口に来てください。いつもより三十分早いです。新しい機械の説明があります。作業着は会社で用意しますが、自分の安全靴を持ってきてください。\n山本", []question{
		{"田中さんは明日、何時にどこへ行きますか。", [4]string{"八時半に工場の入口へ行きます。", "九時に工場の入口へ行きます。", "八時半に駅へ行きます。", "九時に会社の食堂へ行きます。"}, "Pesan menyebut 明日の朝、八時半に工場の入口. Waktu biasanya lebih lambat, tetapi besok 08.30."},
		{"田中さんが自分で持っていく物は何ですか。", [4]string{"安全靴", "作業着", "新しい機械", "昼ご飯の材料"}, "安全靴 harus dibawa sendiri, sedangkan 作業着 disediakan perusahaan."},
	})
	add("reading", "読解 · Pesan dan pengumuman", "ごみの出し方\n燃えるごみ：月曜・木曜\n缶とびん：水曜\nごみは指定の袋に入れて、朝八時までに集積所へ出してください。前の夜に出さないでください。缶とびんは中を水で洗ってから出してください。", []question{
		{"缶とびんは、いつ出しますか。", [4]string{"水曜の朝八時まで", "木曜の夜八時まで", "月曜の前の夜", "毎日いつでも"}, "Jadwal kaleng dan botol adalah Rabu, sebelum atau sampai pukul delapan pagi. Malam sebelumnya dilarang."},
		{"缶とびんを出す前に、何をしますか。", [4]string{"中を水で洗います。", "中に紙を入れます。", "食べ物と一緒にします。", "前の夜に外へ置きます。"}, "Petunjuk 洗ってから出す berarti bilas bagian dalam dengan air sebelum membuang."},
	})
	add("reading", "読解 · Pesan dan pengumuman", "みなさんへ\n今週の日本語教室は先生の都合でお休みです。来週はいつもどおり日曜日の午前十時からです。宿題は教科書の二十ページです。分からない問題は、次の授業で一緒に確認しましょう。", []question{
		{"日本語教室は次にいつありますか。", [4]string{"来週の日曜の午前十時", "今週の日曜の午前十時", "来週の土曜の午後十時", "明日の午前二十時"}, "Minggu ini libur. Pertemuan berikutnya Minggu pekan depan pukul 10 pagi."},
		{"宿題で分からない問題があるとき、どうしますか。", [4]string{"次の授業で一緒に確認します。", "教科書を捨てます。", "今週の授業へ行きます。", "先生の家へ必ず行きます。"}, "Pengumuman meminta soal yang belum dipahami diperiksa bersama di pelajaran berikutnya."},
	})
	add("reading", "読解 · Pencarian informasi", "さくら食堂　ランチメニュー\nA 魚定食：700円（魚・ご飯・みそ汁）\nB 鶏肉定食：750円（鶏肉・ご飯・みそ汁）\nC 野菜カレー：650円（肉なし）\nD ラーメン：600円（豚肉入り）\nご飯の大盛りは100円追加です。定食だけ、ご飯の量を選べます。ランチは11:00〜14:00。火曜日は休みです。", []question{
		{"肉を食べない人が、魚も使っていない料理を選びます。どれがいいですか。", [4]string{"C 野菜カレー", "A 魚定食", "B 鶏肉定食", "D ラーメン"}, "C menyebut 肉なし dan tidak mencantumkan ikan. A ikan, B ayam, D babi."},
		{"魚定食でご飯を大盛りにすると、いくらですか。", [4]string{"800円", "700円", "750円", "850円"}, "魚定食 700円 ditambah 大盛り 100円 menjadi 800円."},
	})
	add("reading", "読解 · Pencarian informasi", "市立図書館\n開館：火曜〜金曜 9:00〜19:00／土曜・日曜 10:00〜17:00\n休館：月曜\n本は一人五冊まで、二週間借りられます。初めて借りる人は、住所が分かる身分証明書を持って、受付でカードを作ってください。本を返すとき、図書館が閉まっていたら、入口の返却ボックスが使えます。", []question{
		{"初めて本を借りたい人は、何を持っていきますか。", [4]string{"住所が分かる身分証明書", "五冊の自分の本", "学校の宿題だけ", "料理のメニュー"}, "Kartu peminjaman dibuat dengan identitas yang menunjukkan alamat. Tidak diminta membawa buku sendiri."},
		{"月曜日に本を返したいとき、どうしますか。", [4]string{"入口の返却ボックスに入れます。", "受付で新しいカードを作ります。", "午後七時まで中で待ちます。", "本を二週間追加で必ず借ります。"}, "Senin tutup, tetapi tersedia kotak pengembalian di pintu masuk saat perpustakaan tutup."},
	})
	add("reading", "読解 · Pencarian informasi", "青葉駅から病院へのバス\n1番：市役所行き／9:00、9:30／病院には止まりません。\n2番：中央病院行き／9:10、9:40／病院まで20分／運賃200円\n3番：空港行き／9:15、9:45／病院には止まりません。\nバスは駅の東口から出ます。運賃は降りるときに払ってください。", []question{
		{"九時五分に駅に着きました。病院へ行くために、次にどのバスに乗りますか。", [4]string{"九時十分の2番のバス", "九時の1番のバス", "九時十五分の3番のバス", "九時半の1番のバス"}, "Bus berikutnya menuju rumah sakit adalah nomor 2 pukul 09.10. Nomor 1 dan 3 tidak berhenti di rumah sakit."},
		{"九時十分のバスに乗ったら、病院に何時ごろ着きますか。", [4]string{"九時半ごろ", "九時十分ごろ", "九時四十分ごろ", "十時ごろ"}, "Perjalanan 20 menit: 09.10 + 20 menit = 09.30."},
	})
	return items
}
