package store

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"math"
	"nihongo-master-backend/internal/anki"
	"nihongo-master-backend/internal/ankitransfer"
	"nihongo-master-backend/internal/domain"
	"sort"
	"strconv"
	"time"
)

var ErrAnkiConflict = errors.New("jadwal kartu berubah; muat ulang sesi belajar")

type AnkiCardState struct {
	Version     int       `json:"version"`
	Due         time.Time `json:"due"`
	Interval    float64   `json:"interval"`
	Ease        int       `json:"ease"`
	Repetitions int       `json:"repetitions"`
	Lapses      int       `json:"lapses"`
	RequestID   string    `json:"requestId"`
	Rating      string    `json:"rating"`
}
type AnkiDeckState struct {
	Revision string                `json:"revision"`
	NextNew  int                   `json:"nextNew"`
	Cards    map[int]AnkiCardState `json:"cards"`
}
type AnkiStats struct {
	Seen    int        `json:"seen"`
	Due     int        `json:"due"`
	New     int        `json:"new"`
	NextDue *time.Time `json:"nextDue"`
}
type AnkiDeckView struct {
	anki.Deck
	Stats        AnkiStats `json:"stats"`
	Locked       bool      `json:"locked"`
	Personal     bool      `json:"personal"`
	RequiredTier string    `json:"requiredTier"`
	SourceID     string    `json:"sourceId,omitempty"`
}
type AnkiStudyCard struct {
	anki.Card
	Version   int               `json:"version"`
	Intervals map[string]string `json:"intervals"`
	AudioURL  string            `json:"audioUrl,omitempty"`
}
type AnkiSession struct {
	Title        string          `json:"title"`
	Level        string          `json:"level"`
	RequiredTier string          `json:"requiredTier"`
	Instructions string          `json:"instructions"`
	DeckID       string          `json:"deckId"`
	Cards        []AnkiStudyCard `json:"cards"`
	Stats        AnkiStats       `json:"stats"`
}
type AnkiReviewInput struct {
	Index     int    `json:"index"`
	Version   int    `json:"version"`
	Rating    string `json:"rating"`
	RequestID string `json:"requestId"`
}
type AnkiReviewResult struct {
	Saved   bool      `json:"saved"`
	Due     time.Time `json:"due"`
	Version int       `json:"version"`
	// XPAwarded is what this review earned (1 until the daily Anki cap), so
	// the session end can show "+N XP".
	XPAwarded int `json:"xpAwarded"`
}

func ankiState(state *gamificationState, id string) *AnkiDeckState {
	return ankiStateForRevision(state, id, anki.Revision())
}
func ankiStateForRevision(state *gamificationState, id, revision string) *AnkiDeckState {
	if state != nil && state.Anki != nil && state.Anki[id] != nil && state.Anki[id].Revision == revision {
		return state.Anki[id]
	}
	return &AnkiDeckState{Revision: revision, Cards: map[int]AnkiCardState{}}
}
func ankiStats(state *AnkiDeckState, total int, now time.Time) AnkiStats {
	stats := AnkiStats{New: total}
	for index, card := range state.Cards {
		if index < 0 || index >= total {
			continue
		}
		stats.Seen++
		stats.New--
		if !card.Due.After(now) {
			stats.Due++
		} else if stats.NextDue == nil || card.Due.Before(*stats.NextDue) {
			due := card.Due
			stats.NextDue = &due
		}
	}
	return stats
}
func ankiPremium(deck anki.Deck) bool {
	if openLearningAccess {
		return false
	}
	return deck.Level == "N1" || deck.Level == "N2" || deck.Level == "N3"
}
func (s *MemoryStore) AnkiDeck(user, id string) (AnkiDeckView, error) {
	return s.AnkiDeckFor(user, id, false)
}

