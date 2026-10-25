package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/text/language"
)

const (
	localeIndonesian = "id"
	localeEnglish    = "en"
)

// localizedLearningResponses selects the response language per request. The
// content lives in the source Indonesian datasets, so English is an immutable
// response transformation; it never changes a shared store or quiz attempt.
func localizedLearningResponses(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLocalizedLearningPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		locale := requestContentLanguage(r.Header.Get("Accept-Language"))
		if locale == localeEnglish && !englishLearningContentEnabled() {
			locale = localeIndonesian
		}
		addVaryHeader(w.Header(), "Accept-Language")
		w.Header().Set("Content-Language", locale)
		if !isOfflinePracticePath(r.URL.Path) {
			w.Header().Set("Cache-Control", "private, no-store")
		}

		buffered := &localizedResponseBuffer{ResponseWriter: w}
		next.ServeHTTP(buffered, r)
		body := buffered.body.Bytes()
		if locale == localeEnglish && (buffered.status != http.StatusOK || !isOfflinePracticePath(r.URL.Path)) {
			if localized, err := translateLearningJSONForPath(body, r.URL.Path); err == nil {
				body = localized
				w.Header().Del("Content-Length")
			}
		}
		status := buffered.status
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)
		if len(body) > 0 {
			_, _ = w.Write(body)
		}
	})
}

type localizedResponseBuffer struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *localizedResponseBuffer) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}

func (w *localizedResponseBuffer) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(body)
}

func (w *localizedResponseBuffer) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func requestContentLanguage(header string) string {
	if strings.TrimSpace(header) == "" {
		return localeIndonesian
	}
	tags, qualities, err := language.ParseAcceptLanguage(header)
	if err != nil {
		return localeIndonesian
	}
	for index, tag := range tags {
		if index < len(qualities) && qualities[index] <= 0 {
			continue
		}
		base, _ := tag.Base()
		switch strings.ToLower(base.String()) {
		case localeEnglish:
			return localeEnglish
		case localeIndonesian:
			return localeIndonesian
		}
	}
	// Existing clients do not send Accept-Language. Unsupported languages also
	// retain the API's established Indonesian behavior; the app sends `en` when
	// it applies its English fallback for an unsupported device language.
	return localeIndonesian
}

func isLocalizedLearningPath(path string) bool {
	for _, prefix := range []string{
		"/v1/home", "/v1/learning/", "/v1/exams", "/v1/dictionary/", "/v1/vocabulary/",
		"/v1/grammar/", "/v1/kanji/", "/v1/reading/", "/v1/simulation/",
		"/v1/exam-attempts", "/v1/diagnostic/", "/v1/ssw/", "/v1/offline/practice/",
	} {
		base := strings.TrimSuffix(prefix, "/")
		if path == base || strings.HasPrefix(path, base+"/") {
			return true
		}
	}
	return false
}

func isOfflinePracticePath(path string) bool {
	return strings.HasPrefix(path, "/v1/offline/practice/") && !strings.HasPrefix(path, "/v1/offline/practice-attempts/")
}

func addVaryHeader(header http.Header, value string) {
	for _, existing := range header.Values("Vary") {
		for _, item := range strings.Split(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(item), value) {
				return
			}
		}
	}
	header.Add("Vary", value)
}

func translateLearningJSON(body []byte) ([]byte, error) {
	return translateLearningJSONForPath(body, "")
}

func translateLearningJSONForPath(body []byte, path string) ([]byte, error) {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	domain := ""
	materialContextID := ""
	if strings.HasPrefix(path, "/v1/reading/") {
		domain = "reading"
	} else if strings.HasPrefix(path, "/v1/kanji/") {
		domain = "kanji"
		if strings.HasPrefix(path, "/v1/kanji/lessons/") {
			lessonID := strings.TrimPrefix(path, "/v1/kanji/lessons/")
			lessonID = strings.SplitN(lessonID, "/", 2)[0]
			parts := strings.Split(lessonID, "-")
			if len(parts) >= 3 && strings.HasPrefix(parts[1], "n") {
				materialContextID = "kanji:" + strings.ToUpper(parts[1]) + ":" + lessonID
			}
		}
	} else if strings.HasPrefix(path, "/v1/vocabulary/") {
		domain = "vocabulary"
	} else if strings.HasPrefix(path, "/v1/grammar/") {
		domain = "grammar"
	}
	translateLearningValueInContext(payload, "", nil, domain, materialContextID)
	return json.Marshal(payload)
}

