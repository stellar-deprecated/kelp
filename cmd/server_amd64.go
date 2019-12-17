package cmd

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/kelp/gui"
	"github.com/stellar/kelp/gui/backend"
	"github.com/stellar/kelp/support/kelpos"
	"github.com/stellar/kelp/support/prefs"
)

const urlOpenDelayMillis = 1500
const kelpPrefsDirectory = ".kelp"
const kelpAssetsPath = "/assets"
const trayIconName = "kelp-icon@1-8x.png"

type serverInputs struct {
	port              *uint16
	dev               *bool
	devAPIPort        *uint16
	horizonTestnetURI *string
	horizonPubnetURI  *string
	noHeaders         *bool
}

func init() {
	hasUICapability = true

	options := serverInputs{}
	options.port = serverCmd.Flags().Uint16P("port", "p", 8000, "port on which to serve")
	options.dev = serverCmd.Flags().Bool("dev", false, "run in dev mode for hot-reloading of JS code")
	options.devAPIPort = serverCmd.Flags().Uint16("dev-api-port", 8001, "port on which to run API server when in dev mode")
	options.horizonTestnetURI = serverCmd.Flags().String("horizon-testnet-uri", "https://horizon-testnet.stellar.org", "URI to use for the horizon instance connected to the Stellar Test Network (must contain the word 'test')")
	options.horizonPubnetURI = serverCmd.Flags().String("horizon-pubnet-uri", "https://horizon.stellar.org", "URI to use for the horizon instance connected to the Stellar Public Network (must not contain the word 'test')")
	options.noHeaders = serverCmd.Flags().Bool("no-headers", false, "do not set X-App-Name and X-App-Version headers on requests to horizon")

	serverCmd.Run = func(ccmd *cobra.Command, args []string) {
		log.Printf("Starting Kelp GUI Server: %s [%s]\n", version, gitHash)

		checkInitRootFlags()
		if !strings.Contains(*options.horizonTestnetURI, "test") {
			panic("'horizon-testnet-uri' argument must contain the word 'test'")
		}
		if strings.Contains(*options.horizonPubnetURI, "test") {
			panic("'horizon-pubnet-uri' argument must not contain the word 'test'")
		}

		kos := kelpos.GetKelpOS()
		horizonTestnetURI := strings.TrimSuffix(*options.horizonTestnetURI, "/")
		horizonPubnetURI := strings.TrimSuffix(*options.horizonPubnetURI, "/")
		log.Printf("using horizonTestnetURI: %s\n", horizonTestnetURI)
		log.Printf("using horizonPubnetURI: %s\n", horizonPubnetURI)
		log.Printf("using ccxtRestUrl: %s\n", *rootCcxtRestURL)
		apiTestNet := &horizonclient.Client{
			HorizonURL: horizonTestnetURI,
			HTTP:       http.DefaultClient,
		}
		apiPubNet := &horizonclient.Client{
			HorizonURL: horizonPubnetURI,
			HTTP:       http.DefaultClient,
		}
		if !*options.noHeaders {
			apiTestNet.AppName = "kelp-ui"
			apiTestNet.AppVersion = version
			apiPubNet.AppName = "kelp-ui"
			apiPubNet.AppVersion = version

			p := prefs.Make(prefsFilename)
			if p.FirstTime() {
				log.Printf("Kelp sets the `X-App-Name` and `X-App-Version` headers on requests made to Horizon. These headers help us track overall Kelp usage, so that we can learn about general usage patterns and adapt Kelp to be more useful in the future. These can be turned off using the `--no-headers` flag. See `kelp trade --help` for more information.\n")
				e := p.SetNotFirstTime()
				if e != nil {
					log.Println("")
					log.Printf("unable to create preferences file: %s", e)
					// we can still proceed with this error
				}
			}
		}
		s, e := backend.MakeAPIServer(kos, *options.horizonTestnetURI, apiTestNet, *options.horizonPubnetURI, apiPubNet, *rootCcxtRestURL, *options.noHeaders)
		if e != nil {
			panic(e)
		}

		if env == envDev && *options.dev {
			checkHomeDir()
			// the frontend app checks the REACT_APP_API_PORT variable to be set when serving
			os.Setenv("REACT_APP_API_PORT", fmt.Sprintf("%d", *options.devAPIPort))
			go runAPIServerDevBlocking(s, *options.port, *options.devAPIPort)
			runWithYarn(kos, options)
			return
		}

		options.devAPIPort = nil
		// the frontend app checks the REACT_APP_API_PORT variable to be set when serving
		os.Setenv("REACT_APP_API_PORT", fmt.Sprintf("%d", *options.port))

		if env == envDev {
			checkHomeDir()
			generateStaticFiles(kos)
		}

		r := chi.NewRouter()
		setMiddleware(r)
		backend.SetRoutes(r, s)
		// gui.FS is automatically compiled based on whether this is a local or deployment build
		gui.FileServer(r, "/", gui.FS)

		portString := fmt.Sprintf(":%d", *options.port)
		log.Printf("Serving frontend and API server on HTTP port: %d\n", *options.port)
		// local mode (non --dev) and release binary should open browser (since --dev already opens browser via yarn and returns)
		go func() {
			url := fmt.Sprintf("http://localhost:%d", *options.port)
			log.Printf("A browser window will open up automatically to %s\n", url)
			time.Sleep(urlOpenDelayMillis * time.Millisecond)
			openBrowser(kos, url)
		}()

		if env == envDev {
			e = http.ListenAndServe(portString, r)
			if e != nil {
				log.Fatal(e)
			}
		} else {
			_ = http.ListenAndServe(portString, r)
		}
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

func runWithYarn(kos *kelpos.KelpOS, options serverInputs) {
	// yarn requires the PORT variable to be set when serving
	os.Setenv("PORT", fmt.Sprintf("%d", *options.port))

	log.Printf("Serving frontend via yarn on HTTP port: %d\n", *options.port)
	e := kos.StreamOutput(exec.Command("yarn", "--cwd", "gui/web", "start"))
	if e != nil {
		panic(e)
	}
}

func generateStaticFiles(kos *kelpos.KelpOS) {
	log.Printf("generating contents of gui/web/build ...\n")

	e := kos.StreamOutput(exec.Command("yarn", "--cwd", "gui/web", "build"))
	if e != nil {
		panic(e)
	}

	log.Printf("... finished generating contents of gui/web/build\n")
	log.Println()
}

func writeTrayIcon(kos *kelpos.KelpOS) (string, error) {
	binDirectory, e := getBinaryDirectory()
	if e != nil {
		return "", errors.Wrap(e, "could not get binary directory")
	}
	log.Printf("binDirectory: %s", binDirectory)
	assetsDirPath := filepath.Join(binDirectory, kelpPrefsDirectory, kelpAssetsPath)
	log.Printf("assetsDirPath: %s", assetsDirPath)
	trayIconPath := filepath.Join(assetsDirPath, trayIconName)
	log.Printf("trayIconPath: %s", trayIconPath)
	if _, e := os.Stat(trayIconPath); !os.IsNotExist(e) {
		// file exists, don't write again
		return trayIconPath, nil
	}

	trayIconBytes, e := resourcesKelpIcon18xPngBytes()
	if e != nil {
		return "", errors.Wrap(e, "could not fetch tray icon image bytes")
	}

	img, _, e := image.Decode(bytes.NewReader(trayIconBytes))
	if e != nil {
		return "", errors.Wrap(e, "could not decode bytes as image data")
	}

	// create dir if not exists
	if _, e := os.Stat(assetsDirPath); os.IsNotExist(e) {
		log.Printf("making assetsDirPath: %s ...", assetsDirPath)
		e = kos.Mkdir(assetsDirPath)
		if e != nil {
			return "", errors.Wrap(e, "could not make directories for assetsDirPath: "+assetsDirPath)
		}
		log.Printf("... made assetsDirPath (%s)", assetsDirPath)
	}

	trayIconFile, e := os.Create(trayIconPath)
	if e != nil {
		return "", errors.Wrap(e, "could not create tray icon file")
	}
	defer trayIconFile.Close()

	e = png.Encode(trayIconFile, img)
	if e != nil {
		return "", errors.Wrap(e, "could not write png encoded icon")
	}

	return trayIconPath, nil
}

func getBinaryDirectory() (string, error) {
	return filepath.Abs(filepath.Dir(os.Args[0]))
}

func openBrowser(kos *kelpos.KelpOS, url string) {
	trayIconPath, e := writeTrayIcon(kos)
	if e != nil {
		log.Fatal(errors.Wrap(e, "could not write tray icon"))
	}

	e = bootstrap.Run(bootstrap.Options{
		AstilectronOptions: astilectron.Options{
			AppName:            "Kelp",
			AppIconDefaultPath: "resources/kelp-icon@2x.png",
		},
		Debug: false,
		Windows: []*bootstrap.Window{&bootstrap.Window{
			Homepage: url,
			Options: &astilectron.WindowOptions{
				Center:   astilectron.PtrBool(true),
				Width:    astilectron.PtrInt(1280),
				Height:   astilectron.PtrInt(960),
				Closable: astilectron.PtrBool(false),
			},
		}},
		TrayOptions: &astilectron.TrayOptions{
			Image: astilectron.PtrStr(trayIconPath),
		},
		TrayMenuOptions: []*astilectron.MenuItemOptions{
			&astilectron.MenuItemOptions{
				Label:   astilectron.PtrStr("Quit"),
				Visible: astilectron.PtrBool(true),
				Enabled: astilectron.PtrBool(true),
				OnClick: astilectron.Listener(func(e astilectron.Event) (deleteListener bool) {
					quit()
					return false
				}),
			},
		},
	})
	if e != nil {
		log.Fatal(e)
	}

	quit()
}

func quit() {
	log.Printf("quitting...")
	os.Exit(0)
}
