package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"casino/internal/logger"
	"casino/internal/transactions/domain"
)

type ProcessTransaction struct {
	ID              string
	UserID          string
	TransactionType string
	Amount          int64
	Timestamp       time.Time
}

type ProcessTransactionHandler struct {
	repo   domain.Repository
	logger logger.Logger
}

func NewProcessTransactionHandler(repo domain.Repository, logger logger.Logger) ProcessTransactionHandler {
	return ProcessTransactionHandler{repo: repo, logger: logger}
}

func (h ProcessTransactionHandler) Handle(ctx context.Context, cmd ProcessTransaction) error {
	tType, err := domain.ParseTransactionType(cmd.TransactionType)
	if err != nil {
		h.logger.ErrorContext(ctx, err.Error(), "transaction_type", cmd.TransactionType)
		return err
	}

	id, err := uuid.Parse(cmd.ID)
	if err != nil {
		return fmt.Errorf("invalid transaction id %q: %w", cmd.ID, err)
	}

	transaction, err := domain.NewTransaction(id, cmd.UserID, tType, cmd.Amount, cmd.Timestamp)
	if err != nil {
		h.logger.ErrorContext(ctx, err.Error(), "cmd", cmd)
		return err
	}

	err = h.repo.Save(ctx, transaction)
	if err != nil {
		h.logger.ErrorContext(ctx, err.Error(), "transaction", transaction)
	}

	return err
}
