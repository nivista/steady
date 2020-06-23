package timer

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

// Timer definition
type (
	Timer struct {
		ID             uuid.UUID
		Account        string
		ExecutionCount uint32
		Task
		Schedule
		Meta `json:"Meta"`
	}

	// Task represents a timer task.
	Task interface {
		isTask()
		encoding.BinaryMarshaler
		encoding.BinaryUnmarshaler
	}

	// Schedule represents the "when" configuration of a timer.
	Schedule interface {
		isSchedule()
		encoding.BinaryMarshaler
		encoding.BinaryUnmarshaler
	}

	Meta struct {
		CreationTime time.Time
	}
)

// HTTP Task represents an HTTP request.
type (
	HTTP struct {
		URL string
		Method
		Body    string
		Headers []string
	}

	// Method is the method of the HTTP request.
	Method int
)

func (HTTP) isTask() {}

func (t *HTTP) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *HTTP) UnmarshalBinary(bytes []byte) error {
	return json.Unmarshal(bytes, t)
}

// Method definitions
const (
	GET Method = iota
	POST
)

// Schedule definitions
type (
	// Cron is a schedule configured by a cron string.
	Cron struct {
		Start      time.Time
		Cron       string
		Executions int32
	}

	// Interval is a schedule configured by a fixed interval.
	Interval struct {
		Start      time.Time
		Interval   int32
		Executions int32
	}
)

func (Cron) isSchedule() {}

func (s *Cron) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Cron) UnmarshalBinary(bytes []byte) error {
	return json.Unmarshal(bytes, s)
}

func (Interval) isSchedule() {}

func (s *Interval) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Interval) UnmarshalBinary(bytes []byte) error {
	return json.Unmarshal(bytes, s)
}

func (t *Timer) MarshalBinary() ([]byte, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	fmt.Printf(">> %v\n", string(bytes))

	var m map[string]interface{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}

	var sched string
	switch t.Schedule.(type) {
	case *Cron:
		sched = "cron"
	case *Interval:
		sched = "interval"
	default:
		panic("Can't marshal schedule, unkownn type.")
	}
	m["sched"] = sched

	var task string
	switch t.Task.(type) {
	case *HTTP:
		task = "http"
	default:
		panic("Can't marshal task, unknown type.")
	}
	m["task"] = task

	bytes, err = json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return bytes, nil

}

func (t *Timer) UnmarshalBinary(bytes []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}
	var sched Schedule
	var ok bool
	switch m["sched"] {
	case "cron":
		var cron Cron
		bytes, _ = json.Marshal(m["Schedule"])
		err = json.Unmarshal(bytes, &cron)
		if err != nil {
			return errors.New("Can't unmarshal cron.")
		}
		sched = &cron
	case "interval":
		var interval Interval
		bytes, _ = json.Marshal(m["Schedule"])
		err = json.Unmarshal(bytes, &interval)
		if err != nil {
			return errors.New("Can't unmarshal interval.")
		}
		sched = &interval
	default:
		return errors.New("Unknown schedule type.")
	}

	var task Task
	switch m["task"] {
	case "http":
		var http HTTP
		bytes, _ = json.Marshal(m["Task"])
		err = json.Unmarshal(bytes, &http)
		if err != nil {
			return errors.New("Can't unmarshal HTTP.")
		}
		task = &http
	default:
		return errors.New("Unknown task type.")
	}

	fmt.Println(reflect.TypeOf(m["Meta"]))
	idString, ok := m["ID"].(string)
	if !ok {
		return errors.New("Bad id.")
	}
	id, err := uuid.Parse(idString)
	if err != nil {
		return err
	}

	account, ok := m["Account"].(string)
	if !ok {
		return errors.New("Bad account.")
	}

	execCountFloat, ok := m["ExecutionCount"].(float64)
	if !ok {
		return errors.New("Bad executioncount.")
	}
	execCount := uint32(execCountFloat)

	bytes, err = json.Marshal(m["Meta"])
	var meta Meta
	err = json.Unmarshal(bytes, &meta)
	if err != nil {
		return errors.New("Bad Meta")
	}

	*t = Timer{
		ID:             id,
		ExecutionCount: execCount,
		Task:           task,
		Schedule:       sched,
		Meta:           meta,
		Account:        account,
	}
	return nil
}
