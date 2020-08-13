package consumer

import (
	"fmt"
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/internal/.gen/protos/messaging"
	"github.com/nivista/steady/runtimers/coordinator"
	"google.golang.org/protobuf/proto"
)

// Consumer consumes from the database and runs its partition of timers.
type Consumer struct {
	producer   sarama.AsyncProducer
	coord      *coordinator.Coordinator
	partitions int32
	nodeID     string
	topic      string
}

// NewConsumer returns a new Consumer.
func NewConsumer(producer sarama.AsyncProducer, coord *coordinator.Coordinator, partitions int32, nodeID, topic string) sarama.ConsumerGroupHandler {
	return &Consumer{
		producer:   producer,
		coord:      coord,
		partitions: partitions,
		nodeID:     nodeID,
		topic:      topic,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	partitions := session.Claims()[c.topic]

	for _, partition := range partitions {
		// if we already have this partition, do nothing
		if c.coord.HasPartition(partition) {
			continue
		}

		// read from beggining
		session.ResetOffset(c.topic, partition, -1, "")

		key := messaging.Key{Key: &messaging.Key_Dummy_{Dummy: &messaging.Key_Dummy{}}}
		keyBytes, err := proto.Marshal(&key)
		if err != nil {
			panic(err)
		}

		// send dummy message to know when "present" is
		c.producer.Input() <- &sarama.ProducerMessage{
			Topic: c.topic,
			Key:   sarama.ByteEncoder(keyBytes),
			Value: nil,
			Headers: []sarama.RecordHeader{{
				Key:   []byte("generationID"),
				Value: []byte(strconv.Itoa(int(session.GenerationID()))),
			}},
			Partition: partition,
		}
	}

	// drop and stop whatever partitions you need to.
	c.coord.HandleRepartition(partitions)

	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	var (
		man = c.coord.GetManager(claim.Partition())

		// set of the keys of all the potentially stale progresses
		staleProgressCandiates = map[string]struct{}{}
	)

	// this happens concurrently with the firing of timers and affects messages produced.
	// we want the gaurantee that we will send producer messages w/ monotonically increasing generationIDs.
	// anyways this is bad because it's undefined behavior
	// instead we can use a RWLock() for updating generation ID
	man.GenerationID = strconv.Itoa(int(session.GenerationID()))

	for msg := range claim.Messages() {
		var key messaging.Key
		err := proto.Unmarshal(msg.Key, &key)
		if err != nil {
			// we didn't mark the message here oops
			fmt.Println("unmarshal key error:", err)
			continue
		}
		var headers = map[string]string{}

		for _, header := range msg.Headers {
			headers[string(header.Key)] = string(header.Value)
		}

		switch k := key.Key.(type) {
		// A timer Create or Delete
		case *messaging.Key_CreateTimer_:
			var (
				id, err = uuid.Parse(k.CreateTimer.TimerUuid)
				domain  = k.CreateTimer.Domain
			)

			if err != nil {
				fmt.Println("unmarshal id err:", err)
			}

			if msg.Value == nil {
				man.RemoveTimer(id)
				break
			}

			var timer messaging.CreateTimer
			err = proto.Unmarshal(msg.Value, &timer)
			if err != nil {
				fmt.Println("consume claim unmarshal timer:", err.Error())
				break
			}

			err = man.AddTimer(id, domain, &timer)
			if err != nil {
				fmt.Println("consume claim add timer:", err.Error())
				break
			}

		// A Progress update or delete
		case *messaging.Key_ExecuteTimer_:
			id, err := uuid.Parse(k.ExecuteTimer.TimerUuid)
			if err != nil {
				fmt.Println("unmarshal id execute err:", err)
			}

			if msg.Value == nil {
				// the timer associated with this progress must've been deleted.
				delete(staleProgressCandiates, string(msg.Key))
				break
			}

			var prog common.Progress
			err = proto.Unmarshal(msg.Value, &prog)
			if err != nil {
				fmt.Println("consume claim unmarshal progress:", err.Error())
				break
			}

			// we're reading historical progress updates to get to the current state
			if man.HasTimer(id) && !man.Active {
				man.UpdateTimerProgress(id, &prog)
				break
			}

			// this progress update is from a timer we already have running
			if man.HasTimer(id) {
				break
			}

			// in this case we definitley already sent a progress delete
			if man.GenerationID == headers["generationID"] {
				break
			}

			// in this case this progress update probably doesn't have a delete coming yet.
			if man.Active {
				c.producer.Input() <- &sarama.ProducerMessage{
					Topic: c.topic,
					Key:   sarama.ByteEncoder(msg.Key),
					Value: nil,
					Headers: []sarama.RecordHeader{{
						Key:   []byte("generationID"),
						Value: []byte(man.GenerationID),
					}},
					Partition: msg.Partition,
				}
				break
			}

			// there might be a delete just later in the queue
			staleProgressCandiates[string(msg.Key)] = struct{}{}

		case *messaging.Key_Dummy_:
			// is this me
			if headers["generationID"] != man.GenerationID {
				break
			}

			// if so start the manager
			man.Start()

			// delete stale progress updates
			for key := range staleProgressCandiates {

				c.producer.Input() <- &sarama.ProducerMessage{
					Topic: c.topic,
					Key:   sarama.ByteEncoder(key),
					Value: nil,
					Headers: []sarama.RecordHeader{{
						Key:   []byte("generationID"),
						Value: []byte(man.GenerationID),
					}},
					Partition: msg.Partition,
				}

			}

		default:
			fmt.Println("Unknown key type:", string(msg.Key))
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
