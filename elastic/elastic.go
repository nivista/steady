package elastic

import "encoding/json"

type (
	// Index is a request to index a new document.
	Index struct {
		Doc interface{} `json:"doc"`
	}

	// Get is the response to a get request.
	Get struct {
		Found  bool `json:"found"`
		Source struct {
			Doc json.RawMessage `json:"doc"`
		} `json:"_source"`
	}

	// User is a user document.
	User struct {
		HashedAPIKey string `json:"hashed_api_key"`
	}
)
