package places

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	_ "golang.org/x/image/webp"

	"nihongo-master-backend/internal/imageopt"
)

// DefaultUserAgent follows the Wikimedia User-Agent policy: an honest client
// name and a contact page. Override with MAP_SOURCE_USER_AGENT.
const DefaultUserAgent = "NihongoMaster/1.0 (+https://nihongo.codeverta.com/support)"

const (
	maxJSONBytes  = 4 << 20
	maxImageBytes = 15 << 20
	requestGap    = 300 * time.Millisecond
)

// Fetcher talks to the public Wikimedia APIs without keys. Requests are
// sequential and spaced out; it is only used by admin imports and refreshes.
type Fetcher struct {
	HTTP      *http.Client
	UserAgent string
	// Base URLs are overridable for tests.
	WikidataAPI string
	CommonsAPI  string
	RestAPI     func(project, lang string) string
	// Gap is the minimum delay between requests (polite crawling).
	Gap time.Duration

	mu   sync.Mutex
	last time.Time
}

func NewFetcher(userAgent string) *Fetcher {
	if strings.TrimSpace(userAgent) == "" {
		userAgent = DefaultUserAgent
	}
	return &Fetcher{
		HTTP:        &http.Client{Timeout: 20 * time.Second},
		UserAgent:   userAgent,
		Gap:         requestGap,
		WikidataAPI: "https://www.wikidata.org/w/api.php",
		CommonsAPI:  "https://commons.wikimedia.org/w/api.php",
		RestAPI: func(project, lang string) string {
			return "https://" + lang + "." + project + ".org/api/rest_v1"
		},
	}
}

type Page struct{ Text, URL string }

type FetchedImage struct {
	JPEG                                   []byte
	Author, License, LicenseURL, SourceURL string
}

// Fetched is everything one import or refresh collects.
type Fetched struct {
	WikidataID string
	Labels     map[string]string // en, id, ja
	Lat, Lng   *float64
	Wikipedia  map[string]Page // en, id, ja (ja: URL only)
	Wikivoyage map[string]Page // en, id
	Image      *FetchedImage
}

var (
	qidPattern       = regexp.MustCompile(`^Q[1-9][0-9]{0,11}$`)
	wikidataURL      = regexp.MustCompile(`^https?://(?:www\.|m\.)?wikidata\.org/wiki/(Q[1-9][0-9]{0,11})$`)
	enWikipediaURL   = regexp.MustCompile(`^https?://en\.(?:m\.)?wikipedia\.org/wiki/(.+)$`)
	htmlTag          = regexp.MustCompile(`<[^>]*>`)
	whitespaceRun    = regexp.MustCompile(`\s+`)
	allowedFileTypes = regexp.MustCompile(`(?i)\.(jpe?g|png|gif|webp|tiff?|svg)$`)
)

// ParseInput accepts a Wikidata ID ("Q123"), a Wikidata URL, an en.wikipedia
// URL, or an en.wikipedia title. It returns either the ID or the title.
func ParseInput(input string) (qid, title string, err error) {
	input = strings.TrimSpace(input)
	if input == "" || !utf8.ValidString(input) || utf8.RuneCountInString(input) > 300 {
		return "", "", ErrInvalid
	}
	if qidPattern.MatchString(strings.ToUpper(input)) {
		return strings.ToUpper(input), "", nil
	}
	if m := wikidataURL.FindStringSubmatch(input); m != nil {
		return m[1], "", nil
	}
	if m := enWikipediaURL.FindStringSubmatch(input); m != nil {
		decoded, err := url.PathUnescape(strings.SplitN(m[1], "#", 2)[0])
		if err != nil {
			return "", "", ErrInvalid
		}
		input = decoded
	} else if strings.Contains(input, "://") {
		return "", "", ErrInvalid
	}
	title = strings.TrimSpace(strings.ReplaceAll(input, "_", " "))
	if title == "" || strings.ContainsAny(title, "#<>[]{}|") {
		return "", "", ErrInvalid
	}
	return "", title, nil
}

