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

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/att-comdev/jarvis-connector/gerrit"
)

// gerritChecker run formatting checks against a gerrit server.
type gerritChecker struct {
	server *gerrit.Server

	todo chan *gerrit.PendingChecksInfo
}

// TektonListenerPayload to be recieved by trigger
type TektonListenerPayload struct {
	RepoRoot       string `json:"repoRoot"`
	Project        string `json:"project"`
	ChangeNumber   string `json:"changeNumber"`
	PatchSetNumber int    `json:"patchSetNumber"`
	CheckerUUID    string `json:"checkerUUID"`
}

// checkerScheme is the scheme by which we are registered in the Gerrit server.
const checkerScheme = "jarvis"

// ListCheckers returns all the checkers for our scheme.
func (gc *gerritChecker) ListCheckers() ([]*gerrit.CheckerInfo, error) {
	c, err := gc.server.GetPath("a/plugins/checks/checkers/")
	if err != nil {
		log.Fatalf("ListCheckers: %v", err)
	}

	var out []*gerrit.CheckerInfo
	if err := gerrit.Unmarshal(c, &out); err != nil {
		return nil, err
	}

	filtered := out[:0]
	for _, o := range out {
		if !strings.HasPrefix(o.UUID, checkerScheme+":") {
			continue
		}
		if _, ok := checkerPrefix(o.UUID); !ok {
			continue
		}

		filtered = append(filtered, o)
	}
	return filtered, nil
}

