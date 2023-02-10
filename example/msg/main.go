package main

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"msg/consumer"
	"os"
	"os/signal"
	"snowflake"
	"sync"
	"syscall"
)

func main() {
	c := consumer.NewConsumerGroupConf(
		consumer.WithBrokers([]string{"127.0.0.1:9092"}),
		consumer.WithGroup("example15"),
		consumer.WithVersion("3.3.1"),
		consumer.WithTopics([]string{"quickstart-events"}),
		consumer.WithAssignor("range"),
		consumer.WithOldest(true),
		consumer.WithVerbose(false),
	)
	client, err := consumer.New(c)
	if err != nil {
		log.Panicf("Error from new consumer: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	keepRunning := true
	consumptionIsPaused := false
	w := snowflake.NewWorker(5, 5)
	consumner01 := consumer.NewConsumer()
	id01, _ := w.NextID()
	consumner01.SetId(id01)
	fmt.Println(consumner01.GetId())
	consumerList := make([]*consumer.Consumer, 0, 3)
	consumerList = append(consumerList, consumner01)
	wg := &sync.WaitGroup{}
	wg.Add(len(consumerList))
	for _, cl := range consumerList {
		go func(cl *consumer.Consumer) {
			defer wg.Done()
			for {
				if err := client.Consume(ctx, c.GetTopics(), cl); err != nil {
					log.Panicf("Error from consumer: %v", err)
				}
				if ctx.Err() != nil {
					return
				}
				consumner01.SetReady()
			}
		}(cl)
	}
	// Await till the consumer has been set up
	<-consumner01.GetReady()
	log.Println("Sarama consumer up and running!...")
	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")
			keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(client, &consumptionIsPaused)
		}
	}

	cancel()
	wg.Wait()
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Println("Resuming consumption")
	} else {
		client.PauseAll()
		log.Println("Pausing consumption")
	}
	*isPaused = !*isPaused
}
