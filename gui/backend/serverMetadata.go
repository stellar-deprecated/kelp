package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/stellar/kelp/support/guiconfig"
)

// ServerMetadataResponse is the response from the /serverMetadata endpoint
type ServerMetadataResponse struct {
	DisablePubnet bool `json:"disable_pubnet"`
	EnableKaas    bool `json:"enable_kaas"`
	GuiConfig	  guiconfig.GUIConfig `json:"guiconfig"`
}

func (s *APIServer) serverMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := ServerMetadataResponse{
		DisablePubnet: s.disablePubnet,
		EnableKaas:    s.enableKaas,
		GuiConfig:	   s.guiConfig,
	}

	b, e := json.Marshal(metadata)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("unable to marshal json with indentation: %s", e))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
