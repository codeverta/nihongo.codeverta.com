package unlock

import (
	"context"
	"sort"
	"strings"
	"time"
)

// Group is how the admin dashboard groups kinds: one group per ad switch.
type Group string

const (
	GroupLessons    Group = "lessons"
	GroupAnki       Group = "anki"
	GroupSimulation Group = "simulation"
	GroupNews       Group = "news"
	GroupJobs       Group = "jobs"
	GroupXP         Group = "xp" // double XP after a finished activity
)

// Groups lists every group in display order.
var Groups = []Group{GroupLessons, GroupAnki, GroupSimulation, GroupNews, GroupJobs, GroupXP}

// GroupOf maps a stored content key to its group. A spent simulation credit is
// stored as a simulation attempt and still counts as a simulation.
func GroupOf(contentKey string) (Group, bool) {
	kind, _, found := strings.Cut(contentKey, ":")
	if !found {
		return "", false
	}
	switch k := Kind(kind); {
	case k.IsLesson():
		return GroupLessons, true
	case k == KindAnki:
		return GroupAnki, true
	case k == KindSimulation || k == KindSimulationAttempt:
		return GroupSimulation, true
	case k == KindNews:
		return GroupNews, true
	case k == KindJob:
		return GroupJobs, true
	case k == KindXPBoost:
		return GroupXP, true
	}
	return "", false
}

// Activity is the raw material for statistics: what the repository holds
// since a cutoff. Intents are kept for IntentRetention only.
type Activity struct {
	Intents []IntentActivity
	Unlocks []Unlock
}

// IntentActivity is one intent, without anything that identifies the account.
type IntentActivity struct {
	ContentKey string
	CreatedAt  time.Time
	Verified   bool
	Consumed   bool
}

// DayStats counts one WIB day of activity for one group.
type DayStats struct {
	Day      string `json:"day"`
	Group    Group  `json:"group"`
	Intents  int    `json:"intents"`  // ads the app asked to show
	Verified int    `json:"verified"` // ads Google confirmed by SSV
	Rewarded int    `json:"rewarded"` // items opened with a verified ad
	Fallback int    `json:"fallback"` // items opened without an ad (no ad could load)
}

// Report is the statistics for a window of days.
type Report struct {
	Since  string     `json:"since"`
	Days   []DayStats `json:"days"`   // newest day first, groups in display order
	Totals []DayStats `json:"totals"` // one row per group, Day empty
	// Accounts are the accounts that opened something in the window. The API
	// counts how many of them hold Lifetime Access; they are not shown.
	Accounts []string `json:"-"`
}

// MaxStatsDays bounds the statistics window.
const MaxStatsDays = 30

// Stats summarises the last days days (1 to MaxStatsDays), by WIB day.
func (s *Service) Stats(ctx context.Context, days int) (Report, error) {
	if days < 1 {
		days = 1
	}
	if days > MaxStatsDays {
		days = MaxStatsDays
	}
	now := s.now().In(jakarta)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jakarta).AddDate(0, 0, -(days - 1))
	activity, err := s.Repo.Activity(ctx, start)
	if err != nil {
		return Report{}, err
	}
	return summarise(start, days, activity), nil
}

func summarise(start time.Time, days int, activity Activity) Report {
	type cell struct {
		day   string
		group Group
	}
	counts := map[cell]*DayStats{}
	at := func(t time.Time, group Group) *DayStats {
		key := cell{Day(t), group}
		if counts[key] == nil {
			counts[key] = &DayStats{Day: key.day, Group: group}
		}
		return counts[key]
	}
	for _, intent := range activity.Intents {
		if group, ok := GroupOf(intent.ContentKey); ok {
			row := at(intent.CreatedAt, group)
			row.Intents++
			if intent.Verified {
				row.Verified++
			}
			// Double XP never becomes an unlock: a consumed intent is the reward.
			if group == GroupXP && intent.Consumed {
				row.Rewarded++
			}
		}
	}
	accounts := map[string]bool{}
	for _, u := range activity.Unlocks {
		group, ok := GroupOf(u.ContentKey)
		if !ok {
			continue
		}
		row := at(u.CreatedAt, group)
		if u.Source == SourceFallback {
			row.Fallback++
		} else {
			row.Rewarded++
		}
		accounts[u.UserID] = true
	}

	report := Report{Since: Day(start)}
	totals := map[Group]*DayStats{}
	for _, group := range Groups {
		totals[group] = &DayStats{Group: group}
	}
	for i := days - 1; i >= 0; i-- {
		day := Day(start.AddDate(0, 0, i))
		for _, group := range Groups {
			row := counts[cell{day, group}]
			if row == nil {
				row = &DayStats{Day: day, Group: group}
			}
			report.Days = append(report.Days, *row)
			total := totals[group]
			total.Intents += row.Intents
			total.Verified += row.Verified
			total.Rewarded += row.Rewarded
			total.Fallback += row.Fallback
		}
	}
	for _, group := range Groups {
		report.Totals = append(report.Totals, *totals[group])
	}
	for account := range accounts {
		report.Accounts = append(report.Accounts, account)
	}
	sort.Strings(report.Accounts)
	return report
}
