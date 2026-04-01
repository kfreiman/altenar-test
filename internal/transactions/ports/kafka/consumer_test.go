package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"casino/internal/transactions/app"
	"casino/internal/transactions/domain"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"io"
)

func TestConsumer_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nopLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("successfully process message", func(t *testing.T) {
		mockReader := NewMockMessageReader(t)
		mockRepo := domain.NewMockRepository(t)
		
		application := app.New(mockRepo, nopLogger)
		
		consumer := &Consumer{
			app:    application,
			reader: mockReader,
			logger: nopLogger,
		}

		msg := transactionMessage{
			UserID:          "user-1",
			TransactionType: "bet",
			Amount:          100,
			Timestamp:       time.Now().UTC(),
		}
		data, _ := json.Marshal(msg)

		mockReader.EXPECT().ReadMessage(mock.Anything).Return(kafka.Message{
			Value:  data,
			Offset: 123,
		}, nil).Once()

		// After the first successful message, we make the next call return an error to break the loop
		mockReader.EXPECT().ReadMessage(mock.Anything).Return(kafka.Message{}, errors.New("stop")).Once()

		mockRepo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(tr *domain.Transaction) bool {
			return tr.UserID() == msg.UserID && 
				tr.Type() == domain.TransactionTypeBet && 
				tr.Amount() == msg.Amount
		})).Return(nil).Once()

		err := consumer.Run(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stop")
	})

	t.Run("skip invalid JSON", func(t *testing.T) {
		mockReader := NewMockMessageReader(t)
		mockRepo := domain.NewMockRepository(t)
		
		application := app.New(mockRepo, nopLogger)
		
		consumer := &Consumer{
			app:    application,
			reader: mockReader,
			logger: nopLogger,
		}

		mockReader.EXPECT().ReadMessage(mock.Anything).Return(kafka.Message{
			Value: []byte("invalid json"),
		}, nil).Once()

		mockReader.EXPECT().ReadMessage(mock.Anything).Return(kafka.Message{}, errors.New("stop")).Once()

		err := consumer.Run(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stop")
		// mockRepo.Save should NOT be called
	})

	t.Run("skip on handler error", func(t *testing.T) {
		mockReader := NewMockMessageReader(t)
		mockRepo := domain.NewMockRepository(t)
		
		application := app.New(mockRepo, nopLogger)
		
		consumer := &Consumer{
			app:    application,
			reader: mockReader,
			logger: nopLogger,
		}

		msg := transactionMessage{
			UserID:          "user-1",
			TransactionType: "invalid-type", // This will cause Handle to return error
			Amount:          100,
			Timestamp:       time.Now().UTC(),
		}
		data, _ := json.Marshal(msg)

		mockReader.EXPECT().ReadMessage(mock.Anything).Return(kafka.Message{
			Value: data,
		}, nil).Once()

		mockReader.EXPECT().ReadMessage(mock.Anything).Return(kafka.Message{}, errors.New("stop")).Once()

		err := consumer.Run(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stop")
		// mockRepo.Save should NOT be called because of invalid transaction type
	})
	
	t.Run("return error on ReadMessage error", func(t *testing.T) {
		mockReader := NewMockMessageReader(t)
		mockRepo := domain.NewMockRepository(t)
		
		application := app.New(mockRepo, nopLogger)
		
		consumer := &Consumer{
			app:    application,
			reader: mockReader,
			logger: nopLogger,
		}

		expectedErr := errors.New("kafka connection lost")
		mockReader.EXPECT().ReadMessage(mock.Anything).Return(kafka.Message{}, expectedErr).Once()

		err := consumer.Run(ctx)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expectedErr))
	})
}

func TestConsumer_Close(t *testing.T) {
	mockReader := NewMockMessageReader(t)
	consumer := &Consumer{
		reader: mockReader,
	}

	mockReader.EXPECT().Close().Return(nil).Once()
	err := consumer.Close()
	assert.NoError(t, err)
}
