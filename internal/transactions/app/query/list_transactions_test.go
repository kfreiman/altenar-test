package query_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"casino/internal/logger"
	"casino/internal/transactions/app/query"
	"casino/internal/transactions/domain"
)

func TestListTransactionsHandler_Handle(t *testing.T) {
	ctx := context.Background()
	mockRepo := domain.NewMockRepository(t)
	mockLogger := logger.NewMockLogger(t)
	handler := query.NewListTransactionsHandler(mockRepo, mockLogger)

	t.Run("success", func(t *testing.T) {
		userID := "user-123"
		tType := domain.TransactionTypeBet
		pagination := domain.NewPagination(nil, 10)
		q := query.ListTransactions{
			UserID:          &userID,
			TransactionType: &tType,
			Pagination:      pagination,
		}

		expectedResult := &domain.PageResult{
			Transactions: []*domain.Transaction{},
			HasMore:      false,
		}

		mockRepo.EXPECT().List(ctx, &userID, &tType, pagination).Return(expectedResult, nil).Once()

		result, err := handler.Handle(ctx, q)

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("repository error", func(t *testing.T) {
		q := query.ListTransactions{}
		repoErr := errors.New("repository error")

		mockRepo.EXPECT().List(ctx, (*string)(nil), (*domain.TransactionType)(nil), (*domain.Pagination)(nil)).Return(nil, repoErr).Once()

		result, err := handler.Handle(ctx, q)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoErr, err)
	})
}
