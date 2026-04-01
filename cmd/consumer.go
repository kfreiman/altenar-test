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
	tkafka "casino/internal/transactions/ports/kafka"

	"github.com/spf13/cobra"
)

type consumerConfig struct {
	Database   postgres.Config `env-prefix:"DB_" json:"db"`
	Kafka      tkafka.Config   `env-prefix:"KAFKA_" json:"kafka"`
	Logger     logger.Config   `env-prefix:"LOG_" json:"log"`
	HealthPort string          `env:"HEALTH_PORT" env-default:"8081" json:"health_port"`
}

var consumerCmd = &cobra.Command{
	Use:   "consumer",
	Short: "Start the Kafka consumer",
	Long: `Start the Kafka consumer for processing transaction events.
The consumer connects to Kafka using KAFKA_URL and to the
database using DB_URL environment variables.`,
	RunE: runConsumer,
}

func runConsumer(cmd *cobra.Command, _ []string) error {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	cfg, err := LoadConfig[consumerConfig]()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := logger.New(cfg.Logger)
	log.InfoContext(ctx, "configuration loaded", "config", cfg)

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	queries := db.New(pool)
	repo := postgres.NewTransactionRepository(queries)
	application := app.New(repo, log)
	consumer := tkafka.NewConsumer(cfg.Kafka, application, log)

	healthSrv := &http.Server{
		Addr: ":" + cfg.HealthPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}),
	}

	go func() {
		log.InfoContext(ctx, "starting health check server", "port", cfg.HealthPort)
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.ErrorContext(ctx, "health check server failed", "error", err)
		}
	}()

	consumeErr := make(chan error, 1)
	go func() {
		log.InfoContext(ctx, "Starting Kafka consumer")
		if err := consumer.Run(ctx); err != nil {
			consumeErr <- err
		}
	}()

	select {
	case err := <-consumeErr:
		return fmt.Errorf("consumer error: %w", err)
	case <-ctx.Done():
		log.InfoContext(ctx, "shutdown signal received")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	log.InfoContext(shutdownCtx, "shutting down consumer")
	if err := consumer.Close(); err != nil {
		return fmt.Errorf("consumer shutdown failed: %w", err)
	}

	if err := healthSrv.Shutdown(shutdownCtx); err != nil {
		log.ErrorContext(shutdownCtx, "health check server shutdown failed", "error", err)
	}

	log.InfoContext(shutdownCtx, "consumer stopped gracefully")
	return nil
}

func init() {
	rootCmd.AddCommand(consumerCmd)
}
