package jobfetch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxBytes = 12 << 20
const maxPostings = 2000

var sources = []Source{
	{ID: "lever-mujininc", Name: "Mujin · Karier resmi", Company: "Mujin", Domain: "api.lever.co", Endpoint: "https://api.lever.co/v0/postings/mujininc?mode=json", Website: "https://jobs.lever.co/mujininc", TermsURL: "https://github.com/lever/postings-api"},
	{ID: "lever-woven-by-toyota", Name: "Woven by Toyota · Karier resmi", Company: "Woven by Toyota", Domain: "api.lever.co", Endpoint: "https://api.lever.co/v0/postings/woven-by-toyota?mode=json", Website: "https://jobs.lever.co/woven-by-toyota", TermsURL: "https://github.com/lever/postings-api"},
	{ID: "smartrecruiters-accor-japan", Name: "Accor · Karier Jepang resmi", Company: "Accor", Domain: "api.smartrecruiters.com", Endpoint: "https://api.smartrecruiters.com/v1/companies/AccorHotel/postings?limit=100&country=jp&destination=PUBLIC", Website: "https://careers.accor.com/", TermsURL: "https://developers.smartrecruiters.com/docs/posting-api"},
	{ID: "tokyodev", Name: "TokyoDev · lowongan teknologi", Company: "TokyoDev employers", Domain: "www.tokyodev.com", Endpoint: "https://www.tokyodev.com/", Website: "https://www.tokyodev.com/", RequestBudget: 4, JobPathPrefix: "/companies/", JobPathContains: "/jobs/"},
	{ID: "japan-dev", Name: "Japan Dev · lowongan teknologi", Company: "Japan Dev employers", Domain: "japan-dev.com", Endpoint: "https://japan-dev.com/jobs", Website: "https://japan-dev.com/", RequestBudget: 4, JobPathPrefix: "/jobs/", JobPathMinSegments: 3},
	{ID: "jobs-in-japan", Name: "Jobs in Japan · lowongan umum", Company: "Jobs in Japan employers", Domain: "jobsinjapan.com", Endpoint: "https://jobsinjapan.com/jobs/", Website: "https://jobsinjapan.com/", RequestBudget: 4, JobPathPrefix: "/jobs/"},
	{ID: "findy", Name: "Findy · lowongan teknologi", Company: "Findy employers", Domain: "findy-code.io", Endpoint: "https://findy-code.io/", Website: "https://findy-code.io/", RequestBudget: 4, JobPathPrefix: "/companies/", JobPathSuffix: "/jobs"},
	{ID: "lapras", Name: "LAPRAS · lowongan teknologi", Company: "LAPRAS employers", Domain: "lapras.com", Endpoint: "https://lapras.com/", Website: "https://lapras.com/", RequestBudget: 4, JobPathPrefix: "/jobs/"},
	{ID: "hireplanner", Name: "HirePlanner · lowongan umum", Company: "HirePlanner employers", Domain: "www.hireplanner.com", Endpoint: "https://www.hireplanner.com/", Website: "https://www.hireplanner.com/", RequestBudget: 4, JobPathPrefix: "/jobs/"},
	{ID: "agri-navi", Name: "AgriNavi · lowongan agrikultur", Company: "AgriNavi employers", Domain: "www.agri-navi.com", Endpoint: "https://www.agri-navi.com/", Website: "https://www.agri-navi.com/", RequestBudget: 4, JobPathPrefix: "/jobdetail/", Category: "agriculture"},
}

func Sources() []Source {
	out := append([]Source(nil), sources...)
	for i := range out {
		if out[i].RequestBudget == 0 {
			out[i].RequestBudget = 1
		}
	}
	return out
}

type Client struct {
	http         *http.Client
	requestDelay time.Duration
}

func NewClient() *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil // fixed public endpoints must not inherit a private proxy
	transport.ResponseHeaderTimeout = 15 * time.Second
	transport.MaxConnsPerHost = 1
	return &Client{requestDelay: 3 * time.Second, http: &http.Client{Transport: transport, Timeout: 35 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// A feed redirect changes the reviewed endpoint: stop and require review.
		return http.ErrUseLastResponse
	}}}
}

