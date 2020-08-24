package executer

import (
	"encoding/json"

	"github.com/nivista/steady/.gen/protos/common"
)

type (
	Executer interface {
		Execute() []byte
	}

	// how's this?
	executer struct {
		task *common.Task
	}

	// this is for marshalling to a json that looks like { Error: "description" }
	errorResult struct {
		Error string
	}
)

// New validation here? meaning return error?
func New(t *common.Task) Executer {
	return &executer{task: t}
}

func (e *executer) Execute() []byte {
	switch task := e.task.Task.(type) {
	case *common.Task_Http:
		return executeHTTP(task)
	default:
		return getError("unknown task")
	}
}

func getError(err string) []byte {
	data, _ := json.Marshal(errorResult{err})
	return data
}
