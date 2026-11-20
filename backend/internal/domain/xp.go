package domain

// XPGain is what one finished activity earned, returned with its result so
// the app can show "+N XP" and the current streak. CompletionID names the
// activity for a later "double XP" claim.
type XPGain struct {
	Awarded      int    `json:"awarded"`
	Total        int    `json:"total"`
	Streak       int    `json:"streak"`
	CompletionID string `json:"completionId"`
}
