package networking

import "net/http"

// AuthLevel specifies the level of authentication needed for an endpoint.
type AuthLevel int

const (
	// NoAuth means that no authentication is required.
	NoAuth AuthLevel = iota
	// GoogleAuth means that a valid Google email is needed to access the endpoint.
	GoogleAuth
)

// Endpoint represents an API endpoint that implements GetHandlerFunc
// which returns a http.HandlerFunc specifying the behavior when this
// endpoint is hit. It's also required to implement GetAuthLevel, which
// returns the level of authentication that's required to access this endpoint.
// Currently, the values can be NoAuth or GoogleAuth. Lastly, GetPath returns the
// path that routes to this endpoint.
type Endpoint interface {
	GetHandlerFunc() http.HandlerFunc
	GetAuthLevel() AuthLevel
	GetPath() string
}
