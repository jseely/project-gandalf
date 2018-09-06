package authentication

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func New() *authenticator {
	return &authenticator{}
}

type authenticator struct{}

func (a *authenticator) Authenticate(w http.ResponseWriter, req *http.Request) (http.Header, error) {
	auth := req.Header.Get("Authorization")
	userpass, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(userpass), ":")
	header := http.Header{}
	header.Set("User-Principal-Name", parts[0])
	header.Set("Username", parts[0])
	header.Set("Shortname", parts[0])
	return header, nil
}
