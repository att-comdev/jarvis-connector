// Copyright 2019 Google Inc. All rights reserved.
// Copyright 2021 AT&T Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gerrit

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	JarvisMergeHashtag string = "jarvis-merge"
)

// Server represents a single Gerrit host.
type Server struct {
	UserAgent string
	URL       url.URL
	Client    http.Client

	// Issue trace requests.
	Debug bool

	Authenticator Authenticator
}

type Authenticator interface {
	// Authenticate adds an authentication header to an outgoing request.
	Authenticate(req *http.Request) error
}

// BasicAuth adds the "Basic Authorization" header to an outgoing request.
type BasicAuth struct {
	// Base64 encoded user:secret string.
	EncodedBasicAuth string
}

// NewBasicAuth creates a BasicAuth authenticator. |who| should be a
// "user:secret" string.
func NewBasicAuth(who string) *BasicAuth {
	auth := strings.TrimSpace(who)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(auth)))
	base64.StdEncoding.Encode(encoded, []byte(auth))
	return &BasicAuth{
		EncodedBasicAuth: string(encoded),
	}
}

func (b *BasicAuth) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Basic "+string(b.EncodedBasicAuth))
	return nil
}

// New creates a Gerrit Server for the given URL.
func New(u url.URL) *Server {
	g := &Server{
		URL: u,
	}

	g.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return nil
	}

	return g
}

// GetPath runs a Get on the given path.
func (g *Server) GetPath(p string) ([]byte, error) {
	u := g.URL
	u.Path = path.Join(u.Path, p)
	if strings.HasSuffix(p, "/") && !strings.HasSuffix(u.Path, "/") {
		// Ugh.
		u.Path += "/"
	}
	return g.Get(&u)
}

// Do runs a HTTP request against the remote server.
func (g *Server) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", g.UserAgent)
	if g.Authenticator != nil {
		if err := g.Authenticator.Authenticate(req); err != nil {
			return nil, err
		}
	}

	if g.Debug {
		if req.URL.RawQuery != "" {
			req.URL.RawQuery += "&trace=0x1"
		} else {
			req.URL.RawQuery += "trace=0x1"
		}
	}
	return g.Client.Do(req)
}

// Get runs a HTTP GET request on the given URL.
func (g *Server) Get(u *url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	rep, err := g.Do(req)
	if err != nil {
		return nil, err
	}
	if rep.StatusCode/100 != 2 {
		return nil, fmt.Errorf("Get %s: status %d", u.String(), rep.StatusCode)
	}

	defer rep.Body.Close()
	return ioutil.ReadAll(rep.Body)
}

