package jobfetch

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func boardFixtureClient(t *testing.T, responses map[string]struct {
	status int
	body   string
}) *Client {
	t.Helper()
	c := &Client{requestDelay: 0}
	c.http = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		fixture, ok := responses[r.URL.String()]
		if !ok {
			t.Errorf("unexpected request %s", r.URL)
			fixture = struct {
				status int
				body   string
			}{404, ""}
		}
		return &http.Response{StatusCode: fixture.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(fixture.body)), Request: r}, nil
	})}
	return c
}

func TestTokyoDevSamplesStructuredPostingsWithinBudget(t *testing.T) {
	source := sources[3]
	listing := `<a href="/companies/acme/jobs/backend-engineer">one</a><a href="/companies/acme/jobs/sre">two</a><a href="https://evil.example/companies/x/jobs/y">bad</a>`
	detail := `<script type="application/ld+json">{"@context":"https://schema.org","@type":"JobPosting","title":"Backend Engineer","description":"Build services","datePosted":"2026-01-02T00:00:00Z","employmentType":"FULL_TIME","hiringOrganization":{"name":"Acme"},"jobLocation":{"address":{"addressLocality":"Tokyo","addressCountry":"Japan"}},"baseSalary":{"currency":"JPY","value":{"minValue":6000000,"maxValue":9000000,"unitText":"YEAR"}}}</script>`
	responses := map[string]struct {
		status int
		body   string
	}{
		"https://www.tokyodev.com/robots.txt":                           {200, "User-agent: *\nAllow: /\n"},
		"https://www.tokyodev.com/":                                     {200, listing},
		"https://www.tokyodev.com/companies/acme/jobs/backend-engineer": {200, detail},
		"https://www.tokyodev.com/companies/acme/jobs/sre":              {200, detail},
	}
	result, err := boardFixtureClient(t, responses).Fetch(context.Background(), source.ID, Validators{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Requests != 4 || result.Complete || len(result.Jobs) != 2 {
		t.Fatalf("unexpected result %+v", result)
	}
	job := result.Jobs[0]
	if job.Title != "Backend Engineer" || job.Company != "Acme" || job.Country != "JP" || job.Prefecture != "Tokyo" || job.SalaryMin == nil || *job.SalaryMin != 6000000 || !SafeLink(job.SourceURL) {
		t.Fatalf("bad normalized posting %+v", job)
	}
}

func TestBoardRobotsAndRedirectsAreRespected(t *testing.T) {
	source := sources[4]
	c := boardFixtureClient(t, map[string]struct {
		status int
		body   string
	}{"https://japan-dev.com/robots.txt": {200, "User-agent: *\nDisallow: /jobs\n"}})
	result, err := c.Fetch(context.Background(), source.ID, Validators{})
	if err == nil || result.Requests != 1 || result.Complete || !result.Blocked {
		t.Fatalf("robots denial not respected: %+v %v", result, err)
	}
	c = boardFixtureClient(t, map[string]struct {
		status int
		body   string
	}{"https://japan-dev.com/robots.txt": {200, "User-agent: *\nAllow: /\n"}, "https://japan-dev.com/jobs": {302, ""}})
	result, err = c.Fetch(context.Background(), source.ID, Validators{})
	if err == nil || result.Requests != 2 || result.Complete {
		t.Fatalf("redirect response accepted: %+v %v", result, err)
	}
}

func TestBoardStopsAtDetailBlockAndPreservesRetryAfter(t *testing.T) {
	source := sources[3]
	listing := `<a href="/companies/acme/jobs/blocked">one</a><a href="/companies/acme/jobs/should-not-fetch">two</a>`
	requested := 0
	c := &Client{requestDelay: 0}
	c.http = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requested++
		status, body := http.StatusOK, ""
		header := make(http.Header)
		switch r.URL.String() {
		case "https://www.tokyodev.com/robots.txt":
			body = "User-agent: *\nAllow: /\n"
		case "https://www.tokyodev.com/":
			body = listing
		case "https://www.tokyodev.com/companies/acme/jobs/blocked":
			status = http.StatusTooManyRequests
			header.Set("Retry-After", "3600")
		case "https://www.tokyodev.com/companies/acme/jobs/should-not-fetch":
			t.Error("request continued after 429")
		}
		return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})}
	result, err := c.Fetch(context.Background(), source.ID, Validators{})
	if err == nil || requested != 3 || result.Requests != 3 || !result.Blocked || result.RetryAfter.Before(time.Now().Add(59*time.Minute)) {
		t.Fatalf("block result did not stop and back off: %+v requests=%d error=%v", result, requested, err)
	}
}

