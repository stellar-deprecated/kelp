package networking

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/lechengfan/googleauth"
)

// Config is a struct storing configuration parameters for the monitoring server.
type Config struct {
	// PermittedEmails is a map of emails that are allowed access to the endpoints protected by Google authentication
	PermittedEmails map[string]bool
	// GoogleClientID - client ID of the Google application. It should only be left empty if no endpoints require
	// Google authentication
	GoogleClientID string
	// GoogleClientSecret - client secret of the Google application. It should only be left empty if no endpoints require
	// Google authentication
	GoogleClientSecret string
}

// WebServer defines an interface for a generic HTTP/S server with a StartServer function.
// If certFile and certKey are specified, then the server will serve through HTTPS. StartServer
// will run synchronously and return a non-nil error.
type WebServer interface {
	StartServer(port uint16, certFile string, keyFile string) error
}

type server struct {
	router             *http.ServeMux
	googleClientID     string
	googleClientSecret string
	permittedEmails    map[string]bool
}

// MakeServerWithGoogleAuth creates a WebServer that's responsible for serving all the endpoints passed into it with google authentication.
func MakeServerWithGoogleAuth(cfg *Config, endpoints []Endpoint) (WebServer, error) {
	mux := new(http.ServeMux)
	s := &server{
		router:             mux,
		googleClientID:     cfg.GoogleClientID,
		googleClientSecret: cfg.GoogleClientSecret,
		permittedEmails:    cfg.PermittedEmails,
	}
	// Router for endpoints that require authentication
	authMux := new(http.ServeMux)
	googleAuthRequired := false
	for _, endpoint := range endpoints {
		if endpoint.GetAuthLevel() == GoogleAuth {
			if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
				return nil, fmt.Errorf("error registering a GoogleAuth endpoint - google client ID or client secret is empty")
			}
			googleAuthRequired = true
			authMux.HandleFunc(endpoint.GetPath(), endpoint.GetHandlerFunc())
		} else {
			mux.HandleFunc(endpoint.GetPath(), endpoint.GetHandlerFunc())
		}
	}
	if googleAuthRequired {
		mux.Handle("/", s.googleAuthHandler(authMux))
	}
	return s, nil
}

// StartServer starts the server by using the router on the server struct
func (s *server) StartServer(port uint16, certFile string, keyFile string) error {
	return StartServer(s.router, port, certFile, keyFile)
}

// StartServer starts the monitoring server by listening on the specified port and serving requests
// according to its handlers. If certFile and keyFile aren't empty, then the server will use TLS.
// This call will block or return a non-nil error.
func StartServer(handler http.Handler, port uint16, certFile string, keyFile string) error {
	addr := ":" + strconv.Itoa(int(port))
	if certFile != "" && keyFile != "" {
		_, e := os.Stat(certFile)
		if e != nil {
			return fmt.Errorf("provided tls cert file cannot be found")
		}
		_, e = os.Stat(keyFile)
		if e != nil {
			return fmt.Errorf("provided tls key file cannot be found")
		}
		return http.ListenAndServeTLS(addr, certFile, keyFile, handler)
	}
	return http.ListenAndServe(addr, handler)
}

// googleAuthHandler creates a random key to encrypt/decrypt session cookies
// and returns a handler for the Google OAuth process, where the user is asked
// to sign in via Google. If successful, h is used as the callback.
func (s *server) googleAuthHandler(h http.Handler) http.Handler {
	key := make([]byte, 64)
	_, e := rand.Read(key)
	if e != nil {
		log.Printf("error encountered generating random key for google auth")
		panic(e)
	}
	k := sha256.Sum256(key)
	return &googleauth.Handler{
		PermittedEmails: s.permittedEmails,
		Keys:            []*[32]byte{&k},
		ClientID:        s.googleClientID,
		ClientSecret:    s.googleClientSecret,
		MaxAge:          28 * 24 * time.Hour,
		Handler:         h,
	}
}

// AddHTTPSUpgrade adds an entry on the passed in path to redirect to an https connection
func AddHTTPSUpgrade(mux *chi.Mux, path string) {
	mux.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("received request on http port, redirecting to https connection using a temporary redirect (http status code 307)")
		http.Redirect(w, r, fmt.Sprintf("https://%s%s", r.Host, path), http.StatusTemporaryRedirect)
	}))
}
