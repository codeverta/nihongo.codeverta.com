package ankitransfer

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
	_ "modernc.org/sqlite"
)

const (
	legacy1               = 1
	legacy2               = 2
	latest                = 3
	answerSeparatorMarker = "anki-transfer-answer-separator:v1"
)

type archiveEntry struct {
	file *zip.File
	name string
}

type mediaManifestEntry struct {
	name       string
	size       uint64
	sha1       []byte
	legacyName *uint64
}

type ankiNotetype struct {
	fields    []string
	kind      int
	templates int
}

func parseAPKG(filename string, raw []byte) (Document, error) {
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return Document{}, fmt.Errorf("invalid APKG ZIP archive: %w", err)
	}
	entries, err := indexArchive(reader)
	if err != nil {
		return Document{}, err
	}
	version, err := packageVersion(entries)
	if err != nil {
		return Document{}, err
	}
	collectionName := map[int]string{legacy1: "collection.anki2", legacy2: "collection.anki21", latest: "collection.anki21b"}[version]
	collection, ok := entries[strings.ToLower(collectionName)]
	if !ok {
		return Document{}, fmt.Errorf("APKG is missing %s", collectionName)
	}
	collectionBytes, err := readZipEntry(collection.file, MaxExpandedBytes)
	if err != nil {
		return Document{}, fmt.Errorf("read %s: %w", collectionName, err)
	}
	if version == latest {
		collectionBytes, err = decodeZstdLimited(collectionBytes, MaxExpandedBytes)
		if err != nil {
			return Document{}, fmt.Errorf("decompress %s: %w", collectionName, err)
		}
	}
	if int64(len(collectionBytes)) > MaxExpandedBytes {
		return Document{}, fmt.Errorf("collection database exceeds %d MiB", MaxExpandedBytes>>20)
	}
	media, byName, err := readPackageMedia(entries, version)
	if err != nil {
		return Document{}, err
	}
	expanded := len(collectionBytes)
	for _, item := range media {
		expanded += len(item.Data)
	}
	if expanded > MaxExpandedBytes {
		return Document{}, fmt.Errorf("APKG expands beyond %d MiB", MaxExpandedBytes>>20)
	}
	doc, err := readAnkiCollection(filename, collectionBytes, byName)
	if err != nil {
		return Document{}, err
	}
	doc.Version = DocumentVersion
	doc.Media = media
	if entry, ok := entries["nihongo-master.json"]; ok {
		raw, err := readZipEntry(entry.file, 128<<10)
		if err != nil {
			return Document{}, err
		}
		var meta csvMetadata
		if json.Unmarshal(raw, &meta) != nil || meta.Version != DocumentVersion {
			return Document{}, fmt.Errorf("invalid app package metadata")
		}
		doc.Title, doc.Description, doc.SourceURL, doc.SourceID, doc.Level = meta.Title, meta.Description, meta.SourceURL, meta.SourceID, meta.Level
	}
	return doc, nil
}

func indexArchive(reader *zip.Reader) (map[string]archiveEntry, error) {
	if len(reader.File) == 0 || len(reader.File) > maxMediaCount+32 {
		return nil, fmt.Errorf("APKG has an invalid number of ZIP entries")
	}
	entries := make(map[string]archiveEntry, len(reader.File))
	var expanded uint64
	for _, file := range reader.File {
		name := file.Name
		if name == "" || strings.ContainsAny(name, `/\\:`) || name == "." || name == ".." || strings.TrimSpace(name) != name {
			return nil, fmt.Errorf("APKG contains unsafe ZIP entry name %q", name)
		}
		if file.FileInfo().IsDir() {
			return nil, fmt.Errorf("APKG contains unexpected directory entry %q", name)
		}
		key := strings.ToLower(name)
		if _, duplicate := entries[key]; duplicate {
			return nil, fmt.Errorf("APKG contains duplicate ZIP entry %q", name)
		}
		if file.UncompressedSize64 > MaxExpandedBytes {
			return nil, fmt.Errorf("APKG ZIP entry %q is oversized", name)
		}
		expanded += file.UncompressedSize64
		if expanded > MaxExpandedBytes {
			return nil, fmt.Errorf("APKG expands beyond %d MiB", MaxExpandedBytes>>20)
		}
		entries[key] = archiveEntry{file: file, name: name}
	}
	return entries, nil
}

