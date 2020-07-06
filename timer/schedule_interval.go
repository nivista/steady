package timer

import (
	"time"

	"github.com/nivista/steady/.gen/protos/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	// Cron is a schedule configured by a cron string.

	// Interval is a schedule configured by a fixed interval.
	interval struct {
		start      time.Time
		interval   int
		executions int
	}
)

func (i interval) schedule(progress *progress, task executer, updateProgress chan<- int, stop <-chan int) {
	nextFire := i.helper(progress, task, updateProgress)
	select {
	case <-stop:
	case <-time.NewTimer(time.Until(nextFire)).C:
		nextFire = i.helper(progress, task, updateProgress)
	}
}

func (i interval) helper(progress *progress, task executer, updateProgress chan<- int) time.Time {
	intervalNano := i.interval * 1e9
	intervalDuration := time.Duration(intervalNano)

	intervalsAhead := progress.completed + progress.skipped
	nextFire := i.start.Add(time.Duration(intervalNano * intervalsAhead))
	diff := time.Now().Sub(nextFire)

	// next fire is in the past
	if diff > 0 {
		// fire once now
		task.execute()

		diff -= intervalDuration

		progress.completed++
		updateProgress <- 0

		// for every other missed fire just skip
		for diff > 0 {
			diff -= intervalDuration
			progress.skipped++
			updateProgress <- 0
		}

		intervalsAhead = progress.completed + progress.skipped
		nextFire = i.start.Add(time.Duration(intervalNano * intervalsAhead))
	}

	return nextFire
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
