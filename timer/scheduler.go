package timer

import (
	"time"

	rcron "github.com/robfig/cron/v3"
)

type cron struct {
	sched               rcron.SpecSchedule
	startTime, stopTime time.Time
	maxExecutions       int
}

var zeroTime = time.Time{}

func (c cron) schedule(prog Progress, now time.Time) (nextFire time.Time, done bool) {
	if c.maxExecutions != InfiniteExecutions && prog.CompletedExecutions >= c.maxExecutions {
		done = true
		return
	}

	for {
		nextFire = c.sched.Next(now)

		if c.stopTime != zeroTime && nextFire.After(c.stopTime) {
			done = true
			return
		}

		if !now.Before(nextFire) {
			return
		}
	}
}
