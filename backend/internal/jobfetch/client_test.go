package jobfetch

import (
	"context"

	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func testClient(t *testing.T, status int, body string, headers http.Header) *Client {
	t.Helper()
	return &Client{http: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != Sources()[0].Endpoint || r.Method != "GET" {
			t.Fatalf("unexpected endpoint %s", r.URL)
		}
		if !strings.Contains(r.UserAgent(), "NihongoMaster") {
			t.Fatal("missing identifying user agent")
		}
		return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
}

const japanPosting = `{"id":"abc-123","text":"Mechanical engineer","country":"JP","categories":{"location":"Tokyo, Japan","commitment":"Full-Time","team":"Robotics","department":"Operations"},"descriptionPlain":"Build robots.","lists":[{"text":"Requirements","content":"<ul><li>Japanese reading</li><li>Two years experience</li></ul><script>alert(1)</script>"}],"additionalPlain":"Apply directly.","hostedUrl":"https://jobs.lever.co/mujininc/abc-123","applyUrl":"https://jobs.lever.co/mujininc/abc-123/apply","workplaceType":"on-site","createdAt":1700000000000,"salaryRange":{"currency":"JPY","interval":"per year","min":4000000,"max":7000000}}`

func TestJapanNormalization(t *testing.T) {
	result, err := testClient(t, 200, "["+japanPosting+"]", nil).Fetch(context.Background(), "lever-mujininc", Validators{})
	if err != nil || !result.Complete || len(result.Jobs) != 1 {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	j := result.Jobs[0]
	if j.Country != "JP" || j.Prefecture != "Tokyo" || j.Category != "manufacturing" || j.Remote != "onsite" || j.EmploymentType != "full-time" {
		t.Fatalf("bad normalized job: %+v", j)
	}
	if j.SalaryMin == nil || *j.SalaryMin != 4000000 || j.SalaryCurrency != "JPY" || j.SalaryPeriod != "year" {
		t.Fatalf("bad salary: %+v", j)
	}
	if j.JapaneseLevel != "unknown" || j.Visa != "unknown" {
		t.Fatal("inferred unsupported immigration/language requirement")
	}
	if strings.Contains(j.Description, "<") || strings.Contains(j.Qualifications, "alert") || !strings.Contains(j.Qualifications, "Two years") {
		t.Fatalf("unsafe/missing plain text: %s", j.Qualifications)
	}
	if j.PublishedAt == nil {
		t.Fatal("missing original publication time")
	}
}

func TestCountryEvidenceRequired(t *testing.T) {
	for _, country := range []string{"US", "", "CN"} {
		body := strings.ReplaceAll(japanPosting, `"country":"JP"`, `"country":"`+country+`"`)
		body = strings.ReplaceAll(body, "Mechanical engineer", "Japan Tokyo developer")
		result, err := testClient(t, 200, "["+body+"]", nil).Fetch(context.Background(), "lever-mujininc", Validators{})
		if err != nil || len(result.Jobs) != 0 || !result.Complete {
			t.Fatalf("country=%s included or failed: %+v %v", country, result, err)
		}
	}
}

func TestConditionalsAndBackoff(t *testing.T) {
	client := testClient(t, 304, "", http.Header{"Etag": []string{`"v2"`}})
	base := client.http.Transport
	client.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("If-None-Match") != `"v1"` || r.Header.Get("If-Modified-Since") != "yesterday" {
			t.Fatal("missing validators")
		}
		return base.RoundTrip(r)
	})
	result, err := client.Fetch(context.Background(), "lever-mujininc", Validators{ETag: `"v1"`, LastModified: "yesterday"})
	if err != nil || !result.NotModified || result.Complete || result.ETag != `"v2"` {
		t.Fatalf("bad304: %+v %v", result, err)
	}
	for _, status := range []int{403, 429} {
		result, err = testClient(t, status, "", http.Header{"Retry-After": []string{"86400"}}).Fetch(context.Background(), "lever-mujininc", Validators{})
		if err == nil || !result.Blocked || result.RetryAfter.Before(time.Now().Add(23*time.Hour)) {
			t.Fatalf("bad block: %+v %v", result, err)
		}
	}
}

func TestInvalidSnapshotDoesNotRetireJobs(t *testing.T) {
	for _, body := range []string{"{}", "null", "not json", "[" + japanPosting + "," + japanPosting + "]", strings.ReplaceAll("["+japanPosting+"]", "https://jobs.lever.co/mujininc/abc-123/apply", "http://127.0.0.1/apply")} {
		result, err := testClient(t, 200, body, nil).Fetch(context.Background(), "lever-mujininc", Validators{})
		if err == nil || result.Complete {
			t.Fatalf("invalid snapshot marked complete: %s", body)
		}
	}
	for _, status := range []int{301, 302, 500} {
		result, err := testClient(t, status, "", nil).Fetch(context.Background(), "lever-mujininc", Validators{})
		if err == nil || result.Complete {
			t.Fatalf("HTTP %d accepted", status)
		}
	}
	client := testClient(t, 200, strings.Repeat(" ", maxBytes+1), nil)
	if result, err := client.Fetch(context.Background(), "lever-mujininc", Validators{}); err == nil || result.Complete {
		t.Fatal("oversize feed accepted")
	}
	rows := strings.Repeat(japanPosting+",", maxPostings) + japanPosting
	if result, err := testClient(t, 200, "["+rows+"]", nil).Fetch(context.Background(), "lever-mujininc", Validators{}); err == nil || result.Complete {
		t.Fatal("too many postings accepted")
	}
}

func TestNoArbitraryEndpoints(t *testing.T) {
	client := NewClient()
	if _, err := client.Fetch(context.Background(), "https://127.0.0.1/secrets", Validators{}); err == nil {
		t.Fatal("arbitrary source accepted")
	}
	for _, link := range []string{"http://jobs.lever.co/x", "https://jobs.lever.co@127.0.0.1/x", "https://jobs.lever.co.evil.example/x", "https://jobs.lever.co:444/x", "https://localhost/x", "javascript:alert(1)", "https://127.0.0.1/x"} {
		if SafeLink(link) {
			t.Fatalf("unsafe link accepted %s", link)
		}
	}
	if !SafeLink("https://jobs.lever.co/mujininc/abc/apply") {
		t.Fatal("safe employer link rejected")
	}
	request, _ := http.NewRequest("GET", "https://api.lever.co/elsewhere", nil)
	if err := client.http.CheckRedirect(request, nil); err != http.ErrUseLastResponse {
		t.Fatal("redirect allowed")
	}
	copy := Sources()
	copy[0].Endpoint = "http://localhost"
	if Sources()[0].Endpoint == copy[0].Endpoint {
		t.Fatal("sources mutated")
	}
}

func TestSalaryAndLocationUnknownPreserved(t *testing.T) {
	for _, fragment := range []string{`"salaryRange":null`, `"salaryRange":{"currency":"JPY","interval":"per year","min":-1,"max":0}`, `"salaryRange":{"currency":"JPY","interval":"per year","min":900,"max":100}`} {
		body := strings.Replace(japanPosting, `"salaryRange":{"currency":"JPY","interval":"per year","min":4000000,"max":7000000}`, fragment, 1)

		result, err := testClient(t, 200, "["+body+"]", nil).Fetch(context.Background(), "lever-mujininc", Validators{})
		if err != nil {
			t.Fatal(err)
		}
		if result.Jobs[0].SalaryMin != nil || result.Jobs[0].SalaryMax != nil {
			t.Fatalf("fabricated salary for %s", fragment)
		}
	}
	if prefecture("Japan") != "" || prefecture("Yokohama, Japan") != "Kanagawa" {
		t.Fatal("location normalization failed")
	}

}
