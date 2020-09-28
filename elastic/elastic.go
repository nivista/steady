package elastic

import "encoding/json"

type (
	User struct {
		HashedAPIKey string `json:"hashed_api_key"`
	}

	Get struct {
		Found  bool `json:"found"`
		Source struct {
			Doc json.RawMessage `json:"doc"`
		} `json:"_source"`
	}

	Index struct {
		Doc interface{} `json:"doc"`
	}
)
