package timer

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMarshalInterval(t *testing.T) {
	i := Interval{
		Start:      time.Now().UTC(),
		Interval:   100,
		Executions: 10,
	}

	bytes, err := i.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(string(bytes))
	t.Logf("len of encoding: %v\n", len(bytes))

	var i2 Interval
	err = i2.UnmarshalBinary(bytes)

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(i, i2) {
		t.Fatal("Not equal.")
	}
}

func TestMarshalTimer(t *testing.T) {
	tim := getMockTimer()

	bytes, err := tim.MarshalBinary()

	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
	t.Logf("Encoding size: %v\n", len(bytes))

	var tim2 Timer

	err = tim2.UnmarshalBinary(bytes)
	if err != nil {
		t.Fatal(err)
	}
	//bytes, _ = tim2.MarshalBinary()
	t.Log(tim)
	t.Log(tim2)
	if !reflect.DeepEqual(tim.Task, tim2.Task) {
		t.Fatal("Not equal.")
	}
}

func getMockTimer() Timer {
	id, _ := uuid.Parse("09ef702a-3475-4cf4-8503-ad915f9abb5f")

	timer := Timer{
		ID:             id,
		Account:        "Hello",
		ExecutionCount: 10,
		Task: &HTTP{
			URL:     "http://www.example.com/",
			Method:  GET,
			Body:    "hello go",
			Headers: []string{"set-cookie:yum"},
		},
		Schedule: &Cron{
			Start:      time.Unix(0, 0).UTC(), // important, we're not encoding location so make sure it's set to nil (UTC)
			Cron:       "* * * * *",
			Executions: 10,
		},
		Meta: Meta{
			CreationTime: time.Unix(0, 0).UTC(),
		},
	}

	return timer
}
