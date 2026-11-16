package places

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type cannedTransport map[string]string

func (c cannedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	body, ok := c[r.URL.Host]
	if !ok {
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}, Request: r}, nil
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}, Request: r}, nil
}

func lenientFetcher(wikidata string) *Fetcher {
	fetcher := NewFetcher("")
	fetcher.Gap = 0
	fetcher.HTTP.Transport = cannedTransport{"www.wikidata.org": wikidata}
	return fetcher
}

// Wikidata (formatversion=2) answers a title with no item with
// "missing": true. That used to fail JSON decoding and surface as
// "invalid JSON" instead of "not found".
func TestFetchMissingTitleIsNotFound(t *testing.T) {
	fetcher := lenientFetcher(`{"entities":{"-1":{"site":"enwiki","title":"Nope","missing":true}}}`)
	_, err := fetcher.Fetch(context.Background(), "", "Nope")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// MediaWiki serializes empty maps as [] which must not break the import.
func TestFetchToleratesEmptyArrays(t *testing.T) {
	fetcher := lenientFetcher(`{"entities":{"Q1":{"id":"Q1","labels":[],"sitelinks":[],"claims":[]}}}`)
	got, err := fetcher.Fetch(context.Background(), "Q1", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.WikidataID != "Q1" || got.Lat != nil || len(got.Labels) != 0 {
		t.Fatalf("unexpected result %+v", got)
	}
}

func TestCommonsMetadataToleratesOddValues(t *testing.T) {
	for _, in := range []string{`[]`, `null`, `{"Artist":{"value":"A"},"Count":{"value":3}}`} {
		var m commonsMetadata
		if err := m.UnmarshalJSON([]byte(in)); err != nil {
			t.Fatalf("%s: %v", in, err)
		}
	}
	var m commonsMetadata
	_ = m.UnmarshalJSON([]byte(`{"Artist":{"value":"A"},"Count":{"value":3}}`))
	if m["Artist"] != "A" || m["Count"] != "" {
		t.Fatalf("unexpected %v", m)
	}
}
