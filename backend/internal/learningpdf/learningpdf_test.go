package learningpdf_test

import (
	"bytes"
	"strconv"
	"testing"
	"time"

	"nihongo-master-backend/internal/learningpdf"
	"nihongo-master-backend/internal/store"
)

func TestRendersEveryLevelAsPDF(t *testing.T) {
	memory := store.NewMemoryStore()
	defer func() { _ = memory.Close() }()
	for _, level := range []string{"N5", "N4", "N3", "N2", "N1"} {
		for _, locale := range []string{"id", "en"} {
			started := time.Now()
			grammar := memory.GrammarLevelLessons("", level)
			if len(grammar) == 0 {
				t.Fatalf("%s: no grammar lessons", level)
			}
			out, err := learningpdf.Grammar(level, locale, grammar)
			if err != nil || !bytes.HasPrefix(out, []byte("%PDF")) {
				t.Fatalf("grammar %s/%s: %v", level, locale, err)
			}
			kanji := memory.KanjiLevelLessons("", level)
			if len(kanji) == 0 {
				t.Fatalf("%s: no kanji lessons", level)
			}
			kout, err := learningpdf.Kanji(level, locale, kanji)
			if err != nil || !bytes.HasPrefix(kout, []byte("%PDF")) {
				t.Fatalf("kanji %s/%s: %v", level, locale, err)
			}
			t.Logf("%s/%s grammar=%dKB kanji=%dKB in %s", level, locale, len(out)/1024, len(kout)/1024, time.Since(started).Round(time.Millisecond))
		}
	}
}

func TestRendersEverySimulationPaper(t *testing.T) {
	memory := store.NewMemoryStore()
	defer func() { _ = memory.Close() }()
	levels := []string{"N5", "N4", "N3", "N2", "N1", "JFT", "NAT"}
	for _, base := range []string{"N5", "N4", "N3", "N2", "N1"} {
		for paper := 2; paper <= 10; paper++ {
			levels = append(levels, base+"-"+strconv.Itoa(paper))
		}
	}
	for paper := 2; paper <= 10; paper++ {
		levels = append(levels, "JFT-"+strconv.Itoa(paper), "NAT-"+strconv.Itoa(paper))
	}
	levels = append(levels, store.SSWLevels()...)
	for _, level := range levels {
		paper, err := memory.SimulationPaper(level)
		if err != nil {
			t.Fatalf("%s: %v", level, err)
		}
		if len(paper.Reviews) == 0 || paper.Title == "" {
			t.Fatalf("%s: empty paper", level)
		}
		firstCorrectFirst := 0
		for _, review := range paper.Reviews {
			if review.Question.Options[0].ID == review.CorrectOptionID {
				firstCorrectFirst++
			}
		}
		if firstCorrectFirst == len(paper.Reviews) {
			t.Fatalf("%s: the correct answer must not always be option 1", level)
		}
		out, err := learningpdf.Exam(learningpdf.ExamPaper{Title: paper.Title, Sections: paper.Sections, Omitted: paper.Omitted, Reviews: paper.Reviews}, "id")
		if err != nil || !bytes.HasPrefix(out, []byte("%PDF")) {
			t.Fatalf("%s: %v", level, err)
		}
	}
	t.Logf("rendered %d simulation papers", len(levels))
	if _, err := memory.SimulationPaper("N9"); err == nil {
		t.Fatal("unknown level must be rejected")
	}
}
