//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"casino/internal/logger"
	"casino/internal/transactions/adapters/postgres"
	"casino/internal/transactions/adapters/postgres/db"
	"casino/internal/transactions/app"
	"casino/internal/transactions/domain"
	ports_http "casino/internal/transactions/ports/http"
	"casino/tests/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1. Setup Containers
	containers, err := helpers.SetupContainers(ctx, "../..")
	require.NoError(t, err)
	defer containers.Postgres.Terminate(ctx)
	defer containers.Kafka.Terminate(ctx)

	// 2. Connect to DB
	pgCfg := postgres.Config{
		Host:    containers.PostgresConfig.Host,
		Port:    containers.PostgresConfig.Port,
		User:    containers.PostgresConfig.User,
		Pass:    containers.PostgresConfig.Password,
		Name:    containers.PostgresConfig.Database,
		PoolMin: 2,
		PoolMax: 10,
	}
	dbPool, err := postgres.NewPool(ctx, pgCfg)
	require.NoError(t, err)
	defer dbPool.Close()

	// 3. Setup App and HTTP Server
	l := logger.NewNoop()
	queries := db.New(dbPool)
	repo := postgres.NewTransactionRepository(queries)
	application := app.New(repo, l)
	httpCfg := ports_http.Config{
		Port:            "0", // random port
		ShutdownTimeout: time.Second,
	}
	server := ports_http.NewServer(application, httpCfg)

	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			// Using t.Log instead of t.Error as it might be racey with Shutdown in tests
			t.Logf("HTTP server stopped: %v", err)
		}
	}()
	defer server.Shutdown(ctx)

	// Refresh addr
	serverAddr, err := server.Addr()
	require.NoError(t, err)

	// 4. Test Health Check
	t.Run("HealthCheck", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(fmt.Sprintf("http://%s/v1/health", serverAddr))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 5. Create Transaction via Repo and Verify via API
	t.Run("Create then find API", func(t *testing.T) {
		id, _ := uuid.NewV7()
		userID := "user-1"
		txn, err := domain.NewTransaction(id, userID, domain.TransactionTypeBet, 100, time.Now())
		require.NoError(t, err)

		err = repo.Save(ctx, txn)
		require.NoError(t, err)

		// Call API
		client := &http.Client{Timeout: 5 * time.Second}
		apiURL := fmt.Sprintf("http://%s/v1/transactions?userId=%s", serverAddr, userID)
		resp, err := client.Get(apiURL)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
