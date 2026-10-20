package jobfetch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const boardMaxBytes = 4 << 20

var jobIDPath = regexp.MustCompile(`^/jobs/[0-9]+/`)

// FetchPublicBoard runs the bounded robots/listing/structured-detail workflow
// for a fixed Source. JobPathPrefix and related path constraints must be
// reviewed constants, and the source host must be present in SafeLink's exact
// hostname allowlist. The request budget is capped at four.
func (c *Client) FetchPublicBoard(ctx context.Context, source Source) (Result, error) {
	endpoint, err := url.Parse(source.Endpoint)
	if err != nil || endpoint.Host != source.Domain || !SafeLink(source.Endpoint) || source.RequestBudget < 2 || source.RequestBudget > 4 || source.JobPathPrefix == "" {
		return Result{}, errors.New("public board source is not in the reviewed fixed-source format")
	}
	return c.fetchBoard(ctx, source)
}

// fetchBoard makes at most four requests: robots.txt, one listing page, and up
// to two posting pages. This collector intentionally produces partial snapshots
// because a small sample cannot authoritatively retire old listings.
func (c *Client) fetchBoard(ctx context.Context, source Source) (Result, error) {
	result := Result{}
	base := "https://" + source.Domain
	robots, err := c.boardGet(ctx, base+"/robots.txt", &result)
	if err != nil {
		return result, err
	}
	if robots.StatusCode != http.StatusOK {
		robots.Body.Close()
		markBoardResponse(&result, robots)
		return result, fmt.Errorf("robots.txt request returned HTTP %d", robots.StatusCode)
	}
	robotsText, err := readLimited(robots.Body, boardMaxBytes)
	robots.Body.Close()
	if err != nil {
		return result, err
	}
	indexURL, _ := url.Parse(source.Endpoint)
	if !robotsAllows(string(robotsText), indexURL) {
		result.Blocked = true
		return result, errors.New("robots.txt disallows configured listing path")
	}

	index, err := c.boardGet(ctx, source.Endpoint, &result)
	if err != nil {
		return result, err
	}
	if index.StatusCode != http.StatusOK {
		index.Body.Close()
		markBoardResponse(&result, index)
		return result, fmt.Errorf("listing request returned HTTP %d", index.StatusCode)
	}
	page, err := readLimited(index.Body, boardMaxBytes)
	index.Body.Close()
	if err != nil {
		return result, err
	}
	links := boardLinks(source, page)
	var problems []string
	for _, link := range links {
		if result.Requests >= source.RequestBudget {
			break
		}
		u, _ := url.Parse(link)
		if !robotsAllows(string(robotsText), u) {
			continue
		}
		resp, reqErr := c.boardGet(ctx, link, &result)
		if reqErr != nil {
			problems = append(problems, reqErr.Error())
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			markBoardResponse(&result, resp)
			if result.Blocked {
				return result, fmt.Errorf("posting request paused after HTTP %d", resp.StatusCode)
			}
			problems = append(problems, fmt.Sprintf("posting request returned HTTP %d", resp.StatusCode))
			continue
		}
		body, readErr := readLimited(resp.Body, boardMaxBytes)
		resp.Body.Close()
		if readErr != nil {
			problems = append(problems, readErr.Error())
			continue
		}
		job, ok := parseJobPosting(source, link, body)
		if ok {
			result.Jobs = append(result.Jobs, job)
		}
	}
	if len(result.Jobs) == 0 {
		result.Unsupported = true
		return result, errors.New("public board sample contained no extractable schema.org JobPosting")
	}
	if len(problems) > 0 {
		return result, errors.New(strings.Join(problems, "; "))
	}
	return result, nil
}

func markBoardResponse(result *Result, response *http.Response) {
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusTooManyRequests {
		result.Blocked = true
		result.RetryAfter = retryAfter(response.Header.Get("Retry-After"), time.Now().UTC())
	}
}