// AnkiDeckFor is AnkiDeck for an account that may have opened this built-in
// deck by watching an ad (unlocked). It never affects imported decks.
func (s *MemoryStore) AnkiDeckFor(user, id string, unlocked bool) (AnkiDeckView, error) {
	if isUserAnkiDeckID(id) {
		return s.userAnkiDeck(user, id)
	}
	deck, err := anki.Get(id)
	if errors.Is(err, anki.ErrNotFound) {
		return AnkiDeckView{}, ErrNotFound
	}
	if err != nil {
		return AnkiDeckView{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	locked := ankiPremium(deck) && s.profiles[user].Entitlement.Tier != domain.TierPremium && !unlocked
	requiredTier := "free"
	if ankiPremium(deck) {
		requiredTier = "lifetime"
	}
	if locked {
		deck.Preview = nil
	}
	return AnkiDeckView{Deck: deck, Stats: ankiStats(ankiState(s.gamification[user], id), deck.CardCount, s.gamificationNow()), Locked: locked, RequiredTier: requiredTier}, nil
}
func (s *MemoryStore) AnkiStudy(user, id string, limit int) (AnkiSession, error) {
	return s.AnkiStudyFor(user, id, limit, false)
}

// AnkiStudyFor is AnkiStudy for an account that may have opened this built-in
// deck by watching an ad.
func (s *MemoryStore) AnkiStudyFor(user, id string, limit int, unlocked bool) (AnkiSession, error) {
	view, err := s.AnkiDeckFor(user, id, unlocked)
	if err != nil {
		return AnkiSession{}, err
	}
	if view.Locked {
		return AnkiSession{}, ErrPremium
	}
	if !view.Available {
		return AnkiSession{}, ErrInvalid
	}
	if limit < 1 || limit > 20 {
		return AnkiSession{}, ErrInvalid
	}
	s.mu.RLock()
	stateRevision := anki.Revision()
	var personalCards []ankitransfer.Card
	if isUserAnkiDeckID(id) {
		record := s.userAnkiDecks[id]
		if record == nil || record.UserID != user {
			s.mu.RUnlock()
			return AnkiSession{}, ErrNotFound
		}
		stateRevision = record.Revision
		personalCards = copyAnkiTransferCards(record.Cards)
	}
	state := ankiStateForRevision(s.gamification[user], id, stateRevision)
	now := s.gamificationNow().UTC()
	type item struct {
		index int
		state AnkiCardState
	}
	due := []item{}
	for index, card := range state.Cards {
		if index >= 0 && index < view.CardCount && !card.Due.After(now) {
			due = append(due, item{index, card})
		}
	}
	sort.Slice(due, func(i, j int) bool {
		if due[i].state.Due.Equal(due[j].state.Due) {
			return due[i].index < due[j].index
		}
		return due[i].state.Due.Before(due[j].state.Due)
	})
	selected := []item{}
	for _, card := range due {
		if len(selected) >= limit {
			break
		}
		selected = append(selected, card)
	}
	for index := state.NextNew; index < view.CardCount && len(selected) < limit; index++ {
		if _, seen := state.Cards[index]; !seen {
			selected = append(selected, item{index, AnkiCardState{}})
		}
	}
	stats := ankiStats(state, view.CardCount, now)
	s.mu.RUnlock()
	requiredTier := view.RequiredTier
	if requiredTier == "" {
		requiredTier = "free"
	}
	result := AnkiSession{Title: view.Title, Level: view.Level, RequiredTier: requiredTier, Instructions: view.Instructions, DeckID: id, Stats: stats, Cards: []AnkiStudyCard{}}
	for _, entry := range selected {
		var card anki.Card
		audioURL := ""
		if isUserAnkiDeckID(id) {
			transferCard := personalCards[entry.index]
			card = anki.Card{Index: entry.index, SourceIndex: entry.index, Front: transferCard.Front, Answers: append([]string(nil), transferCard.Answers...), Meaning: transferCard.Meaning, ImageURL: ankiMediaURL(id, transferCard.ImageURL)}
			audioURL = ankiMediaURL(id, transferCard.Audio)
		} else {
			var err error
			card, err = anki.ReadCard(id, entry.index)
			if err != nil {
				return AnkiSession{}, err
			}
			card.ImageURL = ankiMediaURL(id, card.ImageURL)
		}
		intervals := map[string]string{}
		for _, rating := range []string{"again", "hard", "good", "easy"} {
			next := ankiSchedule(entry.state, rating, now)
			duration := next.Due.Sub(now)
			if duration < time.Hour {
				intervals[rating] = "1 menit"
			} else if duration < 24*time.Hour {
				intervals[rating] = "±" + strconv.Itoa(int(math.Round(duration.Hours()))) + " jam"
			} else {
				intervals[rating] = timeDaysLabel(duration)
			}
		}
		result.Cards = append(result.Cards, AnkiStudyCard{Card: card, Version: entry.state.Version, Intervals: intervals, AudioURL: audioURL})
	}
	return result, nil
}
func timeDaysLabel(duration time.Duration) string {
	days := int(math.Round(duration.Hours() / 24))
	if days < 1 {
		days = 1
	}
	return fmtDays(days)
}
func fmtDays(days int) string { // no locale-dependent durations in the study contract
	return strconv.Itoa(days) + " hari"
}
func ankiSchedule(current AnkiCardState, rating string, now time.Time) AnkiCardState {
	next := current
	if next.Ease < 130 {
		next.Ease = 250
	}
	interval := current.Interval
	switch rating {
	case "again":
		next.Repetitions = 0
		next.Lapses++
		next.Ease -= 20
		interval = 1.0 / (24 * 60)
	case "hard":
		next.Repetitions++
		next.Ease -= 15
		interval = math.Max(.25, interval*1.2)
	case "good":
		next.Repetitions++
		if current.Repetitions == 0 {
			interval = 1
		} else {
			interval = math.Max(1, interval*float64(next.Ease)/100)
		}
	case "easy":
		next.Repetitions++
		next.Ease += 15
		interval = math.Max(4, interval*float64(next.Ease)/100*1.3)
	}
	if next.Ease < 130 {
		next.Ease = 130
	}
	if next.Ease > 350 {
		next.Ease = 350
	}
	interval = math.Min(365, interval)
	next.Interval = interval
	next.Due = now.Add(time.Duration(interval * float64(24*time.Hour)))
	next.Version++
	next.Rating = rating
	return next
}
func (s *MemoryStore) AnkiReview(user, id string, input AnkiReviewInput) (AnkiReviewResult, error) {
	return s.AnkiReviewFor(user, id, input, false)
}

// AnkiReviewFor is AnkiReview for an account that may have opened this
// built-in deck by watching an ad.
func (s *MemoryStore) AnkiReviewFor(user, id string, input AnkiReviewInput, unlocked bool) (AnkiReviewResult, error) {
	var deck anki.Deck
	deckRevision := anki.Revision()
	if isUserAnkiDeckID(id) {
		view, err := s.AnkiDeck(user, id)
		if err != nil {
			return AnkiReviewResult{}, ErrNotFound
		}
		deck = view.Deck
		deckRevision = ""
	} else {
		var err error
		deck, err = anki.Get(id)
		if err != nil {
			return AnkiReviewResult{}, ErrNotFound
		}
	}
	requestID, err := uuid.Parse(input.RequestID)
	if err != nil || requestID == uuid.Nil || input.Index < 0 || input.Index >= deck.CardCount || !deck.Available || input.Version < 0 {
		return AnkiReviewResult{}, ErrInvalid
	}
	if input.Rating != "again" && input.Rating != "hard" && input.Rating != "good" && input.Rating != "easy" {
		return AnkiReviewResult{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.profiles[user]; !exists {
		return AnkiReviewResult{}, ErrNotFound
	}
	if isUserAnkiDeckID(id) {
		record := s.userAnkiDecks[id]
		if record == nil || record.UserID != user {
			return AnkiReviewResult{}, ErrNotFound
		}
		deckRevision = record.Revision
		if userAnkiDeckPremium(record) && s.profiles[user].Entitlement.Tier != domain.TierPremium {
			return AnkiReviewResult{}, ErrPremium
		}
	} else if ankiPremium(deck) && s.profiles[user].Entitlement.Tier != domain.TierPremium && !unlocked {
		return AnkiReviewResult{}, ErrPremium
	}
	state := ankiStateForRevision(s.gamification[user], id, deckRevision)
	current := state.Cards[input.Index]
	now := s.gamificationNow().UTC()
	if current.RequestID == input.RequestID {
		if current.Rating != input.Rating {
			return AnkiReviewResult{}, ErrInvalid
		}
		return AnkiReviewResult{Saved: true, Due: current.Due, Version: current.Version}, nil
	}
	if current.Version != input.Version || (!current.Due.IsZero() && current.Due.After(now)) || (current.Version == 0 && input.Index >= state.NextNew+20) {
		return AnkiReviewResult{}, ErrAnkiConflict
	}
	base := s.gamification[user]
	if base == nil {
		base = newGamificationState()
	}
	raw, err := json.Marshal(base)
	if err != nil {
		return AnkiReviewResult{}, err
	}
	next := newGamificationState()
	if err = json.Unmarshal(raw, next); err != nil {
		return AnkiReviewResult{}, err
	}
	if next.Anki == nil {
		next.Anki = map[string]*AnkiDeckState{}
	}
	progress := ankiStateForRevision(next, id, deckRevision)
	next.Anki[id] = progress
	scheduled := ankiSchedule(current, input.Rating, now)
	scheduled.RequestID = input.RequestID
	progress.Cards[input.Index] = scheduled
	for progress.NextNew < deck.CardCount {
		if _, seen := progress.Cards[progress.NextNew]; !seen {
			break
		}
		progress.NextNew++
	}
	// +1 XP per reviewed card, at most 30 XP per study day. Enough reviews in
	// a day also make it count for the streak.
	var awards []xpAward
	studyDay := now.In(studyTimezone(next)).Format("2006-01-02")
	ledger := next.xpDayLocked(studyDay)
	ledger.Reviews++
	if ledger.Reviews >= s.ankiStreakCardsLocked() {
		if _, counted := next.Days[studyDay]; !counted {
			next.Days[studyDay] = 0
		}
	}
	if ledger.Anki < xpAnkiMaxPerDay {
		ledger.Anki += xpAnkiCard
		awards = append(awards, xpAward{xpAnkiCard, "anki"})
		next.XP += xpAnkiCard
		next.pruneXPDays(now)
	}
	// Review-count achievements do not depend on the daily XP cap.
	unlockAchievements(next, now)
	if err = s.persistXPLocked(user, next, awards, now); err != nil {
		return AnkiReviewResult{}, err
	}
	s.recordXPLocked(user, next.XP, totalXP(awards), now)
	s.gamification[user] = next
	if len(awards) > 0 {
		profile := s.profiles[user]
		profile.XP = next.XP
		profile.Level = 1 + next.XP/500
		s.profiles[user] = profile
	}
	return AnkiReviewResult{Saved: true, Due: scheduled.Due, Version: scheduled.Version, XPAwarded: totalXP(awards)}, nil
}
