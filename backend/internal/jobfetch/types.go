// Package jobfetch reads fixed public feeds and bounded, robots-aware samples
// of job-board listings. It never follows arbitrary URLs or uses private APIs.
package jobfetch

import "time"

type Source struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Company            string `json:"company"`
	Domain             string `json:"domain"`
	Endpoint           string `json:"endpoint"`
	Website            string `json:"website"`
	TermsURL           string `json:"termsUrl"`
	RequestBudget      int    `json:"requestBudget"`
	JobPathPrefix      string `json:"jobPathPrefix,omitempty"`
	JobPathContains    string `json:"jobPathContains,omitempty"`
	JobPathSuffix      string `json:"jobPathSuffix,omitempty"`
	JobPathMinSegments int    `json:"jobPathMinSegments,omitempty"`
	Category           string `json:"category,omitempty"`
}

type Job struct {
	ExternalID     string     `json:"externalId"`
	Title          string     `json:"title"`
	Company        string     `json:"company"`
	Description    string     `json:"description"`
	Qualifications string     `json:"qualifications"`
	Benefits       string     `json:"benefits"`
	Location       string     `json:"location"`
	Country        string     `json:"country"`
	Prefecture     string     `json:"prefecture"`
	Category       string     `json:"category"`
	Industry       string     `json:"industry"`
	EmploymentType string     `json:"employmentType"`
	Remote         string     `json:"remote"`
	Visa           string     `json:"visa"`
	JapaneseLevel  string     `json:"japaneseLevel"`
	SalaryMin      *float64   `json:"salaryMin"`
	SalaryMax      *float64   `json:"salaryMax"`
	SalaryCurrency string     `json:"salaryCurrency"`
	SalaryPeriod   string     `json:"salaryPeriod"`
	SalaryText     string     `json:"salaryText"`
	SourceURL      string     `json:"sourceUrl"`
	ApplyURL       string     `json:"applyUrl"`
	PublishedAt    *time.Time `json:"publishedAt"`
	ExpiresAt      *time.Time `json:"expiresAt"`
}

type Validators struct{ ETag, LastModified string }

type Result struct {
	Jobs        []Job
	Requests    int
	NotModified bool
	// Complete means the entire configured employer feed was successfully read;
	// absent jobs may be retired. It never means every job in Japan is covered.
	Complete     bool
	ETag         string
	LastModified string
	RetryAfter   time.Time
	Blocked      bool
	Unsupported  bool
}
