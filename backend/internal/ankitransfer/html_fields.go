package ankitransfer

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type cardMediaRef struct {
	kind  string
	value string
}

var soundReference = regexp.MustCompile(`(?i)\[sound:([^\]]+)\]`)
var clozeReference = regexp.MustCompile(`\{\{c[0-9]+::`)

func hasClozeSyntax(value string) bool { return clozeReference.MatchString(value) }

func sanitizeAnkiField(raw string, bundled map[string]Media) (string, []cardMediaRef, error) {
	refs := []cardMediaRef{}
	soundIndexes := soundReference.FindAllStringSubmatchIndex(raw, 21)
	if len(soundIndexes) > 20 {
		return "", nil, fmt.Errorf("field contains too many audio references")
	}
	var noSound strings.Builder
	last := 0
	for _, match := range soundIndexes {
		noSound.WriteString(raw[last:match[0]])
		name := raw[match[2]:match[3]]
		ref, err := normalizeMediaReference("audio", name, bundled)
		if err != nil {
			return "", nil, err
		}
		refs = append(refs, ref)
		last = match[1]
	}
	noSound.WriteString(raw[last:])

	var out strings.Builder
	tokenizer := html.NewTokenizer(strings.NewReader(noSound.String()))
	tokenizer.SetMaxBuf(MaxExpandedBytes)
	skipTag := ""
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			if tokenizer.Err() == nil || tokenizer.Err().Error() == "EOF" {
				break
			}
			return "", nil, fmt.Errorf("invalid HTML field: %w", tokenizer.Err())
		}
		token := tokenizer.Token()
		name := strings.ToLower(token.Data)
		if skipTag != "" {
			if tokenType == html.EndTagToken && name == skipTag {
				skipTag = ""
			}
			continue
		}
		switch tokenType {
		case html.StartTagToken, html.SelfClosingTagToken:
			if name == "script" || name == "style" || name == "iframe" || name == "object" {
				skipTag = name
				continue
			}
			if isBlockTag(name) {
				writeBreak(&out)
			}
			if name == "img" {
				src, alt := tokenAttr(token, "src"), tokenAttr(token, "alt")
				if src != "" {
					ref, err := normalizeMediaReference("image", src, bundled)
					if err != nil {
						return "", nil, err
					}
					refs = append(refs, ref)
				}
				if alt != "" {
					out.WriteString(alt)
				}
			}
			if name == "audio" || name == "source" {
				src := tokenAttr(token, "src")
				if src != "" {
					ref, err := normalizeMediaReference("audio", src, bundled)
					if err != nil {
						return "", nil, err
					}
					refs = append(refs, ref)
				}
			}
			if name == "br" || name == "hr" {
				writeBreak(&out)
			}
			if tokenType == html.SelfClosingTagToken && isBlockTag(name) {
				writeBreak(&out)
			}
		case html.EndTagToken:
			if isBlockTag(name) {
				writeBreak(&out)
			}
		case html.TextToken:
			out.WriteString(token.Data)
		case html.CommentToken:
			if strings.TrimSpace(token.Data) == answerSeparatorMarker {
				out.WriteString("\x00")
			}
		}
	}
	return cleanPlainText(out.String()), refs, nil
}

func normalizeMediaReference(kind, raw string, bundled map[string]Media) (cardMediaRef, error) {
	value := strings.TrimSpace(raw)
	parsed, err := url.Parse(value)
	if err != nil {
		return cardMediaRef{}, fmt.Errorf("invalid %s reference", kind)
	}
	if strings.EqualFold(parsed.Scheme, "http") || strings.EqualFold(parsed.Scheme, "https") {
		if parsed.Host == "" {
			return cardMediaRef{}, fmt.Errorf("invalid external %s reference", kind)
		}
		return cardMediaRef{kind: kind, value: parsed.String()}, nil
	}
	if parsed.Scheme != "" || parsed.Host != "" || strings.HasPrefix(value, "//") {
		return cardMediaRef{}, fmt.Errorf("unsafe %s reference", kind)
	}
	path, err := url.PathUnescape(parsed.Path)
	if err != nil {
		return cardMediaRef{}, fmt.Errorf("invalid escaped %s reference", kind)
	}
	if strings.ContainsAny(path, `/\\`) {
		return cardMediaRef{}, fmt.Errorf("unsafe local %s reference", kind)
	}
	if err := validateMediaName(path); err != nil {
		return cardMediaRef{}, err
	}
	if media, ok := bundled[strings.ToLower(path)]; ok {
		path = media.Name
	}
	if bundled != nil {
		if _, ok := bundled[strings.ToLower(path)]; !ok {
			return cardMediaRef{}, fmt.Errorf("APKG is missing referenced media %q", path)
		}
	}
	return cardMediaRef{kind: kind, value: "media:" + path}, nil
}

func tokenAttr(token html.Token, name string) string {
	for _, attr := range token.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val
		}
	}
	return ""
}

func assignCardMedia(card *Card, refs []cardMediaRef, cardIndex int) error {
	seen := map[string]map[string]bool{"image": {}, "audio": {}}
	values := map[string]string{}
	for _, ref := range refs {
		if ref.kind != "image" && ref.kind != "audio" {
			continue
		}
		if seen[ref.kind][ref.value] {
			continue
		}
		seen[ref.kind][ref.value] = true
		if values[ref.kind] != "" && values[ref.kind] != ref.value {
			return fmt.Errorf("Anki card %d has multiple %s references; the transfer format supports one per card", cardIndex, ref.kind)
		}
		values[ref.kind] = ref.value
	}
	card.ImageURL = values["image"]
	card.Audio = values["audio"]
	return nil
}

func splitEncodedAnswers(value string) []string {
	parts := strings.Split(value, "\x00")
	if len(parts) > 1 {
		for i := range parts {
			parts[i] = strings.Trim(parts[i], "\n")
		}
	}
	if len(parts) == 1 {
		return []string{value}
	}
	return parts
}

func cleanPlainText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	for strings.Contains(value, "\n\n\n") {
		value = strings.ReplaceAll(value, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(value)
}

func writeBreak(out *strings.Builder) {
	if out.Len() == 0 {
		return
	}
	value := out.String()
	if !strings.HasSuffix(value, "\n") {
		out.WriteByte('\n')
	}
}

func isBlockTag(name string) bool {
	switch name {
	case "address", "article", "aside", "blockquote", "div", "dl", "dt", "dd", "fieldset", "figcaption", "figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6", "header", "li", "main", "nav", "ol", "p", "pre", "section", "table", "tbody", "td", "tfoot", "th", "thead", "tr", "ul":
		return true
	default:
		return false
	}
}
