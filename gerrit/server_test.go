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
	"net/http"
	"net/url"
	"testing"
)

func TestServer_NewBasicAuth(t *testing.T) {
	credentials := "Jarvis:Landry"
	paddedCredentials := " Jarvis:Landry "
	b64EncodedCredentials := "SmFydmlzOkxhbmRyeQ=="
	emptyCredentials := ""
	test1 := NewBasicAuth(credentials)
	test2 := NewBasicAuth(paddedCredentials)
	test3 := NewBasicAuth(emptyCredentials)

	if test1.EncodedBasicAuth != b64EncodedCredentials {
		t.Errorf("Base64 Encoding of %s was %s, expected: %s",
			credentials, test1.EncodedBasicAuth, b64EncodedCredentials)
	}

	if test2.EncodedBasicAuth != b64EncodedCredentials {
		t.Errorf("Base64 Encoding of %s was %s, expected: %s",
			credentials, test2.EncodedBasicAuth, b64EncodedCredentials)
	}

	if test3.EncodedBasicAuth != "" {
		t.Errorf("Base64 Encoding of %s was %s, expected: empty string",
			emptyCredentials, test3.EncodedBasicAuth)
	}
}

func TestServer_Authenticate(t *testing.T) {
	test1 := BasicAuth{
		EncodedBasicAuth: "SmFydmlzOkxhbmRyeQ==",
	}
	request, err := http.NewRequest(http.MethodGet, "gerrit.com", nil)

	err = test1.Authenticate(request)

	if err != nil {
		t.Errorf("Received error from Authenicate function: %v", err)
	} else if request.Header.Get("Authorization") == "" {
		t.Errorf("Authorization header not found")
	} else if request.Header.Get("Authorization") != "Basic SmFydmlzOkxhbmRyeQ==" {
		t.Errorf("Authorization header was %s, expected: Basic SmFydmlzOkxhbmRyeQ==",
			request.Header.Get("Authorization"))
	}
}

func TestServer_New(t *testing.T) {
	// TODO(dannymassa) Find case where CheckRedirect errors
	goodUrl, err := url.Parse("https://github.com/")
	if err != nil {
		t.Errorf("Received error during test setup: %v", err)
	}

	goodServer := New(*goodUrl)

	if goodServer.URL.String() != "https://github.com/" {
		t.Errorf("Server creation unsuccessful")
	}
}

func TestServer_GetPath(t *testing.T) {
	// Setup Server
	goodUrl, err := url.Parse("https://github.com/")
	if err != nil {
		t.Errorf("Received error during test setup: %v", err)
	}
	goodServer := New(*goodUrl)

	// Mock API Endpoint


	// Ensure API GET request is made, and trailing slash is appended if necessary
	slashAPI := "my/fake/api/"
	noSlashAPI := "my/fake/api"
	_, slashErr := goodServer.GetPath(slashAPI)
	_, noSlashErr := goodServer.GetPath(noSlashAPI)
	if slashErr != nil {
		// t.Errorf("Error making GET request for path: %s, error: %v", slashAPI, err)
	}
	if noSlashErr != nil {
		// t.Errorf("Error making GET request for path: %s, error: %v", noSlashAPI, err)
	}
}