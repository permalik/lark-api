package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

func ProducePromptRaw() {
	seeds := []string{"localhost:9092"}
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup("llm"),
		kgo.ConsumeTopics("prompt.raw"),
	)
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)
	record := &kgo.Record{Topic: "prompt.raw", Value: []byte("Why is the sky blue?")}
	cl.Produce(ctx, record, func(_ *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			fmt.Printf("Record had a produce error: %v\n", err)
		}
	})
	wg.Wait()
}
