package coordinator

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/internal/.gen/protos/messaging/create"
	"github.com/nivista/steady/runtimers/db"

	"github.com/nivista/steady/internal/.gen/protos/messaging/execute"

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

	timerDatas map[string]*timerData
	producer   chan<- *sarama.ProducerMessage
	clock      clockwork.Clock
}

// timerData is a wrapper around timer that holds the state of progress, the cancelFn, and the timers identifiers.
type timerData struct {
	timer    timer.Timer
	pk       string
	cancelFn func()
}

func newManager(producer chan<- *sarama.ProducerMessage, db db.Client, createTopic, executeTopic string, partition int32, clock clockwork.Clock) *Manager {
	return &Manager{
		timerDatas:   make(map[string]*timerData),
		db:           db,
		producer:     producer,
		clock:        clock,
		partition:    partition,
		createTopic:  createTopic,
		executeTopic: executeTopic,
	}
}

func (m *Manager) AddTimer(pk string, t *create.Value) error {

	myTimer, err := timer.New(t)
	if err != nil {
		return err
	}

	m.timerDatas[pk] = &timerData{
		timer: myTimer,
		pk:    pk,
	}

	if m.Active {
		m.startTimerData(pk, &execute.Progress{})
	}

	return nil
}

// TODO DELETE
func (m *Manager) HasTimer(pk string) bool {
	_, ok := m.timerDatas[pk]
	return ok
}

func (m *Manager) RemoveTimer(pk string) {
	timerData, ok := m.timerDatas[pk]
	if !ok {
		return
	}

	if m.Active {
		timerData.cancelFn()
	}

	delete(m.timerDatas, pk)
}

// TODO go to elastic for data
func (m *Manager) Start(ctx context.Context) error {
	if m.Active == true {
		return nil
	}
	m.Active = true

	var timerPks = make([]string, len(m.timerDatas))
	i := 0
	for pk := range m.timerDatas {
		timerPks[i] = pk
		i++
	}
	progs, err := m.db.GetTimerProgresses(ctx, timerPks)
	if err != nil {
		return err
	}

	for pk := range m.timerDatas {
		prog, ok := progs[pk]
		if !ok {
			prog = &execute.Progress{}
		}
		m.startTimerData(pk, prog)
	}
	return nil
}

func (m *Manager) stop() {
	for _, data := range m.timerDatas {
		data.cancelFn()
	}
}

func (m *Manager) startTimerData(pk string, prog *execute.Progress) {
	var w = m.timerDatas[pk]

	var executeTimerFunc = func(exec *execute.Value) {

		bytes, err := proto.Marshal(exec)
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

	var finishTimerFunc = func() {
		m.producer <- &sarama.ProducerMessage{
			Topic:     m.createTopic,
			Key:       sarama.StringEncoder(w.pk),
			Value:     nil,
			Partition: m.partition,
			Headers: []sarama.RecordHeader{{
				Key:   []byte("generationID"),
				Value: []byte(m.GenerationID),
			}},
		}

		m.producer <- &sarama.ProducerMessage{
			Topic:     m.executeTopic,
			Key:       sarama.StringEncoder(w.pk),
			Value:     nil,
			Partition: m.partition,
		}
		// TODO remove timer from manager?? or wait until we get Kafka message?
	}

	w.cancelFn = w.timer.Run(executeTimerFunc, finishTimerFunc, prog, m.clock)
}
