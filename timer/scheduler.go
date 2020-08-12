package timer

import (
	"time"

	"github.com/nivista/steady/.gen/protos/common"
	rcron "github.com/robfig/cron/v3"
)

type cron struct {
	sched               rcron.Schedule
	startTime, stopTime time.Time
	maxExecutions       int
}

var parser = rcron.NewParser(rcron.Minute | rcron.Hour | rcron.Dom | rcron.Month | rcron.Dow | rcron.SecondOptional | rcron.Descriptor)

func (c cron) schedule(prog Progress, now time.Time) (nextFire time.Time, done bool) {
	if c.maxExecutions != InfiniteExecutions && prog.CompletedExecutions >= c.maxExecutions {
		done = true
		return
	}

	if prog.LastExecution.IsZero() {
		// subtract one second so sched.Next will return startTime if it is valid
		nextFire = c.startTime.Add(-1 * time.Second)
	} else {
		nextFire = prog.LastExecution
	}

	for {
		nextFire = c.sched.Next(nextFire)
		if c.stopTime.IsZero() && nextFire.After(c.stopTime) {
			done = true
			return
		}

		// !(now > nextFire) --> now <= nextFire
		if !now.After(nextFire) {
			return
		}
	}
}

func (c *cron) fromProto(p *common.Schedule) error {
	var newCron cron

	sched, err := parser.Parse(p.Cron)
	if err != nil {
		return err
	}

	newCron.sched = sched
	newCron.maxExecutions = int(p.MaxExecutions)
	newCron.startTime = p.StartTime.AsTime().UTC()
	newCron.stopTime = p.StopTime.AsTime().UTC()

	*c = newCron
	return nil
}
