package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

var client *kgo.Client

func InitProducer() {
	seeds := []string{"localhost:9092"}
	var err error
	client, err = kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup("llm"),
		kgo.ConsumeTopics("prompt.raw"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("Producer initialized")
}

func ProducePromptRaw(prompt string) {
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)
	record := &kgo.Record{Topic: "prompt.raw", Value: []byte(prompt)}
	client.Produce(ctx, record, func(_ *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			fmt.Printf("Record had a produce error: %v\n", err)
		}
	})
	wg.Wait()
}

func ShutdownProducer() {
	client.Close()
}
