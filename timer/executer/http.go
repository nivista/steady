package executer

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/nivista/steady/.gen/protos/common"
)

const maxSize = 10e5

func executeHTTP(config *common.Task_Http) []byte {
	var res *http.Response
	var err error
	switch config.Http.Method {
	case common.Method_GET:
		res, err = http.Get(config.Http.Url)

	case common.Method_POST:
		return getError("post not implemented")
	default:
		return getError("unknown http method")
	}

	if err != nil {
		return getError(err.Error())
	}

	var result httpResponse
	result.Proto = res.Proto
	result.StatusCode = res.StatusCode
	result.Headers = res.Header

	if config.Http.SaveResponseBody {
		var data bytes.Buffer

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			result.Error = err.Error()
		} else if len(body) > maxSize {
			result.Error = "body too big to write"
		} else {
			data.Write(body)
			result.Body = data.String()
		}
	}

	res.Body.Close()

	json, _ := json.Marshal(result)
	return json
}

type httpResponse struct {
	StatusCode         int
	Error, Proto, Body string
	Headers            map[string][]string
}
