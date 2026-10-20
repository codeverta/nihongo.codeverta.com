package jobs

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/jobfetch"
	"nihongo-master-backend/internal/testdb"
)

type fakeFetcher struct {
	mu         sync.Mutex
	results    map[string]jobfetch.Result
	errors     map[string]error
	calls      []string
	validators []jobfetch.Validators
}

func (f *fakeFetcher) Fetch(ctx context.Context, id string, v jobfetch.Validators) (jobfetch.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, id)
	f.validators = append(f.validators, v)
	return f.results[id], f.errors[id]
}
func testRepository(t *testing.T) (*Repository, *gorm.DB) {
	t.Helper()
	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	r, err := New(db, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	r.requestSpacing = 0
	r.sources = r.sources[:2]
	return r, db
}
func number(v float64) *float64 { return &v }
func sample(id string) jobfetch.Job {
	return jobfetch.Job{ExternalID: id, Title: "Software Engineer " + id, Company: "Example Employer", Description: "Original public role description", Qualifications: "Experience as described by employer", Location: "Tokyo, Japan", Country: "JP", Prefecture: "13", Category: "engineering", EmploymentType: "full-time", Remote: "hybrid", Visa: "unknown", JapaneseLevel: "unknown", SalaryMin: number(6000000), SalaryMax: number(9000000), SalaryCurrency: "JPY", SalaryPeriod: "year", SourceURL: "https://jobs.lever.co/example/" + id, ApplyURL: "https://jobs.lever.co/example/" + id + "/apply"}
}
func baseFilter() Filter { return Filter{Limit: 20} }

func TestFiltersAndPublicURLs(t *testing.T) {
	if CanonicalCategory("it") != "engineering" || CanonicalCategory("administration") != "office" {
		t.Fatal("collector categories were lost")
	}
	for _, v := range []string{"", "168", "84"} {
		if _, err := Interval(v); err != nil {
			t.Fatal(err)
		}
	}
	for _, v := range []string{"1", "24", "167", "999", "-1"} {
		if _, err := Interval(v); err == nil {
			t.Fatal("unsafe schedule", v)
		}
	}
	for _, q := range []string{"limit=31", "limit=0", "salaryMin=NaN&salaryPeriod=year", "salaryMin=Inf&salaryPeriod=year", "salaryMin=2000", "prefecture=99", "remote=yes", "category=made-up", "q=" + string(make([]byte, 161))} {
		v, _ := url.ParseQuery(q)
		if _, err := ParseFilter(v); err == nil {
			t.Fatal("invalid filter accepted", q)
		}
	}
	for _, link := range []string{"http://example.com", "https://localhost/x", "https://127.0.0.1/x", "https://10.0.0.1/x", "https://user:pass@example.com", "javascript:alert(1)", "https://example.com:8443/x"} {
		if PublicHTTPS(link) {
			t.Fatal("unsafe link", link)
		}
	}
	if !PublicHTTPS("https://jobs.lever.co/example/a/apply") {
		t.Fatal("public link rejected")
	}
	if canonical("https://JOBS.lever.co/a/?utm_source=x&ref=hi") != canonical("https://jobs.lever.co/a") {
		t.Fatal("tracking variants not deduplicated")
	}
}

func TestSelectedSyncRequiresWholeDomainAndKeepsCooldown(t *testing.T) {
	r, _ := testRepository(t)
	f := &fakeFetcher{results: map[string]jobfetch.Result{}, errors: map[string]error{}}
	ctx := context.Background()
	if _, err := r.SyncSelected(ctx, f, []string{r.sources[0].ID}); !errors.Is(err, ErrInvalid) {
		t.Fatal("partial shared domain allowed", err)
	}
	if _, err := r.SyncSelected(ctx, f, []string{"unknown"}); !errors.Is(err, ErrInvalid) {
		t.Fatal("unknown source allowed", err)
	}
	ids := []string{r.sources[0].ID, r.sources[1].ID}
	if _, err := r.SyncSelected(ctx, f, ids); err != nil {
		t.Fatal(err)
	}
	if _, err := r.SyncDue(ctx, f); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 2 {
		t.Fatal("manual selection bypassed weekly deadline")
	}
}

func TestCapturedMetadataRepairDoesNotRollBackNewerData(t *testing.T) {
	r, db := testRepository(t)
	r.sources = r.sources[:1]
	ctx := context.Background()
	source := r.sources[0]
	now := time.Now().UTC()
	newer := sample("repair")
	newer.Benefits = "Paid leave and commuting allowance"
	newer.Category = "general"
	newer.Prefecture = ""
	f := &fakeFetcher{results: map[string]jobfetch.Result{source.ID: {Jobs: []jobfetch.Job{newer}, Complete: true}}, errors: map[string]error{}}
	if _, err := r.SyncDue(ctx, f); err != nil {
		t.Fatal(err)
	}
	var before listingRow
	if err := db.First(&before).Error; err != nil {
		t.Fatal(err)
	}
	var deadline domainRow
	db.Where("domain = ?", source.Domain).First(&deadline)
	older := newer
	older.Category = "it"
	older.Prefecture = "Tokyo"
	older.Description = "Old description must not replace current content"
	if _, err := r.ImportCaptured(ctx, source.ID, jobfetch.Result{Jobs: []jobfetch.Job{older}, Complete: true}, now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	var after listingRow
	db.First(&after)
	if after.Category != "engineering" || after.Prefecture != "13" {
		t.Fatal("missing normalized metadata", after.Category, after.Prefecture)
	}
	if before.Benefits != newer.Benefits || after.Benefits != before.Benefits || after.Description != before.Description || !after.LastSeenAt.Equal(before.LastSeenAt) || after.Status != before.Status {
		t.Fatal("older response rolled back fresh data")
	}
	var afterDeadline domainRow
	db.Where("domain = ?", source.Domain).First(&afterDeadline)
	if !afterDeadline.NextFetchAt.Equal(deadline.NextFetchAt) {
		t.Fatal("capture changed network deadline")
	}
	if _, err := r.ImportCaptured(ctx, source.ID, jobfetch.Result{}, now.Add(-25*time.Hour)); !errors.Is(err, ErrInvalid) {
		t.Fatal("old capture accepted")
	}
}

func TestWeeklyDomainLeasePaginationDedupAndRestart(t *testing.T) {
	r, db := testRepository(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 18, 6, 0, 0, 0, time.UTC)
	r.now = func() time.Time { return now }
	a, b := r.sources[0].ID, r.sources[1].ID
	first, second := sample("a"), sample("b")
	duplicate := first
	duplicate.ExternalID = "another-source-a"
	duplicate.ApplyURL += "?utm_source=other"
	f := &fakeFetcher{results: map[string]jobfetch.Result{a: {Jobs: []jobfetch.Job{first}, Complete: true, ETag: `"v1"`}, b: {Jobs: []jobfetch.Job{duplicate, second}, Complete: true}}, errors: map[string]error{}}
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := r.SyncDue(ctx, f); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if len(f.calls) != 2 {
		t.Fatal("concurrent workers duplicated network", f.calls)
	}
	page, err := r.List(ctx, Filter{Limit: 1})
	if err != nil || page.Total != 2 || !page.HasMore {
		t.Fatal(page, err)
	}
	next, err := r.List(ctx, Filter{Limit: 1, Cursor: page.NextCursor})
	if err != nil || len(next.Items) != 1 || next.Items[0].ID == page.Items[0].ID {
		t.Fatal("bad cursor", err)
	}
	if _, err := r.List(ctx, Filter{Limit: 1, Cursor: page.NextCursor, Category: "hospitality"}); !errors.Is(err, ErrInvalid) {
		t.Fatal("cursor reused across filters")
	}
	filtered, err := r.List(ctx, Filter{Limit: 20, Q: "software", Prefecture: "13", Remote: "hybrid", SalaryMin: number(8000000), SalaryPeriod: "year"})
	if err != nil || filtered.Total != 2 {
		t.Fatal(filtered, err)
	}
	filtered, err = r.List(ctx, Filter{Limit: 20, SalaryMin: number(10000000), SalaryPeriod: "year"})
	if err != nil || filtered.Total != 0 {
		t.Fatal("salary bound not applied")
	}
	restarted, err := New(db, 168*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	restarted.sources = r.sources
	restarted.now = r.now
	restarted.requestSpacing = 0
	report, err := restarted.SyncDue(ctx, f)
	if err != nil || report.Requests != 0 || len(f.calls) != 2 {
		t.Fatal("restart bypassed cooldown", report, err)
	}
	now = now.Add(84 * time.Hour)
	report, err = r.SyncDue(ctx, f)
	if err != nil || report.Requests != 0 {
		t.Fatal("weekly source requested twice", report, err)
	}
	now = now.Add(84 * time.Hour)
	f.results[a] = jobfetch.Result{NotModified: true, ETag: `"v1"`}
	f.results[b] = jobfetch.Result{Jobs: []jobfetch.Job{second}, Complete: true}
	if _, err = r.SyncDue(ctx, f); err != nil {
		t.Fatal(err)
	}
	if f.validators[2].ETag != `"v1"` {
		t.Fatal("conditional validator lost", f.validators)
	}
	page, err = r.List(ctx, baseFilter())
	if err != nil || page.Total != 2 {
		t.Fatal("304 failed to retain active job", err)
	}
	now = now.Add(7 * 24 * time.Hour)
	f.results[a] = jobfetch.Result{Complete: true}
	if _, err = r.SyncDue(ctx, f); err != nil {
		t.Fatal(err)
	}
	page, err = r.List(ctx, baseFilter())
	if err != nil || page.Total != 1 || page.Items[0].Title != second.Title {
		t.Fatal("absent job not retired", page, err)
	}
	var rows []listingRow
	db.Find(&rows)
	for _, row := range rows {
		if row.Title == first.Title {
			detail, err := r.Get(ctx, row.ID)
			if err != nil || detail.Status != "closed" {
				t.Fatal("closed deep link", err)
			}
		}
	}
	now = now.Add(31 * 24 * time.Hour)
	page, _ = r.List(ctx, baseFilter())
	if page.Total != 0 {
		t.Fatal("stale jobs remained active")
	}
}

func TestFailedPartialAndRateLimitedSyncNeverClosesGoodData(t *testing.T) {
	r, db := testRepository(t)
	r.sources = r.sources[:1]
	ctx := context.Background()
	now := time.Date(2026, 10, 18, 1, 0, 0, 0, time.UTC)
	r.now = func() time.Time { return now }
	id := r.sources[0].ID
	f := &fakeFetcher{results: map[string]jobfetch.Result{id: {Jobs: []jobfetch.Job{sample("good")}, Complete: true}}, errors: map[string]error{}}
	if _, err := r.SyncDue(ctx, f); err != nil {
		t.Fatal(err)
	}
	now = now.Add(7 * 24 * time.Hour)
	f.results[id] = jobfetch.Result{Complete: false}
	if _, err := r.SyncDue(ctx, f); err != nil {
		t.Fatal(err)
	}
	page, _ := r.List(ctx, baseFilter())
	if page.Total != 1 {
		t.Fatal("partial snapshot closed jobs")
	}
	now = now.Add(7 * 24 * time.Hour)
	f.results[id] = jobfetch.Result{Blocked: true, RetryAfter: now.Add(14 * 24 * time.Hour)}
	f.errors[id] = errors.New("HTTP 429")
	report, err := r.SyncDue(ctx, f)
	if err != nil || report.Requests != 1 || len(report.Errors) != 1 {
		t.Fatal(report, err)
	}
	var d domainRow
	db.First(&d)
	if d.NextFetchAt.Before(now.Add(14 * 24 * time.Hour)) {
		t.Fatal("Retry-After ignored")
	}
	now = now.Add(7 * 24 * time.Hour)
	report, err = r.SyncDue(ctx, f)
	if err != nil || report.Requests != 0 {
		t.Fatal("backoff bypassed")
	}
	page, _ = r.List(ctx, baseFilter())
	if page.Total != 1 {
		t.Fatal("temporary error closed known job")
	}
}

func TestIngestFencingAndDatabaseFailureRollback(t *testing.T) {
	r, db := testRepository(t)
	r.sources = r.sources[:1]
	ctx := context.Background()
	source := r.sources[0]
	owner, _, claimed, err := r.claim(ctx, source.Domain)
	if err != nil || !claimed {
		t.Fatal(err)
	}
	db.Model(&domainRow{}).Where("domain = ?", source.Domain).Update("lease_owner", "00000000-0000-4000-8000-000000000001")
	_, err = r.ingest(ctx, source, source.Domain, owner, jobfetch.Result{Jobs: []jobfetch.Job{sample("x")}, Complete: true}, time.Now().UTC())
	if !errors.Is(err, ErrLeaseLost) {
		t.Fatal("stale worker wrote data", err)
	}
	db.Model(&domainRow{}).Where("domain = ?", source.Domain).Update("lease_owner", owner)
	err = db.Callback().Create().Before("gorm:create").Register("job_test_failure", func(tx *gorm.DB) {
		if tx.Statement.Table == "job_listings" {
			tx.AddError(fmt.Errorf("test database write failure"))
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.ingest(ctx, source, source.Domain, owner, jobfetch.Result{Jobs: []jobfetch.Job{sample("x")}, Complete: true}, time.Now().UTC())
	if err == nil {
		t.Fatal("write failure hidden")
	}
	var count int64
	db.Model(&listingRow{}).Count(&count)
	if count != 0 {
		t.Fatal("failed write published listing")
	}
	db.Model(&observationRow{}).Count(&count)
	if count != 0 {
		t.Fatal("failed write published observation")
	}
}

func TestSalaryUnknownExpiryAndLiteralSearch(t *testing.T) {
	r, _ := testRepository(t)
	r.sources = r.sources[:1]
	ctx := context.Background()
	now := time.Now().UTC()
	r.now = func() time.Time { return now }
	id := r.sources[0].ID
	unknown := sample("unknown")
	unknown.SalaryMin = nil
	unknown.SalaryMax = nil
	unknown.JapaneseLevel = ""
	unknown.Visa = ""
	unknown.Remote = ""
	unknown.Title = "100% hands_on"
	expired := sample("expired")
	past := now.Add(-time.Hour)
	expired.ExpiresAt = &past
	f := &fakeFetcher{results: map[string]jobfetch.Result{id: {Jobs: []jobfetch.Job{unknown, expired}, Complete: true}}, errors: map[string]error{}}
	if _, err := r.SyncDue(ctx, f); err != nil {
		t.Fatal(err)
	}
	page, err := r.List(ctx, Filter{Limit: 20, Q: "%"})
	if err != nil || page.Total != 1 {
		t.Fatal("search wildcard not literal", page, err)
	}
	detail, err := r.Get(ctx, page.Items[0].ID)
	if err != nil || detail.SalaryMin != nil || detail.Visa != "unknown" || detail.JapaneseLevel != "unknown" || detail.Remote != "unknown" {
		t.Fatal("unknown facts fabricated", detail, err)
	}
	filtered, _ := r.List(ctx, Filter{Limit: 20, SalaryMin: number(1), SalaryPeriod: "year"})
	if filtered.Total != 0 {
		t.Fatal("unknown salary passed minimum")
	}
	directory, err := r.Sources(ctx)
	if err != nil || len(directory) < 53 {
		t.Fatal("directory missing", err)
	}
	for _, s := range directory {
		if s.Mode == "directory" && (s.Status != "directory_only" || s.NextFetchAt != nil) {
			t.Fatal("directory falsely scheduled", s)
		}
	}
}
