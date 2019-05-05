package backend

import "net/http"

func one(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World one"))
}
