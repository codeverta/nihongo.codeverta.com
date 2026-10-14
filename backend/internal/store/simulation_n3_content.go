package store

// Original N3 written practice paper: 35 vocabulary + 23 grammar + 16 reading.
// Correct choices are authored first; StartSimulation randomizes their position.
func simulationN3Content() []simulationItem {
	type question struct {
		prompt      string
		choices     [4]string
		explanation string
	}
	items := make([]simulationItem, 0, 74)
	add := func(section, group, instructions, passage string, questions []question) {
		for _, q := range questions {
			items = append(items, simulationItem{section, group, instructions, passage, q.prompt, q.choices, q.explanation})
		}
	}
	add("vocabulary", "文字・語彙 問題1 · 漢字の読み", "＿＿の言葉はどう読みますか。いちばんよいものを一つ選んでください。", "", []question{
		{"会議の日程を[[確認]]してください。", [4]string{"かくにん", "かくねん", "こうにん", "かっにん"}, "確認（かくにん） berarti memastikan atau mengecek. 日程を確認する = mengecek jadwal."},
		{"駅の近くで部屋を[[探して]]います。", [4]string{"さがして", "さぐして", "たがして", "そうして"}, "探す dibaca さがす; 探して adalah bentuk て, mencari sesuatu."},
		{"この会社では新しい社員を[[募集]]しています。", [4]string{"ぼしゅう", "ぼじゅう", "もしゅう", "ぼしゅ"}, "募集（ぼしゅう） berarti membuka perekrutan atau mengumpulkan peserta. しゅう memakai vokal panjang."},
		{"荷物を受付に[[預けました]]。", [4]string{"あずけました", "あつけました", "あげました", "あづきました"}, "預ける（あずける） berarti menitipkan. 預かる berarti menerima titipan, sehingga perannya berbeda."},
		{"この薬の[[効果]]について説明を受けた。", [4]string{"こうか", "こか", "ごうか", "こうが"}, "効果（こうか） berarti efek atau hasil. Vokal こう panjang."},
		{"出発までに旅行の[[準備]]を終えたい。", [4]string{"じゅんび", "しゅんび", "じゅんぴ", "じゅび"}, "準備（じゅんび） berarti persiapan. 音 じゅ berbeda dari しゅ."},
		{"その色の服はあなたによく[[似合う]]。", [4]string{"にあう", "じあう", "しあう", "にごう"}, "似合う（にあう） berarti cocok pada seseorang, misalnya warna pakaian."},
		{"問題の原因を[[調べる]]必要がある。", [4]string{"しらべる", "えらべる", "ならべる", "くらべる"}, "調べる（しらべる） = menyelidiki. 選ぶ = memilih, 並べる = menyusun, 比べる = membandingkan."},
	})
	add("vocabulary", "文字・語彙 問題2 · 表記", "＿＿の言葉を漢字で書くとき、いちばんよいものを一つ選んでください。", "", []question{
		{"台風の[[えいきょう]]で電車が止まった。", [4]string{"影響", "映響", "影郷", "映郷"}, "影響（えいきょう） = pengaruh. Bentuk 影 dan 響 membentuk kata yang tepat."},
		{"この道は夜になると[[きけん]]だ。", [4]string{"危険", "危検", "期険", "機検"}, "危険（きけん） berarti berbahaya. 険 memakai radikal 阝, bukan 検 dengan 木."},
		{"子どもの[[せいちょう]]を見守る。", [4]string{"成長", "生長期", "正長", "成張"}, "成長（せいちょう） adalah pertumbuhan atau perkembangan anak. 生長期 merupakan kata lebih panjang dengan arti periode pertumbuhan."},
		{"参加を[[きぼう]]する人は連絡してください。", [4]string{"希望", "期望", "希忘", "起望"}, "希望（きぼう） berarti harapan atau keinginan. 参加を希望する = ingin ikut."},
		{"水や電気を[[せつやく]]しましょう。", [4]string{"節約", "接約", "説役", "節役"}, "節約（せつやく） = menghemat penggunaan sumber daya atau uang."},
		{"この仕事には大きな[[せきにん]]がある。", [4]string{"責任", "積任", "責認", "席認"}, "責任（せきにん） = tanggung jawab. 責 menandai kewajiban, bukan 積 yang berarti menumpuk."},
	})
	add("vocabulary", "文字・語彙 問題3 · 文脈", "（　）に入るいちばんよい言葉を一つ選んでください。", "", []question{
		{"祖母は足が弱いので、歩くときは家族が体を（　）いる。", [4]string{"支えて", "断って", "減らして", "破って"}, "支える berarti menopang. Keluarga menopang tubuh nenek ketika berjalan."},
		{"お腹がすいたが、夕食まであと少しなので（　）することにした。", [4]string{"我慢", "案内", "交換", "発見"}, "我慢する = menahan diri; menahan lapar sampai waktu makan."},
		{"（　）約束の時間を忘れてしまい、友達に謝った。", [4]string{"うっかり", "しっかり", "ぐっすり", "はっきり"}, "うっかり menunjukkan kelalaian tanpa sengaja. しっかり = dengan mantap, ぐっすり = tidur nyenyak."},
		{"週末のレストランは（　）だったので、外で三十分待った。", [4]string{"満席", "満足", "完全", "十分"}, "満席 = semua tempat duduk terisi. Akibatnya tamu harus menunggu di luar."},
		{"雨がひどくなったため、運動会は来週に（　）された。", [4]string{"延期", "出発", "到着", "完成"}, "延期する = menunda ke tanggal berikutnya. Acara tidak dibatalkan selamanya."},
		{"二つの商品を（　）てから、安い方を買った。", [4]string{"比べ", "並び", "運び", "増え"}, "比べる = membandingkan. 比べて lalu memilih yang lebih murah sesuai konteks."},
		{"誰も使わない部屋の電気をつけておくのは、エネルギーの（　）だ。", [4]string{"無駄", "限界", "方法", "予定"}, "無駄 = pemborosan atau sesuatu yang sia-sia. Lampu di ruang kosong memboroskan energi."},
		{"どの大学に入るか、まだ（　）いる。", [4]string{"迷って", "片づけて", "届けて", "続けて"}, "迷う = bimbang memilih. Masih belum menentukan universitas."},
		{"彼は毎日（　）を重ねて、ついに大会で優勝した。", [4]string{"努力", "予約", "伝言", "故障"}, "努力を重ねる berarti terus melakukan usaha. Usaha sehari-hari menghasilkan kemenangan."},
		{"集合時間に遅れそうだったので、駅まで（　）足で歩いた。", [4]string{"急ぎ", "帰り", "下り", "入り"}, "急ぎ足 = langkah tergesa-gesa. Istilah ini sesuai dengan kekhawatiran terlambat."},
		{"旅行の前に、現地の天気や交通について（　）を集めた。", [4]string{"情報", "事情", "表情", "感情"}, "情報を集める = mengumpulkan informasi. 表情 = ekspresi wajah dan 感情 = perasaan."},
	})
	add("vocabulary", "文字・語彙 問題4 · 言い換え", "＿＿の言葉に意味が最も近いものを一つ選んでください。", "", []question{
		{"その仕事を[[断った]]。", [4]string{"引き受けなかった", "すぐ始めた", "最後まで続けた", "誰かに頼んだ"}, "断る = menolak, sehingga tidak menerima pekerjaan tersebut. Bukan sekadar menyerah setelah mulai."},
		{"参加者は[[およそ]]百人だった。", [4]string{"だいたい", "ちょうど", "少なくとも", "必ず"}, "およそ = kira-kira. ちょうど berarti tepat, bukan perkiraan."},
		{"工事が[[ついに]]終わった。", [4]string{"とうとう", "すでに何度も", "たまたま", "とりあえず"}, "ついに = akhirnya setelah proses atau waktu yang cukup lama; とうとう paling dekat."},
		{"試合の結果を聞いて[[がっかりした]]。", [4]string{"期待どおりでなくて残念に思った", "予想よりよくて喜んだ", "意味が分からなくて困った", "急な知らせで驚いた"}, "がっかりする = kecewa karena hasil tidak memenuhi harapan, bukan sekadar terkejut."},
		{"彼女は[[じっと]]座っていた。", [4]string{"動かずに", "急いで", "何度も", "少しずつ"}, "じっと = diam tanpa bergerak. Konteksnya duduk tenang."},
	})
	add("vocabulary", "文字・語彙 問題5 · 用法", "次の言葉の使い方として最もよい文を一つ選んでください。", "", []question{
		{"参加する", [4]string{"来週の会議に参加するつもりだ。", "新しい靴を参加して歩いた。", "窓が参加しているので閉めた。", "駅まで参加して十五分かかった。"}, "参加する digunakan dengan acara atau kegiatan: 会議に参加する = mengikuti rapat."},
		{"気づく", [4]string{"電車に乗ってから、財布を忘れたことに気づいた。", "毎朝七時に学校を気づいている。", "母は野菜を小さく気づいた。", "新しい仕事を気づいて応募した。"}, "気づく berarti menyadari fakta, memakai ～に気づく. Bukan berangkat, memotong, atau mencari pekerjaan."},
		{"そろそろ", [4]string{"もう十時だから、そろそろ帰ろう。", "この料理はそろそろ甘い。", "彼はそろそろ三人の兄がいる。", "昨日、そろそろ一度も寝なかった。"}, "そろそろ digunakan saat waktu suatu tindakan sudah mendekat: sudah waktunya pulang."},
		{"関係", [4]string{"この問題は私の仕事と関係がある。", "駅から会社まで関係で歩いた。", "このかばんは関係が重い。", "先生に漢字を関係してもらった。"}, "Aと関係がある = memiliki kaitan dengan A. 関係 bukan cara berjalan atau kata sifat berat."},
		{"減る", [4]string{"運動を続けたら、体重が少し減った。", "机の上の本を減って並べた。", "友達から手紙を減った。", "バスに減って駅へ行った。"}, "減る adalah verba intransitif: jumlah/berat が減る. Verba transitif untuk mengurangi adalah 減らす."},
	})
	add("grammar", "文法 問題1 · 文の文法", "（　）に入るいちばんよいものを一つ選んでください。", "", []question{
		{"毎日練習したので、簡単なニュースなら聞き取れる（　）。", [4]string{"ようになった", "ことにした", "ところだった", "ままになった"}, "V可能形＋ようになった menyatakan perubahan kemampuan: sekarang bisa memahami berita sederhana."},
		{"健康のために、できるだけ階段を使う（　）いる。", [4]string{"ようにして", "ことになって", "ところにして", "ばかりにして"}, "V辞書形＋ようにしている = membiasakan diri atau berusaha melakukannya. Memilih tangga merupakan usaha untuk sehat."},
		{"この店は値段が安い（　）、料理もおいしい。", [4]string{"うえに", "ために", "かわりに", "とおりに"}, "～うえに menambahkan kelebihan lain: murah dan juga enak. ために tidak menghubungkan dua evaluasi tambahan ini."},
		{"雨が降った（　）、予定どおり試合が行われた。", [4]string{"にもかかわらず", "おかげで", "たびに", "ついでに"}, "～にもかかわらず = meskipun. Pertandingan tetap berlangsung berlawanan dengan keadaan hujan."},
		{"この本は子ども（　）書かれているので、説明がやさしい。", [4]string{"向けに", "に代わって", "に比べて", "によって"}, "N向けに = ditujukan untuk kelompok tersebut. Buku ini ditulis untuk anak, bukan oleh anak."},
		{"寝る前にスマートフォンを見る（　）、眠れなくなった。", [4]string{"せいで", "おかげで", "ためなら", "一方で"}, "～せいで menyatakan sebab dari akibat negatif. おかげで biasanya dipakai untuk hasil yang menguntungkan."},
		{"先生に説明してもらった（　）、問題の意味が分かった。", [4]string{"おかげで", "せいで", "くせに", "だけにして"}, "～おかげで menunjukkan bantuan sebagai sebab hasil yang baik: dapat memahami soal."},
		{"京都へ行く（　）、この店でお土産を買う。", [4]string{"たびに", "うちに", "ところに", "ばかりに"}, "V辞書形＋たびに = setiap kali. Membeli oleh-oleh dilakukan setiap kunjungan ke Kyoto."},
		{"この仕事は私一人ではできない。みんなに手伝ってもらう（　）。", [4]string{"しかない", "ほどない", "わけない", "さえない"}, "V辞書形＋しかない = tidak ada pilihan selain melakukan tindakan tersebut."},
		{"日本語を長く勉強しているが、何でも分かる（　）。", [4]string{"わけではない", "ことにしている", "ようになる", "ところである"}, "～わけではない membatasi kesimpulan: lama belajar tidak berarti memahami segala hal."},
		{"申込書は明日までに出す（　）。忘れないでください。", [4]string{"ことになっています", "つもりになっています", "ばかりになっています", "ところにしています"}, "～ことになっている menyatakan ketentuan atau aturan yang sudah ditetapkan, yaitu tenggat formulir."},
		{"あの人とは初めて会ったのに、以前会ったことがある（　）。", [4]string{"ような気がする", "ことにする", "ばかりだと思う", "ままにしている"}, "～ような気がする menyatakan kesan subjektif yang belum pasti: terasa seperti pernah bertemu."},
		{"説明を聞けば聞く（　）、この研究に興味がわいてくる。", [4]string{"ほど", "だけで", "しか", "さえ"}, "Vば＋V辞書形＋ほど = semakin melakukan, semakin berubah. 聞けば聞くほど = semakin mendengarkan."},
	})
	add("grammar", "文法 問題2 · 文の組み立て", "＿＿に言葉を入れて文を完成させてください。★に入るものを一つ選んでください。", "", []question{
		{"日本語が ＿＿ ＿＿ ★ ＿＿ 、毎日練習している。", [4]string{"ように", "話せる", "なるために", "もっと自然に"}, "Urutan: もっと自然に／話せる／ように／なるために。日本語がもっと自然に話せるようになる = agar bisa berbicara lebih alami. ★ adalah slot ketiga: ように。"},
		{"忙しい ＿＿ ★ ＿＿ ＿＿ つもりだ。", [4]string{"勉強を", "ときでも", "続ける", "ようにする"}, "Urutan: ときでも／勉強を／続ける／ようにする。Meskipun sibuk, berencana berusaha terus belajar. ★ di slot kedua adalah 勉強を。"},
		{"この町は ＿＿ ＿＿ ★ ＿＿ 住みやすい。", [4]string{"交通も", "静かな", "うえに", "便利なので"}, "Urutan: 静かな／うえに／交通も／便利なので。Kota ini tenang dan transportasinya nyaman, sehingga mudah ditinggali. ★ slot ketiga adalah 交通も。"},
		{"資料を ＿＿ ★ ＿＿ ＿＿ ください。", [4]string{"説明した", "先生が", "とおりに", "並べて"}, "Urutan: 先生が／説明した／とおりに／並べて。Susun dokumen sesuai penjelasan guru. ★ slot kedua adalah 説明した, bukan とおりに。"},
		{"電車が ＿＿ ＿＿ ＿＿ ★ ことになった。", [4]string{"会議を延期する", "止まった", "せいで", "参加者が集まれず"}, "Urutan: 止まった／せいで／参加者が集まれず／会議を延期する。Kereta berhenti sehingga peserta tidak dapat berkumpul; rapat harus ditunda. ★ slot keempat adalah 会議を延期する。"},
	})
	text := "先月、私は地域の清掃活動に参加した。以前は、休みの日は家でゆっくりしたいと思っていた。しかし、近所の人に誘われた【1】、一度行ってみることにした。\n最初は知らない人ばかりで緊張したが、いっしょにごみを集める【2】、少しずつ話せるようになった。活動が終わると、道がきれいになっていた。それを見て、自分も町の役に立てた【3】気がした。\n私は、この活動は町をきれいにする【4】、近所の人と知り合う機会にもなると気づいた。次は友達も誘って参加しよう【5】。"
	add("grammar", "文法 問題3 · 文章の文法", "文章を読んで、【1】から【5】に入るいちばんよいものを選んでください。", text, []question{
		{"【1】に入るものはどれですか。", [4]string{"のをきっかけに", "のに対して", "わけがなく", "ことがないので"}, "～のをきっかけに = dengan kejadian itu sebagai pemicu. Ajakan tetangga menjadi awal keputusan ikut."},
		{"【2】に入るものはどれですか。", [4]string{"うちに", "ためには", "ばかりに", "として"}, "V辞書形＋うちに = selama melakukan kegiatan, perubahan terjadi. Selama mengumpulkan sampah, penulis menjadi bisa berbicara dengan peserta lain."},
		{"【3】に入るものはどれですか。", [4]string{"ような", "ための", "はずの", "ばかりの"}, "～ような気がした = merasa seolah-olah. Penulis merasa dirinya berguna bagi kota."},
		{"【4】に入るものはどれですか。", [4]string{"だけでなく", "にもかかわらず", "わけではなく", "しかないので"}, "～だけでなく、～にもなる = tidak hanya membersihkan kota, tetapi juga kesempatan mengenal tetangga."},
		{"【5】に入るものはどれですか。", [4]string{"と思っている", "ことになれない", "ように見える", "に違いない"}, "V意向形＋と思っている menyatakan rencana pribadi yang sedang dimiliki: berniat mengajak teman dan ikut lagi."},
	})
	read := func(group, passage string, questions []question) {
		add("grammar", group, "次の文章を読んで、質問の答えとして最もよいものを一つ選んでください。", passage, questions)
	}
	read("読解 問題4 · 短文", "佐藤さんへ\n明日の打ち合わせですが、取引先の都合で開始が午後二時から三時に変わりました。場所は予定どおり本社の会議室です。資料は私が準備しますので、佐藤さんは先週の売上の数字だけ確認しておいてください。変更のお知らせは、私から参加者全員に送っておきます。\n田中", []question{
		{"佐藤さんは打ち合わせの前に何をする必要がありますか。", [4]string{"先週の売上の数字を確認する。", "参加者全員に時間の変更を知らせる。", "会議室を別の場所に変更する。", "田中さんの代わりにすべての資料を作る。"}, "Instruksi untuk Sato adalah 売上の数字だけ確認しておいてください. Tanaka menyiapkan dokumen dan memberi tahu peserta."},
	})
	read("読解 問題4 · 短文", "駅前のパン屋は、毎晩、売れ残ったパンを近くの福祉施設に届けている。以前は閉店の直前に値段を下げて売っていたが、それでも残ることがあった。店長は『食べられるものを捨てたくないと思って始めました。必要としている人が喜んでくれることも、続ける理由です』と話している。", []question{
		{"店長がこの取り組みを始めた主な理由は何ですか。", [4]string{"まだ食べられるパンを捨てたくなかったから。", "施設にパンを売って利益を増やしたかったから。", "閉店の時間を早くしたかったから。", "値下げするといつも全部売れていたから。"}, "Penulis mengutip alasan awal 食べられるものを捨てたくない. Kegembiraan penerima menjadi alasan melanjutkan, bukan tujuan menjual."},
	})
	read("読解 問題4 · 短文", "図書館からのお知らせ\n今月から、予約した本を受け取れる期間は、連絡した日の翌日から七日間となります。期間を過ぎると予約は取り消されます。来館が難しい場合は、期限までに電話をいただければ一度だけ三日間延長できます。ウェブサイトからの延長はできません。", []question{
		{"期限までに本を受け取れない人は、どうすればよいですか。", [4]string{"期限までに電話して、三日間の延長を頼む。", "ウェブサイトで何度でも延長する。", "期限が過ぎてから電話して予約を残す。", "何もせず、次の七日間が始まるのを待つ。"}, "Pemberitahuan meminta telepon sebelum tenggat; perpanjangan hanya satu kali selama tiga hari. Situs tidak dapat dipakai memperpanjang."},
	})
	read("読解 問題4 · 短文", "私は料理の本を見ながら初めてカレーを作った。材料の量も時間も本に書いてあるとおりにしたが、思ったより辛くなった。あとで母に聞くと、同じ名前の調味料でも製品によって辛さが違うそうだ。次からは、量を一度に入れず、味を確かめながら少しずつ加えようと思う。", []question{
		{"筆者は次にカレーを作るとき、どうするつもりですか。", [4]string{"味を見ながら調味料を少しずつ入れる。", "本に書いてある時間より長く煮る。", "同じ調味料を前より多く入れる。", "母にすべて作ってもらう。"}, "Kalimat terakhir menyatakan rencana menambah sedikit demi sedikit sambil memastikan rasa. Tidak ada rencana mengubah durasi atau menyerahkan masakan kepada ibu."},
	})
	read("読解 問題5 · 中文", "私の会社では、会議の前に必ず資料を送ることになっていた。しかし、多くの人は忙しくて読めず、会議の最初の二十分は担当者の説明を聞くだけだった。意見を出す時間が足りなくなり、結論を次の会議まで待つこともあった。\nそこで、ある部署が会議の最初に十分間だけ、全員で資料を読む時間を作った。資料の説明はやめ、分からない点だけを質問するようにした。すると、参加者は同じ内容を理解してから話し合えるようになり、以前より短い時間で結論が出るようになった。\n最初は『読む時間を会議中に取るのは無駄だ』という意見もあった。でも、担当者は、読む時間を入れることで、説明を聞くだけの時間や後で確認する時間が減ったと考えている。大切なのは会議をただ短くすることではなく、必要な話し合いができるようにすることなのだ。", []question{
		{"以前の会議には、どんな問題がありましたか。", [4]string{"説明に時間を使い、十分に話し合えないことがあった。", "資料が一度も送られなかった。", "担当者が質問に答えなかった。", "参加者が結論を出すことを禁止されていた。"}, "Awalnya dua puluh menit dipakai untuk penjelasan, sehingga 意見を出す時間が足りなくなり dan keputusan bisa ditunda."},
		{"新しい方法で全員が資料を読む時間を取るのは、何のためですか。", [4]string{"同じ内容を理解したうえで、必要な話し合いをするため。", "担当者が資料を作らなくてもよいようにするため。", "質問をせず、必ず十分で会議を終えるため。", "事前に資料を送ることを禁止するため。"}, "Tujuannya menyamakan pemahaman agar diskusi yang diperlukan bisa berlangsung. Sepuluh menit hanya untuk membaca, bukan batas seluruh rapat."},
	})
	read("読解 問題5 · 中文", "私は半年ほど前から、休日の朝に近くの公園を走っている。走り始めたころは、毎回どれくらい速く走れたかを気にしていた。前の週より遅いと、体調が悪いわけでもないのに、失敗したような気分になった。\nある日、公園でよく会う年配の人に、そのことを話した。その人は『速さだけを比べるより、走った後にどう感じたかも考えてみたらどうですか』と言った。私は試しに、走る速さだけでなく、帰宅したときの気分も記録することにした。\nすると、ゆっくり走った日でも、頭がすっきりして、午後の仕事に集中できていることに気づいた。今は、疲れている日は無理に速度を上げないようにしている。速く走れるようになることも目標の一つだが、生活を気持ちよく続けるために走るという目的を忘れないようにしたい。", []question{
		{"筆者は以前、どんなときに失敗したように感じましたか。", [4]string{"前の週より走る速さが遅かったとき。", "公園で知っている人に会ったとき。", "午後の仕事に集中できたとき。", "走った後に気分を記録したとき。"}, "Penulis menyatakan 前の週より遅いと...失敗したような気分になった. Masalah awalnya menilai keberhasilan hanya dari kecepatan."},
		{"年配の人の助言を受けて、筆者は何に気づきましたか。", [4]string{"ゆっくり走っても、気分や集中に良い影響がある。", "速く走らない日は必ず体調が悪い。", "気分を記録すれば走る必要はなくなる。", "疲れているときほど速度を上げるべきだ。"}, "Catatan baru menunjukkan efek baik pada kejernihan pikiran dan konsentrasi walaupun berlari pelan. Penulis justru tidak memaksa diri saat lelah."},
	})
	read("読解 問題5 · 中文", "私たちの町では、使わなくなった物を持ち寄る交換会が開かれている。参加者は、本や服、食器などを持ってきて、ほかの人の物を自由に選べる。ただし、壊れている物や汚れがひどい物は受け付けない。売買ではないので、値段を決める必要もない。\n私は初めて参加したとき、自分には不要な物を誰かが欲しいと思うだろうか、と心配していた。ところが、持っていった小さな鍋を見た学生が『一人暮らしを始めたばかりなので、助かります』と喜んでくれた。私は大きな花瓶をもらい、家で使うことにした。\n交換会では、自分が持ってきた数と同じ数だけ持ち帰る必要はない。何も持ち帰らない人もいる。主催者は、一つ一つの交換を公平に数えるよりも、まだ使える物が必要な人のところへ移ることを大切にしているという。", []question{
		{"この交換会で受け付けてもらえるのは、どんな物ですか。", [4]string{"不要になったが、まだ使えるきれいな物。", "壊れていても高い値段を付けた物。", "主催者が買うと約束した物だけ。", "必ず新品で、まだ一度も使っていない物だけ。"}, "物 harus masih dapat dipakai dan tidak sangat kotor. Tidak diwajibkan baru; benda rusak ditolak."},
		{"主催者が最も大切にしていることは何ですか。", [4]string{"使える物が、それを必要とする人に渡ること。", "全員が持ってきた物と同じ数を持ち帰ること。", "参加者からできるだけ多くの代金を集めること。", "物を持ってきた人だけが会場に入れること。"}, "Paragraf terakhir menegaskan pemindahan benda ke orang yang membutuhkan, bukan kesetaraan jumlah atau uang."},
	})
	read("読解 問題6 · 長文", "新しいことを学ぶとき、間違えないように十分準備してから始めたいと思う人は多い。私も以前はそうだった。写真を始めようと思ったとき、まずカメラの使い方を説明する本を何冊も読んだ。専門的な言葉を理解するのに時間がかかり、カメラを持って出かける前に、写真は難しいものだと思ってしまった。\nそのころ、写真が趣味の友人に話すと、『まず身近な物を一枚撮ってみたら』と言われた。私は家の窓から見える木を撮った。思ったより暗く写ったので、本で明るさの調整方法を調べた。調整して撮り直すと、今度は木の形がよく見えた。その小さな変化がうれしくて、翌日もカメラを使った。\n同じ本を読んでいるのに、以前より説明が分かりやすく感じられた。最初に読んだときは、何のための設定なのか、具体的に想像できなかったからだろう。失敗した写真があったからこそ、自分が知りたいことがはっきりし、本の説明を自分の経験と結び付けられたのだ。\nもちろん、準備が必要ないと言いたいのではない。危険を伴う活動なら、安全のために先に確認すべきことがある。ただ、すべてを理解してからでないと始められない、と考える必要はないと思う。小さく試して、分からないところを調べ、もう一度試す。その繰り返しによって、知識はただ覚えた言葉ではなく、使えるものになる。", []question{
		{"筆者が撮影後に本の説明を分かりやすく感じたのは、なぜですか。", [4]string{"自分の経験によって、知りたいことや設定の目的が具体的になったから。", "友人が本の内容を全部暗記させてくれたから。", "専門的な言葉が本からすべてなくなったから。", "最初に撮った写真が完全に成功したから。"}, "Kegagalan foto membuat masalah terang/gelap konkret, sehingga penulis bisa menghubungkan penjelasan dengan pengalamannya. Buku tidak berubah."},
		{"新しいことを学ぶ方法について、筆者はどう考えていますか。", [4]string{"必要な安全確認をしながら、小さく試して調べることを繰り返すとよい。", "どんな活動でも準備や安全確認は不要だ。", "すべての知識を理解するまで、何も試してはいけない。", "失敗したら、別の趣味をすぐに選ぶべきだ。"}, "Kesimpulan menggabungkan percobaan kecil dan pemeriksaan yang diperlukan. Penulis secara eksplisit tidak menolak persiapan atau keselamatan."},
	})
	read("読解 問題6 · 長文", "私の住む地域では、商店街の空いている店を利用して、誰でも立ち寄れる小さな交流スペースを作った。机と椅子、本棚があり、飲み物を持ち込んで休むこともできる。最初は、毎週いろいろなイベントを開けば、たくさんの人が来るだろうと考えられていた。料理教室や音楽会などが企画され、開催日には確かに多くの人が集まった。\nしかし、イベントのない日は、ほとんど人が来なかった。近所の人に理由を聞くと、『何かに参加しないと入れない場所だと思っていた』という声があった。案内には自由に利用できると書いてあったのに、イベントの情報が目立ちすぎて、普段の使い方が伝わっていなかったのだ。\nそこで、運営する人たちは、入口にその日の特別な予定だけでなく、『何もしなくても、ここで休めます』という案内を置いた。また、運営者自身が、イベントのない日に本を読んだり、近所の人と話したりする様子を見せるようにした。すると、買い物の帰りに少し休む人や、家で一人でいるよりここで過ごしたいという人が増えてきた。\nこの場所を作った目的は、参加者の数を増やすことだけではない。地域の人が必要なときに気軽に立ち寄れる場所を持つことだった。にぎやかな催しも大切だが、静かな時間にも安心して利用できることが、その目的に合っている。使い方を説明する言葉だけでなく、実際に使われている様子も見せる必要があると、運営者は考えている。", []question{
		{"イベントのない日に人が来なかった主な理由は何ですか。", [4]string{"イベントに参加しないと入れない場所だと思われていたから。", "入口に椅子や机がまったくなかったから。", "飲み物を持ち込むことが禁止されていたから。", "運営者が毎日すべての店を閉めていたから。"}, "Wawancara menunjukkan kesalahpahaman bahwa masuk harus ikut acara. Informasi acara terlalu menonjol menutupi penggunaan sehari-hari."},
		{"筆者が紹介している運営者の考えに合うものはどれですか。", [4]string{"普段の利用の様子も見せ、必要なときに立ち寄りやすい場所にしたい。", "イベントの参加人数だけを場所の成功の基準にしたい。", "自由に使えるという案内をやめ、予約制にしたい。", "静かに過ごす人を減らし、毎日音楽会を開きたい。"}, "Tujuan ruang adalah kemudahan singgah bagi warga, termasuk waktu yang tenang. Pengelola ingin menunjukkan cara penggunaan nyata, bukan hanya kata-kata."},
	})
	read("読解 問題7 · 情報検索", "みどり市文化センター　秋の講座\n\nA 写真入門：土曜 10:00〜12:00／全3回／3,000円／定員12名／カメラを持参\nB 家庭料理：水曜 18:30〜20:30／全4回／4,000円／定員16名／エプロンを持参\nC はじめての書道：土曜 13:00〜15:00／全3回／2,500円／定員10名／道具は会場で貸し出し\nD 会話英語：金曜 19:00〜20:00／全6回／5,000円／定員20名／教材費込み\n\n申込期間：9月1日〜10日。センターの窓口またはウェブサイトで申し込んでください。電話では申し込めません。\n定員を超えた場合は抽選を行います。申込順ではありません。結果は9月15日にメールで通知します。受講料は当選後、9月20日までに支払ってください。", []question{
		{"山田さんは平日の夜は働いています。土曜日の午後に参加でき、自分で道具を用意しなくてもよい講座を探しています。どれがよいですか。", [4]string{"C はじめての書道", "A 写真入門", "B 家庭料理", "D 会話英語"}, "C berlangsung Sabtu sore dan alat dipinjamkan. A Sabtu pagi dengan kamera sendiri; B dan D malam hari kerja."},
		{"申込みについて、正しいものはどれですか。", [4]string{"定員を超えると抽選になり、当選した人が9月20日までに払う。", "9月15日までなら電話で申し込める。", "早く申し込んだ人から必ず参加できる。", "申し込む前に全員が受講料を払う。"}, "Pemberitahuan menyebut undian bila melebihi kapasitas, bukan urutan pendaftaran. Pembayaran setelah menang undian sampai 20 September."},
	})
	return items
}
