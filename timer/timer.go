package timer

import (
	"time"
)

type (
	// Timer represents the configuration for the execution of a recurring task.
	Timer struct {
		executer  executer
		scheduler scheduler
	}

	// Progress records the progress of a recurring task.
	Progress struct {
		// CompletedExecutions indicated the number of successfully completed executions.
		CompletedExecutions int

		// LastExecution indicates which execution was recorded last.
		LastExecution time.Time
	}

	executer interface {
		execute()
	}

	scheduler interface {
		schedule(prog Progress, now time.Time) (nextFire time.Time, done bool)
	}
)
