// Package manga imports a curated Japanese-only catalog. Public reads never
// contact a publisher; page bytes are served exclusively from our COS cache.
package manga

import "time"

const RefreshInterval = 7 * 24 * time.Hour

type Series struct {
	ID                string     `json:"id" gorm:"primaryKey;size:64"`
	URL               string     `json:"sourceUrl" gorm:"type:text"`
	Title             string     `json:"title"`
	JapaneseTitle     string     `json:"japaneseTitle"`
	Summary           string     `json:"summary" gorm:"type:text"`
	Author            string     `json:"author"`
	Genres            []string   `json:"genres" gorm:"serializer:json;type:json"`
	CoverURL          string     `json:"coverUrl" gorm:"type:text"`
	Language          string     `json:"language" gorm:"size:8"`
	PublicationStatus string     `json:"publicationStatus"`
	Status            string     `json:"status" gorm:"size:20;index"`
	Enabled           bool       `json:"enabled"`
	Reviewed          bool       `json:"reviewed"`
	ChapterCount      int        `json:"chapterCount"`
	FetchedAt         *time.Time `json:"fetchedAt"`
	NextFetchAt       time.Time  `json:"nextFetchAt" gorm:"index"`
	LastError         string     `json:"-" gorm:"type:text"`
}

func (Series) TableName() string { return "manga_series" }

type Page struct {
	Index  int    `json:"index"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Chapter struct {
	ID            string     `json:"id" gorm:"primaryKey;size:64"`
	SeriesID      string     `json:"seriesId" gorm:"size:64;index"`
	URL           string     `json:"sourceUrl" gorm:"type:text"`
	Title         string     `json:"title"`
	Number        string     `json:"number" gorm:"size:32"`
	Position      float64    `json:"-"`
	Status        string     `json:"status" gorm:"size:20;index"`
	Pages         []Page     `json:"pages" gorm:"serializer:json;type:json"`
	SourcePages   []string   `json:"-" gorm:"serializer:json;type:json"`
	FailureCount  int        `json:"-"`
	RequestedAt   *time.Time `json:"-" gorm:"index"`
	NextAttemptAt time.Time  `json:"retryAt"`
	FetchedAt     *time.Time `json:"fetchedAt"`
	LastError     string     `json:"-" gorm:"type:text"`
}

func (Chapter) TableName() string { return "manga_chapters" }

type Asset struct {
	ID            string `gorm:"primaryKey;size:64"`
	SourceURL     string `gorm:"type:text"`
	URL           string `gorm:"type:text"`
	ObjectKey     string `gorm:"size:160"`
	Width, Height int
	Bytes         int
	CreatedAt     time.Time
}

func (Asset) TableName() string { return "manga_assets" }

type syncState struct {
	ID            string `gorm:"primaryKey;size:40"`
	LeaseOwner    string `gorm:"size:36"`
	LeaseUntil    time.Time
	NextRunAt     time.Time
	LastSuccessAt *time.Time
}

func (syncState) TableName() string { return "manga_sync_state" }

type Catalog struct {
	Items        []Series `json:"items"`
	Total        int64    `json:"total"`
	Page         int      `json:"page"`
	PageSize     int      `json:"pageSize"`
	SourceName   string   `json:"sourceName"`
	SourceURL    string   `json:"sourceUrl"`
	CacheEnabled bool     `json:"cacheEnabled"`
}
type Detail struct {
	Series
	Chapters   []Chapter `json:"chapters"`
	SourceName string    `json:"sourceName"`
}

type AdminSeries struct {
	Series
	ImportError string `json:"importError"`
}

// Submission is an editorial decision, never an arbitrary URL submitted by a reader.
type Submission struct {
	URL              string `json:"sourceUrl"`
	Enabled          bool   `json:"enabled"`
	JapaneseVerified bool   `json:"japaneseVerified"`
	NonAdultReviewed bool   `json:"nonAdultReviewed"`
}
