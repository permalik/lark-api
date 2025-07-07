package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
)

type config struct {
	env  string
	port int
}

type application struct {
	config config
	ctx    context.Context
	logger *log.Logger
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env", err)
	}

	var cfg config
	flag.StringVar(&cfg.env, "env", "development", "environment (development|staging|production)")
	flag.IntVar(&cfg.port, "port", 5555, "Network port (default 5555)")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app := &application{
		config: cfg,
		ctx:    ctx,
		logger: logger,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.Router(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		logger.Printf("Starting %s server on port %s\n", cfg.env, srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	<-ctx.Done()
	logger.Println("Shutting down server..")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server shutdown failed:", err)
	}

	logger.Println("Server gracefully exited.")
}
