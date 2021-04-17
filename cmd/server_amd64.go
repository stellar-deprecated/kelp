package cmd

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/asticode/go-astilog"
	"github.com/denisbrodbeck/machineid"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/pkg/browser"
	"github.com/rs/cors"
	"github.com/spf13/cobra"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/kelp/gui"
	"github.com/stellar/kelp/gui/backend"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/kelpos"
	"github.com/stellar/kelp/support/logger"
	"github.com/stellar/kelp/support/networking"
	"github.com/stellar/kelp/support/prefs"
	"github.com/stellar/kelp/support/sdk"
	"github.com/stellar/kelp/support/utils"
)

const kelpAssetsPath = "/assets"
const uiLogsDir = "/ui_logs"
const vendorDirectory = "/vendor"
const trayIconName = "kelp-icon@1-8x.png"
const kelpCcxtPath = "/ccxt"
const ccxtDownloadBaseURL = "https://github.com/stellar/kelp/releases/download/ccxt-rest_v0.0.4"
const ccxtBinaryName = "ccxt-rest"
const ccxtWaitSeconds = 60
const versionPlaceholder = "VERSION_PLACEHOLDER"
const stringPlaceholder = "PLACEHOLDER_URL"
const redirectPlaceholder = "REDIRECT_URL"
const pingPlaceholder = "PING_URL"
const sleepNumSecondsBeforeReadyString = 1
const readyPlaceholder = "READY_STRING"
const readyStringIndicator = "Serving frontend and API server on HTTP port"
const downloadCcxtUpdateIntervalLogMillis = 1000

type serverInputOptions struct {
	port              *uint16
	ports             *uint16
	dev               *bool
	devAPIPort        *uint16
	horizonTestnetURI *string
	horizonPubnetURI  *string
	noHeaders         *bool
	verbose           *bool
	noElectron        *bool
	disablePubnet     *bool
	enableKaas        *bool
	tlsCertFile       *string
	tlsKeyFile        *string
}

// String is the stringer method impl.
func (o serverInputOptions) String() string {
	return fmt.Sprintf("serverInputOptions[port=%d, dev=%v, devAPIPort=%d, horizonTestnetURI='%s', horizonPubnetURI='%s', noHeaders=%v, verbose=%v, noElectron=%v, disablePubnet=%v, enableKaas=%v]",
		*o.port, *o.dev, *o.devAPIPort, *o.horizonTestnetURI, *o.horizonPubnetURI, *o.noHeaders, *o.verbose, *o.noElectron, *o.disablePubnet, *o.enableKaas)
}

