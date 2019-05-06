package backend

import (
	"fmt"
	"net/http"
)

func version(w http.ResponseWriter, r *http.Request) {
	b, e := runCommand("./bin/kelp version | grep version | cut -d':' -f3")
	if e != nil {
		w.Write([]byte(fmt.Sprintf("error in version command: %s\n", e)))
		return
	}
	w.Write(b)
}
