package services_test

import (
	"net/http"
	"testing"

	"github.com/att-comdev/jarvis-connector/services"
)


func TestNewBasicAuth(t *testing.T) {
	credentials := "Jarvis:Landry"
	paddedCredentials := " Jarvis:Landry "
	b64EncodedCredentials := "SmFydmlzOkxhbmRyeQ=="
	emptyCredentials := ""
	test1 := services.NewBasicAuth(credentials)
	test2 := services.NewBasicAuth(paddedCredentials)
	test3 := services.NewBasicAuth(emptyCredentials)

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

func TestAuthenticate(t *testing.T) {
	test1 := services.BasicAuth{
		EncodedBasicAuth: "SmFydmlzOkxhbmRyeQ==",
	}
	request, err := http.NewRequest(http.MethodGet, "gerrit.com", nil)
	if err != nil {
		t.Errorf("Received error setting up TestAuthenticate: %v", err)
	}

	err = test1.Authenticate(request)

	if err != nil { //nolint
		t.Errorf("Received error from Authenicate function: %v", err)
	} else if request.Header.Get("Authorization") == "" {
		t.Errorf("Authorization header not found")
	} else if request.Header.Get("Authorization") != "Basic SmFydmlzOkxhbmRyeQ==" {
		t.Errorf("Authorization header was %s, expected: Basic SmFydmlzOkxhbmRyeQ==",
			request.Header.Get("Authorization"))
	}
}
