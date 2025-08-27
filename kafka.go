package main

import (
	"context"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

func (app *application) InitKafka() {
	seeds := []string{"localhost:9092"}
	var err error
	app.client, err = kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ConsumerGroup("lark-api"),
		kgo.ConsumeTopics("response"),
	)
	if err != nil {
		app.logger.Panic(err)
	}
	app.logger.Info("Producer initialized")
}

func (app *application) ProducePromptRaw(prompt string) {
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)
	record := &kgo.Record{Topic: "prompt", Value: []byte(prompt)}
	app.client.Produce(ctx, record, func(_ *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			app.logger.Errorw("failed to produce:",
				"err", err)
		}
	})
	wg.Wait()
}

func (app *application) ShutdownKafka() {
	app.client.Close()
}
