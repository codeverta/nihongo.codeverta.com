package store

import "strings"

type natSimulationQuestion struct {
	prompt  string
	choices [4]string
	why     string
}

func natAdditionalQuestions(paper int) []natSimulationQuestion {
	switch paper {
	case 2:
		return []natSimulationQuestion{
			{"[[近所]]の人にあいさつしました。", [4]string{"きんじょ", "きんしょ", "ちかじょ", "こんじょ"}, "近所（きんじょ） berarti lingkungan sekitar."},
			{"この道は[[危険]]ですから、気をつけてください。", [4]string{"きけん", "きげん", "きれん", "あぶけん"}, "危険（きけん） berarti berbahaya."},
			{"姉は料理がとても（　）です。", [4]string{"上手", "丈夫", "大事", "親切"}, "上手（じょうず） berarti mahir."},
			{"手紙に（　）を貼って、ポストに入れます。", [4]string{"切手", "切符", "切身", "切口"}, "切手（きって） adalah perangko yang ditempel di surat."},
			{"雨で服が（　）しまいました。", [4]string{"ぬれて", "こわれて", "おれて", "やせて"}, "ぬれる berarti menjadi basah."},
			{"駅の階段を上がると、出口が（　）にあります。", [4]string{"左", "低", "遠", "弱"}, "左（ひだり） berarti kiri."},
			{"この店は新鮮な魚を毎朝（　）います。", [4]string{"売って", "買って", "借りて", "習って"}, "Toko menjual ikan: 魚を売る."},
			{"仕事のあとで、友達と食事を（　）ました。", [4]string{"しました", "いました", "なりました", "ありました"}, "食事をする berarti makan."},
			{"電車の中にかばんを（　）きました。", [4]string{"忘れて", "覚えて", "決めて", "調べて"}, "忘れる berarti lupa atau tertinggal."},
			{"この部屋は広くて、窓から山が（　）。", [4]string{"見えます", "聞こえます", "使えます", "直せます"}, "見える menyatakan sesuatu terlihat."},
			{"旅行のあと、家族に（　）を買いました。", [4]string{"お土産", "お見舞い", "お祝い", "お知らせ"}, "お土産 adalah oleh-oleh yang dibeli untuk keluarga setelah bepergian."},
			{"日本へ行く（　）、日本語を勉強しています。", [4]string{"ために", "そうに", "ところに", "ばかりに"}, "Tujuan yang disengaja: V辞書形 + ために."},
			{"この薬を飲めば、すぐ元気に（　）と思います。", [4]string{"なる", "する", "いる", "ある"}, "なる menyatakan perubahan menjadi sehat."},
			{"歩いて行くの（　）、バスで行くのと、どちらがいいですか。", [4]string{"と", "が", "を", "に"}, "のと membentuk frasa nomina untuk membandingkan dua cara."},
			{"ドアが開いていますから、入って（　）いいですよ。", [4]string{"も", "は", "が", "を"}, "～てもいい menyatakan izin."},
			{"この漢字は難しくて、私には（　）ません。", [4]string{"読め", "読み", "読む", "読ん"}, "Bentuk potensial negatif 読めません menyatakan tidak bisa membaca."},
			{"昨日買ったケーキは、もう全部食べて（　）。", [4]string{"しまいました", "おきました", "みました", "ありました"}, "～てしまう menyatakan tindakan selesai seluruhnya."},
			{"友達が駅まで迎えに来て（　）ので、助かりました。", [4]string{"くれた", "あげた", "もらった", "いただいた"}, "Teman melakukan kebaikan untuk saya: 来てくれた."},
			{"毎日練習すれば、漢字をもっと（　）ようになります。", [4]string{"覚えられる", "覚える", "覚えた", "覚えさせる"}, "Kemampuan yang akan berkembang: 覚えられるようになります."},
			{"この機械の使い方を（　）ください。", [4]string{"説明して", "説明する", "説明した", "説明し"}, "Permintaan memakai bentuk て + ください."},
			{"宿題をして（　）から、友達と遊びます。", [4]string{"から", "まで", "しか", "でも"}, "Urutan kegiatan: 宿題をしてから遊ぶ."},
			{"この町に来てから、まだ一度も雪を（　）ことがありません。", [4]string{"見た", "見る", "見て", "見ます"}, "Pengalaman negatif: 見たことがありません."},
			{"先生は学生に作文をもう一度書かせ（　）。", [4]string{"ました", "られました", "ていました", "ようでした"}, "Kausatif lampau: 書かせました, menyuruh siswa menulis lagi."},
		}
	case 3:
		return []natSimulationQuestion{
			{"[[景色]]がきれいな場所で写真を撮りました。", [4]string{"けしき", "けいしょく", "けしょく", "けいき"}, "景色（けしき） berarti pemandangan."},
			{"空港で飛行機の時間を（　）ました。", [4]string{"調べ", "集め", "並べ", "比べ"}, "調べる berarti memeriksa atau mencari informasi."},
			{"旅行の前にホテルを（　）しました。", [4]string{"予約", "用意", "案内", "相談"}, "Hotel dipesan: 予約する."},
			{"道に迷ったので、交番で（　）ました。", [4]string{"聞き", "呼び", "売り", "借り"}, "聞く berarti bertanya atau mendengar."},
			{"[[荷物]]を運ぶのを手伝ってください。", [4]string{"にもつ", "にぶつ", "にもじ", "にもの"}, "荷物（にもつ） berarti barang bawaan."},
			{"旅行の写真をアルバムに（　）ました。", [4]string{"入れ", "出し", "消し", "乗り"}, "Foto dimasukkan ke album: 入れる."},
			{"ホテルの部屋から海がよく（　）。", [4]string{"見えます", "見せます", "見ます", "見られますか"}, "見える berarti terlihat dari suatu tempat."},
			{"電車をホームで（　）ました。", [4]string{"待ち", "渡り", "通り", "曲がり"}, "駅のホームで電車を待つ: 待つ berarti menunggu."},
			{"この町の[[名物]]は何ですか。", [4]string{"めいぶつ", "みょうぶつ", "めいもの", "みょうもの"}, "名物（めいぶつ） berarti produk atau makanan khas suatu tempat."},
			{"旅行中に財布を（　）ました。", [4]string{"なくし", "見つけ", "預け", "返し"}, "なくす berarti kehilangan; 財布をなくしました = saya kehilangan dompet."},
			{"山の上は空気が（　）です。ごみも少ないです。", [4]string{"きれい", "きれく", "きれいな", "きれいに"}, "きれい adalah na-adjective yang berarti bersih."},
			{"友達はもう空港に着いた（　）です。", [4]string{"そう", "よう", "らしいな", "みたいに"}, "Kabar yang didengar: 着いたそうです."},
			{"雨が降っていた（　）、試合は予定どおり行われました。", [4]string{"のに", "ので", "から", "ため"}, "のに menyatakan hasil berlawanan dengan dugaan hujan seharusnya mengganggu pertandingan."},
			{"切符を買って（　）から、改札を通ってください。", [4]string{"から", "まで", "だけ", "しか"}, "Urutan tindakan memakai Vてから."},
			{"旅行の写真を見せて（　）ませんか。", [4]string{"もらえ", "あげ", "くれ", "いただき"}, "Permintaan bantuan sopan: 見せてもらえませんか."},
			{"ホテルに着いた（　）、フロントで名前を伝えてください。", [4]string{"とき", "まで", "しか", "ほど"}, "「着いたとき」は到着した時点を表します。"},
			{"雨が降りそうなので、出かける前に傘を持って（　）。", [4]string{"いきます", "きます", "あります", "しまいます"}, "出かける人が傘を持参するため、持っていきますが自然です。"},
			{"旅行会社に電話して、出発時間をもう一度確かめて（　）。", [4]string{"みました", "おきました", "しまいました", "ありました"}, "試しに確認する意味の「～てみる」: 確かめてみました。"},
			{"電車の中で大きな声で話さない（　）しています。", [4]string{"ように", "ために", "そうに", "ことが"}, "Kebiasaan/berusaha menghindari: 話さないようにしています."},
			{"この切符は、明日まで使う（　）ができます。", [4]string{"こと", "もの", "ところ", "ため"}, "Kemungkinan atau izin: 使うことができます."},
			{"先生が来る（　）、ここで待っていてください。", [4]string{"まで", "から", "しか", "ほど"}, "まで menunjukkan batas waktu menunggu sampai guru datang."},
			{"窓を閉めても、外の音が（　）聞こえます。", [4]string{"まだ", "もう", "ぜひ", "たぶん"}, "まだ聞こえる = suaranya masih terdengar."},
			{"ホテルを予約して（　）、旅行の日を決めました。", [4]string{"から", "まで", "しか", "ほど"}, "Urutan tindakan dinyatakan dengan Vてから."},
		}
	case 4:
		return []natSimulationQuestion{
			{"[[台風]]が近づいているので、早く帰ります。", [4]string{"たいふう", "だいふう", "たいぷう", "だいぷう"}, "台風（たいふう） berarti topan."},
			{"駅の店で旅行の（　）を買いました。", [4]string{"切符", "切手", "切身", "切口"}, "切符（きっぷ） berarti tiket."},
			{"明日は大切な試験なので、今晩は早く（　）ます。", [4]string{"寝", "起き", "休み", "止まり"}, "寝る berarti tidur."},
			{"約束の時間に[[遅刻]]しました。", [4]string{"ちこく", "ちごく", "ちこう", "じこく"}, "遅刻（ちこく） berarti terlambat datang."},
			{"使った道具を元の場所に（　）ください。", [4]string{"戻して", "返して", "直して", "届けて"}, "元の場所に戻す = mengembalikan ke tempat semula."},
			{"風が強くて、帽子が遠くまで（　）。", [4]string{"飛びました", "落としました", "消しました", "折りました"}, "帽子が飛ぶ berarti topi terbang tertiup angin."},
			{"店員にサイズを（　）から、服を買いました。", [4]string{"確かめて", "決まって", "続けて", "伝わって"}, "確かめる berarti memastikan atau mengecek."},
			{"このカレーは辛くて、水が（　）です。", [4]string{"必要", "安全", "有名", "静か"}, "必要（ひつよう） berarti diperlukan; pedasnya kari membuat air diperlukan."},
			{"駅の近くに新しい病院が（　）ました。", [4]string{"でき", "なり", "おき", "あい"}, "建物ができる berarti bangunan baru selesai/terbentuk."},
			{"春になると、この公園に花がたくさん（　）。", [4]string{"咲きます", "鳴ります", "焼きます", "建てます"}, "咲く dipakai untuk bunga mekar."},
			{"今日は雲がなく、空がとても（　）です。", [4]string{"青い", "辛い", "弱い", "浅い"}, "空が青い berarti langit berwarna biru."},
			{"窓を開けたまま出かけて（　）。", [4]string{"しまいました", "おきました", "みました", "ありました"}, "～てしまう menunjukkan akibat yang tidak diinginkan."},
			{"雨が降る（　）もしれませんから、傘を持っていきます。", [4]string{"か", "と", "なら", "ても"}, "Kemungkinan: ～かもしれません."},
			{"道が混んでいた（　）、バスが遅れました。", [4]string{"ので", "のに", "でも", "なら"}, "Alasan keterlambatan dijelaskan dengan ので."},
			{"この店は値段が安い（　）、品物の種類も多いです。", [4]string{"し", "ので", "のに", "だけ"}, "～し menambahkan alasan atau keunggulan."},
			{"書類を出すときは、本人確認のカードを（　）ください。", [4]string{"持ってきて", "持っていって", "持っていて", "持ってみて"}, "持ってくる digunakan saat meminta orang membawa sesuatu ke tempat pembicara."},
			{"六名以上で予約する場合は、前日（　）に電話してください。", [4]string{"まで", "までに", "までで", "までも"}, "Batas waktu tindakan memakai 前日までに."},
			{"午後から雨が降るそうです。洗濯物は昼までに（　）ほうがいいでしょう。", [4]string{"取り込んだ", "取り込む", "取り込んで", "取り込まない"}, "Saran memakai bentuk lampau + ほうがいい: 取り込んだほうがいい."},
			{"窓の外が暗くなって（　）から、電気をつけました。", [4]string{"きた", "いった", "あった", "しまった"}, "Perubahan menuju keadaan kini: 暗くなってきた."},
			{"この荷物は大きすぎるので、宅配便で（　）ことにしました。", [4]string{"送る", "送って", "送った", "送られる"}, "Keputusan: 辞書形 + ことにする."},
			{"雨がやんだ（　）、空に虹が出ました。", [4]string{"あとで", "まえに", "うちに", "ながら"}, "Setelah hujan berhenti: やんだあとで."},
			{"説明書を読んでも、使い方がよく（　）ません。", [4]string{"分かり", "知り", "思い", "考え"}, "分かりません menyatakan belum paham."},
			{"この店では、買った商品を一週間以内なら交換する（　）ができます。", [4]string{"こと", "もの", "ため", "ところ"}, "Aturan/kemungkinan memakai ～ことができます."},
		}
	case 5:
		return []natSimulationQuestion{
			{"[[神社]]でおまつりがありました。", [4]string{"じんじゃ", "しんしゃ", "じんしゃ", "かみやしろ"}, "神社（じんじゃ） berarti kuil Shinto."},
			{"日本の文化について本で（　）ました。", [4]string{"学び", "習い", "教え", "働き"}, "学ぶ berarti belajar atau menelaah."},
			{"この町は古い建物が多くて、とても（　）があります。", [4]string{"歴史", "予定", "経験", "生活"}, "Bangunan tua memberi kesan memiliki sejarah: 歴史があります."},
			{"お祭りで太鼓の音が大きく（　）ました。", [4]string{"聞こえ", "見え", "思い", "考え"}, "聞こえる berarti terdengar."},
			{"家族におみやげを（　）ました。", [4]string{"選び", "決まり", "集まり", "変わり"}, "選ぶ berarti memilih."},
			{"鳥居を通って、神社の（　）へ入りました。", [4]string{"境内", "教室", "工場", "台所"}, "境内（けいだい） adalah area di dalam kompleks kuil."},
			{"説明を聞いてから、展示を（　）ました。", [4]string{"見", "聞き", "読み", "会い"}, "Pameran dilihat: 展示を見る."},
			{"お寺に入る前に、入口で手を（　）ます。", [4]string{"洗い", "浴び", "磨き", "拭き"}, "手を洗う berarti mencuci tangan."},
			{"祭りの日は、町の広場に人がたくさん（　）。", [4]string{"集まります", "集めます", "集められます", "集めさせます"}, "人が集まる berarti orang-orang berkumpul."},
			{"祖父は若いころ、この町で（　）をしていました。", [4]string{"大工", "台風", "景色", "切符"}, "大工（だいく） adalah pekerjaan tukang kayu."},
			{"祭りで太鼓を（　）人を見ました。", [4]string{"たたく", "弾く", "吹く", "歌う"}, "太鼓をたたく adalah kolokasi untuk memainkan taiko."},
			{"人が多かったので、前のほうからは何も（　）。", [4]string{"見えませんでした", "見ませんでした", "見せませんでした", "見られたいです"}, "Sesuatu tidak terlihat dari posisi tertentu: 見えませんでした."},
			{"古い町並みを残す（　）、市民が活動しています。", [4]string{"ために", "そうに", "ことに", "ところに"}, "Tujuan yang diusahakan: ～ために."},
			{"説明が難しかったので、もう一度話して（　）ました。", [4]string{"もらい", "あげ", "くれ", "いただけ"}, "Menerima bantuan: 話してもらいました."},
			{"展示品にさわらない（　）お願いします。", [4]string{"ように", "ために", "そうに", "ことへ"}, "Permintaan agar tidak menyentuh: ～ないようにお願いします."},
			{"祭りの日に雨が降った（　）、会場を公民館に変えます。", [4]string{"場合は", "のに", "ながら", "ほど"}, "Kondisi yang mungkin terjadi memakai「～場合は」."},
			{"発表会が始まる前に、資料を先生に見て（　）。", [4]string{"もらいました", "あげました", "くれました", "いただきますか"}, "Menerima bantuan dari guru dinyatakan dengan「見てもらいました」."},
			{"町の歴史を知る（　）、博物館へ行きました。", [4]string{"ために", "のに", "そうに", "まで"}, "Tujuan melakukan tindakan memakai「辞書形 + ために」."},
			{"この資料は来週の会議で使うので、コピーして（　）ください。", [4]string{"おいて", "しまって", "みて", "あって"}, "Persiapan untuk nanti: コピーしておく."},
			{"祭りの歴史を調べる（　）、図書館へ行きました。", [4]string{"ために", "のに", "そうに", "まで"}, "Tujuan melakukan tindakan memakai「辞書形 + ために」."},
			{"約束の時間に遅れそうなときは、先に連絡する（　）です。", [4]string{"つもり", "ところ", "ばかり", "はず"}, "Niat atau rencana: 連絡するつもりです."},
			{"この町には、百年以上前に建てられた家がまだ（　）います。", [4]string{"残って", "残して", "残り", "残らせて"}, "Bangunan masih ada: 残っている."},
			{"写真を撮るとき、ほかの人の迷惑に（　）ようにしてください。", [4]string{"ならない", "しない", "いない", "できない"}, "Ungkapan: 迷惑にならないようにする."},
		}
	}
	return nil
}

