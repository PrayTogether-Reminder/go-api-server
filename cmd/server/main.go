package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/delivery/http/routes"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/infrastructure/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/infrastructure/database"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/infrastructure/server"
)

func main() {
	// Parse command line flags
	var env string
	flag.StringVar(&env, "env", "local", "Environment (local|dev|prod)")
	flag.Parse()

	// Initialize structured logger
	setupLogger(env)

	// Load configuration
	cfg, err := config.Load(env)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}()

	// Bootstrap server with common setup (Clean Architecture: no DB in bootstrap)
	bootstrap := server.NewBootstrap(cfg)
	router := bootstrap.SetupEngine()

	// Setup application-specific routes
	routes.Setup(router, cfg, db)

	// Create and start server
	srv := server.New(cfg, router)

	// Channel to receive server errors
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		slog.Info("Server starting",
			"environment", env,
			"port", cfg.App.Port,
		)
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
			slog.Error("Server error", "error", err)
			// Still perform cleanup via deferred functions
			return
		}
	case sig := <-quit:
		// Received shutdown signal
		slog.Info("Shutting down server", "signal", sig.String())
	}

	// Graceful shutdown with timeout (only if we received a signal)
	// If server errored on startup, it's already stopped
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		// Don't call os.Exit here - let deferred functions run
	}

	slog.Info("Server shutdown complete")
}

// setupLogger configures the global slog logger based on environment
func setupLogger(env string) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if env == "prod" {
		// Production: JSON format, error level
		opts.Level = slog.LevelError
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Development: Text format, debug level
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
