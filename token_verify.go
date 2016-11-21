package firebase

import (
	"errors"
	"fmt"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
	"golang.org/x/net/context"
)

func (a *Auth) VerifyIDToken(ctx context.Context, token string) (*Token, error) {
	decodedJWT, err := jws.ParseJWT([]byte(token))
	if err != nil {
		return nil, err
	}

	decodedJWS, ok := decodedJWT.(jws.JWS)
	if !ok {
		return nil, errors.New("Firebase Auth ID Token cannot be decoded")
	}

	keys := func(j jws.JWS) ([]interface{}, error) {
		kid, ok := j.Protected().Get("kid").(string)
		if !ok {
			return nil, errors.New("Firebase Auth ID Token has no 'kid' claim")
		}
		cert, err := certs.Get(ctx, kid)
		if err != nil {
			return nil, err
		}
		return []interface{}{cert.PublicKey}, nil
	}

	if err := decodedJWS.VerifyCallback(keys,
		[]crypto.SigningMethod{crypto.SigningMethodRS256},
		&jws.SigningOpts{Number: 1, Indices: []int{0}}); err != nil {
		return nil, err
	}

	ks, _ := keys(decodedJWS)
	key := ks[0]
	if err := decodedJWT.Validate(key, crypto.SigningMethodRS256, validator(a.app.creds.ProjectID)); err != nil {
		return nil, err
	}

	return &Token{decodedJWT}, nil
}

func validator(projectID string) *jwt.Validator {
	v := &jwt.Validator{}
	v.EXP = acceptableExpSkew
	v.SetAudience(projectID)
	v.SetIssuer(fmt.Sprintf("https://securetoken.google.com/%s", projectID))
	v.Fn = func(claims jwt.Claims) error {
		subject, ok := claims.Subject()
		if !ok || len(subject) == 0 || len(subject) > 128 {
			return jwt.ErrInvalidSUBClaim
		}
		return nil
	}
	return v
}
