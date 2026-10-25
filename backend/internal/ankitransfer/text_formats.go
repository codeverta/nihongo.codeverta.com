package ankitransfer

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func parseJSON(raw []byte) (Document, error) {
	if !utf8.Valid(raw) {
		return Document{}, fmt.Errorf("JSON deck is not valid UTF-8")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var doc Document
	if err := dec.Decode(&doc); err != nil {
		return Document{}, fmt.Errorf("invalid deck JSON: %w", err)
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return Document{}, fmt.Errorf("invalid deck JSON: trailing value")
		}
		return Document{}, fmt.Errorf("invalid deck JSON: %w", err)
	}
	if doc.Version != DocumentVersion {
		return Document{}, fmt.Errorf("unsupported deck JSON version %d", doc.Version)
	}
	return doc, nil
}

type csvMetadata struct {
	Version     int    `json:"version"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	SourceURL   string `json:"sourceUrl,omitempty"`
	SourceID    string `json:"sourceId,omitempty"`
	Level       string `json:"level,omitempty"`
}

var csvHeaders = []string{"front", "answers", "meaning", "imageUrl", "audio"}

func exportJSON(doc Document) ([]byte, error) {
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode deck JSON: %w", err)
	}
	return append(out, '\n'), nil
}

func exportCSV(doc Document) ([]byte, error) {
	var buf bytes.Buffer
	meta, _ := json.Marshal(csvMetadata{Version: doc.Version, Title: doc.Title, Description: doc.Description, SourceURL: doc.SourceURL, SourceID: doc.SourceID, Level: doc.Level})
	buf.WriteString("#nihongo-master:" + base64.StdEncoding.EncodeToString(meta) + "\n")
	w := csv.NewWriter(&buf)
	w.UseCRLF = false
	if err := w.Write(csvHeaders); err != nil {
		return nil, err
	}
	for _, card := range doc.Cards {
		answers, marshalErr := json.Marshal(card.Answers)
		if marshalErr != nil {
			return nil, marshalErr
		}
		image, audio := card.ImageURL, card.Audio
		if strings.HasPrefix(image, "media:") {
			image = ""
		}
		if strings.HasPrefix(audio, "media:") {
			audio = ""
		}
		row := []string{card.Front, string(answers), card.Meaning, image, audio}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func parseCSV(filename string, raw []byte) (Document, error) {
	if !utf8.Valid(raw) {
		return Document{}, fmt.Errorf("CSV deck is not valid UTF-8")
	}
	var meta *csvMetadata
	raw = bytes.TrimPrefix(raw, []byte{0xef, 0xbb, 0xbf})
	if bytes.HasPrefix(raw, []byte("#nihongo-master:")) {
		line, rest, ok := bytes.Cut(raw, []byte("\n"))
		if !ok {
			return Document{}, fmt.Errorf("CSV metadata missing body")
		}
		encoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(line[len("#nihongo-master:"):])))
		if err != nil {
			return Document{}, fmt.Errorf("invalid CSV metadata")
		}
		meta = &csvMetadata{}
		if json.Unmarshal(encoded, meta) != nil || meta.Version != DocumentVersion {
			return Document{}, fmt.Errorf("invalid CSV metadata version")
		}
		raw = rest
	}
	r := csv.NewReader(bytes.NewReader(raw))
	r.FieldsPerRecord = -1
	r.ReuseRecord = false
	first, err := r.Read()
	if err == io.EOF {
		return Document{}, fmt.Errorf("CSV deck is empty")
	}
	if err != nil {
		return Document{}, fmt.Errorf("invalid CSV deck: %w", err)
	}
	doc := Document{Version: DocumentVersion, Title: titleFromFilename(filename), Cards: []Card{}, Media: []Media{}}
	if meta != nil {
		doc.Title, doc.Description, doc.SourceURL, doc.SourceID, doc.Level = meta.Title, meta.Description, meta.SourceURL, meta.SourceID, meta.Level
	}
	header := make(map[string]int, len(first))
	for i, value := range first {
		header[strings.ToLower(strings.TrimSpace(value))] = i
	}
	if _, ok := header["front"]; !ok {
		return Document{}, fmt.Errorf("CSV header must contain front")
	}
	_, back := header["back"]
	_, answers := header["answers"]
	if !back && !answers {
		return Document{}, fmt.Errorf("CSV header must contain back or answers")
	}
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Document{}, fmt.Errorf("invalid CSV: %w", err)
		}
		if len(row) != len(first) {
			return Document{}, fmt.Errorf("CSV row has wrong number of columns")
		}
		if len(doc.Cards) >= MaxCards {
			return Document{}, fmt.Errorf("CSV exceeds card limit")
		}
		if err := appendCSVCard(&doc, row, header); err != nil {
			return Document{}, err
		}
	}
	return doc, nil
}

func appendCSVCard(doc *Document, row []string, header map[string]int) error {
	get := func(name string) string {
		if index, ok := header[name]; ok && index < len(row) {
			return row[index]
		}
		return ""
	}
	card := Card{Front: get("front"), Meaning: get("meaning"), ImageURL: get("imageurl"), Audio: get("audio")}
	answersValue := get("answers")
	if answersValue != "" {
		if err := json.Unmarshal([]byte(answersValue), &card.Answers); err != nil || len(card.Answers) == 0 {
			if strings.HasPrefix(strings.TrimSpace(answersValue), "[") {
				return fmt.Errorf("invalid answers JSON array")
			}
			// Accept a conventional single Back/answers text column as well as the
			// versioned format's JSON array.
			if _, hasBack := header["back"]; hasBack {
				card.Answers = []string{get("back")}
			} else {
				card.Answers = []string{answersValue}
			}
		}
	} else if back := get("back"); back != "" {
		card.Answers = []string{back}
	}
	doc.Cards = append(doc.Cards, card)
	if len(doc.Cards) > MaxCards {
		return fmt.Errorf("deck exceeds %d cards", MaxCards)
	}
	return nil
}

func titleFromFilename(filename string) string {
	name := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	if name == "." || name == "" {
		return "Imported deck"
	}
	return name
}
