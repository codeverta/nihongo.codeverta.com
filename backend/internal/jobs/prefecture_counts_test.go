package jobs

import (
	"context"
	"testing"
	"time"

	"nihongo-master-backend/internal/jobfetch"
)

func TestPrefectureCountsOnlyCountActiveJobs(t *testing.T) {
	r, _ := testRepository(t)
	r.sources = r.sources[:1]
	ctx := context.Background()
	now := time.Date(2026, 10, 18, 1, 0, 0, 0, time.UTC)
	r.now = func() time.Time { return now }
	osaka, unknown := sample("osaka"), sample("unknown")
	osaka.Prefecture = "27"
	unknown.Prefecture, unknown.Location = "", "Japan"
	id := r.sources[0].ID
	f := &fakeFetcher{results: map[string]jobfetch.Result{id: {Jobs: []jobfetch.Job{sample("tokyo-1"), sample("tokyo-2"), osaka, unknown}, Complete: true}}, errors: map[string]error{}}
	if _, err := r.SyncDue(ctx, f); err != nil {
		t.Fatal(err)
	}
	counts, err := r.PrefectureCounts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if counts["13"] != 2 || counts["27"] != 1 || len(counts) != 2 {
		t.Fatalf("unexpected counts %v", counts)
	}
	now = now.Add(StaleAfter + time.Hour)
	if counts, _ = r.PrefectureCounts(ctx); len(counts) != 0 {
		t.Fatalf("stale jobs must not be counted: %v", counts)
	}
}
