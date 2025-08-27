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
	fmt.Println("start")
	var cfg config
	flag.StringVar(&cfg.env, "env", "local", "environment (local|development|staging|production)")
	flag.IntVar(&cfg.port, "port", 4444, "Network port (default 4444)")
	flag.StringVar(&cfg.logDir, "logDir", "./logs", "Log directory (default ./logs)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger, err := newLogger(cfg.logDir)
	if err != nil {
		fmt.Println("failed to initialize logger: ", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	if err := godotenv.Load(cfg.env + ".env"); err != nil {
		sugar.Fatalw("failed to load .env", "err", err)
	}

	app := &application{
		config: cfg,
		ctx:    ctx,
		logger: sugar,
	}

	logger.Info("starting kafka")
	app.InitKafka()
	defer app.ShutdownKafka()

	go app.consumeLoop()

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

func (app *application) consumeLoop() {
	app.logger.Info("starting consumer")
	for {
		fetches := app.client.PollFetches(app.ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			app.logger.Panic(errs)
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			app.logger.Infow("consumed", "response", string(record.Value))
		}
	}
}

type config struct {
	env    string
	port   int
	logDir string
}

type application struct {
	config config
	ctx    context.Context
	logger *zap.SugaredLogger
	client *kgo.Client
}

func newLogger(logDir string) (*zap.Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating log dir: %w", err)
	}

	path := fmt.Sprintf("%s/out.log", logDir)

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