func (c *Client) boardGet(ctx context.Context, raw string, result *Result) (*http.Response, error) {
	if result.Requests > 0 && c.requestDelay > 0 {
		timer := time.NewTimer(c.requestDelay)
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
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/json;q=0.9,*/*;q=0.8")
	req.Header.Set("User-Agent", "NihongoMasterJobSync/1.0 (weekly public listing cache)")
	result.Requests++
	return c.http.Do(req)
}

func readLimited(r io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New("public board response exceeds byte limit")
	}
	return b, nil
}

func boardLinks(source Source, body []byte) []string {
	z := html.NewTokenizer(strings.NewReader(string(body)))
	seen := map[string]bool{}
	var links []string
	base, _ := url.Parse("https://" + source.Domain)
	for z.Next() != html.ErrorToken {
		tok := z.Token()
		typeOfToken := tok.Type
		if typeOfToken != html.StartTagToken && typeOfToken != html.SelfClosingTagToken {
			continue
		}
		if tok.Data != "a" {
			continue
		}
		for _, attr := range tok.Attr {
			if attr.Key != "href" {
				continue
			}
			u, err := url.Parse(strings.TrimSpace(attr.Val))
			if err != nil {
				continue
			}
			u = base.ResolveReference(u)
			if u.Scheme != "https" || !strings.EqualFold(u.Host, source.Domain) || u.User != nil || u.Port() != "" || !isBoardJobPath(source, u.Path) {
				continue
			}
			u.RawQuery = ""
			u.ForceQuery = false
			u.Fragment = ""
			link := u.String()
			if !seen[link] && SafeLink(link) {
				seen[link] = true
				links = append(links, link)
				if len(links) == 2 {
					return links
				}
			}
		}
	}
	return links
}

func isBoardJobPath(source Source, p string) bool {
	if source.JobPathPrefix == "" || !strings.HasPrefix(p, source.JobPathPrefix) {
		return false
	}
	if source.JobPathContains != "" && !strings.Contains(p, source.JobPathContains) {
		return false
	}
	if source.JobPathSuffix != "" && !strings.HasSuffix(p, source.JobPathSuffix) {
		return false
	}
	if source.JobPathMinSegments > 0 && strings.Count(strings.Trim(p, "/"), "/")+1 < source.JobPathMinSegments {
		return false
	}
	if source.ID == "jobs-in-japan" && !jobIDPath.MatchString(p) {
		return false
	}
	if source.ID == "lapras" && !regexp.MustCompile(`^/jobs/[0-9]+$`).MatchString(p) {
		return false
	}
	return true
}

func parseJobPosting(source Source, pageURL string, body []byte) (Job, bool) {
	root, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return Job{}, false
	}
	var posting map[string]any
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if posting != nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "script" {
			isJSONLD := false
			for _, a := range n.Attr {
				if a.Key == "type" && strings.HasPrefix(strings.ToLower(a.Val), "application/ld+json") {
					isJSONLD = true
				}
			}
			if isJSONLD {
				var v any
				if json.Unmarshal([]byte(textContent(n)), &v) == nil {
					posting = findPosting(v)
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	if posting == nil {
		return Job{}, false
	}
	title := jsonString(posting["title"])
	desc := plainHTML(jsonString(posting["description"]))
	if strings.TrimSpace(title) == "" {
		return Job{}, false
	}
	urlValue := jsonString(posting["url"])
	if urlValue == "" {
		urlValue = pageURL
	}
	if !sameBoardLink(source, urlValue) {
		urlValue = pageURL
	}
	company := jsonString(nested(posting, "hiringOrganization", "name"))
	if company == "" {
		company = jsonString(posting["hiringOrganization"])
	}
	if strings.TrimSpace(company) == "" {
		return Job{}, false
	}
	loc := postingLocation(posting["jobLocation"])
	if loc == "" {
		loc = postingLocation(posting["jobLocationType"])
	}
	if !hasJapanCountry(posting["jobLocation"]) && !strings.Contains(strings.ToLower(loc), "japan") && !strings.Contains(loc, "日本") {
		return Job{}, false
	}
	industry := jsonString(posting["industry"])
	qualifications := plainHTML(jsonString(posting["qualifications"]))
	jobCategory := source.Category
	if jobCategory == "" {
		jobCategory = category(title, desc, industry, jsonString(posting["occupationalCategory"]))
	}
	job := Job{ExternalID: boardExternalID(source.ID, urlValue), Title: bounded(title, 500), Company: bounded(company, 300), Description: bounded(desc, 30000), Qualifications: bounded(qualifications, 12000), Benefits: postingBenefits(posting["jobBenefits"]), Location: bounded(loc, 500), Country: "JP", Prefecture: prefecture(loc), Category: jobCategory, Industry: bounded(industry, 300), EmploymentType: boardEmployment(jsonString(posting["employmentType"])), Remote: "unknown", Visa: "unknown", JapaneseLevel: "unknown", SourceURL: urlValue, ApplyURL: urlValue}
	if strings.Contains(strings.ToLower(jsonString(posting["jobLocationType"])), "telecommute") {
		job.Remote = "remote"
	}
	if date, ok := parseBoardDate(jsonString(posting["datePosted"])); ok {
		t := date.UTC()
		job.PublishedAt = &t
	}
	if date, ok := parseBoardDate(jsonString(posting["validThrough"])); ok {
		t := date.UTC()
		job.ExpiresAt = &t
	}
	if sal, ok := posting["baseSalary"].(map[string]any); ok {
		job.SalaryCurrency = bounded(jsonString(sal["currency"]), 8)
		value, _ := sal["value"].(map[string]any)
		job.SalaryMin = jsonNumber(value["minValue"])
		job.SalaryMax = jsonNumber(value["maxValue"])
		job.SalaryPeriod = salaryPeriod(jsonString(value["unitText"]))
	}
	return job, true
}

func findPosting(v any) map[string]any {
	switch x := v.(type) {
	case map[string]any:
		t := strings.ToLower(jsonString(x["@type"]))
		if t == "jobposting" || strings.HasSuffix(t, "/jobposting") || strings.HasSuffix(t, "#jobposting") {
			return x
		}
		for _, value := range x {
			if found := findPosting(value); found != nil {
				return found
			}
		}
	case []any:
		for _, value := range x {
			if found := findPosting(value); found != nil {
				return found
			}
		}
	}
	return nil
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var visit func(*html.Node)
	visit = func(x *html.Node) {
		if x.Type == html.TextNode {
			b.WriteString(x.Data)
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(n)
	return b.String()
}
func jsonString(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	if a, ok := v.([]any); ok && len(a) > 0 {
		return jsonString(a[0])
	}
	return ""
}
func countryValue(v any) string {
	if s := jsonString(v); s != "" {
		return s
	}
	if m, ok := v.(map[string]any); ok {
		for _, key := range []string{"name", "identifier"} {
			if s := countryValue(m[key]); s != "" {
				return s
			}
		}
	}
	if a, ok := v.([]any); ok {
		for _, item := range a {
			if s := countryValue(item); s != "" {
				return s
			}
		}
	}
	return ""
}
func hasJapanCountry(v any) bool {
	switch x := v.(type) {
	case map[string]any:
		for key, value := range x {
			if strings.EqualFold(key, "addressCountry") {
				country := strings.TrimSpace(countryValue(value))
				if strings.EqualFold(country, "JP") || strings.EqualFold(country, "Japan") || country == "日本" {
					return true
				}
			}
			if hasJapanCountry(value) {
				return true
			}
		}
	case []any:
		for _, value := range x {
			if hasJapanCountry(value) {
				return true
			}
		}
	}
	return false
}
func boardEmployment(s string) string {
	normalized := strings.ToLower(strings.ReplaceAll(s, "_", "-"))
	switch normalized {
	case "full-time":
		return "full-time"
	case "part-time":
		return "part-time"
	case "contractor", "contract":
		return "contract"
	case "intern", "internship":
		return "internship"
	default:
		return employment(s)
	}
}
func nested(m map[string]any, keys ...string) any {
	var v any = m
	for _, k := range keys {
		x, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = x[k]
	}
	return v
}
func postingLocation(v any) string {
	if arr, ok := v.([]any); ok {
		var out []string
		for _, x := range arr {
			if s := postingLocation(x); s != "" {
				out = append(out, s)
			}
		}
		return strings.Join(out, "; ")
	}
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	if addr, ok := m["address"].(map[string]any); ok {
		place := jsonString(addr["addressLocality"])
		if place == "" {
			place = jsonString(addr["streetAddress"])
		}
		return strings.Join(nonEmpty(place, jsonString(addr["addressRegion"]), countryValue(addr["addressCountry"])), ", ")
	}
	return jsonString(m["name"])
}
func nonEmpty(s ...string) []string {
	var out []string
	for _, v := range s {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
func jsonNumber(v any) *float64 {
	n, ok := v.(float64)
	if !ok || n <= 0 || n > 1e12 {
		return nil
	}
	return &n
}
func parseBoardDate(s string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02"} {
		if parsed, err := time.Parse(layout, s); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}
func boardExternalID(sourceID, u string) string {
	sum := sha256.Sum256([]byte(u))
	return sourceID + "-" + hex.EncodeToString(sum[:12])
}
func sameBoardLink(source Source, raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && strings.EqualFold(u.Host, source.Domain) && SafeLink(raw)
}

type robotRule struct {
	path  string
	allow bool
}

func robotsAllows(text string, target *url.URL) bool {
	type robotGroup struct {
		agents   []string
		rules    []robotRule
		hasRules bool
	}
	var groups []robotGroup
	var current *robotGroup
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		if line == "" {
			if current != nil && current.hasRules {
				current = nil
			}
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := strings.ToLower(strings.TrimSpace(parts[0])), strings.TrimSpace(parts[1])
		switch key {
		case "user-agent":
			if current == nil || current.hasRules {
				groups = append(groups, robotGroup{})
				current = &groups[len(groups)-1]
			}
			current.agents = append(current.agents, strings.ToLower(val))
		case "allow", "disallow":
			if current != nil {
				current.hasRules = true
				if val != "" {
					current.rules = append(current.rules, robotRule{path: val, allow: key == "allow"})
				}
			}
		}
	}
	const productToken = "nihongomasterjobsync/1.0"
	bestAgent := -1
	var rules []robotRule
	for _, group := range groups {
		score := -1
		for _, agent := range group.agents {
			if agent == "*" {
				if score < 0 {
					score = 0
				}
			} else if strings.Contains(productToken, agent) && len(agent) > score {
				score = len(agent)
			}
		}
		if score < 0 {
			continue
		}
		if score > bestAgent {
			bestAgent = score
			rules = append([]robotRule(nil), group.rules...)
		} else if score == bestAgent {
			rules = append(rules, group.rules...)
		}
	}
	path := target.EscapedPath()
	if path == "" {
		path = "/"
	}
	if target.RawQuery != "" {
		path += "?" + target.RawQuery
	}
	best := -1
	allowed := true
	for _, rule := range rules {
		if !robotPathMatches(rule.path, path) {
			continue
		}
		specificity := len(strings.ReplaceAll(strings.TrimSuffix(rule.path, "$"), "*", ""))
		if specificity > best || (specificity == best && rule.allow) {
			best = specificity
			allowed = rule.allow
		}
	}
	return allowed
}

func robotPathMatches(pattern, path string) bool {
	anchored := strings.HasSuffix(pattern, "$")
	pattern = strings.TrimSuffix(pattern, "$")
	var b strings.Builder
	b.WriteByte('^')
	for _, r := range pattern {
		if r == '*' {
			b.WriteString(".*")
		} else {
			b.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	if anchored {
		b.WriteByte('$')
	}
	matched, err := regexp.MatchString(b.String(), path)
	return err == nil && matched
}

// Preserve explicit employer benefits only, using the listing HTML cleanup.
func postingBenefits(value any) string {
	var sections []string
	switch v := value.(type) {
	case string:
		sections = append(sections, plainHTML(v))
	case []any:
		for _, item := range v {
			if text, ok := item.(string); ok {
				sections = append(sections, plainHTML(text))
			}
		}
	}
	return bounded(strings.TrimSpace(strings.Join(sections, "\n")), 30000)
}
