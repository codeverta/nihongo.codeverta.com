package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nihongo-master-backend/internal/anki"
	"nihongo-master-backend/internal/api"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/jobfetch"
	"nihongo-master-backend/internal/jobs"
	"nihongo-master-backend/internal/news"
	"nihongo-master-backend/internal/testdb"
)

// Phase 2 of docs/rewarded-unlock-spec.md: Anki decks, the N3 simulation,
// news articles and job application links.

const n3SimulationExam = "jlpt-n3-simulation-1"

func gatedAds(extra string) string {
	return `{"enabled":true,"unlockRewarded":{"enabled":true,"unitId":"` + unlockUnit + `"},"unlockDailyCap":20,"unlockFallbackPerDay":5` + extra + `}`
}

func (e *unlockEnv) fallback(t *testing.T, kind, item string) {
	t.Helper()
	data := e.data(t, e.do(http.MethodPost, "/v1/unlocks/fallback", `{"kind":"`+kind+`","itemId":"`+item+`"}`), http.StatusOK)
	if data["unlocked"] != true {
		t.Fatalf("fallback %s/%s = %v", kind, item, data)
	}
}

func (e *unlockEnv) unlockRef(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	body := e.errorBody(t, response, http.StatusPaymentRequired)
	if body["unlockable"] != true {
		t.Fatalf("no ad offered: %v", body)
	}
	ref, _ := body["unlock"].(map[string]any)
	return ref
}

func (e *unlockEnv) unlockKeys(t *testing.T) map[string]bool {
	t.Helper()
	list := e.data(t, e.do(http.MethodGet, "/v1/unlocks", ""), http.StatusOK)
	keys := map[string]bool{}
	for _, item := range list["items"].([]any) {
		keys[item.(string)] = true
	}
	return keys
}

func lockedAnkiDecks(t *testing.T) []string {
	t.Helper()
	decks, err := anki.List("", "")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, deck := range decks {
		if deck.Level == "N3" {
			ids = append(ids, deck.ID)
		}
	}
	if len(ids) < 2 {
		t.Skip("needs two built-in N3 decks")
	}
	return ids
}

func TestAnAdOpensOneAnkiDeck(t *testing.T) {
	e := newUnlockEnv(t, gatedAds(""))
	decks := lockedAnkiDecks(t)
	deck, other := decks[0], decks[1]
	study := "/v1/anki/decks/" + deck + "/study"

	if ref := e.unlockRef(t, e.do(http.MethodGet, study, "")); ref["kind"] != "anki" || ref["itemId"] != deck {
		t.Fatalf("unlock ref = %v", ref)
	}
	if view := e.data(t, e.do(http.MethodGet, "/v1/anki/decks/"+deck, ""), http.StatusOK); view["locked"] != true {
		t.Fatalf("deck view before unlock = %v", view["locked"])
	}

	e.fallback(t, "anki", deck)

	if response := e.do(http.MethodGet, study, ""); response.Code != http.StatusOK {
		t.Fatalf("study an unlocked deck: %d %s", response.Code, response.Body.String())
	}
	if view := e.data(t, e.do(http.MethodGet, "/v1/anki/decks/"+deck, ""), http.StatusOK); view["locked"] != false {
		t.Fatalf("deck view after unlock = %v", view["locked"])
	}
	if response := e.do(http.MethodPost, "/v1/anki/decks/"+deck+"/review", `{}`); response.Code == http.StatusPaymentRequired {
		t.Fatalf("review of an unlocked deck was refused: %s", response.Body.String())
	}
	if response := e.do(http.MethodGet, "/v1/anki/decks/"+other+"/study", ""); response.Code != http.StatusPaymentRequired {
		t.Fatalf("another N3 deck opened: %d", response.Code)
	}
}

func TestAnkiAdOfferFollowsItsSwitch(t *testing.T) {
	e := newUnlockEnv(t, gatedAds(`,"unlockKinds":{"lessons":true,"anki":false,"simulation":true,"news":true,"jobs":true}`))
	deck := lockedAnkiDecks(t)[0]
	body := e.errorBody(t, e.do(http.MethodGet, "/v1/anki/decks/"+deck+"/study", ""), http.StatusPaymentRequired)
	if _, offered := body["unlockable"]; offered {
		t.Fatalf("an ad was offered while Anki unlocks are switched off: %v", body)
	}
	if body := e.errorBody(t, e.do(http.MethodPost, "/v1/unlocks/intents", `{"kind":"anki","itemId":"`+deck+`"}`), http.StatusForbidden); body["code"] != "unlock_unavailable" {
		t.Fatalf("intent = %v", body)
	}
}

