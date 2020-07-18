package timer

import (
	"errors"
	"fmt"

	"github.com/nivista/steady/.gen/protos/common"
)

// HTTP Task represents an HTTP request.
type (
	http struct {
		url     string
		method  method
		body    string
		headers map[string]string
	}

	// Method is the method of the HTTP request.
	method int
)

// Method definitions
const (
	GET method = iota
	POST
)

func (h http) execute() {
	fmt.Printf("Execute http w/ config: %v\n", h)
}

func (h http) toProto() *common.Task {
	return &common.Task{
		Task: &common.Task_HttpConfig{
			HttpConfig: &common.HTTPConfig{
				Url:     h.url,
				Method:  h.method.toProto(),
				Body:    h.body,
				Headers: h.headers,
			},
		},
	}
}

func (h *http) fromProto(p *common.Task_HttpConfig) error {
	httpConfig := p.HttpConfig
	var m method
	err := m.fromProto(httpConfig.Method)
	if err != nil {
		return err
	}
	*h = http{
		url:     httpConfig.Url,
		method:  m,
		body:    httpConfig.Body,
		headers: httpConfig.Headers,
	}
	return nil
}

func (m method) toProto() common.Method {
	switch m {
	case GET:
		return common.Method_GET
	case POST:
		return common.Method_POST
	default:
		panic("Unknown method type")
	}
}

func (m *method) fromProto(p common.Method) error {
	switch p {
	case common.Method_GET:
		*m = GET
	case common.Method_POST:
		*m = POST
	default:
		return errors.New("Invalid method")
	}
	return nil
}
