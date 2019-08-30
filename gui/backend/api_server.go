package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/kelp/support/kelpos"
)

// APIServer is an instance of the API service
type APIServer struct {
	dirPath               string
	binPath               string
	configsDir            string
	logsDir               string
	kos                   *kelpos.KelpOS
	horizonTestnetURI     string
	horizonPubnetURI      string
	ccxtRestUrl           string
	apiTestNet            *horizonclient.Client
	apiPubNet             *horizonclient.Client
	cachedOptionsMetadata metadata
}

// MakeAPIServer is a factory method
func MakeAPIServer(kos *kelpos.KelpOS, horizonTestnetURI string, horizonPubnetURI string, ccxtRestUrl string) (*APIServer, error) {
	binPath, e := filepath.Abs(os.Args[0])
	if e != nil {
		return nil, fmt.Errorf("could not get binPath of currently running binary: %s", e)
	}

	dirPath := filepath.Dir(binPath)
	configsDir := dirPath + "/ops/configs"
	logsDir := dirPath + "/ops/logs"

	horizonTestnetURI = strings.TrimSuffix(horizonTestnetURI, "/")
	horizonPubnetURI = strings.TrimSuffix(horizonPubnetURI, "/")
	log.Printf("using horizonTestnetURI: %s\n", horizonTestnetURI)
	log.Printf("using horizonPubnetURI: %s\n", horizonPubnetURI)
	log.Printf("using ccxtRestUrl: %s\n", ccxtRestUrl)
	apiTestNet := &horizonclient.Client{
		HorizonURL: horizonTestnetURI,
		HTTP:       http.DefaultClient,
	}
	apiPubNet := &horizonclient.Client{
		HorizonURL: horizonPubnetURI,
		HTTP:       http.DefaultClient,
	}

	optionsMetadata, e := loadOptionsMetadata()
	if e != nil {
		return nil, fmt.Errorf("error while loading options metadata when making APIServer: %s", e)
	}

	return &APIServer{
		dirPath:               dirPath,
		binPath:               binPath,
		configsDir:            configsDir,
		logsDir:               logsDir,
		kos:                   kos,
		horizonTestnetURI:     horizonTestnetURI,
		horizonPubnetURI:      horizonPubnetURI,
		ccxtRestUrl:           ccxtRestUrl,
		apiTestNet:            apiTestNet,
		apiPubNet:             apiPubNet,
		cachedOptionsMetadata: optionsMetadata,
	}, nil
}

func (s *APIServer) parseBotName(r *http.Request) (string, error) {
	botNameBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		return "", fmt.Errorf("error when reading request input: %s\n", e)
	}
	return string(botNameBytes), nil
}

func (s *APIServer) writeError(w http.ResponseWriter, message string) {
	log.Print(message)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(message))
}

// ErrorResponse represents an error
type ErrorResponse struct {
	Error string `json:"error"`
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
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	return s.kos.Blocking(namespace, cmdString)
}

func (s *APIServer) runKelpCommandBackground(namespace string, cmd string) (*kelpos.Process, error) {
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	return s.kos.Background(namespace, cmdString)
}
