package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestConfig_DSN(t *testing.T) {
	cfg := Config{
		Host: "localhost",
		Port: 5432,
		User: "user",
		Pass: "pass",
		Name: "db",
	}
	expected := "postgres://user:pass@localhost:5432/db?sslmode=disable"
	assert.Equal(t, expected, cfg.DSN())
}

func TestNewPool(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		Host:    "localhost",
		Port:    5432,
		User:    "user",
		Pass:    "pass",
		Name:    "db",
		PoolMin: 2,
		PoolMax: 10,
	}

	t.Run("ParseConfig error", func(t *testing.T) {
		// Save original
		orig := pgxParseConfig
		defer func() { pgxParseConfig = orig }()

		expectedErr := errors.New("parse error")
		pgxParseConfig = func(s string) (*pgxpool.Config, error) {
			return nil, expectedErr
		}

		pool, err := NewPool(ctx, cfg)
		assert.Nil(t, pool)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("NewWithConfig error", func(t *testing.T) {
		origParse := pgxParseConfig
		origNew := pgxNewWithConfig
		defer func() {
			pgxParseConfig = origParse
			pgxNewWithConfig = origNew
		}()

		pgxParseConfig = func(s string) (*pgxpool.Config, error) {
			return &pgxpool.Config{ConnConfig: &pgx.ConnConfig{}}, nil
		}
		expectedErr := errors.New("new error")
		pgxNewWithConfig = func(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error) {
			return nil, expectedErr
		}

		pool, err := NewPool(ctx, cfg)
		assert.Nil(t, pool)
		assert.Equal(t, expectedErr, err)
	})
}
