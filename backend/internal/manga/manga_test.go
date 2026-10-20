package manga

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/testdb"
)

const testURL = "https://rawkuma.net/manga/safe-story/"

func TestParseSeriesIsScopedAndOrdersDecimalChapters(t *testing.T) {
	body := `<html><script type="application/ld+json">{"@type":["Book","ComicSeries"],"url":"` + testURL + `","name":"Safe Story","description":"日本語の物語","genre":["Comedy","School Life"],"alternateName":"Example,物語です","image":{"url":"https://rawkuma.net/wp-content/uploads/cover.jpg"},"author":{"name":"作家"}}</script>
<a href="` + testURL + `chapter-12.5.999/">12.5</a><a href="` + testURL + `chapter-1.100/">1</a><a href="` + testURL + `chapter-1.100/">duplicate</a><a href="https://evil.example/manga/safe-story/chapter-2.101/">ad</a><a href="https://rawkuma.net/manga/other/chapter-1.200/">related</a></html>`
	got, err := parseSeries([]byte(body), testURL)
	if err != nil || len(got.Chapters) != 2 || got.Chapters[0].Number != "1" || got.Chapters[1].Number != "12.5" || got.Series.JapaneseTitle != "物語です" {
		t.Fatalf("parse: %+v %v", got, err)
	}
	for _, bad := range []string{"Ecchi", "Adult", "Smut", "Hentai", "Mature", "Harem", "R18", "エッチ"} {
		blocked := strings.Replace(body, `"Comedy"`, `"`+bad+`"`, 1)
		if _, err := parseSeries([]byte(blocked), testURL); !errors.Is(err, ErrBlocked) {
			t.Fatalf("tag %s not blocked: %v", bad, err)
		}
	}
	if _, err := parseSeries([]byte(strings.Replace(body, `"genre":["Comedy","School Life"]`, `"genre":[]`, 1)), testURL); !errors.Is(err, ErrBlocked) {
		t.Fatal("unknown classification accepted")
	}
}

func TestReaderExtractsOnlyExactChapterPagesInDOMOrder(t *testing.T) {
	ch := Chapter{URL: testURL + "chapter-1.100/", Number: "1"}
	body := `<img src="https://rawkuma.net/wp-content/uploads/cover.jpg"><script>ad()</script>
<img src="https://cdn.rawkuma.com/images/s/safe-story/1/010.jpg"><img data-src="https://cdn.rawkuma.com/s/safe-story/chapter-1/002.jpg">
<img src="https://cdn.rawkuma.com/s/safe-story/1/001.jpg"><img src="https://cdn.rawkuma.com/s/safe-story/1/001.jpg">
<img src="https://cdn.rawkuma.com/s/other/1/003.jpg"><img src="https://cdn.rawkuma.com/s/safe-story/2/003.jpg"><img src="https://cdn.rawkuma.com/s/safe-story/1/ad.gif"><img src="https://evil.example/tracker.jpg">`
	got, err := parsePages([]byte(body), ch)
	if err != nil || len(got) != 3 || !strings.HasSuffix(got[0], "/010.jpg") || !strings.HasSuffix(got[2], "/001.jpg") {
		t.Fatalf("pages: %v %v", got, err)
	}
	if _, err = parsePages([]byte(`<html>challenge</html>`), ch); err == nil {
		t.Fatal("empty/challenge page accepted")
	}
}

func TestURLAndRobotsGuards(t *testing.T) {
	for _, raw := range []string{"http://rawkuma.net/manga/test/", "https://rawkuma.net.evil.example/manga/test/", "https://user:pass@rawkuma.net/manga/test/", "https://rawkuma.net:443/manga/test/", "https://rawkuma.net/manga/test/?redirect=http://localhost", "https://rawkuma.net/manga/test/#x", "https://rawkuma.net/manga/test/../other/", "https://localhost/manga/test/"} {
		if ValidSeriesURL(raw) {
			t.Fatal("invalid source", raw)
		}
	}
	if !ValidSeriesURL(testURL) || allowedURL("https://rawkuma.net/wp-admin/admin-ajax.php") || allowedURL("https://cdn.rawkuma.com@127.0.0.1/page.jpg") {
		t.Fatal("URL rules")
	}
	rules := "User-agent: *\nDisallow: /\nUser-agent: NihongoMasterManga\nAllow: /manga/*\nDisallow: /manga/private/\n"
	if !robotsAllow(rules, "/manga/safe/") || robotsAllow(rules, "/manga/private/") {
		t.Fatal("specific robots group ignored")
	}
	if robotsAllow("User-agent: *\nDisallow: /*.jpg$", "/a.jpg") || !robotsAllow("User-agent: *\nDisallow:\n", "/manga/a/") {
		t.Fatal("robots wildcard or empty rule")
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestForbiddenAndRateLimitAreNotRetriedOrBypassed(t *testing.T) {
	for _, status := range []int{403, 429} {
		calls := 0
		f := NewFetcher()
		f.gap = 0
		f.client = &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			calls++
			code, body := 200, "User-agent: *\nAllow: /\n"
			if r.URL.Path != "/robots.txt" {
				code = status
				body = "blocked"
			}
			return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		})}
		if _, err := f.Series(context.Background(), testURL); err == nil || calls != 2 {
			t.Fatalf("status %d retried or accepted: %d %v", status, calls, err)
		}
	}
}

