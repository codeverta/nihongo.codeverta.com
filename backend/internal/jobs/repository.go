package jobs

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/jobfetch"
)

//go:embed resources/sources.json
var directoryJSON []byte

var ErrInvalid = errors.New("invalid job filter")
var ErrNotFound = errors.New("job not found")
var ErrLeaseLost = errors.New("job synchronization lease no longer owned")

const StaleAfter = 30 * 24 * time.Hour
const MaxRequestsPerDomain = 4

type Repository struct {
	db             *gorm.DB
	interval       time.Duration
	now            func() time.Time
	sources        []jobfetch.Source
	directory      []Source
	requestSpacing time.Duration
}

func New(db *gorm.DB, interval time.Duration) (*Repository, error) {
	if db == nil {
		return nil, errors.New("jobs require MariaDB")
	}
	if interval != 168*time.Hour && interval != 84*time.Hour {
		return nil, errors.New("jobs interval must be 168 or 84 hours")
	}
	if err := database.Migrate(db, &listingRow{}, &observationRow{}, &sourceRow{}, &domainRow{}, &runRow{}); err != nil {
		return nil, err
	}
	directory, err := DecodeDirectory(directoryJSON)
	if err != nil {
		return nil, err
	}
	r := &Repository{db: db, interval: interval, now: time.Now, sources: jobfetch.Sources(), directory: directory, requestSpacing: 3 * time.Second}
	groups := map[string]int{}
	for _, s := range r.sources {
		groups[s.Domain] += requestBudget(s)
		if groups[s.Domain] > MaxRequestsPerDomain {
			return nil, fmt.Errorf("job source domain %s exceeds weekly request budget", s.Domain)
		}
		if err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&sourceRow{ID: stableID("source:" + s.ID), Code: s.ID, Domain: s.Domain, Status: "idle"}).Error; err != nil {
			return nil, err
		}
		if err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&domainRow{ID: stableID("domain:" + s.Domain), Domain: s.Domain, NextFetchAt: time.Unix(0, 0).UTC()}).Error; err != nil {
			return nil, err
		}
	}
	return r, nil
}
func requestBudget(s jobfetch.Source) int {
	if s.RequestBudget > 0 {
		return s.RequestBudget
	}
	return 1
}
func stableID(s string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("nihongo-master:jobs:"+s)).String()
}
func hash(s string) string { sum := sha256.Sum256([]byte(s)); return hex.EncodeToString(sum[:]) }

