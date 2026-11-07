package news

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/html"
)

const globalFeed = "https://jp.globalvoices.org/feed/"
const osakaIndex = "https://www.city.osaka.lg.jp/shisei/category/3054-1-2-21-3-4-0-0-0-0.html"

var gvPath = regexp.MustCompile(`^/\d{4}/\d{2}/\d{2}/\d+/$`)
var osakaPath = regexp.MustCompile(`^/[a-z_]+/page/\d+\.html$`)
var japaneseDate = regexp.MustCompile(`(\d{4})年\s*(\d{1,2})月\s*(\d{1,2})日`)

func allowedArticleURL(source, value string) bool {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	if source == "global-voices" {
		return u.Host == "jp.globalvoices.org" && gvPath.MatchString(u.Path)
	}
	return source == "osaka" && u.Host == "www.city.osaka.lg.jp" && osakaPath.MatchString(u.Path)
}

type HTTPFetcher struct {
	client         *http.Client
	spacing        time.Duration
	imageDirectory string
	store          ImageStore
	hasImage       func(ctx context.Context, articleURL string) bool
}

// SetImageStore sends imported images to object storage (COS) instead of the
// local image directory.
func (f *HTTPFetcher) SetImageStore(store ImageStore) { f.store = store }

// SetImageLookup lets the fetcher skip articles whose image is already stored.
func (f *HTTPFetcher) SetImageLookup(lookup func(ctx context.Context, articleURL string) bool) {
	f.hasImage = lookup
}

func NewFetcher(imageDirectory ...string) *HTTPFetcher {
	f := &HTTPFetcher{client: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 3 || req.URL.Scheme != "https" || req.URL.Host != via[0].URL.Host {
			return fmt.Errorf("news redirect outside source")
		}
		return nil
	}}, spacing: time.Second}
	if len(imageDirectory) > 0 {
		f.imageDirectory = imageDirectory[0]
	}
	return f
}
func (f *HTTPFetcher) get(ctx context.Context, value string, s Source, conditional bool) ([]byte, http.Header, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, value, nil)
	if err != nil {
		return nil, nil, false, err
	}
	req.Header.Set("User-Agent", "NihongoMasterNews/1.0 (+https://nihongo.codeverta.com)")
	if conditional {
		req.Header.Set("If-None-Match", s.ETag)
		req.Header.Set("If-Modified-Since", s.LastModified)
	}
	res, err := f.client.Do(req)
	if err != nil {
		return nil, nil, false, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotModified {
		return nil, res.Header, true, nil
	}
	if res.StatusCode != 200 {
		return nil, nil, false, fmt.Errorf("source HTTP %d", res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, (8<<20)+1))
	if err == nil && len(data) > 8<<20 {
		err = fmt.Errorf("source exceeds 8 MiB")
	}
	return data, res.Header, false, err
}
func (f *HTTPFetcher) Fetch(ctx context.Context, s Source, since time.Time) (FetchResult, error) {
	if err := prepareAnalyzer(); err != nil {
		return FetchResult{}, err
	}
	endpoint := globalFeed
	if s.ID == "osaka" {
		endpoint = osakaIndex
	} else if s.ID != "global-voices" {
		return FetchResult{}, ErrInvalid
	}
	// Osaka's index validators do not necessarily change when a linked article
	// is corrected. Revisit eligible HTML pages on each scheduled check.
	data, headers, unchanged, err := f.get(ctx, endpoint, s, s.ID == "global-voices")
	if err != nil {
		return FetchResult{}, err
	}
	result := FetchResult{Articles: []Article{}, ETag: headers.Get("ETag"), LastModified: headers.Get("Last-Modified"), NotModified: unchanged}
	if unchanged {
		return result, nil
	}
	if s.ID == "global-voices" {
		items, skipped, err := parseGlobalVoices(data, since)
		result.Articles = items
		result.Skipped = skipped
		if err == nil {
			f.importImages(ctx, result.Articles)
		}
		return result, err
	}
	links, err := parseOsakaIndex(data, since)
	if err != nil {
		return result, err
	}
	for _, link := range links {
		if f.spacing > 0 {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			case <-time.After(f.spacing):
			}
		}
		body, _, _, err := f.get(ctx, link.URL, s, false)
		if err != nil {
			return result, fmt.Errorf("Osaka article: %w", err)
		}
		a, err := parseOsakaArticle(body, link)
		if err != nil {
			return result, err
		}
		result.Articles = append(result.Articles, a)
	}
	f.importImages(ctx, result.Articles)
	return result, nil
}

