package query

import (
	"context"

	"casino/internal/logger"
	"casino/internal/transactions/domain"
)

type ListTransactions struct {
	UserID          *string
	TransactionType *domain.TransactionType
	Pagination      *domain.Pagination
}

type ListTransactionsHandler struct {
	repo   domain.Repository
	logger logger.Logger
}

func NewListTransactionsHandler(repo domain.Repository, logger logger.Logger) ListTransactionsHandler {
	return ListTransactionsHandler{repo: repo, logger: logger}
}

func (h ListTransactionsHandler) Handle(ctx context.Context, q ListTransactions) (*domain.PageResult, error) {
	return h.repo.List(ctx, q.UserID, q.TransactionType, q.Pagination)
}
