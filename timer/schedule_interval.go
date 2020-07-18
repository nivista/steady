package timer

import (
	"time"

	"github.com/nivista/steady/.gen/protos/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// interval is a scheduler configured by a fixed interval.
type interval struct {
	start      time.Time
	interval   int
	executions int
}

func (i interval) schedule(prog progress, now time.Time) (nextFire time.Time, skips int, done bool) {
	intervalNano := i.interval * 1e9

	if i.executions != -1 && prog.completed >= i.executions {
		done = true
		return
	}

	timePassed := now.Sub(i.start)
	if timePassed < 0 {
		nextFire = i.start
		return
	}

	// An executionPoint is a time when a timer was supposed to fire.
	// An executionPoint is said to be satisfied when it is completed or skipped.
	fireTimesSatisfied := prog.completed + prog.skipped

	// All execution points in the past
	fireTimesPassed := int(timePassed)/intervalNano + 1

	fireTimesNotSatisfied := fireTimesPassed - fireTimesSatisfied

	// We are after executions that have not been completed or skipped. Fire once now and record all the other missed executions as skipped.
	if fireTimesNotSatisfied >= 1 {
		skips = fireTimesNotSatisfied - 1
		nextFire = now
		return
	}

	// The next executions that have not been completed or skipped are in the future
	nextFire = i.start.Add(time.Duration(intervalNano * fireTimesSatisfied))
	return
}

func (i interval) toProto() *common.Schedule {
	return &common.Schedule{
		Schedule: &common.Schedule_IntervalConfig{
			IntervalConfig: &common.IntervalConfig{
				StartTime:  timestamppb.New(i.start),
				Interval:   int32(i.interval),
				Executions: int32(i.executions),
			},
		},
	}
}

func (i *interval) fromProto(p *common.Schedule_IntervalConfig) error {
	intConfig := p.IntervalConfig

	*i = interval{
		start:      intConfig.StartTime.AsTime(),
		interval:   int(intConfig.Interval),
		executions: int(intConfig.Executions),
	}

	return nil
}
