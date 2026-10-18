package dictionary

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"nihongo-master-backend/internal/database"
	"nihongo-master-backend/internal/testdb"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const wordFixture = `{"meta":{"status":200},"data":[{"slug":"猫","is_common":true,"japanese":[{"word":"猫","reading":"ねこ"}],"senses":[{"english_definitions":["cat"]}],"attribution":{"jmdict":true}}]}`
const kanjiFixture = `{"kanji":"猫","stroke_count":11,"meanings":["cat"],"on_readings":["ビョウ"]}`

func persistentTestClient(t *testing.T, path string) (*Client, *sql.DB) {
	t.Helper()
	db, err := sql.Open("mysql", path)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	orm, err := database.FromSQL(db)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := NewSQLCache(orm)
	if err != nil {
		t.Fatal(err)
	}
	c := NewClient()
	c.Persistence = cache
	return c, db
}
func TestDatabaseCacheSurvivesRestartAndArchivesUniqueEntries(t *testing.T) {
	path := testdb.DSN(t)
	c, db := persistentTestClient(t, path)
	c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "kanjiapi.dev" {
			return response(200, kanjiFixture), nil
		}
		return response(200, wordFixture), nil
	})
	for _, q := range []string{"neko", "猫"} {
		if _, err := c.Search(context.Background(), q, 1, "", false); err != nil {
			t.Fatal(err)
		}
		c.next = time.Time{}
	}
	if _, err := c.Kanji(context.Background(), "猫"); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM dictionary_responses`).Scan(&count); err != nil || count != 3 {
		t.Fatalf("responses not retained: %d %v", count, err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM dictionary_entries`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("archive duplicated words or lost kanji: %d %v", count, err)
	}
	var raw string
	if err := db.QueryRow(`SELECT body FROM dictionary_entries WHERE source='jisho.org' AND source_entry_id='猫'`).Scan(&raw); err != nil || !strings.Contains(raw, `"jmdict":true`) {
		t.Fatal("source attribution not retained")
	}
	db.Close()
	restarted, _ := persistentTestClient(t, path)
	restarted.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		t.Error("restart caused another external request")
		return nil, errors.New("offline")
	})
	// Different clients/users share the same persisted query, not per-user copies.
	if result, err := restarted.Search(context.Background(), "neko", 1, "", false); err != nil || len(result.Entries) != 1 {
		t.Fatalf("database hit failed after restart: %v", err)
	}
	if _, err := restarted.Kanji(context.Background(), "猫"); err != nil {
		t.Fatal(err)
	}
}
func TestOldDatabaseResponseRemainsPermanent(t *testing.T) {
	c, db := persistentTestClient(t, testdb.DSN(t))
	calls := 0
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { calls++; return response(200, wordFixture), nil })
	if _, err := c.Search(context.Background(), "cat", 1, "", false); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE dictionary_responses SET fetched_at=?`, time.Now().Add(-31*24*time.Hour).UnixMilli()); err != nil {
		t.Fatal(err)
	}
	c.cache = map[string]cached{}
	c.next = time.Time{}
	if _, err := c.Search(context.Background(), "cat", 1, "", false); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("permanent response unexpectedly refreshed")
	}
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM dictionary_entries`).Scan(&count)
	if count != 1 {
		t.Fatal("refresh duplicated archive entries")
	}
}
func TestInvalidPayloadNeverPersistsAndWritesAreAtomic(t *testing.T) {
	c, db := persistentTestClient(t, testdb.DSN(t))
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		return response(200, `{"meta":{"status":200},"data":[{"slug":"bad"}]}`), nil
	})
	if _, err := c.Search(context.Background(), "bad", 1, "", false); !errors.Is(err, ErrUnavailable) {
		t.Fatal("bad payload accepted")
	}
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM dictionary_responses`).Scan(&count)
	if count != 0 {
		t.Fatal("bad payload persisted")
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_archive BEFORE INSERT ON dictionary_entries FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='test database failure'`); err != nil {
		t.Fatal(err)
	}
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { return response(200, wordFixture), nil })
	c.next = time.Time{}
	if _, err := c.Search(context.Background(), "cat", 1, "", false); !errors.Is(err, ErrUnavailable) {
		t.Fatal("write failure was reported as success")
	}
	db.QueryRow(`SELECT COUNT(*) FROM dictionary_responses`).Scan(&count)
	if count != 0 {
		t.Fatal("partial response committed without its archive")
	}
	db.Close()
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		t.Error("DB outage triggered an unnecessary upstream request")
		return nil, nil
	})
	if _, err := c.Search(context.Background(), "dog", 1, "", false); !errors.Is(err, ErrUnavailable) {
		t.Fatal("DB outage was ignored")
	}
}
func TestFilterAndPageCacheKeysDoNotMixResults(t *testing.T) {
	c, _ := persistentTestClient(t, testdb.DSN(t))
	calls := 0
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { calls++; return response(200, wordFixture), nil })
	for _, input := range []struct {
		page   int
		jlpt   string
		common bool
	}{{1, "", false}, {2, "", false}, {1, "N5", false}, {1, "N5", true}, {1, "N5", true}} {
		if _, err := c.Search(context.Background(), "cat", input.page, input.jlpt, input.common); err != nil {
			t.Fatal(err)
		}
		c.next = time.Time{}
	}
	if calls != 4 {
		t.Fatal("page/filter keys were mixed or duplicate fetched")
	}
}

func TestMariaDBRetainsMoreThanFiveThousandResponses(t *testing.T) {
	c, db := persistentTestClient(t, testdb.DSN(t))
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`INSERT INTO dictionary_responses(id,cache_key,source,body,fetched_at) VALUES(?,?,?,?,?)`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 5001; i++ {
		if _, err := stmt.Exec(uuid.NewString(), responseKey(fmt.Sprint(i)), "jisho.org", []byte(wordFixture), i); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := c.Persistence.save(context.Background(), "https://jisho.org/api/v1/search/words?keyword=cat", cached{body: []byte(wordFixture), at: time.Now()}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM dictionary_responses`).Scan(&count); err != nil || count != 5002 {
		t.Fatalf("permanent cache retention: %d %v", count, err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM dictionary_entries WHERE source_entry_id='猫'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("archive lost: %d %v", count, err)
	}
}