func init() {
	options := serverInputOptions{}
	options.port = serverCmd.Flags().Uint16P("port", "p", 8000, "port on which to serve HTTP")
	options.ports = serverCmd.Flags().Uint16P("ports", "P", 8001, "port on which to serve HTTPS (only applicable if tls cert and key provided)")
	options.dev = serverCmd.Flags().Bool("dev", false, "run in dev mode for hot-reloading of JS code")
	options.devAPIPort = serverCmd.Flags().Uint16("dev-api-port", 8002, "port on which to run API server when in dev mode")
	options.horizonTestnetURI = serverCmd.Flags().String("horizon-testnet-uri", "https://horizon-testnet.stellar.org", "URI to use for the horizon instance connected to the Stellar Test Network (must contain the word 'test')")
	options.horizonPubnetURI = serverCmd.Flags().String("horizon-pubnet-uri", "https://horizon.stellar.org", "URI to use for the horizon instance connected to the Stellar Public Network (must not contain the word 'test')")
	options.noHeaders = serverCmd.Flags().Bool("no-headers", false, "do not use Amplitude or set X-App-Name and X-App-Version headers on requests to horizon")
	options.verbose = serverCmd.Flags().BoolP("verbose", "v", false, "enable verbose log lines typically used for debugging")
	options.noElectron = serverCmd.Flags().Bool("no-electron", false, "open in browser instead of using electron, only applies when not in KaaS mode")
	options.disablePubnet = serverCmd.Flags().Bool("disable-pubnet", false, "disable pubnet option")
	options.enableKaas = serverCmd.Flags().Bool("enable-kaas", false, "enable kelp-as-a-service (KaaS) mode, which does not bring up browser or electron")
	options.tlsCertFile = serverCmd.Flags().String("tls-cert-file", "", "path to TLS certificate file")
	options.tlsKeyFile = serverCmd.Flags().String("tls-key-file", "", "path to TLS key file")

	serverCmd.Run = func(ccmd *cobra.Command, args []string) {
		isLocalMode := env == envDev
		isLocalDevMode := isLocalMode && *options.dev
		kos := kelpos.GetKelpOS()
		var e error
		if isLocalMode {
			wd, e := os.Getwd()
			if e != nil {
				panic(errors.Wrap(e, "could not get working directory"))
			}
			if filepath.Base(wd) != "kelp" {
				e := fmt.Errorf("need to invoke from the root 'kelp' directory")
				utils.PrintErrorHintf(e.Error())
				panic(e)
			}
		}

		e = backend.InitBotNameRegex()
		if e != nil {
			panic(errors.Wrap(e, "could not init BotNameRegex: "))
		}

		var logFilepath *kelpos.OSPath
		if !isLocalDevMode {
			l := logger.MakeBasicLogger()
			t := time.Now().Format("20060102T150405MST")
			logFilename := fmt.Sprintf("kelp-ui_%s.log", t)

			uiLogsDirPath := kos.GetDotKelpWorkingDir().Join(uiLogsDir)
			log.Printf("calling mkdir on uiLogsDirPath: %s ...", uiLogsDirPath.AsString())
			// no need to pass a userID since we are not running under the context of any user at this point
			e = kos.Mkdir("_", uiLogsDirPath)
			if e != nil {
				panic(errors.Wrap(e, "could not mkdir on uiLogsDirPath: "+uiLogsDirPath.AsString()))
			}

			// don't use explicit unix filepath here since it uses os.Open directly and won't work on windows
			logFilepath = uiLogsDirPath.Join(logFilename)
			setLogFile(l, logFilepath.Native())

			if *options.verbose {
				astilog.SetDefaultLogger()
			}
		}

		log.Printf("initialized server with cli flag inputs: %s", options)

		if runtime.GOOS == "windows" {
			if !*options.noElectron {
				log.Printf("input options had specified noElectron=false for windows, but electron is not supported on windows yet. force setting noElectron=true for windows.\n")
				*options.noElectron = true
			}
		}

		// create a latch to trigger the browser opening once the backend server is loaded
		openBrowserWg := &sync.WaitGroup{}
		openBrowserWg.Add(1)
		if !isLocalDevMode {
			// don't use explicit unix filepath here since it uses os.Create directly and won't work on windows
			assetsDirPath := kos.GetDotKelpWorkingDir().Join(kelpAssetsPath)
			log.Printf("assetsDirPath: %s", assetsDirPath.AsString())
			trayIconPath := assetsDirPath.Join(trayIconName)
			log.Printf("trayIconPath: %s", trayIconPath.AsString())
			e = writeTrayIcon(kos, trayIconPath, assetsDirPath)
			if e != nil {
				log.Fatal(errors.Wrap(e, "could not write tray icon"))
			}

			htmlContent := tailFileHTML
			if runtime.GOOS == "windows" {
				htmlContent = windowsInitialFile
			}

			appURL := fmt.Sprintf("http://localhost:%d", *options.port)
			pingURL := fmt.Sprintf("http://localhost:%d/ping", *options.port)
			// write out tail.html after setting the file to be tailed
			tailFileCompiled1 := strings.Replace(htmlContent, stringPlaceholder, logFilepath.Native(), -1)
			tailFileCompiled2 := strings.Replace(tailFileCompiled1, redirectPlaceholder, appURL, -1)
			tailFileCompiled3 := strings.Replace(tailFileCompiled2, readyPlaceholder, readyStringIndicator, -1)
			version := strings.TrimSpace(fmt.Sprintf("%s (%s)", guiVersion, version))
			tailFileCompiled4 := strings.Replace(tailFileCompiled3, versionPlaceholder, version, -1)
			tailFileCompiled5 := strings.Replace(tailFileCompiled4, pingPlaceholder, pingURL, -1)
			tailFileCompiled := tailFileCompiled5

			var electronURL string
			if runtime.GOOS == "windows" {
				// start a new web server to serve the tail file since windows does not allow accessing a file directly in electron
				// likely because of the way the file path is specified
				tailFilePort := startTailFileServer(tailFileCompiled)
				electronURL = fmt.Sprintf("http://localhost:%d", tailFilePort)
			} else {
				tailFilepath := kos.GetDotKelpWorkingDir().Join("tail.html")
				fileContents := []byte(tailFileCompiled)
				e := ioutil.WriteFile(tailFilepath.Native(), fileContents, 0644)
				if e != nil {
					panic(fmt.Sprintf("could not write tailfile to path '%s': %s", tailFilepath, e))
				}

				electronURL = tailFilepath.Native()
			}

			// only open browser or electron when not running in kaas mode
			if !*options.enableKaas {
				// kick off the desktop window for UI feedback to the user
				// local mode (non --dev) and release binary should open browser (since --dev already opens browser via yarn and returns)
				go func() {
					if *options.noElectron {
						openBrowser(appURL, openBrowserWg)
					} else {
						openElectron(trayIconPath, electronURL)
					}
				}()
			}
		}

		log.Printf("Starting Kelp GUI Server, gui=%s, cli=%s [%s]\n", guiVersion, version, gitHash)

		checkInitRootFlags()
		if !strings.Contains(*options.horizonTestnetURI, "test") {
			panic("'horizon-testnet-uri' argument must contain the word 'test'")
		}
		if strings.Contains(*options.horizonPubnetURI, "test") {
			panic("'horizon-pubnet-uri' argument must not contain the word 'test'")
		}

		horizonTestnetURI := strings.TrimSuffix(*options.horizonTestnetURI, "/")
		horizonPubnetURI := strings.TrimSuffix(*options.horizonPubnetURI, "/")
		log.Printf("using horizonTestnetURI: %s\n", horizonTestnetURI)
		log.Printf("using horizonPubnetURI: %s\n", horizonPubnetURI)

		if *rootCcxtRestURL == "" {
			*rootCcxtRestURL = "http://localhost:3000"
			e := sdk.SetBaseURL(*rootCcxtRestURL)
			if e != nil {
				panic(fmt.Errorf("unable to set CCXT-rest URL to '%s': %s", *rootCcxtRestURL, e))
			}
		}
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
			if *options.enableKaas {
				apiTestNet.AppName = "kelp--gui-kaas--admin"
				apiPubNet.AppName = "kelp--gui-kaas--admin"
			} else {
				if *options.noElectron {
					apiTestNet.AppName = "kelp--gui-desktop--admin-browser"
					apiPubNet.AppName = "kelp--gui-desktop--admin-browser"
				} else {
					apiTestNet.AppName = "kelp--gui-desktop--admin-electron"
					apiPubNet.AppName = "kelp--gui-desktop--admin-electron"
				}
			}

			apiTestNet.AppVersion = version
			apiPubNet.AppVersion = version

			p := prefs.Make(prefsFilename)
			if p.FirstTime() {
				log.Printf("Kelp sets the `X-App-Name` and `X-App-Version` headers on requests made to Horizon. These headers help us track overall Kelp usage, so that we can learn about general usage patterns and adapt Kelp to be more useful in the future. Kelp also uses Amplitude for metric tracking. These can be turned off using the `--no-headers` flag. See `kelp trade --help` for more information.\n")
				e := p.SetNotFirstTime()
				if e != nil {
					log.Println("")
					log.Printf("unable to create preferences file: %s", e)
					// we can still proceed with this error
				}
			}
		}
		log.Printf("using apiTestNet.AppName = '%s' and apiPubNet.AppName = '%s'", apiTestNet.AppName, apiPubNet.AppName)

		if isLocalDevMode {
			log.Printf("not checking ccxt in local dev mode")
		} else {
			// we need to check twice because sometimes the ccxt process lingers between runs so we can get a false positive on the first check
			e := checkIsCcxtUpTwice(*rootCcxtRestURL)
			ccxtRunning := e == nil
			log.Printf("checked if CCXT is already running, ccxtRunning = %v", ccxtRunning)

			if !ccxtRunning {
				// start ccxt before we make API server (which loads exchange list)
				ccxtGoos := runtime.GOOS
				if ccxtGoos == "windows" {
					ccxtGoos = "linux"
				}

				ccxtDirPath := kos.GetDotKelpWorkingDir().Join(kelpCcxtPath)
				ccxtFilenameNoExt := fmt.Sprintf("ccxt-rest_%s-x64", ccxtGoos)
				filenameWithExt := fmt.Sprintf("%s.zip", ccxtFilenameNoExt)
				ccxtDestDir := ccxtDirPath.Join(ccxtFilenameNoExt)
				ccxtBinPath := ccxtDestDir.Join(ccxtBinaryName)

				log.Printf("mkdir ccxtDirPath: %s ...", ccxtDirPath.AsString())
				// no need to pass a userID since we are not running under the context of any user at this point
				e := kos.Mkdir("_", ccxtDirPath)
				if e != nil {
					panic(fmt.Errorf("could not mkdir for ccxtDirPath: %s", e))
				}

				if runtime.GOOS == "windows" {
					ccxtSourceDir := kos.GetBinDir().Join("ccxt").Join(ccxtFilenameNoExt)
					// no need to pass a userID since we are not running under the context of any user at this point
					e = copyCcxtFolder(kos, "_", ccxtSourceDir, ccxtDestDir)
					if e != nil {
						panic(e)
					}
				} else {
					ccxtBundledZipPath := kos.GetBinDir().Join("ccxt").Join(filenameWithExt)
					ccxtZipDestPath := ccxtDirPath.Join(filenameWithExt)
					// no need to pass a userID since we are not running under the context of any user at this point
					e = copyOrDownloadCcxtBinary(kos, "_", ccxtBundledZipPath, ccxtZipDestPath, filenameWithExt)
					if e != nil {
						panic(e)
					}

					// no need to pass a userID since we are not running under the context of any user at this point
					unzipCcxtFile(kos, "_", ccxtDirPath, ccxtBinPath, filenameWithExt)
				}

				// no need to pass a userID since we are not running under the context of any user at this point
				e = runCcxtBinary(kos, "_", ccxtBinPath)
				if e != nil {
					panic(e)
				}
			}
		}

		var metricsTracker *plugins.MetricsTracker
		if isLocalDevMode {
			log.Printf("metric - not sending data metrics in dev mode")
		} else {
			deviceID, e := machineid.ID()
			if e != nil {
				panic(fmt.Errorf("could not generate machine id: %s", e))
			}
			userID := deviceID                      // reuse for now
			isTestnetOnly := *options.disablePubnet // needs to always match the fronntend
			metricsTracker, e = plugins.MakeMetricsTracker(
				http.DefaultClient,
				amplitudeAPIKey,
				userID,
				"", // use an empty guiUserID because it is sent from the web request via the frontend for each request
				deviceID,
				time.Now(),         // TODO: Find proper time.
				*options.noHeaders, // disable metrics if the CLI specified no headers
				plugins.MakeCommonProps(
					version,
					gitHash,
					env,
					runtime.GOOS,
					runtime.GOARCH,
					"unknown_todo", // TODO DS Determine how to get GOARM.
					runtime.Version(),
					0,
					isTestnetOnly,
					guiVersion,
				),
				nil,
			)
			if e != nil {
				panic(e)
			}
		}

		dataPath := kos.GetDotKelpWorkingDir().Join("bot_data")
		botConfigsPath := dataPath.Join("configs")
		botLogsPath := dataPath.Join("logs")
		s, e := backend.MakeAPIServer(
			kos,
			botConfigsPath,
			botLogsPath,
			*options.horizonTestnetURI,
			apiTestNet,
			*options.horizonPubnetURI,
			apiPubNet,
			*rootCcxtRestURL,
			*options.disablePubnet,
			*options.enableKaas,
			*options.noHeaders,
			quit,
			metricsTracker,
		)
		if e != nil {
			panic(e)
		}

		e = s.InitBackend()
		if e != nil {
			panic(e)
		}

		guiWebPath := kos.GetBinDir().Join("../gui/web")
		if isLocalDevMode {
			// the frontend app checks the REACT_APP_API_PORT variable to be set when serving
			os.Setenv("REACT_APP_API_PORT", fmt.Sprintf("%d", *options.devAPIPort))
			go runAPIServerDevBlocking(s, *options.port, *options.devAPIPort)
			runWithYarn(kos, *options.port, guiWebPath)

			log.Printf("should not have reached here after running yarn")
			return
		}

		options.devAPIPort = nil
		// the frontend app checks the REACT_APP_API_PORT variable to be set when serving
		os.Setenv("REACT_APP_API_PORT", fmt.Sprintf("%d", *options.port))

		if isLocalMode {
			generateStaticFiles(kos, guiWebPath)
		}

		r := chi.NewRouter()
		setMiddleware(r)
		backend.SetRoutes(r, s)
		// gui.FS is automatically compiled based on whether this is a local or deployment build
		gui.FileServer(r, "/", gui.FS)

		isTLS := *options.tlsCertFile != "" && *options.tlsKeyFile != ""
		threadTracker := multithreading.MakeThreadTracker()
		e = threadTracker.TriggerGoroutine(func(inputs []interface{}) {
			port := *options.port
			if isTLS {
				port = *options.ports
			}
			log.Printf("starting server on port %d (TLS enabled = %v)\n", port, isTLS)
			e1 := networking.StartServer(r, port, *options.tlsCertFile, *options.tlsKeyFile)
			if e1 != nil {
				log.Fatal(e1)
			}
		}, nil)
		if e != nil {
			log.Fatal(e)
		}
		if isTLS {
			// we want a new server to redirect traffic from http to https
			httpRedirectMux := chi.NewRouter()
			networking.AddHTTPSUpgrade(httpRedirectMux, "/")
			log.Printf("starting server on port %d to upgrade HTTP requests on the root path '/' to HTTPS connections\n", *options.port)
			e1 := networking.StartServer(httpRedirectMux, *options.port, "", "")
			if e1 != nil {
				log.Fatal(e1)
			}
		}

		log.Printf("sleeping for %d seconds before showing the ready string indicator...\n", sleepNumSecondsBeforeReadyString)
		time.Sleep(sleepNumSecondsBeforeReadyString * time.Second)

		log.Printf("%s: %d\n", readyStringIndicator, *options.port)
		openBrowserWg.Done()
		threadTracker.Wait()

		log.Printf("should not have reached here after starting the backend server")
	}
}

