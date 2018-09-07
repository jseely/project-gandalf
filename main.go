package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jseely/project-gandalf/aad/authentication"
)

var auth authenticator

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	header, err := auth.Authenticate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Server error: %v", err)
		return
	}
	if header != nil {
		fmt.Fprintf(w, "UPN: %s\n", header.Get("User-Principal-Name"))
		fmt.Fprintf(w, "Username: %s\n", header.Get("Username"))
		fmt.Fprintf(w, "Shortname: %s\n", header.Get("Shortname"))
		return
	}
	fmt.Fprintf(w, "No reponse")
}

func main() {
	auth = authentication.New(os.Getenv("AZURE_AD_CLIENT_ID"), os.Getenv("AZURE_AD_CLIENT_SECRET"))
	http.Handle("/", http.HandlerFunc(handler))
	log.Fatal(http.ListenAndServe("0.0.0.0:6743", nil))
}
