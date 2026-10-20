package jobfetch

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExplicitJobBenefits(t *testing.T) {
	got := postingBenefits([]any{"<p>Paid leave</p>", "<b>Commuting allowance</b>", 42})
	if !strings.Contains(got, "Paid leave") || !strings.Contains(got, "Commuting allowance") || strings.Contains(got, "<") {
		t.Fatal(got)
	}
	if postingBenefits(nil) != "" {
		t.Fatal("missing benefits invented")
	}
	var p leverPosting
	if err := json.Unmarshal([]byte(`{"id":"one","text":"Engineer","country":"JP","hostedUrl":"https://jobs.lever.co/example/one","applyUrl":"https://jobs.lever.co/example/one/apply","lists":[{"text":"Benefits","content":"<ul><li>Paid leave</li><li>Transport allowance</li></ul>"},{"text":"Requirements","content":"<p>Three years experience</p>"}]}`), &p); err != nil {
		t.Fatal(err)
	}
	job, ok, _ := normalizeLever(Source{ID: "lever-example", Company: "Example"}, p)
	if !ok || !strings.Contains(job.Benefits, "Paid leave") || strings.Contains(job.Benefits, "experience") {
		t.Fatal("benefits mixed with requirements", job.Benefits)
	}
}
