package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	vocabularyData "nihongo-master-backend/data/vocabulary"
	"nihongo-master-backend/internal/domain"
)

type vocabularyAttemptState struct {
	result       *domain.VocabularyAttemptResult
	attempt      domain.VocabularyAttempt
	userID       string
	correct      map[string]string
	explanations map[string]string
	answers      map[string]domain.VocabularyAnswerFeedback
}

func (s *MemoryStore) VocabularyLevels(userID string) []domain.VocabularyLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.VocabularyLevel, 0, len(s.vocabulary))
	for _, level := range s.vocabulary {
		if PremiumLevel(level.Level) {
			level.RequiredTier = domain.TierPremium
		}
		lessons := make([]domain.VocabularyLessonSummary, 0, len(level.Lessons))
		total := 0
		for _, lesson := range level.Lessons {
			asset, questionCount := "", len(lesson.Words)
			if len(lesson.Words) > 0 {
				asset = lesson.Words[0].ImageAsset
			}
			if level.Level == "N5" {
				questionCount *= 3
			}
			progress := s.progress[userID][lesson.ID]
			total += progress.Completion
			lessons = append(lessons, domain.VocabularyLessonSummary{
				ImageAsset: asset, ImageURL: s.assetURLLocked(asset), QuestionCount: questionCount,
				ID: lesson.ID, Order: lesson.Order, Title: lesson.Title,
				Description: lesson.Description, WordCount: len(lesson.Words),
				BestScore: progress.Completion, Completed: progress.Completed,
			})
		}
		percentage := 0
		if len(lessons) > 0 {
			percentage = total / len(lessons)
		}
		result = append(result, domain.VocabularyLevel{
			Level: level.Level, Title: level.Title, Description: level.Description,
			SourceName: level.SourceName, SourceURL: level.SourceURL, SourceLicense: level.SourceLicense,
			RequiredTier: level.RequiredTier, Progress: percentage, Lessons: lessons,
		})
	}
	return result
}

func (s *MemoryStore) VocabularyLesson(userID, lessonID string) (domain.VocabularyLesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	level, lesson := s.findVocabularyLessonLocked(lessonID)
	if lesson == nil {
		return domain.VocabularyLesson{}, ErrNotFound
	}
	progress := s.progress[userID][lessonID]
	words := append([]domain.VocabularyWord(nil), lesson.Words...)
	for i := range words {
		if words[i].ImageAsset != "" {
			words[i].ImageURL = s.assetURLLocked(words[i].ImageAsset)
		}
	}
	return domain.VocabularyLesson{
		ID: lesson.ID, Level: level.Level, Title: lesson.Title,
		Description: lesson.Description, BestScore: progress.Completion,
		Completed: progress.Completed,
		ImageURL:  s.assetURLLocked("assets/images/vocabulary/study-journey.webp"),
		Words:     words,
	}, nil
}

func (s *MemoryStore) StartVocabularyAttempt(userID, lessonID string) (domain.VocabularyAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	level, lesson := s.findVocabularyLessonLocked(lessonID)
	if lesson == nil {
		return domain.VocabularyAttempt{}, ErrNotFound
	}
	options := make([]domain.QuizOption, 0, len(lesson.Words))
	for _, word := range lesson.Words {
		options = append(options, domain.QuizOption{ID: word.ID, Text: word.Meaning})
	}
	questions := make([]domain.VocabularyQuestion, 0, len(lesson.Words))
	correct := make(map[string]string, len(lesson.Words))
	explanations := make(map[string]string, len(lesson.Words))
	for index, word := range lesson.Words {
		questionID := lesson.ID + "-q" + fmt.Sprint(index+1)
		rotated := append([]domain.QuizOption(nil), options...)
		shiftOptions(rotated, index)
		questions = append(questions, domain.VocabularyQuestion{
			ID: questionID, Prompt: fmt.Sprintf("Apa arti %s (%s)?", word.Term, word.Reading), Options: rotated,
		})
		correct[questionID] = word.ID
		explanations[questionID] = fmt.Sprintf("%s berarti %s. %s Contoh: %s — %s", word.Term, word.Meaning, word.Note, word.Example, word.ExampleMeaning)
	}
	if level.Level == "N5" {
		questions, correct, explanations = s.n5VocabularyQuestionsLocked(lesson)
	}
	id := fmt.Sprintf("vocab-attempt-%d", s.attemptID.Add(1))
	attempt := domain.VocabularyAttempt{
		ID: id, LessonID: lessonID, Questions: questions,
		ImageURL:  s.assetURLLocked("assets/images/vocabulary/quiz-mastery.webp"),
		StartedAt: time.Now().UTC(),
	}
	s.vocabAttempts[id] = &vocabularyAttemptState{
		attempt: attempt, userID: userID, correct: correct,
		explanations: explanations, answers: map[string]domain.VocabularyAnswerFeedback{},
	}
	return attempt, nil
}

