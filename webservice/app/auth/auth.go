package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"

	"github.com/nivista/steady/webservice/db"
	"github.com/nivista/steady/webservice/util"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	goauth "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

var config = oauth2.Config{
	ClientID:     "1097208656533-2dks16lohh6hc2567ttpotbk63q0uhsq.apps.googleusercontent.com",
	ClientSecret: "EMBaIOABTgXahL2h-EQi6SB7",
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://127.0.0.1:8080/auth/apikey/callback",
	Scopes:       []string{"openid"},
}

type authHandler struct {
	db db.Client
}

func NewAuth(db db.Client) http.Handler {

	return authHandler{
		db: db,
	}
}

func (a authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var head string
	head, r.URL.Path = util.ShiftPath(r.URL.Path)

	switch head {
	case "apikey":
		head, r.URL.Path = util.ShiftPath(r.URL.Path)
		switch head {
		case "":
			a.getAPIKey(w, r)
		case "callback":
			a.getAPIKeyCallback(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// first step in getting API key
// redirects to google oauthHandler
func (a authHandler) getAPIKey(res http.ResponseWriter, req *http.Request) {
	url := config.AuthCodeURL("state")
	http.Redirect(res, req, url, http.StatusFound)
}

// callback with authorization grant
// we then go to google, verify it's legit, do an update to the user and give the ID.
func (a authHandler) getAPIKeyCallback(res http.ResponseWriter, req *http.Request) {
	token, err := config.Exchange(req.Context(), req.URL.Query().Get("code"))
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(err.Error()))
		return
	}

	service, err := goauth.NewService(req.Context(), option.WithTokenSource(config.TokenSource(req.Context(), token)))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}

	userInfoRes, err := service.Userinfo.Get().Do()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}

	apiKeyInt, err := rand.Int(rand.Reader, big.NewInt(1<<31))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}
	apiKeyString := apiKeyInt.String()
	apiKeyBytes := []byte(apiKeyString)

	hashedApiKey, err := bcrypt.GenerateFromPassword(apiKeyBytes, bcrypt.DefaultCost)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}

	err = a.db.UpsertUser(req.Context(), userInfoRes.Id, string(hashedApiKey))
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
		return
	}

	_, err = res.Write([]byte(fmt.Sprintf("API Token: %v\nAPI Secret: %v\n", userInfoRes.Id, apiKeyString)))

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(err.Error()))
	}
}
