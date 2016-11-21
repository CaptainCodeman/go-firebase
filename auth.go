package firebase

import (
	"time"

	"github.com/SermoDigital/jose/jwt"
)

const (
	// Audience to use for Firebase Auth Custom tokens
	firebaseAudienceURL = "https://identitytoolkit.googleapis.com/google.identity.identitytoolkit.v1.IdentityToolkit"

	// expiry leeway.
	acceptableExpSkew = 300 * time.Second
)

type (
	claims struct {
		jwt.Claims
	}

	Auth struct {
		app *App
	}
)
