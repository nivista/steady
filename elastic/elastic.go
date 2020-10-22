package elastic

import (
	"encoding/json"
	"time"
)

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

	// ExecuteTimer is an execution record.
	ExecuteTimer struct {
		TimerUUID      string          `json:"timer_uuid"`
		KafkaTimestamp time.Time       `json:"kafka_timestamp"`
		Result         json.RawMessage `json:"result"`
	}

	// Progress is the value of a timers progress.
	Progress struct {
		LastExecution       time.Time `json:"last_execution"`
		CompletedExecutions int       `json:"result"`
	}
)
