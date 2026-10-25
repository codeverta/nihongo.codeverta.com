package api

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode"

	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

func TestEnglishExamCatalogCoversDiagnosticIndonesianFields(t *testing.T) {
	sections := map[string]string{
		"Kosakata": "Vocabulary",
		"Kanji":    "Kanji",
		"Grammar":  "Grammar",
		"Reading":  "Reading",
	}
	for source, want := range sections {
		assertExamTranslation(t, "diagnostic.section", source, want)
	}
	for source, want := range map[string]string{
		"Sarapan":     "Breakfast",
		"Makan siang": "Lunch",
		"Makan malam": "Dinner",
		"Minuman":     "A drink",
	} {
		assertExamTranslation(t, "diagnostic.choice", source, want)
	}
	// These are all Indonesian explanatory strings in the diagnostic source.
	// The Japanese prompt, passages, and answer options remain source text.
	for _, source := range []string{
		"あさごはん（朝ご飯）berarti sarapan.",
		"あつい berarti panas. Permintaan membuka jendela sesuai dengan kamar yang panas.",
		"パンや adalah toko roti; で menandai tempat membeli.",
		"Hari setelah きのう (kemarin) adalah きょう (hari ini).",
		"かきます（書きます）berarti menulis. Nama ditulis menggunakan pensil.",
		"ちかい berarti dekat; lawannya とおい, jauh.",
		"Tas ini 1.000 yen, tas itu 5.000 yen. Tas ini lebih murah: やすい.",
		"山 dibaca やま, gunung.",
		"毎日 dibaca まいにち dan berarti setiap hari.",
		"学校 dibaca がっこう, sekolah. Perhatikan bunyi kecil っ.",
		"みず berarti air dan ditulis 水.",
		"三人 berarti tiga orang dan dibaca さんにん; 三つ menghitung benda.",
		"に menandai waktu tertentu: 六時に, pada pukul enam.",
		"きのう menunjukkan masa lampau. Bentuk sopan lampau afirmatif adalah 見ました.",
		"あります digunakan untuk keberadaan benda mati, seperti buku.",
		"あまり biasanya dipakai dengan bentuk negatif: tidak begitu besar, 大きくないです.",
		"Pola meminta izin adalah bentuk て + もいいですか: とってもいいですか.",
		"より menandai pembanding: bahasa Jepang lebih sulit daripada bahasa Inggris menurut kalimat ini.",
		"～たい melekat pada batang bentuk ます: 行きます → 行きたい.",
		"母が作った menerangkan ケーキ: kue yang dibuat ibu. が menandai pelakunya.",
		"八時に家を出ます menyatakan Tanaka keluar rumah pukul delapan.",
		"会社までバスで二十分: menuju kantor dengan bus selama dua puluh menit.",
		"十時に入口で会いましょう: bertemu di pintu masuk perpustakaan pukul sepuluh.",
		"一緒に勉強しませんか adalah ajakan belajar bersama Yui.",
		"Sabtu pukul 19.00 masih dalam jam buka 09.00–20.00, Senin sampai Sabtu. Minggu tutup.",
	} {
		if _, ok := englishExamSSWText("diagnostic.explanation", source); !ok {
			t.Errorf("diagnostic explanation has no exact English catalog entry: %q", source)
		}
	}
}

func TestEnglishExamCatalogPreservesJapaneseStudyText(t *testing.T) {
	got, ok := englishExamSSWText("diagnostic.explanation", "学校 dibaca がっこう, sekolah. Perhatikan bunyi kecil っ.")
	if !ok {
		t.Fatal("diagnostic explanation missing from catalog")
	}
	for _, japanese := range []string{"学校", "がっこう", "っ"} {
		if !strings.Contains(got, japanese) {
			t.Errorf("English explanation %q lost Japanese study text %q", got, japanese)
		}
	}
	if got, ok := englishExamSSWText("diagnostic.prompt", "「あさごはん」はどれですか。"); ok {
		t.Errorf("Japanese-only diagnostic prompt should stay unchanged by source fallback, got translation %q", got)
	}
}

func TestEnglishExamCatalogCoversAvailablePaperLabelsAndSharedCopy(t *testing.T) {
	for source, want := range map[string]string{
		"Kaigo · Simulasi Keterampilan 1":           "Kaigo · Skills Practice 1",
		"Kaigo · Simulasi Bahasa Jepang 1":          "Kaigo · Japanese Language Practice 1",
		"Pembersihan Gedung · Simulasi 1":           "Building Cleaning · Practice 1",
		"Pertanian · Tanaman · Simulasi Tertulis 1": "Agriculture · Crop Production · Written Practice 1",
		"Perikanan Tangkap · Simulasi 1":            "Capture Fisheries · Practice 1",
	} {
		assertExamTranslation(t, "ssw.paper.title", source, want)
	}
	assertExamTranslation(t, "simulation.instruction", "Pilih satu jawaban paling tepat. Soal orisinal Nihongo Master.", "Choose the single best answer. Original questions by Nihongo Master.")
}

