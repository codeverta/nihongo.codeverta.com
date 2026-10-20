// Package news owns imported, attributed news and its durable calendar scheduler.
package news

import "time"

type Token struct {
	Start         int    `json:"start"` // UTF-16 code units, matching Dart String offsets.
	End           int    `json:"end"`
	Surface       string `json:"surface"`
	Base          string `json:"base"`
	Reading       string `json:"reading"`
	ReadingSource string `json:"readingSource"`
}
type Paragraph struct {
	Text   string  `json:"text"`
	Kind   string  `json:"kind"`
	Tokens []Token `json:"tokens"`
}
type Credit struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Role string `json:"role"`
}
type Article struct {
	ID              string           `json:"id" gorm:"primaryKey;size:36"`
	SourceID        string           `json:"sourceId" gorm:"size:40;index"`
	SourceName      string           `json:"sourceName"`
	Kind            string           `json:"kind"`
	URL             string           `json:"url" gorm:"size:512;uniqueIndex"`
	Title           string           `json:"title" gorm:"type:text"`
	TitleTokens     []Token          `json:"titleTokens" gorm:"serializer:json;type:json"`
	Summary         string           `json:"summary" gorm:"type:text"`
	Paragraphs      []Paragraph      `json:"paragraphs,omitempty" gorm:"serializer:json;type:json"`
	Credits         []Credit         `json:"credits" gorm:"serializer:json;type:json"`
	References      []Credit         `json:"references,omitempty" gorm:"serializer:json;type:json"`
	License         string           `json:"license"`
	LicenseURL      string           `json:"licenseUrl"`
	Changes         string           `json:"changes"`
	PublishedAt     time.Time        `json:"publishedAt" gorm:"index"`
	OriginalDate    string           `json:"originalDate" gorm:"-"`
	DateType        string           `json:"dateType"`
	FetchedAt       time.Time        `json:"fetchedAt"`
	ContentHash     string           `json:"revision" gorm:"size:64"`
	ImageURL        string           `json:"imageUrl"`
	ImageCredit     string           `json:"imageCredit"`
	ImageLicense    string           `json:"imageLicense"`
	ImageSourceURL  string           `json:"imageSourceUrl"`
	ImageLicenseURL string           `json:"imageLicenseUrl"`
	imageCandidates []imageCandidate `gorm:"-"`
	Hidden          bool             `json:"hidden"`
	Deleted         bool             `json:"-"`
}

func (Article) TableName() string { return "news_articles" }

type Source struct {
	Status        string     `json:"status" gorm:"-"`
	ID            string     `json:"id" gorm:"primaryKey;size:40"`
	Name          string     `json:"name"`
	Kind          string     `json:"kind"`
	URL           string     `json:"url"`
	Timezone      string     `json:"timezone"`
	Days          []int      `json:"days" gorm:"serializer:json;type:json"` // Sunday = 0.
	Hour          int        `json:"hour"`
	Minute        int        `json:"minute"`
	Enabled       bool       `json:"enabled"`
	NextFetchAt   time.Time  `json:"nextFetchAt"`
	LastStartedAt *time.Time `json:"lastStartedAt"`
	LastSuccessAt *time.Time `json:"lastSuccessAt"`
	LastError     string     `json:"lastError" gorm:"type:text"`
	LastImported  int        `json:"lastImported"`
	LastSkipped   int        `json:"lastSkipped"`
	ETag          string     `json:"-"`
	LastModified  string     `json:"-"`
	LeaseOwner    string     `json:"-" gorm:"size:36"`
	LeaseUntil    *time.Time `json:"-"`
}

func (Source) TableName() string { return "news_sources" }

type Schedule struct {
	Timezone string `json:"timezone"`
	Days     []int  `json:"days"`
	Hour     int    `json:"hour"`
	Minute   int    `json:"minute"`
	Enabled  bool   `json:"enabled"`
}
type Page struct {
	Items    []Article `json:"items"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
	Sources  []Source  `json:"sources"`
}
