package timer

import (
	"errors"

	"github.com/nivista/steady/.gen/protos/common"
)

type execute func() []byte

// newExecute assumes a task has already been validated
func newExecute(t *common.Task) execute {
	switch task := t.Task.(type) {
	case *common.Task_Http:
		return newHTTP(task.Http)
	default:
		panic("unknown task")
	}
}

func validateTask(t *common.Task) error {
	switch task := t.Task.(type) {
	case *common.Task_Http:
		_, err := httpIsValid(task.Http)
		return err
	default:
		return errors.New("unknown task")
	}
}

func getErrorJSON(err string) []byte {
	return []byte(`{"Error":"` + err + `"}`)
}
