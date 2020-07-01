package timer

import (
	"time"

	"github.com/google/uuid"
)

// Timer definition
type (
	Timer struct {
		ID        uuid.UUID
		Domain    string
		Executer  Executer
		Scheduler Scheduler
		meta      meta
		progress  progress
	}

	// Executer is responsible for a single timer execution.
	Executer interface {
		Execute()
	}

	// Scheduler can execute a recurring task. It updates progress through updateP and is stopped through
	// the stop channel. updateP will be closed when Scheduler is done on its own.
	Scheduler interface {
		Schedule(initialP progress, task Executer, updateP <-chan progress, stop <-chan int)
	}

	meta struct {
		CreationTime time.Time
	}

	progress struct {
		completed int
		skipped   int
	}
)