func (f *Fetcher) Fetch(ctx context.Context, qid, title string) (Fetched, error) {
	out := Fetched{Labels: map[string]string{}, Wikipedia: map[string]Page{}, Wikivoyage: map[string]Page{}}
	query := url.Values{
		"action": {"wbgetentities"}, "format": {"json"}, "formatversion": {"2"},
		"props":      {"labels|sitelinks|claims"},
		"languages":  {"en|id|ja"},
		"sitefilter": {"enwiki|idwiki|jawiki|enwikivoyage|idwikivoyage"},
	}
	if qid != "" {
		query.Set("ids", qid)
	} else {
		query.Set("sites", "enwiki")
		query.Set("titles", title)
		query.Set("normalize", "1")
	}
	var entities struct {
		Entities map[string]struct {
			ID      string  `json:"id"`
			Missing *string `json:"missing"`
			Labels  map[string]struct {
				Value string `json:"value"`
			} `json:"labels"`
			Sitelinks map[string]struct {
				Title string `json:"title"`
			} `json:"sitelinks"`
			Claims map[string][]wikidataClaim `json:"claims"`
		} `json:"entities"`
	}
	if err := f.getJSON(ctx, f.WikidataAPI+"?"+query.Encode(), &entities); err != nil {
		return out, err
	}
	found := false
	for key, entity := range entities.Entities {
		if entity.Missing != nil || !qidPattern.MatchString(entity.ID) || strings.HasPrefix(key, "-") {
			continue
		}
		found = true
		out.WikidataID = entity.ID
		for _, lang := range []string{"en", "id", "ja"} {
			if label := strings.TrimSpace(entity.Labels[lang].Value); label != "" {
				out.Labels[lang] = label
			}
		}
		for _, claim := range preferred(entity.Claims["P625"]) {
			var coordinate struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			}
			if json.Unmarshal(claim, &coordinate) == nil && coordinate.Latitude >= 20 && coordinate.Latitude <= 46 && coordinate.Longitude >= 122 && coordinate.Longitude <= 154 {
				lat, lng := coordinate.Latitude, coordinate.Longitude
				out.Lat, out.Lng = &lat, &lng
				break
			}
		}
		var imageFile string
		for _, claim := range preferred(entity.Claims["P18"]) {
			if json.Unmarshal(claim, &imageFile) == nil && allowedFileTypes.MatchString(imageFile) {
				break
			}
			imageFile = ""
		}
		for site, project := range map[string][2]string{
			"enwiki": {"wikipedia", "en"}, "idwiki": {"wikipedia", "id"}, "jawiki": {"wikipedia", "ja"},
			"enwikivoyage": {"wikivoyage", "en"}, "idwikivoyage": {"wikivoyage", "id"},
		} {
			pageTitle := entity.Sitelinks[site].Title
			if pageTitle == "" {
				continue
			}
			if project[1] == "ja" {
				out.Wikipedia["ja"] = Page{URL: articleURL(project[0], "ja", pageTitle)}
				continue
			}
			page, err := f.summary(ctx, project[0], project[1], pageTitle)
			if err != nil {
				return out, err
			}
			if page.Text == "" {
				continue
			}
			if project[0] == "wikipedia" {
				out.Wikipedia[project[1]] = page
			} else {
				out.Wikivoyage[project[1]] = page
			}
		}
		if imageFile != "" {
			img, err := f.commonsImage(ctx, imageFile)
			if err != nil {
				return out, err
			}
			out.Image = img
		}
		break
	}
	if !found {
		return out, ErrNotFound
	}
	return out, nil
}

// preferred returns claim values, preferred-rank first, deprecated excluded.
type wikidataClaim struct {
	Mainsnak struct {
		Datavalue struct {
			Value json.RawMessage `json:"value"`
		} `json:"datavalue"`
	} `json:"mainsnak"`
	Rank string `json:"rank"`
}

func preferred(claims []wikidataClaim) []json.RawMessage {
	var first, rest []json.RawMessage
	for _, claim := range claims {
		value := claim.Mainsnak.Datavalue.Value
		switch {
		case len(value) == 0 || claim.Rank == "deprecated":
		case claim.Rank == "preferred":
			first = append(first, value)
		default:
			rest = append(rest, value)
		}
	}
	return append(first, rest...)
}

func articleURL(project, lang, title string) string {
	return "https://" + lang + "." + project + ".org/wiki/" + url.PathEscape(strings.ReplaceAll(title, " ", "_"))
}

