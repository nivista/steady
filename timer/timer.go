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

		// LastExecution indicates which execution was recorded last. It may be greater than CompletedExecutions if a fire
		// was missed after a long outage. For example there have been 4 successfully completed executions but the last
		// execution was 5th of the scheduled executions.
		LastExecution int
	}

	executer interface {
		execute()
	}

	scheduler interface {
		schedule(prog Progress, now time.Time) (nextFire time.Time, executionNumber int, done bool)
	}
)

// InfiniteExecutions indicates a timer should fire until cancelled.
const InfiniteExecutions = 0

// Canceller is used to cancel a goroutine started by a Run function.
type (
	Canceller interface {
		Cancel()
	}

	canceller struct {
		stopped int32
		stopCh  chan struct{}
	}
)

func (c *canceller) Cancel() {
	if atomic.CompareAndSwapInt32(&c.stopped, 0, 1) {
		close(c.stopCh)
	}
}

// Run starts the timers execution.
func (t *Timer) Run(updateProgress func(Progress), finishTimer func(), intialProgress Progress, clock clockwork.Clock) Canceller {

	progress := intialProgress
	myCanceller := canceller{stopCh: make(chan struct{})}

	go func() {
		defer fmt.Println("stopped")

		for {
			nextFire, executionNumber, done := t.scheduler.schedule(progress, clock.Now())
			if done {
				// Send delete message
				finishTimer()
				return
			}
			fmt.Println(executionNumber)
			progress.LastExecution = executionNumber

			select {
			case <-clock.After(nextFire.Sub(clock.Now())):
				t.executer.execute()
				progress.CompletedExecutions++

				updateProgress(progress)

			case <-myCanceller.stopCh:
				fmt.Println("stopped by cancelled")
				return
			}
		}
	}()

	return &myCanceller
}

// FromMessageProto gets a Timer from a messaging.CreateTimer.
func (t *Timer) FromMessageProto(p *messaging.CreateTimer) error {
	if p.Task == nil {
		return errors.New("(*Timer)FromMessageProto got nil CreateTimer")
	}

	var newTimer Timer
	fmt.Println(p)

	switch task := p.Task.Task.(type) {
	case *common.Task_HttpConfig:
		var h http
		if err := h.fromProto(task); err != nil {
			return err
		}
		newTimer.executer = h
	default:
		return errors.New("Unknown Task")
	}

	switch sched := p.Schedule.Schedule.(type) {
	case *common.Schedule_CronConfig:
		var c cron
		if err := c.fromProto(sched); err != nil {
			return err
		}
		newTimer.scheduler = c
	case *common.Schedule_IntervalConfig:
		var i interval
		if err := i.fromProto(sched); err != nil {
			return err
		}
		newTimer.scheduler = i
	default:
		return errors.New("Unknown Schedule")
	}

	*t = newTimer
	return nil
}