func checkIsCcxtUpTwice(ccxtURL string) error {
	e := isCcxtUp(ccxtURL)
	if e != nil {
		return fmt.Errorf("ccxt-rest was not running on first check: %s", e)
	}

	// tiny pause before second check
	time.Sleep(100 * time.Millisecond)
	e = isCcxtUp(ccxtURL)
	if e != nil {
		return fmt.Errorf("ccxt-rest was not running on second check: %s", e)
	}

	// return nil for no error when it is running
	return nil
}

func setMiddleware(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
}

func copyCcxtFolder(
	kos *kelpos.KelpOS,
	userID string,
	ccxtSourceDir *kelpos.OSPath,
	ccxtDestDir *kelpos.OSPath,
) error {
	log.Printf("copying ccxt directory from %s to location %s ...", ccxtSourceDir.AsString(), ccxtDestDir.AsString())

	cpCmd := fmt.Sprintf("cp -a %s %s", ccxtSourceDir.Unix(), ccxtDestDir.Unix())
	_, e := kos.Blocking(userID, "cp-ccxt", cpCmd)
	if e != nil {
		return fmt.Errorf("unable to copy ccxt directory from %s to %s: %s", ccxtSourceDir.AsString(), ccxtDestDir.AsString(), e)
	}
	log.Printf("... done copying ccxt from %s to location %s", ccxtSourceDir.AsString(), ccxtDestDir.AsString())

	return nil
}

