package ankitransfer

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func exportAPKG(doc Document) ([]byte, error) {
	mediaByName := make(map[string]Media, len(doc.Media))
	mediaNames := make([]string, 0, len(doc.Media))
	var mediaBytes int64
	for _, media := range doc.Media {
		mediaByName[strings.ToLower(media.Name)] = media
		mediaNames = append(mediaNames, media.Name)
		mediaBytes += int64(len(media.Data))
	}
	sort.Slice(mediaNames, func(i, j int) bool { return mediaNames[i] < mediaNames[j] })
	if mediaBytes > MaxMediaBytes {
		return nil, fmt.Errorf("media exceeds %d MiB", MaxMediaBytes>>20)
	}

	tmp, err := os.MkdirTemp("", "anki-transfer-export-*")
	if err != nil {
		return nil, fmt.Errorf("prepare Anki package: %w", err)
	}
	defer os.RemoveAll(tmp)
	dbPath := filepath.Join(tmp, "collection.anki2")
	if err = writeAnkiDatabase(dbPath, doc, mediaByName); err != nil {
		return nil, err
	}
	stat, err := os.Stat(dbPath)
	if err != nil {
		return nil, err
	}
	if stat.Size()+mediaBytes > MaxExpandedBytes {
		return nil, fmt.Errorf("APKG expands beyond %d MiB", MaxExpandedBytes>>20)
	}
	dbFile, err := os.Open(dbPath)
	if err != nil {
		return nil, err
	}
	defer dbFile.Close()

	var out bytes.Buffer
	z := zip.NewWriter(&out)
	collection, err := z.Create("collection.anki2")
	if err == nil {
		_, err = io.Copy(collection, dbFile)
	}
	if err != nil {
		_ = z.Close()
		return nil, fmt.Errorf("write APKG collection: %w", err)
	}
	manifest := make(map[string]string, len(mediaNames))
	for index, name := range mediaNames {
		zipName := strconv.Itoa(index)
		manifest[zipName] = name
		entry, createErr := z.Create(zipName)
		if createErr != nil {
			_ = z.Close()
			return nil, createErr
		}
		if _, err = entry.Write(mediaByName[strings.ToLower(name)].Data); err != nil {
			_ = z.Close()
			return nil, fmt.Errorf("write APKG media %q: %w", name, err)
		}
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		_ = z.Close()
		return nil, err
	}
	manifestEntry, err := z.Create("media")
	if err != nil {
		_ = z.Close()
		return nil, err
	}
	if _, err = manifestEntry.Write(manifestBytes); err != nil {
		_ = z.Close()
		return nil, err
	}
	metadata, _ := json.Marshal(csvMetadata{Version: doc.Version, Title: doc.Title, Description: doc.Description, SourceURL: doc.SourceURL, SourceID: doc.SourceID, Level: doc.Level})
	metadataEntry, err := z.Create("nihongo-master.json")
	if err != nil {
		_ = z.Close()
		return nil, err
	}
	if _, err = metadataEntry.Write(metadata); err != nil {
		_ = z.Close()
		return nil, err
	}
	if err = z.Close(); err != nil {
		return nil, fmt.Errorf("finish APKG: %w", err)
	}
	return out.Bytes(), nil
}

