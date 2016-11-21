package main

import (
	"net/http"

	"github.com/captaincodeman/go-firebase"
	"github.com/rs/cors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var auth *firebase.Auth

func init() {
	// default firebase app, uses firebase-credentials.json file
	fb, _ := firebase.New()
	auth = fb.Auth()

	// for calling remotely from our front-end
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Authorization"},
	})
	mux := c.Handler(http.HandlerFunc(handler))
	http.Handle("/", mux)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// allow authorization token to be sent in querystring (which
	// would avoid a CORS preflight OPTIONS request) or using the
	// Authorization http header (in the format "Bearer token")
	authorization, err := firebase.AuthorizationFromRequest(r)
	if err != nil {
		log.Errorf(ctx, "AuthorizationFromRequest %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// check that it's valid
	token, err := auth.VerifyIDToken(ctx, authorization)
	if err != nil {
		log.Errorf(ctx, "VerifyIDToken %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get the firebase user id
	userID, _ := token.UID()

	// Here is where we'd lookup the user and set the custom claims
	// that we want to be added to the token we're going to produce
	developerClaims := make(firebase.Claims)
	developerClaims["uid"] = 1 // our internal system id
	developerClaims["roles"] = []string{
		"admin",
		"operator",
	}

	// mint a custom token
	tokenString, err := auth.CreateCustomToken(userID, &developerClaims)
	if err != nil {
		log.Errorf(ctx, "CreateCustomToken %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: set headers for no-cacheability

	// write it as text
	w.Write([]byte(tokenString))
}
