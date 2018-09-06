package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	ba "github.com/jseely/project-gandalf/basic/authentication"
	bgm "github.com/jseely/project-gandalf/basic/groupmembership"
)

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

func updateHeader(h http.Header, req *http.Request) *http.Request {
	for key, values := range h {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return req
}

func main() {
	auth := ba.New()
	gm := bgm.New()
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		println("Authenticate")
		header, err := auth.Authenticate(w, req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error authenticating request: %v", err)
			return
		}
		if header == nil {
			//redirect request to the ResponseWriter
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
			//redirect request to the ResponseWriter
			return
		}

		updateHeader(header, req)

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		proxyReq, err := http.NewRequest(req.Method, "http://localhost:12345", bytes.NewReader(body))

		proxyReq.Header = cloneHeader(req.Header)

		resp, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		//Update ResponseWriter
		w.WriteHeader(resp.StatusCode)
		w.Write([]byte(resp.Status))
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(responseData))

	})

	log.Fatal(http.ListenAndServe("0.0.0.0:6701", nil))

}