func natBaseGrammarIndexes(paper int) map[int]bool {
	indexes := map[int]bool{}
	var grammar []int
	switch paper {
	case 2:
		grammar = []int{3, 4, 5, 6, 7, 8, 9}
	case 3:
		grammar = []int{2, 5, 6, 7, 8, 9}
	case 4:
		grammar = []int{2, 3, 4, 5, 6, 7}
	case 5:
		grammar = []int{7, 8, 9}
	}
	for _, index := range grammar {
		indexes[index] = true
	}
	return indexes
}

func natReplacementVocabulary(paper int) []natSimulationQuestion {
	switch paper {
	case 2:
		return []natSimulationQuestion{
			{"[[病院]]の受付で診察券を出しました。", [4]string{"うけつけ", "うけづけ", "うけつき", "うけつ"}, "受付（うけつけ） berarti loket atau meja pendaftaran."},
			{"この料理はとても（　）です。砂糖を入れすぎました。", [4]string{"甘い", "苦い", "辛い", "酸っぱい"}, "甘い（あまい） berarti manis; gula berlebih menjadi petunjuk."},
			{"電車の中で大きな声を出すのは（　）です。", [4]string{"迷惑", "用事", "約束", "景色"}, "迷惑 berarti gangguan atau ketidaknyamanan bagi orang lain."},
			{"兄は毎朝、新聞を（　）から会社へ行きます。", [4]string{"読み", "飲み", "書き", "聞き"}, "新聞を読む berarti membaca koran."},
			{"道が分からないので、地図を（　）います。", [4]string{"見て", "聞いて", "呼んで", "待って"}, "Peta dilihat: 地図を見る."},
			{"試験の答えを答案用紙に（　）ください。", [4]string{"書いて", "消して", "貸して", "習って"}, "Jawaban ditulis pada lembar jawaban: 答えを書く."},
			{"この店は魚が新しくて、値段も（　）です。", [4]string{"安い", "深い", "薄い", "遅い"}, "値段が安い berarti harganya murah."},
		}
	case 3:
		return []natSimulationQuestion{
			{"[[港]]から船で島へ渡りました。", [4]string{"みなと", "みなみ", "こう", "みなとぐち"}, "港（みなと） berarti pelabuhan."},
			{"旅行の荷物をスーツケースに（　）ました。", [4]string{"入れ", "消し", "折り", "比べ"}, "Barang dimasukkan ke koper: 荷物を入れる."},
			{"駅で買ったお弁当を新幹線の中で（　）。", [4]string{"食べました", "作りました", "洗いました", "直しました"}, "お弁当を食べる adalah kolokasi yang tepat."},
			{"このホテルは駅に近くて、とても（　）です。", [4]string{"便利", "不便", "危険", "残念"}, "便利 berarti praktis; dekat stasiun mendukung makna ini."},
			{"山の頂上から町の景色を（　）ました。", [4]string{"眺め", "集め", "決め", "止め"}, "眺める berarti memandang pemandangan."},
			{"旅行の記念に、友達へ小さな（　）を買いました。", [4]string{"お土産", "お見舞い", "お祝い", "お知らせ"}, "お土産 berarti oleh-oleh perjalanan."},
		}
	case 4:
		return []natSimulationQuestion{
			{"[[風邪]]をひいたので、今日は学校を休みます。", [4]string{"かぜ", "ふうじゃ", "かざ", "ふうせ"}, "風邪（かぜ） berarti selesma atau flu."},
			{"この店は駅前にあって、とても（　）です。", [4]string{"便利", "不便", "危ない", "苦手"}, "便利 berarti praktis; lokasi di depan stasiun menjadi petunjuk."},
			{"店員に聞いたら、商品の値段がよく（　）ました。", [4]string{"分かり", "始まり", "渡り", "止まり"}, "Harga menjadi jelas: 値段が分かる."},
			{"このスープは塩を入れすぎて、とても（　）です。", [4]string{"しょっぱい", "にがい", "すっぱい", "からい"}, "しょっぱい berarti asin; terlalu banyak garam menjadi petunjuk."},
			{"雨の日は道が（　）ので、気をつけてください。", [4]string{"滑りやすい", "深い", "親切", "有名"}, "滑りやすい berarti mudah licin; hujan membuat jalan licin."},
			{"地下の駐車場へは、エレベーターで（　）ました。", [4]string{"降り", "上がり", "通り", "着き"}, "降りる berarti turun; tujuan berada di lantai bawah tanah."},
		}
	case 5:
		return []natSimulationQuestion{
			{"[[伝統]]的な踊りを見ました。", [4]string{"でんとう", "でんどう", "てんとう", "てんどう"}, "伝統（でんとう） berarti tradisi."},
			{"祭りの会場で、地元の（　）を食べました。", [4]string{"名物", "目的", "習慣", "景色"}, "名物 berarti hidangan atau produk khas suatu tempat."},
			{"お祭りでは、太鼓の（　）が町中に聞こえました。", [4]string{"音", "色", "味", "形"}, "音（おと） berarti suara."},
		}
	}
	return nil
}

