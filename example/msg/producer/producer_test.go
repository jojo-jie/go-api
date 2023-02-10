package producer

import (
	"github.com/Shopify/sarama"
	"os"
	"os/signal"
	"sync"
	"testing"
)

const ip = "127.0.0.1:9092"

func TestSyncPro(t *testing.T) {
	producer, err := NewSyncProducer([]string{ip})
	if err != nil {
		t.Error(err)
	}
	partition, offset, err := producer.sendMsg(&sarama.ProducerMessage{
		Topic: "quickstart-events",
		Value: sarama.StringEncoder("走着走"),
	})
	if err != nil {
		t.Logf("FAILED to send message: %s\n", err)
	} else {
		t.Logf("> message sent to partition %d at offset %d\n", partition, offset)
	}
}

func TestAsyncPro(t *testing.T) {
	producer, err := NewAsyncProducer([]string{ip})
	if err != nil {
		return
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range producer.producer.Successes() {

		}
	}()
}
