package places

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestPrefectureAt(t *testing.T) {
	for _, tc := range []struct {
		name     string
		lat, lng float64
		want     string
	}{
		{"Fushimi Inari", 34.9671, 135.7727, "26"},
		{"Shibuya Crossing", 35.6595, 139.7005, "13"},
		{"Sapporo", 43.0642, 141.3469, "01"},
		{"Itsukushima Shrine", 34.2960, 132.3198, "34"},
		{"Hashima Island (tiny, nearest outline)", 32.6278, 129.7386, "42"},
		{"Shuri Castle (real position, not inset)", 26.2170, 127.7195, "47"},
		{"Pacific Ocean", 30.0, 142.0, ""},
	} {
		if got := PrefectureAt(tc.lat, tc.lng); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
	if !ValidPrefecture("47") || ValidPrefecture("48") || ValidPrefecture("") {
		t.Fatal("prefecture validation")
	}
}

func TestParseInput(t *testing.T) {
	for _, tc := range []struct{ input, qid, title string }{
		{"Q1194434", "Q1194434", ""},
		{"q42", "Q42", ""},
		{"https://www.wikidata.org/wiki/Q42", "Q42", ""},
		{"https://en.wikipedia.org/wiki/Fushimi_Inari-taisha", "", "Fushimi Inari-taisha"},
		{"https://en.m.wikipedia.org/wiki/D%C5%8Dtonbori#History", "", "Dōtonbori"},
		{"  Shibuya Crossing ", "", "Shibuya Crossing"},
	} {
		qid, title, err := ParseInput(tc.input)
		if err != nil || qid != tc.qid || title != tc.title {
			t.Errorf("%q: %q %q %v", tc.input, qid, title, err)
		}
	}
	for _, bad := range []string{"", "https://example.com/wiki/X", "a|b", "javascript:alert(1)://"} {
		if _, _, err := ParseInput(bad); !errors.Is(err, ErrInvalid) {
			t.Errorf("%q should be invalid", bad)
		}
	}
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func pngBytes() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 40, 30))
	for x := 0; x < 40; x++ {
		for y := 0; y < 30; y++ {
			img.Set(x, y, color.RGBA{200, 30, 30, 255})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

const wikidataFixture = `{"entities":{"Q1194434":{"id":"Q1194434","labels":{"en":{"value":"Fushimi Inari-taisha"},"ja":{"value":"伏見稲荷大社"}},
"sitelinks":{"enwiki":{"title":"Fushimi Inari-taisha"},"jawiki":{"title":"伏見稲荷大社"},"enwikivoyage":{"title":"Fushimi"}},
"claims":{"P625":[{"rank":"normal","mainsnak":{"datavalue":{"value":{"latitude":34.9671,"longitude":135.7727}}}}],
"P18":[{"rank":"deprecated","mainsnak":{"datavalue":{"value":"Old.jpg"}}},{"rank":"normal","mainsnak":{"datavalue":{"value":"Fushimi Inari Taisha.jpg"}}}]}}}}`

type fakeWikimedia struct {
	mu       sync.Mutex
	requests []string
	agents   []string
	imageLic string
}

func (f *fakeWikimedia) transport() http.RoundTripper {
	return roundTrip(func(r *http.Request) (*http.Response, error) {
		f.mu.Lock()
		f.requests = append(f.requests, r.URL.String())
		f.agents = append(f.agents, r.Header.Get("User-Agent"))
		f.mu.Unlock()
		body, status := "", 200
		switch {
		case r.URL.Host == "www.wikidata.org":
			body = wikidataFixture
			if r.URL.Query().Get("titles") == "Nope" {
				body = `{"entities":{"-1":{"missing":""}}}`
			}
		case strings.Contains(r.URL.Path, "/page/summary/Fushimi_Inari-taisha"):
			body = `{"type":"standard","extract":"Fushimi Inari-taisha is the head shrine of the kami Inari.","content_urls":{"desktop":{"page":"https://en.wikipedia.org/wiki/Fushimi_Inari-taisha"}}}`
		case strings.Contains(r.URL.Path, "/page/summary/Fushimi"):
			body = `{"type":"standard","extract":"Fushimi is a district in southern Kyoto.","content_urls":{"desktop":{"page":"https://en.wikivoyage.org/wiki/Fushimi"}}}`
		case r.URL.Host == "commons.wikimedia.org":
			lic := f.imageLic
			body = `{"query":{"pages":[{"imageinfo":[{"url":"https://upload.wikimedia.org/a.png","thumburl":"https://upload.wikimedia.org/thumb/a.png","descriptionurl":"https://commons.wikimedia.org/wiki/File:Fushimi_Inari_Taisha.jpg","extmetadata":{"Artist":{"value":"<a href=\"//commons.wikimedia.org/wiki/User:X\">Jane &amp; Doe</a>"},"LicenseShortName":{"value":"` + lic + `"},"LicenseUrl":{"value":"https://creativecommons.org/licenses/by-sa/4.0"}}}]}]}}`
		case r.URL.Host == "upload.wikimedia.org":
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(bytes.NewReader(pngBytes()))}, nil
		default:
			status, body = 404, `{}`
		}
		return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
}

type fakeImages struct{ puts int }

func (f *fakeImages) Put(_ context.Context, jpeg []byte) (string, error) {
	f.puts++
	if len(jpeg) < 3 || jpeg[0] != 0xFF || jpeg[1] != 0xD8 {
		return "", errors.New("not a JPEG")
	}
	return "https://cdn.example/map/images/a.jpg", nil
}

func newTestService(t *testing.T) (*Service, *fakeWikimedia, *fakeImages) {
	t.Helper()
	wm := &fakeWikimedia{imageLic: "CC BY-SA 4.0"}
	fetcher := NewFetcher("")
	fetcher.HTTP.Transport = wm.transport()
	fetcher.Gap = 0
	images := &fakeImages{}
	return &Service{Store: NewMemoryStore(), Fetcher: fetcher, Images: images, Reading: func(s string) string { return "ふしみいなりたいしゃ" }}, wm, images
}

func TestImportFetchesOnceAndStoresDraft(t *testing.T) {
	s, wm, images := newTestService(t)
	spot, err := s.Import(context.Background(), "https://en.wikipedia.org/wiki/Fushimi_Inari-taisha", "shrine")
	if err != nil {
		t.Fatal(err)
	}
	if spot.WikidataID != "Q1194434" || spot.Prefecture != "26" || spot.Published || !spot.Publishable() {
		t.Fatalf("unexpected spot %+v", spot)
	}
	if spot.NameJA != "伏見稲荷大社" || spot.Reading != "ふしみいなりたいしゃ" || spot.NameID != "" {
		t.Fatalf("names: %q %q %q", spot.NameJA, spot.Reading, spot.NameID)
	}
	if !strings.HasPrefix(spot.SummaryEN, "Fushimi Inari-taisha is") || spot.TipsEN == "" || spot.WikipediaJA == "" {
		t.Fatalf("texts not fetched: %+v", spot)
	}
	if images.puts != 1 || spot.ImageAuthor != "Jane & Doe" || spot.ImageLicense != "CC BY-SA 4.0" || spot.ImageSourceURL == "" {
		t.Fatalf("image: puts=%d %+v", images.puts, spot)
	}
	for _, agent := range wm.agents {
		if agent != DefaultUserAgent {
			t.Fatalf("unexpected user agent %q", agent)
		}
	}
	calls := len(wm.requests)
	if _, err := s.Import(context.Background(), "Q1194434", "shrine"); !errors.Is(err, ErrExists) {
		t.Fatalf("duplicate import should be rejected: %v", err)
	}
	if len(wm.requests) != calls {
		t.Fatal("a known Wikidata ID must not hit the sources again")
	}
	if _, err := s.Import(context.Background(), "Nope", "shrine"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing title: %v", err)
	}
	if _, err := s.Import(context.Background(), "Q5", "unknown"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid category: %v", err)
	}
}

func TestImportSkipsUnlicensedImages(t *testing.T) {
	s, wm, images := newTestService(t)
	wm.imageLic = ""
	spot, err := s.Import(context.Background(), "Q1194434", "shrine")
	if err != nil || spot.ImageURL != "" || images.puts != 0 {
		t.Fatalf("image without a stated license must be skipped: %v %+v", err, spot)
	}
}

func TestRefreshKeepsManualWork(t *testing.T) {
	s, _, _ := newTestService(t)
	spot, err := s.Import(context.Background(), "Q1194434", "shrine")
	if err != nil {
		t.Fatal(err)
	}
	lat, lng := 34.9675, 135.7730
	edited, err := s.Update(context.Background(), spot.ID, Edit{
		Category: "culture", Lat: &lat, Lng: &lng, ManualLocation: true,
		NameEN: spot.NameEN, NameID: "Kuil Fushimi Inari", NameJA: spot.NameJA, Reading: "ふしみいなりたいしゃ",
		ManualSummaryID: "Kuil dengan ribuan gerbang torii merah.",
		Phrases:         []Phrase{{Japanese: "鳥居", Reading: "とりい", MeaningID: "gerbang kuil"}},
		Published:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !edited.Published || edited.Prefecture != "26" {
		t.Fatalf("edit: %+v", edited)
	}
	refreshed, err := s.Refresh(context.Background(), spot.ID)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.ManualSummaryID == "" || refreshed.Category != "culture" || !refreshed.Published || *refreshed.Lat != lat || len(refreshed.Phrases()) != 1 {
		t.Fatalf("refresh lost manual work: %+v", refreshed)
	}
	if refreshed.NameID != "" {
		t.Fatal("names are source fields and follow Wikidata on refresh")
	}
}

func TestUpdateValidation(t *testing.T) {
	s, _, _ := newTestService(t)
	spot, _ := s.Import(context.Background(), "Q1194434", "shrine")
	base := Edit{Category: "shrine", NameEN: "X", Published: false}
	bad := []Edit{
		{Category: "nope"},
		func() Edit { e := base; e.Phrases = []Phrase{{Japanese: "鳥居"}}; return e }(),
		func() Edit { e := base; e.ManualLocation = true; return e }(),
		func() Edit {
			e := base
			lat, lng := 10.0, 100.0
			e.ManualLocation, e.Lat, e.Lng = true, &lat, &lng
			return e
		}(),
	}
	for i, e := range bad {
		if _, err := s.Update(context.Background(), spot.ID, e); !errors.Is(err, ErrInvalid) {
			t.Errorf("case %d should be invalid: %v", i, err)
		}
	}
	// A spot without coordinates cannot be published.
	s.Store.Save(context.Background(), &Spot{ID: "x", WikidataID: "Q9", Category: "food", NameEN: "Ramen"})
	if _, err := s.Update(context.Background(), "x", Edit{Category: "food", NameEN: "Ramen", Published: true}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("publishing without location: %v", err)
	}
}

func TestPublicTextFallback(t *testing.T) {
	spot := Spot{ID: "a", Category: "shrine", Prefecture: "26", NameEN: "Fushimi Inari-taisha",
		SummaryEN: "English summary.", TipsEN: "English tips.", WikipediaEN: "https://en.wikipedia.org/wiki/F", WikidataID: "Q1"}
	_ = spot.SetPhrases([]Phrase{{Japanese: "鳥居", MeaningID: "gerbang"}})
	public := spot.Public()
	if public.Name["id"] != "Fushimi Inari-taisha" || public.Summary["id"].SourceLang != "en" || public.Summary["en"].SourceLang != "en" {
		t.Fatalf("fallback: %+v", public)
	}
	spot.ManualSummaryID = "Ringkasan manual."
	if got := spot.Public().Summary["id"]; got.Text != "Ringkasan manual." || !got.Manual || got.SourceLang != "id" {
		t.Fatalf("manual text should win: %+v", got)
	}
	if len(public.Sources) != 2 || public.Sources[1].License != "CC0" {
		t.Fatalf("sources: %+v", public.Sources)
	}
}
