package news

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/html"
	"gorm.io/gorm"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/testdb"
)

func TestCalendarWeekdaysAndTimezones(t *testing.T) {
	boundary, _ := time.Parse(time.RFC3339, "2026-10-13T15:00:00Z")
	if sourceDate(boundary) != "2026-10-14" {
		t.Fatal("publisher date shifted to the previous day in WIB")
	}
	s := Source{Days: []int{2, 5}, Hour: 18, Timezone: "Asia/Jakarta"}
	for _, test := range []struct{ after, want string }{
		{"2026-10-19T12:00:00Z", "2026-10-20T11:00:00Z"},
		{"2026-10-20T11:00:00Z", "2026-10-23T11:00:00Z"},
		{"2026-10-24T11:00:00Z", "2026-10-27T11:00:00Z"},
	} {
		after, _ := time.Parse(time.RFC3339, test.after)
		next, err := NextScheduled(s, after)
		if err != nil || next.Format(time.RFC3339) != test.want {
			t.Fatalf("%s: %v %v", test.after, next, err)
		}
	}
	for _, bad := range []Source{{Days: []int{7}, Timezone: "UTC"}, {Days: []int{1, 1}, Timezone: "UTC"}, {Days: []int{1}, Hour: 24, Timezone: "UTC"}, {Days: []int{1}, Timezone: "Mars/Unknown"}} {
		if _, err := NextScheduled(bad, time.Now()); err == nil {
			t.Fatal("invalid schedule accepted", bad)
		}
	}
}
func TestTokensUseUTF16AndLemma(t *testing.T) {
	text := "😀東京で食べました。"
	tokens := tokenize(text)
	if len(tokens) == 0 {
		t.Fatal("no tokens")
	}
	foundCity, foundVerb := false, false
	for _, token := range tokens {
		if utf16slice(text, token.Start, token.End) != token.Surface {
			t.Fatalf("invalid offset %+v", token)
		}
		if token.Surface == "東京" {
			foundCity = token.Start == 2 && token.Reading == "とうきょう"
		}
		if token.Surface == "食べ" {
			foundVerb = token.Base == "食べる"
		}
	}
	if !foundCity || !foundVerb {
		t.Fatalf("reading or lemma missing: %+v", tokens)
	}
}
func TestSourceRubyAndMediaRemoval(t *testing.T) {
	doc, _ := html.Parse(strings.NewReader(`<div><p>😀<ruby>東京<rt>とうきょう</rt></ruby>で<ruby>食<rt>た</rt></ruby>べる。</p><div class="wp-caption"><img src="https://publisher/image.jpg"><p>PHOTO CAPTION</p></div><script>BAD</script><p><a href="https://example.com/facts">根拠</a>です。</p><div class="gv-rss-footer"><p>CREDITS</p></div></div>`))
	paragraphs, refs := parseParagraphs(doc, "https://jp.globalvoices.org/2026/10/09/123/")
	if len(paragraphs) != 2 || paragraphs[0].Text != "😀東京で食べる。" || len(refs) != 1 {
		t.Fatalf("bad text extraction: %+v %+v", paragraphs, refs)
	}
	found := false
	for _, token := range paragraphs[0].Tokens {
		if token.Surface == "東京" {
			found = token.ReadingSource == "source" && token.Start == 2 && token.End == 4
		}
	}
	if !found {
		t.Fatal("source ruby lost", paragraphs)
	}
}
func TestFeedsDatesCreditsAndPartnerExclusion(t *testing.T) {
	const body = `<rss xmlns:content="http://purl.org/rss/1.0/modules/content/"><channel><item><title>日本のニュース</title><link>https://jp.globalvoices.org/2026/10/09/123/</link><pubDate>Fri, 09 Oct 2026 11:00:00 +0000</pubDate><content:encoded><![CDATA[<p>東京で新しい文化の催しが始まりました。多くの人が来ています。</p><div class="gv-rss-footer"><div class="text-credits-section">記者 <a class="user-link" href="https://globalvoices.org/author/example/">Author</a></div><div class="text-credits-section">翻訳 <a class="user-link" href="https://jp.globalvoices.org/author/example/">Translator</a></div></div>]]></content:encoded></item><item><title>Partner</title><link>https://jp.globalvoices.org/2026/10/09/456/</link><pubDate>Fri, 09 Oct 2026 11:00:00 +0000</pubDate><content:encoded><![CDATA[<p>コンテンツ共有契約に基づき転載する。</p>]]></content:encoded></item></channel></rss>`
	since := time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC)
	articles, skipped, err := parseGlobalVoices([]byte(body), since)
	if err != nil || len(articles) != 1 || skipped != 1 || len(articles[0].Credits) != 2 || articles[0].Credits[1].Role != "translator" {
		t.Fatalf("feed: %+v %d %v", articles, skipped, err)
	}
	const index = `<ul><li class="cat_topics_icon_none"><a href="https://www.city.osaka.lg.jp/shimin/page/0000123456.html">暮らしの案内</a><span class="date">2026年10月14日</span></li><li class="cat_topics_icon_none"><a href="https://www.city.osaka.lg.jp/a/cmsfiles/a.pdf">PDF</a>2026年10月14日</li></ul>`
	links, err := parseOsakaIndex([]byte(index), since)
	if err != nil || len(links) != 1 || links[0].Date.Hour() != 15 {
		t.Fatalf("Osaka %v %v", links, err)
	}
	for _, u := range []string{"http://jp.globalvoices.org/2026/10/09/123/", "https://jp.globalvoices.org.evil.com/2026/10/09/123/", "https://jp.globalvoices.org:443/2026/10/09/123/", "https://jp.globalvoices.org/2026/10/09/123/?url=http://localhost"} {
		if allowedArticleURL("global-voices", u) {
			t.Fatal("unsafe source URL", u)
		}
	}
}

