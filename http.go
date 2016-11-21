package firebase

import (
	"fmt"

	"net/http"
)

const bearer = "Bearer"

func AuthorizationFromParam(req *http.Request) (string, error) {
	return req.URL.Query().Get("authorization"), nil
}

func AuthorizationFromHeader(req *http.Request) (string, error) {
	header := req.Header.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("Authorization header not found")
	}

	l := len(bearer)
	if len(header) > l+1 && header[:l] == bearer {
		return header[l+1:], nil
	}

	return "", fmt.Errorf("Authorization header format must be 'Bearer {token}'")
}

func AuthorizationFromRequest(req *http.Request) (string, error) {
	authorization, err := AuthorizationFromParam(req)
	if authorization == "" {
		authorization, err = AuthorizationFromHeader(req)
		if err != nil {
			return "", err
		}
	}
	return authorization, nil
}

// TODO: add some convenient middleware handlers to extract claims and provide
// "is signed in?" or "does have role?" checks (with callback for custom checks)
