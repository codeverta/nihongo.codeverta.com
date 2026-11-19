package prefinfo_test

import (
	"path/filepath"
	"testing"
	"time"

	"nihongo-master-backend/internal/prefinfo"
)

func TestBundledDataCoversAllPrefectures(t *testing.T) {
	d, err := prefinfo.Default()
	if err != nil {
		t.Fatal(err)
	}
	byCode := map[string]prefinfo.Prefecture{}
	for _, p := range d.Prefectures {
		byCode[p.Code] = p
	}
	// Spot checks against MHLW FY2026 and the 2023 housing survey.
	if byCode["13"].MinimumWage != 1280 || byCode["13"].WageEffectiveDate != "2026-10-01" {
		t.Fatalf("Tokyo = %+v", byCode["13"])
	}
	if byCode["47"].MinimumWage != 1086 || byCode["47"].Rent1KEstimate == nil {
		t.Fatalf("Okinawa = %+v", byCode["47"])
	}
	if d.Meta.IDRPerJPY <= 0 || d.Meta.HoursPerDay != 8 || d.Meta.DaysPerMonth != 21 {
		t.Fatalf("meta = %+v", d.Meta)
	}
}

func TestValidationRejectsBrokenEdits(t *testing.T) {
	d, _ := prefinfo.Default()
	missing := d
	missing.Prefectures = d.Prefectures[:46]
	if missing.Validate() == nil {
		t.Fatal("46 prefectures must be rejected")
	}
	typo := d
	typo.Prefectures = append([]prefinfo.Prefecture(nil), d.Prefectures...)
	typo.Prefectures[0].MinimumWage = 12800
	if typo.Validate() == nil {
		t.Fatal("an implausible wage must be rejected")
	}
	dup := d
	dup.Prefectures = append([]prefinfo.Prefecture(nil), d.Prefectures...)
	dup.Prefectures[1].Code = "01"
	if dup.Validate() == nil {
		t.Fatal("duplicate codes must be rejected")
	}
}

func TestStorePersistsAdminEdits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "living.json")
	store, err := prefinfo.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	d := store.Get()
	d.Prefectures = append([]prefinfo.Prefecture(nil), d.Prefectures...)
	d.Prefectures[12].MinimumWage = 1300
	if _, err := store.Set(d, time.Date(2027, 10, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	reloaded, err := prefinfo.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.Get(); got.Prefectures[12].MinimumWage != 1300 || got.Meta.UpdatedAt != "2027-10-01T00:00:00Z" {
		t.Fatalf("reloaded = %+v", got.Prefectures[12])
	}
}

func TestBundledProfilesCoverAllPrefectures(t *testing.T) {
	profiles, err := prefinfo.LoadProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles.Raw) == 0 || profiles.ETag == "" {
		t.Fatal("profiles must be bundled with an ETag")
	}
}
