package kafka

import "github.com/Shopify/sarama"

type message interface {
	FromConsumerMessage(*sarama.ConsumerMessage) error
	ToProducerMessage() (*sarama.ProducerMessage, error)
}