func natPaperThree() []natSimulationQuestion {
	return []natSimulationQuestion{
		{"[[空港]]まで電車で行きます。", [4]string{"くうこう", "くうこ", "そらみなと", "からこう"}, "空港（くうこう） berarti bandara."},
		{"あの人は私の[[妹]]です。", [4]string{"いもうと", "おとうと", "あね", "おば"}, "妹（いもうと） adalah adik perempuan."},
		{"毎晩、日記にその日の出来事を（　）から寝ます。", [4]string{"書いて", "買って", "歩いて", "泳いで"}, "日記に出来事を書く berarti menulis peristiwa di buku harian."},
		{"このシャツは大きいです。もっと（　）シャツを見せてください。", [4]string{"小さい", "細い", "短い", "低い"}, "Pakaian yang lebih kecil: 小さい."},
		{"会議は三時に（　）ます。", [4]string{"始まり", "終わり", "決まり", "集まり"}, "会議が始まる berarti rapat dimulai."},
		{"兄は毎朝、公園を（　）います。", [4]string{"走って", "作って", "待って", "売って"}, "走る berarti berlari."},
		{"窓を（　）ください。部屋が暑いです。", [4]string{"開けて", "閉めて", "消して", "つけて"}, "Membuka jendela membantu mendinginkan ruangan."},
		{"この漢字の読み方を（　）ください。", [4]string{"教えて", "習って", "覚えて", "考えて"}, "教えてください adalah permintaan untuk memberi tahu."},
		{"日曜日はどこ（　）行きませんでした。", [4]string{"へも", "へが", "へを", "へで"}, "どこへも + bentuk negatif berarti tidak pergi ke mana pun."},
		{"この薬を飲んで、早く寝た（　）がいいです。", [4]string{"ほう", "もの", "ため", "こと"}, "Saran menggunakan ～たほうがいい."},
		{"日本へ行く前に、日本語を勉強して（　）。", [4]string{"おきます", "しまいます", "みますか", "ありますか"}, "Persiapan sebelumnya memakai ～ておく."},
		{"電車が遅れたので、授業に（　）。", [4]string{"間に合いませんでした", "間に合います", "間に合わせます", "間に合いたい"}, "遅れた akibatnya terlambat; bentuk lampau negatif cocok."},
		{"この町は前より便利に（　）。", [4]string{"なりました", "しました", "ありました", "いました"}, "なります menyatakan perubahan keadaan."},
		{"先生に質問したいことが（　）。", [4]string{"あります", "います", "します", "できますか"}, "Hal yang ingin ditanyakan adalah benda/perkara: あります."},
		{"説明を聞きましたが、まだよく（　）。", [4]string{"分かりません", "知りませんでしたか", "分かりたい", "知っています"}, "Belum paham: まだよく分かりません."},
		{"山のホテルは駅から遠いですが、毎朝駅までバスが出ています。朝食は七時から九時までです。", [4]string{"ホテルから駅までバスがあります。", "朝食は十時からです。", "ホテルは駅の隣です。", "朝食は夜にあります。"}, "Teks menyebut bus dari hotel ke stasiun setiap pagi."},
		{"市民プール：水曜休み。大人600円、子ども300円。水泳帽が必要です。受付で150円で買えます。", [4]string{"水曜日はプールが休みです。", "大人は300円です。", "水泳帽は無料です。", "子どもは600円です。"}, "Pengumuman menyatakan kolam tutup pada Rabu."},
		{"ユキさんへ。借りた本を返しに図書館へ行きます。四時ごろ戻ります。冷蔵庫にケーキがありますから、食べてください。母", [4]string{"母は本を返しに出かけています。", "ユキさんが本を借ります。", "母は三時に戻ります。", "ケーキはありません。"}, "Ibu pergi ke perpustakaan untuk mengembalikan buku."},
		{"旅行の日、朝は雨でした。駅で友達と会い、電車に乗りました。昼ごろ晴れたので、海辺を散歩して写真を撮りました。", [4]string{"昼ごろ、海辺を散歩しました。", "朝、海で泳ぎました。", "駅で雨がやみました。", "友達に会わずに出発しました。"}, "Teks menyatakan jalan-jalan di tepi laut saat cuaca cerah sekitar siang."},
		{"料理教室は金曜日の午後六時から八時までです。参加する人は、エプロンと筆記用具を持ってきてください。材料はこちらで用意します。", [4]string{"参加者はエプロンを持っていきます。", "教室は土曜日です。", "材料を自分で全部買います。", "午後八時に始まります。"}, "Peserta diminta membawa celemek dan alat tulis."},
	}
}

