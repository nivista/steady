package elastic

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"github.com/nivista/steady/frontend/util"
)

type elastic struct {
	timersIndex, executionsIndex string
	proxy                        *httputil.ReverseProxy
}

// NewElastic returns a new handler that redirects to elatic.
func NewElastic(timersIndex, executionsIndex string, elasticURL *url.URL) http.Handler {
	return elastic{
		timersIndex:     timersIndex,
		executionsIndex: executionsIndex,
		proxy:           httputil.NewSingleHostReverseProxy(elasticURL),
	}
}

func (e elastic) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientID, ok := util.GetClientID(r.Context())
	if !ok {
		fmt.Println("elastic servehttp called without authentication")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var head string
	head, r.URL.Path = util.ShiftPath(r.URL.Path)
	switch head {
	case e.executionsIndex:
		r.URL.Path = path.Join(e.executionsIndex+"-"+clientID) + r.URL.Path
	case e.timersIndex:
		r.URL.Path = path.Join(e.timersIndex+"-"+clientID) + r.URL.Path
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("index must be one of %v or %v.", e.executionsIndex, e.timersIndex)))
		return
	}

	clientID, ok = util.GetClientID(r.Context())
	if !ok {
		fmt.Println("elastic servehttp got context with no clientID.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	e.proxy.ServeHTTP(w, r)
}
