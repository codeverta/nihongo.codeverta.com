package api

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"nihongo-master-backend/internal/anki"
	"nihongo-master-backend/internal/jobfetch"
	"nihongo-master-backend/internal/jobs"
	"nihongo-master-backend/internal/store"
)

func decodeAccountPayload(t *testing.T, raw string) any {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode %q: %v", raw, err)
	}
	return payload
}

func marshalAccountPayload(t *testing.T, payload any) string {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func normalizeAccountJSON(t *testing.T, raw string) string {
	t.Helper()
	return marshalAccountPayload(t, decodeAccountPayload(t, raw))
}

// encodeAccountValue returns value as the middleware would see it: marshalled
// to JSON and decoded with UseNumber.
func encodeAccountValue(t *testing.T, value any) any {
	t.Helper()
	return decodeAccountPayload(t, marshalAccountPayload(t, value))
}

func TestTranslateAccountPayloadCases(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		input   string
		want    string
		changed bool
	}{
		{
			name: "gamification achievements by id",
			path: "/v1/me/gamification",
			input: `{"data":{"xp":120,"level":1,"timezone":"Asia/Tokyo","achievements":[
				{"id":"first-lesson","name":"First Step","description":"Lulus satu kuis lesson dengan nilai minimal 80%","icon":"footprints","progress":1,"target":1,"unlockedAt":"2026-01-02T03:04:05Z"},
				{"id":"lessons-10","name":"Pelajar Tekun","description":"Lulus 10 lesson berbeda","icon":"book","progress":2,"target":10},
				{"id":"custom","name":"Pelajar Tekun","description":"Lulus 10 lesson berbeda","icon":"book","progress":2,"target":10}]}}`,
			want: `{"data":{"xp":120,"level":1,"timezone":"Asia/Tokyo","achievements":[
				{"id":"first-lesson","name":"First Step","description":"Pass one lesson quiz with a score of at least 80%.","icon":"footprints","progress":1,"target":1,"unlockedAt":"2026-01-02T03:04:05Z"},
				{"id":"lessons-10","name":"Dedicated Learner","description":"Pass 10 different lessons.","icon":"book","progress":2,"target":10},
				{"id":"custom","name":"Pelajar Tekun","description":"Lulus 10 lesson berbeda","icon":"book","progress":2,"target":10}]}}`,
			changed: true,
		},
		{
			name:    "achievements array",
			path:    "/v1/me/achievements",
			input:   `{"data":[{"id":"streak-7","name":"7-Day Streak","description":"Selesaikan kuis atau ujian selama 7 hari berturut-turut","icon":"flame","progress":3,"target":7},{"id":"all-subjects","name":"Serba Bisa","description":"Lulus lesson di keempat bidang","icon":"star","progress":1,"target":4}]}`,
			want:    `{"data":[{"id":"streak-7","name":"7-Day Streak","description":"Complete a quiz or exam for 7 consecutive days.","icon":"flame","progress":3,"target":7},{"id":"all-subjects","name":"All-Rounder","description":"Pass lessons in all four subjects.","icon":"star","progress":1,"target":4}]}`,
			changed: true,
		},
		{
			name:    "profile keeps user-entered values",
			path:    "/v1/me/profile",
			input:   `{"data":{"id":"u1","displayName":"Pelajar Tekun","email":"a@example.com","target":"JLPT N5","level":1,"xp":10,"streakDays":2,"entitlement":{"tier":"free","isLifetime":false}}}`,
			want:    `{"data":{"id":"u1","displayName":"Pelajar Tekun","email":"a@example.com","target":"JLPT N5","level":1,"xp":10,"streakDays":2,"entitlement":{"tier":"free","isLifetime":false}}}`,
			changed: false,
		},
		{
			name:    "progress is unchanged",
			path:    "/v1/me/progress/vocabulary-n5-1",
			input:   `{"data":{"lessonId":"vocabulary-n5-1","completion":100,"lastPosition":3,"completed":true,"updatedAt":"2026-01-02T03:04:05Z"}}`,
			want:    `{"data":{"lessonId":"vocabulary-n5-1","completion":100,"lastPosition":3,"completed":true,"updatedAt":"2026-01-02T03:04:05Z"}}`,
			changed: false,
		},
		{
			name: "learning history titles",
			path: "/v1/me/learning-history",
			input: `{"data":{"items":[
				{"kind":"grammar","lessonId":"grammar-n3-01","title":"Bagi seseorang dan sebagai suatu peran","level":"N3","position":2,"total":5,"completed":false,"locked":true,"available":true,"accessedAt":"2026-01-02T03:04:05Z"},
				{"kind":"vocabulary","lessonId":"vocabulary-x","title":"Judul pelajaran yang tidak dikenal","level":"N5","position":0,"total":8,"completed":true,"locked":false,"available":true,"accessedAt":"2026-01-01T03:04:05Z"},
				{"kind":"kanji","lessonId":"kanji-n3-01","title":"日本語の漢字","level":"N3","position":1,"total":4,"completed":false,"locked":false,"available":true,"accessedAt":"2026-01-01T01:04:05Z"}]}}`,
			want: `{"data":{"items":[
				{"kind":"grammar","lessonId":"grammar-n3-01","title":"From Someone's Perspective and in a Role","level":"N3","position":2,"total":5,"completed":false,"locked":true,"available":true,"accessedAt":"2026-01-02T03:04:05Z"},
				{"kind":"vocabulary","lessonId":"vocabulary-x","title":"Judul pelajaran yang tidak dikenal","level":"N5","position":0,"total":8,"completed":true,"locked":false,"available":true,"accessedAt":"2026-01-01T03:04:05Z"},
				{"kind":"kanji","lessonId":"kanji-n3-01","title":"日本語の漢字","level":"N3","position":1,"total":4,"completed":false,"locked":false,"available":true,"accessedAt":"2026-01-01T01:04:05Z"}]}}`,
			changed: true,
		},
		{
			name:    "learning history single item",
			path:    "/v1/me/learning-history/grammar/grammar-n3-01",
			input:   `{"data":{"kind":"grammar","lessonId":"grammar-n3-01","title":"Bagi seseorang dan sebagai suatu peran","level":"N3","position":2,"total":5}}`,
			want:    `{"data":{"kind":"grammar","lessonId":"grammar-n3-01","title":"From Someone's Perspective and in a Role","level":"N3","position":2,"total":5}}`,
			changed: true,
		},
		{
			name:    "plans features",
			path:    "/v1/plans",
			input:   `{"data":[{"id":"free","name":"Free","billingType":"free","priceIdr":0,"features":["JLPT N5–N4 dasar","Daily practice","Progress dasar","Fitur baru yang belum dikenal"],"recommended":false},{"id":"premium-lifetime","name":"Lifetime Access","billingType":"lifetime","priceIdr":99000,"features":["Materi N3–N1 yang tersedia","Seluruh tes yang tersedia","Restore pembelian","Sekali bayar tanpa perpanjangan"],"recommended":true}]}`,
			want:    `{"data":[{"id":"free","name":"Free","billingType":"free","priceIdr":0,"features":["Basic JLPT N5–N4","Daily practice","Basic progress","Fitur baru yang belum dikenal"],"recommended":false},{"id":"premium-lifetime","name":"Lifetime Access","billingType":"lifetime","priceIdr":99000,"features":["Available N3–N1 materials","All available tests","Restore purchases","One-time payment with no renewal"],"recommended":true}]}`,
			changed: true,
		},
		{
			name:    "billing status is unchanged",
			path:    "/v1/billing",
			input:   `{"data":{"productId":"nihongo_master_lifetime","accountToken":"token","priceIdr":99000,"billingType":"lifetime","platforms":["android","ios"],"entitlement":{"tier":"premium","isLifetime":true}}}`,
			want:    `{"data":{"productId":"nihongo_master_lifetime","accountToken":"token","priceIdr":99000,"billingType":"lifetime","platforms":["android","ios"],"entitlement":{"tier":"premium","isLifetime":true}}}`,
			changed: false,
		},
		{
			name:    "pricing is unchanged",
			path:    "/v1/pricing",
			input:   `{"data":{"lifetimePriceIdr":99000}}`,
			want:    `{"data":{"lifetimePriceIdr":99000}}`,
			changed: false,
		},
		{
			name:    "auth session is unchanged",
			path:    "/v1/auth/login",
			input:   `{"data":{"token":"abc","user":{"id":"u1","displayName":"Budi","target":"JLPT N5","entitlement":{"tier":"free","isLifetime":false}}}}`,
			want:    `{"data":{"token":"abc","user":{"id":"u1","displayName":"Budi","target":"JLPT N5","entitlement":{"tier":"free","isLifetime":false}}}}`,
			changed: false,
		},
		{
			name:    "manga catalog is source content",
			path:    "/v1/manga",
			input:   `{"data":{"items":[{"id":"m1","title":"Judul Manga","summary":"Ringkasan dari situs sumber","genres":["Aksi"],"status":"ready"}],"total":1,"sourceName":"Rawkuma"}}`,
			want:    `{"data":{"items":[{"id":"m1","title":"Judul Manga","summary":"Ringkasan dari situs sumber","genres":["Aksi"],"status":"ready"}],"total":1,"sourceName":"Rawkuma"}}`,
			changed: false,
		},
		{
			name:    "app update policy is unchanged",
			path:    "/v1/app-update-policy",
			input:   `{"data":{"android":{"enabled":true,"latestBuild":2,"message":"Perbarui aplikasi sekarang."}}}`,
			want:    `{"data":{"android":{"enabled":true,"latestBuild":2,"message":"Perbarui aplikasi sekarang."}}}`,
			changed: false,
		},
		{
			name: "jobs metadata labels",
			path: "/v1/jobs/meta",
			input: `{"data":{
				"categories":[{"id":"nursing-care","label":"Kaigo"},{"id":"office","label":"Administrasi & Korporat"},{"id":"engineering","label":"IT & Engineering"},{"id":"zzz","label":"Kategori Baru"}],
				"employmentTypes":[{"id":"full-time","label":"Penuh waktu"},{"id":"unknown","label":"Tidak dicantumkan"}],
				"japaneseLevels":[{"id":"N3","label":"JLPT N3"},{"id":"not-required","label":"Jepang tidak diwajibkan"}],
				"visaOptions":[{"id":"work-permit-required","label":"Izin kerja wajib dimiliki"}],
				"sectors":[{"id":"railway","label":"Perkeretaapian"}],
				"remoteOptions":[{"id":"onsite","label":"Di lokasi"},{"id":"hybrid","label":"Hybrid"}],
				"prefectures":[{"id":"13","label":"Tokyo"}]}}`,
			want: `{"data":{
				"categories":[{"id":"nursing-care","label":"Nursing Care (Kaigo)"},{"id":"office","label":"Administration & Corporate"},{"id":"engineering","label":"IT & Engineering"},{"id":"zzz","label":"Kategori Baru"}],
				"employmentTypes":[{"id":"full-time","label":"Full-time"},{"id":"unknown","label":"Not specified"}],
				"japaneseLevels":[{"id":"N3","label":"JLPT N3"},{"id":"not-required","label":"Japanese not required"}],
				"visaOptions":[{"id":"work-permit-required","label":"Work permit required"}],
				"sectors":[{"id":"railway","label":"Railway"}],
				"remoteOptions":[{"id":"onsite","label":"On-site"},{"id":"hybrid","label":"Hybrid"}],
				"prefectures":[{"id":"13","label":"Tokyo"}]}}`,
			changed: true,
		},
		{
			name:    "job listing keeps external text",
			path:    "/v1/jobs",
			input:   `{"data":{"items":[{"id":"j1","title":"Staf gudang","company":"PT Contoh","category":"logistics","status":"active","sourceName":"Mujin · Karier resmi","description":"Deskripsi lowongan dalam bahasa Indonesia","salaryText":"Gaji bulanan","salaryMin":200000}],"total":1,"nextCursor":"","hasMore":false}}`,
			want:    `{"data":{"items":[{"id":"j1","title":"Staf gudang","company":"PT Contoh","category":"logistics","status":"active","sourceName":"Mujin · Official careers","description":"Deskripsi lowongan dalam bahasa Indonesia","salaryText":"Gaji bulanan","salaryMin":200000}],"total":1,"nextCursor":"","hasMore":false}}`,
			changed: true,
		},
		{
			name:    "job detail source name",
			path:    "/v1/jobs/abc",
			input:   `{"data":{"id":"abc","title":"Engineer","sourceName":"Findy · lowongan teknologi","sourceId":"findy"}}`,
			want:    `{"data":{"id":"abc","title":"Engineer","sourceName":"Findy · tech jobs","sourceId":"findy"}}`,
			changed: true,
		},
		{
			name:    "job detail unknown source name",
			path:    "/v1/jobs/abc",
			input:   `{"data":{"id":"abc","sourceName":"Sumber Baru · lowongan misterius"}}`,
			want:    `{"data":{"id":"abc","sourceName":"Sumber Baru · lowongan misterius"}}`,
			changed: false,
		},
		{
			name: "job sources",
			path: "/v1/jobs/sources",
			input: `{"data":{"items":[
				{"id":"tokyodev","name":"TokyoDev · lowongan teknologi","url":"https://www.tokyodev.com/","category":"engineering","scope":"Seluruh Jepang","mode":"directory","status":"directory_only","reason":"Tersedia sebagai direktori. Pengambilan lowongan otomatis dari situs ini belum tersedia."},
				{"id":"green","name":"Green","scope":"Tokyo","mode":"directory","status":"directory_only","reason":"Alasan baru yang belum dikenal"},
				{"id":"lever-mujininc","name":"Mujin · Karier resmi","scope":"Lowongan Jepang pada perusahaan ini","mode":"api","status":"blocked","reason":"Situs membatasi akses otomatis. Pengambilan dihentikan sampai jadwal berikutnya.","intervalHours":24}]}}`,
			want: `{"data":{"items":[
				{"id":"tokyodev","name":"TokyoDev · tech jobs","url":"https://www.tokyodev.com/","category":"engineering","scope":"All of Japan","mode":"directory","status":"directory_only","reason":"Available as a directory. Automatic job fetching from this site is not available yet."},
				{"id":"green","name":"Green","scope":"Tokyo","mode":"directory","status":"directory_only","reason":"Alasan baru yang belum dikenal"},
				{"id":"lever-mujininc","name":"Mujin · Official careers","scope":"Job listings in Japan from this company","mode":"api","status":"blocked","reason":"The site limits automated access. Fetching is paused until the next scheduled run.","intervalHours":24}]}}`,
			changed: true,
		},
		{
			name:    "anki credits",
			path:    "/v1/anki/credits",
			input:   `{"data":{"repository":"https://github.com/mistval/kotoba","revision":"r1","license":"MIT","notice":"Data dan arti mengikuti sumber Kotoba. Sebagian deck memuat atribusi sumber pihak ketiga pada detailnya. Tautan media eksternal tetap mengikuti ketersediaan penyedianya."}}`,
			want:    `{"data":{"repository":"https://github.com/mistval/kotoba","revision":"r1","license":"MIT","notice":"Data and meanings follow the Kotoba source. Some decks include third-party source attribution in their details. External media links remain subject to their providers' availability."}}`,
			changed: true,
		},
		{
			name: "anki decks keep category, card content and personal text",
			path: "/v1/anki/decks",
			input: `{"data":{"source":"Kotoba / mistval","revision":"r1","decks":[
				{"id":"words-nature","title":"Alam","category":"Kana & Kanji","level":"","description":"Susun kembali huruf di bagian depan menjadi kata bahasa Inggris. Bagian belakang memuat variasi jawaban dari sumber.","instructions":"Susun huruf menjadi kata.","originalName":"Alam","meaningLanguage":"Mengikuti sumber; kebanyakan arti menggunakan bahasa Inggris.","available":true,"unavailableReason":"","personal":false,"preview":[{"front":"Alam","answers":["Stasiun"],"meaning":"Nama & Geografi"}]},
				{"id":"user-abc","title":"Alam","category":"Personal","description":"Susun huruf menjadi kata.","instructions":"Ingat jawaban sebelum membalik kartu; cocokkan dengan semua jawaban yang tersedia.","personal":true,"available":true}]}}`,
			want: `{"data":{"source":"Kotoba / mistval","revision":"r1","decks":[
				{"id":"words-nature","title":"Nature","category":"Kana & Kanji","level":"","description":"Rearrange the letters on the front to form an English word. The back shows answer variants from the source.","instructions":"Arrange the letters to form a word.","originalName":"Alam","meaningLanguage":"As in the source; most meanings are in English.","available":true,"unavailableReason":"","personal":false,"preview":[{"front":"Alam","answers":["Stasiun"],"meaning":"Nama & Geografi"}]},
				{"id":"user-abc","title":"Alam","category":"Personal","description":"Susun huruf menjadi kata.","instructions":"Recall the answer before flipping the card; compare it with all available answers.","personal":true,"available":true}]}}`,
			changed: true,
		},
		{
			name:    "anki study session",
			path:    "/v1/anki/decks/words-nature/study",
			input:   `{"data":{"title":"Alam","level":"","requiredTier":"free","instructions":"Susun huruf menjadi kata.","deckId":"words-nature","cards":[{"index":0,"front":"Alam","answers":["自然"],"meaning":"Alam","version":1,"intervals":{"again":"1 menit","hard":"±2 jam","good":"1 hari","easy":"4 hari"}}],"stats":{"seen":0,"due":0,"new":1}}}`,
			want:    `{"data":{"title":"Nature","level":"","requiredTier":"free","instructions":"Arrange the letters to form a word.","deckId":"words-nature","cards":[{"index":0,"front":"Alam","answers":["自然"],"meaning":"Alam","version":1,"intervals":{"again":"1 min","hard":"±2 hr","good":"1 day","easy":"4 days"}}],"stats":{"seen":0,"due":0,"new":1}}}`,
			changed: true,
		},
		{
			name:    "personal anki study session keeps the user title",
			path:    "/v1/anki/decks/user-abc/study",
			input:   `{"data":{"title":"Alam","instructions":"Ingat jawaban sebelum membalik kartu; cocokkan dengan semua jawaban yang tersedia.","deckId":"user-abc","cards":[]}}`,
			want:    `{"data":{"title":"Alam","instructions":"Recall the answer before flipping the card; compare it with all available answers.","deckId":"user-abc","cards":[]}}`,
			changed: true,
		},
		{
			name:    "branding default tagline",
			path:    "/v1/branding",
			input:   `{"data":{"appName":"Nihongo Master","tagline":"Selangkah lebih dekat dengan Jepang","iconUrl":"","splashUrl":"","backgroundColor":"#FAFAFA"}}`,
			want:    `{"data":{"appName":"Nihongo Master","tagline":"One step closer to Japan","iconUrl":"","splashUrl":"","backgroundColor":"#FAFAFA"}}`,
			changed: true,
		},
		{
			name:    "branding custom tagline",
			path:    "/v1/branding",
			input:   `{"data":{"appName":"Selangkah","tagline":"Belajar setiap hari","backgroundColor":"#FAFAFA"}}`,
			want:    `{"data":{"appName":"Selangkah","tagline":"Belajar setiap hari","backgroundColor":"#FAFAFA"}}`,
			changed: false,
		},
		{
			name:    "paths outside the account groups",
			path:    "/v1/meow",
			input:   `{"data":{"achievements":[{"id":"first-lesson","name":"First Step","description":"Lulus satu kuis lesson dengan nilai minimal 80%"}]}}`,
			want:    `{"data":{"achievements":[{"id":"first-lesson","name":"First Step","description":"Lulus satu kuis lesson dengan nilai minimal 80%"}]}}`,
			changed: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := decodeAccountPayload(t, tt.input)
			if changed := translateAccountPayload(tt.path, payload); changed != tt.changed {
				t.Fatalf("changed = %v, want %v", changed, tt.changed)
			}
			if got, want := marshalAccountPayload(t, payload), normalizeAccountJSON(t, tt.want); got != want {
				t.Fatalf("payload mismatch\n got: %s\nwant: %s", got, want)
			}
			// A second pass must be a no-op.
			if translateAccountPayload(tt.path, payload) {
				t.Fatalf("second translation pass reported a change")
			}
		})
	}
}

