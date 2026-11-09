// Package notifications sends public announcements only. It never includes account data.
package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const Topic = "nihongo_announcements"
const Channel = "nihongo_announcements"

// TopicEnglish receives the English version. Indonesian keeps the original
// topic so installs that predate language-specific topics still work.
const TopicEnglish = "nihongo_announcements_en"

const LocaleEnglish = "en"

// Message holds the Indonesian text plus an optional English version. Locale
// selects which topic a single send goes to and is never stored or exposed.
type Message struct {
	Title   string `json:"title"`
	Body    string `json:"body"`
	TitleEn string `json:"titleEn,omitempty"`
	BodyEn  string `json:"bodyEn,omitempty"`
	Route   string `json:"route"`
	Locale  string `json:"-"`
}

// HasEnglish reports whether an English translation was written.
func (m Message) HasEnglish() bool { return m.TitleEn != "" && m.BodyEn != "" }

// English returns the message to send to English readers. Without a written
// translation it falls back to the Indonesian text so nobody is left out.
func (m Message) English() Message {
	out := Message{Title: m.Title, Body: m.Body, Route: m.Route, Locale: LocaleEnglish}
	if m.HasEnglish() {
		out.Title, out.Body = m.TitleEn, m.BodyEn
	}
	return out
}

type Sender interface {
	Send(context.Context, Message, string, bool) (string, error)
}

// SendError distinguishes a definite rejection from a request whose delivery is unknown.
type SendError struct {
	Message   string
	Uncertain bool
}

func (e *SendError) Error() string { return e.Message }
func IsUncertain(err error) bool   { var e *SendError; return errors.As(err, &e) && e.Uncertain }

type FCM struct {
	project  string
	client   *http.Client
	endpoint string
}

func NewFCM(ctx context.Context, project, credentialPath string) (*FCM, error) {
	if project == "" || credentialPath == "" {
		return nil, nil
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9-]{4,61}[a-z0-9]$`).MatchString(project) {
		return nil, fmt.Errorf("invalid FCM_PROJECT_ID")
	}
	raw, err := os.ReadFile(credentialPath)
	if err != nil {
		return nil, fmt.Errorf("read FCM credentials: %w", err)
	}
	var metadata struct {
		Type    string `json:"type"`
		Project string `json:"project_id"`
	}
	if json.Unmarshal(raw, &metadata) != nil || metadata.Type != "service_account" || metadata.Project != project {
		return nil, fmt.Errorf("FCM credentials must be a service account for FCM_PROJECT_ID")
	}
	// Bound OAuth credential refresh too; it uses its own HTTP request context.
	authContext := context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: 10 * time.Second})
	credentials, err := google.CredentialsFromJSON(authContext, raw, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return nil, fmt.Errorf("parse FCM credentials")
	}
	client := oauth2.NewClient(authContext, credentials.TokenSource)
	client.Timeout = 15 * time.Second
	return &FCM{project: project, client: client, endpoint: "https://fcm.googleapis.com/v1/projects/" + project + "/messages:send"}, nil
}
func marshalPayload(m Message, id string, dryRun bool) []byte {
	topic := Topic
	if m.Locale == LocaleEnglish {
		topic = TopicEnglish
	}
	payload := map[string]any{"validate_only": dryRun, "message": map[string]any{
		"topic": topic, "notification": map[string]string{"title": m.Title, "body": m.Body},
		"data":    map[string]string{"route": m.Route, "announcementId": id},
		"android": map[string]any{"ttl": "86400s", "priority": "NORMAL", "notification": map[string]string{"channel_id": Channel, "icon": "ic_notification", "tag": id}},
	}}
	raw, _ := json.Marshal(payload)
	return raw
}

// Topic messages have a smaller payload budget; count JSON escapes and UTF-8 too.
func ValidateSize(m Message) error {
	const sample = "00000000-0000-0000-0000-000000000000"
	if len(marshalPayload(m, sample, false)) > 2048 || len(marshalPayload(m.English(), sample, false)) > 2048 {
		return fmt.Errorf("pesan terlalu besar untuk notifikasi; kurangi isi atau emoji")
	}
	return nil
}
func (f *FCM) Send(ctx context.Context, m Message, id string, dryRun bool) (string, error) {
	if err := ValidateSize(m); err != nil {
		return "", &SendError{Message: err.Error()}
	}
	raw := marshalPayload(m, id, dryRun)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.endpoint, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := f.client.Do(req)
	if err != nil {
		return "", &SendError{Message: "Koneksi Firebase terputus. Status pengiriman belum dapat dipastikan.", Uncertain: true}
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 64<<10))
	if err != nil {
		return "", &SendError{Message: "Jawaban Firebase tidak lengkap. Periksa riwayat sebelum mengirim ulang.", Uncertain: true}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message := "Firebase menolak pesan. Periksa konfigurasi pengirim."
		if response.StatusCode == 401 || response.StatusCode == 403 {
			message = "Firebase menolak akses. Periksa API Cloud Messaging dan izin service account pengirim."
		}
		if response.StatusCode == 429 {
			message = "Batas pengiriman Firebase tercapai. Coba kembali nanti."
		}
		return "", &SendError{Message: message, Uncertain: response.StatusCode >= 500}
	}
	var result struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(data, &result) != nil || result.Name == "" {
		return "", &SendError{Message: "Firebase menerima permintaan tetapi ID pesan tidak terbaca.", Uncertain: true}
	}
	return result.Name, nil
}
