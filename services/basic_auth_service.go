package services

import (
	"encoding/base64"
	"net/http"
	"strings"
)

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

type Authenticator interface {
	// Authenticate adds an authentication header to an outgoing request.
	Authenticate(req *http.Request) error
}

// BasicAuth adds the "Basic Authorization" header to an outgoing request.
type BasicAuth struct {
	// Base64 encoded user:secret string.
	EncodedBasicAuth string
}

func (b BasicAuth) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Basic "+b.EncodedBasicAuth)
	return nil
}
