package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gophermart/internal/app/handlers"
	"gophermart/internal/app/services/worker"
	"gophermart/internal/infrastructure/accrual"
	"gophermart/internal/infrastructure/config"
	"gophermart/internal/infrastructure/storage"
)

const (
	serverReadTimeout     = 10 * time.Second
	serverWriteTimeout    = 10 * time.Second
	serverIdleTimeout     = 60 * time.Second
	serverShutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if cfg.DatabaseURI == "" {
		return errors.New("database URI is required")
	}

	ctx := context.Background()
	store, err := storage.NewStorage(ctx, cfg.DatabaseURI, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	defer store.Close()

	router := handlers.NewRouter(store, logger)

	server := &http.Server{
		Addr:         cfg.RunAddress,
		Handler:      router.Routes(logger),
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
		IdleTimeout:  serverIdleTimeout,
	}

	if cfg.AccrualSystemAddress != "" {
		accrualClient := accrual.NewClient(cfg.AccrualSystemAddress, logger)
		w := worker.NewWorker(store, accrualClient, logger)

		workerCtx, workerCancel := context.WithCancel(ctx)
		defer workerCancel()

		go w.Start(workerCtx)
	}

	go func() {
		logger.Info("starting server", "address", cfg.RunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(ctx, serverShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	logger.Info("server stopped")
	return nil
}
