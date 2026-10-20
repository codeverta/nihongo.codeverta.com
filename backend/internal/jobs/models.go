// Package jobs serves a durable local catalog. Reads never contact job websites.
package jobs

import "time"

type listingRow struct {
	ID             string     `gorm:"type:char(36);primaryKey"`
	CanonicalKey   string     `gorm:"type:char(64);not null;uniqueIndex"`
	Title          string     `gorm:"type:varchar(512);not null"`
	Company        string     `gorm:"type:varchar(255);not null;index"`
	Category       string     `gorm:"type:varchar(40);not null;index"`
	Sector         string     `gorm:"type:varchar(64);index"`
	Prefecture     string     `gorm:"type:varchar(2);index"`
	Location       string     `gorm:"type:varchar(512)"`
	Country        string     `gorm:"type:char(2);not null"`
	EmploymentType string     `gorm:"type:varchar(32);index"`
	Remote         string     `gorm:"type:varchar(16);index"`
	Visa           string     `gorm:"type:varchar(32);index"`
	JapaneseLevel  string     `gorm:"type:varchar(20);index"`
	SalaryMin      *float64   `gorm:"type:decimal(16,2)"`
	SalaryMax      *float64   `gorm:"type:decimal(16,2)"`
	SalaryCurrency string     `gorm:"type:varchar(8)"`
	SalaryPeriod   string     `gorm:"type:varchar(16)"`
	SalaryText     string     `gorm:"type:varchar(1024)"`
	Description    string     `gorm:"type:longtext"`
	Qualifications string     `gorm:"type:longtext"`
	Benefits       string     `gorm:"type:longtext"`
	SourceID       string     `gorm:"type:varchar(80);not null;index"`
	SourceName     string     `gorm:"type:varchar(255)"`
	SourceURL      string     `gorm:"type:text"`
	ApplyURL       string     `gorm:"type:text"`
	PostedAt       *time.Time `gorm:"type:datetime(6)"`
	ExpiresAt      *time.Time `gorm:"type:datetime(6);index"`
	FirstSeenAt    time.Time  `gorm:"type:datetime(6);not null"`
	LastSeenAt     time.Time  `gorm:"type:datetime(6);not null;index"`
	SortAt         int64      `gorm:"not null;index:job_feed_sort,priority:2"`
	Status         string     `gorm:"type:varchar(16);not null;index:job_feed_sort,priority:1"`
}

func (listingRow) TableName() string { return "job_listings" }

type observationRow struct {
	ID         string    `gorm:"type:char(36);primaryKey"`
	SourceID   string    `gorm:"type:varchar(80);not null;uniqueIndex:job_source_external,priority:1;index"`
	ExternalID string    `gorm:"type:varchar(200);not null;uniqueIndex:job_source_external,priority:2"`
	ListingID  string    `gorm:"type:char(36);not null;index"`
	LastSeenAt time.Time `gorm:"type:datetime(6);not null"`
	Active     bool      `gorm:"not null;index"`
}

func (observationRow) TableName() string { return "job_observations" }

type sourceRow struct {
	ID                string     `gorm:"type:char(36);primaryKey"`
	Code              string     `gorm:"type:varchar(80);not null;uniqueIndex"`
	Domain            string     `gorm:"type:varchar(255);index"`
	ETag              string     `gorm:"column:etag;type:varchar(512)"`
	LastModified      string     `gorm:"type:varchar(128)"`
	LastAttemptAt     *time.Time `gorm:"type:datetime(6)"`
	LastSuccessAt     *time.Time `gorm:"type:datetime(6)"`
	Status            string     `gorm:"type:varchar(32)"`
	LastError         string     `gorm:"type:varchar(512)"`
	ConsecutiveErrors int
}

func (sourceRow) TableName() string { return "job_sources" }

type domainRow struct {
	ID            string     `gorm:"type:char(36);primaryKey"`
	Domain        string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	NextFetchAt   time.Time  `gorm:"type:datetime(6);not null;index"`
	LastStartedAt *time.Time `gorm:"type:datetime(6)"`
	LeaseOwner    string     `gorm:"type:char(36)"`
	LeaseUntil    *time.Time `gorm:"type:datetime(6)"`
}

func (domainRow) TableName() string { return "job_sync_domains" }

type runRow struct {
	ID         string     `gorm:"type:char(36);primaryKey"`
	Domain     string     `gorm:"type:varchar(255);index"`
	StartedAt  time.Time  `gorm:"type:datetime(6);not null;index"`
	FinishedAt *time.Time `gorm:"type:datetime(6)"`
	Requests   int
	JobsSeen   int
	Status     string `gorm:"type:varchar(32)"`
	Error      string `gorm:"type:varchar(512)"`
}

func (runRow) TableName() string { return "job_sync_runs" }

type Job struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Company        string     `json:"company"`
	Category       string     `json:"category"`
	Sector         string     `json:"sector"`
	Prefecture     string     `json:"prefecture"`
	Location       string     `json:"location"`
	Country        string     `json:"country"`
	EmploymentType string     `json:"employmentType"`
	Remote         string     `json:"remote"`
	Visa           string     `json:"visa"`
	JapaneseLevel  string     `json:"japaneseLevel"`
	SalaryMin      *float64   `json:"salaryMin"`
	SalaryMax      *float64   `json:"salaryMax"`
	SalaryCurrency string     `json:"salaryCurrency"`
	SalaryPeriod   string     `json:"salaryPeriod"`
	SalaryText     string     `json:"salaryText"`
	Description    string     `json:"description"`
	Qualifications string     `json:"qualifications"`
	Benefits       string     `json:"benefits"`
	SourceID       string     `json:"sourceId"`
	SourceName     string     `json:"sourceName"`
	SourceURL      string     `json:"sourceUrl"`
	ApplyURL       string     `json:"applyUrl"`
	PostedAt       *time.Time `json:"postedAt"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	LastSeenAt     time.Time  `json:"lastSeenAt"`
	Status         string     `json:"status"`
}

func view(r listingRow) Job {
	return Job{r.ID, r.Title, r.Company, r.Category, r.Sector, r.Prefecture, r.Location, r.Country, r.EmploymentType, r.Remote, r.Visa, r.JapaneseLevel, r.SalaryMin, r.SalaryMax, r.SalaryCurrency, r.SalaryPeriod, r.SalaryText, r.Description, r.Qualifications, r.Benefits, r.SourceID, r.SourceName, r.SourceURL, r.ApplyURL, r.PostedAt, r.ExpiresAt, r.LastSeenAt, r.Status}
}

type Option struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
type Source struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	Category      string     `json:"category"`
	Scope         string     `json:"scope"`
	Mode          string     `json:"mode"`
	Status        string     `json:"status"`
	TermsURL      string     `json:"termsUrl"`
	Reason        string     `json:"reason"`
	LastSuccessAt *time.Time `json:"lastSuccessAt"`
	NextFetchAt   *time.Time `json:"nextFetchAt"`
	IntervalHours int        `json:"intervalHours"`
}