func natPaperFour() []natSimulationQuestion {
	return []natSimulationQuestion{
		{"[[図書館]]で本を借りました。", [4]string{"としょかん", "としょがん", "ずしょかん", "としょうかん"}, "図書館（としょかん） berarti perpustakaan."},
		{"今日は[[特別]]な日です。", [4]string{"とくべつ", "とくべち", "どくべつ", "とくべっ"}, "特別（とくべつ） berarti khusus."},
		{"部屋が暗いので、電気を（　）ください。", [4]string{"つけて", "けして", "あけて", "しめて"}, "Lampu dinyalakan: 電気をつける."},
		{"この道は車が多くて（　）です。", [4]string{"危ない", "優しい", "正しい", "楽しい"}, "危ない berarti berbahaya."},
		{"試験の前に、名前を（　）ください。", [4]string{"書いて", "消して", "切って", "貸して"}, "Instruksi meminta menulis nama."},
		{"父は新聞を（　）います。", [4]string{"読んで", "飲んで", "呼んで", "選んで"}, "新聞を読む berarti membaca koran."},
		{"この荷物をあそこまで（　）ください。", [4]string{"運んで", "選んで", "並んで", "休んで"}, "運ぶ berarti mengangkut atau membawa barang."},
		{"駅の前で友達と（　）ました。", [4]string{"会い", "合い", "開い", "洗い"}, "友達に会う berarti bertemu teman."},
		{"私は辛い食べ物があまり（　）。", [4]string{"好きではありません", "好きでした", "好きになります", "好きでしょう"}, "あまり lazim berpasangan dengan bentuk negatif."},
		{"雨が降りそうですから、かさを持って（　）。", [4]string{"行きます", "来ますか", "います", "あります"}, "持って行く berarti membawa pergi."},
		{"母は料理を（　）ながら、音楽を聞きます。", [4]string{"作り", "作る", "作って", "作った"}, "Dua kegiatan bersamaan memakai Vます-stem + ながら."},
		{"この部屋では靴を（　）なくてもいいです。", [4]string{"脱が", "脱ぎ", "脱いで", "脱ぐ"}, "Tidak perlu melakukan: Vない-form + なくてもいい, 脱がなくてもいい."},
		{"日本語の新聞が少し（　）ようになりました。", [4]string{"読める", "読む", "読んだ", "読みたい"}, "Kemampuan yang mulai didapat: 読めるようになりました."},
		{"駅に着く（　）、友達から電話がありました。", [4]string{"と", "なら", "ても", "しか"}, "～と menghubungkan begitu sampai dengan kejadian berikutnya."},
		{"この店は安い（　）、とてもおいしいです。", [4]string{"し", "ので", "まで", "しか"}, "～し dapat menghubungkan alasan atau keunggulan tambahan."},
		{"動物園は九時に開きます。入園料は大人800円、子ども400円です。月曜日は休みですが、祝日の月曜日は開いています。", [4]string{"祝日の月曜日は開いています。", "毎週月曜日は必ず休みです。", "子どもは800円です。", "開園は十時です。"}, "Pengecualian hari libur disebutkan langsung."},
		{"スーパーの特売：卵10個180円、牛乳1本220円。今日は野菜を二つ買うと、合計から100円引きます。", [4]string{"野菜を二つ買うと100円安くなります。", "卵は220円です。", "牛乳は180円です。", "野菜は一つだけ買えます。"}, "Promo memberi potongan total 100 yen jika membeli dua sayur."},
		{"ケンさんへ。先にレストランに入っていてください。私は仕事が終わってから行きます。予約は七時で、名前はスズキです。", [4]string{"予約の名前はスズキです。", "ケンさんは仕事の後に来ます。", "予約は八時です。", "二人は家で会います。"}, "Catatan menyebut reservasi atas nama Suzuki pukul tujuh."},
		{"先月、新しい自転車を買いました。駅まで歩くと十五分かかりますが、自転車なら五分です。雨の日はバスを使います。", [4]string{"自転車で駅まで五分です。", "駅まで歩いて五分です。", "雨の日も自転車に乗ります。", "自転車を先月売りました。"}, "Perjalanan dengan sepeda memerlukan lima menit."},
		{"学校祭では、学生が歌やダンスを発表します。食べ物の店もあります。体育館の発表は一時に始まり、校庭の店は午前十時からです。", [4]string{"体育館の発表は一時に始まります。", "店は午後一時からです。", "学校祭に食べ物の店はありません。", "学生は発表しません。"}, "Jadwal pertunjukan di gedung olahraga pukul satu."},
	}
}

