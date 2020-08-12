package timer

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
)

type (
	// Timer represents the configuration for the execution of a recurring task.
	Timer struct {
		executer  executer
		scheduler scheduler
	}

	// Progress records the progress of a recurring task.
	Progress struct {
		// CompletedExecutions indicated the number of successfully completed executions.
		CompletedExecutions int

		// LastExecution indicates which execution was recorded last.
		LastExecution time.Time
	}

	executer interface {
		execute()
	}

	scheduler interface {
		schedule(prog Progress, now time.Time) (nextFire time.Time, done bool)
	}
)

func getCanceller() (stopCh chan struct{}, cancelFn func()) {
	var stopped int32

	stopCh = make(chan struct{})

	cancelFn = func() {
		if atomic.CompareAndSwapInt32(&stopped, 0, 1) {
			close(stopCh)
		}
	}

	return
}

// Run starts the timers execution.
func (t *Timer) Run(updateProgress func(Progress), finishTimer func(), initialProgress Progress, clock clockwork.Clock) (cancel func()) {

	var progress = initialProgress

	stopCh, cancel := getCanceller()

	go func() {
		for {
			nextFire, done := t.scheduler.schedule(progress, clock.Now())

			if done {
				finishTimer()
				cancel()
				return
			}

			select {
			case <-clock.After(nextFire.Sub(clock.Now())):
				t.executer.execute()
				progress.CompletedExecutions++
				progress.LastExecution = nextFire
				updateProgress(progress)

			case <-stopCh:
				return
			}
		}

	}()

	return cancel
}

// FromMessageProto gets a Timer from a messaging.CreateTimer.
func (t *Timer) FromMessageProto(p *messaging.CreateTimer) error {
	if p.Task == nil {
		return errors.New("(*Timer)FromMessageProto got nil Task")
	}

	if p.Schedule == nil {
		return errors.New("(*Timer)FromMessageProto got nil Schedule")
	}

	var newTimer Timer

	switch task := p.Task.Task.(type) {
	case *common.Task_HttpConfig:
		var h http
		if err := h.fromProto(task); err != nil {
			return err
		}
		newTimer.executer = h
	default:
		return errors.New("(*Timer)FromMessageProto got unknown Task")
	}

	var c cron
	err := c.fromProto(p.Schedule)
	if err != nil {
		return fmt.Errorf("(*Timer)FromMessageProto: %w", err)
	}

	newTimer.scheduler = c

	*t = newTimer
	return nil
}