func (s *MemoryStore) AssetURL(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.assetURLLocked(key)
}

func (s *MemoryStore) assetURLLocked(key string) string {
	if key == "" {
		return ""
	}
	key = strings.TrimLeft(key, "/")
	if strings.HasPrefix(key, "assets/images/") && strings.HasSuffix(key, ".png") {
		key = strings.TrimSuffix(key, ".png") + ".webp"
	}
	url := s.assetBaseURL + "/" + key
	if strings.HasPrefix(key, "assets/images/") {
		// Bump assetImageVersion whenever images are re-encoded under the same
		// file name, so app caches keyed by URL fetch the new file.
		url += "?v=" + assetImageVersion
	}
	return url
}

// assetImageVersion 2 = full-resolution images restored after the 2026-10-22
// compression (commit 177f67e) made them blurry.
const assetImageVersion = "2"

func (s *MemoryStore) AnswerVocabularyQuestion(userID, attemptID, questionID, optionID string) (domain.VocabularyAnswerFeedback, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.vocabAttempts[attemptID]
	if !ok || state.userID != userID {
		return domain.VocabularyAnswerFeedback{}, ErrNotFound
	}
	correctOption, ok := state.correct[questionID]
	if !ok || strings.TrimSpace(optionID) == "" {
		return domain.VocabularyAnswerFeedback{}, ErrInvalid
	}
	validOption := false
	for _, question := range state.attempt.Questions {
		if question.ID != questionID {
			continue
		}
		for _, option := range question.Options {
			if option.ID == optionID {
				validOption = true
				break
			}
		}
	}
	if !validOption {
		return domain.VocabularyAnswerFeedback{}, ErrInvalid
	}
	if existing, answered := state.answers[questionID]; answered {
		return existing, nil
	}
	feedback := domain.VocabularyAnswerFeedback{
		QuestionID: questionID, Correct: optionID == correctOption,
		CorrectOptionID: correctOption, Explanation: state.explanations[questionID],
	}
	state.answers[questionID] = feedback
	return feedback, nil
}

func (s *MemoryStore) CompleteVocabularyAttempt(userID, attemptID string) (domain.VocabularyAttemptResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.vocabAttempts[attemptID]
	if !ok || state.userID != userID {
		return domain.VocabularyAttemptResult{}, ErrNotFound
	}
	if state.result != nil {
		return *state.result, nil
	}
	if len(state.answers) != len(state.attempt.Questions) {
		return domain.VocabularyAttemptResult{}, ErrInvalid
	}
	correct := 0
	for _, feedback := range state.answers {
		if feedback.Correct {
			correct++
		}
	}
	score := correct * 100 / len(state.attempt.Questions)
	passed := score >= 80
	if s.progress[userID] == nil {
		s.progress[userID] = map[string]domain.LessonProgress{}
	}
	previous := s.progress[userID][state.attempt.LessonID]
	best := previous.Completion
	if score > best {
		best = score
	}
	progress := domain.LessonProgress{
		LessonID: state.attempt.LessonID, Completion: best,
		LastPosition: len(state.attempt.Questions) * 60,
		Completed:    previous.Completed || passed, UpdatedAt: time.Now().UTC(),
	}
	now := s.gamificationNow().UTC()
	progress.UpdatedAt = now
	if err := s.recordStudyLocked(userID, studyEvent{ID: "vocabulary:" + attemptID, Kind: "lesson", Subject: "vocabulary", LessonID: state.attempt.LessonID, Score: score, Answered: len(state.answers), Seconds: int(now.Sub(state.attempt.StartedAt).Seconds()), At: now, Progress: &progress}); err != nil {
		return domain.VocabularyAttemptResult{}, err
	}
	result := domain.VocabularyAttemptResult{
		AttemptID: attemptID, LessonID: state.attempt.LessonID, Score: score,
		Correct: correct, Total: len(state.attempt.Questions), Passed: passed,
		BestScore: best, Progress: s.vocabularyProgressLocked(userID), CanRetry: true,
	}
	state.result = &result
	return result, nil
}

