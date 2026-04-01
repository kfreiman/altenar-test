package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTransactionType(t *testing.T) {
	tests := []struct {
		s       string
		want    TransactionType
		wantErr error
	}{
		{"bet", TransactionTypeBet, nil},
		{"win", TransactionTypeWin, nil},
		{"invalid", "", ErrInvalidType},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got, err := ParseTransactionType(tt.s)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTransaction_Methods(t *testing.T) {
	now := time.Now()
	id := generateValidUUIDv7()
	parsedID, err := uuid.Parse(id)
	require.NoError(t, err)

	tr, err := NewTransaction(parsedID, "u1", TransactionTypeBet, 100, now)
	require.NoError(t, err)

	assert.Equal(t, "u1", tr.UserID())
	assert.Equal(t, TransactionTypeBet, tr.Type())
	assert.Equal(t, int64(100), tr.Amount())
	assert.True(t, tr.Timestamp().Equal(now))
	assert.Equal(t, "bet", tr.Type().String())
}

func TestNewTransaction(t *testing.T) {
	id := generateValidUUIDv7()
	parsedID, err := uuid.Parse(id)
	require.NoError(t, err)

	tests := []struct {
		name            string
		id              uuid.UUID
		userID          string
		transactionType TransactionType
		amount          int64
		timestamp       time.Time
		wantErr         error
	}{
		{
			name:            "valid bet",
			id:              parsedID,
			userID:          "user-1",
			transactionType: TransactionTypeBet,
			amount:          100,
			timestamp:       time.Now(),
			wantErr:         nil,
		},
		{
			name:            "valid win",
			id:              parsedID,
			userID:          "user-2",
			transactionType: TransactionTypeWin,
			amount:          500,
			timestamp:       time.Now(),
			wantErr:         nil,
		},
		{
			name:            "invalid amount - zero",
			id:              parsedID,
			userID:          "user-1",
			transactionType: TransactionTypeBet,
			amount:          0,
			timestamp:       time.Now(),
			wantErr:         ErrInvalidAmount,
		},
		{
			name:            "invalid amount - negative",
			id:              parsedID,
			userID:          "user-1",
			transactionType: TransactionTypeBet,
			amount:          -100,
			timestamp:       time.Now(),
			wantErr:         ErrInvalidAmount,
		},
		{
			name:            "invalid type",
			id:              parsedID,
			userID:          "user-1",
			transactionType: "invalid",
			amount:          100,
			timestamp:       time.Now(),
			wantErr:         ErrInvalidType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTransaction(tt.id, tt.userID, tt.transactionType, tt.amount, tt.timestamp)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateUUIDv7(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "valid UUIDv7",
			id:      generateValidUUIDv7(),
			wantErr: nil,
		},
		{
			name:    "another valid UUIDv7",
			id:      generateValidUUIDv7(),
			wantErr: nil,
		},
		{
			name:    "invalid format - not a UUID",
			id:      "not-a-valid-uuid",
			wantErr: ErrInvalidUUIDv7,
		},
		{
			name:    "invalid format - empty string",
			id:      "",
			wantErr: ErrInvalidUUIDv7,
		},
		{
			name:    "invalid format - partial",
			id:      "0191e73c-5c45-7b80",
			wantErr: ErrInvalidUUIDv7,
		},
		{
			name:    "wrong version - UUIDv4",
			id:      generateUUIDv4(),
			wantErr: ErrInvalidUUIDVersion,
		},
		{
			name:    "wrong version - UUIDv1",
			id:      generateUUIDv1(),
			wantErr: ErrInvalidUUIDVersion,
		},
		{
			name:    "timestamp too old",
			id:      generateOldUUIDv7(),
			wantErr: ErrUUIDTimestampOutOfRange,
		},
		{
			name:    "timestamp too far in future",
			id:      generateFutureUUIDv7(),
			wantErr: ErrUUIDTimestampOutOfRange,
		},
		{
			name:    "invalid variant",
			id:      generateInvalidVariantUUIDv7(),
			wantErr: ErrInvalidUUIDv7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUIDv7(tt.id)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateUUIDv7_LargeSet(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := generateValidUUIDv7()
		err := ValidateUUIDv7(id)
		require.NoError(t, err, "valid UUIDv7 should pass: %s", id)
	}
}

func TestNewTransactionWithID(t *testing.T) {
	validID := generateValidUUIDv7()
	validUUID, err := uuid.Parse(validID)
	require.NoError(t, err)

	invalidID := generateUUIDv4()
	invalidUUID, err := uuid.Parse(invalidID)
	require.NoError(t, err)

	tests := []struct {
		name            string
		id              uuid.UUID
		userID          string
		transactionType TransactionType
		amount          int64
		timestamp       time.Time
		wantErr         error
	}{
		{
			name:            "valid transaction",
			id:              validUUID,
			userID:          "user-1",
			transactionType: TransactionTypeBet,
			amount:          1000,
			timestamp:       time.Now(),
			wantErr:         nil,
		},
		{
			name:            "wrong UUID version",
			id:              invalidUUID,
			userID:          "user-1",
			transactionType: TransactionTypeWin,
			amount:          500,
			timestamp:       time.Now(),
			wantErr:         ErrInvalidUUIDVersion,
		},
		{
			name:            "invalid amount",
			id:              validUUID,
			userID:          "user-1",
			transactionType: TransactionTypeBet,
			amount:          0,
			timestamp:       time.Now(),
			wantErr:         ErrInvalidAmount,
		},
		{
			name:            "invalid type",
			id:              validUUID,
			userID:          "user-1",
			transactionType: "invalid",
			amount:          100,
			timestamp:       time.Now(),
			wantErr:         ErrInvalidType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewTransaction(tt.id, tt.userID, tt.transactionType, tt.amount, tt.timestamp)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.id, tr.ID())
			assert.Equal(t, tt.userID, tr.UserID())
			assert.Equal(t, tt.transactionType, tr.Type())
			assert.Equal(t, tt.amount, tr.Amount())
			assert.True(t, tt.timestamp.Equal(tr.Timestamp()))
		})
	}
}

func TestTransaction_SetID(t *testing.T) {
	initialID, err := uuid.Parse(generateValidUUIDv7())
	require.NoError(t, err)

	tx, err := NewTransaction(initialID, "user-1", TransactionTypeBet, 100, time.Now())
	require.NoError(t, err)

	t.Run("valid UUIDv7", func(t *testing.T) {
		validID, err := uuid.Parse(generateValidUUIDv7())
		require.NoError(t, err)
		err = tx.SetID(validID)
		require.NoError(t, err)
		assert.Equal(t, validID, tx.ID())
	})

	t.Run("invalid UUIDv7 - wrong version", func(t *testing.T) {
		invalidID, err := uuid.Parse(generateUUIDv4())
		require.NoError(t, err)
		err = tx.SetID(invalidID)
		require.ErrorIs(t, err, ErrInvalidUUIDVersion)
	})

	t.Run("invalid UUIDv7 - timestamp out of range", func(t *testing.T) {
		oldID, err := uuid.Parse(generateOldUUIDv7())
		require.NoError(t, err)
		err = tx.SetID(oldID)
		require.ErrorIs(t, err, ErrUUIDTimestampOutOfRange)
	})
}

func TestExtractUUIDv7Timestamp(t *testing.T) {
	testID := generateValidUUIDv7()
	parsed, err := uuid.Parse(testID)
	require.NoError(t, err)

	extracted := extractUUIDv7Timestamp(parsed)

	now := time.Now()
	assert.WithinDuration(t, now, extracted, MaxTimestampDrift)
}

func TestValidateUUIDv7_IDCaseInsensitive(t *testing.T) {
	upperID := strings.ToUpper(generateValidUUIDv7())
	err := ValidateUUIDv7(upperID)
	assert.NoError(t, err)
}

func TestNewPagination(t *testing.T) {
	cursor := &Cursor{ID: uuid.MustParse(generateValidUUIDv7())}

	tests := []struct {
		name     string
		cursor   *Cursor
		pageSize int
		wantSize int
	}{
		{
			name:     "valid pageSize",
			cursor:   cursor,
			pageSize: 20,
			wantSize: 20,
		},
		{
			name:     "nil cursor",
			cursor:   nil,
			pageSize: 20,
			wantSize: 20,
		},
		{
			name:     "zero pageSize",
			cursor:   cursor,
			pageSize: 0,
			wantSize: DefaultPageSize,
		},
		{
			name:     "negative pageSize",
			cursor:   cursor,
			pageSize: -1,
			wantSize: DefaultPageSize,
		},
		{
			name:     "exceeds MaxPageSize",
			cursor:   cursor,
			pageSize: MaxPageSize + 1,
			wantSize: MaxPageSize,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPagination(tt.cursor, tt.pageSize)
			assert.Equal(t, tt.cursor, got.Cursor)
			assert.Equal(t, tt.wantSize, got.PageSize)
		})
	}
}

func generateValidUUIDv7() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic("failed to generate UUIDv7: " + err.Error())
	}
	return id.String()
}