func TestCDNMissingRobotsRedirectDoesNotVisitLegacyHomepage(t *testing.T) {
	f := NewFetcher()
	f.gap = 0
	calls := 0
	f.client.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "cdn.rawkuma.com" {
			t.Fatal("visited legacy homepage", r.URL)
		}
		code, body := 200, "image"
		headers := make(http.Header)
		if r.URL.Path == "/robots.txt" {
			code = 302
			headers.Set("Location", "https://rawkuma.com/")
		}
		return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: headers, Request: r}, nil
	})
	for i := 0; i < 2; i++ {
		if _, err := f.Image(context.Background(), "https://cdn.rawkuma.com/r/test/1/001.jpg", testURL); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 3 {
		t.Fatal("CDN robots not reused", calls)
	}
}

func imageBytes() []byte {
	var b bytes.Buffer
	_ = png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 80, 120)))
	return b.Bytes()
}
func TestImagesRejectHTMLAndOversizedDimensions(t *testing.T) {
	if _, _, err := inspectImage([]byte(`<html>cloudflare</html>`)); err == nil {
		t.Fatal("HTML image accepted")
	}
	if _, _, err := inspectImage(imageBytes()); err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	_ = png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	if _, _, err := inspectImage(b.Bytes()); err == nil {
		t.Fatal("tracking pixel accepted")
	}
}

type fakeFetcher struct {
	seriesCalls, pagesCalls, imageCalls atomic.Int32
	failImage                           atomic.Bool
	transient                           atomic.Bool
	adult                               atomic.Bool
}