func translateLearningValueInContext(value any, field string, questionContext map[string]any, domain, materialContextID string) {
	switch value := value.(type) {
	case map[string]any:
		if _, isQuestion := value["prompt"].(string); isQuestion {
			questionContext = make(map[string]any, len(value))
			for key, item := range value {
				questionContext[key] = item
			}
		}
		if field == "error" {
			code, _ := value["code"].(string)
			if message, ok := value["message"].(string); ok {
				value["message"] = translateLearningAPIError(code, message)
			}
		}
		for key, child := range value {
			if text, ok := child.(string); ok {
				value[key] = translateLearningTextInQuestionContext(key, text, questionContext, domain, materialContextID)
				continue
			}
			translateLearningValueInContext(child, key, questionContext, domain, materialContextID)
		}
	case []any:
		for index, child := range value {
			if text, ok := child.(string); ok {
				value[index] = translateLearningTextInQuestionContext(field, text, questionContext, domain, materialContextID)
				continue
			}
			translateLearningValueInContext(child, field, questionContext, domain, materialContextID)
		}
	}
}

func translateLearningTextInQuestionContext(field, text string, questionContext map[string]any, domain, materialContextID string) string {
	if domain == "reading" || domain == "kanji" {
		if domain == "kanji" && materialContextID != "" {
			key := jlptKanjiMaterialEnglishKey{contextID: materialContextID, field: field, source: text}
			if translated, ok := jlptKanjiN2Lessons19To44English[key]; ok {
				return translated
			}
			if translated, ok := jlptKanjiN2Lessons45To74English[key]; ok {
				return translated
			}
			if translated, ok := jlptKanjiN3EarlyMaterialEnglish[key]; ok {
				return translated
			}
			if translated, ok := jlptKanjiN3Lessons39To77English[key]; ok {
				return translated
			}
			if translated, ok := jlptKanjiN2Lessons09To18English[key]; ok {
				return translated
			}
			if translated, ok := jlptKanjiN2MnemonicNoteEnglish[key]; ok {
				return translated
			}
			if translated, ok := jlptKanjiN3QuizEnglishCatalog[key]; ok {
				return translated
			}
		}
		if domain == "kanji" && materialContextID != "" {
			key := jlptKanjiMaterialEnglishKey{contextID: materialContextID, field: field, source: text}
			if translated, ok := jlptKanjiMaterialEnglishCatalog[key]; ok {
				return translated
			}
		}
		candidates := []string{field, "reading." + field, "kanji." + field}
		if field == "text" {
			candidates = append(candidates, "choices", "reading.choices", "kanji.choices")
		}
		for _, candidate := range candidates {
			if translated, ok := jlptKanjiReadingEnglish[candidate][text]; ok {
				return translated
			}
		}
		switch field {
		case "title", "chapter", "description", "meaning", "prompt", "explanation", "choices", "text", "section", "group", "instructions", "imageDescription":
			return text
		}
	}
	if field == "text" && questionContext != nil {
		kind, _ := questionContext["kind"].(string)
		prompt, _ := questionContext["prompt"].(string)
		if kind == "Arti kata" || (kind == "" && isVocabularyMeaningQuestion(prompt)) {
			if translated, ok := englishJLPTVocabularyGrammar("vocabulary.meaning", text); ok {
				return translated
			}
			// Meaning choices belong to the vocabulary meaning catalog. Do not
			// borrow a different sense from another JLPT choice catalog.
			return text
		}
	}
	if domain == "vocabulary" || domain == "grammar" {
		candidates := []string{domain + "." + field}
		if domain == "vocabulary" {
			switch field {
			case "kind":
				candidates = append(candidates, "vocabulary.quiz.kind")
			case "prompt":
				candidates = append(candidates, "vocabulary.quiz.prompt")
			case "explanation":
				candidates = append(candidates, "vocabulary.quiz.explanation")
			case "imageDescription":
				candidates = append(candidates, "vocabulary.quiz.imageDescription")
			}
		}
		if domain == "grammar" && field == "meaning" {
			candidates = append(candidates, "grammar.exampleMeaning")
		}
		for _, candidate := range candidates {
			if translated, ok := englishJLPTVocabularyGrammar(candidate, text); ok {
				return translated
			}
		}
		switch field {
		case "title", "chapter", "description", "meaning", "exampleMeaning", "note", "pattern", "prompt", "explanation", "text", "kind", "imageDescription":
			return text
		}
	}
	return translateLearningText(field, text)
}

