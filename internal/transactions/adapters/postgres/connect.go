package postgres

import (
	"context"
	"fmt"
	"time"

	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrExitAfterRollback = fmt.Errorf("exit after rollback")

type Config struct {
	Port    int    `env:"PORT" env-default:"5432" env-description:"PostgreSQL port" json:"port"`
	Host    string `env:"HOST" env-description:"PostgreSQL host address" json:"host" validate:"required"`
	Name    string `env:"NAME" env-description:"PostgreSQL database name" json:"name" validate:"required"`
	User    string `env:"USER" env-description:"PostgreSQL username" json:"user" validate:"required"`
	Pass    string `env:"PASS" env-description:"PostgreSQL password" json:"-" validate:"required"`
	PoolMin int    `env:"POOL_MIN" env-default:"2" env-description:"Minimum pool connections" json:"pool_min"`
	PoolMax int    `env:"POOL_MAX" env-default:"10" env-description:"Maximum pool connections" json:"pool_max"`
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v?sslmode=disable",
		c.User,
		c.Pass,
		c.Host,
		c.Port,
		c.Name,
	)
}

// Injected for testing
var (
	pgxParseConfig   = pgxpool.ParseConfig
	pgxNewWithConfig = pgxpool.NewWithConfig
)

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn := cfg.DSN()
	poolConfig, err := pgxParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	poolConfig.MinConns = int32(cfg.PoolMin)
	poolConfig.MaxConns = int32(cfg.PoolMax)
	// Add some reasonable timeouts for the ping/initial connection
	poolConfig.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxNewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
