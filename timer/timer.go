package timer

import (
	"sync"

	"github.com/jonboulle/clockwork"

	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"github.com/nivista/steady/timer/executer"
	"github.com/nivista/steady/timer/scheduler"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	// Timer represents the configuration for the execution of a recurring task.
	Timer interface {
		Start(executeTimer func(execMsg *messaging.Execute, pk string), finishTimer func(pk string), initialProgress *messaging.Progress, clock clockwork.Clock)
		Stop()
	}

	timer struct {
		pk        string
		executer  executer.Executer
		scheduler scheduler.Scheduler

		active bool
		ch     chan struct{}
		mux    sync.Mutex
	}
)

// Run starts the timers execution.
func (t *timer) Start(executeTimer func(execMsg *messaging.Execute, pk string), finishTimer func(pk string), initialProgress *messaging.Progress, clock clockwork.Clock) {
	t.mux.Lock()
	if t.active {
		t.mux.Unlock()
		return
	}

	t.active = true
	t.ch = make(chan struct{})
	t.mux.Unlock()

	var progress = initialProgress
	if progress == nil {
		progress = &messaging.Progress{}
	}

	go func() {
		for {
			nextFire, done := t.scheduler.Schedule(progress, clock.Now())

			if done {
				finishTimer(t.pk)
				return
			}

			select {
			case currentFire := <-clock.After(nextFire.Sub(clock.Now())):
				res := t.executer.Execute()
				progress.CompletedExecutions++
				progress.LastExecution = timestamppb.New(currentFire)
				executeTimer(&messaging.Execute{
					Progress: progress,
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
		close(t.ch)
	}
}

// New creates a Timer from a *messaging.CreateTimer.
func New(p *messaging.Create, pk string) (Timer, error) {
	var t timer

	t.pk = pk
	t.executer = executer.New(p.Task)

	var err error
	t.scheduler, err = scheduler.New(p.Schedule)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