// Links are returned to an external browser, never fetched by this repository.
func PublicHTTPS(value string) bool {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || len(value) > 4096 || u.Opaque != "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if strings.ContainsAny(value, "\r\n\x00") || host == "localhost" || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") || !strings.Contains(host, ".") {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsUnspecified() && !ip.IsLinkLocalUnicast() && !ip.IsMulticast()
	}
	return u.Port() == "" || u.Port() == "443"
}
func canonical(value string) string {
	u, _ := url.Parse(value)
	u.Fragment = ""
	u.Scheme = "https"
	u.Host = strings.ToLower(u.Host)
	u.Path = strings.TrimRight(u.Path, "/")
	q := u.Query()
	for key := range q {
		if strings.HasPrefix(strings.ToLower(key), "utm_") || key == "source" || key == "ref" || key == "lever-source" {
			q.Del(key)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
func short(s string, n int) string {
	r := []rune(strings.TrimSpace(strings.ReplaceAll(s, "\x00", "")))
	if len(r) > n {
		r = r[:n]
	}
	return string(r)
}
func normalJob(source jobfetch.Source, j jobfetch.Job, at time.Time) (listingRow, error) {
	if j.Country != "JP" || strings.TrimSpace(j.Title) == "" || strings.TrimSpace(j.Company) == "" || len(j.ExternalID) > 200 || j.ExternalID == "" || !PublicHTTPS(j.SourceURL) || !PublicHTTPS(j.ApplyURL) {
		return listingRow{}, ErrInvalid
	}
	if (j.SalaryMin != nil && (!finite(*j.SalaryMin) || *j.SalaryMin < 0 || *j.SalaryMin > 1e12)) || (j.SalaryMax != nil && (!finite(*j.SalaryMax) || *j.SalaryMax < 0 || *j.SalaryMax > 1e12)) {
		return listingRow{}, ErrInvalid
	}
	if j.SalaryMin != nil && j.SalaryMax != nil && *j.SalaryMin > *j.SalaryMax {
		return listingRow{}, ErrInvalid
	}
	prefecture := j.Prefecture
	for _, option := range Prefectures {
		if strings.EqualFold(prefecture, option.Label) {
			prefecture = option.ID
			break
		}
	}
	if !allowed(prefecture, Prefectures) {
		prefecture = ""
	}
	sector := j.Industry
	if !allowed(sector, Sectors) {
		sector = ""
	}
	employment := j.EmploymentType
	if !allowed(employment, EmploymentTypes) {
		employment = "unknown"
	}
	remote := j.Remote
	if !allowed(remote, Metadata()["remoteOptions"].([]Option)) {
		remote = "unknown"
	}
	visa := j.Visa
	if !allowed(visa, VisaOptions) {
		visa = "unknown"
	}
	japanese := j.JapaneseLevel
	if !allowed(japanese, JapaneseLevels) {
		japanese = "unknown"
	}
	period := j.SalaryPeriod
	if period != "hour" && period != "day" && period != "week" && period != "month" && period != "year" {
		period = "unknown"
	}
	sortAt := at.UnixMicro()
	if j.PublishedAt != nil && !j.PublishedAt.After(at.Add(24*time.Hour)) {
		sortAt = j.PublishedAt.UnixMicro()
	}
	key := hash(canonical(j.ApplyURL))
	status := "active"
	if j.ExpiresAt != nil && !j.ExpiresAt.After(at) {
		status = "expired"
	}
	return listingRow{ID: stableID("listing:" + key), CanonicalKey: key, Title: short(j.Title, 512), Company: short(j.Company, 255), Category: CanonicalCategory(j.Category), Sector: sector, Prefecture: prefecture, Location: short(j.Location, 512), Country: "JP", EmploymentType: employment, Remote: remote, Visa: visa, JapaneseLevel: japanese, SalaryMin: j.SalaryMin, SalaryMax: j.SalaryMax, SalaryCurrency: short(strings.ToUpper(j.SalaryCurrency), 8), SalaryPeriod: period, SalaryText: short(j.SalaryText, 1024), Description: short(j.Description, 100000), Qualifications: short(j.Qualifications, 60000), Benefits: short(j.Benefits, 30000), SourceID: source.ID, SourceName: source.Name, SourceURL: j.SourceURL, ApplyURL: j.ApplyURL, PostedAt: j.PublishedAt, ExpiresAt: j.ExpiresAt, FirstSeenAt: at, LastSeenAt: at, SortAt: sortAt, Status: status}, nil
}

type Filter struct {
	Q, Category, Sector, Prefecture, EmploymentType, JapaneseLevel, Visa, Remote, SalaryPeriod, Cursor string
	SalaryMin                                                                                          *float64
	Limit                                                                                              int
}
type Page struct {
	Items      []Job  `json:"items"`
	Total      int64  `json:"total"`
	NextCursor string `json:"nextCursor"`
	HasMore    bool   `json:"hasMore"`
}
type cursor struct {
	At     int64  `json:"at"`
	ID     string `json:"id"`
	Filter string `json:"filter"`
}

func (f Filter) Validate() error {
	if len([]rune(f.Q)) > 160 || len(f.Cursor) > 512 || f.Limit < 1 || f.Limit > 30 {
		return ErrInvalid
	}
	for _, p := range []struct {
		v string
		o []Option
	}{{f.Category, Categories}, {f.Sector, Sectors}, {f.Prefecture, Prefectures}, {f.EmploymentType, EmploymentTypes}, {f.JapaneseLevel, JapaneseLevels}, {f.Visa, VisaOptions}, {f.Remote, Metadata()["remoteOptions"].([]Option)}} {
		if p.v != "" && !allowed(p.v, p.o) {
			return ErrInvalid
		}
	}
	if f.SalaryMin != nil {
		if !finite(*f.SalaryMin) || *f.SalaryMin < 0 || *f.SalaryMin > 1e12 {
			return ErrInvalid
		}
		if f.SalaryPeriod != "hour" && f.SalaryPeriod != "day" && f.SalaryPeriod != "week" && f.SalaryPeriod != "month" && f.SalaryPeriod != "year" {
			return ErrInvalid
		}
	}
	return nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func filterKey(f Filter) string {
	f.Cursor = ""
	f.Limit = 0
	raw, _ := json.Marshal(f)
	return hash(string(raw))
}
func (r *Repository) List(ctx context.Context, f Filter) (Page, error) {
	out := Page{Items: []Job{}}
	if err := f.Validate(); err != nil {
		return out, err
	}
	now := r.now().UTC()
	q := r.db.WithContext(ctx).Model(&listingRow{}).Where("status = ? AND last_seen_at >= ? AND (expires_at IS NULL OR expires_at > ?)", "active", now.Add(-StaleAfter), now)
	for _, p := range []struct{ k, v string }{{"category", f.Category}, {"sector", f.Sector}, {"prefecture", f.Prefecture}, {"employment_type", f.EmploymentType}, {"japanese_level", f.JapaneseLevel}, {"visa", f.Visa}, {"remote", f.Remote}} {
		if p.v != "" {
			q = q.Where(p.k+" = ?", p.v)
		}
	}
	if strings.TrimSpace(f.Q) != "" {
		needle := "%" + strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(strings.ToLower(strings.TrimSpace(f.Q))) + "%"
		q = q.Where("(LOWER(title) LIKE ? OR LOWER(company) LIKE ? OR LOWER(location) LIKE ?)", needle, needle, needle)
	}
	if f.SalaryMin != nil {
		q = q.Where("salary_currency = ? AND salary_period = ? AND COALESCE(salary_max,salary_min) >= ?", "JPY", f.SalaryPeriod, *f.SalaryMin)
	}
	if err := q.Count(&out.Total).Error; err != nil {
		return out, err
	}
	if f.Cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(f.Cursor)
		if err != nil {
			return out, ErrInvalid
		}
		var c cursor
		if json.Unmarshal(raw, &c) != nil || uuid.Validate(c.ID) != nil || c.Filter != filterKey(f) || c.At <= 0 {
			return out, ErrInvalid
		}
		q = q.Where("(sort_at < ? OR (sort_at = ? AND id < ?))", c.At, c.At, c.ID)
	}
	var rows []listingRow
	if err := q.Order("sort_at DESC").Order("id DESC").Limit(f.Limit + 1).Find(&rows).Error; err != nil {
		return out, err
	}
	if len(rows) > f.Limit {
		out.HasMore = true
		rows = rows[:f.Limit]
	}
	for _, row := range rows {
		j := view(row)
		j.Description = ""
		j.Qualifications = ""
		j.Benefits = ""
		out.Items = append(out.Items, j)
	}
	if out.HasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		raw, _ := json.Marshal(cursor{last.SortAt, last.ID, filterKey(f)})
		out.NextCursor = base64.RawURLEncoding.EncodeToString(raw)
	}
	return out, nil
}
func (r *Repository) Get(ctx context.Context, id string) (Job, error) {
	if uuid.Validate(id) != nil {
		return Job{}, ErrNotFound
	}
	var row listingRow
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Job{}, ErrNotFound
	}
	if err != nil {
		return Job{}, err
	}
	now := r.now().UTC()
	if row.Status == "active" {
		if row.ExpiresAt != nil && !row.ExpiresAt.After(now) {
			row.Status = "expired"
		} else if row.LastSeenAt.Before(now.Add(-StaleAfter)) {
			row.Status = "stale"
		}
	}
	return view(row), nil
}
func (r *Repository) Sources(ctx context.Context) ([]Source, error) {
	out := append([]Source{}, r.directory...)
	for i := range out {
		out[i].Mode = "directory"
		out[i].Status = "directory_only"
		out[i].IntervalHours = 0
	}
	var states []sourceRow
	var domains []domainRow
	if err := r.db.WithContext(ctx).Find(&states).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Find(&domains).Error; err != nil {
		return nil, err
	}
	for _, s := range r.sources {
		v := Source{ID: s.ID, Name: s.Name, URL: s.Website, Mode: "api", Status: "idle", TermsURL: s.TermsURL, Scope: "Lowongan Jepang pada perusahaan ini", IntervalHours: int(r.interval / time.Hour)}
		if s.RequestBudget > 1 {
			v.Scope = "Sampel lowongan publik Jepang"
			v.Reason = "Pengambilan terbatas setiap minggu; lihat situs sumber untuk seluruh lowongan."
		}
		for i, entry := range out {
			if entry.ID == s.ID {
				v.Category = entry.Category
				out = append(out[:i], out[i+1:]...)
				break
			}
		}
		for _, state := range states {
			if state.Code == s.ID {
				v.Status = state.Status
				v.LastSuccessAt = state.LastSuccessAt
				switch state.Status {
				case "unsupported":
					v.Reason = "Halaman publik belum menyediakan data lowongan yang dapat dibaca. Buka situs sumber untuk melihat seluruh lowongan."
				case "blocked":
					v.Reason = "Situs membatasi akses otomatis. Pengambilan dihentikan sampai jadwal berikutnya."
				case "error":
					v.Reason = "Pembaruan terakhir belum berhasil. Data tersimpan tetap tersedia selama masih berlaku."
				}
			}
		}
		for _, d := range domains {
			if d.Domain == s.Domain && d.LastStartedAt != nil {
				t := d.NextFetchAt
				v.NextFetchAt = &t
			}
		}
		out = append(out, v)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Mode != out[j].Mode {
			return out[i].Mode == "api"
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}
func ParseFilter(values url.Values) (Filter, error) {
	f := Filter{Q: strings.TrimSpace(values.Get("q")), Category: values.Get("category"), Sector: values.Get("sector"), Prefecture: values.Get("prefecture"), EmploymentType: values.Get("employmentType"), JapaneseLevel: values.Get("japaneseLevel"), Visa: values.Get("visa"), Remote: values.Get("remote"), SalaryPeriod: values.Get("salaryPeriod"), Cursor: values.Get("cursor"), Limit: 20}
	if v := values.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return f, ErrInvalid
		}
		f.Limit = n
	}
	if v := values.Get("salaryMin"); v != "" {
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return f, ErrInvalid
		}
		f.SalaryMin = &n
	}
	return f, f.Validate()
}
