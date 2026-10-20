package manga

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/net/html"
)

var ErrBlocked = errors.New("manga is adult, ecchi, or lacks reviewed metadata")
var seriesPath = regexp.MustCompile(`^/manga/[a-z0-9][a-z0-9-]*/$`)
var chapterName = regexp.MustCompile(`^chapter-([0-9]+(?:\.[0-9]+)?)\.[0-9]+$`)
var ErrInvalid = errors.New("invalid manga request")
var ErrNotFound = errors.New("manga not found")
var ErrBusy = errors.New("manga import queue is full")

type HTTPError struct{ Status int }

func (e HTTPError) Error() string { return fmt.Sprintf("manga source HTTP %d", e.Status) }

func ID(raw string) string { h := sha256.Sum256([]byte(raw)); return hex.EncodeToString(h[:]) }

func ValidSeriesURL(raw string) bool {
	if len(raw) > 500 {
		return false
	}
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Host == "rawkuma.net" && u.User == nil && u.RawQuery == "" && u.Fragment == "" && u.RawPath == "" && seriesPath.MatchString(u.Path)
}

func allowedURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" || strings.Contains(u.Path, "..") {
		return false
	}
	switch u.Host {
	case "rawkuma.net":
		return strings.HasPrefix(u.Path, "/manga/") || strings.HasPrefix(u.Path, "/wp-content/uploads/") || u.Path == "/robots.txt"
	case "cdn.rawkuma.com":
		return u.Path != "/" && u.Path != ""
	}
	return false
}

func safeGenres(genres []string) bool {
	if len(genres) == 0 {
		return false
	}
	for _, g := range genres {
		s := strings.ToLower(strings.TrimSpace(g))
		for _, bad := range []string{"adult", "ecchi", "hentai", "smut", "mature", "erotic", "porn", "lolicon", "shotacon", "harem", "sexual", "18+", "r18", "r-18", "seijin", "18禁", "成人", "エッチ", "エロ", "アダルト"} {
			if strings.Contains(s, bad) {
				return false
			}
		}
	}
	return true
}

type SeriesImport struct {
	Series   Series
	Cover    string
	Chapters []Chapter
}
type Fetcher interface {
	Series(context.Context, string) (SeriesImport, error)
	Pages(context.Context, Chapter) ([]string, error)
	Image(context.Context, string, string) ([]byte, error)
}

type HTTPFetcher struct {
	client *http.Client
	mu     sync.Mutex
	next   time.Time
	robots map[string]string
	gap    time.Duration
}

func NewFetcher() *HTTPFetcher {
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	transport := &http.Transport{MaxConnsPerHost: 1, MaxIdleConnsPerHost: 1, TLSHandshakeTimeout: 15 * time.Second, ResponseHeaderTimeout: 20 * time.Second}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			if !ip.IP.IsGlobalUnicast() || ip.IP.IsPrivate() || ip.IP.IsLoopback() || ip.IP.IsLinkLocalUnicast() {
				return nil, errors.New("manga source resolved to non-public address")
			}
		}
		if len(ips) == 0 {
			return nil, errors.New("manga source has no address")
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
	}
	client := &http.Client{Timeout: 40 * time.Second, Transport: transport, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// This CDN sends nonexistent files (including robots.txt) to its legacy
		// site's homepage. Do not crawl that homepage as if it were robots rules.
		if len(via) > 0 && via[0].URL.Host == "cdn.rawkuma.com" && via[0].URL.Path == "/robots.txt" && req.URL.String() == "https://rawkuma.com/" {
			return http.ErrUseLastResponse
		}
		if len(via) >= 3 || !allowedURL(req.URL.String()) {
			return errors.New("manga redirect outside allowed sources")
		}
		return nil
	}}
	return &HTTPFetcher{client: client, robots: map[string]string{}, gap: 2 * time.Second}
}

// Reset robots once per worker batch, never once per reader request.
func (f *HTTPFetcher) ResetRobots() { f.mu.Lock(); defer f.mu.Unlock(); f.robots = map[string]string{} }

