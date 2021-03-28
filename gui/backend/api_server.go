package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/kelpos"
)

// APIServer is an instance of the API service
type APIServer struct {
	kelpBinPath       *kelpos.OSPath
	botConfigsPath    *kelpos.OSPath
	botLogsPath       *kelpos.OSPath
	kos               *kelpos.KelpOS
	horizonTestnetURI string
	horizonPubnetURI  string
	ccxtRestUrl       string
	apiTestNet        *horizonclient.Client
	apiPubNet         *horizonclient.Client
	disablePubnet     bool
	noHeaders         bool
	quitFn            func()
	metricsTracker    *plugins.MetricsTracker
	kelpErrorMap      map[string]KelpError
	kelpErrorMapLock  *sync.Mutex

	cachedOptionsMetadata metadata
}

// MakeAPIServer is a factory method
func MakeAPIServer(
	kos *kelpos.KelpOS,
	botConfigsPath *kelpos.OSPath,
	botLogsPath *kelpos.OSPath,
	horizonTestnetURI string,
	apiTestNet *horizonclient.Client,
	horizonPubnetURI string,
	apiPubNet *horizonclient.Client,
	ccxtRestUrl string,
	disablePubnet bool,
	noHeaders bool,
	quitFn func(),
	metricsTracker *plugins.MetricsTracker,
) (*APIServer, error) {
	kelpBinPath := kos.GetBinDir().Join(filepath.Base(os.Args[0]))

	optionsMetadata, e := loadOptionsMetadata()
	if e != nil {
		return nil, fmt.Errorf("error while loading options metadata when making APIServer: %s", e)
	}

	kelpErrorMap := map[string]KelpError{}

	return &APIServer{
		kelpBinPath:           kelpBinPath,
		botConfigsPath:        botConfigsPath,
		botLogsPath:           botLogsPath,
		kos:                   kos,
		horizonTestnetURI:     horizonTestnetURI,
		horizonPubnetURI:      horizonPubnetURI,
		ccxtRestUrl:           ccxtRestUrl,
		apiTestNet:            apiTestNet,
		apiPubNet:             apiPubNet,
		disablePubnet:         disablePubnet,
		noHeaders:             noHeaders,
		cachedOptionsMetadata: optionsMetadata,
		quitFn:                quitFn,
		metricsTracker:        metricsTracker,
		kelpErrorMap:          kelpErrorMap,
		kelpErrorMapLock:      &sync.Mutex{},
	}, nil
}

// InitBackend initializes anything required to get the backend ready to serve
func (s *APIServer) InitBackend() error {
	// initial load of bots into memory
	_, e := s.doListBots()
	if e != nil {
		return fmt.Errorf("error listing/loading bots: %s", e)
	}

	return nil
}

func (s *APIServer) parseBotName(r *http.Request) (string, error) {
	botNameBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		return "", fmt.Errorf("error when reading request input: %s", e)
	}
	return string(botNameBytes), nil
}

func (s *APIServer) writeError(w http.ResponseWriter, message string) {
	log.Print(message)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(message))
}

// ErrorResponse represents an error (deprecated)
type ErrorResponse struct {
	Error string `json:"error"`
}

// KelpError represents an error
type KelpError struct {
	UUID       string     `json:"uuid"`
	ObjectType errorType  `json:"object_type"`
	ObjectName string     `json:"object_name"`
	Date       time.Time  `json:"date"`
	Level      errorLevel `json:"level"`
	Message    string     `json:"message"`
}

// String is the Stringer method
func (ke *KelpError) String() string {
	return fmt.Sprintf("KelpError[UUID=%s, objectType=%s, objectName=%s, date=%s, level=%s, message=%s]", ke.UUID, ke.ObjectType, ke.ObjectName, ke.Date.Format("20060102T150405MST"), ke.Level, ke.Message)
}

// KelpErrorResponseWrapper is the outer object that contains the Kelp Error
type KelpErrorResponseWrapper struct {
	KelpError KelpError `json:"kelp_error"`
}