type rss struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}
type rssItem struct {
	Title   string `xml:"title"`
	URL     string `xml:"link"`
	Date    string `xml:"pubDate"`
	Content string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	Media   []struct {
		URL string `xml:"url,attr"`
	} `xml:"http://search.yahoo.com/mrss/ content"`
	Thumbnails []struct {
		URL string `xml:"url,attr"`
	} `xml:"http://search.yahoo.com/mrss/ thumbnail"`
	Enclosures []struct {
		URL  string `xml:"url,attr"`
		Type string `xml:"type,attr"`
	} `xml:"enclosure"`
}

func parseGlobalVoices(data []byte, since time.Time) ([]Article, int, error) {
	var feed rss
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, 0, err
	}
	if len(feed.Channel.Items) == 0 {
		return nil, 0, fmt.Errorf("Global Voices feed structure changed")
	}
	items := []Article{}
	skipped := 0
	for _, item := range feed.Channel.Items {
		published, err := time.Parse(time.RFC1123Z, item.Date)
		if err != nil {
			published, err = time.Parse(time.RFC1123, item.Date)
		}
		if err != nil {
			return nil, skipped, fmt.Errorf("invalid source publication date")
		}
		if published.Before(since) {
			continue
		}
		if !allowedArticleURL("global-voices", item.URL) {
			skipped++
			continue
		}
		doc, err := html.Parse(strings.NewReader(item.Content))
		if err != nil {
			return nil, skipped, err
		}
		// A partner's permission to GV is not a blanket sublicense to our app.
		prefix := plain(doc)
		if len([]rune(prefix)) > 1600 {
			prefix = string([]rune(prefix)[:1600])
		}
		partner := false
		for _, marker := range []string{"コンテンツ共有契約", "共有契約に基づ", "許可を得て転載", "転載許可", "republished with permission", "content-sharing agreement", "もともと掲載", "元々掲載"} {
			if strings.Contains(strings.ToLower(prefix), marker) {
				partner = true
			}
		}
		credits := []Credit{}
		walk(doc, func(n *html.Node) {
			if hasClass(n, "text-credits-section") {
				role := "author"
				if strings.Contains(plain(n), "翻訳") {
					role = "translator"
				}
				walk(n, func(a *html.Node) {
					if a.Data == "a" && hasClass(a, "user-link") {
						credits = append(credits, Credit{Name: plain(a), URL: attr(a, "href"), Role: role})
					}
				})
			}
		})
		for _, credit := range credits {
			for _, name := range []string{"Nepali Times", "Balkan Diskurs", "Meydan TV", "IPS", "Mongabay"} {
				if strings.EqualFold(credit.Name, name) {
					partner = true
				}
			}
		}
		if partner || len(credits) == 0 {
			skipped++
			continue
		}
		paragraphs, refs := parseParagraphs(doc, item.URL)
		if len(paragraphs) == 0 {
			skipped++
			continue
		}
		a := Article{URL: item.URL, Title: strings.TrimSpace(item.Title), TitleTokens: tokenize(item.Title), Paragraphs: paragraphs, References: refs, Credits: credits, PublishedAt: published.UTC(), DateType: "published", License: "CC BY 3.0", LicenseURL: "https://creativecommons.org/licenses/by/3.0/", Changes: "Furigana ditambahkan oleh Nihongo Master; media eksternal tidak disertakan."}
		a.Summary = excerpt(paragraphs[0].Text, 160)
		feedImages := []string{}
		for _, m := range item.Media {
			feedImages = append(feedImages, m.URL)
		}
		for _, m := range item.Thumbnails {
			feedImages = append(feedImages, m.URL)
		}
		for _, m := range item.Enclosures {
			if strings.HasPrefix(m.Type, "image/") {
				feedImages = append(feedImages, m.URL)
			}
		}
		a.imageCandidates = append(licensedImages(doc), plainImageCandidates(item.URL, feedImages, nil, doc)...)
		a.Changes = "Furigana ditambahkan oleh Nihongo Master. Gambar artikel disalin ke penyimpanan Nihongo Master dan dikonversi ke JPEG bila tersedia; kredit gambar mengikuti sumber."
		items = append(items, a)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].PublishedAt.After(items[j].PublishedAt) })
	if len(items) > 20 {
		items = items[:20]
	}
	return items, skipped, nil
}