func (s *MemoryStore) vocabularyProgressLocked(userID string) int {
	total, lessons := 0, 0
	for _, level := range s.vocabulary {
		for _, lesson := range level.Lessons {
			total += s.progress[userID][lesson.ID].Completion
			lessons++
		}
	}
	if lessons == 0 {
		return 0
	}
	return total / lessons
}

func (s *MemoryStore) findVocabularyLessonLocked(lessonID string) (*domain.VocabularyLevelContent, *domain.VocabularyLessonContent) {
	for levelIndex := range s.vocabulary {
		for lessonIndex := range s.vocabulary[levelIndex].Lessons {
			lesson := &s.vocabulary[levelIndex].Lessons[lessonIndex]
			if lesson.ID == lessonID {
				return &s.vocabulary[levelIndex], lesson
			}
		}
	}
	return nil, nil
}

func shiftOptions(options []domain.QuizOption, shift int) {
	if len(options) == 0 {
		return
	}
	shift %= len(options)
	rotated := append(options[shift:], options[:shift]...)
	copy(options, rotated)
	sort.SliceStable(options, func(i, j int) bool { return (i+shift)%2 < (j+shift)%2 })
}

func seedVocabulary() []domain.VocabularyLevelContent {
	levels := []domain.VocabularyLevelContent{
		vocabularyLevel("N5", "Fondasi sehari-hari", "Kosakata paling dasar untuk memperkenalkan diri, makan, dan kebutuhan harian.", "n5-daily", "Kata dasar sehari-hari", "Mulai dari kata yang paling sering ditemui.", []domain.VocabularyWord{
			word("n5-watashi", "私", "わたし", "saya", "私は学生です。", "Saya adalah pelajar.", "Gunakan わたし sebagai kata ganti netral dan sopan."),
			word("n5-gakusei", "学生", "がくせい", "pelajar", "田中さんは学生です。", "Tanaka adalah pelajar.", "学生 dipakai untuk siswa atau mahasiswa."),
			word("n5-taberu", "食べる", "たべる", "makan", "パンを食べます。", "Saya makan roti.", "Bentuk sopannya adalah 食べます."),
			word("n5-mizu", "水", "みず", "air", "水をください。", "Tolong beri saya air.", "ください berarti 'tolong berikan'."),
		}),
		vocabularyLevel("N4", "Aktivitas dan rencana", "Kosakata untuk menjelaskan rencana, kebutuhan, dan pengalaman sederhana.", "n4-plans", "Rencana dan pengalaman", "Bangun percakapan yang lebih panjang.", []domain.VocabularyWord{
			word("n4-yotei", "予定", "よてい", "rencana atau jadwal", "週末の予定は何ですか。", "Apa rencana akhir pekanmu?", "予定 dapat berarti rencana maupun jadwal."),
			word("n4-hitsuyou", "必要", "ひつよう", "perlu", "パスポートが必要です。", "Paspor diperlukan.", "Sering digunakan dengan pola ～が必要です."),
			word("n4-tsuzukeru", "続ける", "つづける", "melanjutkan", "勉強を続けます。", "Saya melanjutkan belajar.", "Ini verba transitif: seseorang melanjutkan sesuatu."),
			word("n4-keiken", "経験", "けいけん", "pengalaman", "日本で働いた経験があります。", "Saya punya pengalaman bekerja di Jepang.", "Pola 経験があります menyatakan pernah mengalami."),
		}),
		vocabularyLevel("N3", "Situasi dan dampak", "Kosakata menengah untuk memahami berita ringan dan lingkungan kerja.", "n3-work", "Komunikasi di tempat kerja", "Pahami masalah, tindakan, dan pengaruhnya.", []domain.VocabularyWord{
			word("n3-kaizen", "改善", "かいぜん", "perbaikan", "サービスを改善します。", "Kami memperbaiki layanan.", "改善 menekankan perubahan menjadi lebih baik."),
			word("n3-joukyou", "状況", "じょうきょう", "situasi", "現在の状況を説明します。", "Saya menjelaskan situasi saat ini.", "Lebih formal daripada 様子 untuk keadaan."),
			word("n3-taiou", "対応", "たいおう", "penanganan atau respons", "問題に対応します。", "Kami menangani masalah.", "Gunakan ～に対応する untuk merespons sesuatu."),
			word("n3-eikyou", "影響", "えいきょう", "pengaruh", "天気は試合に影響します。", "Cuaca memengaruhi pertandingan.", "Pola ～に影響する berarti memengaruhi."),
		}),
		vocabularyLevel("N2", "Bahasa profesional", "Kosakata formal untuk laporan, kebijakan, dan pencapaian kerja.", "n2-business", "Laporan dan kebijakan", "Kuasai kata yang sering muncul pada dokumen formal.", []domain.VocabularyWord{
			word("n2-houshin", "方針", "ほうしん", "kebijakan atau arah", "会社の方針を確認します。", "Saya memastikan kebijakan perusahaan.", "方針 adalah arah dasar dalam mengambil tindakan."),
			word("n2-jisshi", "実施", "じっし", "pelaksanaan", "調査を実施しました。", "Kami telah melaksanakan survei.", "Kata formal, sering dipasangkan dengan する."),
			word("n2-kadai", "課題", "かだい", "tugas atau tantangan", "今後の課題を話し合います。", "Kami membahas tantangan ke depan.", "Dalam bisnis sering berarti masalah yang harus diselesaikan."),
			word("n2-tassei", "達成", "たっせい", "pencapaian", "目標を達成しました。", "Kami telah mencapai target.", "Pola 目標を達成する sangat umum."),
		}),
		vocabularyLevel("N1", "Nuansa akademik", "Kosakata tingkat lanjut untuk argumen, analisis, dan bacaan abstrak.", "n1-analysis", "Analisis dan argumentasi", "Bedakan nuansa kata dalam tulisan kompleks.", []domain.VocabularyWord{
			word("n1-gainen", "概念", "がいねん", "konsep", "基本的な概念を説明します。", "Saya menjelaskan konsep dasar.", "概念 merujuk pada gagasan abstrak yang digeneralisasi."),
			word("n1-datou", "妥当", "だとう", "tepat atau valid", "その判断は妥当です。", "Keputusan itu tepat.", "Menilai bahwa alasan atau keputusan masuk akal."),
			word("n1-fumaeru", "踏まえる", "ふまえる", "mempertimbangkan berdasarkan", "結果を踏まえて計画を直します。", "Kami memperbaiki rencana berdasarkan hasil.", "Pola ～を踏まえて sering muncul dalam tulisan formal."),
			word("n1-ichijirushii", "著しい", "いちじるしい", "mencolok atau signifikan", "技術は著しく進歩しました。", "Teknologi berkembang secara signifikan.", "Digunakan untuk perubahan besar yang jelas terlihat."),
		}),
	}
	var n5 []domain.VocabularyLessonContent
	if err := json.Unmarshal(vocabularyData.N5, &n5); err != nil {
		panic("invalid vocabulary N5: " + err.Error())
	}
	if len(n5) != 20 {
		panic("incomplete vocabulary N5")
	}
	for _, lesson := range n5 {
		if len(lesson.Words) != 8 {
			panic("invalid vocabulary lesson: " + lesson.ID)
		}
		for i, w := range lesson.Words {
			if len(w.Examples) != 2 || w.ImagePanel != i || w.Term == "" || w.Reading == "" || w.Meaning == "" || w.ImageAsset == "" {
				panic("invalid vocabulary word: " + w.ID)
			}
			for _, e := range w.Examples {
				if e.Japanese == "" || e.Reading == "" || e.Meaning == "" {
					panic("invalid vocabulary example: " + w.ID)
				}
			}
		}
	}
	levels[0].Lessons = n5
	levels[0].Description = "20 tema bergambar, 160 entri kosakata, dua contoh per kata, dan tiga jenis latihan."
	var n4 []domain.VocabularyLessonContent
	if err := json.Unmarshal(vocabularyData.N4, &n4); err != nil {
		panic("invalid vocabulary N4: " + err.Error())
	}
	if len(n4) != 34 {
		panic("incomplete vocabulary N4")
	}
	n4WordCount := 0
	for _, lesson := range n4 {
		if lesson.ID == "" || len(lesson.Words) == 0 {
			panic("invalid vocabulary N4 lesson: " + lesson.ID)
		}
		n4WordCount += len(lesson.Words)
		for _, w := range lesson.Words {
			if w.ID == "" || w.Term == "" || w.Reading == "" || w.Meaning == "" || w.Example == "" || w.ExampleMeaning == "" || len(w.Examples) == 0 {
				panic("invalid vocabulary N4 word: " + w.ID)
			}
			for _, example := range w.Examples {
				if example.Japanese == "" || example.Reading == "" || example.Meaning == "" {
					panic("invalid vocabulary N4 example: " + w.ID)
				}
			}
		}
	}
	if n4WordCount != 630 {
		panic("incomplete vocabulary N4 words")
	}
	levels[1].Lessons = n4
	levels[1].Description = "34 tema dan 630 kosakata N4 dengan cara baca, arti Indonesia, dan contoh pemakaian."
	levels[1].SourceName = "OpenJLPT"
	levels[1].SourceURL = "https://github.com/evanclan/OpenJLPT/blob/main/data/json/vocab/n4.json"
	levels[1].SourceLicense = "CC BY-SA 4.0 • diadaptasi dan diterjemahkan"
	var n2 []domain.VocabularyLessonContent
	if err := json.Unmarshal(vocabularyData.N2, &n2); err != nil {
		panic("invalid vocabulary N2: " + err.Error())
	}
	if len(n2) != 5 {
		panic("incomplete vocabulary N2")
	}
	for _, lesson := range n2 {
		if len(lesson.Words) != 8 {
			panic("invalid vocabulary lesson: " + lesson.ID)
		}
	}
	levels[3].Lessons = n2
	levels[3].Description = "5 tema, 40 entri kosakata N2 dengan bacaan, arti Indonesia, dan contoh kalimat."
	var n3 []domain.VocabularyLessonContent
	if err := json.Unmarshal(vocabularyData.N3, &n3); err != nil {
		panic("invalid vocabulary N3: " + err.Error())
	}
	if len(n3) != 6 {
		panic("incomplete vocabulary N3")
	}
	for _, lesson := range n3 {
		if len(lesson.Words) != 12 {
			panic("invalid vocabulary N3 lesson: " + lesson.ID)
		}
		for _, w := range lesson.Words {
			if w.ID == "" || w.Term == "" || w.Reading == "" || w.Meaning == "" {
				panic("invalid vocabulary N3 word: " + w.ID)
			}
		}
	}
	levels[2].Lessons = n3
	levels[2].Description = "6 tema, 72 kosakata tingkat menengah dengan arti dan cara baca."
	var n1 []domain.VocabularyLessonContent
	if err := json.Unmarshal(vocabularyData.N1, &n1); err != nil {
		panic("invalid vocabulary N1: " + err.Error())
	}
	if len(n1) != 8 {
		panic("incomplete vocabulary N1")
	}
	n1WordCount := 0
	for _, lesson := range n1 {
		if lesson.ID == "" || len(lesson.Words) != 8 {
			panic("invalid vocabulary N1 lesson: " + lesson.ID)
		}
		n1WordCount += len(lesson.Words)
		for _, w := range lesson.Words {
			if w.ID == "" || w.Term == "" || w.Reading == "" || w.Meaning == "" || w.Example == "" || w.ExampleMeaning == "" {
				panic("invalid vocabulary N1 word: " + w.ID)
			}
		}
	}
	if n1WordCount != 64 {
		panic("incomplete vocabulary N1 words")
	}
	levels[4].Lessons = n1
	levels[4].Description = "8 tema dan 64 kosakata tingkat lanjut dengan cara baca, arti Indonesia, serta contoh."
	return levels
}

