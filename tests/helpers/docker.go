package helpers

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainers struct {
	Postgres       *postgres.PostgresContainer
	Kafka          *kafka.KafkaContainer
	PostgresConfig PostgresConfig
	KafkaURL       string
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Database)
}

func SetupContainers(ctx context.Context, projectRoot string) (*TestContainers, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("casino"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres: %w", err)
	}

	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")

	pgConfig := PostgresConfig{
		Host:     pgHost,
		Port:     pgPort.Int(),
		User:     "user",
		Password: "password",
		Database: "casino",
	}

	// Run migrations
	u, _ := url.Parse(pgConfig.DSN())
	db := dbmate.New(u)
	db.MigrationsDir = []string{filepath.Join(projectRoot, "internal/transactions/adapters/postgres/migrations")}
	db.SchemaFile = filepath.Join(projectRoot, "internal/transactions/adapters/postgres/schema.sql")
	db.WaitBefore = true
	db.Log = io.Discard

	if err := db.CreateAndMigrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/cp-kafka:7.7.8",
		kafka.WithClusterID("casino-cluster"),
		testcontainers.WithEnv(map[string]string{
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
		}),
	)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to start kafka: %w", err)
	}

	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		pgContainer.Terminate(ctx)
		kafkaContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get kafka brokers: %w", err)
	}

	return &TestContainers{
		Postgres:       pgContainer,
		Kafka:          kafkaContainer,
		PostgresConfig: pgConfig,
		KafkaURL:       brokers[0],
	}, nil
}