type osakaLink struct {
	URL, Title string
	Date       time.Time
}

func parseOsakaIndex(data []byte, since time.Time) ([]osakaLink, error) {
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	items := []osakaLink{}
	seen := map[string]bool{}
	found := false
	walk(doc, func(n *html.Node) {
		if n.Data != "li" || !hasClass(n, "cat_topics_icon_none") {
			return
		}
		found = true
		var link *html.Node
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Data == "a" {
				link = c
				break
			}
		}
		if link == nil {
			return
		}
		u := attr(link, "href")
		if !allowedArticleURL("osaka", u) || seen[u] {
			return
		}
		match := japaneseDate.FindStringSubmatch(plain(n))
		if len(match) != 4 {
			return
		}
		date, err := time.ParseInLocation("2006-1-2", match[1]+"-"+match[2]+"-"+match[3], time.FixedZone("JST", 9*3600))
		if err != nil || date.Before(since) {
			return
		}
		items = append(items, osakaLink{URL: u, Title: plain(link), Date: date.UTC()})
		seen[u] = true
	})
	if !found {
		return nil, fmt.Errorf("Osaka index structure changed")
	}
	if len(items) > 20 {
		items = items[:20]
	}
	return items, nil
}
func parseOsakaArticle(data []byte, link osakaLink) (Article, error) {
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return Article{}, err
	}
	var body *html.Node
	walk(doc, func(n *html.Node) {
		if attr(n, "id") == "mol_contents" {
			body = n
		}
	})
	if body == nil {
		return Article{}, fmt.Errorf("Osaka article structure changed")
	}
	paragraphs, refs := parseParagraphs(body, link.URL)
	if len(paragraphs) == 0 {
		return Article{}, fmt.Errorf("Osaka article has no eligible HTML text")
	}
	return Article{imageCandidates: append(licensedImages(body), plainImageCandidates(link.URL, nil, doc, body)...), URL: link.URL, Title: link.Title, TitleTokens: tokenize(link.Title), Summary: excerpt(paragraphs[0].Text, 160), Paragraphs: paragraphs, References: refs, Credits: []Credit{{Name: "大阪市", URL: link.URL, Role: "publisher"}}, PublishedAt: link.Date, DateType: "source", License: "CC BY 4.0 compatible (Osaka City terms)", LicenseURL: "https://www.city.osaka.lg.jp/main/site_policy/0000000124.html", Changes: "Teks HTML Osaka City; furigana ditambahkan oleh Nihongo Master. Gambar artikel disalin ke penyimpanan Nihongo Master; lampiran tidak disertakan."}, nil
}
func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
func hasClass(n *html.Node, class string) bool {
	for _, v := range strings.Fields(attr(n, "class")) {
		if v == class {
			return true
		}
	}
	return false
}
func walk(n *html.Node, visit func(*html.Node)) {
	visit(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, visit)
	}
}
func plain(n *html.Node) string {
	var b strings.Builder
	var read func(*html.Node)
	read = func(v *html.Node) {
		if v.Data == "script" || v.Data == "style" || v.Data == "rt" || v.Data == "rp" {
			return
		}
		if v.Type == html.TextNode {
			b.WriteString(v.Data)
		}
		for c := v.FirstChild; c != nil; c = c.NextSibling {
			read(c)
		}
	}
	read(n)
	return strings.TrimSpace(b.String())
}
func excerpt(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

type rubySpan struct {
	start, end int
	reading    string
}

func blockText(n *html.Node) (string, []rubySpan) {
	var b strings.Builder
	ruby := []rubySpan{}
	var read func(*html.Node)
	read = func(v *html.Node) {
		if v.Data == "rt" || v.Data == "rp" || v.Data == "script" || v.Data == "style" || v.Data == "iframe" || v.Data == "img" {
			return
		}
		if v.Type == html.TextNode {
			b.WriteString(strings.ReplaceAll(v.Data, "\r", ""))
			return
		}
		if v.Data == "br" {
			b.WriteString("\n")
			return
		}
		start := utf16len(b.String())
		for c := v.FirstChild; c != nil; c = c.NextSibling {
			read(c)
		}
		if v.Data == "ruby" {
			reading := ""
			walk(v, func(c *html.Node) {
				if c.Data == "rt" {
					for child := c.FirstChild; child != nil; child = child.NextSibling {
						reading += plain(child)
					}
				}
			})
			if reading != "" {
				ruby = append(ruby, rubySpan{start, utf16len(b.String()), reading})
			}
		}
	}
	read(n)
	raw := b.String()
	trimmed := strings.TrimLeftFunc(raw, unicode.IsSpace)
	removed := utf16len(raw) - utf16len(trimmed)
	text := strings.TrimSpace(trimmed)
	for i := range ruby {
		ruby[i].start -= removed
		ruby[i].end -= removed
	}
	return text, ruby
}
func parseParagraphs(root *html.Node, base string) ([]Paragraph, []Credit) {
	paragraphs := []Paragraph{}
	references := []Credit{}
	seenRefs := map[string]bool{}
	baseURL, _ := url.Parse(base)
	totalRunes := 0
	overflow := false
	var read func(*html.Node)
	read = func(n *html.Node) {
		if overflow {
			return
		}
		if n.Data == "script" || n.Data == "style" || n.Data == "figure" || n.Data == "iframe" || hasClass(n, "wp-caption") || hasClass(n, "gv-rss-footer") || hasClass(n, "mol_imageblock") {
			return
		}
		isBlock := n.Data == "p" || n.Data == "h2" || n.Data == "h3" || n.Data == "h4" || n.Data == "li" || n.Data == "td" || n.Data == "th"
		if isBlock {
			// A list item containing paragraphs is handled as its smaller blocks.
			if n.Data == "li" {
				nested := false
				walk(n, func(c *html.Node) {
					if c != n && (c.Data == "p" || c.Data == "li") {
						nested = true
					}
				})
				if nested {
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						read(c)
					}
					return
				}
			}
			text, ruby := blockText(n)
			if text == "" {
				return
			}
			totalRunes += len([]rune(text))
			if totalRunes > 60000 || len(paragraphs) >= 400 {
				overflow = true
				return
			}
			tokens := tokenize(text)
			for _, span := range ruby {
				if span.start < 0 || span.end > utf16len(text) || span.start >= span.end {
					continue
				}
				filtered := []Token{}
				baseWord := ""
				for _, t := range tokens {
					if t.Start >= span.end || t.End <= span.start {
						filtered = append(filtered, t)
					} else {
						if t.Start <= span.start && t.End >= span.end {
							baseWord = t.Base
						}
						if t.Start < span.start {
							prefix := t
							prefix.End = span.start
							prefix.Surface = utf16slice(text, prefix.Start, prefix.End)
							prefix.Reading = ""
							prefix.ReadingSource = "unknown"
							filtered = append(filtered, prefix)
						}
						if t.End > span.end {
							suffix := t
							suffix.Start = span.end
							suffix.Surface = utf16slice(text, suffix.Start, suffix.End)
							suffix.Reading = ""
							suffix.ReadingSource = "unknown"
							filtered = append(filtered, suffix)
						}
					}
				}
				surface := utf16slice(text, span.start, span.end)
				if baseWord == "" {
					baseWord = surface
				}
				tokens = append(filtered, Token{Start: span.start, End: span.end, Surface: surface, Base: baseWord, Reading: hiragana(span.reading), ReadingSource: "source"})
			}
			sort.Slice(tokens, func(i, j int) bool { return tokens[i].Start < tokens[j].Start })
			kind := "text"
			if strings.HasPrefix(n.Data, "h") {
				kind = "heading"
			}
			paragraphs = append(paragraphs, Paragraph{Text: text, Kind: kind, Tokens: tokens})
			walk(n, func(a *html.Node) {
				if a.Data != "a" {
					return
				}
				u, err := url.Parse(attr(a, "href"))
				if err != nil {
					return
				}
				u = baseURL.ResolveReference(u)
				if u.Scheme != "https" || u.User != nil || seenRefs[u.String()] {
					return
				}
				seenRefs[u.String()] = true
				references = append(references, Credit{Name: plain(a), URL: u.String(), Role: "reference"})
			})
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			read(c)
		}
	}
	read(root)
	if overflow {
		return nil, nil
	}
	return paragraphs, references
}
func utf16slice(s string, start, end int) string {
	position := 0
	var b strings.Builder
	for _, r := range s {
		width := 1
		if r > 0xffff {
			width = 2
		}
		if position >= start && position < end {
			b.WriteRune(r)
		}
		position += width
	}
	return b.String()
}