// PostChecker creates or changes a checker. It sets up a checker on
// the given repo, for the given prefix.
func (gc *gerritChecker) PostChecker(repo, prefix string, update bool) (*gerrit.CheckerInfo, error) {
	hash := sha1.New()
	hash.Write([]byte(repo))

	uuid := fmt.Sprintf("%s:%s-%x", checkerScheme, prefix, hash.Sum(nil))
	in := gerrit.CheckerInput{
		UUID:        uuid,
		Name:        prefix,
		Description: "check source code formatting.",
		URL:         "",
		Repository:  repo,
		Status:      "ENABLED",
		Blocking:    []string{},
		Query:       "status:open",
	}

	body, err := json.Marshal(&in)
	if err != nil {
		return nil, err
	}

	path := "a/plugins/checks/checkers/"
	if update {
		path += uuid
	}
	content, err := gc.server.PostPath(path, "application/json", body)
	if err != nil {
		return nil, err
	}

	out := gerrit.CheckerInfo{}
	if err := gerrit.Unmarshal(content, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// checkerPrefix extracts the prefix to check for from a checker UUID.
func checkerPrefix(uuid string) (string, bool) {
	uuid = strings.TrimPrefix(uuid, checkerScheme+":")
	fields := strings.Split(uuid, "-")
	if len(fields) != 2 {
		return "", false
	}
	return fields[0], true
}

// NewGerritChecker creates a server that periodically checks a gerrit
// server for pending checks.
func NewGerritChecker(server *gerrit.Server) (*gerritChecker, error) {
	gc := &gerritChecker{
		server: server,
		todo:   make(chan *gerrit.PendingChecksInfo, 5),
	}

	go gc.pendingLoop()
	return gc, nil
}

// errIrrelevant is a marker error value used for checks that don't apply for a change.
var errIrrelevant = errors.New("irrelevant")

// checkChange checks a (change, patchset) for correct formatting in the given prefix. It returns
// a list of complaints, or the errIrrelevant error if there is nothing to do.
func (c *gerritChecker) checkChange(uuid string, repository string, changeID string, psID int, prefix string) ([]string, string, error) {
	log.Printf("checkChange(%s, %d, %q)", changeID, psID, prefix)

	data := TektonListenerPayload{
		RepoRoot:       GerritURL,
		Project:        repository,
		ChangeNumber:   changeID,
		PatchSetNumber: psID,
		CheckerUUID:    uuid,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	body := bytes.NewReader(payloadBytes)

	log.Printf("body: %s", body)

	req, err := http.NewRequest("POST", EventListenerURL, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jarvis", "create")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var msgs []string
	msgs = append(msgs, fmt.Sprintf("%s", "Job has been submitted to tekton"))
	var details string
	details = ""
	return msgs, details, nil
}

// pendingLoop periodically contacts gerrit to find new checks to
// execute. It should be executed in a goroutine.
func (c *gerritChecker) pendingLoop() {
	for {
		// TODO: real rate limiting.
		time.Sleep(10 * time.Second)

		pending, err := c.server.PendingChecksByScheme(checkerScheme)
		if err != nil {
			log.Printf("PendingChecksByScheme: %v", err)
			continue
		}

		if len(pending) == 0 {
			log.Printf("no pending checks")
		}

		for _, pc := range pending {
			select {
			case c.todo <- pc:
			default:
				log.Println("too busy; dropping pending check.")
			}
		}
	}
}

// Serve runs the serve loop, dispatching for checks that
// need it.
func (gc *gerritChecker) Serve() {
	for p := range gc.todo {
		// TODO: parallelism?.
		if err := gc.executeCheck(p); err != nil {
			log.Printf("executeCheck(%v): %v", p, err)
		}
	}
}

// status encodes the checker states.
type status int

var (
	statusUnset      status = 0
	statusIrrelevant status = 4
	statusRunning    status = 1
	statusFail       status = 2
	statusSuccessful status = 3
)

func (s status) String() string {
	return map[status]string{
		statusUnset:      "UNSET",
		statusIrrelevant: "NOT_RELEVANT",
		statusRunning:    "SCHEDULED",
		statusFail:       "FAILED",
		// remember - success here, simply means we have sucessfully informed the event listener of the job...
		statusSuccessful: "SCHEDULED",
	}[s]
}

// executeCheck executes the pending checks specified in the argument.
func (gc *gerritChecker) executeCheck(pc *gerrit.PendingChecksInfo) error {
	log.Println("checking", pc)

	repository := pc.PatchSet.Repository
	changeID := strconv.Itoa(pc.PatchSet.ChangeNumber)
	psID := pc.PatchSet.PatchSetID
	for uuid := range pc.PendingChecks {
		now := gerrit.Timestamp(time.Now())
		checkInput := gerrit.CheckInput{
			CheckerUUID: uuid,
			State:       statusRunning.String(),
			Message:     "Jarvis about to submit job to tekton",
			Started:     &now,
		}
		log.Printf("posted %s", &checkInput)
		_, err := gc.server.PostCheck(
			changeID, psID, &checkInput)
		if err != nil {
			return err
		}

		var status status
		msg := ""
		url := ""
		lang, ok := checkerPrefix(uuid)
		if !ok {
			return fmt.Errorf("uuid %q had unknown prefix", uuid)
		} else {
			msgs, details, err := gc.checkChange(uuid, repository, changeID, psID, lang)
			if err == errIrrelevant {
				status = statusIrrelevant
			} else if err != nil {
				status = statusFail
				log.Printf("failed in attempt to schedule checkChange(%s, %s, %d, %q): %v", uuid, changeID, psID, lang, err)
			} else if len(msgs) != 0 {
				status = statusSuccessful
			} else {
				status = statusFail
				log.Printf("message empty for checkChange(%s, %s, %d, %q): %v", uuid, changeID, psID, lang, err)
			}
			url = details
			msg = strings.Join(msgs, ", ")
			if len(msg) > 1000 {
				msg = msg[:995] + "..."
			}
		}

		log.Printf("status %s for lang %s on %v", status, lang, pc.PatchSet)
		checkInput = gerrit.CheckInput{
			CheckerUUID: uuid,
			State:       status.String(),
			Message:     msg,
			URL:         url,
			Started:     &gerrit.Timestamp{},
		}
		log.Printf("posted %s", &checkInput)

		if _, err := gc.server.PostCheck(changeID, psID, &checkInput); err != nil {
			return err
		}
	}
	return nil
}