func copyOrDownloadCcxtBinary(
	kos *kelpos.KelpOS,
	userID string,
	ccxtBundledZipPath *kelpos.OSPath,
	ccxtZipDestPath *kelpos.OSPath,
	filenameWithExt string,
) error {
	if _, e := os.Stat(ccxtZipDestPath.Native()); !os.IsNotExist(e) {
		return nil
	}

	if _, e := os.Stat(ccxtBundledZipPath.Native()); !os.IsNotExist(e) {
		log.Printf("copying ccxt from %s to location %s ...", ccxtBundledZipPath.Unix(), ccxtZipDestPath.Unix())

		cpCmd := fmt.Sprintf("cp %s %s", ccxtBundledZipPath.Unix(), ccxtZipDestPath.Unix())
		_, e = kos.Blocking(userID, "cp-ccxt", cpCmd)
		if e != nil {
			return fmt.Errorf("unable to copy ccxt zip file from %s to %s: %s", ccxtBundledZipPath.Unix(), ccxtZipDestPath.Unix(), e)
		}
		log.Printf("... done copying ccxt from %s to location %s", ccxtBundledZipPath.Unix(), ccxtZipDestPath.Unix())

		return nil
	}
	log.Printf("did not find ccxt zip file at source %s, proceeding to download", ccxtBundledZipPath.Unix())

	// else download
	downloadURL := fmt.Sprintf("%s/%s", ccxtDownloadBaseURL, filenameWithExt)
	log.Printf("download ccxt from %s to location: %s ...", downloadURL, ccxtZipDestPath.AsString())
	e := networking.DownloadFileWithGrab(
		downloadURL,
		ccxtZipDestPath.Native(),
		downloadCcxtUpdateIntervalLogMillis,
		func(statusCode int, statusString string) {
			log.Printf("  response_status = %s, code = %d\n", statusString, statusCode)
		},
		func(completedBytes float64, sizeBytes float64, speedBytesPerSec float64) {
			log.Printf("  downloaded %.2f / %.2f MB (%.2f%%) at an average speed of %.2f MB/sec\n",
				completedBytes,
				sizeBytes,
				100*(float64(completedBytes)/float64(sizeBytes)),
				speedBytesPerSec,
			)
		},
		func(filename string) {
			log.Printf("  done\n")
			log.Printf("... downloaded file from URL '%s' to destination '%s'\n", downloadURL, filename)
		},
	)
	if e != nil {
		return fmt.Errorf("could not download ccxt from '%s' to location '%s': %s", downloadURL, ccxtZipDestPath.AsString(), e)
	}
	return nil
}

