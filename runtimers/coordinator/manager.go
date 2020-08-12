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
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Manager manages a partition of timers. It is not concurrency safe.
type Manager struct {
	Active       bool
	GenerationID string
	workers      map[uuid.UUID]*worker
	producer     chan<- *sarama.ProducerMessage
	clock        clockwork.Clock
	partition    int32
	topic        string
}

// worker manages a single timer.
type worker struct {
	timer           *timer.Timer
	initialProgress timer.Progress

	timerKey, timerProgressKey sarama.Encoder

	cancelFn func()
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
		timerKey:         keys.NewCreateTimer(domain, id),
		timerProgressKey: keys.NewExecuteTimer(domain, id),
	}

	if m.Active {
		m.startWorker(id)
	}

	return nil
}

func (m *Manager) UpdateTimerProgress(id uuid.UUID, prog *common.Progress) {
	m.workers[id].initialProgress = timer.Progress{
		CompletedExecutions: int(prog.CompletedExecutions),
		LastExecution:       prog.LastExecution.AsTime(),
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

	if m.Active {
		worker.cancelFn()
	}

	delete(m.workers, id)
}

func (m *Manager) Start() {
	if m.Active == true {
		return
	}
	m.Active = true

	for id := range m.workers {
		m.startWorker(id)
	}
}

func (m *Manager) stop() {
	if m.Active == false {
		return
	}
	m.Active = false

	for _, worker := range m.workers {
		worker.cancelFn()
	}
	m.workers = make(map[uuid.UUID]*worker)
}

func (m *Manager) startWorker(id uuid.UUID) {
	var w = m.workers[id]

	var progressUpdateFunc = func(prog timer.Progress) {
		var progPB = common.Progress{
			LastExecution:       timestamppb.New(prog.LastExecution),
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
			Headers: []sarama.RecordHeader{{
				Key:   []byte("generationID"),
				Value: []byte(m.GenerationID),
			}},
		}
	}

	var finishTimerFunc = func() {
		m.producer <- &sarama.ProducerMessage{
			Topic:     m.topic,
			Key:       w.timerKey,
			Value:     nil,
			Partition: m.partition,
			Headers: []sarama.RecordHeader{{
				Key:   []byte("generationID"),
				Value: []byte(m.GenerationID),
			}},
		}

		m.producer <- &sarama.ProducerMessage{
			Topic:     m.topic,
			Key:       w.timerProgressKey,
			Value:     nil,
			Partition: m.partition,
		}
		// TODO remove timer from manager?? or wait until we get Kafka message?
	}

	w.cancelFn = w.timer.Run(progressUpdateFunc, finishTimerFunc, w.initialProgress, m.clock)
}
