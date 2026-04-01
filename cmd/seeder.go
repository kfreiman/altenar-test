/*
Copyright © 2026 Casino
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
)

type TransactionMessage struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	TransactionType string    `json:"transaction_type"`
	Amount          int64     `json:"amount"`
	Timestamp       time.Time `json:"timestamp"`
}

// seederCmd represents the seeder command
var seederCmd = &cobra.Command{
	Use:   "seeder",
	Short: "Seed Kafka with transaction messages",
	Long: `Seed Kafka with transaction messages for testing.

The seeder generates transaction events with UUIDv7 identifiers
and sends them to the configured Kafka topic.`,
	RunE: runSeeder,
}

func runSeeder(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	// Parse flags
	kafkaURL, _ := cmd.Flags().GetString("kafka-url")
	kafkaTopic, _ := cmd.Flags().GetString("kafka-topic")
	count, _ := cmd.Flags().GetInt("count")
	userIDs, _ := cmd.Flags().GetStringSlice("user-ids")
	maxAmount, _ := cmd.Flags().GetInt("amount")

	// Set up Kafka writer
	writer := &kafka.Writer{
		Addr:         kafka.TCP(kafkaURL),
		Topic:        kafkaTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchBytes:   1024 * 1024,
		Async:        false,
		RequiredAcks: kafka.RequireAll,
	}
	defer func() {
		if err := writer.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close Kafka writer: %v\n", err)
		}
	}()

	// Seed random generator
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate and send messages
	messages := make([]kafka.Message, count)
	for i := 0; i < count; i++ {
		// Generate UUIDv7
		id := uuid.Must(uuid.NewV7()).String()

		// Random user ID
		userID := userIDs[rand.Intn(len(userIDs))]

		// Random transaction type
		txType := "bet"
		if rand.Intn(2) == 1 {
			txType = "win"
		}

		// Random amount
		amount := int64(rand.Intn(maxAmount) + 1)

		// Create transaction message
		msg := TransactionMessage{
			ID:              id,
			UserID:          userID,
			TransactionType: txType,
			Amount:          amount,
			Timestamp:       time.Now(),
		}

		// Marshal to JSON
		value, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		messages[i] = kafka.Message{
			Key:   []byte(id),
			Value: value,
		}
	}

	// Send messages to Kafka
	fmt.Printf("Generating %d messages and sending to %s (topic: %s)...\n", count, kafkaURL, kafkaTopic)
	if err := writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("failed to write messages to Kafka: %w", err)
	}

	fmt.Println("Seeding complete.")
	return nil
}

func init() {
	rootCmd.AddCommand(seederCmd)

	// Command flags
	seederCmd.Flags().String("kafka-url", "localhost:9092", "Kafka broker URL")
	seederCmd.Flags().String("kafka-topic", "transactions", "Kafka topic")
	seederCmd.Flags().IntP("count", "n", 10, "Number of messages to generate")
	seederCmd.Flags().StringSliceP("user-ids", "u", []string{"user1", "user2", "user3", "user4", "user5", "user6", "user7", "user8", "user9", "user10"}, "Comma-separated list of user IDs")
	seederCmd.Flags().IntP("amount", "a", 500, "Maximum amount for random generation")
}
