package timer

import (
	"errors"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jonboulle/clockwork"
	"go.uber.org/atomic"

	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	// Timer is a recurring task that will invoke handlers when recurring tasks execute or are finished.
	Timer interface {
		// Start starts the timers execution idempotently.
		Start()

		// Stop stops timer execution synchronously and permanently.
		Stop()
	}

	timer struct {
		execute
		schedule
		progress

		recordExecution   func(*messaging.Execute)
		recordTermination func()

		clock clockwork.Clock

		active     *atomic.Bool
		terminated bool
		stop       chan struct{}
	}

	// making my own type here rather than using protobuf, this can be safely copied.
	progress struct {
		completedExecutions int32
		lastExecution       *time.Time
	}
)

// IsValid validates a create timer message.
func IsValid(pb *messaging.Create) error {
	_, err := New(pb, nil, nil, nil)
	return err
}

// New creates a Timer from the given create message and handlers.
func New(create *messaging.Create, recordExecution func(*messaging.Execute), recordTermination func(), clock clockwork.Clock) (Timer, error) {
	return NewWithProgress(create, nil, recordExecution, recordTermination, clock)
}

// NewWithProgress creates a new timer with the given create message, progress, and handlers.
func NewWithProgress(create *messaging.Create, prog *messaging.Progress, recordExecution func(*messaging.Execute), recordTermination func(), clock clockwork.Clock) (Timer, error) {
	exec, err := newExecute(create.Task)
	if err != nil {
		return nil, errors.New("invalid task: " + err.Error())
	}

	sched, err := newSchedule(create.Schedule)
	if err != nil {
		return nil, errors.New("invalid schedule: " + err.Error())
	}

	return &timer{
		execute:           exec,
		schedule:          sched,
		progress:          progressFromProto(prog),
		clock:             clock,
		recordExecution:   recordExecution,
		recordTermination: recordTermination,
		active:            atomic.NewBool(false),
		stop:              make(chan struct{}),
	}, nil
}

func (t *timer) Start() {
	if t.terminated {
		return
	}

	if !t.active.CAS(false, true) {
		return
	}

	go func() {

		for {
			currFire := t.schedule(t.progress, t.clock.Now())

			if currFire == nil {
				t.recordTermination()
				return
			}

			deadline := currFire.Sub(t.clock.Now())

			select {
			case now := <-t.clock.After(deadline):
				res := t.execute()

				t.progress.completedExecutions++
				t.progress.lastExecution = &now

				t.recordExecution(&messaging.Execute{
					Progress: progressToProto(t.progress),
					Result:   res,
				})

			case <-t.stop:
				return
			}
		}
	}()

}

func (t *timer) Stop() {
	t.terminated = true
	if !t.active.CAS(true, false) {
		return
	}

	t.stop <- struct{}{} // block until timer stops
}

func progressToProto(p progress) *messaging.Progress {
	var last *timestamp.Timestamp
	if p.lastExecution != nil {
		last = timestamppb.New(*p.lastExecution)
	}
	return &messaging.Progress{
		CompletedExecutions: p.completedExecutions,
		LastExecution:       last,
	}
}

func progressFromProto(pb *messaging.Progress) progress {
	var last time.Time
	if pb.LastExecution != nil {
		last = pb.LastExecution.AsTime()
	}
	return progress{
		completedExecutions: pb.CompletedExecutions,
		lastExecution:       &last,
	}
}
