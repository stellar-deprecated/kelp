package backend

import (
	"fmt"
	"net/http"

	"github.com/stellar/go/keypair"
)

func (s *APIServer) newSecretKey(w http.ResponseWriter, r *http.Request) {
	kp, e := keypair.Random()
	if e != nil {
		s.writeError(w, fmt.Sprintf("error generating keypair: %s\n", e))
		return
	}
	seed := kp.Seed()
	w.Write([]byte(seed))
}