func (f *Fetcher) summary(ctx context.Context, project, lang, title string) (Page, error) {
	endpoint := f.RestAPI(project, lang) + "/page/summary/" + url.PathEscape(strings.ReplaceAll(title, " ", "_"))
	var data struct {
		Type        string `json:"type"`
		Extract     string `json:"extract"`
		ContentURLs struct {
			Desktop struct {
				Page string `json:"page"`
			} `json:"desktop"`
		} `json:"content_urls"`
	}
	if err := f.getJSON(ctx, endpoint, &data); err != nil {
		if err == ErrNotFound {
			return Page{}, nil
		}
		return Page{}, err
	}
	if data.Type == "disambiguation" {
		return Page{}, nil
	}
	page := Page{Text: strings.TrimSpace(data.Extract), URL: data.ContentURLs.Desktop.Page}
	if !strings.HasPrefix(page.URL, "https://") {
		page.URL = articleURL(project, lang, title)
	}
	return page, nil
}

func (f *Fetcher) commonsImage(ctx context.Context, file string) (*FetchedImage, error) {
	query := url.Values{
		"action": {"query"}, "format": {"json"}, "formatversion": {"2"},
		"prop": {"imageinfo"}, "iiprop": {"url|extmetadata"}, "iiurlwidth": {"1200"},
		"titles": {"File:" + file},
	}
	var data struct {
		Query struct {
			Pages []struct {
				Missing   bool `json:"missing"`
				ImageInfo []struct {
					URL            string `json:"url"`
					ThumbURL       string `json:"thumburl"`
					DescriptionURL string `json:"descriptionurl"`
					Metadata       map[string]struct {
						Value string `json:"value"`
					} `json:"extmetadata"`
				} `json:"imageinfo"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := f.getJSON(ctx, f.CommonsAPI+"?"+query.Encode(), &data); err != nil {
		if err == ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	if len(data.Query.Pages) == 0 || data.Query.Pages[0].Missing || len(data.Query.Pages[0].ImageInfo) == 0 {
		return nil, nil
	}
	info := data.Query.Pages[0].ImageInfo[0]
	license := plainText(info.Metadata["LicenseShortName"].Value)
	// Commons only hosts freely licensed files; still require a stated license
	// so attribution can always be shown.
	if license == "" {
		return nil, nil
	}
	author := plainText(info.Metadata["Artist"].Value)
	if author == "" {
		author = plainText(info.Metadata["Credit"].Value)
	}
	source := info.ThumbURL
	if source == "" {
		source = info.URL
	}
	if !strings.HasPrefix(source, "https://upload.wikimedia.org/") {
		return nil, nil
	}
	raw, err := f.get(ctx, source, maxImageBytes)
	if err != nil {
		return nil, err
	}
	decoded, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, nil
	}
	jpeg, err := imageopt.JPEG(decoded)
	if err != nil {
		return nil, nil
	}
	return &FetchedImage{
		JPEG:       jpeg,
		Author:     clip(author, 500),
		License:    clip(license, 120),
		LicenseURL: httpsOnly(info.Metadata["LicenseUrl"].Value),
		SourceURL:  httpsOnly(info.DescriptionURL),
	}, nil
}

func plainText(value string) string {
	value = html.UnescapeString(htmlTag.ReplaceAllString(value, " "))
	return strings.TrimSpace(whitespaceRun.ReplaceAllString(value, " "))
}

func httpsOnly(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "//") {
		value = "https:" + value
	}
	if !strings.HasPrefix(value, "https://") || len(value) > 1000 {
		return ""
	}
	return value
}

func clip(text string, limit int) string {
	if utf8.RuneCountInString(text) <= limit {
		return text
	}
	return string([]rune(text)[:limit-1]) + "…"
}

func (f *Fetcher) getJSON(ctx context.Context, endpoint string, target any) error {
	body, err := f.get(ctx, endpoint, maxJSONBytes)
	if err != nil {
		return err
	}
	if json.Unmarshal(body, target) != nil {
		return fmt.Errorf("%w: invalid JSON", ErrSource)
	}
	return nil
}

func (f *Fetcher) get(ctx context.Context, endpoint string, limit int64) ([]byte, error) {
	f.mu.Lock()
	if wait := f.Gap - time.Since(f.last); wait > 0 {
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			f.mu.Unlock()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	f.last = time.Now()
	f.mu.Unlock()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", f.UserAgent)
	req.Header.Set("Api-User-Agent", f.UserAgent)
	resp, err := f.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSource, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrSource, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSource, err)
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("%w: response too large", ErrSource)
	}
	return body, nil
}
