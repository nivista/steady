package timer

import (
	"errors"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jonboulle/clockwork"

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

		active bool
		ch     chan struct{}
		mux    sync.RWMutex
	}

	// zrather than using protobuf, this can be safely copied.
	progress struct {
		completedExecutions int32
		lastExecution       time.Time
	}
)

// IsValid validates a create timer message.
func IsValid(p *messaging.Create) error {
	if _, err := validateSchedule(p.Schedule); err != nil {
		return errors.New("invalid schedule: " + err.Error())
	}

	if err := validateTask(p.Task); err != nil {
		return errors.New("invalid task: " + err.Error())
	}

	return nil
}

// New creates a Timer from a create timer message and a primary key.
func New(p *messaging.Create, pk string) (Timer, error) {
	if err := IsValid(p); err != nil {
		return nil, errors.New("new timer: " + err.Error())
	}

	return &timer{
		pk:       pk,
		execute:  newExecute(p.Task),
		schedule: newSchedule(p.Schedule),
		ch:       make(chan struct{}),
	}, nil
}

func (t *timer) WithProgress(pb *messaging.Progress) Timer {
	return &timer{
		pk:       t.pk,
		execute:  t.execute,
		schedule: t.schedule,
		progress: progressFromProto(pb),
		ch:       make(chan struct{}),
	}
}

func (t *timer) Start(executeTimer func(execMsg *messaging.Execute, pk string), finishTimer func(pk string), clock clockwork.Clock) {
	t.mux.Lock()
	defer t.mux.Unlock()

	if t.active {
		return
	}
	t.active = true

	go func() {

		for {
			var currFire = t.schedule(t.progress, clock.Now())

			if currFire == nil {
				finishTimer(t.pk)
				return
			}

			select {
			case now := <-clock.After(currFire.Sub(clock.Now())):
				res := t.execute()

				t.progress.completedExecutions++
				t.progress.lastExecution = now

				executeTimer(&messaging.Execute{
					Progress: progressToProto(t.progress),
					Result:   res,
				}, t.pk)

			case <-t.ch:
				return
			}
		}
	}()

}

func (t *timer) Stop() {
	t.mux.Lock()
	defer t.mux.Unlock()

	if t.active {
		t.ch <- struct{}{} // block until timer stops
		t.active = false
	}
}

func progressToProto(p progress) *messaging.Progress {
	var last *timestamp.Timestamp
	if !p.lastExecution.IsZero() {
		last = timestamppb.New(p.lastExecution)
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
		lastExecution:       last,
	}
}
