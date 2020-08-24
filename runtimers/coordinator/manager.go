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

	GenerationID string
	Active       bool

	timers   map[string]timer.Timer
	producer chan<- *sarama.ProducerMessage
	clock    clockwork.Clock
}

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

func (m *Manager) AddTimer(pk string, t *messaging.Create) error {

	myTimer, err := timer.New(t, pk)
	if err != nil {
		return err
	}

	m.timers[pk] = myTimer

	if m.Active {
		m.startTimerData(pk, &messaging.Progress{})
	}

	return nil
}

func (m *Manager) RemoveTimer(pk string) {
	timer, ok := m.timers[pk]
	if !ok {
		return
	}

	timer.Stop()

	delete(m.timers, pk)
}

// TODO go to elastic for data
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
		if !ok {
			prog = &messaging.Progress{}
		}
		m.startTimerData(pk, prog)
	}
	return nil
}

func (m *Manager) stop() {
	for _, timer := range m.timers {
		timer.Stop()
	}
}

func (m *Manager) startTimerData(pk string, prog *messaging.Progress) {
	m.timers[pk].Start(m.executeTimerFunc, m.finishTimerFunc, prog, m.clock)
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
		Headers: []sarama.RecordHeader{{
			Key:   []byte("generationID"),
			Value: []byte(m.GenerationID),
		}},
	}

}

func (m *Manager) finishTimerFunc(pk string) {
	m.producer <- &sarama.ProducerMessage{
		Topic:     m.createTopic,
		Key:       sarama.StringEncoder(pk),
		Value:     nil,
		Partition: m.partition,
		Headers: []sarama.RecordHeader{{
			Key:   []byte("generationID"),
			Value: []byte(m.GenerationID),
		}},
	}

	m.producer <- &sarama.ProducerMessage{
		Topic:     m.executeTopic,
		Key:       sarama.StringEncoder(pk),
		Value:     nil,
		Partition: m.partition,
	}

	// TODO remove timer from manager?? or wait until we get Kafka message?
}
