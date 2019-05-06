package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/stellar/kelp/gui"
	"github.com/stellar/kelp/gui/backend"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Serves the Kelp GUI",
}

type serverInputs struct {
	port       *uint16
	dev        *bool
	devAPIPort *uint16
}

func init() {
	options := serverInputs{}
	options.port = serverCmd.Flags().Uint16P("port", "p", 8000, "port on which to serve")
	options.dev = serverCmd.Flags().Bool("dev", false, "run in dev mode for hot-reloading of JS code")
	options.devAPIPort = serverCmd.Flags().Uint16("dev-api-port", 8001, "port on which to run API server when in dev mode")

	serverCmd.Run = func(ccmd *cobra.Command, args []string) {
		s, e := backend.MakeAPIServer()
		if e != nil {
			panic(e)
		}

		if env == envDev && *options.dev {
			checkHomeDir()
			// the frontend app checks the REACT_APP_API_PORT variable to be set when serving
			os.Setenv("REACT_APP_API_PORT", fmt.Sprintf("%d", *options.devAPIPort))
			go runAPIServerDevBlocking(s, *options.port, *options.devAPIPort)
			runWithYarn(options)
			return
		} else {
			options.devAPIPort = nil
			// the frontend app checks the REACT_APP_API_PORT variable to be set when serving
			os.Setenv("REACT_APP_API_PORT", fmt.Sprintf("%d", *options.port))
		}

		if env == envDev {
			checkHomeDir()
			generateStaticFiles()
		}

		r := chi.NewRouter()
		setMiddleware(r)
		backend.SetRoutes(r, s)
		// gui.FS is automatically compiled based on whether this is a local or deployment build
		gui.FileServer(r, "/", gui.FS)

		portString := fmt.Sprintf(":%d", *options.port)
		log.Printf("Serving frontend and API server on HTTP port: %d\n", *options.port)
		e = http.ListenAndServe(portString, r)
		log.Fatal(e)
	}
}

func setMiddleware(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
}

func runAPIServerDevBlocking(s *backend.APIServer, frontendPort uint16, devAPIPort uint16) {
	r := chi.NewRouter()
	// Add CORS middleware around every request since both ports are different when running server in dev mode
	r.Use(cors.New(cors.Options{
		AllowedOrigins: []string{fmt.Sprintf("http://localhost:%d", frontendPort)},
	}).Handler)

	setMiddleware(r)
	backend.SetRoutes(r, s)
	portString := fmt.Sprintf(":%d", devAPIPort)
	log.Printf("Serving API server on HTTP port: %d\n", devAPIPort)
	e := http.ListenAndServe(portString, r)
	log.Fatal(e)
}

func checkHomeDir() {
	op, e := exec.Command("pwd").Output()
	if e != nil {
		panic(e)
	}
	result := strings.TrimSpace(string(op))

	if !strings.HasSuffix(result, "/kelp") {
		log.Fatalf("need to invoke the '%s' command while in the root 'kelp' directory\n", serverCmd.Use)
	}
}

func runWithYarn(options serverInputs) {
	// yarn requires the PORT variable to be set when serving
	os.Setenv("PORT", fmt.Sprintf("%d", *options.port))

	log.Printf("Serving frontend via yarn on HTTP port: %d\n", *options.port)
	e := runCommandStreamOutput(exec.Command("yarn", "--cwd", "gui/web", "start"))
	if e != nil {
		panic(e)
	}
}

func generateStaticFiles() {
	log.Printf("generating contents of gui/web/build ...\n")

	e := runCommandStreamOutput(exec.Command("yarn", "--cwd", "gui/web", "build"))
	if e != nil {
		panic(e)
	}

	log.Printf("... finished generating contents of gui/web/build\n")
	log.Println()
}

func runCommandStreamOutput(command *exec.Cmd) error {
	stdout, e := command.StdoutPipe()
	if e != nil {
		return fmt.Errorf("error while creating Stdout pipe: %s", e)
	}
	command.Start()

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("\t%s\n", line)
	}

	e = command.Wait()
	if e != nil {
		return fmt.Errorf("could not execute command: %s", e)
	}
	return nil
}
