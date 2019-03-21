package cmd

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Serves the Kelp GUI",
}

func init() {
	serverCmd.Run = func(ccmd *cobra.Command, args []string) {
		r := chi.NewRouter()
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome"))
		})
		http.ListenAndServe(":8000", r)
	}
}