func TestEnglishSSWAssessmentBanksCoverEverySourceFieldAndPreserveJapanese(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	wantSectorCounts := map[string]int{
		"accommodation": 36, "agriculture": 160, "automobile-maintenance": 36,
		"aviation": 36, "building-cleaning": 60, "construction": 36,
		"fishery": 120, "food-manufacturing": 36, "food-service": 36,
		"forestry": 36, "industrial-manufacturing": 36, "linen-supply": 48,
		"logistics-warehouse": 36, "nursing-care": 87, "railway": 36,
		"resource-recycling": 36, "road-transport": 36, "shipbuilding": 36,
		"wood-industry": 36,
	}
	if len(source.Banks) != len(wantSectorCounts) {
		t.Fatalf("source contains %d SSW banks, want %d; review this inventory", len(source.Banks), len(wantSectorCounts))
	}
	for sector, want := range wantSectorCounts {
		if got := len(source.Banks[sector]); got != want {
			t.Errorf("source bank %s has %d questions, want %d", sector, got, want)
		}
	}
	// These are Japanese-only learner prompts/options in the current bank. Keep
	// them as Japanese source text rather than treating them as missing English.
	preserve := map[string]map[string]bool{
		"ssw.prompt": {
			"利用者（りようしゃ）が「右（みぎ）の足（あし）が痛（いた）いです」と言（い）いました。どこが痛いですか。":    true,
			"利用者（りようしゃ）：「自分（じぶん）で顔（かお）を洗（あら）いたいです。」返事（へんじ）はどれがよいですか。": true,
			"利用者（りようしゃ）が「寒（さむ）いです」と言（い）いました。適切（てきせつ）な返事（へんじ）はどれですか。":  true,
			"利用者（りようしゃ）とは、だれですか。":                                                            true,
			"利用者（りようしゃ）に聞（き）こえなかったとき、何（なん）と言（い）いますか。":                                        true,
			"連絡（れんらく）：「田中（たなか）さんは、今日（きょう）は入浴（にゅうよく）しません。」正（ただ）しいのはどれですか。":                    true,
			"予定（よてい）：「午後（ごご）二時（じ）に体操（たいそう）。その前（まえ）にトイレを確認（かくにん）。」順番（じゅんばん）はどれですか。":           true,
			"「立（た）つ前（まえ）に、ブレーキを確認（かくにん）してください。」最初（さいしょ）に何（なに）をしますか。":                         true,
			"「むせる」とは、どのような様子（ようす）ですか。":                                                       true,
			"「排泄（はいせつ）」に関係（かんけい）があるのはどれですか。":                                                 true,
			"「車（くるま）いす」の読（よ）み方（かた）はどれですか。":                                                   true,
			"「食後（しょくご）」は、いつですか。":                                                             true,
			"ベッドから車（くるま）いすに移（うつ）る前（まえ）に、何（なん）と言（い）いますか。":                                     true,
			"記録（きろく）：「午前（ごぜん）十時（じ）、水（みず）を百（ひゃく）ミリリットル飲（の）んだ。」何（なに）を飲みましたか。":                  true,
			"申し送（もうしおく）り：「昼食後（ちゅうしょくご）、せきが続（つづ）いています。看護師（かんごし）に報告済（ほうこくず）みです。」報告はどうなっていますか。": true,
		},
		"ssw.option": {
			"おなか":          true,
			"お茶（ちゃ）を1リットル": true,
			"くるまいし":        true,
			"くるまいす":        true,
			"くるまいず":        true,
			"これから車いすに移りましょう。":    true,
			"ご飯を食べる前（まえ）":        true,
			"ご飯（はん）を食（た）べた後（あと）": true,
			"しゃいす":  true,
			"すぐに立つ": true,
			"すでに看護師に報告しました。":                      true,
			"まだ誰（だれ）にも報告していません。":                  true,
			"もう一度（いちど）お話（はな）しいただけますか。":            true,
			"もう食事（しょくじ）は終（お）わりました。":               true,
			"もっと冷（つめ）たい水（みず）を飲（の）んでください。":         true,
			"ゆっくり歩（ある）く":                          true,
			"トイレ":                                 true,
			"トイレの確認はしない":                          true,
			"トイレを確認してから、二時に体操をする":                 true,
			"ブレーキを確認する":                           true,
			"今日（きょう）は休（やす）みです。":                   true,
			"介護（かいご）のサービスを使（つか）う人（ひと）":            true,
			"体操をしてから、トイレを確認する":                    true,
			"何でもいいです。":                            true,
			"何（なに）も言わないでください。":                    true,
			"全員（ぜんいん）が入浴しません。":                    true,
			"分（わ）かりました。必要（ひつよう）なときはお手伝（てつだ）いします。": true,
			"利用者（りようしゃ）が報告を断（ことわ）りました。":           true,
			"午前二時に体操をする":                          true,
			"右の足":                                 true,
			"寒くありません。":                            true,
			"寝（ね）ている間（あいだ）":                       true,
			"左（ひだり）の手（て）":                         true,
			"建物（たてもの）を修理（しゅうり）する人":                true,
			"散歩（さんぽ）":                             true,
			"明日（あした）報告します。":                       true,
			"時間（じかん）がないので全部（ぜんぶ）私（わたし）がします。":      true,
			"服（ふく）を着替（きが）える":                      true,
			"毛布（もうふ）をお持（も）ちしましょうか。":               true,
			"気持（きも）ちよく眠（ねむ）る":                     true,
			"水を100ミリリットル":                         true,
			"洗濯（せんたく）":                            true,
			"牛乳（ぎゅうにゅう）を10ミリリットル":                 true,
			"田中さんは今日はお風呂（ふろ）に入りません。":              true,
			"田中さんは今（いま）入浴しています。":                  true,
			"田中さんは毎日（まいにち）入浴しません。":                true,
			"窓（まど）を閉（し）めました。":                     true,
			"絶対（ぜったい）に違（ちが）います。":                  true,
			"自分では何（なに）もしないでください。":                 true,
			"荷物（にもつ）を売（う）る人":                      true,
			"薬（くすり）を100個（こ）":                      true,
			"試験（しけん）を作（つく）る人":                     true,
			"話さないでください。":                          true,
			"買（か）い物（もの）":                          true,
			"起（お）きる前":                             true,
			"車（くるま）いすを動（うご）かす":                    true,
			"靴（くつ）を捨（す）てる":                        true,
			"頭（あたま）":                              true,
			"顔を洗ってはいけません。":                        true,
			"飲食中（いんしょくちゅう）などに、せきが出（で）る":           true,
		},
	}
	ids := make(map[string]bool, 979)
	answerRows := make([]string, 0, 979)
	questionCount := 0
	for sector, questions := range source.Banks {
		for _, q := range questions {
			questionCount++
			if q.ID == "" || ids[q.ID] {
				t.Errorf("missing or duplicate stable ID in SSW bank %s: %q", sector, q.ID)
			}
			ids[q.ID] = true
			if q.Answer < 0 || q.Answer >= len(q.Options) {
				t.Errorf("SSW bank question %s has answer index %d outside %d options", q.ID, q.Answer, len(q.Options))
			}
			answerRows = append(answerRows, fmt.Sprintf("%s\x00%s\x00%d\n", sector, q.ID, q.Answer))
			for field, value := range map[string]string{
				"ssw.group": q.Group, "ssw.prompt": q.Prompt, "ssw.explanation": q.Explanation,
				"ssw.section": q.Section, "ssw.passage": q.Passage,
			} {
				assertSSWBankFieldCoverage(t, q.ID, field, value, preserve)
			}
			for _, option := range q.Options {
				assertSSWBankFieldCoverage(t, q.ID, "ssw.option", option, preserve)
			}
		}
	}
	if questionCount != 979 {
		t.Fatalf("enumerated %d SSW bank questions, want 979", questionCount)
	}
	for field, allowed := range preserve {
		for value := range allowed {
			if _, ok := englishExamSSWText(field, value); ok {
				t.Errorf("Japanese-only SSW %s was translated; preserve source: %q", field, value)
			}
		}
	}
	sort.Strings(answerRows)
	answerDigest := sha256.Sum256([]byte(strings.Join(answerRows, "")))
	if got, want := hex.EncodeToString(answerDigest[:]), "e6b40559e045c605755556b23dbd4a1589611a7e9084bf664c9d68d9653fe0e2"; got != want {
		t.Errorf("authoritative SSW question IDs/answers changed: digest %s, want %s", got, want)
	}
}

func assertSSWBankFieldCoverage(t *testing.T, questionID, field, source string, preserve map[string]map[string]bool) {
	t.Helper()
	if source == "" {
		return
	}
	if preserve[field][source] {
		if _, ok := englishExamSSWText(field, source); ok {
			t.Errorf("SSW question %s Japanese-only %s should stay in Japanese: %q", questionID, field, source)
		}
		return
	}
	translated, ok := englishExamSSWText(field, source)
	if !ok {
		t.Errorf("SSW question %s has no exact English mapping for %s: %q", questionID, field, source)
		return
	}
	for _, r := range source {
		if isJapaneseStudyRune(r) && !strings.ContainsRune(translated, r) {
			t.Errorf("SSW question %s translation for %s %q lost Japanese study character %q: %q", questionID, field, source, r, translated)
		}
	}
}

func isJapaneseStudyRune(r rune) bool {
	return unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana) || r == 'ー'
}

func TestEnglishSSWAutomobileCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) < 2 {
		t.Fatal("materials inventory is empty")
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string {
		return row[header[name]]
	}
	var rows [][]string
	for _, row := range records[1:] {
		if value(row, "surface") == "SSW learning material" && value(row, "item_type") != "SSW source" && value(row, "sector") == "automobile-maintenance" && (value(row, "status") == "Needs translation" || value(row, "status") == "Exact mapping") {
			rows = append(rows, row)
		}
	}
	if len(rows) != 276 {
		t.Fatalf("automobile-maintenance material inventory changed: got %d translatable rows, want 276", len(rows))
	}
	sort.Slice(rows, func(i, j int) bool { return value(rows[i], "material_id") < value(rows[j], "material_id") })
	hash := sha256.New()
	seen := make(map[examSSWEnglishKey]bool)
	for _, row := range rows {
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%s\n", id, field, source)
		key := examSSWEnglishKey{field: "ssw." + field, source: source}
		translated, ok := translateSSWCurriculumMaterialText("automobile-maintenance", field, source)
		if !ok {
			t.Errorf("automobile material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, r := range translated {
			if isJapaneseStudyRune(r) {
				t.Errorf("automobile material %s translation contains Japanese study text: %q", id, translated)
				break
			}
		}
		seen[key] = true
	}
	const wantHash = "22c597fd0a838147086207ab5020b50ad09b7d5bb806e7225cfa6af08f315bdb"
	if got := hex.EncodeToString(hash.Sum(nil)); got != wantHash {
		t.Errorf("automobile material source IDs or fields changed: got SHA-256 %s, want %s", got, wantHash)
	}
	if len(seen) != 216 {
		t.Errorf("automobile curriculum has %d translated exact field/source pairs, want 216", len(seen))
	}
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "automobile-maintenance" || value(row, "status") != "Preserve Japanese" {
			continue
		}
		source := value(row, "source_text")
		if _, translated := translateSSWCurriculumMaterialText("automobile-maintenance", value(row, "field"), source); translated {
			t.Errorf("Japanese study field %s was translated: %q", value(row, "material_id"), source)
		}
	}
}

func TestEnglishSSWBuildingCleaningCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) < 2 {
		t.Fatal("materials inventory is empty")
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	var rows [][]string
	for _, row := range records[1:] {
		if value(row, "surface") == "SSW learning material" && value(row, "item_type") != "SSW source" && value(row, "sector") == "building-cleaning" && (value(row, "status") == "Needs translation" || value(row, "status") == "Exact mapping") {
			rows = append(rows, row)
		}
	}
	if len(rows) != 660 {
		t.Fatalf("building-cleaning material inventory changed: got %d translatable rows, want 660", len(rows))
	}
	sort.Slice(rows, func(i, j int) bool { return value(rows[i], "material_id") < value(rows[j], "material_id") })
	hash := sha256.New()
	seen := make(map[examSSWEnglishKey]bool)
	for _, row := range rows {
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%s\n", id, field, source)
		key := examSSWEnglishKey{field: "ssw." + field, source: source}
		translated, ok := translateSSWCurriculumMaterialText("building-cleaning", field, source)
		if !ok {
			t.Errorf("building-cleaning material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("building-cleaning material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seen[key] = true
	}
	const wantHash = "d7c08cc8468a1af6b4e7112b9a3a66fc6c2ee31ed44d6c8c507ef93a476004ff"
	if got := hex.EncodeToString(hash.Sum(nil)); got != wantHash {
		t.Errorf("building-cleaning material IDs or fields changed: got SHA-256 %s, want %s", got, wantHash)
	}
	if len(seen) != 531 {
		t.Errorf("building-cleaning curriculum has %d translated field/source pairs, want 531", len(seen))
	}
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "building-cleaning" || value(row, "status") != "Preserve Japanese" {
			continue
		}
		source := value(row, "source_text")
		if _, translated := translateSSWCurriculumMaterialText("building-cleaning", value(row, "field"), source); translated {
			t.Errorf("Japanese study field %s was translated: %q", value(row, "material_id"), source)
		}
	}
}

func TestEnglishSSWConstructionCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	var rows [][]string
	for _, row := range records[1:] {
		if value(row, "surface") == "SSW learning material" && value(row, "item_type") != "SSW source" && value(row, "sector") == "construction" && (value(row, "status") == "Needs translation" || value(row, "status") == "Exact mapping") {
			rows = append(rows, row)
		}
	}
	if len(rows) != 276 {
		t.Fatalf("construction material inventory changed: got %d translatable rows, want 276", len(rows))
	}
	sort.Slice(rows, func(i, j int) bool { return value(rows[i], "material_id") < value(rows[j], "material_id") })
	hash := sha256.New()
	seen := make(map[examSSWEnglishKey]bool)
	for _, row := range rows {
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%s\n", id, field, source)
		translated, ok := translateSSWCurriculumMaterialText("construction", field, source)
		if !ok {
			t.Errorf("construction material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("construction material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seen[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	const wantHash = "4601a9b68849b771455a3c7edfefe8ac0002fd78083a5407a56a49fc0e711bd6"
	if got := hex.EncodeToString(hash.Sum(nil)); got != wantHash {
		t.Errorf("construction material IDs or fields changed: got SHA-256 %s, want %s", got, wantHash)
	}
	if len(seen) != 150 {
		t.Errorf("construction curriculum has %d translated field/source pairs, want 150", len(seen))
	}
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "construction" || value(row, "status") != "Preserve Japanese" {
			continue
		}
		if _, translated := translateSSWCurriculumMaterialText("construction", value(row, "field"), value(row, "source_text")); translated {
			t.Errorf("Japanese study field %s was translated: %q", value(row, "material_id"), value(row, "source_text"))
		}
	}
}

func TestEnglishSSWIndustrialManufacturingCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	var rows [][]string
	for _, row := range records[1:] {
		if value(row, "surface") == "SSW learning material" && value(row, "item_type") != "SSW source" && value(row, "sector") == "industrial-manufacturing" && (value(row, "status") == "Needs translation" || value(row, "status") == "Exact mapping") {
			rows = append(rows, row)
		}
	}
	if len(rows) != 276 {
		t.Fatalf("industrial-manufacturing material inventory changed: got %d translatable rows, want 276", len(rows))
	}
	sort.Slice(rows, func(i, j int) bool { return value(rows[i], "material_id") < value(rows[j], "material_id") })
	hash := sha256.New()
	seen := make(map[examSSWEnglishKey]bool)
	for _, row := range rows {
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%s\n", id, field, source)
		translated, ok := translateSSWCurriculumMaterialText("industrial-manufacturing", field, source)
		if !ok {
			t.Errorf("industrial-manufacturing material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("industrial-manufacturing material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seen[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	const wantHash = "35fe369c2c5e7ffad5e1bd0a178289a812ca5717cee0571737917e3141b1c451"
	if got := hex.EncodeToString(hash.Sum(nil)); got != wantHash {
		t.Errorf("industrial-manufacturing material IDs or fields changed: got SHA-256 %s, want %s", got, wantHash)
	}
	if len(seen) != 150 {
		t.Errorf("industrial-manufacturing curriculum has %d translated field/source pairs, want 150", len(seen))
	}
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "industrial-manufacturing" || value(row, "status") != "Preserve Japanese" {
			continue
		}
		if _, translated := translateSSWCurriculumMaterialText("construction", value(row, "field"), value(row, "source_text")); translated {
			t.Errorf("Japanese study field %s was translated: %q", value(row, "material_id"), value(row, "source_text"))
		}
	}
}

func japaneseStudyRuns(source string) []string {
	var runs []string
	var current strings.Builder
	flush := func() {
		if current.Len() > 0 {
			runs = append(runs, current.String())
			current.Reset()
		}
	}
	for _, r := range source {
		if isJapaneseStudyRune(r) {
			current.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return runs
}

func TestEnglishSSWShipbuildingCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "shipbuilding" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("shipbuilding", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("shipbuilding", field, source)
		if !ok {
			t.Errorf("shipbuilding material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("shipbuilding material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 276 {
		t.Errorf("shipbuilding curriculum has %d translated source IDs, want 276", len(seenIDs))
	}
	if len(seenPairs) != 147 {
		t.Errorf("shipbuilding curriculum has %d translated field/source pairs, want 147", len(seenPairs))
	}
}

func TestEnglishSSWAviationCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "aviation" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("aviation", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("aviation", field, source)
		if !ok {
			t.Errorf("aviation material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("aviation material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 276 {
		t.Errorf("aviation curriculum has %d translated source IDs, want 276", len(seenIDs))
	}
	if len(seenPairs) != 214 {
		t.Errorf("aviation curriculum has %d translated field/source pairs, want 214", len(seenPairs))
	}
}

func assertExamTranslation(t *testing.T, field, source, want string) {
	t.Helper()
	got, ok := englishExamSSWText(field, source)
	if !ok {
		t.Fatalf("missing English catalog entry for %s source %q", field, source)
	}
	if got != want {
		t.Errorf("translation for %s source %q = %q, want %q", field, source, got, want)
	}
}

func TestEnglishSSWForestryCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "forestry" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("forestry", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("forestry", field, source)
		if !ok {
			t.Errorf("forestry material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("forestry material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 276 {
		t.Errorf("forestry curriculum has %d translated source IDs, want 276", len(seenIDs))
	}
	if len(seenPairs) != 213 {
		t.Errorf("forestry curriculum has %d translated field/source pairs, want 213", len(seenPairs))
	}
}

func TestEnglishSSWResourceRecyclingCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "resource-recycling" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("resource-recycling", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("resource-recycling", field, source)
		if !ok {
			t.Errorf("resource-recycling material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("resource-recycling material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 276 {
		t.Errorf("resource-recycling curriculum has %d translated source IDs, want 276", len(seenIDs))
	}
	if len(seenPairs) != 216 {
		t.Errorf("resource-recycling curriculum has %d translated field/source pairs, want 216", len(seenPairs))
	}
}

func TestEnglishSSWFoodManufacturingCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "food-manufacturing" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("food-manufacturing", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("food-manufacturing", field, source)
		if !ok {
			t.Errorf("food-manufacturing material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("food-manufacturing material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 276 {
		t.Errorf("food-manufacturing curriculum has %d translated source IDs, want 276", len(seenIDs))
	}
	if len(seenPairs) != 215 {
		t.Errorf("food-manufacturing curriculum has %d translated field/source pairs, want 215", len(seenPairs))
	}
}

func TestEnglishSSWKaigoCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "kaigo" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("kaigo", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("kaigo", field, source)
		if !ok {
			t.Errorf("kaigo material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("kaigo material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 806 {
		t.Errorf("kaigo curriculum has %d translated source IDs, want 806", len(seenIDs))
	}
	if len(seenPairs) != 688 {
		t.Errorf("kaigo curriculum has %d translated field/source pairs, want 688", len(seenPairs))
	}
}

func TestEnglishSSWFisheryCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "fishery" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("fishery", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("fishery", field, source)
		if !ok {
			t.Errorf("fishery material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("fishery material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 690 {
		t.Errorf("fishery curriculum has %d translated source IDs, want 690", len(seenIDs))
	}
	if len(seenPairs) != 469 {
		t.Errorf("fishery curriculum has %d translated field/source pairs, want 469", len(seenPairs))
	}
}

func TestEnglishSSWRoadTransportCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "road-transport" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("road-transport", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("road-transport", field, source)
		if !ok {
			t.Errorf("road-transport material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("road-transport material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 277 {
		t.Errorf("road-transport curriculum has %d translated source IDs, want 277", len(seenIDs))
	}
	if len(seenPairs) != 237 {
		t.Errorf("road-transport curriculum has %d translated field/source pairs, want 237", len(seenPairs))
	}
}

func TestEnglishSSWLogisticsWarehouseCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "logistics-warehouse" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("logistics-warehouse", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("logistics-warehouse", field, source)
		if !ok {
			t.Errorf("logistics-warehouse material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("logistics-warehouse material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 276 {
		t.Errorf("logistics-warehouse curriculum has %d translated source IDs, want 276", len(seenIDs))
	}
	if len(seenPairs) != 236 {
		t.Errorf("logistics-warehouse curriculum has %d translated field/source pairs, want 236", len(seenPairs))
	}
}

func TestEnglishSSWLinenSupplyCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "linen-supply" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("linen-supply", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("linen-supply", field, source)
		if !ok {
			t.Errorf("linen-supply material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("linen-supply material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 276 {
		t.Errorf("linen-supply curriculum has %d translated source IDs, want 276", len(seenIDs))
	}
	if len(seenPairs) != 199 {
		t.Errorf("linen-supply curriculum has %d translated field/source pairs, want 199", len(seenPairs))
	}
}

func TestEnglishSSWWoodIndustryCurriculumMaterialCoverage(t *testing.T) {
	path := filepath.Join("..", "..", "..", "docs", "english-localization", "materials.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	header := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		header[name] = i
	}
	value := func(row []string, name string) string { return row[header[name]] }
	seenIDs := make(map[string]bool)
	seenPairs := make(map[examSSWEnglishKey]bool)
	for _, row := range records[1:] {
		if value(row, "surface") != "SSW learning material" || value(row, "item_type") == "SSW source" || value(row, "sector") != "wood-industry" {
			continue
		}
		id, field, source := value(row, "material_id"), value(row, "field"), value(row, "source_text")
		if value(row, "status") == "Preserve Japanese" {
			if _, translated := translateSSWCurriculumMaterialText("wood-industry", field, source); translated {
				t.Errorf("Japanese study field %s was translated: %q", id, source)
			}
			continue
		}
		translated, ok := translateSSWCurriculumMaterialText("wood-industry", field, source)
		if !ok {
			t.Errorf("wood-industry material %s (%s) has no exact English translation for %q", id, field, source)
			continue
		}
		for _, japaneseRun := range japaneseStudyRuns(source) {
			if !strings.Contains(translated, japaneseRun) {
				t.Errorf("wood-industry material %s translation dropped Japanese study text %q: %q", id, japaneseRun, translated)
			}
		}
		seenIDs[id] = true
		seenPairs[examSSWEnglishKey{field: "ssw." + field, source: source}] = true
	}
	if len(seenIDs) != 276 {
		t.Errorf("wood-industry curriculum has %d translated source IDs, want 276", len(seenIDs))
	}
	if len(seenPairs) != 213 {
		t.Errorf("wood-industry curriculum has %d translated field/source pairs, want 213", len(seenPairs))
	}
}

func TestEnglishJLPTSimulationCatalogCoversEveryIndonesianExplanation(t *testing.T) {
	s := store.NewMemoryStore()
	defer s.Close()
	const user = "english-jlpt-catalog-coverage"
	token, err := s.BillingAccountToken(user)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SaveVerifiedPurchase(user, domain.Purchase{
		Platform: "android", ID: "english-jlpt-catalog-coverage", ProductID: "nihongo_master_lifetime",
		AccountToken: token, Proof: "coverage-test", State: "purchased", CheckedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	levels := []string{"N5", "N4", "N3", "N2", "N1"}
	questionCount := 0
	translatedExplanations := map[string]bool{}
	for _, level := range levels {
		attempt, err := s.StartSimulation(user, level)
		if err != nil {
			t.Fatalf("start %s simulation: %v", level, err)
		}
		session, err := s.SimulationSession(user, attempt.ID)
		if err != nil {
			t.Fatalf("get %s simulation: %v", level, err)
		}
		for _, section := range session.Sections {
			session, err = s.SubmitSimulationSection(user, attempt.ID, section.ID)
			if err != nil {
				t.Fatalf("advance %s simulation section %s: %v", level, section.ID, err)
			}
		}
		if session.Result == nil {
			t.Fatalf("%s simulation did not return review explanations", level)
		}
		for _, review := range session.Result.Reviews {
			question := review.Question
			questionCount++
			for field, source := range map[string]string{
				"simulation.group":       question.Group,
				"simulation.instruction": question.Instructions,
				"simulation.passage":     question.Passage,
				"simulation.prompt":      question.Prompt,
			} {
				if source != "" {
					if _, ok := englishExamSSWText(field, source); ok {
						t.Errorf("Japanese JLPT %s field unexpectedly has a translation; preserve source exactly: %q", field, source)
					}
				}
			}
			for _, option := range question.Options {
				if _, ok := englishExamSSWText("simulation.choice", option.Text); ok {
					t.Errorf("Japanese JLPT choice unexpectedly has a translation; preserve source exactly: %q", option.Text)
				}
			}
			_, mapped := englishExamSSWText("simulation.explanation", review.Explanation)
			if level == "N5" || level == "N4" || level == "N3" {
				if !mapped {
					t.Errorf("JLPT %s explanation has no exact catalog entry: %q", level, review.Explanation)
				} else {
					translatedExplanations[review.Explanation] = true
				}
			} else if mapped {
				t.Errorf("Japanese-only JLPT %s explanation should remain source text: %q", level, review.Explanation)
			}
		}
	}
	if questionCount != 297 {
		t.Errorf("enumerated %d JLPT questions, want 297 across N1–N5", questionCount)
	}
	if len(translatedExplanations) != 169 {
		t.Errorf("catalog covered %d distinct Indonesian JLPT explanations, want 169", len(translatedExplanations))
	}
}

func TestEnglishJLPTExplanationKeepsJapaneseStudyText(t *testing.T) {
	got, ok := englishExamSSWText("simulation.explanation", "電車 dibaca でんしゃ, kereta listrik. Huruf kecil ゃ membentuk bunyi しゃ; しや bukan bacaan yang sama.")
	if !ok {
		t.Fatal("JLPT explanation missing from exact catalog")
	}
	for _, japanese := range []string{"電車", "でんしゃ", "ゃ", "しゃ", "しや"} {
		if !strings.Contains(got, japanese) {
			t.Errorf("English explanation %q lost Japanese source text %q", got, japanese)
		}
	}
}

func TestEnglishSSWAccommodationCatalogCoversBankAndPaperOccurrences(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	type paper struct {
		Items []question `json:"items"`
	}
	var source struct {
		Banks  map[string][]question `json:"banks"`
		Papers map[string]paper      `json:"papers"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}

	bank := source.Banks["accommodation"]
	paperItems := source.Papers["SSW-ACCOMMODATION"].Items
	if len(bank) != 36 || len(paperItems) != 24 {
		t.Fatalf("source size changed: accommodation bank=%d paper=%d; review the source scope", len(bank), len(paperItems))
	}
	all := append(append([]question(nil), bank...), paperItems...)
	seenFields := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range all {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group":       q.Group,
			"ssw.prompt":      q.Prompt,
			"ssw.explanation": q.Explanation,
			"ssw.section":     q.Section,
			"ssw.passage":     q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("accommodation question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seenFields[field] == nil {
				seenFields[field] = map[string]bool{}
			}
			seenFields[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("accommodation question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seenFields["ssw.option"] == nil {
				seenFields["ssw.option"] = map[string]bool{}
			}
			seenFields["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for question %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 4, "ssw.prompt": 36, "ssw.explanation": 36,
		"ssw.section": 2, "ssw.passage": 12, "ssw.option": 144,
	}
	for field, want := range wantCounts {
		if got := len(seenFields[field]); got != want {
			t.Errorf("enumerated %d unique accommodation %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 36 {
		t.Errorf("enumerated %d stable accommodation question IDs across banks and papers, want 36", len(ids))
	}
}

func TestEnglishSSWAccommodationPreservesJapaneseSources(t *testing.T) {
	for _, source := range []string{"accommodation-01-1", "Data apa yang dicocokkan saat mencari reservasi?"} {
		if source == "accommodation-01-1" {
			continue
		}
		if _, ok := englishExamSSWText("ssw.prompt", source); !ok {
			t.Errorf("expected accommodation prompt to have English mapping: %q", source)
		}
	}
	for _, japanese := range []string{"客室", "予約"} {
		if _, ok := englishExamSSWText("ssw.prompt", japanese); ok {
			t.Errorf("Japanese study text should not be translated away: %q", japanese)
		}
	}
}

func TestEnglishSSWAgricultureCropFoundationsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["agriculture"] {
		if q.ModuleID == "crop-foundations" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 16 {
		t.Fatalf("source size changed: crop-foundations=%d, want 16", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 16, "ssw.explanation": 16,
		"ssw.section": 2, "ssw.passage": 4, "ssw.option": 64,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique crop-foundations %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 16 {
		t.Errorf("enumerated %d stable crop-foundations question IDs, want 16", len(ids))
	}
}

func TestEnglishSSWAgricultureCropNutritionCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["agriculture"] {
		if q.ModuleID == "crop-nutrition" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 16 {
		t.Fatalf("source size changed: crop-nutrition=%d, want 16", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 16, "ssw.explanation": 16,
		"ssw.section": 2, "ssw.passage": 4, "ssw.option": 64,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique crop-nutrition %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 16 {
		t.Errorf("enumerated %d stable crop-nutrition question IDs, want 16", len(ids))
	}
	for source, japanese := range map[string]string{
		`Dalam konteks lesson ini, apa arti 窒素（ちっそ）?`: `窒素（ちっそ）`,
		`Dalam konteks lesson ini, apa arti 湿度（しつど）?`: `湿度（しつど）`,
	} {
		got, ok := englishExamSSWText("ssw.prompt", source)
		if !ok || !strings.Contains(got, japanese) {
			t.Errorf("English prompt %q did not preserve Japanese source %q: %q", source, japanese, got)
		}
	}
}

func TestEnglishSSWAgricultureCropEstablishmentCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["agriculture"] {
		if q.ModuleID == "crop-establishment" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 16 {
		t.Fatalf("source size changed: crop-establishment=%d, want 16", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 16, "ssw.explanation": 16,
		"ssw.section": 2, "ssw.passage": 4, "ssw.option": 64,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique crop-establishment %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 16 {
		t.Errorf("enumerated %d stable crop-establishment question IDs, want 16", len(ids))
	}
	for source, japanese := range map[string]string{
		`Dalam konteks lesson ini, apa arti 発芽（はつが）?`: `発芽（はつが）`,
		`Dalam konteks lesson ini, apa arti 根鉢（ねばち）?`: `根鉢（ねばち）`,
	} {
		got, ok := englishExamSSWText("ssw.prompt", source)
		if !ok || !strings.Contains(got, japanese) {
			t.Errorf("English prompt %q did not preserve Japanese source %q: %q", source, japanese, got)
		}
	}
}

func TestEnglishSSWAgricultureCropProtectionCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["agriculture"] {
		if q.ModuleID == "crop-protection" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 16 {
		t.Fatalf("source size changed: crop-protection=%d, want 16", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 16, "ssw.explanation": 16,
		"ssw.section": 2, "ssw.passage": 4, "ssw.option": 64,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique crop-protection %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 16 {
		t.Errorf("enumerated %d stable crop-protection question IDs, want 16", len(ids))
	}
	for source, japanese := range map[string]string{
		`Dalam konteks lesson ini, apa arti 雑草（ざっそう）?`:      `雑草（ざっそう）`,
		`Dalam konteks lesson ini, apa arti 安全装置（あんぜんそうち）?`: `安全装置（あんぜんそうち）`,
	} {
		got, ok := englishExamSSWText("ssw.prompt", source)
		if !ok || !strings.Contains(got, japanese) {
			t.Errorf("English prompt %q did not preserve Japanese source %q: %q", source, japanese, got)
		}
	}
}

func TestEnglishSSWAgricultureCropHarvestCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["agriculture"] {
		if q.ModuleID == "crop-harvest" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 16 {
		t.Fatalf("source size changed: crop-harvest=%d, want 16", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 16, "ssw.explanation": 16,
		"ssw.section": 2, "ssw.passage": 4, "ssw.option": 64,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique crop-harvest %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 16 {
		t.Errorf("enumerated %d stable crop-harvest question IDs, want 16", len(ids))
	}
	for source, japanese := range map[string]string{
		`Dalam konteks lesson ini, apa arti 摘果（てきか）?`:  `摘果（てきか）`,
		`Dalam konteks lesson ini, apa arti 貯蔵（ちょぞう）?`: `貯蔵（ちょぞう）`,
	} {
		got, ok := englishExamSSWText("ssw.prompt", source)
		if !ok || !strings.Contains(got, japanese) {
			t.Errorf("English prompt %q did not preserve Japanese source %q: %q", source, japanese, got)
		}
	}
}

func TestEnglishSSWAgricultureLivestockFoundationsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["agriculture"] {
		if q.ModuleID == "livestock-foundations" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 16 {
		t.Fatalf("source size changed: livestock-foundations=%d, want 16", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 16, "ssw.explanation": 16,
		"ssw.section": 2, "ssw.passage": 4, "ssw.option": 64,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique livestock-foundations %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 16 {
		t.Errorf("enumerated %d stable livestock-foundations question IDs, want 16", len(ids))
	}
	for source, japanese := range map[string]string{
		`Dalam konteks lesson ini, apa arti 個体（こたい）?`:  `個体（こたい）`,
		`Dalam konteks lesson ini, apa arti 飲水（いんすい）?`: `飲水（いんすい）`,
	} {
		got, ok := englishExamSSWText("ssw.prompt", source)
		if !ok || !strings.Contains(got, japanese) {
			t.Errorf("English prompt %q did not preserve Japanese source %q: %q", source, japanese, got)
		}
	}
}

func TestEnglishSSWAgricultureLivestockEnvironmentCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["agriculture"] {
		if q.ModuleID == "livestock-environment" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 16 {
		t.Fatalf("source size changed: livestock-environment=%d, want 16", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 16, "ssw.explanation": 16,
		"ssw.section": 2, "ssw.passage": 4, "ssw.option": 64,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique livestock-environment %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 16 {
		t.Errorf("enumerated %d stable livestock-environment question IDs, want 16", len(ids))
	}
	for source, japanese := range map[string]string{
		`Dalam konteks lesson ini, apa arti 清掃（せいそう）?`: `清掃（せいそう）`,
		`Dalam konteks lesson ini, apa arti 隔離（かくり）?`:  `隔離（かくり）`,
	} {
		got, ok := englishExamSSWText("ssw.prompt", source)
		if !ok || !strings.Contains(got, japanese) {
			t.Errorf("English prompt %q did not preserve Japanese source %q: %q", source, japanese, got)
		}
	}
}

func TestEnglishSSWAgricultureLivestockCattleCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["agriculture"] {
		if q.ModuleID == "livestock-cattle" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 16 {
		t.Fatalf("source size changed: livestock-cattle=%d, want 16", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 16, "ssw.explanation": 16,
		"ssw.section": 2, "ssw.passage": 4, "ssw.option": 64,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique livestock-cattle %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 16 {
		t.Errorf("enumerated %d stable livestock-cattle question IDs, want 16", len(ids))
	}
	for source, japanese := range map[string]string{
		`Dalam konteks lesson ini, apa arti 出産（しゅっさん）?`: `出産（しゅっさん）`,
		`Dalam konteks lesson ini, apa arti 冷却（れいきゃく）?`: `冷却（れいきゃく）`,
	} {
		got, ok := englishExamSSWText("ssw.prompt", source)
		if !ok || !strings.Contains(got, japanese) {
			t.Errorf("English prompt %q did not preserve Japanese source %q: %q", source, japanese, got)
		}
	}
}

func TestEnglishSSWAgricultureLivestockProductionCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["agriculture"] {
		if q.ModuleID == "livestock-production" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 16 {
		t.Fatalf("source size changed: livestock-production=%d, want 16", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 16, "ssw.explanation": 16,
		"ssw.section": 2, "ssw.passage": 4, "ssw.option": 64,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique livestock-production %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 16 {
		t.Errorf("enumerated %d stable livestock-production question IDs, want 16", len(ids))
	}
	for source, japanese := range map[string]string{
		`Dalam konteks lesson ini, apa arti 増体（ぞうたい）?`: `増体（ぞうたい）`,
		`Dalam konteks lesson ini, apa arti 鶏（にわとり）?`:  `鶏（にわとり）`,
	} {
		got, ok := englishExamSSWText("ssw.prompt", source)
		if !ok || !strings.Contains(got, japanese) {
			t.Errorf("English prompt %q did not preserve Japanese source %q: %q", source, japanese, got)
		}
	}
}

func TestEnglishSSWAgricultureLivestockProductsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["agriculture"] {
		if q.ModuleID == "livestock-products" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 16 {
		t.Fatalf("source size changed: livestock-products=%d, want 16", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 16, "ssw.explanation": 16,
		"ssw.section": 2, "ssw.passage": 4, "ssw.option": 64,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique livestock-products %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 16 {
		t.Errorf("enumerated %d stable livestock-products question IDs, want 16", len(ids))
	}
	for source, japanese := range map[string]string{
		`Dalam konteks lesson ini, apa arti 産卵（さんらん）?`: `産卵（さんらん）`,
		`Dalam konteks lesson ini, apa arti 指示（しじ）?`:   `指示（しじ）`,
	} {
		got, ok := englishExamSSWText("ssw.prompt", source)
		if !ok || !strings.Contains(got, japanese) {
			t.Errorf("English prompt %q did not preserve Japanese source %q: %q", source, japanese, got)
		}
	}
}

func TestEnglishSSWAutomobileIntakeCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["automobile-maintenance"] {
		if q.ModuleID == "intake" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: automobile intake=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique automobile-intake %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable automobile-intake question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWAutomobileInspectionCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["automobile-maintenance"] {
		if q.ModuleID == "inspection" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: automobile inspection=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique automobile-inspection %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable automobile-inspection question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWAutomobileRepairCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["automobile-maintenance"] {
		if q.ModuleID == "repair" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: automobile repair=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique automobile-repair %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable automobile-repair question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWAutomobileQualityCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["automobile-maintenance"] {
		if q.ModuleID == "quality" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: automobile quality=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique automobile-quality %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable automobile-quality question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWAviationApronCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["aviation"] {
		if q.ModuleID == "apron" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: aviation apron=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique aviation-apron %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable aviation-apron question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWAviationBaggageCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["aviation"] {
		if q.ModuleID == "baggage" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: aviation baggage=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique aviation-baggage %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable aviation-baggage question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWAviationServicesCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["aviation"] {
		if q.ModuleID == "services" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: aviation services=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique aviation-services %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable aviation-services question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWAviationMaintenanceCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["aviation"] {
		if q.ModuleID == "maintenance" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: aviation maintenance=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique aviation-maintenance %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable aviation-maintenance question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWBuildingCleaningRoleCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["building-cleaning"] {
		if q.ModuleID == "role" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 6 {
		t.Fatalf("source size changed: building-cleaning role=%d, want 6", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 6, "ssw.explanation": 6,
		"ssw.section": 1, "ssw.option": 24,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique building-cleaning-role %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 6 {
		t.Errorf("enumerated %d stable building-cleaning-role question IDs, want 6", len(ids))
	}
}

func TestEnglishSSWBuildingCleaningSafetyCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["building-cleaning"] {
		if q.ModuleID == "safety" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 6 {
		t.Fatalf("source size changed: building-cleaning safety=%d, want 6", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 6, "ssw.explanation": 6,
		"ssw.section": 1, "ssw.option": 24,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique building-cleaning-safety %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 6 {
		t.Errorf("enumerated %d stable building-cleaning-safety question IDs, want 6", len(ids))
	}
}

func TestEnglishSSWBuildingCleaningChemicalsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["building-cleaning"] {
		if q.ModuleID == "chemicals" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 6 {
		t.Fatalf("source size changed: building-cleaning chemicals=%d, want 6", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 6, "ssw.explanation": 6,
		"ssw.section": 1, "ssw.option": 24,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique building-cleaning-chemicals %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 6 {
		t.Errorf("enumerated %d stable building-cleaning-chemicals question IDs, want 6", len(ids))
	}
}

func TestEnglishSSWBuildingCleaningToolsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["building-cleaning"] {
		if q.ModuleID == "tools" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 6 {
		t.Fatalf("source size changed: building-cleaning tools=%d, want 6", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 6, "ssw.explanation": 6,
		"ssw.section": 1, "ssw.option": 24,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique building-cleaning-tools %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 6 {
		t.Errorf("enumerated %d stable building-cleaning-tools question IDs, want 6", len(ids))
	}
}

func TestEnglishSSWBuildingCleaningFloorsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["building-cleaning"] {
		if q.ModuleID == "floors" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 6 {
		t.Fatalf("source size changed: building-cleaning floors=%d, want 6", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 6, "ssw.explanation": 6,
		"ssw.section": 1, "ssw.option": 24,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique building-cleaning-floors %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 6 {
		t.Errorf("enumerated %d stable building-cleaning-floors question IDs, want 6", len(ids))
	}
}

func TestEnglishSSWBuildingCleaningSanitaryCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["building-cleaning"] {
		if q.ModuleID == "sanitary" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 6 {
		t.Fatalf("source size changed: building-cleaning sanitary=%d, want 6", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 6, "ssw.explanation": 6,
		"ssw.section": 1, "ssw.option": 24,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique building-cleaning-sanitary %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 6 {
		t.Errorf("enumerated %d stable building-cleaning-sanitary question IDs, want 6", len(ids))
	}
}

func TestEnglishSSWBuildingCleaningGlassCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["building-cleaning"] {
		if q.ModuleID == "glass" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 6 {
		t.Fatalf("source size changed: building-cleaning glass=%d, want 6", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 6, "ssw.explanation": 6,
		"ssw.section": 1, "ssw.option": 24,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique building-cleaning-glass %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 6 {
		t.Errorf("enumerated %d stable building-cleaning-glass question IDs, want 6", len(ids))
	}
}

func TestEnglishSSWBuildingCleaningWasteCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["building-cleaning"] {
		if q.ModuleID == "waste" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 6 {
		t.Fatalf("source size changed: building-cleaning waste=%d, want 6", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 6, "ssw.explanation": 6,
		"ssw.section": 1, "ssw.option": 24,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique building-cleaning-waste %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 6 {
		t.Errorf("enumerated %d stable building-cleaning-waste question IDs, want 6", len(ids))
	}
}

func TestEnglishSSWBuildingCleaningPlanningCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["building-cleaning"] {
		if q.ModuleID == "planning" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 6 {
		t.Fatalf("source size changed: building-cleaning planning=%d, want 6", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 6, "ssw.explanation": 6,
		"ssw.section": 1, "ssw.option": 24,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique building-cleaning-planning %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 6 {
		t.Errorf("enumerated %d stable building-cleaning-planning question IDs, want 6", len(ids))
	}
}

func TestEnglishSSWConstructionSiteSafetyCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["construction"] {
		if q.ModuleID == "site-safety" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: construction site-safety=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique construction-site-safety %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable construction-site-safety question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWConstructionDrawingsLayoutCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["construction"] {
		if q.ModuleID == "drawings-layout" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: construction drawings-layout=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique construction-drawings-layout %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable construction-drawings-layout question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWConstructionExecutionCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["construction"] {
		if q.ModuleID == "execution" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: construction execution=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique construction-execution %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable construction-execution question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWConstructionQualityHandoverCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["construction"] {
		if q.ModuleID == "quality-handover" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: construction quality-handover=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 36,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique construction-quality-handover %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable construction-quality-handover question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWFisheryCaptureGeneralCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["fishery"] {
		if q.ModuleID == "capture-general" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 12 {
		t.Fatalf("source size changed: fishery capture-general=%d, want 12", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 12, "ssw.explanation": 12,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 48,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique fishery-capture-general %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 12 {
		t.Errorf("enumerated %d stable fishery-capture-general question IDs, want 12", len(ids))
	}
}

func TestEnglishSSWFisheryCaptureSafetyCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["fishery"] {
		if q.ModuleID == "capture-safety" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 12 {
		t.Fatalf("source size changed: fishery capture-safety=%d, want 12", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 12, "ssw.explanation": 12,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 48,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique fishery-capture-safety %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 12 {
		t.Errorf("enumerated %d stable fishery-capture-safety question IDs, want 12", len(ids))
	}
}

func TestEnglishSSWFisheryCaptureEquipmentCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["fishery"] {
		if q.ModuleID == "capture-equipment" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 12 {
		t.Fatalf("source size changed: fishery capture-equipment=%d, want 12", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 12, "ssw.explanation": 12,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 48,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique fishery-capture-equipment %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 12 {
		t.Errorf("enumerated %d stable fishery-capture-equipment question IDs, want 12", len(ids))
	}
}

func TestEnglishSSWFisheryCaptureMethodsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["fishery"] {
		if q.ModuleID == "capture-methods" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 12 {
		t.Fatalf("source size changed: fishery capture-methods=%d, want 12", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 12, "ssw.explanation": 12,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 48,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique fishery-capture-methods %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 12 {
		t.Errorf("enumerated %d stable fishery-capture-methods question IDs, want 12", len(ids))
	}
}

func TestEnglishSSWFisheryCaptureProductsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["fishery"] {
		if q.ModuleID == "capture-products" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 12 {
		t.Fatalf("source size changed: fishery capture-products=%d, want 12", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 12, "ssw.explanation": 12,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 48,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique fishery-capture-products %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 12 {
		t.Errorf("enumerated %d stable fishery-capture-products question IDs, want 12", len(ids))
	}
}

func TestEnglishSSWFisheryAquacultureGeneralCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["fishery"] {
		if q.ModuleID == "aqua-general" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 12 {
		t.Fatalf("source size changed: fishery aqua-general=%d, want 12", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 12, "ssw.explanation": 12,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 48,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique fishery-aqua-general %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 12 {
		t.Errorf("enumerated %d stable fishery-aqua-general question IDs, want 12", len(ids))
	}
}

func TestEnglishSSWFisheryAquacultureStockCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["fishery"] {
		if q.ModuleID == "aqua-stock" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 12 {
		t.Fatalf("source size changed: fishery aqua-stock=%d, want 12", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 12, "ssw.explanation": 12,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 48,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique fishery-aqua-stock %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 12 {
		t.Errorf("enumerated %d stable fishery-aqua-stock question IDs, want 12", len(ids))
	}
}

func TestEnglishSSWFisheryAquacultureFeedCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["fishery"] {
		if q.ModuleID == "aqua-feed" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 12 {
		t.Fatalf("source size changed: fishery aqua-feed=%d, want 12", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 12, "ssw.explanation": 12,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 48,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique fishery-aqua-feed %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 12 {
		t.Errorf("enumerated %d stable fishery-aqua-feed question IDs, want 12", len(ids))
	}
}

func TestEnglishSSWFisheryAquacultureSystemsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["fishery"] {
		if q.ModuleID == "aqua-systems" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 12 {
		t.Fatalf("source size changed: fishery aqua-systems=%d, want 12", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 12, "ssw.explanation": 12,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 48,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique fishery-aqua-systems %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 12 {
		t.Errorf("enumerated %d stable fishery-aqua-systems question IDs, want 12", len(ids))
	}
}

func TestEnglishSSWFisheryAquacultureProductsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["fishery"] {
		if q.ModuleID == "aqua-products" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 12 {
		t.Fatalf("source size changed: fishery aqua-products=%d, want 12", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 12, "ssw.explanation": 12,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 48,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique fishery-aqua-products %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 12 {
		t.Errorf("enumerated %d stable fishery-aqua-products question IDs, want 12", len(ids))
	}
}

func TestEnglishSSWFoodManufacturingHygieneCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["food-manufacturing"] {
		if q.ModuleID == "hygiene" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: food-manufacturing hygiene=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 30,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique food-manufacturing-hygiene %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable food-manufacturing-hygiene question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWFoodManufacturingMaterialsCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["food-manufacturing"] {
		if q.ModuleID == "materials" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: food-manufacturing materials=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 30,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique food-manufacturing-materials %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable food-manufacturing-materials question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWFoodManufacturingProcessingCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["food-manufacturing"] {
		if q.ModuleID == "processing" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: food-manufacturing processing=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 30,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique food-manufacturing-processing %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable food-manufacturing-processing question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWFoodManufacturingQualityCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["food-manufacturing"] {
		if q.ModuleID == "quality" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: food-manufacturing quality=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 30,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique food-manufacturing-quality %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable food-manufacturing-quality question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWFoodServiceHygieneCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["food-service"] {
		if q.ModuleID == "hygiene" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: food-service hygiene=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 30,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique food-service-hygiene %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable food-service-hygiene question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWFoodServicePreparationCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["food-service"] {
		if q.ModuleID == "preparation" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: food-service preparation=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 30,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique food-service-preparation %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable food-service-preparation question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWFoodServiceHospitalityCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["food-service"] {
		if q.ModuleID == "hospitality" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: food-service hospitality=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 30,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique food-service-hospitality %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable food-service-hospitality question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWFoodServiceClosingCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["food-service"] {
		if q.ModuleID == "closing" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: food-service closing=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 30,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique food-service-closing %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable food-service-closing question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWForestrySiteCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["forestry"] {
		if q.ModuleID == "site" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: forestry site=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 30,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique forestry-site %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable forestry-site question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWForestrySilvicultureCoversEverySourceField(t *testing.T) {
	type question struct {
		ID          string   `json:"id"`
		ModuleID    string   `json:"moduleId"`
		Group       string   `json:"group"`
		Prompt      string   `json:"prompt"`
		Options     []string `json:"options"`
		Answer      int      `json:"answer"`
		Explanation string   `json:"explanation"`
		Section     string   `json:"section"`
		Passage     string   `json:"passage"`
	}
	var source struct {
		Banks map[string][]question `json:"banks"`
	}
	raw, err := os.ReadFile("../store/ssw_resources/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	var questions []question
	for _, q := range source.Banks["forestry"] {
		if q.ModuleID == "silviculture" {
			questions = append(questions, q)
		}
	}
	if len(questions) != 9 {
		t.Fatalf("source size changed: forestry silviculture=%d, want 9", len(questions))
	}
	seen := map[string]map[string]bool{}
	ids := map[string]bool{}
	for _, q := range questions {
		stableID, answer := q.ID, q.Answer
		for field, value := range map[string]string{
			"ssw.group": q.Group, "ssw.prompt": q.Prompt,
			"ssw.explanation": q.Explanation, "ssw.section": q.Section,
			"ssw.passage": q.Passage,
		} {
			if value == "" {
				continue
			}
			if _, ok := englishExamSSWText(field, value); !ok {
				t.Errorf("question %s has no exact catalog entry for %s: %q", q.ID, field, value)
			}
			if seen[field] == nil {
				seen[field] = map[string]bool{}
			}
			seen[field][value] = true
		}
		for _, option := range q.Options {
			if _, ok := englishExamSSWText("ssw.option", option); !ok {
				t.Errorf("question %s has no exact catalog entry for option: %q", q.ID, option)
			}
			if seen["ssw.option"] == nil {
				seen["ssw.option"] = map[string]bool{}
			}
			seen["ssw.option"][option] = true
		}
		if q.ID != stableID || q.Answer != answer {
			t.Errorf("catalog lookup changed stable answer data for %s", stableID)
		}
		ids[q.ID] = true
	}
	wantCounts := map[string]int{
		"ssw.group": 1, "ssw.prompt": 9, "ssw.explanation": 9,
		"ssw.section": 2, "ssw.passage": 3, "ssw.option": 30,
	}
	for field, want := range wantCounts {
		if got := len(seen[field]); got != want {
			t.Errorf("enumerated %d unique forestry-silviculture %s sources, want %d", got, field, want)
		}
	}
	if len(ids) != 9 {
		t.Errorf("enumerated %d stable forestry-silviculture question IDs, want 9", len(ids))
	}
}

func TestEnglishSSWSharedLabelsAndCurriculumTitlesHaveExactMappings(t *testing.T) {
	file, err := os.Open("../../../docs/english-localization/shared_text.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) < 2 {
		t.Fatal("shared_text.csv has no source rows")
	}
	columns := make(map[string]int, len(rows[0]))
	for index, name := range rows[0] {
		columns[name] = index
	}
	get := func(row []string, name string) string {
		return row[columns[name]]
	}
	seen := map[string]map[string]bool{}
	for _, row := range rows[1:] {
		surface := get(row, "surface")
		field := get(row, "field")
		if surface != "SSW curriculum" && surface != "SSW simulation" {
			continue
		}
		switch field {
		case "Module title", "Lesson title", "Paper title":
		case "Section label · vocabulary", "Section label · grammar":
		case "Instructions":
		case "Group label":
		default:
			continue
		}
		source := get(row, "source_text")
		if source == "" {
			continue
		}
		key := surface + "/" + field
		if seen[key] == nil {
			seen[key] = map[string]bool{}
		}
		seen[key][source] = true
		translated := coverageReportSharedTextTranslation(coverageReportSharedText{
			Surface: surface, Sector: get(row, "sector"), Field: field, SourceText: source,
		})
		if translated.Missing {
			t.Errorf("missing exact English mapping for shared %s / %s source %q", surface, field, source)
		}
	}
	want := map[string]int{
		"SSW curriculum/Module title":               102,
		"SSW curriculum/Lesson title":               315,
		"SSW simulation/Paper title":                22,
		"SSW simulation/Section label · vocabulary": 8,
		"SSW simulation/Section label · grammar":    4,
	}
	for key, count := range want {
		if got := len(seen[key]); got != count {
			t.Errorf("shared source inventory for %s has %d unique strings, want %d", key, got, count)
		}
	}
}