func packageVersion(entries map[string]archiveEntry) (int, error) {
	meta, ok := entries["meta"]
	if !ok {
		if _, modern := entries["collection.anki21"]; modern {
			return legacy2, nil
		}
		if _, legacy := entries["collection.anki2"]; legacy {
			return legacy1, nil
		}
		return 0, fmt.Errorf("APKG is missing Anki package metadata and a legacy collection")
	}
	data, err := readZipEntry(meta.file, 64<<10)
	if err != nil {
		return 0, fmt.Errorf("read Anki package metadata: %w", err)
	}
	version, err := parseMetaVersion(data)
	if err != nil {
		return 0, err
	}
	if version < legacy1 || version > latest {
		return 0, fmt.Errorf("unsupported Anki package version %d", version)
	}
	return version, nil
}

func readZipEntry(file *zip.File, limit int64) ([]byte, error) {
	if file.UncompressedSize64 > uint64(limit) {
		return nil, fmt.Errorf("entry exceeds size limit")
	}
	r, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("entry exceeds size limit")
	}
	if uint64(len(data)) != file.UncompressedSize64 {
		return nil, fmt.Errorf("entry size does not match ZIP directory")
	}
	return data, nil
}

func decodeZstdLimited(compressed []byte, limit int64) ([]byte, error) {
	decoder, err := zstd.NewReader(bytes.NewReader(compressed),
		zstd.WithDecoderConcurrency(1),
		zstd.WithDecoderMaxMemory(uint64(limit)),
		zstd.WithDecoderMaxWindow(uint64(limit)),
	)
	if err != nil {
		return nil, err
	}
	defer decoder.Close()
	data, err := io.ReadAll(io.LimitReader(decoder, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("zstd content exceeds size limit")
	}
	return data, nil
}

func parseMetaVersion(data []byte) (int, error) {
	var version int
	if err := walkProto(data, func(field int, wire int, value uint64, bytesValue []byte) error {
		if field == 1 {
			if wire != 0 || value > 255 {
				return fmt.Errorf("invalid Anki package metadata version")
			}
			version = int(value)
		}
		return nil
	}); err != nil {
		return 0, fmt.Errorf("invalid Anki package metadata: %w", err)
	}
	if version == 0 {
		return 0, fmt.Errorf("Anki package metadata has no version")
	}
	return version, nil
}

func walkProto(data []byte, visit func(field int, wire int, value uint64, bytesValue []byte) error) error {
	for len(data) != 0 {
		key, n := readUvarint(data)
		if n <= 0 || key>>3 == 0 || key>>3 > 1<<29-1 {
			return fmt.Errorf("malformed protobuf tag")
		}
		data = data[n:]
		field, wire := int(key>>3), int(key&7)
		var value uint64
		var blob []byte
		switch wire {
		case 0:
			var used int
			value, used = readUvarint(data)
			if used <= 0 {
				return fmt.Errorf("malformed protobuf varint")
			}
			data = data[used:]
		case 1:
			if len(data) < 8 {
				return fmt.Errorf("truncated protobuf fixed64")
			}
			data = data[8:]
		case 2:
			length, used := readUvarint(data)
			if used <= 0 || length > uint64(len(data)-used) {
				return fmt.Errorf("truncated protobuf bytes")
			}
			blob = data[used : used+int(length)]
			data = data[used+int(length):]
		case 5:
			if len(data) < 4 {
				return fmt.Errorf("truncated protobuf fixed32")
			}
			data = data[4:]
		default:
			return fmt.Errorf("unsupported protobuf wire type %d", wire)
		}
		if err := visit(field, wire, value, blob); err != nil {
			return err
		}
	}
	return nil
}

func readUvarint(data []byte) (uint64, int) {
	var value uint64
	for i, b := range data {
		if i == 10 || (i == 9 && b > 1) {
			return 0, -1
		}
		value |= uint64(b&0x7f) << (7 * i)
		if b < 0x80 {
			return value, i + 1
		}
	}
	return 0, 0
}

func readPackageMedia(entries map[string]archiveEntry, version int) ([]Media, map[string]Media, error) {
	manifest, ok := entries["media"]
	if !ok {
		return []Media{}, map[string]Media{}, nil
	}
	manifestBytes, err := readZipEntry(manifest.file, 4<<20)
	if err != nil {
		return nil, nil, fmt.Errorf("read APKG media manifest: %w", err)
	}
	if version == latest {
		manifestBytes, err = decodeZstdLimited(manifestBytes, 4<<20)
		if err != nil {
			return nil, nil, fmt.Errorf("decompress APKG media manifest: %w", err)
		}
	}
	var mediaEntries []mediaManifestEntry
	if version == latest {
		mediaEntries, err = parseModernMediaManifest(manifestBytes)
	} else {
		mediaEntries, err = parseLegacyMediaManifest(manifestBytes)
	}
	if err != nil {
		return nil, nil, err
	}
	if len(mediaEntries) > maxMediaCount {
		return nil, nil, fmt.Errorf("APKG has too many media files")
	}
	media := make([]Media, 0, len(mediaEntries))
	byName := make(map[string]Media, len(mediaEntries))
	indices := make(map[string]struct{}, len(mediaEntries))
	var total int64
	for position, entry := range mediaEntries {
		if err := validateMediaName(entry.name); err != nil {
			return nil, nil, fmt.Errorf("APKG media manifest: %w", err)
		}
		nameKey := strings.ToLower(entry.name)
		if _, duplicate := byName[nameKey]; duplicate {
			return nil, nil, fmt.Errorf("APKG has duplicate media name %q", entry.name)
		}
		index := uint64(position)
		if entry.legacyName != nil {
			index = *entry.legacyName
		}
		zipName := strconv.FormatUint(index, 10)
		zipEntry, exists := entries[strings.ToLower(zipName)]
		if !exists {
			return nil, nil, fmt.Errorf("APKG is missing media entry %q", zipName)
		}
		if _, duplicate := indices[strings.ToLower(zipName)]; duplicate {
			return nil, nil, fmt.Errorf("APKG media manifest references entry %q more than once", zipName)
		}
		indices[strings.ToLower(zipName)] = struct{}{}
		remaining := MaxMediaBytes - total
		if remaining < 0 {
			return nil, nil, fmt.Errorf("APKG media exceeds %d MiB", MaxMediaBytes>>20)
		}
		packed, readErr := readZipEntry(zipEntry.file, remaining)
		if readErr != nil {
			return nil, nil, fmt.Errorf("read APKG media %q: %w", entry.name, readErr)
		}
		data := packed
		if version == latest {
			data, err = decodeZstdLimited(packed, remaining)
			if err != nil {
				return nil, nil, fmt.Errorf("decompress APKG media %q: %w", entry.name, err)
			}
		}
		if entry.size != 0 && uint64(len(data)) != entry.size {
			return nil, nil, fmt.Errorf("APKG media %q size does not match manifest", entry.name)
		}
		if len(entry.sha1) != 0 {
			sum := sha1.Sum(data)
			if !bytes.Equal(sum[:], entry.sha1) {
				return nil, nil, fmt.Errorf("APKG media %q checksum mismatch", entry.name)
			}
		}
		total += int64(len(data))
		if total > MaxMediaBytes {
			return nil, nil, fmt.Errorf("APKG media exceeds %d MiB", MaxMediaBytes>>20)
		}
		mimeType := mediaMIME(entry.name, data)
		item := Media{Name: entry.name, MIME: mimeType, Data: data}
		media = append(media, item)
		byName[nameKey] = item
	}
	if len(mediaEntries) > 0 {
		for key, archive := range entries {
			if isNumericEntry(archive.name) {
				if _, ok := indices[key]; !ok {
					return nil, nil, fmt.Errorf("APKG contains unlisted media entry %q", archive.name)
				}
			}
		}
	}
	return media, byName, nil
}

func isNumericEntry(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func parseLegacyMediaManifest(data []byte) ([]mediaManifestEntry, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		return nil, fmt.Errorf("invalid legacy APKG media manifest")
	}
	result := []mediaManifestEntry{}
	seen := map[uint64]struct{}{}
	for dec.More() {
		keyToken, keyErr := dec.Token()
		if keyErr != nil {
			return nil, fmt.Errorf("invalid legacy APKG media manifest: %w", keyErr)
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, fmt.Errorf("invalid legacy APKG media index")
		}
		index, parseErr := strconv.ParseUint(key, 10, 32)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid legacy APKG media index %q", key)
		}
		if _, duplicate := seen[index]; duplicate {
			return nil, fmt.Errorf("duplicate legacy APKG media index %q", key)
		}
		seen[index] = struct{}{}
		var name string
		if err = dec.Decode(&name); err != nil {
			return nil, fmt.Errorf("invalid legacy APKG media filename: %w", err)
		}
		idx := index
		result = append(result, mediaManifestEntry{name: name, legacyName: &idx})
	}
	if _, err = dec.Token(); err != nil {
		return nil, fmt.Errorf("invalid legacy APKG media manifest: %w", err)
	}
	var trailing any
	if err = dec.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("invalid legacy APKG media manifest trailing data")
	}
	return result, nil
}

