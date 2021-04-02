package services

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/att-comdev/jarvis-connector/types"
)

var (
	GerritServer        serverService = &ServerImpl{}
	EventListenerServer serverService = &ServerImpl{}
)

type serverService interface {
	GetPath(pathing string, headers []types.Header) ([]byte, error)
	PostPath(pathing string, headers []types.Header, content []byte) ([]byte, error)
	Get(u *url.URL) ([]byte, error)
	Init(url url.URL, authenticator Authenticator, testPath string)
	GetURL() url.URL
	GetRepoRoot() string
}

type ServerImpl struct {
	UserAgent string
	URL       url.URL
	Client    http.Client

	// Issue trace requests.
	Debug bool

	Authenticator Authenticator
}

// Returns value of URL, allowing functions to build requests
func (service *ServerImpl) GetURL() url.URL {
	return service.URL
}

// Returns value of URL as a string
func (service *ServerImpl) GetRepoRoot() string {
	return service.URL.String()
}

// Init assigns necessary values to object and sends preliminary request to destination server
func (service *ServerImpl) Init(url url.URL, authenticator Authenticator, testPath string) {
	service.URL = url
	service.UserAgent = "JarvisConnector"
	service.Authenticator = authenticator

	service.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return nil
	}

	if _, err := service.GetPath(testPath, []types.Header{}); err != nil {
		log.Fatalf("NewServer error: %v \n    url: %v \n    testPath: %v", err, url, testPath)
	}
}

// GetPath runs a GetPath on the given path.
func (service *ServerImpl) GetPath(pathing string, headers []types.Header) ([]byte, error) {
	u := service.URL
	u.Path = path.Join(u.Path, pathing)
	if strings.HasSuffix(pathing, "/") && !strings.HasSuffix(u.Path, "/") {
		// Ugh.
		u.Path += "/"
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	for _, header := range headers {
		req.Header.Set(header.Key, header.Value)
	}

	rep, err := service.do(req)
	if err != nil {
		return nil, err
	}
	if rep.StatusCode/100 != 2 {
		return nil, fmt.Errorf("GetPath %s: status %d", u.String(), rep.StatusCode)
	}

	defer rep.Body.Close()
	return ioutil.ReadAll(rep.Body)
}

// PostPath posts the given data onto a path.
func (service *ServerImpl) PostPath(pathing string, headers []types.Header, content []byte) ([]byte, error) {
	u := service.URL
	u.Path = path.Join(u.Path, pathing)
	if strings.HasSuffix(pathing, "/") && !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}
	for _, header := range headers {
		req.Header.Set(header.Key, header.Value)
	}

	rep, err := service.do(req)
	if err != nil {
		return nil, err
	}
	if rep.StatusCode/100 != 2 {
		return nil, fmt.Errorf("PostPath %s: status %d", u.String(), rep.StatusCode)
	}

	defer rep.Body.Close()
	return ioutil.ReadAll(rep.Body)
}

// Get runs a HTTP GET request on the given URL.
func (service *ServerImpl) Get(u *url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	rep, err := service.do(req)
	if err != nil {
		return nil, err
	}
	if rep.StatusCode/100 != 2 {
		return nil, fmt.Errorf("Get %s: status %d", u.String(), rep.StatusCode)
	}

	defer rep.Body.Close()
	return ioutil.ReadAll(rep.Body)
}

// do runs a HTTP request against the remote server.
func (service *ServerImpl) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", service.UserAgent)
	if service.Authenticator != nil {
		if err := service.Authenticator.Authenticate(req); err != nil {
			return nil, err
		}
	}

	if service.Debug {
		if req.URL.RawQuery != "" {
			req.URL.RawQuery += "&trace=0x1"
		} else {
			req.URL.RawQuery += "trace=0x1"
		}
	}
	return service.Client.Do(req)
}