func (c *Client) Fetch(ctx context.Context, sourceID string, v Validators) (Result, error) {
	var source Source
	for _, item := range sources {
		if item.ID == sourceID {
			source = item
			break
		}
	}
	if source.ID == "" {
		return Result{}, errors.New("unknown job source")
	}
	if source.RequestBudget == 0 {
		source.RequestBudget = 1
	}
	if source.RequestBudget > 1 {
		return c.FetchPublicBoard(ctx, source)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.Endpoint, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "NihongoMasterJobSync/1.0 (+https://github.com/lever/postings-api; weekly employer feed cache)")
	if validHeader(v.ETag) {
		req.Header.Set("If-None-Match", v.ETag)
	}
	if validHeader(v.LastModified) {
		req.Header.Set("If-Modified-Since", v.LastModified)
	}
	response, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("job feed request failed: %w", err)
	}
	defer response.Body.Close()
	result := Result{Requests: 1, ETag: cleanHeader(response.Header.Get("ETag")), LastModified: cleanHeader(response.Header.Get("Last-Modified"))}
	result.RetryAfter = retryAfter(response.Header.Get("Retry-After"), time.Now().UTC())
	if response.StatusCode == http.StatusNotModified {
		result.NotModified = true
		return result, nil
	}
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusTooManyRequests {
		result.Blocked = true
		return result, fmt.Errorf("job feed access paused: HTTP %d", response.StatusCode)
	}
	if response.StatusCode != http.StatusOK {
		return result, fmt.Errorf("job feed HTTP %d", response.StatusCode)
	}
	if response.ContentLength > maxBytes {
		return result, errors.New("job feed exceeds byte limit")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return result, err
	}
	if len(body) > maxBytes {
		return result, errors.New("job feed exceeds byte limit")
	}
	if source.ID == "smartrecruiters-accor-japan" {
		return parseAccor(body, result)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(body)), "[") {
		return result, errors.New("unexpected job feed format")
	}
	var rows []leverPosting
	if err = json.Unmarshal(body, &rows); err != nil {
		return result, fmt.Errorf("invalid job feed: %w", err)
	}
	if len(rows) > maxPostings {
		return result, errors.New("job feed exceeds posting limit")
	}
	result.Jobs = make([]Job, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		job, ok, invalid := normalizeLever(source, row)
		if invalid {
			return result, errors.New("job feed contains invalid posting; snapshot not applied")
		}
		if !ok {
			continue
		}
		if seen[job.ExternalID] {
			return result, errors.New("duplicate job identifier")
		}
		seen[job.ExternalID] = true
		result.Jobs = append(result.Jobs, job)
	}
	result.Complete = true
	return result, nil
}

func validHeader(s string) bool { return s != "" && len(s) <= 1024 && !strings.ContainsAny(s, "\r\n") }
func cleanHeader(s string) string {
	if validHeader(s) {
		return s
	}
	return ""
}
func retryAfter(s string, now time.Time) time.Time {
	if seconds, err := strconv.ParseInt(s, 10, 64); err == nil && seconds >= 0 {
		if seconds > 365*86400 {
			seconds = 365 * 86400
		}
		return now.Add(time.Duration(seconds) * time.Second)
	}
	if parsed, err := http.ParseTime(s); err == nil && parsed.After(now) {
		return parsed.UTC()
	}
	return time.Time{}
}

// SafeLink only allows ordinary public HTTPS links. Collectors never fetch these
// links; the mobile client opens the original application website externally.
func SafeLink(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Host == "" || u.Port() != "" || len(raw) > 2048 || strings.ContainsAny(raw, "\r\n\\") {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if !strings.Contains(host, ".") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".internal") {
		return false
	}
	for _, r := range host {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '.' || r == '-') {
			return false
		}
	}
	return (host == "jobs.lever.co" || host == "jobs.smartrecruiters.com" || host == "www.tokyodev.com" || host == "japan-dev.com" || host == "jobsinjapan.com" || host == "findy-code.io" || host == "lapras.com" || host == "www.hireplanner.com" || host == "www.agri-navi.com") && strings.HasPrefix(u.Path, "/")
}
