package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"nihongo-master-backend/internal/notifications"
	"nihongo-master-backend/internal/store"
)

type fakeNotificationSender struct {
	mu       sync.Mutex
	calls    int
	dryCalls int
	err      error
	sent     []notifications.Message
}

func (f *fakeNotificationSender) Send(_ context.Context, m notifications.Message, _ string, dry bool) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if dry {
		f.dryCalls++
	} else {
		f.calls++
		f.sent = append(f.sent, m)
	}
	return "projects/test/messages/123", f.err
}
func notificationHandler(t *testing.T, path string, sender notifications.Sender) http.Handler {
	t.Helper()
	s := NewServer(store.NewMemoryStore(), nil)
	if err := s.ConfigureNotifications(path, "admin-test", sender); err != nil {
		t.Fatal(err)
	}
	return s.Handler()
}
func notificationRequest(h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	var raw []byte
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(raw))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func createTestAnnouncement(t *testing.T, h http.Handler) Announcement {
	t.Helper()
	w := notificationRequest(h, "POST", "/v1/admin/notifications", "admin-test", notifications.Message{Title: "Berita baru", Body: "Yuk belajar kanji!", Route: "/home"})
	if w.Code != 201 {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var out struct{ Data Announcement }
	json.Unmarshal(w.Body.Bytes(), &out)
	return out.Data
}
func TestNotificationAdminAuthValidationAndDrafts(t *testing.T) {
	f := &fakeNotificationSender{}
	h := notificationHandler(t, filepath.Join(t.TempDir(), "history.json"), f)
	for _, c := range []struct{ method, path string }{{"GET", ""}, {"POST", ""}, {"POST", "/check"}, {"PUT", "/test"}, {"POST", "/test/send"}} {
		if w := notificationRequest(h, c.method, "/v1/admin/notifications"+c.path, "wrong", nil); w.Code != 401 {
			t.Fatalf("missing auth %v: %d", c, w.Code)
		}
	}
	for _, m := range []notifications.Message{{Title: "", Body: "test", Route: "/home"}, {Title: "test", Body: "body", Route: "https://evil.example"}, {Title: "test", Body: "body", Route: "/admin/"}} {
		if w := notificationRequest(h, "POST", "/v1/admin/notifications", "admin-test", m); w.Code != 422 {
			t.Fatal("invalid draft accepted")
		}
	}
	a := createTestAnnouncement(t, h)
	if a.Status != "draft" || f.calls != 0 {
		t.Fatal("draft sent before confirmation")
	}
	if w := notificationRequest(h, "POST", "/v1/admin/notifications/"+a.ID+"/send", "admin-test", map[string]string{"confirmId": "wrong"}); w.Code != 422 {
		t.Fatal("send without correct confirmation")
	}
	if f.calls != 0 {
		t.Fatal("invalid confirmation sent notification")
	}
	if w := notificationRequest(h, "POST", "/v1/admin/notifications/check", "admin-test", map[string]any{}); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if f.dryCalls != 1 || f.calls != 0 {
		t.Fatal("connection check broadcast a message")
	}
}
func TestNotificationSendIsDurableAndDeduplicated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	f := &fakeNotificationSender{}
	h := notificationHandler(t, path, f)
	a := createTestAnnouncement(t, h)
	var wg sync.WaitGroup
	codes := make(chan int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codes <- notificationRequest(h, "POST", "/v1/admin/notifications/"+a.ID+"/send", "admin-test", map[string]string{"confirmId": a.ID}).Code
		}()
	}
	wg.Wait()
	close(codes)
	ok, conflict := 0, 0
	for c := range codes {
		if c == 200 {
			ok++
		}
		if c == 409 {
			conflict++
		}
	}
	// One send per language topic (Indonesian + English), never more.
	if f.calls != 2 || ok != 1 || conflict != 1 {
		t.Fatalf("duplicate dispatch: calls %d, status %d/%d", f.calls, ok, conflict)
	}
	h = notificationHandler(t, path, f)
	if w := notificationRequest(h, "POST", "/v1/admin/notifications/"+a.ID+"/send", "admin-test", map[string]string{"confirmId": a.ID}); w.Code != 409 {
		t.Fatal("restart allowed resend")
	}
	w := notificationRequest(h, "GET", "/v1/admin/notifications", "admin-test", nil)
	if !bytes.Contains(w.Body.Bytes(), []byte(`"status":"sent"`)) {
		t.Fatal("sent history not persisted")
	}
}
func TestNotificationUncertainCannotBeRetriedAndFailedCanBeEdited(t *testing.T) {
	for _, uncertain := range []bool{false, true} {
		t.Run(fmt.Sprint(uncertain), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "history.json")
			f := &fakeNotificationSender{err: &notifications.SendError{Message: "failure", Uncertain: uncertain}}
			h := notificationHandler(t, path, f)
			a := createTestAnnouncement(t, h)
			notificationRequest(h, "POST", "/v1/admin/notifications/"+a.ID+"/send", "admin-test", map[string]string{"confirmId": a.ID})
			h = notificationHandler(t, path, f)
			w := notificationRequest(h, "PUT", "/v1/admin/notifications/"+a.ID, "admin-test", a.Message)
			expected := 200
			if uncertain {
				expected = 409
			}
			if w.Code != expected {
				t.Fatalf("editing rejected/uncertain: %d want %d", w.Code, expected)
			}
		})
	}
}
func TestNotificationCrashAndWriteFailureNeverResend(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	f := &fakeNotificationSender{}
	h := notificationHandler(t, path, f)
	a := createTestAnnouncement(t, h)
	a.Status = "sending"
	raw, _ := json.Marshal([]Announcement{a})
	os.WriteFile(path, raw, 0600)
	h = notificationHandler(t, path, f)
	if w := notificationRequest(h, "POST", "/v1/admin/notifications/"+a.ID+"/send", "admin-test", map[string]string{"confirmId": a.ID}); w.Code != 409 {
		t.Fatal("crash allowed resend")
	}
	h = notificationHandler(t, path, f)
	a = createTestAnnouncement(t, h)
	os.Remove(path)
	os.Mkdir(path, 0700)
	if w := notificationRequest(h, "POST", "/v1/admin/notifications/"+a.ID+"/send", "admin-test", map[string]string{"confirmId": a.ID}); w.Code != 500 {
		t.Fatal("write failure accepted")
	}
	if f.calls != 0 {
		t.Fatal("send happened without durable reservation")
	}
}

