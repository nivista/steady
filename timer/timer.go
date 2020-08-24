package timer

import (
	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/internal/.gen/protos/messaging/create"
	"github.com/nivista/steady/internal/.gen/protos/messaging/execute"

	"github.com/nivista/steady/timer/executer"
	"github.com/nivista/steady/timer/scheduler"
	"go.uber.org/atomic"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	// Timer represents the configuration for the execution of a recurring task.
	Timer interface {
		Run(executeTimer func(*execute.Value), finishTimer func(), initialProgress *execute.Progress, clock clockwork.Clock) (cancel func())
	}

	timer struct {
		executer  executer.Executer
		scheduler scheduler.Scheduler
	}
)

func getCanceller() (cancelled <-chan struct{}, cancel func()) {
	var stopped atomic.Bool

	ch := make(chan struct{})

	cancel = func() {
		if stopped.Swap(true) {
			close(ch)
		}
	}

	return ch, cancel
}

// Run starts the timers execution.
func (t *timer) Run(executeTimer func(*execute.Value), finishTimer func(), initialProgress *execute.Progress, clock clockwork.Clock) (cancel func()) {

	var progress = initialProgress
	if progress == nil {
		progress = &execute.Progress{}
	}
	cancelled, cancel := getCanceller()

	go func() {
		for {
			nextFire, done := t.scheduler.Schedule(progress, clock.Now())

			if done {
				finishTimer()
				return
			}

			select {
			case currentFire := <-clock.After(nextFire.Sub(clock.Now())):
				res := t.executer.Execute()
				progress.CompletedExecutions++
				progress.LastExecution = timestamppb.New(currentFire)
				executeTimer(&execute.Value{
					Progress: progress,
					Result:   res,
				})

			case <-cancelled:
				return
			}
		}

	}()

	return cancel
}

// New creates a Timer from a *messaging.CreateTimer.
func New(p *create.Value) (Timer, error) {
	var t timer

	t.executer = executer.New(p.Task)

	var err error
	t.scheduler, err = scheduler.New(p.Schedule)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
