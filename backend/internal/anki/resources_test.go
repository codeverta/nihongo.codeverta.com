package anki

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestAllKotobaFilesAreLosslessAndStudyResourcesAreValid(t *testing.T) {
	raw, err := files.ReadFile("resources/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Revision string `json:"revision"`
		Files    []struct {
			File  string `json:"file"`
			SHA   string `json:"sha256"`
			Bytes int    `json:"bytes"`
		} `json:"files"`
	}
	if err = json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) != 191 || manifest.Revision != Revision() {
		t.Fatal("incomplete source mirror")
	}
	for _, entry := range manifest.Files {
		raw, err := files.ReadFile("resources/source/" + entry.File + ".gz")
		if err != nil {
			t.Fatal(err)
		}
		reader, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.New()
		count, err := io.Copy(hash, reader)
		reader.Close()
		if err != nil || count != int64(entry.Bytes) || hex.EncodeToString(hash.Sum(nil)) != entry.SHA {
			t.Fatal("source changed:", entry.File, err)
		}
	}
	decks, err := List("", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(decks) != 191 {
		t.Fatal("catalog incomplete")
	}
	total, unavailable := 0, 0
	for _, deck := range decks {
		if deck.Description == "" || deck.Instructions == "" || deck.SourceFile == "" {
			t.Fatal("missing deck details", deck.ID)
		}
		if !deck.Available {
			unavailable++
			if deck.UnavailableReason == "" {
				t.Fatal("missing unavailable explanation")
			}
			continue
		}
		for start := 0; start < deck.CardCount; start += deck.ChunkSize {
			raw, err := files.ReadFile(fmt.Sprintf("resources/cards/%s/%06d.json.gz", deck.ID, start/deck.ChunkSize))
			if err != nil {
				t.Fatal(err)
			}
			reader, err := gzip.NewReader(bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}
			var cards []Card
			err = json.NewDecoder(io.LimitReader(reader, 4<<20)).Decode(&cards)
			reader.Close()
			if err != nil {
				t.Fatal(deck.ID, err)
			}
			if len(cards) == 0 || len(cards) > deck.ChunkSize {
				t.Fatal("invalid chunk")
			}
			for index, card := range cards {
				if card.Index != start+index || strings.TrimSpace(card.Front) == "" || len(card.Answers) == 0 {
					t.Fatal("invalid study card", deck.ID, card.Index)
				}
				total++
			}
		}
		if _, err = ReadCard(deck.ID, deck.CardCount-1); err != nil {
			t.Fatal(err)
		}
	}
	if total != 1416940 || unavailable != 4 {
		t.Fatal("unexpected import counts", total, unavailable)
	}
	if _, err = ReadCard("../source/wakayama", 0); err == nil {
		t.Fatal("unsafe deck path")
	}
	if _, err = ReadCard("n5", -1); err == nil {
		t.Fatal("negative index accepted")
	}
	wakayama, err := Get("wakayama")
	if err != nil || !strings.Contains(wakayama.Description, "prefektur") {
		t.Fatal("missing place explanation")
	}
	if !strings.Contains(License(), "MIT License") {
		t.Fatal("missing source license")
	}
}
