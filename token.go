package firebase

import (
	"github.com/SermoDigital/jose/jwt"
)

type (
	Token struct {
		jwt.JWT
	}
)

// UID returns the uid for this token.
func (t *Token) UID() (string, bool) {
	return t.Claims().Subject()
}

// Issuer returns the issuer for this token.
func (t *Token) Issuer() (string, bool) {
	return t.Claims().Issuer()
}

// Name returns the user's display name.
func (t *Token) Name() (string, bool) {
	name, ok := t.Claims().Get("name").(string)
	return name, ok
}

// Picture returns the URI string of the user's profile photo.
func (t *Token) Picture() (string, bool) {
	picture, ok := t.Claims().Get("picture").(string)
	return picture, ok
}

// Email returns the email address for this user, or nil if it's unavailable.
func (t *Token) Email() (string, bool) {
	email, ok := t.Claims().Get("email").(string)
	return email, ok
}

// IsEmailVerified indicates if the email address returned by Email() has been
// verified as good.
func (t *Token) IsEmailVerified() (bool, bool) {
	emailVerified, ok := t.Claims().Get("email_verified").(bool)
	return emailVerified, ok
}
