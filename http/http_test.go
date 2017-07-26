package http

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func httpGet(url string) (string, int, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", 0, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", 0, err
	}
	return string(data), res.StatusCode, nil
}

func httpRequest(method, url, body string) (string, int, error) {
	r, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", 0, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", 0, err
	}
	return string(data), res.StatusCode, nil
}

func httpPost(url, body string) (string, int, error) {
	return httpRequest(http.MethodPost, url, body)
}

func httpDelete(url, body string) (string, int, error) {
	return httpRequest(http.MethodDelete, url, body)
}

func testServer() (*httptest.Server, string) {
	file, err := ioutil.TempFile("", "wakeonlan")
	if err != nil {
		panic(err)
	}
	api := Server{
		wakeFunc:  func(net.IP, net.HardwareAddr) error { return nil },
		cacheFile: file.Name(),
	}
	log.SetOutput(ioutil.Discard)
	return httptest.NewServer(api.Handler()), file.Name()
}

func TestRequests(t *testing.T) {
	server, cacheFile := testServer()
	defer os.Remove(cacheFile)
	defer server.Close()

	var tests = []struct {
		method   string
		body     string
		url      string
		response string
		status   int
	}{
		// Unknown resources
		{"GET", "", "/not-found", "404 page not found\n", 404},
		{"GET", "", "/api/not-found", `{"status":404,"message":"Resource not found"}`, 404},
		// Invalid JSON
		{"POST", "", "/api/v1/wake", `{"status":400,"message":"Malformed JSON"}`, 400},
		// Invalid MAC address
		{"POST", `{"macAddress":"foo"}`, "/api/v1/wake", `{"status":400,"message":"Invalid MAC address: foo"}`, 400},
		// List devices
		{"GET", "", "/api/v1/wake", `{"devices":[]}`, 200},
		// Wake device
		{"POST", `{"macAddress":"AB:CD:EF:12:34:56"}`, "/api/v1/wake", "", 204},
		{"GET", "", "/api/v1/wake", `{"devices":[{"macAddress":"AB:CD:EF:12:34:56"}]}`, 200},
		// Waking same device does not result in duplicates
		{"POST", `{"macAddress":"AB:CD:EF:12:34:56"}`, "/api/v1/wake", "", 204},
		{"GET", "", "/api/v1/wake", `{"devices":[{"macAddress":"AB:CD:EF:12:34:56"}]}`, 200},
		// Delete
		{"DELETE", `{"macAddress":"AB:CD:EF:12:34:56"}`, "/api/v1/wake", "", 204},
		{"GET", "", "/api/v1/wake", `{"devices":[]}`, 200},
		// Add multiple devices
		{"POST", `{"macAddress":"AB:CD:EF:12:34:56"}`, "/api/v1/wake", "", 204},
		{"POST", `{"macAddress":"12:34:56:AB:CD:EF"}`, "/api/v1/wake", "", 204},
		{"GET", "", "/api/v1/wake", `{"devices":[{"macAddress":"12:34:56:AB:CD:EF"},{"macAddress":"AB:CD:EF:12:34:56"}]}`, 200},
		{"DELETE", `{"macAddress":"AB:CD:EF:12:34:56"}`, "/api/v1/wake", "", 204},
		{"DELETE", `{"macAddress":"12:34:56:AB:CD:EF"}`, "/api/v1/wake", "", 204},
		// Add device with name
		{"POST", `{"name":"foo","macAddress":"AB:CD:EF:12:34:56"}`, "/api/v1/wake", "", 204},
		{"GET", "", "/api/v1/wake", `{"devices":[{"name":"foo","macAddress":"AB:CD:EF:12:34:56"}]}`, 200},
	}

	for _, tt := range tests {
		var (
			data   string
			status int
			err    error
		)
		switch tt.method {
		case http.MethodGet:
			data, status, err = httpGet(server.URL + tt.url)
		case http.MethodPost:
			data, status, err = httpPost(server.URL+tt.url, tt.body)
		case http.MethodDelete:
			data, status, err = httpDelete(server.URL+tt.url, tt.body)
		default:
			t.Fatal("invalid method: " + tt.method)
		}
		if err != nil {
			t.Fatal(err)
		}
		if got := status; status != tt.status {
			t.Errorf("want status %d for %q, got %d", tt.status, tt.url, got)
		}
		if got := string(data); got != tt.response {
			t.Errorf("want response %q for %s, got %q", tt.response, tt.url, got)
		}
	}
}
