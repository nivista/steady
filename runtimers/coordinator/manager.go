package coordinator

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"github.com/nivista/steady/timer"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Manager manages a partition of timers. It is not concurrency safe.
type Manager struct {
	topic     string
	partition int32

	GenerationID string
	Active       bool

	timerDatas map[uuid.UUID]*timerData
	producer   chan<- *sarama.ProducerMessage
	clock      clockwork.Clock
}

// timerData is a wrapper around timer that holds the state of progress, the cancelFn, and the timers identifiers.
type timerData struct {
	timer           *timer.Timer
	initialProgress timer.Progress

	timerKey, timerProgressKey sarama.Encoder

	cancelFn func()
}

func newManager(producer chan<- *sarama.ProducerMessage, topic string, partition int32, clock clockwork.Clock) *Manager {
	return &Manager{
		timerDatas: make(map[uuid.UUID]*timerData),
		producer:   producer,
		clock:      clock,
		partition:  partition,
		topic:      topic,
	}
}

func (m *Manager) AddTimer(id uuid.UUID, domain string, t *messaging.CreateTimer) error {
	var timer timer.Timer

	var err = timer.FromMessageProto(t)
	if err != nil {
		return err
	}

	createKey := messaging.Key{
		Key: &messaging.Key_CreateTimer_{
			CreateTimer: &messaging.Key_CreateTimer{
				Domain:    domain,
				TimerUuid: id.String(),
			},
		},
	}

	createKeyBytes, err := proto.Marshal(&createKey)

	if err != nil {
		return nil
	}

	executeKey := messaging.Key{
		Key: &messaging.Key_ExecuteTimer_{
			ExecuteTimer: &messaging.Key_ExecuteTimer{
				Domain:    domain,
				TimerUuid: id.String(),
			},
		},
	}
	executeKeyBytes, err := proto.Marshal(&executeKey)
	if err != nil {
		return nil
	}
	m.timerDatas[id] = &timerData{
		timer:            &timer,
		timerKey:         sarama.ByteEncoder(createKeyBytes),
		timerProgressKey: sarama.ByteEncoder(executeKeyBytes),
	}

	if m.Active {
		m.startTimerData(id)
	}

	return nil
}

func (m *Manager) UpdateTimerProgress(id uuid.UUID, prog *common.Progress) {
	m.timerDatas[id].initialProgress = timer.Progress{
		CompletedExecutions: int(prog.CompletedExecutions),
		LastExecution:       prog.LastExecution.AsTime(),
	}
}

func (m *Manager) HasTimer(id uuid.UUID) bool {
	_, ok := m.timerDatas[id]
	return ok
}
func (m *Manager) RemoveTimer(id uuid.UUID) {
	timerData, ok := m.timerDatas[id]
	if !ok {
		return
	}

	if m.Active {
		timerData.cancelFn()
	}

	delete(m.timerDatas, id)
}

func (m *Manager) Start() {
	if m.Active == true {
		return
	}
	m.Active = true

	for id := range m.timerDatas {
		m.startTimerData(id)
	}
}

func (m *Manager) stop() {
	if m.Active == false {
		return
	}
	m.Active = false

	for _, timerData := range m.timerDatas {
		timerData.cancelFn()
	}
	m.timerDatas = make(map[uuid.UUID]*timerData)
}

func (m *Manager) startTimerData(id uuid.UUID) {
	var w = m.timerDatas[id]

	var progressUpdateFunc = func(prog timer.Progress) {
		var pb = messaging.ExecuteTimer{
			Progress: &common.Progress{
				LastExecution:       timestamppb.New(prog.LastExecution),
				CompletedExecutions: int32(prog.CompletedExecutions),
			},
		}

		bytes, err := proto.Marshal(&pb)
		if err != nil {
			fmt.Printf("progress update fn timerData w/ id %v, err proto.Marshal: %v\n", id, err.Error())
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
