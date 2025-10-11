package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/bootstrap"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/router"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/database"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/logger"
)

func main() {
	// Parse command line flags
	env := parseFlags()

	// Initialize logger
	logger.Setup(env)
	slog.Info("Starting application", "env", env)

	// Run application
	if err := run(env); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Application shutdown complete")
}

// parseFlags parses command line arguments
func parseFlags() string {
	env := flag.String("env", "local", "Environment (local|dev|production)")
	flag.Parse()
	return *env
}

// run contains the main application logic
func run(env string) error {
	// Create root context for application lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.Load(env)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	slog.Info("Configuration loaded successfully", "port", cfg.App.Port)

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}()

	// Setup server
	srv := setupServer(cfg, db)

	// Start server with graceful shutdown
	return startWithGracefulShutdown(ctx, srv, cfg.Server.GracefulTimeout)
}

// setupServer initializes and configures the HTTP server
func setupServer(cfg *config.Config, db *database.DB) *bootstrap.Server {
	// Bootstrap server with common setup
	boot := bootstrap.NewBootstrap(cfg)
	ginRouter := boot.SetupEngine()

	// Setup application-specific routes
	router.Setup(ginRouter, cfg, db)

	slog.Info("Server configured successfully",
		"port", cfg.App.Port,
		"env", cfg.App.Env,
	)

	return bootstrap.New(cfg, ginRouter)
}

// startWithGracefulShutdown starts the server and handles graceful shutdown
func startWithGracefulShutdown(ctx context.Context, srv *bootstrap.Server, gracefulTimeout time.Duration) error {
	// Channel to receive server errors
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		slog.Info("Server starting", "port", srv.Port())
		serverErrors <- srv.Start()
	}()

	// Channel to receive OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either server error or interrupt signal
	select {
	case err := <-serverErrors:
		// Server failed to start or stopped unexpectedly
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil

	case sig := <-quit:
		// Received shutdown signal
		slog.Info("Shutdown signal received", "signal", sig.String())

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(ctx, gracefulTimeout)
		defer cancel()

		// Attempt graceful shutdown
		slog.Info("Initiating graceful shutdown", "timeout", gracefulTimeout)
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server forced shutdown: %w", err)
		}

		slog.Info("Server shutdown gracefully")
		return nil
	}
}
