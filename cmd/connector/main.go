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
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/att-comdev/jarvis-connector/cmd/connector/controllers"
	"github.com/att-comdev/jarvis-connector/services"

	flag "github.com/spf13/pflag"
)

var (
	// GerritURL is the URL of the gerrit instance
	GerritURL string
	// EventListenerURL is the URL of the tekton eventlistener for jarvis
	EventListenerURL string
	register         bool
	update           bool
	list             bool
	authFile         string
	repo             string
	prefix           string
)

func main() {
	// define and parse incoming flags
	flag.StringVar(&GerritURL, "gerrit", "", "URL to gerrit host")
	flag.StringVar(&EventListenerURL, "event_listener", "", "URL of the Tekton EventListener")
	flag.BoolVar(&register, "register", false, "Register the connector with gerrit")
	flag.BoolVar(&update, "update", false, "Update an existing check")
	flag.BoolVar(&list, "list", false, "List pending checks")
	flag.StringVar(&authFile, "auth_file", "", "file containing user:password")
	flag.StringVar(&repo, "repo", "", "the repository (project) name to apply the checker to.")
	flag.StringVar(
		&prefix,
		"prefix",
		"",
		"the prefix that the checker should use for jobs, this is also used as the job name in gerrit.")
	flag.Parse()

	if GerritURL == "" {
		log.Fatal("must set --gerrit")
	}
	gerritURLObj, err := url.Parse(GerritURL)
	if err != nil {
		log.Fatalf("url.Parse: %v", err)
	}

	if authFile == "" {
		log.Fatal("must set --auth_file")
	}
	authFileContent, err := ioutil.ReadFile(authFile)
	if err != nil {
		log.Fatal(err)
	}

	services.GerritServer.Init(*gerritURLObj, services.NewBasicAuth(string(authFileContent)), "a/accounts/self")

	if list { //nolint
		err = controllers.Checker.ListCheckers()

		if err != nil {
			log.Fatalf("ListCheckers: %v", err)
		}

		os.Exit(0)
	} else if register || update {
		if repo == "" {
			log.Fatalf("must set --repo")
		}

		if prefix == "" {
			log.Fatalf("must set --prefix")
		}

		ch, err := controllers.Checker.PostChecker(repo, prefix, update)
		if err != nil {
			log.Fatalf("CreateChecker: %v", err)
		}
		log.Printf("CreateChecker result: %v", ch)

		os.Exit(0)
	} else {
		if EventListenerURL == "" {
			log.Fatal("must set --event_listener")
		}
		eventListenerURLObj, err := url.Parse(EventListenerURL)
		if err != nil {
			log.Fatal(err)
		}

		services.EventListenerServer.Init(*eventListenerURLObj, nil, "/")

		go controllers.Connector.ServeCheck()
		go controllers.Connector.ServeSubmit()
		controllers.Connector.PendingLoop()
	}
}
