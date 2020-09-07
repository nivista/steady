package elastic

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"github.com/nivista/steady/webservice/util"
)

type elastic struct {
	elasticURL, timersIndex, executionsIndex string
}

func NewElastic(timersIndex, executionsIndex, elasticURL string) http.Handler {
	return elastic{
		timersIndex:     timersIndex,
		executionsIndex: executionsIndex,
		elasticURL:      elasticURL,
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

	addr, err := url.Parse("http://127.0.0.1:9200")
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(addr)

	clientID, ok = util.GetClientID(r.Context())
	if !ok {
		fmt.Println("elastic servehttp got context with no clientID.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	proxy.ServeHTTP(w, r)
}
