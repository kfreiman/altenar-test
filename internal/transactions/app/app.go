package app

import (
	"casino/internal/logger"
	"casino/internal/transactions/app/command"
	"casino/internal/transactions/app/query"
	"casino/internal/transactions/domain"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

func New(repo domain.Repository, logger logger.Logger) Application {
	return Application{
		Commands: Commands{
			ProcessTransaction: command.NewProcessTransactionHandler(repo, logger),
		},
		Queries: Queries{
			ListTransactions: query.NewListTransactionsHandler(repo, logger),
		},
	}
}

type Commands struct {
	ProcessTransaction command.ProcessTransactionHandler
}

type Queries struct {
	ListTransactions query.ListTransactionsHandler
}