func (f *fakeFetcher) Series(_ context.Context, raw string) (SeriesImport, error) {
	f.seriesCalls.Add(1)
	genre := "Comedy"
	if f.adult.Load() {
		genre = "Ecchi"
	}
	chs := []Chapter{}
	for _, num := range []string{"1", "2"} {
		url := raw + "chapter-" + num + ".100/"
		p := float64(len(chs) + 1)
		chs = append(chs, Chapter{ID: ID(url), SeriesID: ID(raw), URL: url, Number: num, Title: "Chapter " + num, Position: p})
	}
	return SeriesImport{Series: Series{URL: raw, Title: "Test manga", Genres: []string{genre}}, Cover: "https://rawkuma.net/wp-content/uploads/test.jpg", Chapters: chs}, nil
}
func (f *fakeFetcher) Pages(_ context.Context, ch Chapter) ([]string, error) {
	f.pagesCalls.Add(1)
	return []string{"https://cdn.rawkuma.com/s/safe-story/" + ch.Number + "/001.jpg", "https://cdn.rawkuma.com/s/safe-story/" + ch.Number + "/002.jpg"}, nil
}
func (f *fakeFetcher) Image(_ context.Context, raw, _ string) ([]byte, error) {
	f.imageCalls.Add(1)
	if f.failImage.Load() && strings.HasSuffix(raw, "/002.jpg") {
		if f.transient.Load() {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, HTTPError{Status: 403}
	}
	return imageBytes(), nil
}

type fakeStore struct{ calls atomic.Int32 }

func (s *fakeStore) Put(_ context.Context, body []byte) (Asset, error) {
	n := s.calls.Add(1)
	return Asset{URL: fmtURL(n), ObjectKey: "manga/page.jpg", Bytes: len(body), Width: 80, Height: 120}, nil
}
func fmtURL(n int32) string { return "https://our-cos.example/manga/" + strconv.Itoa(int(n)) + ".png" }

func testRepository(t *testing.T) (*Repository, *fakeFetcher, *fakeStore, *time.Time) {
	t.Helper()
	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	t.Cleanup(func() { _ = conn.Close() })
	store := &fakeStore{}
	r, err := New(db, store)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Model(&Series{}).Where("1 = 1").Update("enabled", false).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	r.now = func() time.Time { return now }
	s := Series{ID: ID(testURL), URL: testURL, Enabled: true, Reviewed: true, Language: "ja", Genres: []string{}, Status: "pending", NextFetchAt: time.Unix(1, 0).UTC()}
	if err = db.Create(&s).Error; err != nil {
		t.Fatal(err)
	}
	return r, &fakeFetcher{}, store, &now
}

func TestMariaDBWeeklyCacheSurvivesReadsAndRestarts(t *testing.T) {
	r, f, storage, now := testRepository(t)
	ctx := context.Background()
	if err := r.Sync(ctx, f); err != nil {
		t.Fatal(err)
	}
	if f.seriesCalls.Load() != 1 || f.pagesCalls.Load() != 1 || storage.calls.Load() != 3 {
		t.Fatal("unexpected initial fetch count")
	}
	detail, err := r.Detail(ctx, ID(testURL))
	if err != nil || len(detail.Chapters) != 2 {
		t.Fatalf("detail: %+v %v", detail, err)
	}
	if catalog, err := r.List(ctx, "test", "Comedy", 1, 20); err != nil || catalog.Total != 1 {
		t.Fatal("case-insensitive cached search failed", err, catalog.Total)
	}
	for i := 0; i < 20; i++ {
		_, _ = r.List(ctx, "Test", "Comedy", 1, 20)
		_, _ = r.Detail(ctx, ID(testURL))
		ch, e := r.Chapter(ctx, detail.Chapters[0].ID)
		if e != nil || ch.Status != "ready" || len(ch.Pages) != 2 {
			t.Fatal("cache incomplete", e)
		}
		if _, e = r.Prepare(ctx, ch.ID); e != nil {
			t.Fatal(e)
		}
		body, _ := json.Marshal(ch)
		if strings.Contains(string(body), "cdn.rawkuma.com") {
			t.Fatal("upstream image leaked")
		}
	}
	if f.seriesCalls.Load() != 1 || f.pagesCalls.Load() != 1 || f.imageCalls.Load() != 3 {
		t.Fatal("reads re-fetched upstream")
	}
	// Re-opening the repo must not reset persisted weekly deadlines.
	r2, err := New(r.db, storage)
	if err != nil {
		t.Fatal(err)
	}
	r2.now = r.now
	*now = now.Add(6 * time.Minute)
	if err = r2.Sync(ctx, f); err != nil {
		t.Fatal(err)
	}
	if f.seriesCalls.Load() != 1 || f.pagesCalls.Load() != 1 {
		t.Fatal("restart bypassed cooldown")
	}
	*now = now.Add(RefreshInterval)
	if err = r2.Sync(ctx, f); err != nil {
		t.Fatal(err)
	}
	if f.seriesCalls.Load() != 2 || f.pagesCalls.Load() != 1 || f.imageCalls.Load() != 3 {
		t.Fatal("weekly detail refresh re-downloaded cached assets")
	}
}

func TestMariaDBQueueIsSharedIdempotentAndLeasedAcrossWorkers(t *testing.T) {
	r, f, _, now := testRepository(t)
	ctx := context.Background()
	if err := r.Sync(ctx, f); err != nil {
		t.Fatal(err)
	}
	detail, _ := r.Detail(ctx, ID(testURL))
	second := detail.Chapters[1].ID
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := r.Prepare(ctx, second); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	var count int64
	r.db.Model(&Chapter{}).Where("status = ?", "queued").Count(&count)
	if count != 1 {
		t.Fatal("duplicate queue", count)
	}
	*now = now.Add(6 * time.Minute)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.Sync(ctx, f); err != nil && !errors.Is(err, ErrBusy) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if f.pagesCalls.Load() != 2 {
		t.Fatal("two workers downloaded the same chapter", f.pagesCalls.Load())
	}
	ch, err := r.Chapter(ctx, second)
	if err != nil || !ready(ch) {
		t.Fatal("queued chapter not durable", err)
	}
}
func ready(ch Chapter) bool { return ch.Status == "ready" && len(ch.Pages) == 2 }

func TestMariaDBPartialFailureIsNotPublishedAndDoesNotHammerSource(t *testing.T) {
	r, f, _, now := testRepository(t)
	ctx := context.Background()
	f.failImage.Store(true)
	if err := r.Sync(ctx, f); err == nil {
		t.Fatal("expected failed image")
	}
	detail, err := r.Detail(ctx, ID(testURL))
	if err != nil {
		t.Fatal(err)
	}
	ch, err := r.Chapter(ctx, detail.Chapters[0].ID)
	if err != nil || ch.Status != "failed" || len(ch.Pages) != 0 {
		t.Fatal("partial chapter published", ch, err)
	}
	calls := f.imageCalls.Load()
	*now = now.Add(6 * time.Minute)
	if _, err = r.Prepare(ctx, ch.ID); err != nil {
		t.Fatal(err)
	}
	if err = r.Sync(ctx, f); err != nil {
		t.Fatal(err)
	}
	if f.imageCalls.Load() != calls || f.pagesCalls.Load() != 1 {
		t.Fatal("failed import retried before seven days")
	}
	f.failImage.Store(false)
	*now = now.Add(RefreshInterval)
	if _, err = r.Prepare(ctx, ch.ID); err != nil {
		t.Fatal(err)
	}
	if err = r.Sync(ctx, f); err != nil {
		t.Fatal(err)
	}
	ch, err = r.Chapter(ctx, ch.ID)
	if err != nil || !ready(ch) {
		t.Fatal("failed chapter not recoverable", err)
	}
	if f.imageCalls.Load() != calls+1 {
		t.Fatal("already cached pages were downloaded again")
	}
}

func TestMariaDBEditorialGateAndAdultMetadataHideAllChapters(t *testing.T) {
	r, f, _, now := testRepository(t)
	ctx := context.Background()
	if err := r.Sync(ctx, f); err != nil {
		t.Fatal(err)
	}
	detail, _ := r.Detail(ctx, ID(testURL))
	if err := r.Submit(ctx, Submission{URL: testURL, Enabled: false}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Chapter(ctx, detail.Chapters[0].ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("hidden chapter accessible")
	}
	if err := r.Submit(ctx, Submission{URL: testURL, Enabled: true}); !errors.Is(err, ErrInvalid) {
		t.Fatal("unreviewed Japanese/content accepted")
	}
	if err := r.Submit(ctx, Submission{URL: testURL, Enabled: true, JapaneseVerified: true, NonAdultReviewed: true}); err != nil {
		t.Fatal(err)
	}
	f.adult.Store(true)
	*now = now.Add(RefreshInterval)
	if err := r.Sync(ctx, f); err == nil {
		t.Fatal("adult import accepted")
	}
	if _, err := r.Detail(ctx, ID(testURL)); !errors.Is(err, ErrNotFound) {
		t.Fatal("adult metadata remained visible")
	}
	items, err := r.List(ctx, "", "", 1, 20)
	if err != nil || items.Total != 0 {
		t.Fatal("adult in public catalog")
	}
}

func TestMariaDBTransientPageFailureResumesWithoutRefetchingChapterDetail(t *testing.T) {
	r, f, _, now := testRepository(t)
	ctx := context.Background()
	f.failImage.Store(true)
	f.transient.Store(true)
	if err := r.Sync(ctx, f); err == nil {
		t.Fatal("expected transient asset failure")
	}
	detail, _ := r.Detail(ctx, ID(testURL))
	ch, err := r.Chapter(ctx, detail.Chapters[0].ID)
	if err != nil || len(ch.SourcePages) != 2 || !ch.NextAttemptAt.Equal(now.Add(5*time.Minute)) {
		t.Fatal("manifest/backoff not durable", err, ch.NextAttemptAt)
	}
	f.failImage.Store(false)
	*now = now.Add(6 * time.Minute)
	if err := r.Sync(ctx, f); err != nil {
		t.Fatal(err)
	}
	ch, err = r.Chapter(ctx, ch.ID)
	if err != nil || !ready(ch) {
		t.Fatal("missing assets were not resumed", err)
	}
	if f.seriesCalls.Load() != 1 || f.pagesCalls.Load() != 1 || f.imageCalls.Load() != 4 {
		t.Fatal("resume repeated a detail or cached image fetch", f.pagesCalls.Load(), f.imageCalls.Load())
	}
}

func TestMariaDBMissingCOSNeverFetchesAndQueueIsBounded(t *testing.T) {
	r, f, _, _ := testRepository(t)
	ctx := context.Background()
	r.images = nil
	if err := r.Sync(ctx, f); err == nil || f.seriesCalls.Load() != 0 {
		t.Fatal("source fetched without storage")
	}
	if _, err := r.Prepare(ctx, "missing"); err == nil {
		t.Fatal("queue accepted without COS")
	}
	r.images = &fakeStore{}
	if err := r.db.Model(&Series{}).Where("id = ?", ID(testURL)).Update("status", "ready").Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 65; i++ {
		raw := testURL + "chapter-" + strconv.Itoa(i+1) + ".100/"
		status := "queued"
		if i == 64 {
			status = "available"
		}
		ch := Chapter{ID: ID(raw), SeriesID: ID(testURL), URL: raw, Status: status, Pages: []Page{}, NextAttemptAt: time.Unix(1, 0).UTC()}
		if err := r.db.Create(&ch).Error; err != nil {
			t.Fatal(err)
		}
		if i == 64 {
			if _, err := r.Prepare(ctx, ch.ID); !errors.Is(err, ErrBusy) {
				t.Fatal("unbounded chapter queue", err)
			}
		}
	}
}
