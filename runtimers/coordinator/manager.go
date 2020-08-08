package coordinator

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"github.com/nivista/steady/keys"
	"github.com/nivista/steady/timer"
	"google.golang.org/protobuf/proto"
)

// manager manages a partition of timers. It is not concurrency safe.
type Manager struct {
	active    bool
	workers   map[uuid.UUID]*worker
	producer  chan<- *sarama.ProducerMessage
	clock     clockwork.Clock
	partition int32
	topic     string
}

// worker manages a single timer.
type worker struct {
	timer           *timer.Timer
	initialProgress timer.Progress

	timerKey, timerProgressKey sarama.Encoder

	timer.Canceller
}

func newManager(producer chan<- *sarama.ProducerMessage, topic string, partition int32, clock clockwork.Clock) *Manager {
	return &Manager{
		workers:   make(map[uuid.UUID]*worker),
		producer:  producer,
		clock:     clock,
		partition: partition,
		topic:     topic,
	}
}

func (m *Manager) AddTimer(id uuid.UUID, domain string, t *messaging.CreateTimer) error {
	var timer timer.Timer

	var err = timer.FromMessageProto(t)
	if err != nil {
		return err
	}

	m.workers[id] = &worker{
		timer:            &timer,
		timerKey:         keys.NewTimer(domain, id),
		timerProgressKey: keys.NewTimerProgress(domain, id),
	}

	if m.active {
		m.startWorker(id)
	}

	return nil
}

func (m *Manager) UpdateTimerProgress(id uuid.UUID, prog *common.Progress) {
	m.workers[id].initialProgress = timer.Progress{
		CompletedExecutions: int(prog.CompletedExecutions),
		LastExecution:       int(prog.LastExecution),
	}
}

func (m *Manager) HasTimer(id uuid.UUID) bool {
	_, ok := m.workers[id]
	return ok
}
func (m *Manager) RemoveTimer(id uuid.UUID) {
	worker, ok := m.workers[id]
	if !ok {
		return
	}

	worker.Cancel()

	delete(m.workers, id)
}

func (m *Manager) Start() {
	if m.active == true {
		return
	}
	m.active = true

	for id := range m.workers {
		m.startWorker(id)
	}
}

func (m *Manager) IsActive() bool {
	return m.active
}

func (m *Manager) stop() {
	if m.active == false {
		return
	}
	m.active = false

	for _, worker := range m.workers {
		worker.Cancel()
	}
	m.workers = nil
}

func (m *Manager) startWorker(id uuid.UUID) {
	var w = m.workers[id]

	var progressUpdateFunc = func(prog timer.Progress) {
		fmt.Println("progressupdatefunc: ", prog)
		var progPB = common.Progress{
			LastExecution:       int32(prog.LastExecution),
			CompletedExecutions: int32(prog.CompletedExecutions),
		}

		bytes, err := proto.Marshal(&progPB)
		if err != nil {
			fmt.Printf("progress update fn worker w/ id %v, err proto.Marshal: %v\n", id, err.Error())
			return
		}

		m.producer <- &sarama.ProducerMessage{
			Topic:     m.topic,
			Key:       w.timerProgressKey,
			Value:     sarama.ByteEncoder(bytes),
			Partition: m.partition,
			Headers:   []sarama.RecordHeader{{[]byte("sender"), []byte("runtimers prog update")}},
		}
		fmt.Println("progressupdatefunc sent")
	}

	var finishTimerFunc = func() {
		fmt.Println("finish timer func")
		m.producer <- &sarama.ProducerMessage{
			Topic:     m.topic,
			Key:       w.timerKey,
			Value:     nil,
			Partition: m.partition,
			Headers:   []sarama.RecordHeader{{[]byte("sender"), []byte("runtimers timer delete")}},
		}

		m.producer <- &sarama.ProducerMessage{
			Topic:     m.topic,
			Key:       w.timerProgressKey,
			Value:     nil,
			Partition: m.partition,
			Headers:   []sarama.RecordHeader{{[]byte("sender"), []byte("runtimers prog delete")}},
		}
	}

	w.Canceller = w.timer.Run(progressUpdateFunc, finishTimerFunc, w.initialProgress, m.clock)
}
