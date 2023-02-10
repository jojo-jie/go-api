package producer

import (
	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
)

type SyncProducer struct {
	producer sarama.SyncProducer
}

func (p *SyncProducer) sendMsg(message *sarama.ProducerMessage) (int32, int64, error) {
	return p.producer.SendMessage(message)
}

func NewSyncProducer(host []string) (*SyncProducer, error) {
	producer, err := sarama.NewSyncProducer(host, nil)
	if err != nil {
		return nil, err
	}

	return &SyncProducer{
		producer: producer,
	}, nil
}

type AsyncProducer struct {
	producer sarama.AsyncProducer
}

func NewAsyncProducer(host []string) (*AsyncProducer, error) {
	config := mocks.NewTestConfig()
	producer, err := sarama.NewAsyncProducer(host, config)
	if err != nil {
		return nil, err
	}
	return &AsyncProducer{
		producer: producer,
	}, nil
}
