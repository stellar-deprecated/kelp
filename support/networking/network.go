package networking

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cavaliercoder/grab"
)

// JSONRequestDynamicHeaders submits an HTTP web request and parses the response into the responseData object as JSON
func JSONRequestDynamicHeaders(
	httpClient *http.Client,
	method string,
	reqURL string,
	data string,
	headers map[string]HeaderFn,
	responseData interface{}, // the passed in responseData should be a pointer
	errorKey string,
) error {
	headersMap := map[string]string{}
	for header, fn := range headers {
		headersMap[header] = fn(method, reqURL, data)
	}

	return JSONRequest(
		httpClient,
		method,
		reqURL,
		data,
		headersMap,
		responseData,
		errorKey,
	)
}

// JSONRequest submits an HTTP web request and parses the response into the responseData object as JSON
func JSONRequest(
	httpClient *http.Client,
	method string,
	reqURL string,
	data string,
	headers map[string]string,
	responseData interface{}, // the passed in responseData should be a pointer
	errorKey string,
) error {
	// create http request
	req, e := http.NewRequest(method, reqURL, strings.NewReader(data))
	if e != nil {
		return fmt.Errorf("could not create http request: %s", e)
	}

	// add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// execute request
	resp, e := httpClient.Do(req)
	if e != nil {
		return fmt.Errorf("could not execute http request: %s", e)
	}
	defer resp.Body.Close()

	// read response
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return fmt.Errorf("could not read http response: %s", e)
	}
	bodyString := string(body)

	// ensure Content-Type is json
	contentType, _, e := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if e != nil {
		return fmt.Errorf("could not read 'Content-Type' header in http response: %s | response body: %s", e, bodyString)
	}
	if contentType != "application/json" && contentType != "application/hal+json" {
		return fmt.Errorf("invalid 'Content-Type' header in http response ('%s'), expecting 'application/json' or 'application/hal+json', response body: %s", contentType, bodyString)
	}

	if errorKey != "" {
		var errorResponse interface{}
		e = json.Unmarshal(body, &errorResponse)
		if e != nil {
			return fmt.Errorf("could not unmarshall response body to check for an error response: %s | bodyString: %s", e, bodyString)
		}

		switch er := errorResponse.(type) {
		case map[string]interface{}:
			if _, ok := er[errorKey]; ok {
				return fmt.Errorf("error in response, bodyString: %s", bodyString)
			}
		}
	}

	if responseData != nil {
		// parse response, the passed in responseData should be a pointer
		e = json.Unmarshal(body, responseData)
		if e != nil {
			return fmt.Errorf("could not unmarshall response body into json: %s | response body: %s", e, bodyString)
		}
	}

	return nil
}

// DownloadFile downloads a URL to a file on the local disk as it downloads it.
func DownloadFile(url string, filepath string) error {
	outfile, e := os.Create(filepath)
	if e != nil {
		return fmt.Errorf("could not create file at filepath (%s): %s", filepath, e)
	}
	defer outfile.Close()

	resp, e := http.Get(url)
	if e != nil {
		return fmt.Errorf("could not get file at URL (%s): %s", url, e)
	}
	defer resp.Body.Close()

	// do the download and write to file on disk as we download
	_, e = io.Copy(outfile, resp.Body)
	if e != nil {
		return fmt.Errorf("could not download from URL (%s) to file (%s) in a streaming manner: %s", url, filepath, e)
	}
	return nil
}

// DownloadFileWithGrab is a download function that uses the grab third-party library
func DownloadFileWithGrab(
	url string,
	filepath string,
	updateIntervalMillis int,
	statusCodeHandler func(statusCode int, statusString string),
	updateHandler func(completedMB float64, sizeMB float64, speedMBPerSec float64),
	finishHandler func(filename string),
) error {
	// create client
	client := grab.NewClient()
	req, e := grab.NewRequest(filepath, url)
	if e != nil {
		return fmt.Errorf("could not make new grab request: %s", e)
	}
	req.NoResume = true

	// start download
	resp := client.Do(req)
	if resp.HTTPResponse != nil {
		statusCodeHandler(resp.HTTPResponse.StatusCode, resp.HTTPResponse.Status)
	} else {
		statusCodeHandler(-1, "nil resp.HTTPResponse")
	}
	tic := time.Now().UnixNano()

	// start UI loop
	t := time.NewTicker(time.Duration(updateIntervalMillis) * time.Millisecond)
	defer t.Stop()

	invokeUpdateHandler := func() {
		toc := time.Now().UnixNano()
		timeElapsedSec := float64(toc-tic) / float64(time.Second)
		mbCompleted := float64(resp.BytesComplete()) / 1024 / 1024
		speedMBPerSec := float64(mbCompleted) / float64(timeElapsedSec)
		sizeMB := float64(resp.Size()) / 1024 / 2014
		updateHandler(mbCompleted, sizeMB, speedMBPerSec)
	}
Loop:
	for {
		select {
		case <-t.C:
			invokeUpdateHandler()

		case <-resp.Done:
			// download is complete
			invokeUpdateHandler()
			break Loop
		}
	}

	// check for errors
	if e = resp.Err(); e != nil {
		return fmt.Errorf("error while downloading: %s", e)
	}

	finishHandler(resp.Filename)
	return nil
}
