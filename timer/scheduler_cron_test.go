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
		cron                    cron
		progress                Progress
		time                    time.Time
		expectedNextFire        time.Time
		expectedExecutionNumber int
		expectedDone            bool
	}

	testCases := []testCase{
		{
			cron: everyDay,
			progress: Progress{
				CompletedExecutions: 10,
				LastExecution:       20,
			},
			time:                    time.Date(2000, time.January, 20, 0, 0, 0, 0, time.UTC),
			expectedExecutionNumber: 21,
			expectedNextFire:        time.Date(2000, time.January, 20, 11, 0, 0, 0, time.UTC),
		},
		{
			cron: everyDay,
			progress: Progress{
				CompletedExecutions: 10,
				LastExecution:       20,
			},
			time:                    time.Date(2000, time.January, 22, 0, 0, 0, 0, time.UTC),
			expectedNextFire:        time.Date(2000, time.January, 22, 0, 0, 0, 0, time.UTC),
			expectedExecutionNumber: 22,
		},
		{
			cron: everyDay,
			progress: Progress{
				CompletedExecutions: 10,
				LastExecution:       32,
			},
			expectedExecutionNumber: 33,
			time:                    time.Date(2000, time.January, 0, 0, 0, 0, 0, time.UTC),
			expectedNextFire:        time.Date(2000, time.February, 1, 11, 0, 0, 0, time.UTC),
		},
		{
			cron:                    sundaysEveryHour,
			time:                    time.Date(2020, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectedNextFire:        time.Date(2020, time.July, 19, 0, 0, 0, 0, time.UTC),
			expectedExecutionNumber: 1,
		},
		{
			cron: sundaysEveryHour,
			progress: Progress{
				CompletedExecutions: 1,
				LastExecution:       1,
			},
			time:                    time.Date(2020, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectedNextFire:        time.Date(2020, time.July, 19, 1, 0, 0, 0, time.UTC),
			expectedExecutionNumber: 2,
		},
		{
			cron: sundaysEveryHour,
			progress: Progress{
				CompletedExecutions: 24,
				LastExecution:       24,
			},
			time:                    time.Date(2020, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectedNextFire:        time.Date(2020, time.July, 26, 0, 0, 0, 0, time.UTC),
			expectedExecutionNumber: 25,
		},
		{
			cron: startOfEachMonth,
			progress: Progress{
				CompletedExecutions: 10,
				LastExecution:       10,
			},
			expectedDone: true,
		},
		{
			cron: startOfEachMonth,
			progress: Progress{
				CompletedExecutions: 9,
				LastExecution:       9,
			},
			expectedExecutionNumber: 10,
			time:                    time.Date(2010, time.April, 1, 0, 0, 0, 0, time.UTC),
			expectedNextFire:        time.Date(2011, time.February, 1, 10, 20, 0, 0, time.UTC),
		},
	}

	for idx, test := range testCases {
		nextFire, executionNumber, done := test.cron.schedule(test.progress, test.time)
		if done != test.expectedDone {
			t.Errorf("test at idx %v: expected '%v' for done, got '%v'", idx, test.expectedDone, done)
		}

		if executionNumber != test.expectedExecutionNumber {
			t.Errorf("test at idx %v: expected '%v' for executionNumber, got '%v'", idx, test.expectedExecutionNumber, executionNumber)
		}

		if nextFire != test.expectedNextFire {
			t.Errorf("test at idx %v: expected '%v' for nextFire, got '%v'", idx, test.expectedNextFire, nextFire)
		}
	}
}
