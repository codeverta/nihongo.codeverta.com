package dictionary

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"nihongo-master-backend/internal/testdb"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCanonicalQueries(t *testing.T) {
	for _, tc := range []struct{ query, mode, want, kind string }{
		{" KONNICHIWA ", "auto", "こんにちは", "japanese"},
		{"konnichiha", "japanese", "こんにちは", "japanese"},
		{"コンニチハ", "auto", "こんにちは", "japanese"},
		{"ｺﾝﾆﾁﾊ", "auto", "こんにちは", "japanese"},
		{"こんにちわ", "auto", "こんにちは", "japanese"},
		{"ＮＥＫＯ", "auto", "ねこ", "japanese"},
		{"pan", "auto", "pan", "raw"}, {"pan", "japanese", "ぱん", "japanese"}, {"pan", "english", "pan", "english"},
		{"gakkou", "japanese", "がっこう", "japanese"}, {"kan'i", "japanese", "かんい", "japanese"},
		{"shinbun", "japanese", "しんぶん", "japanese"}, {"ryokou", "japanese", "りょこう", "japanese"},
		{"matcha", "japanese", "まっちゃ", "japanese"},
		{"Tōkyō", "japanese", "とうきょう", "japanese"},
		{"kōhī", "japanese", "こーひー", "japanese"},
		{"コー", "japanese", "こー", "japanese"}, {"こう", "japanese", "こう", "japanese"},
		{"橋", "auto", "橋", "japanese"}, {"箸", "auto", "箸", "japanese"},
		{"じ", "auto", "じ", "japanese"}, {"ぢ", "auto", "ぢ", "japanese"},
		{"cat #common", "auto", "cat #common", "raw"},
	} {
		t.Run(tc.query+tc.mode, func(t *testing.T) {
			q, kind, err := canonicalQuery(tc.query, tc.mode)
			if err != nil || q != tc.want || kind != tc.kind {
				t.Fatalf("%q %q %v", q, kind, err)
			}
		})
	}
	if _, _, err := canonicalQuery("xyz", "japanese"); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid romaji accepted")
	}
	if _, _, err := canonicalQuery("cat", "other"); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid mode accepted")
	}
	endpoint, _ := searchEndpoint("pan", 1, "N5", true, "english")
	u, _ := url.Parse(endpoint)
	if u.Query().Get("keyword") != "\"pan\" #jlpt-n5 #common" {
		t.Fatal(endpoint)
	}
}

func TestEquivalentScriptsSharePermanentDatabaseAcrossUsersAndRestart(t *testing.T) {
	dsn := testdb.DSN(t)
	c, db := persistentTestClient(t, dsn)
	calls := 0
	c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Query().Get("keyword") != "こんにちは" {
			t.Error("upstream query not canonical")
		}
		return response(200, wordFixture), nil
	})
	if _, err := c.Search(context.Background(), "konnichiwa", 1, "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE dictionary_responses SET fetched_at=1"); err != nil {
		t.Fatal(err)
	}
	second, _ := persistentTestClient(t, dsn)
	second.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		t.Error("second user hit API")
		return nil, errors.New("offline")
	})
	for _, q := range []string{"こんにちは", "コンニチハ", "ｺﾝﾆﾁﾊ", "KONNICHIWA"} {
		if _, err := second.Search(context.Background(), q, 1, "", false); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 || second.Stats()["databaseHits"] != 1 {
		t.Fatal("alias reuse failed")
	}
	rows, total, err := second.Persistence.List(context.Background(), 1, "こんにちは")
	if err != nil || total != 1 || len(rows) != 1 || rows[0].Mode != "japanese" {
		t.Fatalf("metadata missing: %+v %v", rows, err)
	}
}

