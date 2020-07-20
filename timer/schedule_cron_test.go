package timer

import (
	"testing"
	"time"
)

func TestCronSchedule(t *testing.T) {
	everyDay := cron{
		start:      time.Date(2000, time.January, 0, 0, 0, 0, 0, time.UTC), // start firing Jan 1, 2000
		executions: 20,                                                     // fire 20 times
		min:        0,                                                      // start of every minute
		hour:       11,                                                     // 12PM, or 12 in military time
		dayOfMonth: -1,                                                     // All days of month
		month:      -1,                                                     // All months
		dayOfWeek:  -1,                                                     // All days of week
	}

	sundaysEveryHour := cron{
		start:      time.Date(2020, time.July, 18, 0, 0, 0, 0, time.UTC),
		executions: -1,
		min:        0,
		hour:       -1,
		dayOfMonth: -1,
		month:      -1,
		dayOfWeek:  0,
	}

	startOfEachMonth := cron{
		start:      time.Date(2010, time.April, 15, 0, 0, 0, 0, time.UTC),
		executions: 10,
		min:        20,
		hour:       10,
		dayOfMonth: 1,
		month:      -1,
		dayOfWeek:  -1,
	}

	type testCase struct {
		cron             cron
		progress         progress
		time             time.Time
		expectedNextFire time.Time
		expectedSkips    int
		expectedDone     bool
	}

	testCases := []testCase{
		{
			cron: everyDay,
			progress: progress{
				completedExecutions: 10,
				lastExecution:       10,
			},
			time:             time.Date(2000, time.January, 20, 0, 0, 0, 0, time.UTC),
			expectedNextFire: time.Date(2000, time.January, 20, 11, 0, 0, 0, time.UTC),
		},
		{
			cron: everyDay,
			progress: progress{
				completedExecutions: 20,
				lastExecution:       30,
			},
			time:             time.Date(2000, time.January, 22, 0, 0, 0, 0, time.UTC),
			expectedNextFire: time.Date(2000, time.January, 22, 0, 0, 0, 0, time.UTC),
			expectedSkips:    1,
		},
		{
			cron: everyDay,
			progress: progress{
				completedExecutions: 20,
				lastExecution:       22,
			},
			time:             time.Date(2000, time.January, 0, 0, 0, 0, 0, time.UTC),
			expectedNextFire: time.Date(2000, time.February, 1, 11, 0, 0, 0, time.UTC),
		},
		{
			cron:             sundaysEveryHour,
			progress:         progress{},
			time:             time.Date(2020, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectedNextFire: time.Date(2020, time.July, 19, 0, 0, 0, 0, time.UTC),
		},
		{
			cron: sundaysEveryHour,
			progress: progress{
				lastExecution:       1,
				completedExecutions: 1,
			},
			time:             time.Date(2020, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectedNextFire: time.Date(2020, time.July, 19, 1, 0, 0, 0, time.UTC),
		},
		{
			cron: sundaysEveryHour,
			progress: progress{
				completedExecutions: 24,
			},
			time:             time.Date(2020, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectedNextFire: time.Date(2020, time.July, 26, 0, 0, 0, 0, time.UTC),
		},
		{
			cron: startOfEachMonth,
			progress: progress{
				completedExecutions: 10,
			},
			expectedDone: true,
		},
		{
			cron: startOfEachMonth,
			progress: progress{
				completedExecutions: 9,
			},
			expectedDone:     false,
			time:             time.Date(2010, time.April, 1, 0, 0, 0, 0, time.UTC),
			expectedNextFire: time.Date(2011, time.February, 1, 10, 20, 0, 0, time.UTC),
		},
	}

	for idx, test := range testCases {
		nextFire, skips, done := test.cron.schedule(test.progress, test.time)
		if done != test.expectedDone {
			t.Errorf("test at idx %v: expected '%v' for done, got '%v'", idx, test.expectedDone, done)
		}

		if skips != test.expectedSkips {
			t.Errorf("test at idx %v: expected '%v' for skips, got '%v'", idx, test.expectedSkips, skips)
		}

		if nextFire != test.expectedNextFire {
			t.Errorf("test at idx %v: expected '%v' for nextFire, got '%v'", idx, test.expectedNextFire, nextFire)
		}
	}
}
