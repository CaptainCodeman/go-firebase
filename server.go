package firebase

import (
	"encoding/json"
	"net/http"

	"github.com/rs/cors"
	"golang.org/x/net/context"
)

type (
	CreateClaimsFunc func(context.Context, *Token) (*Claims, error)

	Server struct {
		auth     *Auth
		claimsFn CreateClaimsFunc
	}
)

func (a *Auth) Server(claimsFn CreateClaimsFunc) http.Handler {
	s := &Server{
		auth:     a,
		claimsFn: claimsFn,
	}

	// endpoints to issue and verify tokens
	m := http.NewServeMux()
	m.HandleFunc("/token", s.generateHandler)
	m.HandleFunc("/verify", s.verifyHandler)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // TODO: pass in as parameter to restrict?
		AllowedHeaders: []string{"Authorization"},
	})

	return c.Handler(m)
}

func (s *Server) generateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, _ := RequestContext(r)

	// allow authorization token to be sent in querystring (which
	// would avoid a CORS preflight OPTIONS request) or using the
	// Authorization http header (in the format "Bearer token")
	authorization, err := AuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// check that it's valid
	token, err := s.auth.VerifyIDToken(ctx, authorization)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get the firebase user id
	userID, _ := token.UID()

	// call the app-provided function to generate custom claims
	claims, err := s.claimsFn(ctx, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// mint a custom token
	tokenString, err := s.auth.CreateCustomToken(userID, claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write it as text
	w.Write([]byte(tokenString))
}

func (s *Server) verifyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, _ := RequestContext(r)

	authorization, err := AuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check that it's valid
	token, err := s.auth.VerifyIDToken(ctx, authorization)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	enc := json.NewEncoder(w)
	enc.Encode(token.Claims())
}
