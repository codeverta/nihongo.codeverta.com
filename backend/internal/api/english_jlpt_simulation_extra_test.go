package api

import (
	"testing"
	"time"

	"nihongo-master-backend/internal/domain"
	"nihongo-master-backend/internal/store"
)

func TestEnglishCatalogCoversExtraJLPTPapers(t *testing.T) {
	s := store.NewMemoryStore()
	defer s.Close()
	const user = "english-jlpt-extra-coverage"
	token, err := s.BillingAccountToken(user)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SaveVerifiedPurchase(user, domain.Purchase{
		Platform: "android", ID: user, ProductID: "nihongo_master_lifetime",
		AccountToken: token, Proof: "coverage-test", State: "purchased", CheckedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	for _, level := range []string{"N5-2", "N5-3", "N5-4", "N5-5", "N4-2", "N4-3", "N4-4", "N4-5", "N5-6", "N5-7", "N4-6", "N4-7", "N5-8", "N5-9", "N5-10", "N4-8", "N4-9", "N4-10", "N3-2", "N3-3", "N2-2", "N2-3", "N1-2", "N1-3", "N3-4", "N3-5", "N3-6", "N3-7", "N3-8", "N3-9", "N3-10", "N2-4", "N2-5", "N2-6", "N2-7", "N2-8", "N2-9", "N2-10", "N1-4", "N1-5", "N1-6", "N1-7", "N1-8", "N1-9", "N1-10"} {
		attempt, err := s.StartSimulation(user, level)
		if err != nil {
			t.Fatalf("start %s: %v", level, err)
		}
		session, err := s.SimulationSession(user, attempt.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, section := range session.Sections {
			if session, err = s.SubmitSimulationSection(user, attempt.ID, section.ID); err != nil {
				t.Fatalf("advance %s: %v", level, err)
			}
		}
		if session.Result == nil || len(session.Result.Reviews) == 0 {
			t.Fatalf("%s returned no reviews", level)
		}
		for _, review := range session.Result.Reviews {
			english, ok := englishExamSSWText("simulation.explanation", review.Explanation)
			if !ok || english == "" || english == review.Explanation {
				t.Errorf("%s explanation has no English translation: %q", level, review.Explanation)
			}
		}
	}
}
