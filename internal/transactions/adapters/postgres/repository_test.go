package postgres

import (
	"context"
	"testing"
	"time"

	"casino/internal/transactions/adapters/postgres/db"
	"casino/internal/transactions/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionRepository_Save(t *testing.T) {
	mockDB := &db.MockDBTX{}
	queries := db.New(mockDB)
	repo := NewTransactionRepository(queries)

	ctx := context.Background()
	id, _ := uuid.NewV7()
	userID := "user-123"
	now := time.Now().Truncate(time.Millisecond)

	t.Run("Success with existing ID", func(t *testing.T) {
		txn, _ := domain.NewTransaction(id, userID, domain.TransactionTypeBet, 100, now)

		mockDB.On("Exec", ctx, mock.Anything,
			mock.MatchedBy(func(args []interface{}) bool {
				if len(args) != 5 {
					return false
				}
				// Verify UserID, Type, Amount (ignoring ID and CreatedAt as they might be hard to match exactly due to types)
				return args[1] == userID && args[2] == "bet" && args[3] == int64(100)
			}),
		).Return(pgconn.NewCommandTag("INSERT 1"), nil).Once()

		err := repo.Save(ctx, txn)
		assert.NoError(t, err)
	})
}

func TestTransactionRepository_List(t *testing.T) {
	mockDB := &db.MockDBTX{}
	queries := db.New(mockDB)
	repo := NewTransactionRepository(queries)

	ctx := context.Background()

	t.Run("Success with filters", func(t *testing.T) {
		userID := "1d466472-a675-2388-fcc2-e470f53a7cbf"
		tType := domain.TransactionTypeBet
		pagination := &domain.Pagination{PageSize: 10}

		mockRows := &db.MockRows{}
		mockDB.On("Query", ctx, mock.Anything, mock.MatchedBy(func(args []interface{}) bool {
			// ListAllTransactionsParams order: CursorID, UserID, TransactionType, PageSize
			return len(args) == 4 && args[3] == int32(11) // limit is pageSize + 1
		})).Return(mockRows, nil).Once()

		id1, _ := uuid.NewV7()
		id2, _ := uuid.NewV7()

		// Mock two rows
		mockRows.On("Next").Return(true).Twice()
		mockRows.On("Next").Return(false)
		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			destID := args.Get(0).(*pgtype.UUID)
			destUserID := args.Get(1).(*string)
			destType := args.Get(2).(*string)
			destAmount := args.Get(3).(*int64)
			destCreatedAt := args.Get(4).(*pgtype.Timestamptz)

			*destID = pgtype.UUID{Bytes: id1, Valid: true}
			*destUserID = userID
			*destType = "bet"
			*destAmount = int64(100)
			*destCreatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		}).Return(nil).Once()

		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			destID := args.Get(0).(*pgtype.UUID)
			destUserID := args.Get(1).(*string)
			destType := args.Get(2).(*string)
			destAmount := args.Get(3).(*int64)
			destCreatedAt := args.Get(4).(*pgtype.Timestamptz)

			*destID = pgtype.UUID{Bytes: id2, Valid: true}
			*destUserID = userID
			*destType = "bet"
			*destAmount = int64(200)
			*destCreatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		}).Return(nil).Once()

		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return()

		result, err := repo.List(ctx, &userID, &tType, pagination)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Transactions, 2)
		assert.False(t, result.HasMore, "HasMore should be false when number of rows is not page size + 1")
		assert.Equal(t, id2, result.Transactions[1].ID())
	})

	t.Run("Has more results", func(t *testing.T) {
		pagination := &domain.Pagination{PageSize: 1}
		mockRows := &db.MockRows{}
		mockDB.On("Query", ctx, mock.Anything, mock.MatchedBy(func(args []interface{}) bool {
			return len(args) == 4 && args[3] == int32(2)
		})).Return(mockRows, nil).Once()

		id1, _ := uuid.NewV7()
		id2, _ := uuid.NewV7()

		mockRows.On("Next").Return(true).Twice()
		mockRows.On("Next").Return(false)
		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: id1, Valid: true}
			*args.Get(1).(*string) = "user-1"
			*args.Get(2).(*string) = "bet"
			*args.Get(3).(*int64) = int64(100)
			*args.Get(4).(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		}).Return(nil).Once()
		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: id2, Valid: true}
			*args.Get(1).(*string) = "user-1"
			*args.Get(2).(*string) = "bet"
			*args.Get(3).(*int64) = int64(200)
			*args.Get(4).(*pgtype.Timestamptz) = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		}).Return(nil).Once()
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return()

		result, err := repo.List(ctx, nil, nil, pagination)

		assert.NoError(t, err)
		assert.Len(t, result.Transactions, 1)
		assert.True(t, result.HasMore)
		assert.NotNil(t, result.NextCursor)
		assert.Equal(t, id1, result.NextCursor.ID)
	})
}
