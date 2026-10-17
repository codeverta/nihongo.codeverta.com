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

type Message struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Route string `json:"route"`
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
	credentials, err := google.CredentialsFromJSON(ctx, raw, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return nil, fmt.Errorf("parse FCM credentials")
	}
	client := oauth2.NewClient(ctx, credentials.TokenSource)
	client.Timeout = 15 * time.Second
	return &FCM{project: project, client: client, endpoint: "https://fcm.googleapis.com/v1/projects/" + project + "/messages:send"}, nil
}
func marshalPayload(m Message, id string, dryRun bool) []byte {
	payload := map[string]any{"validate_only": dryRun, "message": map[string]any{
		"topic": Topic, "notification": map[string]string{"title": m.Title, "body": m.Body},
		"data":    map[string]string{"route": m.Route, "announcementId": id},
		"android": map[string]any{"ttl": "86400s", "priority": "NORMAL", "notification": map[string]string{"channel_id": Channel, "icon": "ic_notification", "tag": id}},
	}}
	raw, _ := json.Marshal(payload)
	return raw
}

// Topic messages have a smaller payload budget; count JSON escapes and UTF-8 too.
func ValidateSize(m Message) error {
	if len(marshalPayload(m, "00000000-0000-0000-0000-000000000000", false)) > 2048 {
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
