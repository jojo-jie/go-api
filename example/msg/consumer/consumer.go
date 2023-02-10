package consumer

import (
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"os"
)

type InitConf struct {
	brokers  []string
	group    string
	version  string
	topics   []string
	assignor string
	oldest   bool
	verbose  bool
}

func (c *InitConf) GetBrokers() []string {
	return c.brokers
}

func (c *InitConf) GetGroup() string {
	return c.group
}

func (c *InitConf) GetVersion() string {
	return c.version
}

func (c *InitConf) GetTopics() []string {
	return c.topics
}

func (c *InitConf) GetAssignor() string {
	return c.assignor
}

func (c *InitConf) GetOldest() bool {
	return c.oldest
}

func (c *InitConf) GetVerbose() bool {
	return c.verbose
}

type Option func(c *InitConf)

func WithBrokers(brokers []string) Option {
	return func(c *InitConf) {
		c.brokers = brokers
	}
}

func WithGroup(group string) Option {
	return func(c *InitConf) {
		c.group = group
	}
}

func WithVersion(version string) Option {
	return func(c *InitConf) {
		c.version = version
	}
}

func WithTopics(topics []string) Option {
	return func(c *InitConf) {
		c.topics = topics
	}
}

func WithAssignor(assignor string) Option {
	return func(c *InitConf) {
		c.assignor = assignor
	}
}

func WithOldest(oldest bool) Option {
	return func(c *InitConf) {
		c.oldest = oldest
	}
}

func WithVerbose(verbose bool) Option {
	return func(c *InitConf) {
		c.verbose = verbose
	}
}

func NewConsumerGroupConf(opts ...Option) *InitConf {
	c := &InitConf{}
	for _, o := range opts {
		o(c)
	}
	return c
}

func New(c *InitConf) (sarama.ConsumerGroup, error) {
	if c.verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}
	version, err := sarama.ParseKafkaVersion(c.version)
	if err != nil {
		return nil, err
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
	config.Version = version
	switch c.assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategySticky}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRoundRobin}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}
	default:
		return nil, errors.New(fmt.Sprintf("Unrecognized consumer group partition assignor: %s", c.assignor))
	}
	if c.oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	return sarama.NewConsumerGroup(c.brokers, c.group, config)
}

type Consumer struct {
	ready chan bool
	id    uint64
}

func NewConsumer() *Consumer {
	return &Consumer{
		ready: make(chan bool),
	}
}

func (consumer *Consumer) SetReady() {
	consumer.ready = make(chan bool)
}

func (consumer *Consumer) GetReady() chan bool {
	return consumer.ready
}

func (consumer *Consumer) GetId() uint64 {
	return consumer.id
}

func (consumer *Consumer) SetId(id uint64) {
	consumer.id = id
}

func (consumer *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	log.Println("============")
	log.Println(session.Claims(), consumer.id)
	// 重置偏移量 offset
	session.ResetOffset("quickstart-events", 0, 0, "")
	return nil
}

func (consumer *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("Message consumer id: value = %d", consumer.id)

	for {
		select {
		case message := <-claim.Messages():
			log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s, consumner_id = %d", string(message.Value), message.Timestamp, message.Topic, consumer.id)
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
