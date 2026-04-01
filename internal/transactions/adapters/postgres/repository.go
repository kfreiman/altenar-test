package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"casino/internal/transactions/adapters/postgres/db"
	"casino/internal/transactions/domain"
)

type TransactionRepository struct {
	queries *db.Queries
}

func NewTransactionRepository(queries *db.Queries) *TransactionRepository {
	return &TransactionRepository{queries: queries}
}

func (r *TransactionRepository) Save(ctx context.Context, t *domain.Transaction) error {
	if t.ID() == uuid.Nil {
		newID := uuid.New()
		if err := t.SetID(newID); err != nil {
			return fmt.Errorf("failed to set transaction ID: %w", err)
		}
	}

	return r.queries.CreateTransaction(ctx, db.CreateTransactionParams{
		ID:              pgtype.UUID{Bytes: t.ID(), Valid: true},
		UserID:          t.UserID(),
		TransactionType: t.Type().String(),
		Amount:          t.Amount(),
		CreatedAt:       pgtype.Timestamptz{Time: t.Timestamp(), Valid: true},
	})
}

func (r *TransactionRepository) List(ctx context.Context, userID *string, transactionType *domain.TransactionType, pagination *domain.Pagination) (*domain.PageResult, error) {
	pageSize := int32(domain.DefaultPageSize)
	if pagination != nil && pagination.PageSize > 0 {
		pageSize = int32(pagination.PageSize)
	}

	limit := pageSize + 1

	cursorID := uuid.Nil
	if pagination != nil && pagination.Cursor != nil && pagination.Cursor.ID != uuid.Nil {
		cursorID = pagination.Cursor.ID
	}
	cursorUUID := pgtype.UUID{Bytes: cursorID, Valid: true}

	var userIDParam pgtype.Text
	if userID != nil {
		userIDParam = pgtype.Text{String: *userID, Valid: true}
	}

	var typeParam pgtype.Text
	if transactionType != nil {
		typeParam = pgtype.Text{String: string(*transactionType), Valid: true}
	}

	rows, err := r.queries.ListAllTransactions(ctx, db.ListAllTransactionsParams{
		CursorID:        cursorUUID,
		UserID:          userIDParam,
		TransactionType: typeParam,
		PageSize:        limit,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return &domain.PageResult{
				Transactions: []*domain.Transaction{},
				NextCursor:   nil,
				HasMore:      false,
			}, nil
		}
		return nil, err
	}

	transactions, lastID, hasMore, err := processTransactionRows(convertToTxRows(rows), limit)
	if err != nil {
		return nil, err
	}

	var nextCursor *domain.Cursor
	if hasMore && lastID.Valid {
		nextCursor = &domain.Cursor{
			ID: uuid.UUID(lastID.Bytes),
		}
	}

	return &domain.PageResult{
		Transactions: transactions,
		NextCursor:   nextCursor,
		HasMore:      hasMore,
	}, nil
}

func convertToTxRows(rows []*db.ListAllTransactionsRow) []*txRow {
	result := make([]*txRow, len(rows))
	for i, r := range rows {
		result[i] = &txRow{
			ID:              r.ID,
			UserID:          r.UserID,
			TransactionType: r.TransactionType,
			Amount:          r.Amount,
			CreatedAt:       r.CreatedAt,
		}
	}
	return result
}

type txRow struct {
	ID              pgtype.UUID
	UserID          string
	TransactionType string
	Amount          int64
	CreatedAt       pgtype.Timestamptz
}

func processTransactionRows(rows []*txRow, limit int32) ([]*domain.Transaction, pgtype.UUID, bool, error) {
	if len(rows) == 0 {
		return []*domain.Transaction{}, pgtype.UUID{}, false, nil
	}

	hasMore := int32(len(rows)) == limit
	count := len(rows)
	if hasMore {
		count--
	}

	transactions := make([]*domain.Transaction, 0, count)

	var lastID pgtype.UUID

	for i := 0; i < count; i++ {
		row := rows[i]
		txn, err := domain.NewTransaction(
			uuid.UUID(row.ID.Bytes),
			row.UserID,
			domain.TransactionType(row.TransactionType),
			row.Amount,
			row.CreatedAt.Time,
		)
		if err != nil {
			continue
		}
		transactions = append(transactions, txn)
		lastID = row.ID
	}

	return transactions, lastID, hasMore, nil
}
