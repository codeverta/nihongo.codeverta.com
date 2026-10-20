package jobfetch

import (
	"math"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/html"
)

type leverPosting struct {
	ID               string `json:"id"`
	Text             string `json:"text"`
	Country          string `json:"country"`
	DescriptionPlain string `json:"descriptionPlain"`
	Description      string `json:"description"`
	AdditionalPlain  string `json:"additionalPlain"`
	Additional       string `json:"additional"`
	CreatedAt        int64  `json:"createdAt"`
	HostedURL        string `json:"hostedUrl"`
	ApplyURL         string `json:"applyUrl"`
	WorkplaceType    string `json:"workplaceType"`
	Categories       struct {
		Location     string   `json:"location"`
		AllLocations []string `json:"allLocations"`
		Commitment   string   `json:"commitment"`
		Team         string   `json:"team"`
		Department   string   `json:"department"`
	} `json:"categories"`
	Lists []struct {
		Text    string `json:"text"`
		Content string `json:"content"`
	} `json:"lists"`
	SalaryRange *struct {
		Currency string   `json:"currency"`
		Interval string   `json:"interval"`
		Min      *float64 `json:"min"`
		Max      *float64 `json:"max"`
	} `json:"salaryRange"`
	SalaryDescriptionPlain string `json:"salaryDescriptionPlain"`
}

func normalizeLever(source Source, p leverPosting) (Job, bool, bool) {
	// Only structured employer country evidence is accepted. A title mentioning
	// Japan or an unrestricted remote job is not evidence of Japan eligibility.
	if !strings.EqualFold(p.Country, "JP") {
		return Job{}, false, false
	}
	if strings.TrimSpace(p.ID) == "" || len(p.ID) > 128 || strings.ContainsAny(p.ID, "/\\") || strings.TrimSpace(p.Text) == "" || !SafeLink(p.HostedURL) || !SafeLink(p.ApplyURL) {
		return Job{}, false, true
	}
	expected := strings.TrimPrefix(source.ID, "lever-")
	if !strings.HasPrefix(p.HostedURL, "https://jobs.lever.co/"+expected+"/") || !strings.HasPrefix(p.ApplyURL, "https://jobs.lever.co/"+expected+"/") {
		return Job{}, false, true
	}
	job := Job{ExternalID: p.ID, Title: bounded(p.Text, 500), Company: source.Company, Location: bounded(p.Categories.Location, 500), Country: "JP", Category: category(p.Categories.Team, p.Categories.Department, p.Text), EmploymentType: employment(p.Categories.Commitment), Remote: "unknown", Visa: "unknown", JapaneseLevel: "unknown", SourceURL: p.HostedURL, ApplyURL: p.ApplyURL}
	job.Prefecture = prefecture(p.Categories.Location)
	description := p.DescriptionPlain
	if description == "" {
		description = plainHTML(p.Description)
	}
	var sections []string
	if p.Categories.Department != "" {
		sections = append(sections, "Department: "+p.Categories.Department)
	}
	if description != "" {
		sections = append(sections, description)
	}
	var qualifications []string
	var benefits []string
	for _, list := range p.Lists {
		text := strings.TrimSpace(plainHTML(list.Content))
		heading := strings.TrimSpace(plainHTML(list.Text))
		if text == "" {
			continue
		}
		section := heading + "\n" + text
		sections = append(sections, section)
		lower := strings.ToLower(heading)
		for _, key := range []string{"benefits", "perks", "福利厚生", "待遇"} {
			if strings.Contains(lower, key) {
				benefits = append(benefits, text)
				break
			}
		}
		for _, key := range []string{"qualif", "require", "skill", "応募", "必須", "歓迎", "求める", "必要"} {
			if strings.Contains(lower, key) {
				qualifications = append(qualifications, section)
				break
			}
		}
	}
	additional := p.AdditionalPlain
	if additional == "" {
		additional = plainHTML(p.Additional)
	}
	if additional != "" {
		sections = append(sections, additional)
	}
	job.Description = bounded(strings.Join(sections, "\n\n"), 80000)
	job.Qualifications = bounded(strings.Join(qualifications, "\n\n"), 30000)
	job.Benefits = bounded(strings.Join(benefits, "\n\n"), 30000)
	switch strings.ToLower(p.WorkplaceType) {
	case "onsite", "on-site":
		job.Remote = "onsite"
	case "hybrid":
		job.Remote = "hybrid"
	case "remote":
		job.Remote = "remote"
	}
	if p.CreatedAt > 0 {
		created := time.UnixMilli(p.CreatedAt).UTC()
		if created.Year() >= 2000 && created.Year() <= 2100 {
			job.PublishedAt = &created
		}
	}
	if salary := p.SalaryRange; salary != nil {
		if positive(salary.Min) {
			job.SalaryMin = salary.Min
		}
		if positive(salary.Max) {
			job.SalaryMax = salary.Max
		}
		if job.SalaryMin != nil || job.SalaryMax != nil {
			job.SalaryCurrency = bounded(strings.ToUpper(salary.Currency), 10)
			job.SalaryPeriod = salaryPeriod(salary.Interval)
		}
		if job.SalaryMin != nil && job.SalaryMax != nil && *job.SalaryMin > *job.SalaryMax {
			job.SalaryMin = nil
			job.SalaryMax = nil
			job.SalaryCurrency = ""
			job.SalaryPeriod = ""
		}
	}
	job.SalaryText = bounded(p.SalaryDescriptionPlain, 4000)
	// Free-form language/visa requirements remain in the original qualifications.
	// No sponsorship or JLPT grade is inferred from a brand or an incidental N2.
	return job, true, false
}