// finishSimulation submits every section of an attempt.
func (e *unlockEnv) finishSimulation(t *testing.T, attemptID string) {
	t.Helper()
	path := "/v1/simulation/attempts/" + attemptID
	session := e.data(t, e.do(http.MethodGet, path, ""), http.StatusOK)
	for _, section := range session["sections"].([]any) {
		id := section.(map[string]any)["id"].(string)
		if response := e.do(http.MethodPost, path+"/submit-section", `{"section":"`+id+`"}`); response.Code != http.StatusOK {
			t.Fatalf("submit %s: %d %s", id, response.Code, response.Body.String())
		}
	}
}

func (e *unlockEnv) startExam(examID string) *httptest.ResponseRecorder {
	return e.do(http.MethodPost, "/v1/exam-attempts", `{"examId":"`+examID+`"}`)
}

func TestOneAdBuysOneN3SimulationAttempt(t *testing.T) {
	e := newUnlockEnv(t, gatedAds(""))

	if ref := e.unlockRef(t, e.startExam(n3SimulationExam)); ref["kind"] != "simulation" || ref["itemId"] != n3SimulationExam {
		t.Fatalf("unlock ref = %v", ref)
	}
	if response := e.do(http.MethodGet, "/v1/simulation/N3", ""); response.Code != http.StatusOK {
		t.Fatalf("the overview must open so the app can offer an ad: %d", response.Code)
	}
	if response := e.startExam("jlpt-n3-simulation-2"); response.Code != http.StatusCreated {
		t.Fatalf("paper 02 is free: %d %s", response.Code, response.Body.String())
	}

	e.fallback(t, "simulation", n3SimulationExam)
	if data := e.data(t, e.do(http.MethodPost, "/v1/unlocks/intents", `{"kind":"simulation","itemId":"`+n3SimulationExam+`"}`), http.StatusOK); data["alreadyUnlocked"] != true {
		t.Fatalf("an unspent credit needs no second ad: %v", data)
	}

	attempt := e.data(t, e.startExam(n3SimulationExam), http.StatusCreated)
	attemptID := attempt["id"].(string)
	if response := e.do(http.MethodGet, "/v1/simulation/attempts/"+attemptID, ""); response.Code != http.StatusOK {
		t.Fatalf("paid attempt: %d %s", response.Code, response.Body.String())
	}
	keys := e.unlockKeys(t)
	if keys["simulation:"+n3SimulationExam] || !keys["simulation-attempt:"+attemptID] {
		t.Fatalf("the credit must turn into this attempt: %v", keys)
	}

	// Resuming the same attempt costs nothing.
	if resumed := e.data(t, e.startExam(n3SimulationExam), http.StatusCreated); resumed["id"] != attemptID {
		t.Fatalf("resume started %v, want %s", resumed["id"], attemptID)
	}

	// After it is finished, the next attempt needs another ad.
	e.finishSimulation(t, attemptID)
	if ref := e.unlockRef(t, e.startExam(n3SimulationExam)); ref["kind"] != "simulation" {
		t.Fatalf("second attempt without a new ad: %v", ref)
	}
	e.fallback(t, "simulation", n3SimulationExam)
	second := e.data(t, e.startExam(n3SimulationExam), http.StatusCreated)
	if second["id"] == attemptID {
		t.Fatal("a finished attempt was reused")
	}
	if response := e.do(http.MethodGet, "/v1/simulation/attempts/"+attemptID, ""); response.Code != http.StatusOK {
		t.Fatalf("the first attempt's review must stay readable: %d", response.Code)
	}
}

func TestOtherPremiumSimulationsStayLifetimeOnly(t *testing.T) {
	e := newUnlockEnv(t, gatedAds(""))
	for _, exam := range []string{"jlpt-n2-simulation-1", "jft-a2-full-1"} {
		body := e.errorBody(t, e.startExam(exam), http.StatusPaymentRequired)
		if _, offered := body["unlockable"]; offered {
			t.Fatalf("%s offered an ad: %v", exam, body)
		}
		if response := e.do(http.MethodPost, "/v1/unlocks/intents", `{"kind":"simulation","itemId":"`+exam+`"}`); response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("%s intent: %d", exam, response.Code)
		}
	}
	e.makeLifetime(t)
	if response := e.startExam(n3SimulationExam); response.Code != http.StatusCreated {
		t.Fatalf("Lifetime start: %d %s", response.Code, response.Body.String())
	}
}

// --- News and jobs need MariaDB-backed repositories. ---

type unlockJobsFetcher struct{ source string }

