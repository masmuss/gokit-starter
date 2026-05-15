package domain

// Session contains the token returned after authentication.
type Session struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int
}
