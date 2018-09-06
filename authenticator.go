package main

import (
	"net/http"
)

type authenticator interface {
	// Authenticate takes a request and provides authentication on top, setting the Username header on the returned request
	Authenticate(w http.ResponseWriter, req *http.Request) (http.Header, error)
}
