package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAmount           = errors.New("transaction amount must be positive")
	ErrInvalidType             = errors.New("transaction type must be 'bet' or 'win'")
	ErrInvalidUUIDv7           = errors.New("must be valid UUID v7 format")
	ErrInvalidUUIDVersion      = errors.New("must be version 7")
	ErrUUIDTimestampOutOfRange = errors.New("timestamp out of allowed range")
)

const DefaultPageSize = 50
const MaxPageSize = 100

const (
	MaxTimestampDrift = 24 * time.Hour
	UUIDv7Version     = 7
)

type Cursor struct {
	ID uuid.UUID
}

type Pagination struct {
	Cursor   *Cursor
	PageSize int
}

func NewPagination(cursor *Cursor, pageSize int) *Pagination {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return &Pagination{
		Cursor:   cursor,
		PageSize: pageSize,
	}
}

type PageResult struct {
	Transactions []*Transaction
	NextCursor   *Cursor
	HasMore      bool
}

type TransactionType string

const (
	TransactionTypeBet TransactionType = "bet"
	TransactionTypeWin TransactionType = "win"
)

func (t TransactionType) String() string {
	return string(t)
}

func ValidateUUIDv7(id string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid uuidv7 %q: %w", id, ErrInvalidUUIDv7)
	}
	return parseAndValidateUUIDv7(parsed)
}

func parseAndValidateUUIDv7(id uuid.UUID) error {
	// Validate RFC4122 variant (bits 6-7 should be 10xx)
	if (id[8] & 0xc0) != 0x80 {
		return fmt.Errorf("invalid uuidv7 variant: %w", ErrInvalidUUIDv7)
	}

	if id.Version() != UUIDv7Version {
		return fmt.Errorf("parsed version %d: %w", id.Version(), ErrInvalidUUIDVersion)
	}

	uuidTimestamp := extractUUIDv7Timestamp(id)
	now := time.Now()

	if uuidTimestamp.Before(now.Add(-MaxTimestampDrift)) || uuidTimestamp.After(now.Add(MaxTimestampDrift)) {
		var drift time.Duration
		if uuidTimestamp.Before(now.Add(-MaxTimestampDrift)) {
			drift = now.Sub(uuidTimestamp) - MaxTimestampDrift
		} else {
			drift = uuidTimestamp.Sub(now) - MaxTimestampDrift
		}
		return fmt.Errorf("check timestamp drift of %v, got: %w", drift, ErrUUIDTimestampOutOfRange)
	}

	return nil
}

func extractUUIDv7Timestamp(u uuid.UUID) time.Time {
	timestampMs := (int64(u[0]) << 40) | (int64(u[1]) << 32) | (int64(u[2]) << 24) |
		(int64(u[3]) << 16) | (int64(u[4]) << 8) | int64(u[5])
	return time.UnixMilli(timestampMs)
}

func ParseTransactionType(s string) (TransactionType, error) {
	switch s {
	case "bet":
		return TransactionTypeBet, nil
	case "win":
		return TransactionTypeWin, nil
	default:
		return "", fmt.Errorf("unsupported transaction type %q: %w", s, ErrInvalidType)
	}
}

type Transaction struct {
	id              uuid.UUID
	userID          string
	transactionType TransactionType
	amount          int64 // in smallest currency unit (e.g. cents)
	timestamp       time.Time
}

func NewTransaction(id uuid.UUID, userID string, transactionType TransactionType, amount int64, timestamp time.Time) (*Transaction, error) {
	if err := parseAndValidateUUIDv7(id); err != nil {
		return nil, fmt.Errorf("invalid transaction id %q: %w", id, err)
	}

	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive, got %d: %w", amount, ErrInvalidAmount)
	}

	if transactionType != TransactionTypeBet && transactionType != TransactionTypeWin {
		return nil, fmt.Errorf("unsupported transaction type %q: %w", transactionType, ErrInvalidType)
	}

	return &Transaction{
		id:              id,
		userID:          userID,
		transactionType: transactionType,
		amount:          amount,
		timestamp:       timestamp,
	}, nil
}

func (t *Transaction) ID() uuid.UUID {
	return t.id
}

func (t *Transaction) SetID(id uuid.UUID) error {
	if err := parseAndValidateUUIDv7(id); err != nil {
		return fmt.Errorf("invalid transaction id %q: %w", id, err)
	}
	t.id = id
	return nil
}



func (t *Transaction) UserID() string {
	return t.userID
}

func (t *Transaction) Type() TransactionType {
	return t.transactionType
}

func (t *Transaction) Amount() int64 {
	return t.amount
}

func (t *Transaction) Timestamp() time.Time {
	return t.timestamp
}

type Repository interface {
	Save(ctx context.Context, t *Transaction) error
	List(ctx context.Context, userID *string, transactionType *TransactionType, pagination *Pagination) (*PageResult, error)
}
