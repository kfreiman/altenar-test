//go:build go1.22

package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"casino/internal/transactions/app"
	"casino/internal/transactions/app/query"
	"casino/internal/transactions/domain"
	"casino/internal/transactions/ports/http/gen"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	ErrCodeInvalidTransactionType = "INVALID_TRANSACTION_TYPE"
	ErrCodeInternalError          = "INTERNAL_ERROR"
)

type Config struct {
	Port            string        `env:"CASINO_HTTP_PORT" env-default:"8080" env-description:"HTTP server port" json:"port"`
	ShutdownTimeout time.Duration `env:"CASINO_HTTP_SHUTDOWN_TIMEOUT" env-default:"3s" env-description:"Graceful shutdown timeout" json:"shutdown_timeout"`
}

type TransactionHandler struct {
	app app.Application
}

func NewTransactionHandler(app app.Application) *TransactionHandler {
	return &TransactionHandler{app: app}
}

func (h *TransactionHandler) HealthCheck(ctx context.Context, request gen.HealthCheckRequestObject) (gen.HealthCheckResponseObject, error) {
	return gen.HealthCheck200JSONResponse{
		Status: "healthy",
	}, nil
}

func (h *TransactionHandler) ListTransactions(ctx context.Context, request gen.ListTransactionsRequestObject) (gen.ListTransactionsResponseObject, error) {
	var domainType *domain.TransactionType
	if request.Params.TransactionType != nil {
		tType, err := domain.ParseTransactionType(string(*request.Params.TransactionType))
		if err != nil {
			return gen.ListTransactions400JSONResponse{
				Code:    ErrCodeInvalidTransactionType,
				Message: err.Error(),
			}, nil
		}
		domainType = &tType
	}

	// Build pagination from request parameters
	var pagination *domain.Pagination
	if request.Params.Cursor != nil || request.Params.PageSize != nil {
		var cursor *domain.Cursor
		if request.Params.Cursor != nil {
			cursor = &domain.Cursor{ID: *request.Params.Cursor}
		}
		pageSize := domain.DefaultPageSize
		if request.Params.PageSize != nil {
			pageSize = *request.Params.PageSize
		}
		pagination = domain.NewPagination(cursor, pageSize)
	}

	q := query.ListTransactions{
		UserID:          request.Params.UserId,
		TransactionType: domainType,
		Pagination:      pagination,
	}

	txResult, err := h.app.Queries.ListTransactions.Handle(ctx, q)
	if err != nil {
		return gen.ListTransactions400JSONResponse{
			Code:    ErrCodeInternalError,
			Message: err.Error(),
		}, nil
	}

	transactions := make([]gen.Transaction, len(txResult.Transactions))
	for i, t := range txResult.Transactions {
		transactions[i] = gen.Transaction{
			Id:              t.ID(),
			UserId:          t.UserID(),
			TransactionType: gen.TransactionType(t.Type().String()),
			Amount:          t.Amount(),
			Timestamp:       t.Timestamp(),
		}
	}

	var nextCursor *openapi_types.UUID
	if txResult.NextCursor != nil {
		uuid := openapi_types.UUID(txResult.NextCursor.ID)
		nextCursor = &uuid
	}

	response := gen.TransactionListResponse{
		Transactions: transactions,
		NextCursor:   nextCursor,
		HasMore:      txResult.HasMore,
	}

	return gen.ListTransactions200JSONResponse(response), nil
}

type Server struct {
	app     app.Application
	cfg     Config
	httpSrv *http.Server
}

func NewServer(app app.Application, cfg Config) *Server {
	handler := NewTransactionHandler(app)
	mux := gen.Handler(gen.NewStrictHandler(handler, nil))

	return &Server{
		app: app,
		cfg: cfg,
		httpSrv: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: mux,
		},
	}
}

func (s *Server) Run() error {
	ln, err := net.Listen("tcp", s.httpSrv.Addr)
	if err != nil {
		return err
	}
	s.httpSrv.Addr = ln.Addr().String()
	return s.httpSrv.Serve(ln)
}

func (s *Server) Addr() (string, error) {
	for i := 0; i < 10; i++ {
		if s.httpSrv.Addr != ":0" && s.httpSrv.Addr != "" && !strings.HasSuffix(s.httpSrv.Addr, ":0") {
			return s.httpSrv.Addr, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return "", fmt.Errorf("server address not available")
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}