var learningAPIErrorMessages = map[string]string{
	"dictionary_unavailable":   "The dictionary is temporarily unavailable. Please try again shortly.",
	"internal_error":           "An unexpected server error occurred.",
	"invalid_activity":         "The learning position or access time is invalid.",
	"invalid_attempt":          "The practice attempt is invalid.",
	"invalid_input":            "Some of the submitted information is invalid. Please check it and try again.",
	"invalid_json":             "The request data is invalid. Please check it and try again.",
	"invalid_level":            "Choose a valid course level and try again.",
	"invalid_practice_package": "The offline practice package is invalid or does not match this account.",
	"invalid_query":            "Check your search terms and filters, then try again.",
	"not_found":                "The requested item could not be found.",
	"premium_required":         "Premium access is required for this item.",
	"session_error":            "Your session could not be created. Please try again.",
	"unauthorized":             "Your session is invalid or has expired. Please sign in again.",
}

func translateLearningAPIError(code, _ string) string {
	if translated, ok := learningAPIErrorMessages[code]; ok {
		return translated
	}
	// Error codes remain machine-readable. Use a safe English fallback for
	// unfamiliar messages instead of returning Indonesian under an English
	// Content-Language header.
	return "The request could not be completed. Please try again."
}

// translateLearningText only translates fields that represent learner-facing
// explanations or labels. Japanese source fields such as `japanese`, `term`,
// `reading`, `pattern`, `character`, and all identifiers are deliberately left
// untouched. Unknown content remains source text until a reviewed translation
// is added for that field and value.
func translateLearningText(field, text string) string {
	if translated, ok := translateReviewedLearningText(field, text); ok {
		return translated
	}
	if translated, ok := learningTranslations[field][text]; ok {
		return translated
	}
	return text
}

func isVocabularyMeaningQuestion(prompt string) bool {
	_, exists := englishJLPTVocabularyGrammar("vocabulary.quiz.prompt", prompt)
	return exists
}

// translateReviewedLearningText looks up exact human-authored entries across
// the learning-content catalogs. API responses expose only the leaf JSON key,
// so try the known domain-qualified variants and accept a result only when all
// matching catalogs agree on the translation.
func translateReviewedLearningText(field, source string) (string, bool) {
	if field == "title" {
		if translated, ok := englishJLPTVocabularyGrammar("grammar.title", source); ok {
			return translated, true
		}
	}
	if field == "chapter" {
		if translated, ok := englishJLPTVocabularyGrammar("grammar.chapter", source); ok {
			return translated, true
		}
	}
	fields := []string{field}
	for _, domain := range []string{
		"vocabulary", "grammar", "kanji", "reading", "diagnostic", "simulation", "ssw", "ssw.paper",
	} {
		fields = append(fields, domain+"."+field)
	}
	switch field {
	case "text", "option":
		fields = append(fields,
			"vocabulary.choice", "grammar.choice", "kanji.choices", "reading.choices",
			"diagnostic.choice", "simulation.choice", "ssw.choice", "choices",
		)
	case "prompt":
		fields = append(fields, "vocabulary.quiz.prompt")
	case "explanation":
		fields = append(fields, "vocabulary.quiz.explanation")
	case "kind":
		fields = append(fields, "vocabulary.quiz.kind")
	case "imageDescription":
		fields = append(fields, "vocabulary.quiz.imageDescription")
	case "options":
		// SSW practice returns options as []string, while the other quiz APIs
		// return QuizOption objects whose learner-facing leaf is `text`.
		fields = append(fields, "ssw.option")
	case "instructions":
		fields = append(fields, "simulation.instruction", "diagnostic.instruction", "ssw.instruction")
	case "name":
		fields = append(fields, "diagnostic.section", "simulation.section", "ssw.section")
	case "section":
		fields = append(fields, "diagnostic.section", "simulation.section", "ssw.section")
	case "title":
		fields = append(fields, "ssw.paper.title", "grammar.title")
	case "chapter":
		fields = append(fields, "grammar.chapter")
	}
	seenFields := make(map[string]struct{}, len(fields))
	translations := make(map[string]struct{})
	for _, candidate := range fields {
		if _, seen := seenFields[candidate]; seen {
			continue
		}
		seenFields[candidate] = struct{}{}
		if translated, ok := englishJLPTVocabularyGrammar(candidate, source); ok {
			translations[translated] = struct{}{}
		}
		if bySource, ok := jlptKanjiReadingEnglish[candidate]; ok {
			if translated, found := bySource[source]; found {
				translations[translated] = struct{}{}
			}
		}
		if translated, ok := englishExamSSWText(candidate, source); ok {
			translations[translated] = struct{}{}
		}
		if translated, ok := sswCurriculumAccommodationEnglish[examSSWEnglishKey{field: candidate, source: source}]; ok {
			translations[translated] = struct{}{}
		}
	}
	if len(translations) != 1 {
		return "", false
	}
	for translated := range translations {
		return translated, true
	}
	return "", false
}

