package ankitransfer

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"github.com/klauspost/compress/zstd"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func fixtureDocument(t *testing.T) Document {
	t.Helper()
	var picture bytes.Buffer
	if err := png.Encode(&picture, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return Document{Version: 1, Title: "日本語の練習", Description: "A quoted, multiline\nsource", SourceID: "n5", SourceURL: "https://example.test/source", Level: "N5",
		Cards: []Card{{Front: "猫、\"ねこ\"\n毎朝", Answers: []string{"ねこ", "猫"}, Meaning: "cat\nhewan", ImageURL: "media:cat.png", Audio: "media:cat.mp3"}},
		Media: []Media{{Name: "cat.png", Data: picture.Bytes()}, {Name: "cat.mp3", Data: append([]byte("ID3\x04\x00\x00\x00\x00\x00\x00"), make([]byte, 30)...)}}}
}

// Exercise modern package compression/protobuf plus normalized SQLite note
// types, rather than only round-tripping our own legacy export.
func TestAPKGModernPackageAndNormalizedSchema(t *testing.T) {
	doc := fixtureDocument(t)
	raw, err := Export(doc, "apkg")
	if err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := indexArchive(reader)
	if err != nil {
		t.Fatal(err)
	}
	collection, err := readZipEntry(entries["collection.anki2"].file, MaxExpandedBytes)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "modern.sqlite")
	if err := os.WriteFile(path, collection, 0600); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE notetypes(id integer PRIMARY KEY,config blob);
 CREATE TABLE fields(ntid integer,ord integer,name text);
 CREATE TABLE templates(ntid integer,ord integer);
 INSERT INTO notetypes SELECT DISTINCT mid, x'' FROM notes;
 INSERT INTO fields SELECT id,0,'Front' FROM notetypes;
 INSERT INTO fields SELECT id,1,'Back' FROM notetypes;
 INSERT INTO fields SELECT id,2,'Meaning' FROM notetypes;
 INSERT INTO templates SELECT id,0 FROM notetypes;
 UPDATE col SET models='',ver=18;`)
	db.Close()
	if err != nil {
		t.Fatal(err)
	}
	collection, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderConcurrency(1))
	if err != nil {
		t.Fatal(err)
	}
	defer encoder.Close()
	var out bytes.Buffer
	writer := zip.NewWriter(&out)
	add := func(name string, data []byte) {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = entry.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	add("meta", []byte{8, 3})
	add("collection.anki21b", encoder.EncodeAll(collection, nil))
	// Minimal protobuf wire encoding is kept local to this fixture.
	blob := func(tag byte, data []byte) []byte {
		value := []byte{tag}
		value = binary.AppendUvarint(value, uint64(len(data)))
		return append(value, data...)
	}
	var manifest []byte
	for index, media := range doc.Media {
		value := blob(10, []byte(media.Name))
		value = append(value, 16)
		value = binary.AppendUvarint(value, uint64(len(media.Data)))
		hash := sha1.Sum(media.Data)
		value = append(value, blob(26, hash[:])...)
		manifest = append(manifest, blob(10, value)...)
		add(strconv.Itoa(index), encoder.EncodeAll(media.Data, nil))
	}
	add("media", encoder.EncodeAll(manifest, nil))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	imported, err := Parse("modern.apkg", out.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(imported.Cards, doc.Cards) || len(imported.Media) != 2 {
		t.Fatalf("modern package lost content: %+v", imported)
	}
}

func TestTransferRoundTrip(t *testing.T) {
	for _, format := range []string{"apkg", "json", "csv"} {
		t.Run(format, func(t *testing.T) {
			original := fixtureDocument(t)
			raw, err := Export(original, format)
			if err != nil {
				t.Fatal(err)
			}
			imported, err := Parse("deck."+format, raw)
			if err != nil {
				t.Fatal(err)
			}
			if imported.Title != original.Title || imported.Level != original.Level || imported.SourceID != original.SourceID || imported.Description != original.Description {
				t.Fatalf("lost metadata: %+v", imported)
			}
			expected := original.Cards[0]
			if format == "csv" {
				expected.ImageURL = ""
				expected.Audio = ""
			}
			if len(imported.Cards) != 1 || !reflect.DeepEqual(imported.Cards[0], expected) {
				t.Fatalf("lost card: %#v, want %#v", imported.Cards, expected)
			}
			if format != "csv" {
				if len(imported.Media) != 2 {
					t.Fatal("missing media")
				}
				for _, item := range imported.Media {
					found := false
					for _, want := range original.Media {
						if item.Name == want.Name {
							found = bytes.Equal(item.Data, want.Data)
						}
					}
					if !found {
						t.Fatal("media changed", item.Name)
					}
				}
			}
		})
	}
}

func TestCSVHeadersAndMalformedInputs(t *testing.T) {
	doc, err := Parse("simple.csv", []byte("front,back\n猫,ねこ\n"))
	if err != nil || len(doc.Cards) != 1 || doc.Cards[0].Answers[0] != "ねこ" {
		t.Fatalf("simple CSV: %+v %v", doc, err)
	}
	for _, raw := range []string{"front,answers\n", "front,back\n猫\n", "front,answers\n猫,\"[bad\"\n"} {
		if _, err := Parse("bad.csv", []byte(raw)); err == nil {
			t.Fatalf("accepted malformed CSV %q", raw)
		}
	}
	for _, name := range []string{"../escape.png", "CON.png", "bad/name.png"} {
		doc := fixtureDocument(t)
		doc.Media[0].Name = name
		raw, _ := json.Marshal(doc)
		if _, err := Parse("bad.json", raw); err == nil {
			t.Fatal("accepted unsafe media name", name)
		}
	}
	doc = fixtureDocument(t)
	doc.Media[0].Data = []byte("<svg onload='evil()'>")
	doc.Media[0].MIME = "image/png"
	raw, _ := json.Marshal(doc)
	if _, err := Parse("bad.json", raw); err == nil {
		t.Fatal("trusted spoofed MIME")
	}
	if _, err := Parse("huge.json", make([]byte, MaxUploadBytes+1)); err == nil {
		t.Fatal("accepted oversized upload")
	}
}

func TestAPKGRejectsArchiveTraversalAndCloze(t *testing.T) {
	var data bytes.Buffer
	z := zip.NewWriter(&data)
	f, _ := z.Create("../collection.anki2")
	f.Write([]byte("bad"))
	z.Close()
	if _, err := Parse("bad.apkg", data.Bytes()); err == nil {
		t.Fatal("accepted ZIP traversal")
	}
	doc := fixtureDocument(t)
	doc.Cards[0].Front = "{{c1::猫}}"
	raw, err := Export(doc, "apkg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse("cloze.apkg", raw); err == nil || !strings.Contains(err.Error(), "cloze") {
		t.Fatal("cloze should be rejected", err)
	}
}