func TestNotificationSendsEnglishVersionToEnglishTopic(t *testing.T) {
	f := &fakeNotificationSender{}
	h := notificationHandler(t, filepath.Join(t.TempDir(), "history.json"), f)
	w := notificationRequest(h, "POST", "/v1/admin/notifications", "admin-test", notifications.Message{
		Title: "Materi baru", Body: "Yuk belajar kanji!", TitleEn: "New lessons", BodyEn: "Let's study kanji!", Route: "/home",
	})
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	var created struct{ Data Announcement }
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id := created.Data.ID
	if w = notificationRequest(h, "POST", "/v1/admin/notifications/"+id+"/send", "admin-test", map[string]string{"confirmId": id}); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if len(f.sent) != 2 {
		t.Fatalf("sends = %d", len(f.sent))
	}
	if f.sent[0].Locale != "" || f.sent[0].Title != "Materi baru" {
		t.Fatalf("indonesian send = %+v", f.sent[0])
	}
	if f.sent[1].Locale != notifications.LocaleEnglish || f.sent[1].Title != "New lessons" || f.sent[1].Body != "Let's study kanji!" {
		t.Fatalf("english send = %+v", f.sent[1])
	}
}

func TestNotificationEnglishFallsBackToIndonesianAndMustBePaired(t *testing.T) {
	f := &fakeNotificationSender{}
	h := notificationHandler(t, filepath.Join(t.TempDir(), "history.json"), f)
	if w := notificationRequest(h, "POST", "/v1/admin/notifications", "admin-test", notifications.Message{Title: "Judul", Body: "Isi", TitleEn: "Only title", Route: "/home"}); w.Code != 422 {
		t.Fatal("half-filled English version accepted")
	}
	a := createTestAnnouncement(t, h)
	if w := notificationRequest(h, "POST", "/v1/admin/notifications/"+a.ID+"/send", "admin-test", map[string]string{"confirmId": a.ID}); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if len(f.sent) != 2 || f.sent[1].Locale != notifications.LocaleEnglish || f.sent[1].Title != f.sent[0].Title {
		t.Fatalf("fallback sends = %+v", f.sent)
	}
}
