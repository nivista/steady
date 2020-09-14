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
		// WithProgress should return a new timer with the given Progress.
		WithProgress(pb *messaging.Progress) Timer

		// Start starts the timers execution. It should be a no-op if the timer is active.
		Start(executeTimer func(execMsg *messaging.Execute, pk string), finishTimer func(pk string), clock clockwork.Clock)

		// Stop should stop timer execution synchronously.
		Stop()
	}

	timer struct {
		pk string

		execute
		schedule
		progress

		ready, active *atomic.Bool
		stop          chan struct{}
	}

	// making my own type here rather than using protobuf, this can be safely copied.
	progress struct {
		completedExecutions int32
		lastExecution       *time.Time
	}
)

// IsValid validates a create timer message.
func IsValid(pb *messaging.Create) error {
	_, err := New(pb, "")
	return err

}

// New creates a Timer from a create timer message and a primary key.
func New(pb *messaging.Create, pk string) (Timer, error) {
	exec, err := newExecute(pb.Task)
	if err != nil {
		return nil, errors.New("invalid task: " + err.Error())
	}

	sched, err := newSchedule(pb.Schedule)
	if err != nil {
		return nil, errors.New("invalid schedule: " + err.Error())
	}

	return &timer{
		pk:       pk,
		execute:  exec,
		schedule: sched,
		ready:    atomic.NewBool(true),
		active:   atomic.NewBool(false),
		stop:     make(chan struct{}),
	}, nil
}

func (t *timer) WithProgress(pb *messaging.Progress) Timer {
	return &timer{
		pk:       t.pk,
		execute:  t.execute,
		schedule: t.schedule,
		progress: progressFromProto(pb),
		ready:    atomic.NewBool(true),
		active:   atomic.NewBool(false),
		stop:     make(chan struct{}),
	}
}

func (t *timer) Start(executeTimer func(execMsg *messaging.Execute, pk string), finishTimer func(pk string), clock clockwork.Clock) {
	if !t.ready.CAS(true, false) { // make it so Start can't be called
		return
	}
	t.active.CAS(false, true) // make it so Stop can be called

	go func() {

		for {
			currFire := t.schedule(t.progress, clock.Now())

			if currFire == nil {
				finishTimer(t.pk)
				return
			}

			select {
			case now := <-clock.After(currFire.Sub(clock.Now())):
				res := t.execute()

				t.progress.completedExecutions++
				t.progress.lastExecution = &now

				executeTimer(&messaging.Execute{
					Progress: progressToProto(t.progress),
					Result:   res,
				}, t.pk)

			case <-t.stop:
				return
			}
		}
	}()

}

func (t *timer) Stop() {
	if !t.active.CAS(true, false) { // make it so Stop can't be called (start already can't be called)
		return
	}

	t.stop <- struct{}{} // block until timer stops

	t.ready.CAS(false, true) // make it so Start can be called
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
