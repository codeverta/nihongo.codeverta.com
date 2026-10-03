package store

// Original practice items for the application's N5 diagnostic, not official JLPT items.
type diagnosticItem struct {
	section, passage, prompt string
	choices                  [4]string
	explanation              string
}

func diagnosticContent() []diagnosticItem {
	return []diagnosticItem{
		{"Kosakata", "", "「あさごはん」はどれですか。", [4]string{"Sarapan", "Makan siang", "Makan malam", "Minuman"}, "あさごはん（朝ご飯）berarti sarapan."},
		{"Kosakata", "", "このへやは（　）です。まどをあけてください。", [4]string{"あつい", "ちかい", "やすい", "わかい"}, "あつい berarti panas. Permintaan membuka jendela sesuai dengan kamar yang panas."},
		{"Kosakata", "", "パンを（　）でかいます。", [4]string{"パンや", "びょういん", "ぎんこう", "としょかん"}, "パンや adalah toko roti; で menandai tempat membeli."},
		{"Kosakata", "", "きのうのつぎは、なんですか。", [4]string{"きょう", "あした", "おととい", "らいしゅう"}, "Hari setelah きのう (kemarin) adalah きょう (hari ini)."},
		{"Kosakata", "", "えんぴつで、なまえを（　）。", [4]string{"かきます", "ききます", "のみます", "およぎます"}, "かきます（書きます）berarti menulis. Nama ditulis menggunakan pensil."},
		{"Kosakata", "", "「ちかい」のはんたいはどれですか。", [4]string{"とおい", "はやい", "おもい", "ひろい"}, "ちかい berarti dekat; lawannya とおい, jauh."},
		{"Kosakata", "", "このかばんは千円です。あのかばんは五千円です。このかばんのほうが（　）です。", [4]string{"やすい", "たかい", "おおい", "ながい"}, "Tas ini 1.000 yen, tas itu 5.000 yen. Tas ini lebih murah: やすい."},
		{"Kanji", "", "「山」のよみかたはどれですか。", [4]string{"やま", "かわ", "うみ", "そら"}, "山 dibaca やま, gunung."},
		{"Kanji", "", "「毎日」のよみかたはどれですか。", [4]string{"まいにち", "まいげつ", "まいとし", "まいしゅう"}, "毎日 dibaca まいにち dan berarti setiap hari."},
		{"Kanji", "", "「学校」のよみかたはどれですか。", [4]string{"がっこう", "がくせい", "せんせい", "きょうしつ"}, "学校 dibaca がっこう, sekolah. Perhatikan bunyi kecil っ."},
		{"Kanji", "", "「みず」はどの漢字ですか。", [4]string{"水", "火", "木", "金"}, "みず berarti air dan ditulis 水."},
		{"Kanji", "", "「三人」のよみかたはどれですか。", [4]string{"さんにん", "さんじん", "みっつ", "みっか"}, "三人 berarti tiga orang dan dibaca さんにん; 三つ menghitung benda."},
		{"Grammar", "", "わたしは毎日六時（　）おきます。", [4]string{"に", "を", "で", "へ"}, "に menandai waktu tertentu: 六時に, pada pukul enam."},
		{"Grammar", "", "きのう、友だちと映画を（　）。", [4]string{"見ました", "見ます", "見ません", "見ましょう"}, "きのう menunjukkan masa lampau. Bentuk sopan lampau afirmatif adalah 見ました."},
		{"Grammar", "", "つくえの上に本が（　）。", [4]string{"あります", "います", "します", "なります"}, "あります digunakan untuk keberadaan benda mati, seperti buku."},
		{"Grammar", "", "この店はあまり（　）。", [4]string{"大きくないです", "大きいです", "大きかったです", "大きくです"}, "あまり biasanya dipakai dengan bentuk negatif: tidak begitu besar, 大きくないです."},
		{"Grammar", "", "ここで写真を（　）もいいですか。", [4]string{"とって", "とる", "とった", "とり"}, "Pola meminta izin adalah bentuk て + もいいですか: とってもいいですか."},
		{"Grammar", "", "日本語は英語（　）むずかしいです。", [4]string{"より", "まで", "だけ", "にも"}, "より menandai pembanding: bahasa Jepang lebih sulit daripada bahasa Inggris menurut kalimat ini."},
		{"Grammar", "", "あした、友だちと公園へ（　）たいです。", [4]string{"行き", "行く", "行って", "行った"}, "～たい melekat pada batang bentuk ます: 行きます → 行きたい."},
		{"Grammar", "", "これは母（　）作ったケーキです。", [4]string{"が", "を", "に", "で"}, "母が作った menerangkan ケーキ: kue yang dibuat ibu. が menandai pelakunya."},
		{"Reading", "田中さんは毎朝七時に起きます。七時半に朝ご飯を食べて、八時に家を出ます。会社までバスで二十分かかります。", "田中さんは何時に家を出ますか。", [4]string{"八時", "七時", "七時半", "八時半"}, "八時に家を出ます menyatakan Tanaka keluar rumah pukul delapan."},
		{"Reading", "田中さんは毎朝七時に起きます。七時半に朝ご飯を食べて、八時に家を出ます。会社までバスで二十分かかります。", "会社までどうやって行きますか。", [4]string{"バスで", "電車で", "歩いて", "車で"}, "会社までバスで二十分: menuju kantor dengan bus selama dua puluh menit."},
		{"Reading", "ゆいさんへ。あした、図書館で一緒に勉強しませんか。十時に入口で会いましょう。昼ご飯は近くの店で食べましょう。はなより。", "二人はどこで会いますか。", [4]string{"図書館の入口", "店の中", "駅の前", "はなさんの家"}, "十時に入口で会いましょう: bertemu di pintu masuk perpustakaan pukul sepuluh."},
		{"Reading", "ゆいさんへ。あした、図書館で一緒に勉強しませんか。十時に入口で会いましょう。昼ご飯は近くの店で食べましょう。はなより。", "はなさんは何をしたいですか。", [4]string{"ゆいさんと勉強したいです", "一人で勉強したいです", "本を買いたいです", "映画を見たいです"}, "一緒に勉強しませんか adalah ajakan belajar bersama Yui."},
		{"Reading", "このスーパーは月曜日から土曜日まで、午前九時から午後八時まで開いています。日曜日は休みです。", "いつスーパーで買い物ができますか。", [4]string{"土曜日の午後七時", "日曜日の午前十時", "月曜日の午前八時", "金曜日の午後九時"}, "Sabtu pukul 19.00 masih dalam jam buka 09.00–20.00, Senin sampai Sabtu. Minggu tutup."},
	}
}
