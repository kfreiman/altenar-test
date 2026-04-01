package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"casino/internal/logger"
	"casino/internal/transactions/app/command"
	"casino/internal/transactions/domain"
)

func TestProcessTransactionHandler_Handle(t *testing.T) {
	ctx := context.Background()
	mockRepo := domain.NewMockRepository(t)
	mockLogger := logger.NewMockLogger(t)
	handler := command.NewProcessTransactionHandler(mockRepo, mockLogger)

	t.Run("success", func(t *testing.T) {
		id, _ := uuid.NewV7()
		cmd := command.ProcessTransaction{
			ID:              id.String(),
			UserID:          "user-1",
			TransactionType: "bet",
			Amount:          100,
			Timestamp:       time.Now(),
		}

		mockRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil).Once()

		err := handler.Handle(ctx, cmd)
		assert.NoError(t, err)
	})

	t.Run("invalid transaction type", func(t *testing.T) {
		id, _ := uuid.NewV7()
		cmd := command.ProcessTransaction{
			ID:              id.String(),
			UserID:          "user-1",
			TransactionType: "invalid",
			Amount:          100,
			Timestamp:       time.Now(),
		}

		mockLogger.EXPECT().ErrorContext(ctx, mock.Anything, []interface{}{"transaction_type", "invalid"}).Return().Once()

		err := handler.Handle(ctx, cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported transaction type")
	})

	t.Run("invalid amount", func(t *testing.T) {
		id, _ := uuid.NewV7()
		cmd := command.ProcessTransaction{
			ID:              id.String(),
			UserID:          "user-1",
			TransactionType: "bet",
			Amount:          -1,
			Timestamp:       time.Now(),
		}

		mockLogger.EXPECT().ErrorContext(ctx, mock.Anything, []interface{}{"cmd", cmd}).Return().Once()

		err := handler.Handle(ctx, cmd)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
	})

	t.Run("repo error", func(t *testing.T) {
		id, _ := uuid.NewV7()
		cmd := command.ProcessTransaction{
			ID:              id.String(),
			UserID:          "user-1",
			TransactionType: "bet",
			Amount:          100,
			Timestamp:       time.Now(),
		}

		repoErr := errors.New("database error")
		mockRepo.EXPECT().Save(ctx, mock.AnythingOfType("*domain.Transaction")).Return(repoErr).Once()
		mockLogger.EXPECT().ErrorContext(ctx, repoErr.Error(), mock.Anything).Return().Once()

		err := handler.Handle(ctx, cmd)
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
	})
}
