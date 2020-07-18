package timer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nivista/steady/.gen/protos/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// cron is a schedule configured by a cron string.
type cron struct {
	start      time.Time
	executions int

	min, hour, dayOfMonth, month, dayOfWeek int
}

var monthLengths = [12]int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

func (c cron) schedule(prog progress, now time.Time) (nextFire time.Time, skips int, done bool) {
	if c.executions != -1 && prog.completed >= c.executions {
		done = true
		return
	}

	nextFire = c.start

	// a fire time is a time when a timer was supposed to fire.
	// a fire time is said to be satisfied if it has been completed or skipped.
	fireTimesSatisfied := prog.completed + prog.skipped

	// advance until nextFire has passed the last satisfied fire time
	for i := 0; i <= fireTimesSatisfied; i++ {
		if i > 0 {
			// nextClosestFireTime won't advance if nextFire is "on" a valid fire time
			nextFire = nextFire.Add(time.Minute)
		}
		nextFire = c.nextClosestFireTime(nextFire)
	}

	// nextFire is in the present or future
	if !now.After(nextFire) {
		return
	}

	// count how many unsatisfied fire times there are between the last satisfied fire time and the present (not inclusive)
	fireTimesNotSatisfied := 0
	for now.After(nextFire) {
		nextFire = nextFire.Add(time.Minute)
		nextFire = c.nextClosestFireTime(nextFire)
		fireTimesNotSatisfied++
	}

	// compensate for one missed fire
	nextFire = now
	skips = fireTimesNotSatisfied - 1
	return
}

// nextClosestFireTime finds the nearest valid fire time based on cron schedule after or equal to t
func (c cron) nextClosestFireTime(t time.Time) time.Time {
	res := t

	if c.min != -1 {
		minutesBehind := c.min - res.Minute()
		if minutesBehind < 0 {
			minutesBehind += 60
		}
		res = res.Add(time.Duration(minutesBehind) * time.Minute)
	}

	if c.hour != -1 {
		hoursBehind := c.hour - res.Hour()
		if hoursBehind < 0 {
			hoursBehind += 24
		}
		res = res.Add(time.Duration(hoursBehind) * time.Hour)
	}

	if c.dayOfWeek != -1 {
		daysBehind := c.dayOfWeek - int(res.Weekday())
		if daysBehind < 0 {
			daysBehind += 7
		}

		res = res.AddDate(0, 0, daysBehind)
	} else {
		var daysBehind, monthsBehind int

		if c.dayOfMonth != -1 {
			daysBehind = c.dayOfMonth - res.Day()
			if daysBehind < 0 {
				daysBehind += monthLengths[res.Month()-1]
			}
		}

		if c.month != -1 {
			monthsBehind = c.month - int(res.Month())
			if monthsBehind < 0 {
				monthsBehind += 12
			}
		}

		res = res.AddDate(0, monthsBehind, daysBehind)
	}

	return res
}

func (c *cron) parseCron(cronString string) error {
	cronStrings := strings.Split(cronString, " ")

	if len(cronStrings) != 5 {
		return fmt.Errorf("parseCron: expected 5 part cron description, got %v", len(cronStrings))
	}

	var parser cronParser

	c.min = parser.parse(cronStrings[0], "minute", 0, 59)
	c.hour = parser.parse(cronStrings[1], "hour", 0, 23)
	c.month = parser.parse(cronStrings[3], "month", 0, 11)

	if c.month != -1 && parser.err == nil {
		c.dayOfMonth = parser.parse(cronStrings[2], "day of month", 1, monthLengths[c.month])
		c.dayOfMonth-- // day of month from time package format one-indexed to zero-indexed
	}
	c.dayOfWeek = parser.parse(cronStrings[4], "day of week", 0, 6)

	if parser.err != nil {
		return fmt.Errorf("parseCron: %w", parser.err)
	}
	return nil
}

type cronParser struct {
	err error
}

func (parser *cronParser) parse(cronDigit, field string, min, max int) int {
	if parser.err != nil {
		return 0
	}

	if cronDigit == "*" {
		return -1
	}

	val, err := strconv.Atoi(cronDigit)
	if err != nil {
		parser.err = fmt.Errorf("%v: %v", field, err)
		return 0
	}

	if val < min || val > max {
		parser.err = fmt.Errorf("%v: got %v expected in range [%v, %v]", field, val, min, max)
		return 0
	}
	return val
}

func (c cron) validateCron() error {
	if c.dayOfWeek != -1 && (c.dayOfMonth != -1 || c.month != -1) {
		return errors.New("validateCron: can't specify day of week and day of month or month")
	}

	if c.dayOfMonth != -1 && c.month != -1 && c.dayOfMonth >= monthLengths[c.month] {
		return errors.New("validateCron: this day will never happen")
	}

	return nil
}

func (c cron) toProto() *common.Schedule {
	cronInts := []int{c.min, c.hour, c.dayOfMonth + 1, c.month, c.dayOfWeek}
	cronStrings := make([]string, 5)
	for idx, val := range cronInts {
		if val == -1 {
			cronStrings[idx] = "*"
		} else {
			cronStrings[idx] = strconv.Itoa(val)
		}
	}

	cronString := strings.Join(cronStrings, " ")

	return &common.Schedule{
		Schedule: &common.Schedule_CronConfig{
			CronConfig: &common.CronConfig{
				StartTime:  timestamppb.New(c.start),
				Cron:       cronString,
				Executions: int32(c.executions),
			},
		},
	}
}

func (c *cron) fromProto(p *common.Schedule_CronConfig) error {
	cronConfig := p.CronConfig

	myCron := cron{
		start:      cronConfig.StartTime.AsTime(),
		executions: int(cronConfig.Executions),
	}

	if err := myCron.parseCron(cronConfig.Cron); err != nil {
		return fmt.Errorf("cron fromProto: %w", err)
	}

	if err := myCron.validateCron(); err != nil {
		return fmt.Errorf("cron fromProto: %w", err)
	}

	*c = myCron
	return nil
}
