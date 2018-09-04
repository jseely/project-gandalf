package authenticator

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func New() *authenticator {
	return &authenticator{}
}

type authenticator struct{}

func (a *authenticator) Authenticate(req *http.Request) (*http.Request, error) {
	auth := req.Header.Get("Authorization")
	userpass, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(userpass), ":")
	req.Header.Set("Username", parts[0])
	return req, nil
}
