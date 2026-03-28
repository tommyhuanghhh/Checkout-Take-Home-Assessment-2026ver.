package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"PaymentGateway/internal/pkg/config"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	
	cfg := config.Load()

	if err := run(logger, cfg); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, cfg *config.Config) error {
	// 1. HTTP Client for the Bank Simulator
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 2. Spin up In-Memory miniredis for zero-dependency idempotency
	mr, err := miniredis.Run()
	if err != nil {
		return fmt.Errorf("failed to start miniredis: %w", err)
	}
	defer mr.Close() // Ensure it cleans up when the app shuts down

	logger.Info("started in-memory miniredis", slog.String("addr", mr.Addr()))

	// Connect the actual go-redis client to the miniredis instance
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	// 3. Initialize the Dependency Graph (Composition Root via Wire)
	router, err := InitializeAPI(logger, redisClient, httpClient, cfg.BankSimulatorURL)
	if err != nil {
		return fmt.Errorf("failed to initialize api: %w", err)
	}

	// 4. Configure HTTP Server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 5. Graceful Shutdown Setup
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting server", slog.String("port", cfg.ServerPort))
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// 6. Block waiting for signal or server error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Info("graceful shutdown initiated", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			if err := srv.Close(); err != nil {
				return fmt.Errorf("could not stop server gracefully: %w", err)
			}
		}
	}

	logger.Info("server shutdown complete")
	return nil
}