func (s *MemoryStore) n5VocabularyQuestionsLocked(lesson *domain.VocabularyLessonContent) ([]domain.VocabularyQuestion, map[string]string, map[string]string) {
	questions := make([]domain.VocabularyQuestion, 0, len(lesson.Words)*3)
	correct, explanations := map[string]string{}, map[string]string{}
	for kind := 0; kind < 3; kind++ {
		for i, w := range lesson.Words {
			id := fmt.Sprintf("%s-q%d", lesson.ID, len(questions)+1)
			texts := []string{}
			for _, other := range lesson.Words {
				switch kind {
				case 0:
					texts = append(texts, other.Meaning)
				case 1:
					if w.Term == w.Reading {
						texts = append(texts, other.Term)
					} else {
						texts = append(texts, other.Reading)
					}
				case 2:
					texts = append(texts, other.Example)
				}
			}
			options := []domain.QuizOption{{ID: id + "-o1", Text: texts[i]}}
			seen := map[string]bool{texts[i]: true}
			for offset := 1; len(options) < 4 && offset < len(texts); offset++ {
				text := texts[(i+offset)%len(texts)]
				if !seen[text] {
					seen[text] = true
					options = append(options, domain.QuizOption{ID: fmt.Sprintf("%s-o%d", id, len(options)+1), Text: text})
				}
			}
			if len(options) != 4 {
				panic("insufficient vocabulary choices")
			}
			correct[id] = options[0].ID
			shiftOptions(options, (i+kind)%4)
			prompt, label := fmt.Sprintf("Apa arti %s?", w.Term), "Arti kata"
			switch kind {
			case 1:
				prompt = fmt.Sprintf("Bagaimana cara membaca %s?", w.Term)
				label = "Cara baca"
				if w.Term == w.Reading {
					prompt = fmt.Sprintf("Mana penulisan bahasa Jepang untuk “%s”?", w.Meaning)
					label = "Penulisan kata"
				}
			case 2:
				prompt = fmt.Sprintf("Kalimat mana yang berarti:\n“%s”", w.ExampleMeaning)
				label = "Pemakaian kalimat"
			}
			q := domain.VocabularyQuestion{ID: id, Prompt: prompt, Options: options, Kind: label}
			if kind == 0 {
				q.ImageAsset = w.ImageAsset
				q.ImageURL = s.assetURLLocked(w.ImageAsset)
				q.ImagePanel = w.ImagePanel
				q.ImageDescription = "Ilustrasi petunjuk konteks kosakata"
			}
			questions = append(questions, q)
			explanations[id] = fmt.Sprintf("%s（%s）berarti %s.\n%s\n%s\n%s", w.Term, w.Reading, w.Meaning, w.Example, w.Examples[0].Reading, w.ExampleMeaning)
		}
	}
	return questions, correct, explanations
}

func vocabularyLevel(level, title, description, lessonID, lessonTitle, lessonDescription string, words []domain.VocabularyWord) domain.VocabularyLevelContent {
	return domain.VocabularyLevelContent{Level: level, Title: title, Description: description, RequiredTier: domain.TierFree, Lessons: []domain.VocabularyLessonContent{{ID: lessonID, Order: 1, Title: lessonTitle, Description: lessonDescription, Words: words}}}
}

func word(id, term, reading, meaning, example, exampleMeaning, note string) domain.VocabularyWord {
	return domain.VocabularyWord{ID: id, Term: term, Reading: reading, Meaning: meaning, Example: example, ExampleMeaning: exampleMeaning, Note: note}
}
