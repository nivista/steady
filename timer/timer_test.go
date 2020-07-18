package timer

import (
	"reflect"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/.gen/protos/common"
)

func TestTimerProto(t *testing.T) {
	timers := []*Timer{
		{
			id: []byte{1},
			meta: meta{
				creationTime: time.Unix(1, 2).UTC(), // strip location, we lose this in protos
			},
			progress: progress{
				completed: 5,
				skipped:   1,
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
			id: []byte{2},
			meta: meta{
				creationTime: time.Unix(1, 2).UTC(),
			},
			progress: progress{
				completed: 5,
				skipped:   1,
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
		id := timer.id
		var timerTwo *Timer = &Timer{}
		if err := timerTwo.FromMessageProto(proto, id); err != nil {
			t.Errorf("FromMessageProto at idx %v failed w/ error: %v\n", idx, err)
			continue
		}

		if !reflect.DeepEqual(timer, timerTwo) {
			t.Errorf("Going to and from proto yielded a different timer at idx %v\n", idx)
		}
	}
}

type mockExecuter struct {
	count int
}

func (m *mockExecuter) execute() {
	m.count++
}

func (m *mockExecuter) toProto() *common.Task {
	return nil
}

type mockScheduler struct {
	nextFires []time.Time
	skips     []int
	dones     []bool
	nextIdx   int
}

func (m *mockScheduler) schedule(progress, time.Time) (nextFire time.Time, skips int, done bool) {
	i := m.nextIdx
	nextFire, skips, done = m.nextFires[i], m.skips[i], m.dones[i]
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
func TestRun(t *testing.T) {
	timers := []*Timer{
		{
			id:       []byte{1},
			executer: &mockExecuter{},
			scheduler: &mockScheduler{
				nextFires: []time.Time{
					time.Date(2000, time.January, 0, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 9, 0, 0, 0, 0, time.UTC), // this is after ctx will be cancelled
				},
				skips: []int{3, 0, 0, 0},
				dones: []bool{false, false, false, false},
			},
		},
		{
			id:       []byte{1},
			executer: &mockExecuter{},
			scheduler: &mockScheduler{
				nextFires: []time.Time{
					time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 9, 0, 0, 0, 0, time.UTC), // this is after ctx will be cancelled
				},
				skips: []int{3, 0, 0, 0},
				dones: []bool{false, false, false, false},
			},
		}, {
			id:       []byte{1},
			executer: &mockExecuter{},
			scheduler: &mockScheduler{
				nextFires: []time.Time{
					time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 4, 0, 0, 0, 0, time.UTC),
					{}, // <-- zero value
				},
				skips: []int{3, 0, 0},
				dones: []bool{false, false, true},
			},
		}}

	ctx := NewContext()
	clock := clockwork.NewFakeClockAt(time.Date(2000, time.January, -1, 0, 0, 0, 0, time.UTC))
	output := make(chan *sarama.ProducerMessage)
	for _, timer := range timers {
		go timer.Run(ctx, output, clock)
	}

	assertProgress := func(expected []int) {
		actual := []int{}
		for _, t := range timers {
			actual = append(actual, t.executer.(*mockExecuter).count)
		}

		equal := true
		if len(actual) != len(expected) {
			equal = false
		}
		for i := 0; equal && i < len(actual); i++ {
			if expected[i] != actual[i] {
				equal = false
			}
		}
		if !equal {
			t.Fatalf("expected progress %v, got %v\n", expected, actual)
		}
	}

	assertProgress([]int{0, 0, 0})

	clock.Advance(time.Hour * 24)
	<-output
	assertProgress([]int{1, 0, 0})

	clock.Advance(time.Hour * 24)
	<-output
	assertProgress([]int{1, 1, 0})

	clock.Advance(time.Hour * 24)
	<-output
	assertProgress([]int{1, 1, 1})

	clock.Advance(time.Hour * 24)
	<-output
	<-output
	assertProgress([]int{2, 2, 1})

	clock.Advance(time.Hour * 24)
	<-output
	assertProgress([]int{2, 2, 2})
	// at this point one of the goroutines should stop on its own

	clock.Advance(time.Hour * 24)
	<-output
	assertProgress([]int{2, 3, 2})

	clock.Advance(time.Hour * 24)
	<-output
	assertProgress([]int{3, 3, 2})

	go ctx.Cancel()
	ctx.Cancel() // should get no error doing this twice or concurrently

	clock.Advance(time.Hour * 24 * 2)
	<-time.After(time.Second * 2) // wait to see if any dangling goroutines send to output

	select {
	case <-output:
		t.Fatalf("Should not be recieving output after ctx is closed")
	default:
	}
}
