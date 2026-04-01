//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"casino/internal/logger"
	"casino/internal/transactions/adapters/postgres"
	"casino/internal/transactions/adapters/postgres/db"
	"casino/internal/transactions/app"
	tkafka "casino/internal/transactions/ports/kafka"
	"casino/tests/helpers"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

func TestConsumerIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Setup Containers
	containers, err := helpers.SetupContainers(ctx, "../..")
	require.NoError(t, err)
	defer containers.Postgres.Terminate(ctx)
	defer containers.Kafka.Terminate(ctx)

	// 2. Connect to DB
	pgCfg := postgres.Config{
		Host:    containers.PostgresConfig.Host,
		Port:    containers.PostgresConfig.Port,
		User:    containers.PostgresConfig.User,
		Pass:    containers.PostgresConfig.Password,
		Name:    containers.PostgresConfig.Database,
		PoolMin: 2,
		PoolMax: 10,
	}
	dbPool, err := postgres.NewPool(ctx, pgCfg)
	require.NoError(t, err)
	defer dbPool.Close()

	// 3. Setup App and Consumer
	l := logger.NewNoop()
	queries := db.New(dbPool)
	repo := postgres.NewTransactionRepository(queries)
	application := app.New(repo, l)

	kafkaCfg := tkafka.Config{
		URL:     containers.KafkaURL,
		Topic:   "transactions-test",
		GroupID: "test-group",
	}
	consumer := tkafka.NewConsumer(kafkaCfg, application, l)

	// 4. Start Consumer in background
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	defer consumerCancel()

	go func() {
		if err := consumer.Run(consumerCtx); err != nil && consumerCtx.Err() == nil {
			t.Logf("Consumer stopped: %v", err)
		}
	}()

	// 5. Produce a message to Kafka
	var writer *kafka.Writer
	require.Eventually(t, func() bool {
		writer = &kafka.Writer{
			Addr:                   kafka.TCP(containers.KafkaURL),
			Topic:                  kafkaCfg.Topic,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		}
		err := writer.WriteMessages(ctx, kafka.Message{
			Value: []byte("test"),
		})
		return err == nil
	}, 20*time.Second, 1*time.Second)
	defer writer.Close()

	userID := "user-123"
	txID := uuid.Must(uuid.NewV7()).String()
	msg := map[string]interface{}{
		"id":               txID,
		"user_id":          userID,
		"transaction_type": "deposit",
		"amount":           1000,
		"timestamp":        time.Now().Format(time.RFC3339),
	}
	msgBytes, err := json.Marshal(msg)
	require.NoError(t, err)

	err = writer.WriteMessages(ctx, kafka.Message{
		Value: msgBytes,
	})
	require.NoError(t, err)

	// 6. Verify data in DB
	require.Eventually(t, func() bool {
		res, err := repo.List(ctx, &userID, nil, nil)
		if err != nil {
			return false
		}
		return len(res.Transactions) == 1 && res.Transactions[0].Amount() == 1000
	}, 15*time.Second, 500*time.Millisecond)
}