func natPaperFive() []natSimulationQuestion {
	return []natSimulationQuestion{
		{"[[旅行]]の写真を見せてください。", [4]string{"りょこう", "りょうこう", "りょこ", "りゅこう"}, "旅行（りょこう） berarti perjalanan."},
		{"毎週、友達と[[練習]]します。", [4]string{"れんしゅう", "れんしょう", "ねんしゅう", "れんじゅう"}, "練習 dibaca れんしゅう, latihan."},
		{"おなかがすいたので、食堂で昼ご飯を（　）ました。", [4]string{"食べ", "飲み", "買い", "書き"}, "食べる berarti makan; 昼ご飯を食べました = makan siang."},
		{"駅まで歩いたので、足が（　）ました。", [4]string{"疲れ", "楽しく", "元気に", "静かに"}, "疲れる berarti lelah: 足が疲れました."},
		{"この漢字の意味を辞書で（　）ください。", [4]string{"調べて", "集めて", "急いで", "始めて"}, "調べる berarti mencari atau memeriksa informasi."},
		{"荷物が多いので、タクシーに（　）ましょう。", [4]string{"乗り", "降り", "入り", "出"}, "乗る untuk naik kendaraan."},
		{"子どもたちは庭で楽しそうに（　）います。", [4]string{"遊んで", "働いて", "習って", "教えて"}, "遊ぶ berarti bermain."},
		{"駅に着いたら、改札口の前に（　）ください。", [4]string{"集まって", "始まって", "止まって", "決まって"}, "集まる berarti berkumpul."},
		{"今日は昨日（　）暖かいです。", [4]string{"より", "まで", "しか", "ほど"}, "Perbandingan memakai より: lebih hangat daripada kemarin."},
		{"日本へ行ったら、京都を（　）たいです。", [4]string{"見", "見る", "見て", "見た"}, "Keinginan memakai akar bentuk ます + たい: 見たい."},
		{"このボタンを押すと、ドアが（　）。", [4]string{"開きます", "開けますか", "開きたい", "開いて"}, "Pintu terbuka: 自動詞 開く."},
		{"傘を持っていませんでしたが、友達が貸して（　）。", [4]string{"くれました", "あげました", "もらいましたか", "ください"}, "Teman melakukan kebaikan kepada saya: 貸してくれました."},
		{"宿題を忘れない（　）、ノートに書きました。", [4]string{"ように", "ために", "そうに", "みたいに"}, "Tujuan agar tidak lupa memakai ～ないように."},
		{"この部屋は静かで、勉強（　）いいです。", [4]string{"するのに", "したので", "すると", "しても"}, "～するのにいい menyatakan sesuatu cocok untuk kegiatan."},
		{"電車の中では携帯電話で話さない（　）してください。", [4]string{"ように", "ことを", "ために", "そうに"}, "Permintaan agar tidak melakukan: ～ないようにしてください."},
		{"美術館：火曜から日曜、午前十時から午後六時。入館は五時半までです。高校生以下は無料です。", [4]string{"高校生は無料で入れます。", "月曜日は開いています。", "入館は六時までです。", "朝九時に開きます。"}, "高校生以下 gratis; masuk terakhir pukul 17.30."},
		{"駅のロッカー：小300円、中500円、大700円。現金のみ使えます。荷物は三日以内に出してください。", [4]string{"ロッカーは現金で使います。", "小さいロッカーは500円です。", "荷物は一週間置けます。", "カードで払えます。"}, "Pengumuman menyatakan pembayaran hanya dengan uang tunai."},
		{"アキさんへ。明日のピクニックは公園の入口に十時集合です。雨なら中止ですが、朝八時までにメールします。飲み物は各自で持ってきてください。", [4]string{"雨の場合、朝八時までに連絡があります。", "集合は駅に九時です。", "飲み物はこちらで用意します。", "雨でも必ず行きます。"}, "Penyelenggara akan mengirim email sebelum pukul delapan jika hujan."},
		{"先週、祖母の家へ行きました。家の近くに大きな川があります。子どものころ、夏休みにそこで魚を見たり、水遊びをしたりしました。今は橋が新しくなっています。", [4]string{"子どものころ、川の近くで遊びました。", "祖母の家は海の近くです。", "先週、古い橋を作りました。", "川で魚を売りました。"}, "Kenangan masa kecilnya bermain air dan melihat ikan di sungai."},
		{"会社の旅行は来月の土曜日です。朝七時に会社の前からバスで出発します。昼ご飯は用意しますが、飲み物は自分で持ってきてください。", [4]string{"飲み物は自分で準備します。", "旅行は今月です。", "バスは八時に出ます。", "昼ご飯を持っていく必要があります。"}, "Informasi menyebut makan siang disediakan, minuman dibawa sendiri."},
	}
}

