package main

import (
	"context"
)

func (app *application) CreatePromptService(ctx context.Context) ([]byte, error) {
	app.logger.Info("Creating prompt service...")
	// var cancel context.CancelFunc
	// ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()
	return nil, nil
}