func unzipCcxtFile(
	kos *kelpos.KelpOS,
	userID string,
	ccxtDir *kelpos.OSPath,
	ccxtBinPath *kelpos.OSPath,
	filenameWithExt string,
) {
	if _, e := os.Stat(ccxtDir.Native()); !os.IsNotExist(e) {
		if _, e := os.Stat(ccxtBinPath.Native()); !os.IsNotExist(e) {
			return
		}
	}

	log.Printf("unzipping file %s ... ", filenameWithExt)
	zipCmd := fmt.Sprintf("cd %s && unzip %s", ccxtDir.Unix(), filenameWithExt)
	_, e := kos.Blocking(userID, "zip", zipCmd)
	if e != nil {
		log.Fatal(errors.Wrap(e, fmt.Sprintf("unable to unzip file %s in directory %s", filenameWithExt, ccxtDir.AsString())))
	}
	log.Printf("done\n")
}

func runCcxtBinary(kos *kelpos.KelpOS, userID string, ccxtBinPath *kelpos.OSPath) error {
	if _, e := os.Stat(ccxtBinPath.Native()); os.IsNotExist(e) {
		return fmt.Errorf("path to ccxt binary (%s) does not exist", ccxtBinPath.AsString())
	}

	log.Printf("running binary %s", ccxtBinPath.AsString())
	// TODO CCXT should be run at the port specified by rootCcxtRestURL, currently it will default to port 3000 even if the config file specifies otherwise
	_, e := kos.Background(userID, "ccxt-rest", ccxtBinPath.Unix())
	if e != nil {
		log.Fatal(errors.Wrap(e, fmt.Sprintf("unable to run ccxt file at location %s", ccxtBinPath.AsString())))
	}

	log.Printf("waiting up to %d seconds for ccxt-rest to start up ...", ccxtWaitSeconds)
	for i := 0; i < ccxtWaitSeconds; i++ {
		e := isCcxtUp(*rootCcxtRestURL)
		ccxtRunning := e == nil

		if ccxtRunning {
			log.Printf("done, waited for ~%d seconds before CCXT was running\n", i)
			return nil
		}

		// wait
		log.Printf("ccxt is not up, sleeping for 1 second (waited so far = %d seconds)\n", i)
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("waited for %d seconds but CCXT was still not running at URL %s", ccxtWaitSeconds, *rootCcxtRestURL)
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

func runWithYarn(kos *kelpos.KelpOS, port uint16, guiWebPath *kelpos.OSPath) {
	// yarn requires the PORT variable to be set when serving
	os.Setenv("PORT", fmt.Sprintf("%d", port))

	log.Printf("Serving frontend via yarn on HTTP port: %d\n", port)
	e := kos.StreamOutput(exec.Command("yarn", "--cwd", guiWebPath.Unix(), "start"))
	if e != nil {
		panic(e)
	}
}

func generateStaticFiles(kos *kelpos.KelpOS, guiWebPath *kelpos.OSPath) {
	log.Printf("generating contents of %s/build ...\n", guiWebPath.Unix())

	e := kos.StreamOutput(exec.Command("yarn", "--cwd", guiWebPath.Unix(), "build"))
	if e != nil {
		panic(e)
	}

	log.Printf("... finished generating contents of %s/build\n", guiWebPath.Unix())
	log.Println()
}

func writeTrayIcon(kos *kelpos.KelpOS, trayIconPath *kelpos.OSPath, assetsDirPath *kelpos.OSPath) error {
	if _, e := os.Stat(trayIconPath.Native()); !os.IsNotExist(e) {
		// file exists, don't write again
		return nil
	}

	// requires icon to be in /resources folder
	trayIconBytes, e := resourcesKelpIcon18xPngBytes()
	if e != nil {
		return errors.Wrap(e, "could not fetch tray icon image bytes")
	}

	img, _, e := image.Decode(bytes.NewReader(trayIconBytes))
	if e != nil {
		return errors.Wrap(e, "could not decode bytes as image data")
	}

	// create dir if not exists
	if _, e := os.Stat(assetsDirPath.Native()); os.IsNotExist(e) {
		log.Printf("mkdir assetsDirPath: %s ...", assetsDirPath.AsString())
		// no need to pass a userID since we are not running under the context of any user at this point
		e = kos.Mkdir("_", assetsDirPath)
		if e != nil {
			return errors.Wrap(e, "could not mkdir for assetsDirPath: "+assetsDirPath.AsString())
		}
		log.Printf("... made assetsDirPath (%s)", assetsDirPath.AsString())
	}

	trayIconFile, e := os.Create(trayIconPath.Native())
	if e != nil {
		return errors.Wrap(e, "could not create tray icon file")
	}
	defer trayIconFile.Close()

	e = png.Encode(trayIconFile, img)
	if e != nil {
		return errors.Wrap(e, "could not write png encoded icon")
	}

	return nil
}

func openBrowser(url string, openBrowserWg *sync.WaitGroup) {
	log.Printf("opening URL in native browser: %s", url)
	openBrowserWg.Wait()

	e := browser.OpenURL(url)
	if e != nil {
		log.Fatal(e)
	}
}

func openElectron(trayIconPath *kelpos.OSPath, url string) {
	log.Printf("opening URL in electron: %s", url)
	quitMenuItemOption := &astilectron.MenuItemOptions{
		Label:   astilectron.PtrStr("Quit"),
		Visible: astilectron.PtrBool(true),
		Enabled: astilectron.PtrBool(true),
		OnClick: astilectron.Listener(func(e astilectron.Event) (deleteListener bool) {
			quit()
			return false
		}),
	}
	mainMenuItemOptions := []*astilectron.MenuItemOptions{
		&astilectron.MenuItemOptions{
			Label: astilectron.PtrStr("File"),
			SubMenu: []*astilectron.MenuItemOptions{
				&astilectron.MenuItemOptions{
					Label: astilectron.PtrStr("Reload"),
					Role:  astilectron.MenuItemRoleReload,
				},
				quitMenuItemOption,
			},
		},
		&astilectron.MenuItemOptions{
			Label: astilectron.PtrStr("Edit"),
			Role:  astilectron.MenuItemRoleEditMenu,
		},
	}

	e := bootstrap.Run(bootstrap.Options{
		AstilectronOptions: astilectron.Options{
			AppName:            "Kelp",
			AppIconDefaultPath: "resources/kelp-icon@2x.png",
			AcceptTCPTimeout:   time.Minute * 2,
		},
		Debug: true,
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
			Image: astilectron.PtrStr(trayIconPath.Native()),
		},
		TrayMenuOptions: []*astilectron.MenuItemOptions{
			quitMenuItemOption,
		},
		MenuOptions: []*astilectron.MenuItemOptions{
			&astilectron.MenuItemOptions{SubMenu: mainMenuItemOptions},
		},
	})
	if e != nil {
		log.Fatal(e)
	}

	quit()
}

func quit() {
	// this is still valid when running in KaaS mode since it doesn't matter. we can disable it (or make it error) if we wanted
	log.Printf("quitting...")
	os.Exit(0)
}

// startTailFileServer takes in anhtml file or a string and serves that on the root of a new url at localhost:port where port is the int returned
func startTailFileServer(tailFileHTML string) int {
	r := chi.NewRouter()
	r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(tailFileHTML))
	}))

	listener, e := net.Listen("tcp", ":0")
	if e != nil {
		log.Fatal(e)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	log.Printf("starting server for tail file on port %d\n", port)
	go func() {
		panic(http.Serve(listener, r))
	}()
	return port
}

