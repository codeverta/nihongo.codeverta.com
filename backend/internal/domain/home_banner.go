package domain

// HomeBanner is managed by the backend and links to a supported app destination.
type HomeBanner struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	Action      string `json:"action"`
	ButtonLabel string `json:"buttonLabel"`
	Order       int    `json:"order"`
	Enabled     bool   `json:"enabled"`
}
