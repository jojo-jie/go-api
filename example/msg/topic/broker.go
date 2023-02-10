package broker

import (
	"fmt"
	"github.com/Shopify/sarama"
	"sync"
)

type InitConf struct {
	host  string
	topic string
}

type Option func(c *InitConf)

func WithHost(host string) Option {
	return func(c *InitConf) {
		c.host = host
	}
}

func WithTopic(topic string) Option {
	return func(c *InitConf) {
		c.topic = topic
	}
}

func NewTopicConf(opts ...Option) *InitConf {
	c := &InitConf{}
	for _, o := range opts {
		o(c)
	}
	return c
}

func New(c *InitConf) (*sarama.Broker, error) {
	broker := sarama.NewBroker(c.host)
	err := broker.Open(nil)
	if err != nil {
		return nil, err
	}
	request := sarama.MetadataRequest{Topics: []string{c.topic}}
	response, err := broker.GetMetadata(&request)
	if err != nil {
		return nil, err
	}
	fmt.Println("There are", len(response.Topics), "topics active in the cluster.")
	return broker, nil
}

var once sync.Once
var client sarama.Client

func GetTopic(c *InitConf) ([]string, error) {
	once.Do(func() {
		client, _ = sarama.NewClient([]string{c.host}, sarama.NewConfig())
	})
	return client.Topics()
}
