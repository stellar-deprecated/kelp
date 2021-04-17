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

// UserData is the json data passed in to represent a user
type UserData struct {
	ID string `json:"id"`
}

// toUser converts to the format understood by kelpos
func (u UserData) toUser() *kelpos.User {
	return kelpos.MakeUser(u.ID)
}

// String is the stringer method
func (u UserData) String() string {
	return fmt.Sprintf("UserData[ID=%s]", u.ID)
}

// kelpErrorDataForUser tracks errors for a given user
type kelpErrorDataForUser struct {
	errorMap map[string]KelpError
	lock     *sync.Mutex
}

// APIServer is an instance of the API service
type APIServer struct {
	kelpBinPath          *kelpos.OSPath
	botConfigsPath       *kelpos.OSPath
	botLogsPath          *kelpos.OSPath
	kos                  *kelpos.KelpOS
	horizonTestnetURI    string
	horizonPubnetURI     string
	ccxtRestUrl          string
	apiTestNet           *horizonclient.Client
	apiPubNet            *horizonclient.Client
	disablePubnet        bool
	enableKaas           bool
	noHeaders            bool
	quitFn               func()
	metricsTracker       *plugins.MetricsTracker
	kelpErrorsByUser     map[string]kelpErrorDataForUser
	kelpErrorsByUserLock *sync.Mutex

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
	enableKaas bool,
	noHeaders bool,
	quitFn func(),
	metricsTracker *plugins.MetricsTracker,
) (*APIServer, error) {
	kelpBinPath := kos.GetBinDir().Join(filepath.Base(os.Args[0]))

	optionsMetadata, e := loadOptionsMetadata()
	if e != nil {
		return nil, fmt.Errorf("error while loading options metadata when making APIServer: %s", e)
	}

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
		enableKaas:            enableKaas,
		noHeaders:             noHeaders,
		cachedOptionsMetadata: optionsMetadata,
		quitFn:                quitFn,
		metricsTracker:        metricsTracker,
		kelpErrorsByUser:      map[string]kelpErrorDataForUser{},
		kelpErrorsByUserLock:  &sync.Mutex{},
	}, nil
}

func (s *APIServer) botConfigsPathForUser(userID string) *kelpos.OSPath {
	return s.botConfigsPath.Join(userID)
}

func (s *APIServer) botLogsPathForUser(userID string) *kelpos.OSPath {
	return s.botLogsPath.Join(userID)
}

func (s *APIServer) kelpErrorsForUser(userID string) kelpErrorDataForUser {
	s.kelpErrorsByUserLock.Lock()
	defer s.kelpErrorsByUserLock.Unlock()

	var kefu kelpErrorDataForUser
	if v, ok := s.kelpErrorsByUser[userID]; ok {
		kefu = v
	} else {
		// create new value and insert in map
		kefu = kelpErrorDataForUser{
			errorMap: map[string]KelpError{},
			lock:     &sync.Mutex{},
		}
		s.kelpErrorsByUser[userID] = kefu
	}

	return kefu
}

// InitBackend initializes anything required to get the backend ready to serve
func (s *APIServer) InitBackend() error {
	// do not do an initial load of bots into memory for now since it's based on the user context which we don't have right now
	// and we don't want to do it for all users right now
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

func (s *APIServer) addKelpErrorToMap(userData UserData, ke KelpError) {
	key := ke.UUID

	kefu := s.kelpErrorsForUser(userData.ID)
	// need to use a lock because we could encounter a "concurrent map writes" error against the map which is being updated by multiple threads
	kefu.lock.Lock()
	defer kefu.lock.Unlock()

	kefu.errorMap[key] = ke
}

// removeKelpErrorUserDataIfEmpty removes user error data if the underlying map is empty
func (s *APIServer) removeKelpErrorUserDataIfEmpty(userData UserData) {
	// issue with this is that someone can hold a reference to this object when it is empty
	// and then we remove from parent map and the other thread will add a value, which would result
	// in the object having an entry in the map but being orphaned.
	//
	// We can get creative with timeouts too but that is all an overoptimizationn
	//
	// We could resolve this by always holding both the higher level lock and the per-user lock to modify
	// values inside a user's error map, but that will slow things down
	//
	// for now we do not remove Kelp error user data even if empty.

	// do nothing
}

func (s *APIServer) writeKelpError(userData UserData, w http.ResponseWriter, kerw KelpErrorResponseWrapper) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Printf("writing error: %s\n", kerw.String())
	s.addKelpErrorToMap(userData, kerw.KelpError)

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

func (s *APIServer) runKelpCommandBlocking(userID string, namespace string, cmd string) ([]byte, error) {
	// There is a weird issue on windows where the absolute path for the kelp binary does not work on the release GUI
	// version because of the unzipped directory name but it will work on the released cli version or if we change the
	// name of the folder in which the GUI version is unzipped.
	// To avoid these issues we only invoke with the binary name as opposed to the absolute path that contains the
	// directory name. see start_bot.go for some experimentation with absolute and relative paths
	cmdString := fmt.Sprintf("%s %s", s.kelpBinPath.Unix(), cmd)
	return s.kos.Blocking(userID, namespace, cmdString)
}

func (s *APIServer) runKelpCommandBackground(userID string, namespace string, cmd string) (*kelpos.Process, error) {
	// There is a weird issue on windows where the absolute path for the kelp binary does not work on the release GUI
	// version because of the unzipped directory name but it will work on the released cli version or if we change the
	// name of the folder in which the GUI version is unzipped.
	// To avoid these issues we only invoke with the binary name as opposed to the absolute path that contains the
	// directory name. see start_bot.go for some experimentation with absolute and relative paths
	cmdString := fmt.Sprintf("%s %s", s.kelpBinPath.Unix(), cmd)
	return s.kos.Background(userID, namespace, cmdString)
}

func (s *APIServer) setupOpsDirectory(userID string) error {
	e := s.kos.Mkdir(userID, s.botConfigsPathForUser(userID))
	if e != nil {
		return fmt.Errorf("error setting up configs directory (%s): %s", s.botConfigsPathForUser(userID).Native(), e)
	}

	e = s.kos.Mkdir(userID, s.botLogsPathForUser(userID))
	if e != nil {
		return fmt.Errorf("error setting up logs directory (%s): %s", s.botLogsPathForUser(userID).Native(), e)
	}

	return nil
}