func (f *HTTPFetcher) get(ctx context.Context, raw, referer string, limit int64, checkRobots bool) ([]byte, error) {
	if !allowedURL(raw) {
		return nil, ErrInvalid
	}
	u, _ := url.Parse(raw)
	if checkRobots {
		f.mu.Lock()
		rules, ok := f.robots[u.Host]
		f.mu.Unlock()
		if !ok {
			body, err := f.get(ctx, "https://"+u.Host+"/robots.txt", "", 256<<10, false)
			if err != nil && !strings.Contains(err.Error(), "HTTP 404") {
				return nil, err
			}
			rules = string(body)
			f.mu.Lock()
			f.robots[u.Host] = rules
			f.mu.Unlock()
		}
		if !robotsAllow(rules, u.Path) {
			return nil, errors.New("manga path disallowed by robots.txt")
		}
	}
	f.mu.Lock()
	wait := time.Until(f.next)
	if wait < 0 {
		wait = 0
	}
	f.next = time.Now().Add(wait + f.gap)
	f.mu.Unlock()
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "NihongoMasterManga/1.0 (weekly curated Japanese manga cache)")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	res, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if !checkRobots && u.Host == "cdn.rawkuma.com" && u.Path == "/robots.txt" && res.StatusCode >= 300 && res.StatusCode < 400 && res.Header.Get("Location") == "https://rawkuma.com/" {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, HTTPError{Status: res.StatusCode}
	}
	if res.ContentLength > limit {
		return nil, errors.New("manga response too large")
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, errors.New("manga response too large")
	}
	return body, nil
}

func robotsAllow(text, target string) bool {
	// Most-specific matching UA group, longest rule; Allow wins a tie.
	type group struct {
		agents []string
		rules  [][2]string
	}
	groups := []group{}
	g := group{}
	hasRules := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			continue
		}
		key, val := strings.ToLower(strings.TrimSpace(kv[0])), strings.TrimSpace(kv[1])
		if key == "user-agent" {
			if hasRules {
				groups = append(groups, g)
				g = group{}
				hasRules = false
			}
			g.agents = append(g.agents, strings.ToLower(val))
		} else if (key == "allow" || key == "disallow") && len(g.agents) > 0 {
			hasRules = true
			if val != "" {
				g.rules = append(g.rules, [2]string{key, val})
			}
		}
	}
	groups = append(groups, g)
	bestAgent := -1
	selected := [][2]string{}
	for _, group := range groups {
		score := -1
		for _, a := range group.agents {
			if a == "*" && score < 0 {
				score = 0
			}
			if a != "" && a != "*" && strings.Contains("nihongomastermanga", a) && len(a) > score {
				score = len(a)
			}
		}
		if score > bestAgent {
			bestAgent = score
			selected = group.rules
		} else if score == bestAgent && score >= 0 {
			selected = append(selected, group.rules...)
		}
	}
	allow, best := true, -1
	for _, rule := range selected {
		pattern := regexp.QuoteMeta(rule[1])
		pattern = strings.ReplaceAll(pattern, `\*`, ".*")
		pattern = strings.TrimSuffix(pattern, `\$`)
		if strings.HasSuffix(rule[1], "$") {
			pattern += "$"
		}
		match, _ := regexp.MatchString("^"+pattern, target)
		score := len(strings.ReplaceAll(rule[1], "*", ""))
		if match && (score > best || (score == best && rule[0] == "allow")) {
			best = score
			allow = rule[0] == "allow"
		}
	}
	return allow
}

func (f *HTTPFetcher) Series(ctx context.Context, raw string) (SeriesImport, error) {
	if !ValidSeriesURL(raw) {
		return SeriesImport{}, ErrInvalid
	}
	body, err := f.get(ctx, raw, "", 6<<20, true)
	if err != nil {
		return SeriesImport{}, err
	}
	return parseSeries(body, raw)
}
func (f *HTTPFetcher) Pages(ctx context.Context, ch Chapter) ([]string, error) {
	body, err := f.get(ctx, ch.URL, "", 6<<20, true)
	if err != nil {
		return nil, err
	}
	return parsePages(body, ch)
}
func (f *HTTPFetcher) Image(ctx context.Context, raw, referer string) ([]byte, error) {
	return f.get(ctx, raw, referer, MaxImageBytes, true)
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
func text(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(text(c))
	}
	return b.String()
}
func walk(n *html.Node, visit func(*html.Node)) {
	visit(n)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, visit)
	}
}

