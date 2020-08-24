package scheduler

import (
	"time"

	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	rcron "github.com/robfig/cron/v3"
)

type (
	// Scheduler schedules timers based on the current time and progress.
	Scheduler interface {
		Schedule(prog *messaging.Progress, now time.Time) (*time.Time, bool)
	}

	cron struct {
		sched               rcron.Schedule
		startTime, stopTime *time.Time
		maxExecutions       common.Executions
	}
)

var parser = rcron.NewParser(rcron.Minute | rcron.Hour | rcron.Dom | rcron.Month | rcron.Dow | rcron.SecondOptional | rcron.Descriptor)

// New returns a new Scheduler.
func New(p *common.Schedule) (Scheduler, error) {
	var c cron

	sched, err := parser.Parse(p.Cron)
	if err != nil {
		return nil, err
	}

	c.sched = sched
	c.maxExecutions = p.MaxExecutions

	if p.StartTime == nil {
		c.startTime = nil
	} else {
		var st = p.StartTime.AsTime()
		c.startTime = &st
	}

	if p.StopTime == nil {
		c.stopTime = nil
	} else {
		var st = p.StopTime.AsTime()
		c.stopTime = &st
	}

	return &c, nil
}

// Schedule returns the timers next scheduled fire, and a bool indicating if the timer should be done firing.
// The *time.Timer will be nil if and only if the bool is true.
func (c *cron) Schedule(prog *messaging.Progress, now time.Time) (*time.Time, bool) {

	// check executions condition
	if c.maxExecutions != common.Executions_INFINITE && prog != nil && prog.CompletedExecutions >= int32(c.maxExecutions) {
		return nil, true
	}

	// check stoptime condition
	if c.stopTime != nil && now.After(*c.stopTime) {
		return nil, true
	}

	var beforeNextFire time.Time

	// handle first execution case
	if prog == nil || prog.LastExecution == nil {
		// subtract one second so sched.Next will return startTime if it is valid
		// assumes startTime != nil (maybe just do now?)
		if c.startTime == nil {
			beforeNextFire = now
		} else {
			beforeNextFire = c.startTime.Add(-1)
		}
	} else {
		beforeNextFire = prog.LastExecution.AsTime()
	}

	// find next fire
	var nextFire = c.sched.Next(beforeNextFire)

	if nextFire.After(now) {
		nextFire = now
	}

	// check stoptime condition again
	if c.stopTime != nil && nextFire.After(*c.stopTime) {
		return nil, true
	}

	return &nextFire, false
}
