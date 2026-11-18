package news

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fixtureTransport serves local files in place of publisher URLs so fetch
// tests never reach the network.
type fixtureTransport map[string]string

func (f fixtureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	name, ok := f[req.URL.String()]
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}, Request: req}, nil
	}
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(data)), Header: http.Header{"Etag": {`"kantei-1"`}}, Request: req}, nil
}

var kanteiSince = time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC)

func TestKanteiFeedFiltersAndOrders(t *testing.T) {
	data, err := os.ReadFile("testdata/kantei_feed.xml")
	if err != nil {
		t.Fatal(err)
	}
	links, err := parseKanteiFeed(data, kanteiSince)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"https://www.kantei.go.jp/jp/105/actions/202611/18kodomo.html",
		"https://www.kantei.go.jp/jp/105/actions/202611/17bousai.html",
		"https://www.kantei.go.jp/jp/tyoukanpress/202611/17_a.html",
	}
	if len(links) != len(want) {
		t.Fatalf("links: %+v", links)
	}
	for i, link := range links {
		if link.URL != want[i] {
			t.Fatalf("link %d = %s, want %s", i, link.URL, want[i])
		}
	}
	// Kantei omits the weekday: "18 Nov 2026 19:52:00 +0900".
	if got := links[0].Date.Format(time.RFC3339); got != "2026-11-18T10:52:00Z" {
		t.Fatalf("date without weekday parsed as %s", got)
	}
	if links[0].Title != "総理は子ども食堂を訪問しました" {
		t.Fatalf("title %q", links[0].Title)
	}
	if _, err := parseKanteiFeed([]byte(`<rss><channel></channel></rss>`), kanteiSince); err == nil {
		t.Fatal("empty feed accepted")
	}
}

func TestKanteiArticle(t *testing.T) {
	data, err := os.ReadFile("testdata/kantei_article.html")
	if err != nil {
		t.Fatal(err)
	}
	link := articleLinkFixture()
	a, err := parseKanteiArticle(data, link)
	if err != nil {
		t.Fatal(err)
	}
	if a.Title != link.Title || a.URL != link.URL || !a.PublishedAt.Equal(link.Date) || a.DateType != "published" {
		t.Fatalf("metadata: %+v", a)
	}
	if len(a.Paragraphs) != 2 {
		t.Fatalf("title heading, carousel caption or script leaked into text: %+v", a.Paragraphs)
	}
	if !strings.HasPrefix(a.Paragraphs[0].Text, "令和") || a.Summary == "" || !strings.HasPrefix(a.Summary, "令和") {
		t.Fatalf("ideographic indent not trimmed: %q / %q", a.Paragraphs[0].Text, a.Summary)
	}
	ruby := false
	for _, token := range a.Paragraphs[1].Tokens {
		if token.Surface == "昼食" {
			ruby = token.Reading == "ちゅうしょく" && token.ReadingSource == "source"
		}
	}
	if !ruby {
		t.Fatal("source furigana lost", a.Paragraphs[1].Tokens)
	}
	if len(a.Credits) != 1 || a.Credits[0].Name != "首相官邸ホームページ" || a.Credits[0].URL != link.URL || a.Credits[0].Role != "publisher" {
		t.Fatalf("attribution: %+v", a.Credits)
	}
	if !strings.Contains(a.License, "PDL1.0") || a.LicenseURL != kanteiLicenseURL || a.Changes == "" {
		t.Fatalf("license: %q %q %q", a.License, a.LicenseURL, a.Changes)
	}
	if len(a.imageCandidates) != 1 {
		t.Fatalf("image candidates: %+v", a.imageCandidates)
	}
	img := a.imageCandidates[0]
	if img.URL != "https://www.kantei.go.jp/contents/topics/70000_ext_1_0.jpg" || img.Unlicensed || img.License != "PDL1.0" || img.SourceURL != link.URL {
		t.Fatalf("image: %+v", img)
	}
	if err := validateArticle(a, Source{ID: "kantei"}, link.Date.Add(time.Hour)); err != nil {
		t.Fatal("parsed article rejected by import validation:", err)
	}

	noText, _ := os.ReadFile("testdata/kantei_no_text.html")
	if _, err := parseKanteiArticle(noText, link); err == nil {
		t.Fatal("page without article text accepted")
	}
}

func TestKanteiURLs(t *testing.T) {
	for _, u := range []string{
		"https://www.kantei.go.jp/jp/105/actions/202611/18kodomo.html",
		"https://www.kantei.go.jp/jp/tyoukanpress/202611/17_a.html",
		"https://www.kantei.go.jp/jp/pages/20261112choukan_toriinfuruenza.html",
	} {
		if !allowedArticleURL("kantei", u) {
			t.Fatal("rejected", u)
		}
	}
	for _, u := range []string{
		"http://www.kantei.go.jp/jp/105/actions/202611/18kodomo.html",
		"https://www.kantei.go.jp.evil.example/jp/105/actions/202611/x.html",
		"https://www.kantei.go.jp:8443/jp/105/actions/202611/x.html",
		"https://www.kantei.go.jp/jp/singi/shiryou.pdf",
		"https://www.kantei.go.jp/jp/105/../../etc/passwd.html",
		"https://www.kantei.go.jp/jp/105/actions/x.html?redirect=https://localhost",
		"https://www.kantei.go.jp/en/105/actions/202611/x.html",
	} {
		if allowedArticleURL("kantei", u) {
			t.Fatal("unsafe Kantei URL accepted", u)
		}
	}
	if !KnownSource("kantei") || !KnownSource("osaka") || !KnownSource("global-voices") || KnownSource("nhk") {
		t.Fatal("source registry")
	}
}