func (f unlockJobsFetcher) Fetch(_ context.Context, id string, _ jobfetch.Validators) (jobfetch.Result, error) {
	if id != f.source {
		return jobfetch.Result{Complete: true}, nil
	}
	salary := 6000000.0
	return jobfetch.Result{Complete: true, Jobs: []jobfetch.Job{{
		ExternalID: "unlock-1", Title: "Software Engineer", Company: "Example Employer", Description: "Role",
		Location: "Tokyo, Japan", Country: "JP", Prefecture: "13", Category: "engineering", EmploymentType: "full-time",
		Remote: "hybrid", Visa: "unknown", JapaneseLevel: "unknown", SalaryMin: &salary, SalaryCurrency: "JPY", SalaryPeriod: "year",
		SourceURL: "https://jobs.lever.co/example/unlock-1", ApplyURL: "https://jobs.lever.co/example/unlock-1/apply",
	}}}, nil
}

type contentEnv struct {
	*unlockEnv
	articleID string
	jobID     string
}

func newContentEnv(t *testing.T, adsConfig string) *contentEnv {
	t.Helper()
	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { _ = sqlDB.Close() })
	ctx := context.Background()

	newsRepo, err := news.New(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newsRepo.Sync(ctx, &newsFixtureFetcher{}, "global-voices", false); err != nil {
		t.Fatal(err)
	}
	page, err := newsRepo.List(ctx, 1, 20, "", false)
	if err != nil || len(page.Items) == 0 {
		t.Fatalf("news seed: %v", err)
	}

	jobsRepo, err := jobs.New(db, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	sources, err := jobsRepo.Sources(ctx)
	if err != nil || len(sources) == 0 {
		t.Fatalf("job sources: %v", err)
	}
	if _, err := jobsRepo.SyncSelected(ctx, unlockJobsFetcher{source: sources[0].ID}, []string{sources[0].ID}); err != nil {
		t.Fatal(err)
	}
	listed, err := jobsRepo.List(ctx, jobs.Filter{Limit: 20})
	if err != nil || len(listed.Items) == 0 {
		t.Fatalf("job seed: %v (%d jobs)", err, len(listed.Items))
	}

	env := newUnlockEnvWith(t, adsConfig, func(server *api.Server) {
		if err := server.ConfigureNews(newsRepo, &newsFixtureFetcher{}, "news-key", t.TempDir(), ctx); err != nil {
			t.Fatal(err)
		}
		server.ConfigureJobs(jobsRepo)
	})
	return &contentEnv{unlockEnv: env, articleID: page.Items[0].ID, jobID: listed.Items[0].ID}
}

func (e *contentEnv) guest(method, path string) *httptest.ResponseRecorder {
	return (&unlockEnv{handler: e.handler}).do(method, path, "")
}

func TestNewsAndJobsStayOpenUntilTheirGatesAreSwitchedOn(t *testing.T) {
	e := newContentEnv(t, gatedAds(""))
	if response := e.guest(http.MethodGet, "/v1/news/"+e.articleID); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "paragraphs") {
		t.Fatalf("guest read with the gate off: %d", response.Code)
	}
	detail := e.data(t, e.guest(http.MethodGet, "/v1/jobs/"+e.jobID), http.StatusOK)
	if detail["applyUrl"] == "" || detail["applyUrl"] == nil {
		t.Fatalf("job links must stay public while the gate is off: %v", detail)
	}
	if link := e.data(t, e.do(http.MethodPost, "/v1/jobs/"+e.jobID+"/apply-link", ""), http.StatusOK); link["applyUrl"] != detail["applyUrl"] {
		t.Fatalf("apply link = %v", link)
	}
	for _, item := range []string{`{"kind":"news","itemId":"` + e.articleID + `"}`, `{"kind":"job","itemId":"` + e.jobID + `"}`} {
		if response := e.do(http.MethodPost, "/v1/unlocks/intents", item); response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("intent for open content %s: %d", item, response.Code)
		}
	}
}

func TestAnAdOpensOneNewsArticleWhenTheGateIsOn(t *testing.T) {
	e := newContentEnv(t, gatedAds(`,"contentGates":{"news":true,"jobs":false}`))
	path := "/v1/news/" + e.articleID

	if body := e.errorBody(t, e.guest(http.MethodGet, path), http.StatusUnauthorized); body["code"] != "login_required" {
		t.Fatalf("guest = %v", body)
	}
	if response := e.guest(http.MethodGet, "/v1/news"); response.Code != http.StatusOK {
		t.Fatalf("the list stays public: %d", response.Code)
	}
	if ref := e.unlockRef(t, e.do(http.MethodGet, path, "")); ref["kind"] != "news" || ref["itemId"] != e.articleID {
		t.Fatalf("unlock ref = %v", ref)
	}

	claimed := e.watchAdFor(t, "news", e.articleID, "tx-news")
	if claimed["contentKey"] != "news:"+e.articleID {
		t.Fatalf("claim = %v", claimed)
	}
	if response := e.do(http.MethodGet, path, ""); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "paragraphs") {
		t.Fatalf("unlocked article: %d %s", response.Code, response.Body.String())
	}
}

