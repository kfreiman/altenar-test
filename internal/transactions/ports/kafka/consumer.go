package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"casino/internal/logger"
	"casino/internal/transactions/app"
	"casino/internal/transactions/app/command"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	URL     string `env:"URL" env-default:"localhost:9092" env-description:"Kafka broker URL" json:"url"`
	Topic   string `env:"TOPIC" env-default:"transactions" env-description:"Kafka topic" json:"topic"`
	GroupID string `env:"GROUP_ID" env-default:"consumer-group" env-description:"Kafka consumer group ID" json:"group_id"`
}

// MessageReader is an interface that wraps the kafka.Reader methods needed by Consumer.
// This allows for easier testing with mock implementations.
type MessageReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

type Consumer struct {
	app    app.Application
	cfg    Config
	reader MessageReader
	logger logger.Logger
}

func NewConsumer(cfg Config, app app.Application, logger logger.Logger) *Consumer {
	reader := newKafkaReader(cfg)
	return &Consumer{
		app:    app,
		cfg:    cfg,
		reader: reader,
		logger: logger,
	}
}

func newKafkaReader(cfg Config) MessageReader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.URL},
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
}

type transactionMessage struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	TransactionType string    `json:"transaction_type"`
	Amount          int64     `json:"amount"`
	Timestamp       time.Time `json:"timestamp"`
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			err = fmt.Errorf("failed to read message: %w", err)
			c.logger.ErrorContext(ctx, err.Error())
			return err
		}

		var msg transactionMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			c.logger.WarnContext(ctx, err.Error(), "message", string(m.Value))
			continue
		}

		cmd := command.ProcessTransaction{
			ID:              msg.ID,
			UserID:          msg.UserID,
			TransactionType: msg.TransactionType,
			Amount:          msg.Amount,
			Timestamp:       msg.Timestamp,
		}

		if err := c.app.Commands.ProcessTransaction.Handle(ctx, cmd); err != nil {
			c.logger.ErrorContext(ctx, err.Error())
			continue
		}

		c.logger.DebugContext(ctx, "message processed", "offset", m.Offset, "msg", msg)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
