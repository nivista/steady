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

var monthLengths = []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

func (c cron) schedule(progress *progress, task executer, updateProgress chan<- int, stop <-chan int) {
	nextFire := c.helper(progress, task, updateProgress)

	select {
	case <-stop:
	case <-time.NewTimer(time.Until(nextFire)).C:
		nextFire = c.helper(progress, task, updateProgress)
	}
}

func (c cron) helper(progress *progress, task executer, updateProgress chan<- int) time.Time {
	nextFire := c.start
	passedFireTimes := progress.completed + progress.skipped
	for i := 0; i < passedFireTimes; i++ {
		if i > 0 {
			nextFire = nextFire.Add(time.Minute)
		}
		nextFire = c.advance(nextFire)
	}

	diff := time.Now().Sub(nextFire)
	if diff < 0 {
		task.execute()

		progress.completed++
		updateProgress <- 0

		nextFire = c.advance(nextFire)

		for diff < 0 {
			progress.skipped++
			updateProgress <- 0

			nextFire = nextFire.Add(time.Minute)
			nextFire = c.advance(nextFire)
		}
	}
	return nextFire
}

func (c cron) advance(t time.Time) time.Time {
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

type cronParseError struct {
	description string
}

func (err cronParseError) Error() string {
	return err.description
}

func (c *cron) parseCron(cronString string) error {
	cronStrings := strings.Split(cronString, " ")

	if len(cronStrings) != 5 {
		return fmt.Errorf("parseCron: expected 5 part cron description, got %v", len(cronStrings))
	}

	var parser cronParser

	c.min = parser.parse(cronStrings[0], "minute", 0, 59)
	c.hour = parser.parse(cronStrings[1], "hour", 0, 23)
	c.dayOfMonth = parser.parse(cronStrings[2], "day of month", 1, 31)
	c.month = parser.parse(cronStrings[3], "month", 0, 11)
	c.dayOfWeek = parser.parse(cronStrings[4], "day of week", 0, 7)

	return parser.err
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
	cronInts := []int{c.min, c.hour, c.dayOfMonth, c.month, c.dayOfWeek}
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
		return fmt.Errorf("cron fromProto: %v", err)
	}

	if err := myCron.validateCron(); err != nil {
		return fmt.Errorf("cron fromProto: %v", err)
	}

	*c = myCron
	return nil
}
