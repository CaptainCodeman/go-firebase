package main

import (
	"net/http"

	"encoding/json"
	"github.com/captaincodeman/go-firebase"
	"github.com/rs/cors"
)

var auth *firebase.Auth

func init() {
	// default firebase app, uses firebase-credentials.json file
	fb, _ := firebase.New()
	auth = fb.Auth()

	// auth server comes with CORS included
	http.Handle("/", auth.Server(customClaims))

	// but we need to add it for the API endpoint
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Authorization"},
	})

	// example API endpoint with role checks
	http.Handle("/api", c.Handler(auth.AnyRole(http.HandlerFunc(apiHandler), "operator")))
}

func customClaims(token *firebase.Token) *firebase.Claims {
	// get the firebase user id for lookup
	// userID, _ := token.UID()

	// Here is where we'd lookup the user and set the custom claims
	// that we want to be added to the token we're going to produce
	claims := make(firebase.Claims)
	claims["uid"] = 1
	claims["roles"] = []string{
		"operator",
	}

	return &claims
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		ID      int    `json:"id"`
		Message string `json:"message"`
	}{
		ID:      1,
		Message: "Hello World",
	}

	enc := json.NewEncoder(w)
	enc.Encode(data)
}
