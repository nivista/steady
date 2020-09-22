package timer

import (
	"errors"
	"fmt"

	"github.com/nivista/steady/.gen/protos/common"
)

type execute func() []byte

func newExecute(t *common.Task) (execute, error) {
	switch task := t.Task.(type) {
	case *common.Task_Http:
		return newHTTP(task.Http)
	default:
		return nil, errors.New("unknown task")
	}
}

func getErrorJSON(err string) []byte {
	return []byte(fmt.Sprintf("{\"Error\":\"%v\"}", err))
}
