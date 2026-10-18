package jobfetch

import (
	"testing"
)

const accorRow = `{"id":"744000155931619","name":"Front Desk Agent","company":{"identifier":"AccorHotel"},"location":{"city":"Osaka","region":"Osaka","country":"jp","fullLocation":"Osaka, Japan"},"typeOfEmployment":{"label":"Full-time"},"visibility":"PUBLIC","releasedDate":"2026-10-16T07:55:53.083Z"}`

func TestAccorPublicSummary(t *testing.T) {
	result, err := parseAccor([]byte(`{"limit":100,"offset":0,"totalFound":1,"content":[`+accorRow+`]}`), Result{})
	if err != nil || !result.Complete || len(result.Jobs) != 1 {
		t.Fatalf("bad feed: %+v %v", result, err)
	}
	job := result.Jobs[0]
	if job.Category != "hospitality" || job.Prefecture != "Osaka" || job.Country != "JP" || job.Description != "" || job.Qualifications != "" || job.SalaryMin != nil || job.Industry != "" {
		t.Fatalf("invented or missing data: %+v", job)
	}
	if job.SourceURL != "https://jobs.smartrecruiters.com/AccorHotel/744000155931619" || !SafeLink(job.SourceURL) {
		t.Fatal("invalid public source link")
	}
	result, err = parseAccor([]byte(`{"limit":100,"offset":0,"totalFound":150,"content":[`+accorRow+`]}`), Result{})
	if err != nil || result.Complete || len(result.Jobs) != 1 {
		t.Fatal("truncated feed incorrectly authoritative")
	}
}

func TestAccorInvalidEnvelope(t *testing.T) {
	for _, body := range []string{`{}`, `{"limit":100,"offset":100,"totalFound":0}`, `{"limit":100,"offset":0,"totalFound":-1}`, `{"limit":100,"offset":0,"totalFound":0,"content":[` + accorRow + `]}`} {
		if result, err := parseAccor([]byte(body), Result{}); err == nil || result.Complete {
			t.Fatalf("invalid envelope accepted %s", body)
		}
	}
}
