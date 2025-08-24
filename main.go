package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var cfg config
	flag.StringVar(&cfg.env, "env", "development", "environment (development|staging|production)")
	flag.IntVar(&cfg.port, "port", 5555, "Network port (default 5555)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger, _ := newLogger()
	defer logger.Sync()
	sugar := logger.Sugar()

	var kgo *kgo.Client

	app := &application{
		config: cfg,
		ctx:    ctx,
		logger: sugar,
		client: kgo,
	}

	err := godotenv.Load()
	if err != nil {
		app.logger.Fatalw("Failed to load .env",
			"err", err)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.Router(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		app.logger.Infow("starting server:",
			"env", cfg.env,
			"addr", srv.Addr,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Fatal(err)
		}
	}()

	app.InitKafka()
	defer app.ShutdownKafka()

	app.logger.Info("starting consumer");
	for {
		fetches := app.client.PollFetches(app.ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			app.logger.panic(errs)
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			app.logger.Infow("consumed",
				"response", string(record.Value)
			)
		}
	}

	<-ctx.Done()
	app.logger.Info("Shutting down server..")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		app.logger.Fatalw("failed server shutdown:",
			"err", err)
	}

	logger.Info("Server gracefully exited.")
}

type config struct {
	env  string
	port int
}

type application struct {
	config config
	ctx    context.Context
	logger *zap.SugaredLogger
	client *kgo.Client
}

func newLogger() (*zap.Logger, error) {
	dir := "/app/logs"
	// TODO: local
	// dir := "/Users/tymalik/Docs/Git/lark-api/logs"
	fileName := "out.log"
	path := fmt.Sprintf("%s/%s", dir, fileName)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Println("failed creating log file:")
			return nil, fmt.Errorf("failed creating log dir: %w", err)
		}
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed opening log file: %w", err)
	}

	// TODO: set chicago time
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(file),
		zap.DebugLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger, nil
}
