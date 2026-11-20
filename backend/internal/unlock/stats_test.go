package unlock_test

import (
	"context"
	"testing"
	"time"

	"nihongo-master-backend/internal/unlock"
)

func scenarioStatsCountActivityByDayAndGroup(t *testing.T, h *harness) {
	ctx := context.Background()
	today := h.now

	// Three days ago: outside a two-day window.
	h.now = today.AddDate(0, 0, -3)
	if _, err := h.service.Fallback(ctx, userA, "grammar:old"); err != nil {
		t.Fatal(err)
	}

	// Yesterday: one verified ad that was claimed.
	h.now = today.AddDate(0, 0, -1)
	intent := h.verifiedIntents(t, userA, []string{"kanji:k1"})[0]
	if _, err := h.service.Claim(ctx, userA, intent.Nonce); err != nil {
		t.Fatal(err)
	}

	// Today: an ad shown but never confirmed, a fallback, and a spent simulation credit.
	h.now = today
	if _, err := h.service.CreateIntent(ctx, userB, "news:a1"); err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.Fallback(ctx, userB, "anki:deck-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := h.service.Fallback(ctx, userA, "simulation:jlpt-n3-simulation-1"); err != nil {
		t.Fatal(err)
	}
	if ok, err := h.service.Convert(ctx, userA, "simulation:jlpt-n3-simulation-1", "simulation-attempt:attempt-1"); err != nil || !ok {
		t.Fatalf("convert: %v %v", ok, err)
	}

	report, err := h.service.Stats(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if report.Since != unlock.Day(today.AddDate(0, 0, -1)) {
		t.Fatalf("since = %s", report.Since)
	}
	if len(report.Days) != 2*len(unlock.Groups) || report.Days[0].Day != unlock.Day(today) {
		t.Fatalf("days = %+v", report.Days)
	}
	total := map[unlock.Group]unlock.DayStats{}
	for _, row := range report.Totals {
		total[row.Group] = row
	}
	want := map[unlock.Group]unlock.DayStats{
		unlock.GroupLessons:    {Group: unlock.GroupLessons, Intents: 1, Verified: 1, Rewarded: 1},
		unlock.GroupNews:       {Group: unlock.GroupNews, Intents: 1},
		unlock.GroupAnki:       {Group: unlock.GroupAnki, Fallback: 1},
		unlock.GroupSimulation: {Group: unlock.GroupSimulation, Fallback: 1},
		unlock.GroupJobs:       {Group: unlock.GroupJobs},
	}
	for group, w := range want {
		if total[group] != w {
			t.Errorf("%s totals = %+v, want %+v", group, total[group], w)
		}
	}
	if len(report.Accounts) != 2 {
		t.Fatalf("accounts = %v, want both accounts that opened something in the window", report.Accounts)
	}

	// The window is clamped.
	if wide, err := h.service.Stats(ctx, 400); err != nil || len(wide.Days) != unlock.MaxStatsDays*len(unlock.Groups) {
		t.Fatalf("clamped window: %d rows, %v", len(wide.Days), err)
	}
}

func TestDaysFollowJakartaTime(t *testing.T) {
	// 17:30 UTC is already the next day in Jakarta (UTC+7).
	if got := unlock.Day(time.Date(2026, 11, 19, 17, 30, 0, 0, time.UTC)); got != "2026-11-20" {
		t.Fatalf("Day = %s", got)
	}
}