func generateUUIDv4() string {
	id := uuid.New()
	id[6] = (id[6] & 0x0f) | 0x40 // Set version to 4
	id[8] = (id[8] & 0x3f) | 0x80 // Set variant to RFC4122
	return id.String()
}

func generateUUIDv1() string {
	id := uuid.New()
	id[6] = (id[6] & 0x0f) | 0x10 // Set version to 1
	id[8] = (id[8] & 0x3f) | 0x80 // Set variant to RFC4122
	return id.String()
}

func generateOldUUIDv7() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic("failed to generate UUIDv7: " + err.Error())
	}
	// Set timestamp to year 2000 to make it too old
	id[0] = 0x00
	id[1] = 0x00
	id[2] = 0x00
	id[3] = 0x00
	id[4] = 0x00
	id[5] = 0x00
	return id.String()
}

func generateFutureUUIDv7() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic("failed to generate UUIDv7: " + err.Error())
	}
	// Set timestamp far in the future (max possible value)
	id[0] = 0x7f
	id[1] = 0xff
	id[2] = 0xff
	id[3] = 0xff
	id[4] = 0xff
	id[5] = 0xff
	return id.String()
}

func generateInvalidVariantUUIDv7() string {
	id, err := uuid.NewV7()
	if err != nil {
		panic("failed to generate UUIDv7: " + err.Error())
	}
	// Set variant to non-RFC4122 (clear bits 6-7 to 00xx)
	id[8] = id[8] & 0x3f
	return id.String()
}
