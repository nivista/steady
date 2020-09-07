package coordinator

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/runtimers/db"

	"github.com/nivista/steady/internal/.gen/protos/messaging"

	"github.com/nivista/steady/timer"
	"google.golang.org/protobuf/proto"
)

// Manager manages a partition of timers. It is not concurrency safe.
type Manager struct {
	createTopic, executeTopic string
	partition                 int32
	db                        db.Client

	Active       bool
	GenerationID string

	timers   map[string]timer.Timer
	producer chan<- *sarama.ProducerMessage
	clock    clockwork.Clock
}

// unexported because the lifecycle of managers is managed by coordinator.
func newManager(producer chan<- *sarama.ProducerMessage, db db.Client, createTopic, executeTopic string, partition int32, clock clockwork.Clock) *Manager {
	return &Manager{
		timers:       make(map[string]timer.Timer),
		db:           db,
		producer:     producer,
		clock:        clock,
		partition:    partition,
		createTopic:  createTopic,
		executeTopic: executeTopic,
	}
}

// AddTimer adds a timer to the manager, and starts it if the manager is active.
func (m *Manager) AddTimer(pk string, t timer.Timer) error {
	m.timers[pk] = t

	if m.Active {
		m.startTimer(pk)
	}

	return nil
}

// RemoveTimer stops a timer if it is running and removes it from the manager.
func (m *Manager) RemoveTimer(pk string) {
	timer, ok := m.timers[pk]
	if !ok {
		return
	}

	go timer.Stop()

	delete(m.timers, pk)
}

// Start starts all the managers timers, and causes new timers that are created to also be started.
func (m *Manager) Start(ctx context.Context) error {
	if m.Active == true {
		return nil
	}
	m.Active = true

	var timerPks = make([]string, len(m.timers))
	i := 0
	for pk := range m.timers {
		timerPks[i] = pk
		i++
	}

	progs, err := m.db.GetTimerProgresses(ctx, timerPks)
	if err != nil {
		return err
	}

	for pk := range m.timers {
		prog, ok := progs[pk]
		if ok {
			m.timers[pk] = m.timers[pk].WithProgress(prog)
		}
		m.startTimer(pk)
	}
	return nil
}

func (m *Manager) stop() {
	for _, timer := range m.timers {
		go timer.Stop()
	}
}

func (m *Manager) startTimer(pk string) {
	m.timers[pk].Start(m.executeTimerFunc, m.finishTimerFunc, m.clock)
}

func (m *Manager) executeTimerFunc(execMsg *messaging.Execute, pk string) {

	bytes, err := proto.Marshal(execMsg)
	if err != nil {
		fmt.Printf("progress update fn timerData w/ id %v, err proto.Marshal: %v\n", pk, err.Error())
		return
	}

	m.producer <- &sarama.ProducerMessage{
		Topic:     m.executeTopic,
		Key:       sarama.StringEncoder(pk),
		Value:     sarama.ByteEncoder(bytes),
		Partition: m.partition,
	}

}

func (m *Manager) finishTimerFunc(pk string) {
	m.producer <- &sarama.ProducerMessage{
		Topic:     m.createTopic,
		Key:       sarama.StringEncoder(pk),
		Value:     nil,
		Partition: m.partition,
	}

	m.producer <- &sarama.ProducerMessage{
		Topic:     m.executeTopic,
		Key:       sarama.StringEncoder(pk),
		Value:     nil,
		Partition: m.partition,
	}

	// TODO remove timer from manager?? or wait until we get Kafka message?
}
