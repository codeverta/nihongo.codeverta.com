package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFCMPayloadAndDryRun(t *testing.T) {
	for _, dry := range []bool{true, false} {
		t.Run(map[bool]string{true: "dry", false: "send"}[dry], func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var p struct {
					Validate bool `json:"validate_only"`
					Message  struct {
						Topic   string            `json:"topic"`
						Token   string            `json:"token"`
						Data    map[string]string `json:"data"`
						Android struct {
							TTL          string            `json:"ttl"`
							Notification map[string]string `json:"notification"`
						} `json:"android"`
					} `json:"message"`
				}
				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					t.Fatal(err)
				}
				if p.Validate != dry || p.Message.Topic != Topic || p.Message.Token != "" || p.Message.Data["route"] != "/home" || p.Message.Android.TTL != "86400s" || p.Message.Android.Notification["channel_id"] != Channel {
					t.Errorf("unexpected payload %+v", p)
				}
				w.Write([]byte(`{"name":"projects/test/messages/id"}`))
			}))
			defer server.Close()
			f := &FCM{client: server.Client(), endpoint: server.URL}
			name, err := f.Send(context.Background(), Message{Title: "Title", Body: "Body", Route: "/home"}, "id", dry)
			if err != nil || name == "" {
				t.Fatalf("send: %s %v", name, err)
			}
		})
	}
}
func TestFCMErrorsAreSafeAndUncertaintyIsPreserved(t *testing.T) {
	for _, code := range []int{403, 429, 500} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
				w.Write([]byte("private server response"))
			}))
			defer s.Close()
			f := &FCM{client: s.Client(), endpoint: s.URL}
			_, err := f.Send(context.Background(), Message{}, "id", false)
			if err == nil || strings.Contains(err.Error(), "private") || IsUncertain(err) != (code >= 500) {
				t.Fatalf("bad error: %v", err)
			}
		})
	}
}

func TestTopicPayloadCountsUTF8AndJSONEscapes(t *testing.T) {
	for _, body := range []string{strings.Repeat("🎉", 500), strings.Repeat("<", 500)} {
		if err := ValidateSize(Message{Title: "Berita", Body: body, Route: "/home"}); err == nil {
			t.Fatal("oversized topic payload accepted")
		}
	}
	if err := ValidateSize(Message{Title: "Berita aplikasi", Body: strings.Repeat("a", 500), Route: "/home"}); err != nil {
		t.Fatal(err)
	}
}