func TestTranslateAccountPayloadPreservesNumbers(t *testing.T) {
	payload := decodeAccountPayload(t, `{"data":[{"id":"xp-500","name":"Level Up","description":"Kumpulkan 500 XP","progress":120,"target":500}]}`)
	if !translateAccountPayload("/v1/me/achievements", payload) {
		t.Fatal("expected a change")
	}
	item := payload.(map[string]any)["data"].([]any)[0].(map[string]any)
	if number, ok := item["progress"].(json.Number); !ok || number.String() != "120" {
		t.Fatalf("progress = %#v, want json.Number 120", item["progress"])
	}
	if item["description"] != "Earn 500 XP." {
		t.Fatalf("description = %#v", item["description"])
	}
}

func TestTranslateAccountPayloadNeverPanicsOnUnexpectedShapes(t *testing.T) {
	newPayloads := func() []any {
		var deep any = "leaf"
		for i := 0; i < 100; i++ {
			deep = []any{deep}
		}
		return []any{
			nil, "text", json.Number("5"), true, []any{}, map[string]any{}, deep,
			[]any{nil, 3, "x", []any{map[string]any{"id": 5, "name": 7, "description": nil}}},
			map[string]any{
				"features": "x", "intervals": []any{1}, "categories": "x", "id": nil, "kind": 5, "title": 9,
				"personal": "yes", "sourceName": 3, "mode": 4, "name": 5, "tagline": 1, "notice": nil,
				"items": []any{nil}, "scope": 1, "reason": 2, "instructions": 3, "description": []any{},
				"sectors":    []any{nil, map[string]any{"id": 1}, map[string]any{"id": "railway", "label": 3}},
				"intervals2": map[string]any{"again": 1},
			},
			map[string]any{"intervals": map[string]any{"again": 1, "hard": nil, "good": []any{}}, "features": []any{1, nil, map[string]any{}}},
			map[string]any{"id": "first-lesson", "name": 1, "description": 2},
			map[string]any{"kind": "grammar", "title": 5, "lessonId": 7},
			map[string]any{"kind": "kanji", "title": "x", "lessonId": "kanji"},
			map[string]any{"kind": "kanji", "title": "x", "lessonId": "kanji-n-"},
		}
	}
	paths := []string{
		"", "/", "/v1", "/v1/me", "/v1/me/gamification", "/v1/me/achievements", "/v1/me/learning-history",
		"/v1/me/learning-history/kanji/kanji-n3-01", "/v1/plans", "/v1/pricing", "/v1/billing", "/v1/jobs",
		"/v1/jobs/meta", "/v1/jobs/sources", "/v1/anki/decks", "/v1/anki/credits", "/v1/branding",
		"/v1/auth/login", "/v1/manga", "/v1/app-update-policy",
	}
	for _, path := range paths {
		for index, payload := range newPayloads() {
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Fatalf("path %q payload %d panicked: %v", path, index, recovered)
					}
				}()
				if translateAccountPayload(path, payload) {
					t.Fatalf("path %q payload %d: unexpected change", path, index)
				}
			}()
		}
	}
}