// PostPath posts the given data onto a path.
func (g *Server) PostPath(p string, contentType string, content []byte) ([]byte, error) {
	u := g.URL
	u.Path = path.Join(u.Path, p)
	if strings.HasSuffix(p, "/") && !strings.HasSuffix(u.Path, "/") {
		// Ugh.
		u.Path += "/"
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	rep, err := g.Do(req)
	if err != nil {
		return nil, err
	}
	if rep.StatusCode/100 != 2 {
		return nil, fmt.Errorf("Post %s: status %d", u.String(), rep.StatusCode)
	}

	defer rep.Body.Close()
	return ioutil.ReadAll(rep.Body)
}

func (s *Server) PendingChecksByScheme(scheme string) ([]*PendingChecksInfo, error) {
	u := s.URL

	// The trailing '/' handling is really annoying.
	u.Path = path.Join(u.Path, "a/plugins/checks/checks.pending/") + "/"

	q := "scheme:" + scheme
	u.RawQuery = "query=" + q
	content, err := s.Get(&u)
	if err != nil {
		return nil, err
	}

	var out []*PendingChecksInfo
	if err := Unmarshal(content, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// PendingChecks returns the checks pending for the given checker.
func (s *Server) PendingChecks(checkerUUID string) ([]*PendingChecksInfo, error) {
	u := s.URL

	// The trailing '/' handling is really annoying.
	u.Path = path.Join(u.Path, "a/plugins/checks/checks.pending/") + "/"

	q := "checker:" + checkerUUID
	u.RawQuery = "query=" + url.QueryEscape(q)

	content, err := s.Get(&u)
	if err != nil {
		return nil, err
	}

	var out []*PendingChecksInfo
	if err := Unmarshal(content, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// PostCheck posts a single check result onto a change.
func (s *Server) PostCheck(changeID string, psID int, input *CheckInput) (*CheckInfo, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	res, err := s.PostPath(fmt.Sprintf("a/changes/%s/revisions/%d/checks/", changeID, psID),
		"application/json", body)
	if err != nil {
		return nil, err
	}

	var out CheckInfo
	if err := Unmarshal(res, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (s *Server) HandleSubmissions() error {
	u := s.URL

	u.Path = path.Join(u.Path, "a/changes/?o=SUBMITTABLE&o=ALL_REVISIONS&q=is:open")
	content, err := s.Get(&u)
	if err != nil {
		return err
	}

	var out []*PendingSubmitInfo
	if err := Unmarshal(content, &out); err != nil {
		return err
	}

	var patchsets []*PendingSubmitInfo
	for i := 0; i < len(out); i++ {
		// Ignore merge conflicts, patchsets without required labels, and patchsets currently being handled by Jarvis
		if out[i].Mergeable == true && out[i].Subittable == true && !contains(out[i].Hashtags, JarvisMergeHashtag) {
			patchsets = append(patchsets, out[i])
		}
	}

	for _, obj := range patchsets {

		// Find the patch set number by checking the number of revisions that have been make to the change
		obj.PatchsetNumber = len(obj.Revisions)

		if err = s.PostHashtag(obj); err != nil {
			return err
		}

		if err = s.CallPipeline(obj); err != nil {
			return err
		}
	}

	return nil
}

type PendingSubmitInfo struct {
	Id              string    `json:"id"`
	Project         string    `json:"project"`
	Branch          string    `json:"branch"`
	Hashtags        []string  `json:"hashtags"`
	ChangeId        string    `json:"change_id"`
	Subject         string    `json:"subject"`
	Status          string    `json:"status"`
	Created         Timestamp `json:"created"`
	Updated         Timestamp `json:"updated"`
	SubmitType      string    `json:"submit_type"`
	Mergeable       bool      `json:"mergeable"`
	Subittable      bool      `json:"submittable"`
	PatchsetNumber  int
	CurrentRevision string                 `json:"current_revision"`
	Revisions       map[string]interface{} `json:"revisions"`
}

type HashtagPayload struct {
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

type TektonMergePayload struct {
	RepoRoot       string `json:"repoRoot"`
	Project        string `json:"project"`
	ChangeNumber   string `json:"changeNumber"`
	PatchSetNumber string `json:"patchSetNumber"`
}

func (s *Server) PostHashtag(patchset *PendingSubmitInfo) error {
	u := s.URL
	u.Path = path.Join(u.Path, fmt.Sprintf("a/changes/%s/hashtags/", patchset.ChangeId)) + "/"
	hashtagPayload := HashtagPayload{
		Add:    []string{JarvisMergeHashtag},
		Remove: []string{},
	}
	body, err := json.Marshal(hashtagPayload)
	if err != nil {
		return err
	}
	_, err = s.PostPath(u.Path, "application/json", body)
	if err != nil {
		// return err
		log.Printf("Error: %v", err)
		log.Printf("If you are seeing this, the PostHashtag function isn't working")
	}

	return nil
}

func (s *Server) CallPipeline(patchset *PendingSubmitInfo) error {
	EventListenerURL := "https://gerrit.jarvis.local/"
	data := TektonMergePayload{
		RepoRoot:       "http://gerrit.jarvis.local/",
		Project:        patchset.Project,
		ChangeNumber:   patchset.ChangeId,
		PatchSetNumber: strconv.Itoa(patchset.PatchsetNumber),
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", EventListenerURL, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jarvis", "merge")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	return nil
}

func contains(list []string, element string) bool {
	for _, obj := range list {
		if obj == element {
			return true
		}
	}

	return false
}