func positive(v *float64) bool { return v != nil && !math.IsNaN(*v) && !math.IsInf(*v, 0) && *v > 0 }
func salaryPeriod(s string) string {
	switch strings.ToLower(s) {
	case "per year", "year", "yearly", "annual":
		return "year"
	case "per month", "month", "monthly":
		return "month"
	case "per hour", "hour", "hourly":
		return "hour"
	case "per week", "week", "weekly":
		return "week"
	default:
		return "unknown"
	}
}
func employment(s string) string {
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "part-time"), strings.Contains(s, "part time"), strings.Contains(s, "アルバイト"):
		return "part-time"
	case strings.Contains(s, "intern"):
		return "internship"
	case strings.Contains(s, "contract"), strings.Contains(s, "契約"):
		return "contract"
	case strings.Contains(s, "full-time"), strings.Contains(s, "full time"), strings.Contains(s, "正社員"):
		return "full-time"
	default:
		return "unknown"
	}
}
func category(parts ...string) string {
	s := strings.ToLower(strings.Join(parts, " "))
	for _, entry := range []struct {
		category string
		keys     []string
	}{{"manufacturing", []string{"manufactur", "mechanical", "robot", "製造", "機械"}}, {"it", []string{"software", "engineer", "data", "security", "ソフト", "エンジニア"}}, {"sales", []string{"sales", "marketing", "business development", "営業"}}, {"administration", []string{"finance", "account", "legal", "human resources", "recruit", "corporate", "人事", "経理"}}, {"logistics", []string{"logistic", "warehouse", "supply chain", "物流"}}} {
		for _, key := range entry.keys {
			if strings.Contains(s, key) {
				return entry.category
			}
		}
	}
	return "other"
}

// Stable English prefecture names form filter values. Japanese and common city
// aliases are matched only inside the structured location field.
func prefecture(s string) string {
	lower := strings.ToLower(s)
	for _, pair := range [][2]string{
		{"Hokkaido", "北海道"}, {"Aomori", "青森"}, {"Iwate", "岩手"}, {"Miyagi", "宮城"},
		{"Akita", "秋田"}, {"Yamagata", "山形"}, {"Fukushima", "福島"}, {"Ibaraki", "茨城"},
		{"Tochigi", "栃木"}, {"Gunma", "群馬"}, {"Saitama", "埼玉"}, {"Chiba", "千葉"},
		{"Tokyo", "東京都"}, {"Kanagawa", "神奈川"}, {"Niigata", "新潟"}, {"Toyama", "富山"},
		{"Ishikawa", "石川"}, {"Fukui", "福井"}, {"Yamanashi", "山梨"}, {"Nagano", "長野"},
		{"Gifu", "岐阜"}, {"Shizuoka", "静岡"}, {"Aichi", "愛知"}, {"Mie", "三重"},
		{"Shiga", "滋賀"}, {"Kyoto", "京都"}, {"Osaka", "大阪"}, {"Hyogo", "兵庫"},
		{"Nara", "奈良"}, {"Wakayama", "和歌山"}, {"Tottori", "鳥取"}, {"Shimane", "島根"},
		{"Okayama", "岡山"}, {"Hiroshima", "広島"}, {"Yamaguchi", "山口"}, {"Tokushima", "徳島"},
		{"Kagawa", "香川"}, {"Ehime", "愛媛"}, {"Kochi", "高知"}, {"Fukuoka", "福岡"},
		{"Saga", "佐賀"}, {"Nagasaki", "長崎"}, {"Kumamoto", "熊本"}, {"Oita", "大分"},
		{"Miyazaki", "宮崎"}, {"Kagoshima", "鹿児島"}, {"Okinawa", "沖縄"},
	} {
		if strings.Contains(lower, strings.ToLower(pair[0])) || strings.Contains(s, pair[1]) {
			return pair[0]
		}
	}
	for _, pair := range [][2]string{{"Yokohama", "Kanagawa"}, {"Nagoya", "Aichi"}, {"Susono", "Shizuoka"}, {"Sapporo", "Hokkaido"}, {"Kobe", "Hyogo"}, {"Sendai", "Miyagi"}, {"Kawasaki", "Kanagawa"}} {
		if strings.Contains(lower, strings.ToLower(pair[0])) {
			return pair[1]
		}
	}
	return ""
}

func bounded(s string, limit int) string {
	s = strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s))
	runes := []rune(s)
	if len(runes) > limit {
		s = string(runes[:limit])
	}
	return s
}

func plainHTML(s string) string {
	node, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return ""
	}
	var b strings.Builder
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style" || n.Data == "iframe") {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		if n.Type == html.ElementNode && (n.Data == "br" || n.Data == "li" || n.Data == "p" || n.Data == "div") {
			b.WriteByte('\n')
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
		if n.Type == html.ElementNode && (n.Data == "li" || n.Data == "p" || n.Data == "div") {
			b.WriteByte('\n')
		}
	}
	visit(node)
	var lines []string
	for _, line := range strings.Split(b.String(), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