func TestEveryStoreAchievementHasEnglishText(t *testing.T) {
	items := store.NewMemoryStore().Achievements("account-localization-test")
	if len(items) == 0 {
		t.Fatal("store returned no achievements")
	}
	seen := map[string]bool{}
	for _, item := range items {
		seen[item.ID] = true
		if _, ok := accountAchievementEnglish[item.ID]; !ok {
			t.Errorf("achievement %q has no English translation", item.ID)
		}
	}
	for id := range accountAchievementEnglish {
		if !seen[id] {
			t.Errorf("translation table has stale achievement id %q", id)
		}
	}

	payload := encodeAccountValue(t, map[string]any{"data": items})
	if !translateAccountPayload("/v1/me/achievements", payload) {
		t.Fatal("expected a change")
	}
	translated := payload.(map[string]any)["data"].([]any)
	if len(translated) != len(items) {
		t.Fatalf("translated %d achievements, want %d", len(translated), len(items))
	}
	for index, raw := range translated {
		item := raw.(map[string]any)
		want := accountAchievementEnglish[items[index].ID]
		if item["name"] != want.Name || item["description"] != want.Description {
			t.Errorf("achievement %q = %q / %q, want %q / %q", items[index].ID, item["name"], item["description"], want.Name, want.Description)
		}
		if item["description"] == items[index].Description {
			t.Errorf("achievement %q description was not translated", items[index].ID)
		}
		if item["id"] != items[index].ID || item["icon"] != items[index].Icon {
			t.Errorf("achievement %q lost its id or icon", items[index].ID)
		}
	}
}