const windowsInitialFile = `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
<html>
	<head>
		<title>Kelp GUI (beta) VERSION_PLACEHOLDER</title>
		<script type="text/javascript">
			if (typeof XMLHttpRequest == "undefined") {
				// this is only for really ancient browsers
				XMLHttpRequest = function () {
					try { return new ActiveXObject("Msxml2.xmlHttp.6.0"); }
					catch (e1) { }
					try { return new ActiveXObject("Msxml2.xmlHttp.3.0"); }
					catch (e2) { }
					try { return new ActiveXObject("Msxml2.xmlHttp"); }
					catch (e3) { }
					throw new Error("This browser does not support xmlHttpRequest.");
				};
			}

			var pingUrl = "PING_URL";
			var redirectUrl = "REDIRECT_URL";
			function checkServerOnline() {
				var ajax = new XMLHttpRequest();
				ajax.open("GET", pingUrl, true);
				ajax.onreadystatechange = function () {
					if ((ajax.readyState == 4) && (ajax.status == 200)) {
						window.location.href = redirectUrl;
					}
				}
				ajax.send(null);
			}
		</script>
	</head>
	<body onLoad='setInterval("checkServerOnline()", 1000);' bgcolor="#0D0208" text="#00FF41">
		<div>
			Loading the backend for Kelp.<br />
			This will take a few minutes.<br />
			<br />
			You will be redirected automatically once loaded.<br />
			<br />
			Please be patient.<br />
		</div>
	</body>
</html>
`