func writeAnkiDatabase(path string, doc Document, mediaByName map[string]Media) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("create Anki collection: %w", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()
	if _, err = db.Exec(`PRAGMA page_size=512; PRAGMA journal_mode=DELETE; PRAGMA foreign_keys=OFF;`); err != nil {
		return fmt.Errorf("prepare Anki collection: %w", err)
	}
	const schema = `
CREATE TABLE col (
 id integer PRIMARY KEY, crt integer NOT NULL, mod integer NOT NULL, scm integer NOT NULL,
 ver integer NOT NULL, dty integer NOT NULL, usn integer NOT NULL, ls integer NOT NULL,
 conf text NOT NULL, models text NOT NULL, decks text NOT NULL, dconf text NOT NULL, tags text NOT NULL
);
CREATE TABLE notes (
 id integer PRIMARY KEY, guid text NOT NULL, mid integer NOT NULL, mod integer NOT NULL,
 usn integer NOT NULL, tags text NOT NULL, flds text NOT NULL, sfld integer NOT NULL,
 csum integer NOT NULL, flags integer NOT NULL, data text NOT NULL
);
CREATE TABLE cards (
 id integer PRIMARY KEY, nid integer NOT NULL, did integer NOT NULL, ord integer NOT NULL,
 mod integer NOT NULL, usn integer NOT NULL, type integer NOT NULL, queue integer NOT NULL,
 due integer NOT NULL, ivl integer NOT NULL, factor integer NOT NULL, reps integer NOT NULL,
 lapses integer NOT NULL, left integer NOT NULL, odue integer NOT NULL, odid integer NOT NULL,
 flags integer NOT NULL, data text NOT NULL
);
CREATE TABLE revlog (
 id integer PRIMARY KEY, cid integer NOT NULL, usn integer NOT NULL, ease integer NOT NULL,
 ivl integer NOT NULL, lastIvl integer NOT NULL, factor integer NOT NULL, time integer NOT NULL,
 type integer NOT NULL
);
CREATE TABLE graves (usn integer NOT NULL, oid integer NOT NULL, type integer NOT NULL);
CREATE INDEX ix_notes_usn ON notes (usn);
CREATE INDEX ix_cards_usn ON cards (usn);
CREATE INDEX ix_revlog_usn ON revlog (usn);
CREATE INDEX ix_cards_nid ON cards (nid);
CREATE INDEX ix_cards_sched ON cards (did, queue, due);
CREATE INDEX ix_revlog_cid ON revlog (cid);
CREATE INDEX ix_notes_csum ON notes (csum);
`
	if _, err = db.Exec(schema); err != nil {
		return fmt.Errorf("create Anki collection schema: %w", err)
	}
	now := time.Now().UTC()
	nowMillis := now.UnixMilli()
	nowSeconds := now.Unix()
	deckID, modelID := nowMillis, nowMillis+1
	model := makeBasicModel(modelID, deckID, nowSeconds)
	decks := map[string]any{strconv.FormatInt(deckID, 10): map[string]any{
		"id": deckID, "name": doc.Title, "mod": nowSeconds, "usn": -1, "desc": deckDescription(doc),
		"dyn": 0, "collapsed": false, "browserCollapsed": false, "conf": 1,
		"extendNew": 10, "extendRev": 50, "newToday": []int{0, 0}, "revToday": []int{0, 0},
		"lrnToday": []int{0, 0}, "timeToday": []int{0, 0}, "terms": []any{}, "resched": true,
	}}
	deckConfig := map[string]any{"1": map[string]any{
		"id": 1, "name": "Default", "mod": nowSeconds, "usn": -1, "maxTaken": 60,
		"autoplay": true, "timer": 0, "replayq": true,
		"new":   map[string]any{"bury": true, "delays": []int{1, 10}, "initialFactor": 2500, "ints": []int{1, 4, 7}, "order": 1, "perDay": 20},
		"rev":   map[string]any{"bury": true, "ease4": 1.3, "ivlFct": 1, "maxIvl": 36500, "perDay": 200},
		"lapse": map[string]any{"delays": []int{10}, "leechAction": 0, "leechFails": 8, "minInt": 1, "mult": 0},
	}}
	models := map[string]any{strconv.FormatInt(modelID, 10): model}
	conf := map[string]any{"activeDecks": []int64{deckID}, "curDeck": deckID, "curModel": modelID,
		"nextPos": len(doc.Cards), "newBury": true, "timeLim": 0, "sortType": "noteFld", "sortBackwards": false,
		"addToCur": true, "collapseTime": 1200, "dueCounts": true, "estTimes": true,
		"schedVer": 1, "dayOffset": 0,
	}
	confJSON, _ := json.Marshal(conf)
	modelsJSON, _ := json.Marshal(models)
	decksJSON, _ := json.Marshal(decks)
	dconfJSON, _ := json.Marshal(deckConfig)
	if _, err = db.Exec(`INSERT INTO col VALUES(1,?,?,?,?,?,?,?,?,?,?,?,?)`, nowSeconds, nowMillis, nowMillis, 11, 0, 0, 0, string(confJSON), string(modelsJSON), string(decksJSON), string(dconfJSON), `{}`); err != nil {
		return fmt.Errorf("initialize Anki collection: %w", err)
	}
	return insertAnkiCards(db, doc, mediaByName, modelID, deckID, nowSeconds, nowMillis)
}

