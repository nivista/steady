package manager

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nivista/steady/messaging"
	"github.com/nivista/steady/timer"
)

type manager struct {
	active  bool
	workers map[uuid.UUID]*worker
}

type worker struct {
	timer *timer.Timer
	stop  chan int
}

func newManager(c *messaging.Client) *manager {
	return &manager{workers: make(map[uuid.UUID]*worker)}
}

func (m *manager) start() {
	for _, worker := range m.workers {
		worker.start()
	}
	m.active = true
}

func (m *manager) stop() {
	for _, worker := range m.workers {
		worker.stop <- 0
	}
	m.active = false
}

func (m *manager) addTimer(t *timer.Timer) {
	newWorker := worker{timer: t}
	m.workers[t.ID] = &newWorker
	if m.active {
		newWorker.start()
	}
}

func (m *manager) removeTimer(id uuid.UUID) {
	worker, ok := m.workers[id]
	if !ok {
		return
	}

	close(worker.stop)
	delete(m.workers, id)
}

func (w *worker) start() {
	w.stop = make(chan int)
	go func(stop <-chan int) {
		fmt.Printf("In goroutine to execute timer w/ id:%v\n", w.timer.ID)
		<-stop
		fmt.Printf("Leaving goroutine to execute timer w/ id:%v\n", w.timer.ID)
	}(w.stop)
}
