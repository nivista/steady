package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nivista/tasktimer/timer"
)

func TestA(t *testing.T) {
	c, err := NewClient(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	id, _ := uuid.NewRandom()
	newTimer := timer.Timer{
		ID:             id,
		Account:        "bob",
		ExecutionCount: 5,
		Task: &timer.HTTP{
			URL:    "www.example.org",
			Method: timer.GET,
		},
		Schedule: &timer.Interval{
			Start: time.Now(),
		},
		Meta: timer.Meta{
			CreationTime: time.Now(),
		},
	}

	err = c.CreateTimer(context.Background(), newTimer)

	if err != nil {
		t.Fatal(err)
	}
}

func TestB(t *testing.T) {
	c, err := NewClient(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.Parse("8d386939-f810-4c5b-83d2-64cd9cea1085")
	if err != nil {
		t.Fatal(err)
	}

	oldTimer, err := c.GetTimer(context.Background(), "bob", id)

	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(oldTimer)

}

func TestC(t *testing.T) {
	c, err := NewClient(context.Background())

	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.Parse("8d386939-f810-4c5b-83d2-64cd9cea1085")
	if err != nil {
		t.Fatal(err)
	}

	err = c.SetExecCount(context.Background(), "bob", id, 9)

	if err != nil {
		t.Fatal(err)
	}
}
