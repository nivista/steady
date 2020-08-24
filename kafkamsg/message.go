package kafka

import "github.com/Shopify/sarama"

type Message interface {
	//FromConsumerMessage(*sarama.ConsumerMessage) error
	ToProducerMessage() (*sarama.ProducerMessage, error)
}