type stubFetcher struct {
	result           FetchResult
	err              error
	calls            atomic.Int32
	started, release chan struct{}
}

func (f *stubFetcher) Fetch(ctx context.Context, s Source, since time.Time) (FetchResult, error) {
	f.calls.Add(1)
	if f.started != nil {
		close(f.started)
		select {
		case <-f.release:
		case <-ctx.Done():
			return FetchResult{}, ctx.Err()
		}
	}
	return f.result, f.err
}
func setup(t *testing.T) (*Repository, *gorm.DB, time.Time) {
	t.Helper()
	db, err := database.Open(testdb.DSN(t))
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	t.Cleanup(func() { conn.Close() })
	r, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	r.now = func() time.Time { return now }
	return r, db, now
}
func sampleArticle(now time.Time) Article {
	return Article{URL: "https://jp.globalvoices.org/2026/10/09/123/", Title: "日本のニュース", Paragraphs: []Paragraph{{Text: "東京で新しい文化の催しが始まりました。多くの人が来ています。", Tokens: []Token{{Start: 0, End: 2, Surface: "東京", Base: "東京", Reading: "とうきょう", ReadingSource: "automatic"}}}}, TitleTokens: []Token{}, Credits: []Credit{{Name: "Writer", Role: "author"}}, References: []Credit{}, PublishedAt: now.Add(-24 * time.Hour), License: "CC BY 3.0", LicenseURL: "https://creativecommons.org/licenses/by/3.0/"}
}
func TestPersistentSyncDedupEditsFailureAndDelete(t *testing.T) {
	r, db, now := setup(t)
	ctx := context.Background()
	f := &stubFetcher{result: FetchResult{Articles: []Article{sampleArticle(now)}, ETag: "first"}}
	if n, err := r.Sync(ctx, f, "global-voices", false); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	restarted, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	restarted.now = r.now
	if _, err := restarted.Sync(ctx, f, "global-voices", false); !errors.Is(err, ErrBusy) {
		t.Fatal("restart retried source", err)
	}
	page, err := restarted.List(ctx, 1, 20, "", false)
	if err != nil || page.Total != 1 {
		t.Fatal(page, err)
	}
	id := page.Items[0].ID
	if err := r.Edit(ctx, id, ArticleEdit{Revision: "stale", Readings: []ReadingEdit{{Paragraph: 0, Token: 0, Reading: "wrong"}}}); !errors.Is(err, ErrConflict) {
		t.Fatal("stale reading edit accepted", err)
	}
	if err := r.Edit(ctx, id, ArticleEdit{Revision: page.Items[0].ContentHash, Hidden: true, Readings: []ReadingEdit{{Paragraph: 0, Token: 0, Reading: "とーきょー"}}}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Minute)
	r.now = func() time.Time { return now }
	if _, err := r.Sync(ctx, f, "global-voices", true); err != nil {
		t.Fatal(err)
	}
	a, err := r.Detail(ctx, id, true)
	if err != nil || !a.Hidden || a.Paragraphs[0].Tokens[0].Reading != "とーきょー" {
		t.Fatal("lost editorial state", a, err)
	}
	public, _ := r.List(ctx, 1, 20, "", false)
	if public.Total != 0 {
		t.Fatal("hidden article exposed")
	}
	f.err = errors.New("upstream timeout")
	now = now.Add(2 * time.Minute)
	if _, err := r.Sync(ctx, f, "global-voices", true); err == nil {
		t.Fatal("failure accepted")
	}
	if _, err := r.Detail(ctx, id, true); err != nil {
		t.Fatal("failed refresh destroyed data", err)
	}
	var source Source
	db.First(&source, "id = ?", "global-voices")
	if source.LastError == "" || source.LeaseOwner != "" {
		t.Fatal("missing durable failure status", source)
	}
	if err := r.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Minute)
	f.err = nil
	if _, err := r.Sync(ctx, f, "global-voices", true); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Detail(ctx, id, true); !errors.Is(err, ErrNotFound) {
		t.Fatal("deleted article returned", err)
	}
	var count int64
	db.Model(&Article{}).Count(&count)
	if count != 1 {
		t.Fatal("tombstone not deduplicated", count)
	}
}
func TestConcurrentWorkerAndAtomicBatch(t *testing.T) {
	r, db, now := setup(t)
	ctx := context.Background()
	f := &stubFetcher{result: FetchResult{Articles: []Article{sampleArticle(now)}}, started: make(chan struct{}), release: make(chan struct{})}
	done := make(chan error, 1)
	go func() { _, err := r.Sync(ctx, f, "global-voices", false); done <- err }()
	<-f.started
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := r.Sync(ctx, f, "global-voices", true); !errors.Is(err, ErrBusy) {
				t.Error("duplicate claim", err)
			}
		}()
	}
	group.Wait()
	close(f.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if f.calls.Load() != 1 {
		t.Fatal("multiple upstream fetches")
	}
	now = now.Add(3 * time.Minute)
	r.now = func() time.Time { return now }
	bad := sampleArticle(now)
	bad.URL = "https://localhost/invalid"
	good := sampleArticle(now)
	good.Title = "Changed"
	f = &stubFetcher{result: FetchResult{Articles: []Article{good, bad}}}
	if _, err := r.Sync(ctx, f, "global-voices", true); err == nil {
		t.Fatal("invalid batch accepted")
	}
	var a Article
	db.First(&a)
	if a.Title != "日本のニュース" {
		t.Fatal("partial transaction persisted", a.Title)
	}
	sources, _ := r.Sources(ctx)
	if sources[0].LastStartedAt == nil {
		t.Fatal("deadline lost")
	}
}
func TestLiveSources(t *testing.T) {
	if os.Getenv("NEWS_LIVE_TEST") != "1" {
		t.Skip("opt-in publisher requests")
	}
	f := NewFetcher()
	for _, s := range []Source{{ID: "global-voices"}, {ID: "osaka"}} {
		result, err := f.Fetch(context.Background(), s, time.Now().AddDate(0, -6, 0))
		if err != nil {
			t.Fatal(s.ID, err)
		}
		if len(result.Articles) == 0 {
			t.Fatal(s.ID, "no eligible articles")
		}
		for _, a := range result.Articles {
			if !allowedArticleURL(s.ID, a.URL) || len(a.Paragraphs) == 0 || len(a.Credits) == 0 {
				t.Fatal("incomplete article", a.URL)
			}
		}
		t.Logf("%s: %d articles; skipped %d; first article %d paragraphs", s.ID, len(result.Articles), result.Skipped, len(result.Articles[0].Paragraphs))
	}
}