const tailFileHTML = `<!-- taken from http://www.davejennifer.com/computerjunk/javascript/tail-dash-f.html with minor modifications -->
<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
<html>
	<head>
		<title>Kelp GUI (beta) VERSION_PLACEHOLDER</title>

		<style>
			.button {
				background-color: #003B00; /* Dark Green */
				color: #00FF41;
				border: 2px solid #00FF41;
				padding: 15px 32px;
				text-align: center;
				text-decoration: none;
				font-size: 16px;
				cursor: pointer;
			}
		</style>
		<script type="text/javascript">
			var lastByte = 0;

			if (typeof XMLHttpRequest == "undefined") {
				// this is only for really ancient browsers
				XMLHttpRequest = function () {
					try { return new ActiveXObject("Msxml2.xmlHttp.6.0"); }
					catch (e1) { }
					try { return new ActiveXObject("Msxml2.xmlHttp.3.0"); }
					catch (e2) { }
					try { return new ActiveXObject("Msxml2.xmlHttp"); }
					catch (e3) { }
					throw new Error("This browser does not support xmlHttpRequest.");
				};
			}

			// Substitute the URL for your server log file here...
			//
			var url = "PLACEHOLDER_URL";

			var visible = false;
			function tailf() {
				var ajax = new XMLHttpRequest();
				ajax.open("POST", url, true);

				if (lastByte == 0) {
					// First request - get everything
				} else {
					//
					// All subsequent requests - add the Range header
					//
					ajax.setRequestHeader("Range", "bytes=" + parseInt(lastByte) + "-");
				}

				ajax.onreadystatechange = function () {
					if (ajax.readyState == 4) {
						if (ajax.status == 200) {
							// only the first request
							lastByte = parseInt(ajax.getResponseHeader("Content-length"));
							document.getElementById("thePlace").innerHTML = ajax.responseText;
							if (visible) {
								document.getElementById("theEnd").scrollIntoView();
							}
						} else if (ajax.status == 206) {
							lastByte += parseInt(ajax.getResponseHeader("Content-length"));
							document.getElementById("thePlace").innerHTML += ajax.responseText;
							if (visible) {
								document.getElementById("theEnd").scrollIntoView();
							}
						} else if (ajax.status == 416) {
							// no new data, so do nothing
						} else {
							//  Some error occurred - just display the status code and response
							alert("Ajax status: " + ajax.status + "\n" + ajax.getAllResponseHeaders());
						}
						
						if (ajax.status == 200 || ajax.status == 206) {
							if (ajax.responseText.includes("READY_STRING")) {
								var redirectURL = "REDIRECT_URL";
								var pingURL = "PING_URL";
								document.getElementById("theEnd").innerHTML = "<br/><br/><b>redirecting to " + redirectURL + " ...</b><br/><br/>";
								document.getElementById("theEnd").scrollIntoView();

								// sleep for 2 seconds so the user sees that we are being redirected
								setTimeout(() => {
									var ajaxPing = new XMLHttpRequest();
									ajaxPing.open("GET", pingURL, true);
									ajaxPing.onreadystatechange = function () {
										if ((ajaxPing.readyState == 4) && (ajaxPing.status == 200)) {
											window.location.href = redirectURL;
										}
									}
									ajaxPing.send(null);
								}, 2000)
							}
						}
					}// ready state 4
				}//orsc function def

				ajax.send(null);

			}// function tailf
		</script>
	
		<script type="text/javascript">
			function onInit() {
				document.getElementById("overHood").style.visibility = "visible";
				document.getElementById("underHood").style.visibility = "hidden";
			}

			function liftHood() {
				document.getElementById("overHood").style.visibility = "hidden";
				document.getElementById("underHood").style.visibility = "visible";
				visible = true;
				document.getElementById("theEnd").scrollIntoView();
			}
		</script>
	</head>

	<body onLoad='onInit(); tailf(); setInterval("tailf()", 250);' bgcolor="#0D0208" text="#00FF41">
		<div>
			<div id="overHood">
				<center>
					<button class="button" onclick='liftHood();'>Show Me What's Under The Hood</button>
				</center>
			</div>
			<div id="underHood">
				<pre id="thePlace"/>
			</div>
			<div id="theEnd"/>
		</div>
	</body>
</html>
`
