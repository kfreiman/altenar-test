package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"casino/internal/logger"
	"casino/internal/transactions/adapters/postgres"
	"casino/internal/transactions/adapters/postgres/db"
	"casino/internal/transactions/app"
	thttp "casino/internal/transactions/ports/http"

	"github.com/spf13/cobra"
)

type apiConfig struct {
	Database postgres.Config `env-prefix:"DB_" json:"db"`
	HTTP     thttp.Config    `env-prefix:"HTTP_" json:"http"`
	Logger   logger.Config   `env-prefix:"LOG_" json:"log"`
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the HTTP API server",
	Long: `Start the HTTP API server for processing transactions.
The server listens on the port specified via environment variables.`,
	RunE: runApi,
}

func runApi(cmd *cobra.Command, _ []string) error {
	// Setup Signal Context
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// Load Config
	cfg, err := LoadConfig[apiConfig]()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize Logger
	log := logger.New(cfg.Logger)
	// Avoid logging 'cfg' directly if it contains DB passwords!
	log.InfoContext(ctx, "configuration loaded", "config", cfg)

	// Database Setup
	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	// Dependency Injection
	queries := db.New(pool)
	repo := postgres.NewTransactionRepository(queries)
	application := app.New(repo, log)
	httpServer := thttp.NewServer(application, cfg.HTTP)

	// 6. Start Server with Error Channel
	srvErr := make(chan error, 1)
	go func() {
		log.InfoContext(ctx, "Starting HTTP API server", "port", cfg.HTTP.Port)
		if err := httpServer.Run(); err != nil && err != http.ErrServerClosed {
			srvErr <- err
		}
	}()

	// Wait for Shutdown Signal or Server Error
	select {
	case err := <-srvErr:
		return fmt.Errorf("server startup failed: %w", err)
	case <-ctx.Done():
		log.InfoContext(ctx, "shutdown signal received")
	}

	// Graceful Shutdown
	// Use a fresh context for shutdown, don't use the one cancelled by the signal
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	if cfg.HTTP.ShutdownTimeout == 0 {
		// Fallback if timeout isn't set in config
		shutdownCtx, shutdownCancel = context.WithTimeout(context.Background(), 10*time.Second)
	}
	defer shutdownCancel()

	log.InfoContext(shutdownCtx, "shutting down HTTP server")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.InfoContext(shutdownCtx, "HTTP server stopped gracefully")
	return nil
}

func init() {
	rootCmd.AddCommand(apiCmd)
}
