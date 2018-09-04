package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/sessions"

	_ "golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const (
	redirectURI string = "http://104.210.55.174:6743/callback"
)

var sessionStoreKeyPairs = [][]byte{
	[]byte("something-very-secret"),
	nil,
}

var store sessions.Store

var (
	clientID string
	config   *oauth2.Config
)

type User struct {
	Email       string
	DisplayName string
}

func init() {
	fsStore := sessions.NewFilesystemStore("", sessionStoreKeyPairs...)
	fsStore.MaxLength(0)
	store = fsStore

	gob.Register(&User{})
	gob.Register(&oauth2.Token{})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	clientID = os.Getenv("AZURE_AD_CLIENT_ID")
	if clientID == "" {
		log.Fatal("AZURE_AD_CLIENT_ID must be set.")
	}

	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: "", // no client secret
		RedirectURL:  redirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/token",
		},
		Scopes: []string{"User.Read", "Group.Read.All"},
	}

	http.Handle("/", handle(IndexHandler))
	http.Handle("/callback", handle(CallbackHandler))

	log.Fatal(http.ListenAndServe("0.0.0.0:6743", nil))
}

type handle func(w http.ResponseWriter, req *http.Request) error

func (h handle) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Handler panic: %v", r)
		}
	}()
	if err := h(w, req); err != nil {
		log.Printf("Handler error: %v", err)

		if httpErr, ok := err.(Error); ok {
			http.Error(w, httpErr.Message, httpErr.Code)
		}
	}
}

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	if e.Message == "" {
		e.Message = http.StatusText(e.Code)
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func IndexHandler(w http.ResponseWriter, req *http.Request) error {
	session, _ := store.Get(req, "session")

	var token *oauth2.Token
	if req.FormValue("logout") != "" {
		session.Values["token"] = nil
		sessions.Save(req, w)
	} else {
		if v, ok := session.Values["token"]; ok {
			token = v.(*oauth2.Token)
		}
	}

	var data = struct {
		Token   *oauth2.Token
		AuthURL string
	}{
		Token:   token,
		AuthURL: config.AuthCodeURL(SessionState(session), oauth2.AccessTypeOnline),
	}

	if data.Token == nil {
		fmt.Fprintf(w, "<a href=\"%s\">Login</a>", data.AuthURL)
	} else {
		fmt.Println(data.Token.AccessToken)
		graphRequest, err := http.NewRequest(http.MethodPost, "https://graph.windows.net/72f988bf-86f1-41af-91ab-2d7cd011db47/users/e6962d84-f613-4015-8498-999a23226bbd/getMemberGroups?api-version=1.6", strings.NewReader("{\"securityEnabledOnly\":false}"))
		if err != nil {
			return fmt.Errorf("Error creating graph request: %v", err)
		}
		graphRequest.Header.Set("Content-Type", "application/json")
		graphRequest.Header.Set("Authorization", "Bearer "+data.Token.AccessToken)
		resp, err := http.DefaultClient.Do(graphRequest)
		if err != nil {
			return fmt.Errorf("Error getting graph data: %v", err)
		}
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			return fmt.Errorf("Error writing response: %v", err)
		}
	}

	return nil
}

func CallbackHandler(w http.ResponseWriter, req *http.Request) error {
	session, _ := store.Get(req, "session")

	if req.FormValue("state") != SessionState(session) {
		return Error{http.StatusBadRequest, "invalid callback state"}
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", clientID)
	form.Set("code", req.FormValue("code"))
	form.Set("redirect_uri", redirectURI)
	form.Set("resource", "https://graph.windows.net")
	form.Set("client_secret", os.Getenv("AZURE_AD_CLIENT_SECRET"))

	tokenReq, err := http.NewRequest(http.MethodPost, config.Endpoint.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("error creating token request: %v", err)
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		return fmt.Errorf("error performing token request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("token response was %s", resp.Status)
	}

	var token oauth2.Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return fmt.Errorf("error decoding JSON response: %v", err)
	}

	session.Values["token"] = &token
	if err := sessions.Save(req, w); err != nil {
		return fmt.Errorf("error saving session: %v", err)
	}

	http.Redirect(w, req, "/", http.StatusFound)
	return nil
}

func SessionState(session *sessions.Session) string {
	return base64.StdEncoding.EncodeToString(sha256.New().Sum([]byte(session.ID)))
}

func dump(v interface{}) {
	spew.Dump(v)
}
