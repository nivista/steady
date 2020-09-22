package timer

import (
	"fmt"
	"time"

	"github.com/nivista/steady/.gen/protos/common"
	"github.com/robfig/cron"
)

type schedule func(prog progress, now time.Time) *time.Time

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

const rollback = -1 * time.Nanosecond

func newSchedule(p *common.Schedule) (schedule, error) {
	sched, err := parser.Parse(p.Cron)
	if err != nil {
		return nil, fmt.Errorf("invalid cron: %w", err)
	}

	if p.StartTime == nil {
		return nil, fmt.Errorf("start time required")
	}

	return func(prog progress, now time.Time) *time.Time {
		// check executions condition
		if p.MaxExecutions != common.Executions_INFINITE && prog.completedExecutions >= int32(p.MaxExecutions) {
			return nil
		}

		var beforeNextFire time.Time
		if prog.lastExecution == nil { // this is the first execution
			beforeNextFire = p.StartTime.AsTime().Add(rollback)
		} else {
			beforeNextFire = *prog.lastExecution
		}

		// find next fire
		nextFire := sched.Next(beforeNextFire)

		if nextFire.Before(now) { // missed fire
			nextFire = now // compensate for missed fire
		}

		// check stoptime condition
		if p.StopTime != nil && nextFire.After(p.StopTime.AsTime()) {
			return nil
		}

		return &nextFire
	}, nil
}
