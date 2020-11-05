package coordinator

import (
	"context"
	"fmt"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/jonboulle/clockwork"
	"github.com/nivista/steady/runtimers/db"
	"go.uber.org/atomic"

	"github.com/nivista/steady/internal/.gen/protos/messaging"

	"github.com/nivista/steady/timer"
	"google.golang.org/protobuf/proto"
)

// Manager manages a partition of timers. It is not concurrency safe.
type Manager struct {
	createTopic, executeTopic string
	partition                 int
	db                        db.Client

	GenerationID string

	haveCreates    bool
	haveProgresses bool
	started        *atomic.Bool

	creates    map[string]*messaging.Create
	progresses map[string]*messaging.Progress
	timers     map[string]timer.Timer
	timersLock sync.Mutex

	producer chan<- *sarama.ProducerMessage
	clock    clockwork.Clock
}

// unexported because the lifecycle of managers is managed by coordinator.
func newManager(producer chan<- *sarama.ProducerMessage, db db.Client, createTopic, executeTopic string, partition int, clock clockwork.Clock) *Manager {
	manager := Manager{
		timers:       make(map[string]timer.Timer),
		progresses:   make(map[string]*messaging.Progress),
		creates:      make(map[string]*messaging.Create),
		db:           db,
		producer:     producer,
		clock:        clock,
		partition:    partition,
		started:      atomic.NewBool(false),
		createTopic:  createTopic,
		executeTopic: executeTopic,
	}
	// query from db
	go func() {
		var err error
		for i := 0; i < 10; i++ {
			manager.progresses, err = db.GetProgresses(context.TODO(), partition)
			if err != nil {
				fmt.Printf("ERROR: getting progresses from db.")
			} else {
				break
			}
		}
		if err != nil {
			// escalate to a panic if we can't reach database
			panic(fmt.Errorf("unable to get progress from db: %w", err))
		}
		manager.haveProgresses = true
		manager.attemptStart()
	}()

	return &manager
}

func (m *Manager) attemptStart() {
	m.timersLock.Lock()
	defer m.timersLock.Unlock()
	fmt.Println(m.haveCreates, m.haveProgresses)
	if m.haveCreates && m.haveProgresses {
		s := m.started.CAS(false, true)
		fmt.Println(s)
		if s {
			for id, create := range m.creates {
				t, err := timer.NewWithProgress(create, m.progresses[id], m.executeTimerFunc(id), m.finishTimerFunc(id), m.clock)
				if err == nil {
					m.timers[id] = t
					t.Start()
				} else {
					fmt.Printf("error timer.NewWithProgress with id %v: %v\n", id, err)
				}
			}
		}
	}
}

// CreateTimer adds a timer to the manager, and starts it if the manager is active.
func (m *Manager) CreateTimer(pk string, create *messaging.Create) {
	if m.started.Load() {
		t, err := timer.New(create, m.executeTimerFunc(pk), m.finishTimerFunc(pk), m.clock)
		if err == nil {
			fmt.Println("error constructing timer with id %v: %v", pk, err.Error())
		} else {
			m.timersLock.Lock()
			m.timers[pk] = t
			t.Start()
			m.timersLock.Unlock()
		}
	} else {
		err := timer.IsValid(create)
		if err == nil {
			m.creates[pk] = create
		} else {
			fmt.Println("recieved invalid create id %v: %v", create, err.Error())
		}
	}
}

// RemoveTimer stops a timer if it is running and removes it from the manager.
func (m *Manager) RemoveTimer(pk string) {
	if m.started.Load() {
		m.timersLock.Lock()
		go m.timers[pk].Stop()
		delete(m.timers, pk)
		m.timersLock.Unlock()
	} else {
		delete(m.creates, pk)
	}
}

// RecievedDummy indicates that we've seen the dummy message, so creates recieved can be processed.
func (m *Manager) RecievedDummy() {
	m.haveCreates = true
	m.attemptStart()
}

func (m *Manager) stop() {
	for _, timer := range m.timers {
		go timer.Stop()
	}
}

func (m *Manager) executeTimerFunc(pk string) func(execMsg *messaging.Execute) {
	return func(execMsg *messaging.Execute) {
		bytes, err := proto.Marshal(execMsg)
		if err != nil {
			fmt.Printf("progress update fn timerData w/ id %v, err proto.Marshal: %v\n", pk, err.Error())
			return
		}

		m.producer <- &sarama.ProducerMessage{
			Topic:     m.executeTopic,
			Key:       sarama.StringEncoder(pk),
			Value:     sarama.ByteEncoder(bytes),
			Partition: int32(m.partition),
		}
	}

}

func (m *Manager) finishTimerFunc(pk string) func() {
	return func() {
		m.producer <- &sarama.ProducerMessage{
			Topic:     m.createTopic,
			Key:       sarama.StringEncoder(pk),
			Value:     nil,
			Partition: int32(m.partition),
		}

		m.producer <- &sarama.ProducerMessage{
			Topic:     m.executeTopic,
			Key:       sarama.StringEncoder(pk),
			Value:     nil,
			Partition: int32(m.partition),
		}
	}
}
