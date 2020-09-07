package app

import (
	"net/http"

	"github.com/nivista/steady/webservice/util"

	"github.com/nivista/steady/webservice/db"
	"github.com/nivista/steady/webservice/queue"

	"github.com/nivista/steady/webservice/app/auth"
	"github.com/nivista/steady/webservice/app/elastic"
)

var clientIDKey int

type app struct {
	db                                       db.Client
	queue                                    queue.Client
	timersIndex, executionsIndex, elasticURL string
}

func NewApp(db db.Client, queue queue.Client, timersIndex, executionsIndex, elasticURL string) http.Handler {
	return app{
		db:              db,
		queue:           queue,
		timersIndex:     timersIndex,
		executionsIndex: executionsIndex,
		elasticURL:      elasticURL,
	}
}

func (a app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = util.ShiftPath(r.URL.Path)
	switch head {
	case "auth":
		auth.NewAuth(a.db).ServeHTTP(w, r)
	case "elastic":
		a.AuthMiddleware(w, r, elastic.NewElastic(a.timersIndex, a.executionsIndex, a.elasticURL))

	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

}

func (a app) AuthMiddleware(w http.ResponseWriter, r *http.Request, handler http.Handler) {
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := a.db.AuthenticateUser(r.Context(), clientID, clientSecret)
	if err != nil {
		switch err.(type) {
		case db.InvalidAPIToken:
			w.WriteHeader(http.StatusUnauthorized)

		case db.InvalidAPISecret:
			w.WriteHeader(http.StatusUnauthorized)

		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}

	reqWithClientID := r.WithContext(util.SetClientID(r.Context(), clientID))
	handler.ServeHTTP(w, reqWithClientID)
}