func TestEverySeededPlanFeatureHasEnglishText(t *testing.T) {
	plans := store.NewMemoryStore().Plans()
	if len(plans) == 0 {
		t.Fatal("store returned no plans")
	}
	for _, plan := range plans {
		for _, feature := range plan.Features {
			if _, ok := accountPlanFeatureEnglish[feature]; !ok {
				t.Errorf("plan %q feature %q has no English translation", plan.ID, feature)
			}
		}
	}
	payload := encodeAccountValue(t, map[string]any{"data": plans})
	translateAccountPayload("/v1/plans", payload)
	for index, raw := range payload.(map[string]any)["data"].([]any) {
		plan := raw.(map[string]any)
		if plan["id"] != plans[index].ID || plan["name"] != plans[index].Name || plan["billingType"] != plans[index].BillingType {
			t.Errorf("plan %q identity fields changed", plans[index].ID)
		}
		for featureIndex, feature := range plan["features"].([]any) {
			if feature != accountPlanFeatureEnglish[plans[index].Features[featureIndex]] {
				t.Errorf("plan %q feature %d = %q", plans[index].ID, featureIndex, feature)
			}
		}
	}
}

func TestEveryJobMetadataOptionHasEnglishLabel(t *testing.T) {
	metadata := encodeAccountValue(t, map[string]any{"data": jobs.Metadata()})
	before := marshalAccountPayload(t, metadata.(map[string]any)["data"].(map[string]any)["prefectures"])
	for group, options := range metadata.(map[string]any)["data"].(map[string]any) {
		if group == "prefectures" {
			continue
		}
		labels, ok := accountJobOptionEnglish[group]
		if !ok {
			t.Errorf("metadata group %q has no translation table", group)
			continue
		}
		for _, raw := range options.([]any) {
			id := raw.(map[string]any)["id"].(string)
			if _, ok := labels[id]; !ok {
				t.Errorf("metadata %s/%s has no English label", group, id)
			}
		}
		if len(labels) != len(options.([]any)) {
			t.Errorf("group %q: translation table has %d entries for %d options", group, len(labels), len(options.([]any)))
		}
	}
	translateAccountPayload("/v1/jobs/meta", metadata)
	data := metadata.(map[string]any)["data"].(map[string]any)
	for group, options := range data {
		if group == "prefectures" {
			continue
		}
		for _, raw := range options.([]any) {
			option := raw.(map[string]any)
			if option["label"] != accountJobOptionEnglish[group][option["id"].(string)] {
				t.Errorf("%s/%v label = %q", group, option["id"], option["label"])
			}
		}
	}
	if after := marshalAccountPayload(t, data["prefectures"]); after != before {
		t.Errorf("prefectures changed: %s", after)
	}
}

