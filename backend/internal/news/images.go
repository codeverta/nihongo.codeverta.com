package news

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"nihongo-master-backend/internal/imageopt"

	_ "golang.org/x/image/webp"
	"golang.org/x/net/html"
)

type imageCandidate struct {
	URL, Credit, License, LicenseURL, SourceURL string
	CommonsFiles                                []string
}

// Only explicit commercial reuse licenses on the individual image caption
// qualify. A publisher's article license does not license its photographs.
var imageCCPath = regexp.MustCompile(`^/licenses/(by|by-sa)/(1\.0|2\.0|2\.5|3\.0|4\.0)/?$`)

func imageLicense(value string) string {
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.Port() != "" {
		return ""
	}
	if u.Hostname() != "creativecommons.org" && u.Hostname() != "www.creativecommons.org" {
		return ""
	}
	if m := imageCCPath.FindStringSubmatch(u.Path); m != nil {
		return "CC " + strings.ToUpper(m[1]) + " " + m[2]
	}
	if u.Path == "/publicdomain/zero/1.0/" {
		return "CC0 1.0"
	}
	if u.Path == "/publicdomain/mark/1.0/" {
		return "Public domain"
	}
	return ""
}

func allowedImageURL(value string) bool {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.Fragment != "" {
		return false
	}
	return (u.Host == "globalvoices.org" || u.Host == "jp.globalvoices.org") && strings.HasPrefix(u.Path, "/wp-content/uploads/") ||
		u.Host == "www.city.osaka.lg.jp" && strings.Contains(u.Path, "/cmsfiles/contents/")
}

func licensedImages(doc *html.Node) []imageCandidate {
	result := []imageCandidate{}
	walk(doc, func(n *html.Node) {
		if !hasClass(n, "wp-caption") && n.Data != "figure" {
			return
		}
		var caption *html.Node
		walk(n, func(c *html.Node) {
			if hasClass(c, "wp-caption-text") || c.Data == "figcaption" {
				caption = c
			}
		})
		if caption == nil {
			return
		}
		candidate := imageCandidate{Credit: plain(caption)}
		ineligible := false
		walk(caption, func(c *html.Node) {
			if c.Data != "a" {
				return
			}
			href := attr(c, "href")
			if u, err := url.Parse(href); err == nil {
				if u.Hostname() == "commons.wikimedia.org" && strings.HasPrefix(u.Path, "/wiki/File:") {
					candidate.CommonsFiles = append(candidate.CommonsFiles, strings.TrimPrefix(u.Path, "/wiki/"))
				}
				if strings.Contains(u.Hostname(), "creativecommons.org") && strings.HasPrefix(u.Path, "/licenses/") && imageLicense(href) == "" {
					ineligible = true
				}
			}
			if license := imageLicense(href); license != "" {
				candidate.License, candidate.LicenseURL = license, strings.Replace(href, "http://", "https://", 1)
			} else if u, err := url.Parse(href); err == nil && u.Scheme == "https" && u.User == nil {
				if candidate.SourceURL == "" {
					candidate.SourceURL = href
				}
			}
		})
		if ineligible || candidate.License == "" || candidate.Credit == "" || len(candidate.Credit) > 1000 {
			return
		}
		walk(n, func(c *html.Node) {
			if c.Data == "img" && candidate.URL == "" && allowedImageURL(attr(c, "src")) {
				candidate.URL = attr(c, "src")
			}
		})
		if candidate.URL != "" {
			if candidate.SourceURL == "" {
				candidate.SourceURL = candidate.URL
			}
			result = append(result, candidate)
		}
	})
	return result
}

