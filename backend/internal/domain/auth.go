package domain

type GoogleIdentity struct {
	Subject     string
	Email       string
	DisplayName string
	PhotoURL    string
}

type AuthSession struct {
	Token string  `json:"token"`
	User  Profile `json:"user"`
}