func TestNewsGateWithoutAdsLeavesOnlyLifetime(t *testing.T) {
	e := newContentEnv(t, gatedAds(`,"contentGates":{"news":true,"jobs":false},"unlockKinds":{"lessons":true,"anki":true,"simulation":true,"news":false,"jobs":true}`))
	body := e.errorBody(t, e.do(http.MethodGet, "/v1/news/"+e.articleID, ""), http.StatusPaymentRequired)
	if _, offered := body["unlockable"]; offered {
		t.Fatalf("an ad was offered while news unlocks are off: %v", body)
	}
	e.makeLifetime(t)
	if response := e.do(http.MethodGet, "/v1/news/"+e.articleID, ""); response.Code != http.StatusOK {
		t.Fatalf("Lifetime read: %d", response.Code)
	}
}

func TestAnAdOpensOneJobApplicationWhenTheGateIsOn(t *testing.T) {
	e := newContentEnv(t, gatedAds(`,"contentGates":{"news":false,"jobs":true}`))

	for _, path := range []string{"/v1/jobs/" + e.jobID, "/v1/jobs"} {
		response := e.guest(http.MethodGet, path)
		if response.Code != http.StatusOK {
			t.Fatalf("%s: %d", path, response.Code)
		}
		raw := response.Body.String()
		if strings.Contains(raw, "jobs.lever.co") || !strings.Contains(raw, `"applyLocked":true`) {
			t.Fatalf("%s leaks a link or misses applyLocked: %s", path, raw)
		}
	}
	if response := e.guest(http.MethodPost, "/v1/jobs/"+e.jobID+"/apply-link"); response.Code != http.StatusUnauthorized {
		t.Fatalf("guest apply link: %d", response.Code)
	}
	if ref := e.unlockRef(t, e.do(http.MethodPost, "/v1/jobs/"+e.jobID+"/apply-link", "")); ref["kind"] != "job" || ref["itemId"] != e.jobID {
		t.Fatalf("unlock ref = %v", ref)
	}

	e.watchAdFor(t, "job", e.jobID, "tx-job")
	link := e.data(t, e.do(http.MethodPost, "/v1/jobs/"+e.jobID+"/apply-link", ""), http.StatusOK)
	if link["applyUrl"] != "https://jobs.lever.co/example/unlock-1/apply" {
		t.Fatalf("apply link = %v", link)
	}
	var detail struct {
		Data map[string]any `json:"data"`
	}
	_ = json.Unmarshal(e.do(http.MethodGet, "/v1/jobs/"+e.jobID, "").Body.Bytes(), &detail)
	if detail.Data["applyUrl"] != "" {
		t.Fatalf("the public detail stays the same for everyone (it is cached publicly): %v", detail.Data["applyUrl"])
	}
}

func TestLifetimeAppliesWithoutAnAd(t *testing.T) {
	e := newContentEnv(t, gatedAds(`,"contentGates":{"news":true,"jobs":true}`))
	e.makeLifetime(t)
	if link := e.data(t, e.do(http.MethodPost, "/v1/jobs/"+e.jobID+"/apply-link", ""), http.StatusOK); link["applyUrl"] == "" {
		t.Fatalf("Lifetime apply link = %v", link)
	}
	if response := e.do(http.MethodGet, "/v1/news/"+e.articleID, ""); response.Code != http.StatusOK {
		t.Fatalf("Lifetime news: %d", response.Code)
	}
}

// watchAdFor runs the verified ad flow for any kind of item.
func (e *unlockEnv) watchAdFor(t *testing.T, kind, item, transaction string) map[string]any {
	t.Helper()
	intent := e.data(t, e.do(http.MethodPost, "/v1/unlocks/intents", `{"kind":"`+kind+`","itemId":"`+item+`"}`), http.StatusOK)
	nonce, _ := intent["nonce"].(string)
	if nonce == "" {
		t.Fatalf("intent = %v", intent)
	}
	if response := e.callback(e.userID, nonce, transaction); response.Code != http.StatusOK {
		t.Fatalf("callback %d", response.Code)
	}
	return e.data(t, e.do(http.MethodPost, "/v1/unlocks/claim", `{"nonce":"`+nonce+`"}`), http.StatusOK)
}
