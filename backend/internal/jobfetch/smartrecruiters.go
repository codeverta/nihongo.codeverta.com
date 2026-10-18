package jobfetch

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type accorPosting struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ReleasedDate string `json:"releasedDate"`
	Visibility   string `json:"visibility"`
	Company      struct {
		Identifier string `json:"identifier"`
	} `json:"company"`
	Location struct {
		City         string `json:"city"`
		Region       string `json:"region"`
		Country      string `json:"country"`
		FullLocation string `json:"fullLocation"`
		Remote       bool   `json:"remote"`
		Hybrid       bool   `json:"hybrid"`
	} `json:"location"`
	TypeOfEmployment struct {
		Label string `json:"label"`
	} `json:"typeOfEmployment"`
}

func parseAccor(body []byte, result Result) (Result, error) {
	var feed struct {
		Limit      int            `json:"limit"`
		Offset     int            `json:"offset"`
		TotalFound *int           `json:"totalFound"`
		Content    []accorPosting `json:"content"`
	}
	if err := json.Unmarshal(body, &feed); err != nil {
		return result, err
	}
	if feed.TotalFound == nil || feed.Offset != 0 || feed.Limit != 100 || len(feed.Content) > 100 || *feed.TotalFound < 0 || *feed.TotalFound < len(feed.Content) {
		return result, errors.New("invalid public employer feed envelope")
	}
	result.Jobs = make([]Job, 0, len(feed.Content))
	seen := map[string]bool{}
	for _, p := range feed.Content {
		if !strings.EqualFold(p.Location.Country, "JP") || p.Visibility != "PUBLIC" {
			continue
		}
		if p.ID == "" || len(p.ID) > 30 || strings.Trim(p.ID, "0123456789") != "" || strings.TrimSpace(p.Name) == "" || p.Company.Identifier != "AccorHotel" || seen[p.ID] {
			return result, errors.New("invalid public employer posting")
		}
		seen[p.ID] = true
		link := "https://jobs.smartrecruiters.com/AccorHotel/" + p.ID
		location := p.Location.FullLocation
		if location == "" {
			location = strings.Join([]string{p.Location.City, p.Location.Region, "Japan"}, ", ")
		}
		workplace := "onsite"
		if p.Location.Hybrid {
			workplace = "hybrid"
		} else if p.Location.Remote {
			workplace = "remote"
		}
		job := Job{ExternalID: p.ID, Title: bounded(p.Name, 500), Company: "Accor", Location: bounded(location, 500), Country: "JP", Prefecture: prefecture(p.Location.Region + " " + p.Location.City), Category: "hospitality", EmploymentType: employment(p.TypeOfEmployment.Label), Remote: workplace, Visa: "unknown", JapaneseLevel: "unknown", SourceURL: link, ApplyURL: link}
		if published, err := time.Parse(time.RFC3339Nano, p.ReleasedDate); err == nil {
			utc := published.UTC()
			job.PublishedAt = &utc
		}
		// The public list does not include requirements, salary or description.
		// Leave them absent and link to the original, rather than fan out into
		// one additional request per advertisement.
		result.Jobs = append(result.Jobs, job)
	}
	// If the board grows beyond this one-page budget, ingest the visible page
	// but never retire omitted jobs as though it were an authoritative snapshot.
	result.Complete = *feed.TotalFound == len(feed.Content)
	return result, nil
}
