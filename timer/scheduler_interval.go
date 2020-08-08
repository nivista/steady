package timer

import (
	"time"

	"github.com/nivista/steady/.gen/protos/common"
)

// interval is a scheduler configured by a fixed interval.
type interval struct {
	start      time.Time
	interval   int
	executions int
}

func (i interval) schedule(prog Progress, now time.Time) (nextFire time.Time, executionNumber int, done bool) {
	if i.executions != InfiniteExecutions && prog.CompletedExecutions >= i.executions {
		done = true
		return
	}

	intervalNano := i.interval * 1e9

	nextFire = i.start.Add(time.Duration(intervalNano * prog.LastExecution))
	executionNumber = prog.LastExecution + 1

	diff := now.Sub(nextFire)

	// nextFire is in the past
	if diff > 0 {
		nextFire = now
		executionNumber += int(diff) / intervalNano
	}

	return
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