var learningTranslations = map[string]map[string]string{
	"chapter": {
		"Kalimat dasar": "Basic sentences",
		"Kata tunjuk":   "Demonstratives",
		"Keberadaan":    "Existence and location",
		"Partikel":      "Particles",
	},
	"title": {
		"Perkenalan diri":               "Self introductions",
		"Keluarga":                      "Family",
		"Benda di sekolah":              "School items",
		"Rumah dan ruangan":             "Homes and rooms",
		"Makanan sehari-hari":           "Everyday food",
		"Minuman dan buah":              "Drinks and fruit",
		"Aktivitas sehari-hari":         "Everyday activities",
		"Belajar dan bekerja":           "Study and work",
		"Tempat di kota":                "Places in town",
		"Transportasi":                  "Transportation",
		"Waktu sehari-hari":             "Everyday time expressions",
		"Hari dan libur":                "Days and holidays",
		"Cuaca dan alam":                "Weather and nature",
		"Hewan":                         "Animals",
		"Tubuh":                         "The body",
		"Pakaian":                       "Clothing",
		"Warna":                         "Colors",
		"Ukuran dan sifat":              "Sizes and qualities",
		"Posisi dan arah":               "Positions and directions",
		"Hobi dan waktu luang":          "Hobbies and free time",
		"Identitas dan penyangkalan":    "Identity and negation",
		"Waktu lampau untuk kata benda": "Past tense for nouns",
		"Topik dan pertanyaan":          "Topics and questions",
		"Kepemilikan dan kesamaan":      "Possession and similarity",
		"Menunjuk benda":                "Pointing to things",
		"Menerangkan benda":             "Describing things",
		"Tempat dan arah sopan":         "Places and polite directions",
	},
	"kind": {
		"Arti kata":         "Word meaning",
		"Cara baca":         "Reading",
		"Penulisan kata":    "Word spelling",
		"Pemakaian kalimat": "Sentence usage",
	},
	"description": {
		"Belajar perkenalan diri melalui gambar, dua contoh per kata, dan latihan arti, cara baca serta kalimat.": "Learn to introduce yourself through illustrations, two examples per word, and practice with meanings, readings, and sentences.",
		"Kosakata tematik dengan spaced repetition":                                                               "Thematic vocabulary with spaced repetition",
		"Grammar JLPT N5 sampai N1: pola, contoh kalimat, dan kuis":                                               "JLPT N5–N1 grammar: patterns, example sentences, and quizzes",
		"Kanji JLPT N5 sampai N1: bacaan, contoh pemakaian, dan kuis":                                             "JLPT N5–N1 kanji: readings, usage examples, and quizzes",
		"Reading JLPT N5 sampai N1: cerita, kosakata, dan kuis":                                                   "JLPT N5–N1 reading: stories, vocabulary, and quizzes",
		"20 tema bergambar, 160 entri kosakata, dua contoh per kata, dan tiga jenis latihan.":                     "20 illustrated themes, 160 vocabulary entries, two examples per word, and three practice types.",
		"34 tema dan 630 kosakata N4 dengan cara baca, arti Indonesia, dan contoh pemakaian.":                     "34 themes and 630 N4 words with readings, Indonesian meanings, and usage examples.",
		"5 tema, 40 entri kosakata N2 dengan bacaan, arti Indonesia, dan contoh kalimat.":                         "5 themes with 40 N2 entries, readings, Indonesian meanings, and example sentences.",
		"6 tema, 72 kosakata tingkat menengah dengan arti dan cara baca.":                                         "6 themes with 72 intermediate words, meanings, and readings.",
		"8 tema dan 64 kosakata tingkat lanjut dengan cara baca, arti Indonesia, serta contoh.":                   "8 themes with 64 advanced words, readings, Indonesian meanings, and examples.",
	},
	"meaning": {
		"saya":                                  "I",
		"pelajar":                               "student",
		"guru":                                  "teacher",
		"teman":                                 "friend",
		"anak":                                  "child",
		"nama":                                  "name",
		"Jepang":                                "Japan",
		"luar negeri":                           "abroad",
		"adalah …":                              "is …",
		"bukan …":                               "is not …",
		"dulu adalah …":                         "was …",
		"yang mana":                             "which one",
		"di sini / di situ / di sana / di mana": "here / there / over there / where",
		"Saya adalah pelajar.":                  "I am a student.",
		"Tanaka adalah pelajar.":                "Tanaka is a student.",
		"Pelajar belajar di perpustakaan.":      "Students study in the library.",
		"Guru ada di kelas.":                    "The teacher is in the classroom.",
		"Saya menyebutkan nama kepada guru.":    "I say my name to the teacher.",
		"Saya pergi ke sekolah bersama teman.":  "I go to school with a friend.",
		"Teman saya ada di taman.":              "My friend is in the park.",
		"Ada dua anak.":                         "There are two children.",
		"Anak makan roti.":                      "The child eats bread.",
		"Tolong tulis nama Anda.":               "Please write your name.",
		"Nama gurunya Tanaka.":                  "The teacher's name is Tanaka.",
		"Saya ingin pergi ke Jepang.":           "I want to go to Japan.",
		"Saya suka makanan Jepang.":             "I like Japanese food.",
		"Saya bepergian ke luar negeri.":        "I travel abroad.",
		"Saya punya teman dari luar negeri.":    "I have a friend from abroad.",
		"Ini adalah buku.":                      "This is a book.",
		"Saya bukan guru.":                      "I am not a teacher.",
		"Ini bukan kamus.":                      "This is not a dictionary.",
		"Kemarin adalah hari libur.":            "Yesterday was a day off.",
		"Ayah dulu seorang guru.":               "My father used to be a teacher.",
		"Saya juga pelajar.":                    "I am a student too.",
		"Saya orang Indonesia.":                 "I am Indonesian.",
		"Hari ini panas.":                       "It is hot today.",
		"Apakah Tanaka pelajar?":                "Is Tanaka a student?",
		"Apakah ini buku Anda?":                 "Is this your book?",
	},
	"formation": {
		"Kata benda + です":            "Noun + です",
		"Kata benda + ではありません":       "Noun + ではありません",
		"Kata benda + でした":           "Noun + でした",
		"Kata benda + は + informasi": "Topic + は + information",
	},
	"explanation": {
		"です menutup pernyataan sopan tentang identitas atau sifat. Subjek yang jelas dari konteks boleh dihilangkan.": "です completes a polite statement about identity or a quality. A subject that is clear from context may be omitted.",
		"Bentuk negatif sopan dari Nです. じゃありません umum dalam percakapan; ではありません lebih formal.":                         "The polite negative form of Nです. じゃありません is common in conversation; ではありません is more formal.",
		"Gunakan でした untuk identitas atau keadaan kata benda pada waktu lampau.":                                      "Use でした for a noun's identity or state in the past.",
		"は dibaca わ saat menjadi partikel. Topik adalah hal yang sedang dibicarakan, tidak selalu pelaku tindakan.":   "は is read as わ when used as a particle. The topic is what the sentence is about; it is not always the person doing the action.",
	},
	"commonMistake": {
		"Jangan menambahkan です setelah kata kerja bentuk ます.": "Do not add です after a verb in ます form.",
		"Negatif kata benda bukan Nないです.":                     "Do not form a noun negative with Nないです.",
		"Jangan membaca partikel は sebagai ha.":               "Do not read the particle は as ha.",
	},
	"exampleMeaning": {
		"Saya adalah pelajar.":                 "I am a student.",
		"Saya belajar bahasa Jepang.":          "I study Japanese.",
		"Tanaka adalah pelajar.":               "Tanaka is a student.",
		"Pelajar belajar di perpustakaan.":     "Students study in the library.",
		"Guru ada di kelas.":                   "The teacher is in the classroom.",
		"Saya menyebutkan nama kepada guru.":   "I say my name to the teacher.",
		"Saya pergi ke sekolah bersama teman.": "I go to school with a friend.",
		"Teman saya ada di taman.":             "My friend is in the park.",
		"Ada dua anak.":                        "There are two children.",
		"Anak makan roti.":                     "The child eats bread.",
		"Tolong tulis nama Anda.":              "Please write your name.",
		"Nama gurunya Tanaka.":                 "The teacher's name is Tanaka.",
		"Saya ingin pergi ke Jepang.":          "I want to go to Japan.",
		"Saya suka makanan Jepang.":            "I like Japanese food.",
		"Saya bepergian ke luar negeri.":       "I travel abroad.",
		"Saya punya teman dari luar negeri.":   "I have a friend from abroad.",
	},
	"sourceLicense": {
		"CC BY-SA 4.0 • diadaptasi dan diterjemahkan": "CC BY-SA 4.0 • adapted and translated",
	},
	"note": {
		"Perhatikan partikel dan pasangan kata pada dua contoh. Gunakan cara baca yang tertera ketika mengucapkan kata.": "Notice the particles and word pairings in the two examples. Use the listed reading when pronouncing the word.",
		"Gunakan わたし sebagai kata ganti netral dan sopan.":                                                               "Use わたし as a neutral, polite pronoun.",
		"学生 dipakai untuk siswa atau mahasiswa.":                                                                         "学生 is used for school and university students.",
	},
	"prompt": {
		"私は学生（　）。 Lengkapi kalimat sopan.":                    "私は学生（　）。 Complete the polite sentence.",
		"Pilih kalimat sopan yang berarti ‘Saya bukan guru’.": "Choose the polite sentence that means ‘I am not a teacher.’",
		"これは本です。 Apa fungsi です?":                              "これは本です。 What is the function of です?",
	},
	"text": {
		"saya":    "I",
		"pelajar": "student",
		"guru":    "teacher",
		"teman":   "friend",
	},
	"section": {
		"Kosakata":    "Vocabulary",
		"Tata bahasa": "Grammar",
		"Kanji":       "Kanji",
		"Membaca":     "Reading",
	},
	"message": {
		"Grammar tersedia untuk N5, N4, N3, N2, dan N1.": "Grammar is available for N5, N4, N3, N2, and N1.",
		"Kanji tersedia untuk N5, N4, N3, N2, dan N1.":   "Kanji is available for N5, N4, N3, N2, and N1.",
		"Reading tersedia untuk N5, N4, N3, N2, dan N1.": "Reading is available for N5, N4, N3, N2, and N1.",
		"Gunakan kata maksimal 100 karakter, halaman 1–100, mode yang valid, atau satu karakter kanji. Jika ejaan romaji belum dikenali pada mode Jepang, coba gunakan kana.": "Enter a word of up to 100 characters, a page from 1 to 100, a valid search mode, or one kanji character. If romaji is not recognized in Japanese mode, try kana.",
		"Detail kanji belum tersedia dari sumber kamus.":                          "Kanji details are not available from the dictionary source yet.",
		"Sumber kamus sedang tidak tersedia. Coba lagi sebentar atau buka Jisho.": "The dictionary source is temporarily unavailable. Please try again shortly or open Jisho.",
	},
}

func localizePracticePackage(raw []byte) ([]byte, error) {
	return translateLearningJSON(raw)
}
