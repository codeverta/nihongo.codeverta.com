package domain

import "time"

// Purchase is populated exclusively from store verification, never client entitlement flags.
type Purchase struct {
	Acknowledged bool      `json:"acknowledged"`
	Platform     string    `json:"platform"`
	ID           string    `json:"id"`
	ProductID    string    `json:"productId"`
	UserID       string    `json:"userId"`
	AccountToken string    `json:"accountToken"`
	Proof        string    `json:"proof"`
	State        string    `json:"state"`
	CheckedAt    time.Time `json:"checkedAt"`
}

type PurchaseProof struct {
	Platform         string `json:"platform"`
	ProductID        string `json:"productId"`
	VerificationData string `json:"verificationData"`
	TransactionID    string `json:"transactionId"`
}
