// Package ankitransfer reads and writes self-contained Anki deck packages and
// the app's portable JSON/CSV deck formats.
package ankitransfer

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	DocumentVersion  = 1
	MaxUploadBytes   = 20 << 20
	MaxExpandedBytes = 60 << 20
	MaxCards         = 10_000
	MaxMediaBytes    = 50 << 20
	maxMediaCount    = 10_000
)

// Document is the versioned transfer representation shared by JSON and CSV.
type Document struct {
	Version     int     `json:"version"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	SourceURL   string  `json:"sourceUrl,omitempty"`
	SourceID    string  `json:"sourceId,omitempty"`
	Level       string  `json:"level,omitempty"`
	Cards       []Card  `json:"cards"`
	Media       []Media `json:"media,omitempty"`
}

// Card contains one prompt and its accepted answers. ImageURL and Audio may
// be remote http(s) URLs or a local reference in the form media:<safe-name>.
type Card struct {
	Front    string   `json:"front"`
	Answers  []string `json:"answers"`
	Meaning  string   `json:"meaning,omitempty"`
	ImageURL string   `json:"imageUrl,omitempty"`
	Audio    string   `json:"audio,omitempty"`
}

// Media is an embedded image/audio file. Data is base64-encoded by JSON.
type Media struct {
	Name string `json:"name"`
	MIME string `json:"mime,omitempty"`
	Data []byte `json:"data"`
}

var (
	ErrUnsupportedFormat = errors.New("unsupported deck format")
	ErrUnsupportedNote   = errors.New("unsupported Anki note type")
)

func Parse(filename string, raw []byte) (Document, error) {
	if len(raw) > MaxUploadBytes {
		return Document{}, fmt.Errorf("upload exceeds %d MiB", MaxUploadBytes>>20)
	}
	format := formatFromFilename(filename)
	if format == "" {
		format = sniffFormat(raw)
	}
	var doc Document
	var err error
	switch format {
	case "apkg":
		doc, err = parseAPKG(filename, raw)
	case "json":
		doc, err = parseJSON(raw)
	case "csv":
		doc, err = parseCSV(filename, raw)
	default:
		return Document{}, ErrUnsupportedFormat
	}
	if err != nil {
		return Document{}, err
	}
	return normalizeDocument(doc, true)
}

func Export(doc Document, format string) ([]byte, error) {
	clean, err := normalizeDocument(doc, false)
	if err != nil {
		return nil, err
	}
	var raw []byte
	switch strings.TrimPrefix(strings.ToLower(strings.TrimSpace(format)), ".") {
	case "apkg", "application/vnd.anki":
		raw, err = exportAPKG(clean)
	case "csv", "text/csv":
		raw, err = exportCSV(clean)
	case "json", "application/json":
		raw, err = exportJSON(clean)
	default:
		return nil, ErrUnsupportedFormat
	}
	if err != nil {
		return nil, err
	}
	if len(raw) > MaxUploadBytes {
		return nil, fmt.Errorf("export exceeds 20 MiB; use a smaller deck or a more compact format")
	}
	return raw, nil
}

func formatFromFilename(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".apkg", ".colpkg":
		return "apkg"
	case ".json":
		return "json"
	case ".csv", ".tsv":
		return "csv"
	default:
		return ""
	}
}

func sniffFormat(raw []byte) string {
	trimmed := strings.TrimSpace(string(raw))
	if len(raw) >= 4 && string(raw[:4]) == "PK\x03\x04" {
		return "apkg"
	}
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return "json"
	}
	if len(raw) != 0 {
		return "csv"
	}
	return ""
}

func normalizeDocument(doc Document, importing bool) (Document, error) {
	if doc.Version == 0 && importing {
		doc.Version = DocumentVersion
	}
	if doc.Version != DocumentVersion {
		return Document{}, fmt.Errorf("unsupported deck JSON version %d", doc.Version)
	}
	if doc.SourceURL != "" && !isHTTPURL(doc.SourceURL) {
		return Document{}, fmt.Errorf("unsafe source URL")
	}
	if len(doc.Cards) > MaxCards {
		return Document{}, fmt.Errorf("deck exceeds %d cards", MaxCards)
	}
	if len(doc.Cards) == 0 {
		return Document{}, fmt.Errorf("deck contains no cards")
	}
	if len(doc.Media) > maxMediaCount {
		return Document{}, fmt.Errorf("deck exceeds %d media files", maxMediaCount)
	}
	if doc.Cards == nil {
		doc.Cards = []Card{}
	}
	if doc.Media == nil {
		doc.Media = []Media{}
	}
	mediaByName := make(map[string]struct{}, len(doc.Media))
	var mediaBytes int64
	for i := range doc.Media {
		media := &doc.Media[i]
		if err := validateMediaName(media.Name); err != nil {
			return Document{}, fmt.Errorf("media %d: %w", i+1, err)
		}
		key := strings.ToLower(media.Name)
		if _, ok := mediaByName[key]; ok {
			return Document{}, fmt.Errorf("duplicate media name %q", media.Name)
		}
		mediaByName[key] = struct{}{}
		detected, err := DetectMediaMIME(media.Data)
		if err != nil {
			return Document{}, fmt.Errorf("media %q: %w", media.Name, err)
		}
		media.MIME = detected
		mediaBytes += int64(len(media.Data))
		if mediaBytes > MaxMediaBytes {
			return Document{}, fmt.Errorf("media exceeds %d MiB", MaxMediaBytes>>20)
		}
	}
	for index, card := range doc.Cards {
		if len(card.Front) > 60000 || len(card.Meaning) > 60000 || len(card.Answers) > 100 || len(card.ImageURL) > 8192 || len(card.Audio) > 8192 {
			return Document{}, fmt.Errorf("card %d exceeds field limits", index+1)
		}
		if strings.TrimSpace(card.Front) == "" {
			return Document{}, fmt.Errorf("card %d has an empty front", index+1)
		}
		if len(card.Answers) == 0 {
			return Document{}, fmt.Errorf("card %d has no answer", index+1)
		}
		for answerIndex, answer := range card.Answers {
			if len(answer) > 60000 {
				return Document{}, fmt.Errorf("card %d answer exceeds field limits", index+1)
			}
			if strings.TrimSpace(answer) == "" {
				return Document{}, fmt.Errorf("card %d answer %d is empty", index+1, answerIndex+1)
			}
		}
		for _, ref := range []string{card.ImageURL, card.Audio} {
			if strings.HasPrefix(ref, "media:") {
				name := strings.TrimPrefix(ref, "media:")
				if err := validateMediaName(name); err != nil {
					return Document{}, fmt.Errorf("card %d: %w", index+1, err)
				}
				if _, ok := mediaByName[strings.ToLower(name)]; !ok {
					return Document{}, fmt.Errorf("card %d references missing media %q", index+1, name)
				}
			} else if ref != "" && !isHTTPURL(ref) {
				return Document{}, fmt.Errorf("card %d has an unsafe media URL", index+1)
			}
		}
	}
	return doc, nil
}

func validateMediaName(name string) error {
	if name == "" || len(name) > 255 || name == "." || name == ".." {
		return fmt.Errorf("unsafe media name %q", name)
	}
	if strings.ContainsAny(name, `/\\:<>"|?*[]`) || strings.TrimSpace(name) != name || strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return fmt.Errorf("unsafe media name %q", name)
	}
	for _, r := range name {
		if unicode.IsControl(r) || r == 0 || r == '\\' || r == '/' || r == ':' {
			return fmt.Errorf("unsafe media name %q", name)
		}
	}
	base := strings.TrimRight(strings.SplitN(name, ".", 2)[0], " .")
	reserved := map[string]bool{"CON": true, "PRN": true, "AUX": true, "NUL": true}
	for i := 1; i <= 9; i++ {
		reserved[fmt.Sprintf("COM%d", i)] = true
		reserved[fmt.Sprintf("LPT%d", i)] = true
	}
	if reserved[strings.ToUpper(base)] {
		return fmt.Errorf("unsafe media name %q", name)
	}
	return nil
}

func isHTTPURL(value string) bool {
	u, err := url.Parse(value)
	return err == nil && (strings.EqualFold(u.Scheme, "http") || strings.EqualFold(u.Scheme, "https")) && u.Host != ""
}
