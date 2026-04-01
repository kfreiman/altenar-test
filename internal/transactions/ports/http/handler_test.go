//go:build go1.22

package http

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"casino/internal/logger"
	"casino/internal/transactions/app"
	"casino/internal/transactions/domain"
	"casino/internal/transactions/ports/http/gen"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionHandler_HealthCheck(t *testing.T) {
	mockRepo := domain.NewMockRepository(t)
	mockLogger := logger.NewMockLogger(t)
	application := app.New(mockRepo, mockLogger)
	handler := NewTransactionHandler(application)

	resp, err := handler.HealthCheck(context.Background(), gen.HealthCheckRequestObject{})

	assert.NoError(t, err)
	assert.IsType(t, gen.HealthCheck200JSONResponse{}, resp)
	assert.Equal(t, "healthy", resp.(gen.HealthCheck200JSONResponse).Status)
}

func TestTransactionHandler_ListTransactions(t *testing.T) {
	userID := "user-123"
	tTypeStr := "bet"
	tType := gen.TransactionType(tTypeStr)
	domainType := domain.TransactionTypeBet

	now := time.Now().Truncate(time.Millisecond)
	// We need a valid UUIDv7 for the mock to return valid Transactions if we were to use NewTransaction
	// but here the handler just takes what the app layer returns.
	// However, domain.NewTransaction validates UUIDv7.
	// Let's generate a valid UUIDv7.
	id, _ := uuid.NewV7()

	tx, _ := domain.NewTransaction(id, userID, domainType, 100, now)

	tests := []struct {
		name          string
		request       gen.ListTransactionsRequestObject
		mockSetup     func(r *domain.MockRepository)
		expectedResp  interface{}
		expectedError error
	}{
		{
			name: "Success",
			request: gen.ListTransactionsRequestObject{
				Params: gen.ListTransactionsParams{
					UserId:          &userID,
					TransactionType: &tType,
				},
			},
			mockSetup: func(r *domain.MockRepository) {
				r.EXPECT().List(mock.Anything, &userID, &domainType, (*domain.Pagination)(nil)).
					Return(&domain.PageResult{
						Transactions: []*domain.Transaction{tx},
						HasMore:      false,
					}, nil)
			},
			expectedResp: gen.ListTransactions200JSONResponse{
				Transactions: []gen.Transaction{
					{
						Id:              id,
						UserId:          userID,
						TransactionType: gen.TransactionType(domainType.String()),
						Amount:          100,
						Timestamp:       now,
					},
				},
				HasMore: false,
			},
		},
		{
			name: "Invalid Transaction Type",
			request: gen.ListTransactionsRequestObject{
				Params: gen.ListTransactionsParams{
					TransactionType: (*gen.TransactionType)(ptr("invalid")),
				},
			},
			mockSetup: func(r *domain.MockRepository) {},
			expectedResp: gen.ListTransactions400JSONResponse{
				Code:    ErrCodeInvalidTransactionType,
				Message: "unsupported transaction type \"invalid\": transaction type must be 'bet' or 'win'",
			},
		},
		{
			name: "Internal Error",
			request: gen.ListTransactionsRequestObject{
				Params: gen.ListTransactionsParams{},
			},
			mockSetup: func(r *domain.MockRepository) {
				r.EXPECT().List(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("db error"))
			},
			expectedResp: gen.ListTransactions400JSONResponse{
				Code:    ErrCodeInternalError,
				Message: "db error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := domain.NewMockRepository(t)
			mockLogger := logger.NewMockLogger(t)
			application := app.New(mockRepo, mockLogger)
			handler := NewTransactionHandler(application)

			tt.mockSetup(mockRepo)

			resp, err := handler.ListTransactions(context.Background(), tt.request)

			assert.Equal(t, tt.expectedError, err)
			assert.Equal(t, tt.expectedResp, resp)
		})
	}
}

func TestServer_RunShutdown(t *testing.T) {
	mockRepo := domain.NewMockRepository(t)
	mockLogger := logger.NewMockLogger(t)
	application := app.New(mockRepo, mockLogger)

	cfg := Config{
		Port:            "0", // random port
		ShutdownTimeout: time.Second,
	}

	srv := NewServer(application, cfg)
	assert.NotNil(t, srv)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run()
	}()

	// Wait a bit to ensure the server has started
	// Using a small sleep is usually enough for local tests with port :0
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	assert.NoError(t, err)

	runErr := <-errCh
	assert.ErrorIs(t, runErr, http.ErrServerClosed)
}

func ptr[T any](v T) *T {
	return &v
}
