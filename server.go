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
		auth           *Auth
		claimsFn       CreateClaimsFunc
		generateURI    string
		verifyURI      string
		allowedOrigins []string
		serveMux       *http.ServeMux
	}
)

// ServerServeMux Uses the existing ServeMux
func ServerServeMux(m *http.ServeMux) func(*Server) {
	return func(s *Server) {
		s.serveMux = m
	}
}

// ServerGenerateURI Sets URI for the token generation
func ServerGenerateURI(uri string) func(*Server) {
	return func(s *Server) {
		s.generateURI = uri
	}
}

// ServerVerifyURI Sets URI for token verification
func ServerVerifyURI(uri string) func(*Server) {
	return func(s *Server) {
		s.verifyURI = uri
	}
}

// ServerAllowedOrigins sets AllowedOrigins for CORS
func ServerAllowedOrigins(origins []string) func(*Server) {
	return func(s *Server) {
		s.allowedOrigins = origins
	}
}

func (a *Auth) Server(claimsFn CreateClaimsFunc, options ...func(*Server)) http.Handler {
	s := &Server{
		auth:     a,
		claimsFn: claimsFn,
	}

	for _, option := range options {
		option(s)
	}

	// Setting defaults if empty
	if len(s.generateURI) == 0 {
		s.generateURI = "/token"
	}

	if len(s.verifyURI) == 0 {
		s.verifyURI = "/verify"
	}

	if len(s.allowedOrigins) == 0 {
		s.allowedOrigins = []string{"*"}
	}

	// endpoints to issue and verify tokens
	m := http.NewServeMux()
	if s.serveMux != nil {
		m = s.serveMux
	}

	m.HandleFunc(s.generateURI, s.generateHandler)
	m.HandleFunc(s.verifyURI, s.verifyHandler)

	c := cors.New(cors.Options{
		AllowedOrigins: s.allowedOrigins,
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

