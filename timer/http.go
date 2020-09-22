package timer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/nivista/steady/.gen/protos/common"
	"github.com/spf13/viper"
)

// upper limit for the size of a reponse body, past this point it will be dropped
var (
	maxRequestBodySizeKey  string = "STEADY_HTTP_MAX_REQUEST_BODY_SIZE"
	maxResponseBodySizeKey string = "STEADY_HTTP_MAX_RESPONSE_BODY_SIZE"

	maxRequestBodySize  int64
	maxResponseBodySize int64
)

func init() {
	viper.SetDefault(maxRequestBodySizeKey, 1e6)
	viper.SetDefault(maxResponseBodySizeKey, 1e6)

	maxRequestBodySize = viper.GetInt64(maxRequestBodySizeKey)
	maxResponseBodySize = viper.GetInt64(maxResponseBodySizeKey)
}

// httpResponse for JSON marshalling
type httpResponse struct {
	StatusCode         int
	Error, Proto, Body string
	Headers            map[string][]string
}

func newHTTP(pb *common.HTTP) (execute, error) {
	var method string
	var ok bool
	if method, ok = common.Method_name[int32(pb.Method)]; !ok {
		return nil, errors.New("unknown method")
	}

	u, err := url.Parse(pb.Url)
	if err != nil {
		return nil, fmt.Errorf("parsing url: %w", err)
	}

	if u.Host == "" {
		return nil, errors.New("relative url not allowed")
	}

	if int64(len(pb.Body)) > maxRequestBodySize {
		return nil, errors.New("request body too large")
	}

	var body io.ReadCloser
	if pb.Body != nil {
		body = ioutil.NopCloser(bytes.NewReader(pb.Body))
	}

	req, err := http.NewRequest(method, pb.Url, body)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	req.ContentLength = int64(len(pb.Body))

	// merge headers
	for key, value := range pb.Headers {
		req.Header[key] = strings.Split(value, ",")
	}

	return getHTTPExecute(req, pb.SaveResponseBody), nil
}

func getHTTPExecute(req *http.Request, saveResponseBody bool) execute {
	return func() []byte {

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return getErrorJSON(err.Error())
		}

		var result httpResponse
		result.Proto = res.Proto
		result.StatusCode = res.StatusCode
		result.Headers = res.Header

		if saveResponseBody {
			limitedReader := io.LimitedReader{R: res.Body, N: maxResponseBodySize}

			body, err := ioutil.ReadAll(&limitedReader)
			if err != nil {
				result.Error = err.Error()

			} else if limitedReader.N <= 0 { // limitedReader ran out of space.
				result.Error = "response body size exceeded limit"

			} else {
				result.Body = string(body)
			}
		}
		res.Body.Close()

		json, err := json.Marshal(result)
		if err != nil { // this should never happen
			fmt.Printf("marshalling result: %v\n", err.Error())
			return getErrorJSON("steady system error.")
		}

		return json
	}
}
