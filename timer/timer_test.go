package timer

import (
	"time"

	"github.com/nivista/steady/.gen/protos/common"
)

/* func TestTimerProto(t *testing.T) {
	timers := []*Timer{
		{
			meta: meta{
				creationTime: time.Unix(1, 2).UTC(), // strip location, we lose this in protos
			},
			progress: progress{
				completedExecutions: 5,
				lastExecution:       5,
			},
			executer: http{
				method:  GET,
				url:     "http://example.com",
				body:    "",
				headers: map[string]string{"set-cookie": "hello:cookie"},
			},
			scheduler: cron{
				start:      time.Unix(1, 2).UTC(),
				executions: 10,
				min:        -1,
				hour:       2,
				dayOfMonth: 3,
				month:      3,
				dayOfWeek:  -1,
			},
		},
		{
			meta: meta{
				creationTime: time.Unix(1, 2).UTC(),
			},
			progress: progress{
				completedExecutions: 5,
				lastExecution:       6,
			},
			executer: http{
				method:  GET,
				url:     "http://example.com",
				body:    "",
				headers: map[string]string{"set-cookie": "hello:cookie"},
			},
			scheduler: interval{
				start:      time.Unix(1, 2).UTC(),
				interval:   2,
				executions: 10,
			},
		}}

	for idx, timer := range timers {
		proto := timer.ToMessageProto()
		var timerTwo *Timer = &Timer{}
		if err := timerTwo.FromMessageProto(proto); err != nil {
			t.Errorf("FromMessageProto at idx %v failed w/ error: %v\n", idx, err)
			continue
		}

		if !reflect.DeepEqual(timer, timerTwo) {
			t.Errorf("Going to and from proto yielded a different timer at idx %v\n", idx)
		}
	}
} */

type mockExecuter struct {
}

func (m *mockExecuter) execute() {
}

func (m *mockExecuter) toProto() *common.Task {
	return nil
}

type mockScheduler struct {
	nextFires        []time.Time
	executionNumbers []int
	dones            []bool
	nextIdx          int
}

func (m *mockScheduler) schedule(Progress, time.Time) (nextFire time.Time, executionNumber int, done bool) {
	i := m.nextIdx
	nextFire, executionNumber, done = m.nextFires[i], m.executionNumbers[i], m.dones[i]
	m.nextIdx++
	return
}

func (m *mockScheduler) toProto() *common.Schedule {
	return nil
}

// this test runs three timers with interleaved schedules. One should terminate on its own
// and two should be terminated by context. Two fire at the same time once.
// in terms of sequence: a b c ba c* b a **
// letters refer to timer executions, no spaces mean they fire at exact same time,
// * means fire then terminate, ** means cancel context.
/* func TestRun(t *testing.T) {
	var executionCounts = [3]int{0, 0, 0}
	fmt.Println(unsafe.Pointer(&executionCounts))
	timers := []*Timer{
		{
			executer: &mockExecuter{},
			scheduler: &mockScheduler{
				nextFires: []time.Time{
					time.Date(2000, time.January, 0, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 9, 0, 0, 0, 0, time.UTC), // this is after ctx will be cancelled
				},
				dones: []bool{false, false, false, false},
			},
		},
		{
			executer: &mockExecuter{},
			scheduler: &mockScheduler{
				nextFires: []time.Time{
					time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 9, 0, 0, 0, 0, time.UTC), // this is after ctx will be cancelled
				},
				dones: []bool{false, false, false, false},
			},
		}, {
			executer: &mockExecuter{},
			scheduler: &mockScheduler{
				nextFires: []time.Time{
					time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 4, 0, 0, 0, 0, time.UTC),
					{}, // <-- zero value
				},
				dones: []bool{false, false, true},
			},
		}}

	clock := clockwork.NewFakeClockAt(time.Date(2000, time.January, -1, 0, 0, 0, 0, time.UTC))

	var expectedExecutionCounts = [][3]int{
		{0, 0, 0},
		{1, 0, 0},
		{1, 1, 0},
		{1, 1, 1},
		{2, 2, 1},
		{2, 2, 2},
		{2, 3, 2},
		{3, 3, 2},
		{3, 3, 2},
	}

	var (
		progressUpdates
		skips

		expectedProgressUpdates
		expectedSkips
	)

	cancellers := make([]Canceller, len(timers))
	for idx, timer := range timers {
		cancellers[idx] = timer.Run(func(Progress) {

		}, func() {

		}, Progress{}, clock)
	}

	for _, expected := range expectedExecutionCounts {
		if expected != executionCounts {
			t.Fatalf("expected for execution counts %v, got %v.\n", expected, executionCounts)
		}
		clock.Advance(time.Hour * 24)
	}

} */

/* func TestRun(t *testing.T) {
	var timer Timer = Timer{
		executer: &mockExecuter{},
		scheduler: &mockScheduler{
			nextFires: []time.Time{
				time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC),
				time.Date(2000, time.January, 3, 0, 0, 0, 0, time.UTC),
				{},
			},
			dones:            []bool{false, false, false, false},
			executionNumbers: []int{1, 2, 3, 0},
		},
	}

	var prog Progress

	var clock = clockwork.NewFakeClockAt(time.Date(2000, time.January, 0, 0, 0, 0, 0, time.UTC))

	var expectedProgressUpdates = []Progress{
		{CompletedExecutions: 1, LastExecution: 1},
		{CompletedExecutions: 2, LastExecution: 2},
		{CompletedExecutions: 3, LastExecution: 3},
	}
	var idx = 0
	var err string
	var handleProgressUpdate = func(p Progress) {
		if err == "" && idx > len(expectedProgressUpdates) {
			err = "Too many calls to handleProgressUpdate"
		}
		if err == "" && expectedProgressUpdates[idx] != p {
			err = fmt.Sprintf("For progress at idx %v, expected %v, got %v", idx, expectedProgressUpdates[idx], p)
		}

		idx++
	}

	var finishes int
	var handleFinish = func() {
		finishes++
	}

	var cancel = timer.Run(handleProgressUpdate, handleFinish, prog, clock)

	var checkErr = func() {
		if err != "" {
			t.Fatal(err)
		}
	}
	for i := 0; i < len(expectedProgressUpdates); i++ {
		checkErr()
		clock.BlockUntil(1)
		checkErr()
		clock.Advance(24 * time.Hour)
		checkErr()
		clock.BlockUntil(0)
	}
	checkErr()
	cancel.Cancel()
	clock.BlockUntil(1)
	clock.BlockUntil(0)
}
*/
