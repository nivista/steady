package manager

import (
	"github.com/nivista/steady/store"
	"github.com/nivista/steady/timer"
)

type manager struct {
}

func newManager(c *store.Client) *manager {
	return &manager{}
}

func (m *manager) start() {

}

func (m *manager) stop() {

}

func (m *manager) addTimer(t *timer.Timer) {

}

func (m *manager) removeTimer(t *timer.Timer) {

}
