package firebase

import (
	"fmt"

	"net/http"
)

const bearer = "Bearer"

type AuthFunc func(*Token) (bool, error)

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

func (a *Auth) Authorize(h http.Handler, authFn AuthFunc) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authorization, err := AuthorizationFromRequest(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// check that it's valid
		ctx, err := RequestContext(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := a.VerifyIDToken(ctx, authorization)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ok, err := authFn(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func (a *Auth) Authenticated(h http.Handler, roles ...string) http.Handler {
	return a.Authorize(h, func(token *Token) (bool, error) {
		return true, nil
	})
}

func (a *Auth) AnyRole(h http.Handler, roles ...string) http.Handler {
	return a.Authorize(h, func(token *Token) (bool, error) {
		authorized := false
		claimedRoles := token.Claims().Get("roles").([]interface{})
		for _, role := range roles {
			for _, claimedRole := range claimedRoles {
				if claimedRole.(string) == role {
					authorized = true
					break
				}
			}
			if authorized {
				break
			}
		}
		return authorized, nil
	})
}

func (a *Auth) AllRoles(h http.Handler, roles ...string) http.Handler {
	return a.Authorize(h, func(token *Token) (bool, error) {
		claimedRoles := token.Claims().Get("roles").([]interface{})
		for _, role := range roles {
			authorized := false
			for _, claimedRole := range claimedRoles {
				if claimedRole.(string) == role {
					authorized = true
					break
				}
			}
			if !authorized {
				return false, nil
			}
		}

		return true, nil
	})
}
