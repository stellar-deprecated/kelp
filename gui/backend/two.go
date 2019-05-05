package backend

import "net/http"

func two(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World two"))
}
