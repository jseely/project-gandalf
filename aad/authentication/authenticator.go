package authentication

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"

	"golang.org/x/oauth2"
)

type User struct {
	Email       string
	DisplayName string
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

func init() {
	gob.Register(&User{})
	gob.Register(&oauth2.Token{})
}

func New(clientId, clientSecret, hostaddr, cookieSecret string) *authenticator {
	fsStore := sessions.NewFilesystemStore("", []byte(cookieSecret))
	fsStore.MaxLength(0)
	redirectURI := fmt.Sprintf("%saad/callback", hostaddr)
	return &authenticator{
		store:        fsStore,
		clientID:     clientId,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		config: &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/common/oauth2/authorize",
				TokenURL: "https://login.microsoftonline.com/common/oauth2/token",
			},
			Scopes: []string{"User.Read", "Group.Read.All"},
		},
	}
}

type authenticator struct {
	store        sessions.Store
	config       *oauth2.Config
	clientID     string
	clientSecret string
	redirectURI  string
}

func (a *authenticator) Authenticate(w http.ResponseWriter, req *http.Request) (http.Header, error) {
	if req.URL.Path == "/aad/callback" {
		return a.HandleAADCallback(w, req)
	}

	session, _ := a.store.Get(req, "session")

	var token string
	if req.FormValue("logout") == "" {
		if _, ok := req.Header["Token"]; ok {
			token = req.Header.Get("Token")
			session.Values["token"] = token
			session.Save(req, w)
		}
		if v, ok := session.Values["token"]; ok {
			token = v.(string)
		}
	} else {
		session.Values["token"] = ""
		session.Save(req, w)
	}

	if token == "" {
		session.Values["redirect_path"] = req.URL.Path
		session.Save(req, w)
		redirectPath := a.config.AuthCodeURL(SessionState(session), oauth2.AccessTypeOnline)
		http.Redirect(w, req, redirectPath, http.StatusFound)
		return nil, nil
	}
	parts := strings.Split(token, ".")
	var claim map[string]interface{}
	var err error
	for len(parts[1])%4 != 0 {
		parts[1] += "="
	}
	b, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling base64 encoded claim: %v", err)
	}
	fmt.Println(string(b))
	err = json.Unmarshal(b, &claim)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling json claim: %v", err)
	}
	header := http.Header{}
	header.Set("Token", token)
	header.Set("User-Principal-Name", claim["upn"].(string))
	header.Set("Username", claim["name"].(string))
	header.Set("Shortname", claim["given_name"].(string))
	return header, nil
}

func (a *authenticator) HandleAADCallback(w http.ResponseWriter, req *http.Request) (http.Header, error) {
	session, _ := a.store.Get(req, "session")
	if req.FormValue("state") != SessionState(session) {
		return nil, Error{http.StatusBadRequest, "invalid callback state"}
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", a.clientID)
	form.Set("code", req.FormValue("code"))
	form.Set("redirect_uri", a.redirectURI)
	form.Set("resource", "https://graph.windows.net")
	form.Set("client_secret", a.clientSecret)

	tokenReq, err := http.NewRequest(http.MethodPost, a.config.Endpoint.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating token request: %v", err)
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("error performing token request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("token response was %s", resp.Status)
	}

	var token oauth2.Token
	if err = json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %v", err)
	}

	session.Values["token"] = token.AccessToken
	if err = sessions.Save(req, w); err != nil {
		return nil, fmt.Errorf("error saving session: %v", err)
	}

	redirectPath, _ := session.Values["redirect_path"].(string)
	http.Redirect(w, req, redirectPath, http.StatusFound)
	return nil, nil
}

func SessionState(session *sessions.Session) string {
	return base64.StdEncoding.EncodeToString(sha256.New().Sum([]byte(session.ID)))
}