func makeBasicModel(modelID, deckID, now int64) map[string]any {
	field := func(name string, ord int) map[string]any {
		return map[string]any{"name": name, "ord": ord, "sticky": false, "rtl": false, "font": "Arial", "size": 20, "media": []any{}}
	}
	tmpl := map[string]any{"name": "Card 1", "ord": 0, "qfmt": "{{Front}}", "afmt": "{{FrontSide}}<hr id=answer>{{Back}}{{#Meaning}}<br><br>{{Meaning}}{{/Meaning}}", "bqfmt": "", "bafmt": "", "did": deckID, "bfont": "", "bsize": 0}
	return map[string]any{
		"id": modelID, "name": "Nihongo Master Transfer", "type": 0, "mod": now, "usn": -1,
		"sortf": 0, "did": deckID, "tmpls": []any{tmpl},
		"flds":      []any{field("Front", 0), field("Back", 1), field("Meaning", 2)},
		"css":       ".card { font-family: arial; font-size: 20px; text-align: center; color: black; background-color: white; }",
		"latexPre":  "\\documentclass[12pt]{article}\n\\special{papersize=3in,5in}\n\\usepackage[utf8]{inputenc}\n\\usepackage{amssymb,amsmath}\n\\pagestyle{empty}\n\\setlength{\\parindent}{0in}\n\\begin{document}\n",
		"latexPost": "\\end{document}", "req": []any{[]any{0, "all", []int{0}}},
		"originalStockKind": 1,
	}
}

func deckDescription(doc Document) string {
	parts := []string{}
	if doc.Description != "" {
		parts = append(parts, "<div>"+html.EscapeString(doc.Description)+"</div>")
	}
	if doc.Level != "" {
		parts = append(parts, "<div>Level: "+html.EscapeString(doc.Level)+"</div>")
	}
	if doc.SourceURL != "" {
		parts = append(parts, "<div>Source: <a href=\""+html.EscapeString(doc.SourceURL)+"\">"+html.EscapeString(doc.SourceURL)+"</a></div>")
	}
	if doc.SourceID != "" {
		parts = append(parts, "<div>Source ID: "+html.EscapeString(doc.SourceID)+"</div>")
	}
	return strings.Join(parts, "\n")
}

func insertAnkiCards(db *sql.DB, doc Document, mediaByName map[string]Media, modelID, deckID, nowSeconds, nowMillis int64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin Anki card insert: %w", err)
	}
	defer tx.Rollback()
	noteStmt, err := tx.Prepare(`INSERT INTO notes (id,guid,mid,mod,usn,tags,flds,sfld,csum,flags,data) VALUES(?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("prepare Anki note insert: %w", err)
	}
	defer noteStmt.Close()
	cardStmt, err := tx.Prepare(`INSERT INTO cards (id,nid,did,ord,mod,usn,type,queue,due,ivl,factor,reps,lapses,left,odue,odid,flags,data) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("prepare Anki card insert: %w", err)
	}
	defer cardStmt.Close()
	for i, card := range doc.Cards {
		front := htmlField(card.Front)
		if card.ImageURL != "" {
			front += "<br>" + imageTag(card.ImageURL, mediaByName)
		}
		if card.Audio != "" {
			front += "<br>" + audioReference(card.Audio, mediaByName)
		}
		answerParts := make([]string, len(card.Answers))
		for j, answer := range card.Answers {
			answerParts[j] = htmlField(answer)
		}
		back := strings.Join(answerParts, "<!--"+answerSeparatorMarker+"--><br>")
		meaning := htmlField(card.Meaning)
		if _, err = noteStmt.Exec(nowMillis+int64(i), newGUID(), modelID, nowSeconds, -1, "", front+"\x1f"+back+"\x1f"+meaning, front, checksum(front), 0, ""); err != nil {
			return fmt.Errorf("write Anki note %d: %w", i+1, err)
		}
		if _, err = cardStmt.Exec(nowMillis+int64(i), nowMillis+int64(i), deckID, 0, nowSeconds, -1, 0, 0, i+1, 0, 2500, 0, 0, 0, 0, 0, 0, ""); err != nil {
			return fmt.Errorf("write Anki card %d: %w", i+1, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("finish Anki collection: %w", err)
	}
	return nil
}

func htmlField(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(html.EscapeString(value), "\r\n", "\n"), "\n", "<br>")
}

func imageTag(ref string, mediaByName map[string]Media) string {
	if strings.HasPrefix(ref, "media:") {
		ref = strings.TrimPrefix(ref, "media:")
		_ = mediaByName[strings.ToLower(ref)]
	}
	return `<img src="` + html.EscapeString(ref) + `">`
}

func audioReference(ref string, mediaByName map[string]Media) string {
	if strings.HasPrefix(ref, "media:") {
		ref = strings.TrimPrefix(ref, "media:")
		_ = mediaByName[strings.ToLower(ref)]
		return "[sound:" + ref + "]"
	}
	return `<audio src="` + html.EscapeString(ref) + `"></audio>`
}

func checksum(value string) uint32 {
	sum := sha1.Sum([]byte(value))
	return binary.BigEndian.Uint32(sum[:4])
}

func newGUID() string {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data[:]))
}
