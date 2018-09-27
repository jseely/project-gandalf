package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jseely/project-gandalf/aad/authentication"
	"github.com/jseely/project-gandalf/microsoft/groupmembership"
)

var (
	auth authenticator
	gm   GroupMembership
)

func updateHeader(h http.Header, req *http.Request) *http.Request {
	for key, value := range h {
		req.Header[key] = value
	}
	return req
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	header, err := auth.Authenticate(w, req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error authenticating request: %v", err)
		return
	}
	if header == nil {
		// If header is nil, auth already wrote a redirect to the ResponseWriter
		return
	}

	updateHeader(header, req)

	header, err = gm.GetGroupMembership(w, req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error getting group membership: %v", err)
		return
	}
	if header == nil {
		// If header is nil, groupmembership already wrote a redirect to the ResponseWriter
		return
	}

	updateHeader(header, req)

	url := os.Getenv("REDIRECT_URL") + req.URL.String()
	proxyReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating new proxy request: %v", err)
		return
	}
	proxyReq.Header = req.Header

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	//Update ResponseWriter
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	auth = authentication.New(os.Getenv("AZURE_AD_CLIENT_ID"), os.Getenv("AZURE_AD_CLIENT_SECRET"), os.Getenv("HOST_ADDR"), os.Getenv("COOKIE_SECRET"))
	gm = groupmembership.New()
	http.HandleFunc("/", handleRequest)

	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDR"), nil))
}