func parseModernMediaManifest(data []byte) ([]mediaManifestEntry, error) {
	var result []mediaManifestEntry
	err := walkProto(data, func(field int, wire int, _ uint64, blob []byte) error {
		if field != 1 {
			return nil
		}
		if wire != 2 {
			return fmt.Errorf("invalid APKG media manifest entry")
		}
		entry := mediaManifestEntry{}
		err := walkProto(blob, func(entryField int, entryWire int, value uint64, valueBytes []byte) error {
			switch entryField {
			case 1:
				if entryWire != 2 {
					return fmt.Errorf("invalid media filename field")
				}
				entry.name = string(valueBytes)
			case 2:
				if entryWire != 0 {
					return fmt.Errorf("invalid media size field")
				}
				entry.size = value
			case 3:
				if entryWire != 2 {
					return fmt.Errorf("invalid media hash field")
				}
				entry.sha1 = append([]byte(nil), valueBytes...)
			case 255:
				if entryWire != 0 {
					return fmt.Errorf("invalid legacy media index field")
				}
				idx := value
				entry.legacyName = &idx
			}
			return nil
		})
		if err != nil {
			return err
		}
		if entry.name == "" {
			return fmt.Errorf("APKG media manifest entry is missing a name")
		}
		if entry.size > MaxMediaBytes {
			return fmt.Errorf("APKG media entry %q is oversized", entry.name)
		}
		result = append(result, entry)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid modern APKG media manifest: %w", err)
	}
	return result, nil
}

func readAnkiCollection(filename string, collection []byte, mediaByName map[string]Media) (Document, error) {
	if len(collection) < 16 || string(collection[:16]) != "SQLite format 3\x00" {
		return Document{}, fmt.Errorf("APKG collection is not a SQLite database")
	}
	tmp, err := os.MkdirTemp("", "anki-transfer-*")
	if err != nil {
		return Document{}, fmt.Errorf("prepare read-only collection: %w", err)
	}
	defer os.RemoveAll(tmp)
	path := filepath.Join(tmp, "collection.anki2")
	if err = os.WriteFile(path, collection, 0600); err != nil {
		return Document{}, fmt.Errorf("stage collection database: %w", err)
	}
	uriPath := filepath.ToSlash(path)
	if !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	fileURL := (&url.URL{Scheme: "file", Path: uriPath}).String()
	fileURL += "?mode=ro&immutable=1&_pragma=query_only(1)&_pragma=trusted_schema(OFF)&_pragma=temp_store(MEMORY)&_pragma=busy_timeout(1000)"
	db, err := sql.Open("sqlite", fileURL)
	if err != nil {
		return Document{}, fmt.Errorf("open collection database: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return Document{}, fmt.Errorf("open collection database: %w", err)
	}
	if err = validateSQLiteSchema(ctx, db); err != nil {
		return Document{}, err
	}
	return extractCollection(ctx, db, filename, mediaByName)
}

func validateSQLiteSchema(ctx context.Context, db *sql.DB) error {
	var objectCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM sqlite_schema`).Scan(&objectCount); err != nil {
		return fmt.Errorf("inspect Anki collection schema: %w", err)
	}
	if objectCount > 256 {
		return fmt.Errorf("Anki collection schema contains too many objects")
	}
	rows, err := db.QueryContext(ctx, `SELECT type, name, sql FROM sqlite_schema ORDER BY name`)
	if err != nil {
		return fmt.Errorf("inspect Anki collection schema: %w", err)
	}
	defer rows.Close()
	tables := map[string]bool{}
	for rows.Next() {
		var objectType, name string
		var sqlText sql.NullString
		if err := rows.Scan(&objectType, &name, &sqlText); err != nil {
			return err
		}
		if objectType == "view" || objectType == "trigger" {
			return fmt.Errorf("Anki collection contains unsupported SQL object %q", name)
		}
		if objectType == "table" {
			if !sqlText.Valid || !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sqlText.String)), "CREATE TABLE") {
				return fmt.Errorf("Anki collection contains an unsupported table %q", name)
			}
			tables[strings.ToLower(name)] = true
		}
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("inspect Anki collection schema: %w", err)
	}
	for _, required := range []string{"col", "notes", "cards"} {
		if !tables[required] {
			return fmt.Errorf("Anki collection is missing table %q", required)
		}
	}
	return nil
}

func extractCollection(ctx context.Context, db *sql.DB, filename string, mediaByName map[string]Media) (Document, error) {
	var modelsValue, decksValue any
	var schemaVersion int
	if err := db.QueryRowContext(ctx, `SELECT models, decks, ver FROM col LIMIT 1`).Scan(&modelsValue, &decksValue, &schemaVersion); err != nil {
		return Document{}, fmt.Errorf("read Anki collection metadata: %w", err)
	}
	modelsJSON, err := sqliteText(modelsValue)
	if err != nil {
		return Document{}, fmt.Errorf("invalid Anki model metadata: %w", err)
	}
	decksJSON, _ := sqliteText(decksValue)
	notetypes, normalized, err := readNotetypes(ctx, db)
	if err != nil {
		return Document{}, err
	}
	if !normalized {
		notetypes, err = parseLegacyNotetypes(modelsJSON)
		if err != nil {
			return Document{}, err
		}
	}
	if len(notetypes) == 0 {
		return Document{}, fmt.Errorf("Anki collection has no note types")
	}
	for id, nt := range notetypes {
		if nt.kind == 1 {
			return Document{}, fmt.Errorf("%w: cloze notetype %d cannot be represented by this deck format", ErrUnsupportedNote, id)
		}
		if nt.templates > 1 {
			return Document{}, fmt.Errorf("%w: notetype %d generates multiple card templates", ErrUnsupportedNote, id)
		}
	}
	if schemaVersion > 0 && schemaVersion < 11 {
		return Document{}, fmt.Errorf("unsupported Anki collection schema version %d", schemaVersion)
	}
	if schemaVersion > 18 {
		return Document{}, fmt.Errorf("unsupported Anki collection schema version %d", schemaVersion)
	}
	rows, err := db.QueryContext(ctx, `SELECT n.id, n.mid, n.flds, c.ord FROM notes n JOIN cards c ON c.nid=n.id ORDER BY n.id, c.ord LIMIT ?`, MaxCards+1)
	if err != nil {
		return Document{}, fmt.Errorf("read Anki cards: %w", err)
	}
	type noteRow struct {
		id, model int64
		fields    string
		ord       int
	}
	notes := make([]noteRow, 0, 128)
	counts := make(map[int64]int)
	for rows.Next() {
		var row noteRow
		if err := rows.Scan(&row.id, &row.model, &row.fields, &row.ord); err != nil {
			rows.Close()
			return Document{}, fmt.Errorf("read Anki card: %w", err)
		}
		counts[row.id]++
		if counts[row.id] > 1 || row.ord != 0 {
			rows.Close()
			return Document{}, fmt.Errorf("%w: reverse or secondary Anki cards cannot be represented", ErrUnsupportedNote)
		}
		notes = append(notes, row)
		if len(notes) > MaxCards {
			rows.Close()
			return Document{}, fmt.Errorf("Anki deck exceeds %d cards", MaxCards)
		}
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return Document{}, fmt.Errorf("read Anki cards: %w", err)
	}
	if err = rows.Close(); err != nil {
		return Document{}, err
	}
	doc := Document{Version: DocumentVersion, Title: collectionTitle(filename, decksJSON, ctx, db), Cards: []Card{}, Media: []Media{}}
	for i, row := range notes {
		if i%256 == 0 {
			if err := ctx.Err(); err != nil {
				return Document{}, fmt.Errorf("read Anki collection: %w", err)
			}
		}
		nt, ok := notetypes[row.model]
		if !ok {
			return Document{}, fmt.Errorf("Anki note %d references missing note type %d", row.id, row.model)
		}
		fieldValues := strings.Split(row.fields, "\x1f")
		if len(fieldValues) > 512 {
			return Document{}, fmt.Errorf("Anki note %d has too many fields", row.id)
		}
		if len(nt.fields) > 0 && len(fieldValues) != len(nt.fields) {
			return Document{}, fmt.Errorf("Anki note %d field count does not match its note type", row.id)
		}
		if len(fieldValues) < 2 {
			return Document{}, fmt.Errorf("Anki note %d has fewer than two fields", row.id)
		}
		frontIndex, answerIndex, meaningIndexes := fieldIndexes(nt.fields, len(fieldValues))
		if hasClozeSyntax(strings.Join(fieldValues, "\n")) {
			return Document{}, fmt.Errorf("%w: cloze deletion in Anki note %d cannot be represented", ErrUnsupportedNote, row.id)
		}
		front, frontMedia, err := sanitizeAnkiField(fieldValues[frontIndex], mediaByName)
		if err != nil {
			return Document{}, fmt.Errorf("Anki note %d front field: %w", row.id, err)
		}
		answer, answerMedia, err := sanitizeAnkiField(fieldValues[answerIndex], mediaByName)
		if err != nil {
			return Document{}, fmt.Errorf("Anki note %d answer field: %w", row.id, err)
		}
		card := Card{Front: front}
		card.Answers = splitEncodedAnswers(answer)
		meaningParts := []string{}
		for _, fieldIndex := range meaningIndexes {
			value, _, fieldErr := sanitizeAnkiField(fieldValues[fieldIndex], mediaByName)
			if fieldErr != nil {
				return Document{}, fmt.Errorf("Anki note %d field %d: %w", row.id, fieldIndex+1, fieldErr)
			}
			if strings.TrimSpace(value) != "" {
				meaningParts = append(meaningParts, value)
			}
		}
		card.Meaning = strings.Join(meaningParts, "\n")
		refs := append(frontMedia, answerMedia...)
		for _, fieldIndex := range meaningIndexes {
			_, fieldRefs, fieldErr := sanitizeAnkiField(fieldValues[fieldIndex], mediaByName)
			if fieldErr != nil {
				return Document{}, fmt.Errorf("Anki note %d field %d: %w", row.id, fieldIndex+1, fieldErr)
			}
			refs = append(refs, fieldRefs...)
		}
		if err := assignCardMedia(&card, refs, i+1); err != nil {
			return Document{}, err
		}
		if strings.TrimSpace(card.Front) == "" || len(card.Answers) == 0 || strings.TrimSpace(card.Answers[0]) == "" {
			return Document{}, fmt.Errorf("Anki note %d does not contain a usable front and answer", row.id)
		}
		doc.Cards = append(doc.Cards, card)
	}
	doc.Title = strings.TrimSpace(doc.Title)
	if doc.Title == "" {
		doc.Title = titleFromFilename(filename)
	}
	return doc, nil
}

func sqliteText(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("expected text, got %T", value)
	}
}

func readNotetypes(ctx context.Context, db *sql.DB) (map[int64]ankiNotetype, bool, error) {
	var exists int
	err := db.QueryRowContext(ctx, `SELECT 1 FROM sqlite_schema WHERE type='table' AND name='notetypes'`).Scan(&exists)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("inspect Anki note types: %w", err)
	}
	for _, table := range []string{"fields", "templates"} {
		var found int
		if err := db.QueryRowContext(ctx, `SELECT 1 FROM sqlite_schema WHERE type='table' AND name=?`, table).Scan(&found); err != nil {
			return nil, false, fmt.Errorf("modern Anki schema is incomplete: %w", err)
		}
	}
	rows, err := db.QueryContext(ctx, `SELECT id, config FROM notetypes ORDER BY id LIMIT 1024`)
	if err != nil {
		return nil, false, fmt.Errorf("read Anki note types: %w", err)
	}
	result := map[int64]ankiNotetype{}
	for rows.Next() {
		var id int64
		var config []byte
		if err := rows.Scan(&id, &config); err != nil {
			rows.Close()
			return nil, false, err
		}
		kind, err := parseNotetypeKind(config)
		if err != nil {
			rows.Close()
			return nil, false, fmt.Errorf("invalid Anki note type %d: %w", id, err)
		}
		result[id] = ankiNotetype{kind: kind}
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, false, err
	}
	if err = rows.Close(); err != nil {
		return nil, false, err
	}
	for id, nt := range result {
		rows, err = db.QueryContext(ctx, `SELECT ord, name FROM fields WHERE ntid=? ORDER BY ord LIMIT 512`, id)
		if err != nil {
			return nil, false, fmt.Errorf("read Anki fields: %w", err)
		}
		for rows.Next() {
			var ord int
			var name string
			if err := rows.Scan(&ord, &name); err != nil {
				rows.Close()
				return nil, false, err
			}
			if ord < 0 || ord > len(nt.fields) {
				rows.Close()
				return nil, false, fmt.Errorf("Anki note type %d has invalid field order", id)
			}
			if ord == len(nt.fields) {
				nt.fields = append(nt.fields, name)
			} else {
				nt.fields[ord] = name
			}
		}
		if err = rows.Err(); err != nil {
			rows.Close()
			return nil, false, err
		}
		rows.Close()
		var count int
		if err = db.QueryRowContext(ctx, `SELECT count(*) FROM templates WHERE ntid=?`, id).Scan(&count); err != nil {
			return nil, false, fmt.Errorf("read Anki card templates: %w", err)
		}
		nt.templates = count
		result[id] = nt
	}
	return result, true, nil
}

func parseNotetypeKind(config []byte) (int, error) {
	kind := 0
	err := walkProto(config, func(field int, wire int, value uint64, _ []byte) error {
		if field == 1 {
			if wire != 0 || value > 1 {
				return fmt.Errorf("invalid note type kind")
			}
			kind = int(value)
		}
		return nil
	})
	return kind, err
}

func parseLegacyNotetypes(raw string) (map[int64]ankiNotetype, error) {
	var models map[string]struct {
		Type int `json:"type"`
		Flds []struct {
			Name string `json:"name"`
			Ord  int    `json:"ord"`
		} `json:"flds"`
		Tmpls []json.RawMessage `json:"tmpls"`
	}
	if err := json.Unmarshal([]byte(raw), &models); err != nil {
		return nil, fmt.Errorf("invalid legacy Anki note types: %w", err)
	}
	result := map[int64]ankiNotetype{}
	for key, model := range models {
		id, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid legacy Anki note type id")
		}
		if len(model.Flds) > 512 || len(model.Tmpls) > 512 {
			return nil, fmt.Errorf("Anki note type %d has too many fields or templates", id)
		}
		sort.Slice(model.Flds, func(i, j int) bool { return model.Flds[i].Ord < model.Flds[j].Ord })
		fields := make([]string, len(model.Flds))
		for i, field := range model.Flds {
			if field.Ord != i {
				return nil, fmt.Errorf("Anki note type %d has invalid field order", id)
			}
			fields[i] = field.Name
		}
		result[id] = ankiNotetype{fields: fields, kind: model.Type, templates: len(model.Tmpls)}
	}
	return result, nil
}

func fieldIndexes(names []string, count int) (front, answer int, meaning []int) {
	if len(names) != count {
		names = make([]string, count)
	}
	front, answer = 0, -1
	for i, name := range names {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "front", "question", "prompt", "word", "kanji", "term", "expression":
			front = i
		}
	}
	for i, name := range names {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "back", "answer", "translation", "reading", "definition", "response":
			if i != front {
				answer = i
				break
			}
		}
		if answer >= 0 {
			break
		}
	}
	if answer < 0 {
		for i := range names {
			if i != front {
				answer = i
				break
			}
		}
	}
	if answer < 0 {
		answer = front
	}
	for i := range names {
		if i != front && i != answer {
			meaning = append(meaning, i)
		}
	}
	return
}

func collectionTitle(filename, decksJSON string, ctx context.Context, db *sql.DB) string {
	var title string
	var exists int
	if db.QueryRowContext(ctx, `SELECT 1 FROM sqlite_schema WHERE type='table' AND name='decks'`).Scan(&exists) == nil {
		_ = db.QueryRowContext(ctx, `SELECT name FROM decks WHERE id != 1 ORDER BY id LIMIT 1`).Scan(&title)
	}
	if strings.TrimSpace(title) == "" {
		var decks map[string]struct {
			Name string `json:"name"`
		}
		if json.Unmarshal([]byte(decksJSON), &decks) == nil {
			ids := make([]int64, 0, len(decks))
			values := map[int64]string{}
			for key, deck := range decks {
				if id, err := strconv.ParseInt(key, 10, 64); err == nil && id != 1 {
					ids = append(ids, id)
					values[id] = deck.Name
				}
			}
			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
			if len(ids) > 0 {
				title = values[ids[0]]
			}
		}
	}
	if strings.TrimSpace(title) == "" {
		title = titleFromFilename(filename)
	}
	return title
}

func mediaMIME(name string, data []byte) string {
	if value := mime.TypeByExtension(strings.ToLower(filepath.Ext(name))); value != "" {
		return value
	}
	if len(data) == 0 {
		return "application/octet-stream"
	}
	return http.DetectContentType(data[:min(len(data), 512)])
}