func (f *HTTPFetcher) importImages(ctx context.Context, articles []Article) {
	if f.imageDirectory == "" {
		return
	}
	for i := range articles {
		for _, candidate := range articles[i].imageCandidates {
			credit, err := f.imageAttribution(ctx, candidate)
			if err != nil {
				log.Printf("news image credit %s: %v", articles[i].URL, err)
				continue
			}
			path, err := f.downloadImage(ctx, candidate.URL)
			if err != nil {
				log.Printf("news image import %s: %v", articles[i].URL, err)
				continue
			}
			articles[i].ImageURL, articles[i].ImageCredit, articles[i].ImageLicense = path, credit, candidate.License
			articles[i].ImageSourceURL, articles[i].ImageLicenseURL = candidate.SourceURL, candidate.LicenseURL
			break
		}
	}
}

// Captions sometimes credit only Commons. Include each file's actual creator,
// including both creators when the article's image is a composite.
func (f *HTTPFetcher) imageAttribution(ctx context.Context, candidate imageCandidate) (string, error) {
	credit := candidate.Credit
	for _, title := range candidate.CommonsFiles {
		query := url.Values{"action": {"query"}, "format": {"json"}, "prop": {"imageinfo"}, "iiprop": {"extmetadata"}, "titles": {title}}
		data, _, _, err := f.get(ctx, "https://commons.wikimedia.org/w/api.php?"+query.Encode(), Source{}, false)
		if err != nil {
			return "", err
		}
		var response struct {
			Query struct {
				Pages map[string]struct {
					ImageInfo []struct {
						ExtMetadata map[string]struct {
							Value any `json:"value"`
						} `json:"extmetadata"`
					} `json:"imageinfo"`
				} `json:"pages"`
			} `json:"query"`
		}
		if err := json.Unmarshal(data, &response); err != nil {
			return "", err
		}
		found := false
		for _, page := range response.Query.Pages {
			for _, info := range page.ImageInfo {
				metadata := info.ExtMetadata
				artistHTML, _ := metadata["Artist"].Value.(string)
				artistDoc, err := html.Parse(strings.NewReader(artistHTML))
				if err != nil {
					return "", err
				}
				artist := plain(artistDoc)
				licenseURL, _ := metadata["LicenseUrl"].Value.(string)
				license := imageLicense(licenseURL)
				if artist == "" || license == "" {
					continue
				}
				credit += " · " + artist + " (" + license + "; https://commons.wikimedia.org/wiki/" + url.PathEscape(title) + ")"
				found = true
			}
		}
		if !found {
			return "", fmt.Errorf("image creator or reusable Commons license unavailable")
		}
	}
	if len(credit) > 8000 {
		return "", fmt.Errorf("image credit exceeds limit")
	}
	return credit, nil
}

func (f *HTTPFetcher) downloadImage(ctx context.Context, remote string) (string, error) {
	if !allowedImageURL(remote) {
		return "", ErrInvalid
	}
	data, _, _, err := f.get(ctx, remote, Source{}, false)
	if err != nil {
		return "", err
	}
	if len(data) > 5<<20 {
		return "", fmt.Errorf("image exceeds 5 MiB")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	if (format != "jpeg" && format != "png" && format != "webp") || config.Width < 1 || config.Height < 1 || config.Width > 4096 || config.Height > 4096 {
		return "", fmt.Errorf("unsupported image dimensions or format")
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	encoded, err := imageopt.JPEG(decoded)
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("%x.jpg", sha256.Sum256(encoded))
	if err := os.MkdirAll(f.imageDirectory, 0755); err != nil {
		return "", err
	}
	target := filepath.Join(f.imageDirectory, name)
	if _, err := os.Stat(target); err == nil {
		return "/news-images/" + name, nil
	}
	temp, err := os.CreateTemp(f.imageDirectory, ".news-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(temp.Name())
	_, err = temp.Write(encoded)
	if err == nil {
		err = temp.Sync()
	}
	closeErr := temp.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(temp.Name(), target)
	}
	if err != nil {
		if _, statErr := os.Stat(target); statErr != nil {
			return "", err
		}
	}
	return "/news-images/" + name, nil
}
