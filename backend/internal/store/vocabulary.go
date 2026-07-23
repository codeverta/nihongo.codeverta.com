package store

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"nihongo-master-backend/internal/domain"
)

type vocabularyAttemptState struct {
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
		lessons := make([]domain.VocabularyLessonSummary, 0, len(level.Lessons))
		total := 0
		for _, lesson := range level.Lessons {
			progress := s.progress[userID][lesson.ID]
			total += progress.Completion
			lessons = append(lessons, domain.VocabularyLessonSummary{
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
	return domain.VocabularyLesson{
		ID: lesson.ID, Level: level.Level, Title: lesson.Title,
		Description: lesson.Description, BestScore: progress.Completion,
		Completed: progress.Completed,
		Words:     append([]domain.VocabularyWord(nil), lesson.Words...),
	}, nil
}

func (s *MemoryStore) StartVocabularyAttempt(userID, lessonID string) (domain.VocabularyAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, lesson := s.findVocabularyLessonLocked(lessonID)
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
	id := fmt.Sprintf("vocab-attempt-%d", s.attemptID.Add(1))
	attempt := domain.VocabularyAttempt{ID: id, LessonID: lessonID, Questions: questions, StartedAt: time.Now().UTC()}
	s.vocabAttempts[id] = &vocabularyAttemptState{
		attempt: attempt, userID: userID, correct: correct,
		explanations: explanations, answers: map[string]domain.VocabularyAnswerFeedback{},
	}
	return attempt, nil
}

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
	s.progress[userID][state.attempt.LessonID] = domain.LessonProgress{
		LessonID: state.attempt.LessonID, Completion: best,
		LastPosition: len(state.attempt.Questions) * 60,
		Completed:    previous.Completed || passed, UpdatedAt: time.Now().UTC(),
	}
	if passed && !previous.Completed {
		profile := s.ensureProfileLocked(userID)
		profile.XP += 50
		profile.Level = 1 + profile.XP/500
		s.profiles[userID] = profile
	}
	delete(s.vocabAttempts, attemptID)
	return domain.VocabularyAttemptResult{
		AttemptID: attemptID, LessonID: state.attempt.LessonID, Score: score,
		Correct: correct, Total: len(state.attempt.Questions), Passed: passed,
		BestScore: best, Progress: s.vocabularyProgressLocked(userID), CanRetry: true,
	}, nil
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
	return []domain.VocabularyLevelContent{
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
}

func vocabularyLevel(level, title, description, lessonID, lessonTitle, lessonDescription string, words []domain.VocabularyWord) domain.VocabularyLevelContent {
	return domain.VocabularyLevelContent{Level: level, Title: title, Description: description, RequiredTier: domain.TierFree, Lessons: []domain.VocabularyLessonContent{{ID: lessonID, Order: 1, Title: lessonTitle, Description: lessonDescription, Words: words}}}
}

func word(id, term, reading, meaning, example, exampleMeaning, note string) domain.VocabularyWord {
	return domain.VocabularyWord{ID: id, Term: term, Reading: reading, Meaning: meaning, Example: example, ExampleMeaning: exampleMeaning, Note: note}
}
