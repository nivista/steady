package timer

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Timer definition
type (
	Timer struct {
		ID       uuid.UUID
		Domain   string
		Task     Task
		Schedule Schedule
		Meta     Meta
		progress progress
	}

	// Task represents a timer task.
	Task interface {
		isTask()
	}

	// Schedule represents the "when" configuration of a timer.
	Schedule interface {
		isSchedule()
	}

	// Meta represents timer metadata.
	Meta struct {
		CreationTime time.Time
	}

	// Progress represents the execution progress of the timer
	progress struct {
		completed int
		skipped   int
	}
)

// HTTP Task represents an HTTP request.
type (
	HTTP struct {
		URL string
		Method
		Body    string
		Headers map[string]string
	}

	// Method is the method of the HTTP request.
	Method int
)

func (HTTP) isTask() {}

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

func (Interval) isSchedule() {}

// MarshalBinary marshals a timer to a []byte.
func (t *Timer) MarshalBinary() ([]byte, error) {
	// Use json package to turn timer into map[string]interface
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}

	// Set "type indicators"
	var sched string
	switch t.Schedule.(type) {
	case *Cron:
		sched = "cron"
	case *Interval:
		sched = "interval"
	default:
		panic("Can't marshal schedule, unknown type.")
	}
	m["scheduleType"] = sched

	var task string
	switch t.Task.(type) {
	case *HTTP:
		task = "http"
	default:
		panic("Can't marshal task, unknown type.")
	}
	m["taskType"] = task

	return json.Marshal(m)
}

// UnmarshalBinary unmarshals a timer into a []byte.
func (t *Timer) UnmarshalBinary(bytes []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}

	*t = Timer{}

	switch m["scheduleType"] {
	case "cron":
		t.Schedule = &Cron{}
	case "interval":
		t.Schedule = &Interval{}
	default:
		return errors.New("unknown schedule type")
	}

	switch m["taskType"] {
	case "http":
		t.Task = &HTTP{}
	default:
		return errors.New("unknown task type")
	}

	return json.Unmarshal(bytes, t)
}