func parseSeries(body []byte, raw string) (SeriesImport, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return SeriesImport{}, err
	}
	var schema struct {
		Name, Description, AlternateName, URL, CreativeWorkStatus string
		Genre                                                     []string
		Image                                                     struct{ URL string }
		Author                                                    struct{ Name string }
	}
	found := false
	walk(doc, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" && attr(n, "type") == "application/ld+json" && strings.Contains(text(n), `"ComicSeries"`) {
			if json.Unmarshal([]byte(text(n)), &schema) == nil && schema.Name != "" && schema.URL == raw {
				found = true
			}
		}
	})
	if !found {
		return SeriesImport{}, errors.New("manga series schema missing or changed")
	}
	if !safeGenres(schema.Genre) {
		return SeriesImport{}, ErrBlocked
	}
	if !allowedURL(schema.Image.URL) {
		return SeriesImport{}, errors.New("manga cover is outside allowed sources")
	}
	jp := ""
	for _, v := range strings.Split(schema.AlternateName, ",") {
		if strings.ContainsFunc(v, func(r rune) bool { return unicode.In(r, unicode.Hiragana, unicode.Katakana) }) {
			jp = strings.TrimSpace(v)
			break
		}
	}
	if len(schema.Name) > 500 || len(schema.Description) > 16000 {
		return SeriesImport{}, ErrInvalid
	}
	out := SeriesImport{Series: Series{ID: ID(raw), URL: raw, Title: schema.Name, JapaneseTitle: jp, Summary: html.UnescapeString(schema.Description), Author: schema.Author.Name, Genres: schema.Genre, Language: "ja", PublicationStatus: schema.CreativeWorkStatus}, Cover: schema.Image.URL, Chapters: []Chapter{}}
	seen := map[string]bool{}
	walk(doc, func(n *html.Node) {
		if n.Type != html.ElementNode || n.Data != "a" {
			return
		}
		link := strings.TrimSpace(attr(n, "href"))
		if !strings.HasPrefix(link, raw+"chapter-") || !strings.HasSuffix(link, "/") || !allowedURL(link) || seen[link] {
			return
		}
		u, _ := url.Parse(link)
		match := chapterName.FindStringSubmatch(path.Base(u.Path))
		if match == nil {
			return
		}
		position, e := strconv.ParseFloat(match[1], 64)
		if e != nil || position < 0 {
			return
		}
		seen[link] = true
		out.Chapters = append(out.Chapters, Chapter{ID: ID(link), SeriesID: ID(raw), URL: link, Title: "Chapter " + match[1], Number: match[1], Position: position, Status: "available", Pages: []Page{}})
	})
	if len(out.Chapters) == 0 || len(out.Chapters) > 5000 {
		return SeriesImport{}, errors.New("manga chapter list missing or exceeds limit")
	}
	sort.Slice(out.Chapters, func(i, j int) bool { return out.Chapters[i].Position < out.Chapters[j].Position })
	return out, nil
}

func parsePages(body []byte, ch Chapter) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(ch.URL)
	if err != nil {
		return nil, ErrInvalid
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 3 {
		return nil, ErrInvalid
	}
	slug := parts[1]
	pages := []string{}
	seen := map[string]bool{}
	walk(doc, func(n *html.Node) {
		if n.Type != html.ElementNode || n.Data != "img" {
			return
		}
		raw := strings.TrimSpace(attr(n, "data-src"))
		if raw == "" {
			raw = strings.TrimSpace(attr(n, "src"))
		}
		p, e := url.Parse(raw)
		// Only numbered images from this exact series/chapter are accepted. Logos,
		// covers, ads, pixels, recommendations and scripts never enter the reader.
		if e != nil || p.Host != "cdn.rawkuma.com" || !allowedURL(raw) || seen[raw] {
			return
		}
		bits := strings.Split(strings.Trim(p.Path, "/"), "/")
		if len(bits) == 5 && bits[0] == "images" {
			bits = bits[1:]
		}
		if len(bits) != 4 || bits[1] != slug || strings.TrimPrefix(bits[2], "chapter-") != ch.Number {
			return
		}
		if !regexp.MustCompile(`^[0-9]+\.(jpg|jpeg|png|webp)$`).MatchString(bits[3]) {
			return
		}
		seen[raw] = true
		pages = append(pages, raw)
	})
	if len(pages) == 0 || len(pages) > 120 {
		return nil, errors.New("manga reader images missing, unsupported source layout, or over 120 pages")
	}
	// Preserve publisher DOM order; filenames are not a reliable ordering signal.
	return pages, nil
}
