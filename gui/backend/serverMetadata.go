package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ServerMetadataResponse is the response from the /serverMetadata endpoint
type ServerMetadataResponse struct {
	DisablePubnet bool `json:"disable_pubnet"`
}

func (s *APIServer) serverMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := ServerMetadataResponse{
		DisablePubnet: s.disablePubnet,
	}

	b, e := json.Marshal(metadata)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("unable to marshal json with indentation: %s", e))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