func TestEveryJobSourceTextHasEnglish(t *testing.T) {
	for _, source := range jobfetch.Sources() {
		if !strings.Contains(source.Name, accountSourceNameSeparator) {
			continue
		}
		translated, ok := translateJobSourceName(source.Name)
		if !ok {
			t.Errorf("source name %q has no English suffix", source.Name)
		}
		if translated == source.Name {
			t.Errorf("source name %q was not translated", source.Name)
		}
	}
	raw, err := os.ReadFile("../jobs/resources/sources.json")
	if err != nil {
		t.Fatal(err)
	}
	directory, err := jobs.DecodeDirectory(raw)
	if err != nil || len(directory) == 0 {
		t.Fatalf("decode directory: %v (%d entries)", err, len(directory))
	}
	for _, source := range directory {
		if source.Scope != "" {
			if _, ok := accountJobSourceScopeEnglish[source.Scope]; !ok {
				t.Errorf("source %q scope %q has no English text", source.ID, source.Scope)
			}
		}
		if source.Reason != "" {
			if _, ok := accountJobSourceReasonEnglish[source.Reason]; !ok {
				t.Errorf("source %q reason %q has no English text", source.ID, source.Reason)
			}
		}
	}
}

func TestEveryAnkiCatalogTextHasEnglish(t *testing.T) {
	decks, err := anki.List("", "")
	if err != nil || len(decks) == 0 {
		t.Fatalf("anki catalog: %v (%d decks)", err, len(decks))
	}
	// Proper names that are identical in English and absent from the client
	// catalog on purpose.
	untranslatedTitles := map[string]bool{
		"Hentaigana": true, "Hiragana": true, "Katakana": true, "Kokuji": true,
		"Kodansha Kanji Learner’s Course": true,
	}
	for _, deck := range decks {
		if _, ok := accountAnkiCatalogEnglish[deck.Title]; !ok && !untranslatedTitles[deck.Title] {
			t.Errorf("deck %q title %q has no English text", deck.ID, deck.Title)
		}
		for field, value := range map[string]string{
			"description":       deck.Description,
			"instructions":      deck.Instructions,
			"unavailableReason": deck.UnavailableReason,
			"meaningLanguage":   deck.MeaningLanguage,
		} {
			if value == "" {
				continue
			}
			if _, ok := accountAnkiCatalogEnglish[value]; !ok {
				t.Errorf("deck %q %s %q has no English text", deck.ID, field, value)
			}
		}
	}

	// A real deck keeps its category and original name through translation.
	deck := decks[0]
	payload := encodeAccountValue(t, map[string]any{"data": deck})
	translateAccountPayload("/v1/anki/decks/"+deck.ID, payload)
	got := payload.(map[string]any)["data"].(map[string]any)
	if got["category"] != deck.Category || got["originalName"] != deck.OriginalName || got["id"] != deck.ID {
		t.Errorf("deck identity fields changed: %v", got)
	}
	if want := accountAnkiCatalogEnglish[deck.Description]; want != "" && got["description"] != want {
		t.Errorf("description = %q, want %q", got["description"], want)
	}
}

func TestTranslateAnkiInterval(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		changed bool
	}{
		{"1 menit", "1 min", true},
		{"15 menit", "15 min", true},
		{"±3 jam", "±3 hr", true},
		{"1 hari", "1 day", true},
		{"12 hari", "12 days", true},
		{"hari", "hari", false},
		{"x hari", "x hari", false},
		{"±jam", "±jam", false},
		{"3 hari lagi", "3 hari lagi", false},
		{"", "", false},
		{"10 days", "10 days", false},
	}
	for _, tt := range tests {
		got, changed := translateAnkiInterval(tt.input)
		if got != tt.want || changed != tt.changed {
			t.Errorf("translateAnkiInterval(%q) = %q, %v; want %q, %v", tt.input, got, changed, tt.want, tt.changed)
		}
	}
}
