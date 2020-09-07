package timer

import (
	"errors"
	"time"

	"github.com/nivista/steady/.gen/protos/common"
	"github.com/robfig/cron"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type schedule func(prog progress, now time.Time) *time.Time

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

func newSchedule(p *common.Schedule) schedule {
	var err error
	var sched cron.Schedule
	if sched, err = validateSchedule(p); err != nil {
		panic("scheduler.New called with invalid Schedule: " + err.Error())
	}

	return func(prog progress, now time.Time) *time.Time {
		// check executions condition
		if p.MaxExecutions != common.Executions_INFINITE && prog.completedExecutions >= int32(p.MaxExecutions) {
			return nil
		}

		var beforeNextFire time.Time
		if prog.lastExecution.IsZero() { // this is the first execution
			if p.StartTime == nil { // start time wasn't specified, start now
				p.StartTime = timestamppb.New(now)
			}
			beforeNextFire = p.StartTime.AsTime().Add(-1) // rollback one nanosecond
		} else {
			beforeNextFire = prog.lastExecution
		}

		// find next fire
		var nextFire = sched.Next(beforeNextFire)

		if nextFire.Before(now) { // missed fire
			nextFire = now // compensate for missed fire
		}

		// check stoptime condition
		if p.StopTime != nil && nextFire.After(p.StopTime.AsTime()) {
			return nil
		}

		return &nextFire
	}
}

func validateSchedule(p *common.Schedule) (cron.Schedule, error) {
	sched, err := parser.Parse(p.Cron)
	if err != nil {
		return nil, errors.New("invalid cron config: " + err.Error())
	}
	return sched, nil
}
