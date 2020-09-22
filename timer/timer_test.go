package timer

import (
	"bytes"
	"sort"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"go.uber.org/atomic"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testCases = []struct {
	timer           *timer
	startTime       time.Time // time when timer.Start should be called.
	expectedResults map[time.Time]*messaging.Execute
	expectFinish    bool // whether the timer is expected to self terminate after executions.
}{
	// Single fire with self termination.
	{
		timer: newTimerOptimistic(
			&common.Schedule{
				// intended execution: fires once 1 second past epoch.
				Cron:          "@every 1s",
				StartTime:     timestamppb.New(time.Unix(1, 0)),
				MaxExecutions: 1,
			},
			progress{}), // no progress
		startTime: time.Unix(0, 0), // recieves timer early
		expectedResults: map[time.Time]*messaging.Execute{
			time.Unix(1, 0): {
				Progress: &messaging.Progress{
					CompletedExecutions: 1,
					LastExecution:       timestamppb.New(time.Unix(1, 0)),
				},
			},
		},
		expectFinish: true,
	},

	// Compensatory fire and self termination.
	{
		timer: newTimerOptimistic(
			&common.Schedule{
				// intended execution: fires 60 and 120 seconds past epoch.
				Cron:          "@every 1m",
				StartTime:     timestamppb.New(time.Unix(60, 0)),
				MaxExecutions: 2,
			},
			progress{
				// correctly did its first fire
				lastExecution:       unixTimePointer(60),
				completedExecutions: 1,
			}),
		startTime: time.Unix(150, 0), // node doesn't get the timer until 150 seconds past epoch
		expectedResults: map[time.Time]*messaging.Execute{
			time.Unix(150, 0): { // compensatory fire
				Progress: &messaging.Progress{
					CompletedExecutions: 2,
					LastExecution:       timestamppb.New(time.Unix(150, 0)),
				},
			},
		},
		expectFinish: true,
	},

	// Zero fires, external termination.
	{
		timer: newTimerOptimistic(
			&common.Schedule{
				// intended execution: fires 60 and 120 seconds past epoch.
				Cron:          "@every 1m",
				StartTime:     timestamppb.New(time.Unix(60, 0)),
				MaxExecutions: 2,
			},
			progress{
				// correctly did its first fire
				lastExecution:       unixTimePointer(60),
				completedExecutions: 1,
			}),
		startTime:       time.Unix(40, 0),                   // node gets the timer before its last fire, shoudn't matter.
		expectedResults: map[time.Time]*messaging.Execute{}, // isn't expected to fire
		expectFinish:    false,
	},
}

func TestTimer(t *testing.T) {
Outer:
	for idx, cfg := range testCases {
		executions, finishes := make(chan *messaging.Execute), make(chan struct{})
		fc := clockwork.NewFakeClockAt(cfg.startTime)

		cfg.timer.clock = fc
		cfg.timer.recordExecution = func(execMsg *messaging.Execute) {
			executions <- execMsg
		}

		cfg.timer.recordTermination = func() {
			finishes <- struct{}{}
		}

		cfg.timer.Start()

		// get the times from the expected results so we can iterate through them in sorted order
		var times = make([]time.Time, 0, len(cfg.expectedResults))
		for time := range cfg.expectedResults {
			times = append(times, time)
		}
		sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })

		for _, exec := range times {
			fc.Advance(exec.Sub(fc.Now())) // wait until next expected result

			select {
			case execMsg := <-executions:
				expected := cfg.expectedResults[exec]
				if execMsg.Progress.CompletedExecutions != expected.Progress.CompletedExecutions {
					t.Errorf("case: %v. time: %v. Unequal 'Progress.CompetedExecution'", idx, exec)
					continue Outer
				}

				if execMsg.Progress.LastExecution.Nanos != expected.Progress.LastExecution.Nanos {
					t.Errorf("case: %v. time: %v. Unequal 'Progress.LastExecution'", idx, exec)
					continue Outer
				}

				if !bytes.Equal(execMsg.Result, expected.Result) {
					t.Errorf("case: %v. time: %v. Unequal 'Result'", idx, exec)
					continue Outer
				}

			case <-time.After(10 * time.Millisecond): // give time for go scheduler to give control to other goroutine
				t.Errorf("case: %v. time: %v. Expected execution.", idx, exec)
				continue Outer
			}
		}

		if cfg.expectFinish {
			select {
			case <-finishes:
			case <-time.After(10 * time.Millisecond):
				t.Errorf("case: %v. Expected finish.", idx)
			}

		} else {

			stopped := make(chan struct{})
			go func() {
				cfg.timer.Stop()
				stopped <- struct{}{}
			}()
			select {
			case <-stopped:
			case <-time.After(10 * time.Millisecond):
				t.Errorf("case: %v. Stop is blocking for too long.", idx)
			}
		}
	}
}

func newTimerOptimistic(sched *common.Schedule, prog progress) *timer {
	s, err := newSchedule(sched)
	if err != nil {
		panic(err)
	}

	return &timer{
		execute:  func() []byte { return nil },
		schedule: s,
		progress: prog,
		active:   atomic.NewBool(false),
		stop:     make(chan struct{}),
	}
}

func unixTimePointer(seconds int64) *time.Time {
	t := time.Unix(seconds, 0)
	return &t
}