func TestKanteiFetchOffline(t *testing.T) {
	f := NewFetcher()
	f.spacing = 0
	f.client.Transport = fixtureTransport{
		kanteiFeed: "kantei_feed.xml",
		"https://www.kantei.go.jp/jp/105/actions/202611/18kodomo.html": "kantei_article.html",
		"https://www.kantei.go.jp/jp/105/actions/202611/17bousai.html": "kantei_article.html",
		"https://www.kantei.go.jp/jp/tyoukanpress/202611/17_a.html":    "kantei_no_text.html",
	}
	result, err := f.Fetch(context.Background(), Source{ID: "kantei"}, kanteiSince)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Articles) != 2 || result.Skipped != 1 || result.ETag != `"kantei-1"` {
		t.Fatalf("articles %d skipped %d etag %q", len(result.Articles), result.Skipped, result.ETag)
	}
	if result.Articles[0].Title != "総理は子ども食堂を訪問しました" || result.Articles[1].Title != "総理は防災訓練に参加しました" {
		t.Fatalf("titles: %q, %q", result.Articles[0].Title, result.Articles[1].Title)
	}
}

func TestDedupKeyIgnoresWidthPunctuationAndSpacing(t *testing.T) {
	a := dedupKey("総理はＡＳＥＡＮ首脳会議に出席しました（２）")
	b := dedupKey("総理は ASEAN 首脳会議に出席しました(2)！")
	if a == "" || a != b {
		t.Fatalf("equivalent titles differ: %q %q", a, b)
	}
	if a == dedupKey("総理はＡＳＥＡＮ首脳会議に出席しました（３）") {
		t.Fatal("different parts of a series collapsed")
	}
	if dedupKey("お知らせ") != "" {
		t.Fatal("generic short title used as a story key")
	}
}

func TestSevenDaySchedule(t *testing.T) {
	daily := Source{Days: []int{0, 1, 2, 3, 4, 5, 6}, Hour: 19, Timezone: "Asia/Jakarta"}
	after := time.Date(2026, 11, 18, 13, 0, 0, 0, time.UTC) // 20:00 WIB, past today's run.
	next, err := NextScheduled(daily, after)
	if err != nil || next.Format(time.RFC3339) != "2026-11-19T12:00:00Z" {
		t.Fatal(next, err)
	}
	daily.Days = []int{0, 1, 2, 3, 4, 5, 6, 0}
	if _, err := NextScheduled(daily, after); err == nil {
		t.Fatal("eight schedule days accepted")
	}
}

func TestCrossSourceDuplicateIsSkipped(t *testing.T) {
	r, db, now := setup(t)
	ctx := context.Background()
	first := sampleArticle(now)
	first.Title = "東京で新しい文化の催しが始まりました"
	if n, err := r.Sync(ctx, &stubFetcher{result: FetchResult{Articles: []Article{first}}}, "global-voices", false); err != nil || n != 1 {
		t.Fatal(n, err)
	}
	same := sampleArticle(now)
	same.URL = "https://www.kantei.go.jp/jp/105/actions/202611/18bunka.html"
	same.Title = "東京で、新しい文化の催しが始まりました！"
	same.PublishedAt = now.Add(-2 * time.Hour)
	other := sampleArticle(now)
	other.URL = "https://www.kantei.go.jp/jp/105/actions/202611/18kodomo.html"
	other.Title = "総理は子ども食堂を訪問しました"
	n, err := r.Sync(ctx, &stubFetcher{result: FetchResult{Articles: []Article{same, other}}}, "kantei", false)
	if err != nil || n != 1 {
		t.Fatal("expected only the new story to be imported", n, err)
	}
	page, err := r.List(ctx, 1, 20, "", false)
	if err != nil || page.Total != 2 {
		t.Fatal("duplicate story listed twice", page.Total, err)
	}
	var source Source
	db.First(&source, "id = ?", "kantei")
	if source.LastImported != 1 || source.LastSkipped != 1 {
		t.Fatalf("status imported %d skipped %d", source.LastImported, source.LastSkipped)
	}
	// Re-syncing the surviving story updates it in place instead of skipping.
	now = now.Add(2 * time.Minute)
	r.now = func() time.Time { return now }
	if n, err := r.Sync(ctx, &stubFetcher{result: FetchResult{Articles: []Article{other}}}, "kantei", true); err != nil || n != 1 {
		t.Fatal("refresh of stored article", n, err)
	}
}

func articleLinkFixture() articleLink {
	return articleLink{URL: "https://www.kantei.go.jp/jp/105/actions/202611/18kodomo.html", Title: "総理は子ども食堂を訪問しました", Date: time.Date(2026, 11, 18, 10, 52, 0, 0, time.UTC)}
}
