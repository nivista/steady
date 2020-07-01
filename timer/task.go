package timer

import "fmt"

// HTTP Task represents an HTTP request.
type (
	http struct {
		url     string
		method  string
		body    string
		headers map[string]string
	}

	// Method is the method of the HTTP request.
	Method int
)

// Method definitions
const (
	GET Method = iota
	POST
)

func (h http) Execute() {
	fmt.Printf("Execute http w/ config: %v\n", h)
}