func makeKelpErrorResponseWrapper(
	objectType errorType,
	objectName string,
	date time.Time,
	level errorLevel,
	message string,
) KelpErrorResponseWrapper {
	uuid, e := uuid.NewRandom()
	if e != nil {
		// TODO NS - panic here instead of returning and handling smoothly because interface is a lot cleaner without returning error
		// need to find a better solution that does not require a panic
		panic(fmt.Errorf("unable to generate new uuid: %s", e))
	}

	return KelpErrorResponseWrapper{
		KelpError: KelpError{
			UUID:       uuid.String(),
			ObjectType: objectType,
			ObjectName: objectName,
			Date:       date,
			Level:      level,
			Message:    message,
		},
	}
}

// String is the Stringer method
func (kerw *KelpErrorResponseWrapper) String() string {
	return fmt.Sprintf("KelpErrorResponseWrapper[kelp_error=%s]", kerw.KelpError.String())
}

func (s *APIServer) writeErrorJson(w http.ResponseWriter, message string) {
	log.Println(message)
	w.WriteHeader(http.StatusInternalServerError)

	marshalledJson, e := json.MarshalIndent(ErrorResponse{Error: message}, "", "    ")
	if e != nil {
		log.Printf("unable to marshal json with indentation: %s\n", e)
		w.Write([]byte(fmt.Sprintf("unable to marshal json with indentation: %s\n", e)))
		return
	}
	w.Write(marshalledJson)
}

func (s *APIServer) addKelpErrorToMap(ke KelpError) {
	key := ke.UUID

	// need to use a lock because we could encounter a "concurrent map writes" error against the map which is being updated by multiple threads
	s.kelpErrorMapLock.Lock()
	defer s.kelpErrorMapLock.Unlock()

	s.kelpErrorMap[key] = ke
}

func (s *APIServer) writeKelpError(w http.ResponseWriter, kerw KelpErrorResponseWrapper) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Printf("writing error: %s\n", kerw.String())
	s.addKelpErrorToMap(kerw.KelpError)

	marshalledJSON, e := json.MarshalIndent(kerw, "", "    ")
	if e != nil {
		log.Printf("unable to marshal json with indentation: %s\n", e)
		w.Write([]byte(fmt.Sprintf("unable to marshal json with indentation: %s\n", e)))
		return
	}
	w.Write(marshalledJSON)
}

func (s *APIServer) writeJson(w http.ResponseWriter, v interface{}) {
	s.writeJsonWithLog(w, v, true)
}

func (s *APIServer) writeJsonWithLog(w http.ResponseWriter, v interface{}, doLog bool) {
	marshalledJson, e := json.MarshalIndent(v, "", "    ")
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("unable to marshal json with indentation: %s", e))
		return
	}

	if doLog {
		log.Printf("responseJson: %s\n", string(marshalledJson))
	}
	w.WriteHeader(http.StatusOK)
	w.Write(marshalledJson)
}

func (s *APIServer) runKelpCommandBlocking(namespace string, cmd string) ([]byte, error) {
	// There is a weird issue on windows where the absolute path for the kelp binary does not work on the release GUI
	// version because of the unzipped directory name but it will work on the released cli version or if we change the
	// name of the folder in which the GUI version is unzipped.
	// To avoid these issues we only invoke with the binary name as opposed to the absolute path that contains the
	// directory name. see start_bot.go for some experimentation with absolute and relative paths
	cmdString := fmt.Sprintf("%s %s", s.kelpBinPath.Unix(), cmd)
	return s.kos.Blocking(namespace, cmdString)
}

func (s *APIServer) runKelpCommandBackground(namespace string, cmd string) (*kelpos.Process, error) {
	// There is a weird issue on windows where the absolute path for the kelp binary does not work on the release GUI
	// version because of the unzipped directory name but it will work on the released cli version or if we change the
	// name of the folder in which the GUI version is unzipped.
	// To avoid these issues we only invoke with the binary name as opposed to the absolute path that contains the
	// directory name. see start_bot.go for some experimentation with absolute and relative paths
	cmdString := fmt.Sprintf("%s %s", s.kelpBinPath.Unix(), cmd)
	return s.kos.Background(namespace, cmdString)
}

func (s *APIServer) setupOpsDirectory() error {
	e := s.kos.Mkdir(s.botConfigsPath)
	if e != nil {
		return fmt.Errorf("error setting up configs directory (%s): %s\n", s.botConfigsPath, e)
	}

	e = s.kos.Mkdir(s.botLogsPath)
	if e != nil {
		return fmt.Errorf("error setting up logs directory (%s): %s\n", s.botLogsPath, e)
	}

	return nil
}