func TestRobotsAgentGroupsWildcardAndAllowTie(t *testing.T) {
	emptyRule := "User-agent: *\nDisallow:\nUser-agent: OtherBot\nDisallow: /\n"
	emptyRuleURL, _ := url.Parse("https://www.tokyodev.com/jobs/public")
	if !robotsAllows(emptyRule, emptyRuleURL) {
		t.Fatal("empty disallow merged unrelated bot groups")
	}
	robots := "User-agent: *\nDisallow: /\n\nUser-agent: NihongoMasterJobSync\nAllow: /jobs/*\nDisallow: /jobs/private\n"
	public, _ := url.Parse("https://www.tokyodev.com/jobs/public")
	private, _ := url.Parse("https://www.tokyodev.com/jobs/private/role")
	if !robotsAllows(robots, public) || robotsAllows(robots, private) {
		t.Fatal("specific user-agent rules did not take precedence over wildcard")
	}
	wildcard := "User-agent: *\nDisallow: /jobs/*\nAllow: /jobs/public$\n"
	if !robotsAllows(wildcard, public) || robotsAllows(wildcard, private) {
		t.Fatal("robots wildcard or longest matching rule was applied incorrectly")
	}
	tie := "User-agent: *\nDisallow: /jobs/public\nAllow: /jobs/public\n"
	if !robotsAllows(tie, public) {
		t.Fatal("Allow did not win a same-length robots rule tie")
	}
	query := "User-agent: *\nDisallow: /*?secret=*$\n"
	withQuery, _ := url.Parse("https://www.tokyodev.com/jobs?secret=1")
	if robotsAllows(query, withQuery) {
		t.Fatal("robots wildcard query rule was ignored")
	}
}

func TestPostingCountryArraysAndCompanyRequirement(t *testing.T) {
	source := sources[9] // AgriNavi
	pageURL := "https://www.agri-navi.com/jobdetail/13954"
	page := `<script type="application/ld+json">{"@type":"JobPosting","title":"農業スタッフ","hiringOrganization":{"name":"農園"},"jobLocation":[{"address":{"addressCountry":{"@type":"Country","name":"Japan"},"addressRegion":"Hokkaido"}},{"address":{"addressCountry":"JP"}}]}</script>`
	job, ok := parseJobPosting(source, pageURL, []byte(page))
	if !ok || job.Country != "JP" || job.Category != "agriculture" || job.Company != "農園" {
		t.Fatalf("Japan country in structured location array was lost: %+v %v", job, ok)
	}
	withoutCompany := strings.Replace(page, `,"hiringOrganization":{"name":"農園"}`, "", 1)
	if _, ok := parseJobPosting(source, pageURL, []byte(withoutCompany)); ok {
		t.Fatal("posting without a company was accepted")
	}
}

func TestBoardWithoutStructuredPostingIsExplicitlyUnsupported(t *testing.T) {
	source := sources[7] // LAPRAS
	result, err := boardFixtureClient(t, map[string]struct {
		status int
		body   string
	}{
		"https://lapras.com/robots.txt": {200, "User-agent: *\nAllow: /\n"},
		"https://lapras.com/":           {200, `<a href="/jobs/123">job</a>`},
		"https://lapras.com/jobs/123":   {200, `<title>job with no schema</title>`},
	}).Fetch(context.Background(), source.ID, Validators{})
	if err == nil || !result.Unsupported || result.Complete || result.Requests != 3 {
		t.Fatalf("missing unsupported result: %+v %v", result, err)
	}
}

func TestSourceRequestBudgetsAndBoardHosts(t *testing.T) {
	all := Sources()
	for _, s := range all[:3] {
		if s.RequestBudget != 1 {
			t.Fatalf("API source %s budget=%d", s.ID, s.RequestBudget)
		}
	}
	for _, s := range all[3:] {
		if s.RequestBudget != 4 {
			t.Fatalf("board source %s budget=%d", s.ID, s.RequestBudget)
		}
	}
	for _, link := range []string{"https://www.tokyodev.com/jobs/a", "https://japan-dev.com/jobs/a/b", "https://jobsinjapan.com/jobs/12/a/", "https://www.agri-navi.com/jobdetail/13954"} {
		if !SafeLink(link) {
			t.Errorf("board link rejected: %s", link)
		}
	}
	for _, link := range []string{"https://tokyodev.com/jobs/a", "https://evil.example/jobs/a"} {
		if SafeLink(link) {
			t.Errorf("unverified host accepted: %s", link)
		}
	}
}
