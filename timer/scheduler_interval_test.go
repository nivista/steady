package timer

import (
	"testing"
	"time"
)

func TestIntervalSchedule(t *testing.T) {
	type testCase struct {
		interval                interval
		progress                Progress
		time                    time.Time
		expectedNextFire        time.Time
		expectedExecutionNumber int
		expectedDone            bool
	}

	testCases := []testCase{
		{
			interval: interval{
				start:      time.Unix(0, 0),
				interval:   10,
				executions: 4,
			},
			progress:                Progress{LastExecution: 1},
			time:                    time.Unix(0, 0),
			expectedNextFire:        time.Unix(10, 0),
			expectedExecutionNumber: 2,
			expectedDone:            false,
		},
		{
			interval: interval{
				start:      time.Unix(0, 0),
				interval:   10,
				executions: 4,
			},
			progress: Progress{
				LastExecution:       5,
				CompletedExecutions: 4,
			},
			time:         time.Unix(0, 0),
			expectedDone: true,
		},
		{
			interval: interval{
				start:      time.Date(2012, time.April, 0, 0, 0, 0, 132, time.UTC),
				interval:   100,
				executions: 4,
			},
			progress: Progress{
				LastExecution:       5,
				CompletedExecutions: 3,
			},
			expectedExecutionNumber: 6,
			time:                    time.Date(2012, time.April, 0, 0, 0, 0, 131, time.UTC).Add(600e9),
			expectedNextFire:        time.Date(2012, time.April, 0, 0, 0, 0, 131, time.UTC).Add(600e9),
		},
		{
			interval: interval{
				start:      time.Date(2012, time.March, 20, 0, 0, 0, 0, time.UTC),
				interval:   60,
				executions: InfiniteExecutions,
			},
			progress: Progress{
				LastExecution:       32,
				CompletedExecutions: 30,
			},
			time:                    time.Date(2012, time.March, 20, 0, 35, 0, 0, time.UTC),
			expectedNextFire:        time.Date(2012, time.March, 20, 0, 35, 0, 0, time.UTC),
			expectedExecutionNumber: 36,
		},
		{
			interval: interval{
				start:      time.Date(2012, time.March, 20, 0, 0, 0, 0, time.UTC),
				interval:   60,
				executions: InfiniteExecutions,
			},
			progress: Progress{
				LastExecution:       32,
				CompletedExecutions: 30,
			},
			expectedExecutionNumber: 33,
			time:                    time.Date(2012, time.March, 20, 0, 30, 0, 0, time.UTC),
			expectedNextFire:        time.Date(2012, time.March, 20, 0, 32, 0, 0, time.UTC),
		},
	}

	for idx, test := range testCases {
		nextFire, executionNumber, done := test.interval.schedule(test.progress, test.time)
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
