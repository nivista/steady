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

func (i interval) schedule(prog progress, now time.Time) (nextFire time.Time, executionNumber int, done bool) {
	intervalNano := i.interval * 1e9

	if i.executions != -1 && prog.completedExecutions >= i.executions {
		done = true
		return
	}

	// lastExecution is one indexed. lastExecution == 0 means we haven't fired, so nextFire should be i.start.
	nextFire = i.start.Add(time.Duration(intervalNano * prog.lastExecution))

	// nextFire > now  // do i want >= ?
	if nextFire.After(now) {
		executionNumber = prog.lastExecution + 1
		return
	}

	// how many timer fires are between the "ideal nextFire" and now (half-open (nextFire, now] )
	skippedExecutions := int(now.Sub(nextFire)) / intervalNano

	executionNumber = prog.lastExecution + skippedExecutions + 1
	nextFire = now
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
