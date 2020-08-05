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
	if i.executions != -1 && prog.completedExecutions >= i.executions {
		done = true
		return
	}

	intervalNano := i.interval * 1e9

	nextFire = i.start.Add(time.Duration(intervalNano * prog.lastExecution))
	executionNumber = prog.lastExecution + 1

	diff := now.Sub(nextFire)

	// nextFire is in the past
	if diff > 0 {
		nextFire = now
		executionNumber += int(diff) / intervalNano
	}

	return
}

func (i interval) toProto() *common.Schedule {
	return &common.Schedule{
		Schedule: &common.Schedule_IntervalConfig{
			IntervalConfig: &common.IntervalConfig{
				StartTime:  timestamppb.New(i.start),
				Interval:   int32(i.interval),
				Executions: common.Executions(i.executions),
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
