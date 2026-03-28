package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Setup Structured Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// ==========================================
	// 1. HTTP Client Connection Pool Setup
	// ==========================================
	customTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second, // Explicitly enables TCP Keep-Alive
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,              // Total idle connections across all hosts
		MaxIdleConnsPerHost:   100,              // Max idle connections per specific host (e.g., our acquiring bank)
		IdleConnTimeout:       90 * time.Second, // How long an idle connection stays in the pool before closing
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   15 * time.Second, // Overall request timeout (including reading response)
		Transport: customTransport,
	}

	// ==========================================
	// 2. Redis Connection Pool Setup
	// ==========================================
	redisClient := redis.NewClient(&redis.Options{
		Addr:            "localhost:6379", // Point this to your miniredis/local redis
		PoolSize:        50,               // Maximum number of socket connections
		MinIdleConns:    10,               // Minimum number of idle connections to keep open
		ConnMaxIdleTime: 5 * time.Minute,  // How long a connection can be idle before closing
		ConnMaxLifetime: 1 * time.Hour,    // Absolute max lifetime of a connection
	})

	// Verify Redis connection on startup
	startupCtx, startupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer startupCancel()
	if err := redisClient.Ping(startupCtx).Err(); err != nil {
		return errors.New("failed to connect to redis: " + err.Error())
	}
	logger.Info("Successfully connected to Redis")

	// ==========================================
	// 3. Dependency Wiring (Wire package)
	// ==========================================
	bankSimulatorURL := "http://localhost:8080"
	router := InitializeRouter(logger, redisClient, httpClient, bankSimulatorURL)

	// ==========================================
	// 4. Server Configuration & Startup
	// ==========================================
	srv := &http.Server{
		Addr:         ":8090",
		Handler:      router,
		ReadTimeout:  10 * time.Second, // Defend against slow-loris attacks
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for OS shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine so it doesn't block
	go func() {
		logger.Info("Starting Payment Gateway API", "port", 8090)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// ==========================================
	// 5. Graceful Shutdown Sequence
	// ==========================================
	<-quit // Block here until we receive SIGTERM or SIGINT
	logger.Info("Shutdown signal received, initiating graceful shutdown...")

	// Give active requests 10 seconds to finish before force-closing
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Cleanup Redis connection pool
	if err := redisClient.Close(); err != nil {
		logger.Error("failed to cleanly close redis pool", "error", err)
	}

	// Shutdown HTTP Server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return errors.New("server forced to shutdown: " + err.Error())
	}

	logger.Info("Server exited gracefully")
	return nil
}
