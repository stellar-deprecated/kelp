package networking

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// WebServer defines an interface for a generic HTTP/S server with a StartServer function.
// If certFile and certKey are specified, then the server will serve through HTTPS. StartServer
// will run synchronously and return a non-nil error.
type WebServer interface {
	StartServer(port uint16, certFile string, keyFile string) error
}

type server struct {
	router *http.ServeMux
}

// MakeServer creates a WebServer that's responsible for serving all the endpoints passed into it.
func MakeServer(endpoints []Endpoint) (WebServer, error) {
	mux := new(http.ServeMux)
	s := &server{router: mux}
	for _, endpoint := range endpoints {
		mux.HandleFunc(endpoint.GetPath(), endpoint.GetHandlerFunc())
	}
	return s, nil
}

// StartServer starts the monitoring server by listening on the specified port and serving requests
// according to its handlers. If certFile and keyFile aren't empty, then the server will use TLS.
// This call will block or return a non-nil error.
func (s *server) StartServer(port uint16, certFile string, keyFile string) error {
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
		return http.ListenAndServeTLS(addr, certFile, keyFile, s.router)
	}
	return http.ListenAndServe(addr, s.router)
}