// Four original, compact NAT 5Q written papers. Each paper combines vocabulary,
// grammar, short reading, and practical information search; listening is omitted.
func simulationNATAdditionalContent(paper int) []simulationItem {
	if paper >= 6 {
		// Papers 06 and later live in simulation_data/*.json (see simulation_json.go).
		return simulationJSONContent("NAT", paper)
	}
	sets := map[int][]natSimulationQuestion{
		2: {
			{"[[病院]]で先生に会いました。", [4]string{"びょういん", "びょうえん", "びょうじん", "びょうにん"}, "病院（びょういん） berarti rumah sakit."},
			{"毎日、六時に[[起きます]]。", [4]string{"おきます", "いきます", "あきます", "ききます"}, "起きます dibaca おきます, bangun tidur."},
			{"この料理は少し（　）です。水をください。", [4]string{"からい", "あまい", "うすい", "ひくい"}, "からい berarti pedas; permintaan air cocok dengan konteks."},
			{"教室に入る前に、くつを（　）ください。", [4]string{"ぬいで", "あらって", "あけて", "ならんで"}, "ぬぐ berarti melepas pakaian atau sepatu."},
			{"バスが来るまで、ここで（　）ください。", [4]string{"待って", "走って", "泳いで", "売って"}, "待つ berarti menunggu."},
			{"この道をまっすぐ行くと、右に郵便局が（　）。", [4]string{"あります", "います", "します", "なります"}, "Benda dan bangunan memakai あります."},
			{"今日は寒いですから、コートを（　）ほうがいいです。", [4]string{"着た", "着る", "着て", "着ない"}, "Saran memakai ～たほうがいい: 着たほうがいい."},
			{"駅に着いたら、電話を（　）ください。", [4]string{"かけて", "かける", "かけた", "かけます"}, "Permintaan memakai bentuk て + ください."},
			{"友達と日曜日（　）映画を見ます。", [4]string{"に", "が", "を", "で"}, "に menandai waktu tertentu: 日曜日に映画を見ます."},
			{"昨日、友達と新しい映画を（　）。", [4]string{"見ました", "見ます", "見ません", "見て"}, "昨日 menandai masa lampau: 見ました."},
			{"A：ここで写真を撮ってもいいですか。 B：すみません、ここでは（　）。", [4]string{"撮らないでください", "撮ってください", "撮りましょう", "撮りたいです"}, "Larangan sopan: 撮らないでください."},
			{"私は日本料理を作った（　）があります。", [4]string{"こと", "もの", "ところ", "ため"}, "Pengalaman memakai ～たことがあります."},
			{"田中さんは来ませんでした。仕事が（　）そうです。", [4]string{"忙しかった", "忙しい", "忙しく", "忙しさ"}, "～そうです untuk kabar yang didengar dapat mengikuti bentuk lampau: 忙しかったそうです."},
			{"雨が降っています（　）、試合は中止です。", [4]string{"から", "まで", "しか", "でも"}, "から menghubungkan alasan dan akibat."},
			{"この箱は重すぎて、一人では（　）。", [4]string{"持てません", "持ちます", "持ちたい", "持って"}, "Potensial negatif 持てません menyatakan tidak mampu mengangkat."},
			{"図書館は月曜日が休みです。火曜日から土曜日は九時から五時まで開いています。日曜日は十時から三時までです。", [4]string{"日曜日は十時に開きます。", "月曜日は五時まで開いています。", "土曜日は休みです。", "毎日九時に開きます。"}, "Informasi menyebut Minggu buka mulai pukul sepuluh."},
			{"カフェのお知らせ：コーヒー350円、紅茶300円。午後二時から五時までケーキを一つ買うと飲み物が50円安くなります。", [4]string{"三時にケーキと紅茶を買うと、紅茶は250円です。", "一時にコーヒーを買うと300円です。", "六時にケーキを買うと飲み物が安くなります。", "紅茶はいつも250円です。"}, "Pada jam promo, teh 300 yen dikurangi 50 yen menjadi 250 yen."},
			{"ミカさんへ。会議は二時からです。資料をコピーしてから行くので、少し遅れます。先に始めてください。佐藤", [4]string{"佐藤さんは資料をコピーしてから会議へ行きます。", "会議は三時からです。", "ミカさんが資料を作ります。", "佐藤さんは会議に行きません。"}, "Pesan menyatakan Sato akan menyalin bahan sebelum datang."},
			{"駅前の店で財布を買いました。黒い小さい財布で、カードが三枚入ります。値段は二千円でした。", [4]string{"財布は黒くて小さいです。", "財布は白いです。", "カードが五枚入ります。", "三千円でした。"}, "Teks menyebut dompet hitam kecil dan muat tiga kartu."},
			{"今朝は寝坊しました。朝ご飯を食べないで家を出ました。駅まで走りましたが、電車に間に合いませんでした。次の電車は十分後でした。", [4]string{"次の電車を十分待ちました。", "朝ご飯を食べてから出ました。", "駅までバスで行きました。", "電車にすぐ乗れました。"}, "Kereta berikutnya datang sepuluh menit kemudian."},
		},
	}
	// Paper 3–5 use distinct original question sets below.
	sets[3] = natPaperThree()
	sets[4] = natPaperFour()
	sets[5] = natPaperFive()
	questions := sets[paper]
	if len(questions) != 20 {
		return nil
	}
	// Add 11 original vocabulary questions and 12 grammar questions, keeping the
	// five reading passages at the end to match the full NAT 5Q paper shape.
	extra := natAdditionalQuestions(paper)
	baseGrammar := natBaseGrammarIndexes(paper)
	replacements := natReplacementVocabulary(paper)
	if len(extra) != 23 || len(replacements) != len(baseGrammar) {
		return nil
	}
	base := questions
	questions = make([]natSimulationQuestion, 0, 43)
	for i, item := range base[:10] {
		if baseGrammar[i] {
			continue
		}
		questions = append(questions, item)
	}
	questions = append(questions, extra[:11]...)
	questions = append(questions, replacements...)
	questions = append(questions, base[10:15]...)
	for i, item := range base[:10] {
		if baseGrammar[i] {
			questions = append(questions, item)
		}
	}
	questions = append(questions, extra[11+len(baseGrammar):]...)
	questions = append(questions, base[15:]...)
	if len(questions) != 43 {
		return nil
	}
	items := make([]simulationItem, 0, len(questions))
	for i, item := range questions {
		section, group, instructions, passage := "grammar", "文法 問題1 · Grammar & Reading", "（　）に入るいちばんよいものを一つ選んでください。", ""
		if i < 21 {
			section, group = "vocabulary", "文字・語彙 問題 · Kosakata & Kanji"
			instructions = "（　）に入るいちばんよいものを一つ選んでください。"
			if strings.Contains(item.prompt, "[[") {
				instructions = "＿＿の言葉はどう読みますか。いちばんよいものを一つ選んでください。"
			}
		} else if i >= 38 {
			group, instructions = "読解 問題 · Bacaan dan informasi", "次の文章を読んで、質問に答えてください。"
			passage = item.prompt
			item.prompt = "文章の内容と合っているものはどれですか。"
		}
		items = append(items, simulationItem{section, group, instructions, passage, item.prompt, item.choices, item.why})
	}
	return items
}
