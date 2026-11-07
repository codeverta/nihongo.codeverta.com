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
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"nihongo-master-backend/internal/imageopt"

	_ "golang.org/x/image/webp"
	"golang.org/x/net/html"
)

type imageCandidate struct {
	URL, Credit, License, LicenseURL, SourceURL string
	CommonsFiles                                []string
	// Unlicensed marks a plain article image (og:image, feed media or the first
	// body image). It is credited to the publisher instead of a license.
	Unlicensed bool
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

// safeImageURL accepts any public HTTPS image URL (publisher CDNs included)
// while refusing credentials, ports, IP literals and local hostnames.
func safeImageURL(value string) bool {
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || len(value) > 1000 {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" || host == "localhost" || !strings.Contains(host, ".") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return false
	}
	return net.ParseIP(host) == nil
}

var skippedImageHints = []string{"gravatar", "emoji", "/smilies/", "logo", "icon", "sprite", "avatar", "spacer", "pixel", ".svg", ".gif"}

func resolveImageURL(base, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.HasPrefix(ref, "data:") {
		return ""
	}
	b, err := url.Parse(base)
	if err != nil {
		return ""
	}
	r, err := b.Parse(ref)
	if err != nil {
		return ""
	}
	r.Fragment = ""
	value := r.String()
	if !safeImageURL(value) {
		return ""
	}
	lower := strings.ToLower(value)
	for _, hint := range skippedImageHints {
		if strings.Contains(lower, hint) {
			return ""
		}
	}
	return value
}

// metaImage returns og:image / twitter:image from a full HTML document.
func metaImage(doc *html.Node, base string) string {
	found := ""
	walk(doc, func(n *html.Node) {
		if n.Data != "meta" || found != "" {
			return
		}
		key := strings.ToLower(attr(n, "property") + attr(n, "name"))
		if key == "og:image" || key == "og:image:secure_url" || key == "twitter:image" {
			found = resolveImageURL(base, attr(n, "content"))
		}
	})
	return found
}

// bodyImages lists article images in document order, skipping tiny images.
func bodyImages(root *html.Node, base string) []string {
	result := []string{}
	seen := map[string]bool{}
	walk(root, func(n *html.Node) {
		if n.Data != "img" {
			return
		}
		for _, dim := range []string{"width", "height"} {
			if v, err := strconv.Atoi(strings.TrimSuffix(attr(n, dim), "px")); err == nil && v < 120 {
				return
			}
		}
		src := attr(n, "src")
		for _, lazy := range []string{"data-src", "data-lazy-src", "data-orig-file"} {
			if v := attr(n, lazy); v != "" {
				src = v
				break
			}
		}
		if value := resolveImageURL(base, src); value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	})
	return result
}

// plainImageCandidates gathers unlicensed fallbacks: feed media first, then the
// page's og:image, then body images.
func plainImageCandidates(base string, feedURLs []string, doc, body *html.Node) []imageCandidate {
	urls := []string{}
	for _, v := range feedURLs {
		if value := resolveImageURL(base, v); value != "" {
			urls = append(urls, value)
		}
	}
	if doc != nil {
		if v := metaImage(doc, base); v != "" {
			urls = append(urls, v)
		}
	}
	if body != nil {
		urls = append(urls, bodyImages(body, base)...)
	}
	result := []imageCandidate{}
	seen := map[string]bool{}
	for _, v := range urls {
		if !seen[v] {
			seen[v] = true
			result = append(result, imageCandidate{URL: v, SourceURL: base, Unlicensed: true})
		}
	}
	if len(result) > 4 {
		result = result[:4]
	}
	return result
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
	if f.imageDirectory == "" && f.store == nil {
		return
	}
	for i := range articles {
		// Articles that already hold a stored image are never re-downloaded
		// from the publisher.
		if f.hasImage != nil && f.hasImage(ctx, articles[i].URL) {
			continue
		}
		for _, candidate := range articles[i].imageCandidates {
			credit, license := candidate.Credit, candidate.License
			if candidate.Unlicensed {
				host := "sumber berita"
				if u, err := url.Parse(articles[i].URL); err == nil && u.Hostname() != "" {
					host = u.Hostname()
				}
				credit, license = "Foto dari "+host, "Hak cipta pemilik sumber"
			} else {
				var err error
				if credit, err = f.imageAttribution(ctx, candidate); err != nil {
					log.Printf("news image credit %s: %v", articles[i].URL, err)
					continue
				}
			}
			path, err := f.downloadImage(ctx, candidate.URL)
			if err != nil {
				log.Printf("news image import %s: %v", articles[i].URL, err)
				continue
			}
			articles[i].ImageURL, articles[i].ImageCredit, articles[i].ImageLicense = path, credit, license
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
	if !safeImageURL(remote) {
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
	if (format != "jpeg" && format != "png" && format != "webp") || config.Width < 100 || config.Height < 100 || config.Width > 8000 || config.Height > 8000 || config.Width*config.Height > 30_000_000 {
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
	if f.store != nil {
		return f.store.Put(ctx, encoded)
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
