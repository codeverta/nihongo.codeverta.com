package store

import (
	"errors"
	"testing"
)

func TestOpenLearningAccessUnlocksStudyAndExamsForFreeAccounts(t *testing.T) {
	previous := SetOpenLearningAccess(true)
	t.Cleanup(func() { SetOpenLearningAccess(previous) })
	s := NewMemoryStore()
	for _, examID := range []string{"jlpt-n3-simulation-1", "jlpt-n2-simulation-1", "jlpt-n1-simulation-1", "jft-a2-full-1", "nat-5q-full-6", "jft-a2-full-10"} {
		if _, err := s.StartExam("free", examID); err != nil {
			t.Fatalf("%s should be open to a free account: %v", examID, err)
		}
		if s.ExamRequiresPremium(examID) {
			t.Fatalf("%s still reports premium", examID)
		}
	}
	if PremiumLevel("N3") || PremiumLevel("N1") {
		t.Fatal("N3-N1 lessons should be open")
	}
	if s.PremiumRequest("/v1/grammar/course", "N3", "", "") || s.PremiumRequest("/v1/simulation/N3", "N3", "", "") || s.PremiumRequest("/v1/ssw/kaigo/practice", "KAIGO", "", "") {
		t.Fatal("PremiumRequest should be false in open mode")
	}
	for _, exam := range s.Exams("") {
		if exam.RequiredTier != "free" {
			t.Fatalf("%s exposes tier %s", exam.ID, exam.RequiredTier)
		}
	}
	if _, err := s.StartExam("free", "nat-5q-full-11"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown paper must stay not found, got %v", err)
	}
}

func TestOpenLearningAccessCanBeSwitchedOff(t *testing.T) {
	previous := SetOpenLearningAccess(false)
	t.Cleanup(func() { SetOpenLearningAccess(previous) })
	s := NewMemoryStore()
	if _, err := s.StartExam("free", "jlpt-n3-simulation-1"); !errors.Is(err, ErrPremium) {
		t.Fatal("gate should return when open access is off")
	}
}