func TestEmptyResultsArePermanentAndRefreshIsAtomic(t *testing.T) {
	c, db := persistentTestClient(t, testdb.DSN(t))
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		return response(200, `{"meta":{"status":200},"data":[]}`), nil
	})
	if _, err := c.Search(context.Background(), "empty", 1, "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE dictionary_responses SET fetched_at=1"); err != nil {
		t.Fatal(err)
	}
	c.cache = map[string]cached{}
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { return nil, errors.New("offline") })
	if result, err := c.Search(context.Background(), "empty", 1, "", false); err != nil || len(result.Entries) != 0 {
		t.Fatal("empty response expired", err)
	}
	rows, _, _ := c.Persistence.List(context.Background(), 1, "")
	c.next = time.Time{}
	if err := c.Refresh(context.Background(), rows[0].ID); err == nil {
		t.Fatal("failed refresh reported success")
	}
	if result, err := c.Search(context.Background(), "empty", 1, "", false); err != nil || len(result.Entries) != 0 {
		t.Fatal("failed refresh destroyed memory", err)
	}
	var fetched int64
	db.QueryRow("SELECT fetched_at FROM dictionary_responses").Scan(&fetched)
	if fetched != 1 {
		t.Fatal("failed refresh modified DB")
	}
	c.next = time.Time{}
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { return response(200, wordFixture), nil })
	if err := c.Refresh(context.Background(), rows[0].ID); err != nil {
		t.Fatal(err)
	}
	if result, err := c.Search(context.Background(), "empty", 1, "", false); err != nil || len(result.Entries) != 1 {
		t.Fatal("refresh did not replace memory", err)
	}
	c.cache = map[string]cached{}
	if result, err := c.Search(context.Background(), "empty", 1, "", false); err != nil || len(result.Entries) != 1 {
		t.Fatal("refresh did not persist", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_refresh BEFORE UPDATE ON dictionary_entries FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='refresh write failure'`); err != nil {
		t.Fatal(err)
	}
	c.next = time.Time{}
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		return response(200, strings.ReplaceAll(wordFixture, `"cat"`, `"kitten"`)), nil
	})
	if err := c.Refresh(context.Background(), rows[0].ID); err == nil {
		t.Fatal("DB write failure reported refresh success")
	}
	c.cache = map[string]cached{}
	if result, err := c.Search(context.Background(), "empty", 1, "", false); err != nil || !strings.Contains(string(result.Entries[0]), `"cat"`) {
		t.Fatal("failed write corrupted persisted data", err)
	}
}

func TestSharedAliasRequestSurvivesFirstCallerCancellation(t *testing.T) {
	c := NewClient()
	entered := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		close(entered)
		select {
		case <-release:
			return response(200, wordFixture), nil
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	first := make(chan error, 1)
	go func() { _, err := c.Search(ctx, "konnichiwa", 1, "", false); first <- err }()
	<-entered
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			q := []string{"こんにちは", "コンニチハ", "ｺﾝﾆﾁﾊ"}[i%3]
			if _, err := c.Search(context.Background(), q, 1, "", false); err != nil {
				t.Error(err)
			}
		}(i)
	}
	cancel()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	close(release)
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatal("equivalent concurrent queries fetched more than once")
	}
}

func TestMemoryCacheEnforcesByteBudget(t *testing.T) {
	c := NewClient()
	for i := 0; i < 20; i++ {
		c.cache[fmt.Sprint(i)] = cached{body: make([]byte, 1<<20), at: time.Now().Add(time.Duration(i) * time.Millisecond)}
	}
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { return response(200, wordFixture), nil })
	if _, err := c.Search(context.Background(), "cat", 1, "", false); err != nil {
		t.Fatal(err)
	}
	if c.Stats()["memoryBytes"] > maxMemoryBytes {
		t.Fatal("memory budget exceeded")
	}
}

func TestModesAndHomophonesKeepIndependentKeys(t *testing.T) {
	c := NewClient()
	calls := 0
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { calls++; return response(200, wordFixture), nil })
	for _, tc := range []struct{ q, mode string }{{"pan", "auto"}, {"pan", "english"}, {"pan", "japanese"}, {"ぱん", "auto"}, {"橋", "auto"}, {"箸", "auto"}} {
		if _, err := c.Search(context.Background(), tc.q, 1, "", false, tc.mode); err != nil {
			t.Fatal(err)
		}
		c.next = time.Time{}
	}
	if calls != 5 {
		t.Fatal("ambiguous modes/homophones mixed", calls)
	}
}

func TestCacheLatency(t *testing.T) {
	c, _ := persistentTestClient(t, testdb.DSN(t))
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		time.Sleep(20 * time.Millisecond)
		return response(200, wordFixture), nil
	})
	start := time.Now()
	if _, err := c.Search(context.Background(), "konnichiwa", 1, "", false); err != nil {
		t.Fatal(err)
	}
	t.Logf("cold fixture API (includes artificial 20ms upstream delay and DB transaction): %s", time.Since(start))
	for _, kind := range []string{"memory", "database"} {
		times := make([]time.Duration, 120)
		before := c.Stats()["upstreamCalls"]
		for i := range times {
			if kind == "database" {
				c.mu.Lock()
				c.cache = map[string]cached{}
				c.mu.Unlock()
			}
			start = time.Now()
			if _, err := c.Search(context.Background(), "コンニチハ", 1, "", false); err != nil {
				t.Fatal(err)
			}
			times[i] = time.Since(start)
		}
		sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
		t.Logf("%s hit n=%d p50=%s p95=%s API calls=%d", kind, len(times), times[59], times[113], c.Stats()["upstreamCalls"]-before)
	}
}

func BenchmarkSearchMemory(b *testing.B) {
	c := NewClient()
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { return response(200, wordFixture), nil })
	if _, err := c.Search(context.Background(), "konnichiwa", 1, "", false); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.Search(context.Background(), "コンニチハ", 1, "", false); err != nil {
			b.Fatal(err)
		}
	}
}

func TestLegacyExactQueryCacheCanBeAdoptedWithoutAPI(t *testing.T) {
	c, _ := persistentTestClient(t, testdb.DSN(t))
	endpoint, _ := searchEndpoint("cat", 1, "", false, "auto")
	legacy := strings.Split(endpoint, "#")[0]
	if err := c.Persistence.save(context.Background(), legacy, cached{body: []byte(wordFixture), at: time.Now()}); err != nil {
		t.Fatal(err)
	}
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		t.Error("legacy hit fetched upstream")
		return nil, errors.New("offline")
	})
	if _, err := c.Search(context.Background(), "cat", 1, "", false); err != nil {
		t.Fatal(err)
	}
}

func TestLiveCanonicalDatabaseReuse(t *testing.T) {
	if os.Getenv("DICTIONARY_LIVE_TEST") != "1" {
		t.Skip("set DICTIONARY_LIVE_TEST=1 for real Jisho + isolated MariaDB test")
	}
	dsn := testdb.DSN(t)
	first, _ := persistentTestClient(t, dsn)
	result, err := first.Search(context.Background(), "konnichiwa", 1, "", false)
	if err != nil || len(result.Entries) == 0 {
		t.Fatal("live canonical query", err)
	}
	second, _ := persistentTestClient(t, dsn)
	second.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		t.Error("second user contacted Jisho")
		return nil, errors.New("network disabled")
	})
	for _, q := range []string{"こんにちは", "コンニチハ", "ｺﾝﾆﾁﾊ"} {
		cached, err := second.Search(context.Background(), q, 1, "", false)
		if err != nil || len(cached.Entries) != len(result.Entries) {
			t.Fatalf("live DB reuse %q %v", q, err)
		}
		for i := range result.Entries {
			if string(cached.Entries[i]) != string(result.Entries[i]) {
				t.Fatal("live entries differ")
			}
		}
	}
	t.Logf("real Jisho returned %d entries; first client API calls=%d; second client DB hits=%d, API calls=%d", len(result.Entries), first.Stats()["upstreamCalls"], second.Stats()["databaseHits"], second.Stats()["upstreamCalls"])
}
