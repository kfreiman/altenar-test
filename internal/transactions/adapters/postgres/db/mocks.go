package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
)

type MockDBTX struct {
	mock.Mock
}

func (m *MockDBTX) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	call := m.Called(ctx, sql, args)
	return call.Get(0).(pgconn.CommandTag), call.Error(1)
}

func (m *MockDBTX) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	call := m.Called(ctx, sql, args)
	if call.Get(0) == nil {
		return nil, call.Error(1)
	}
	return call.Get(0).(pgx.Rows), call.Error(1)
}

func (m *MockDBTX) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	call := m.Called(ctx, sql, args)
	return call.Get(0).(pgx.Row)
}

type MockRows struct {
	mock.Mock
}

func (m *MockRows) Close() {
	m.Called()
}

func (m *MockRows) Err() error {
	return m.Called().Error(0)
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	return m.Called().Get(0).(pgconn.CommandTag)
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return m.Called().Get(0).([]pgconn.FieldDescription)
}

func (m *MockRows) Next() bool {
	return m.Called().Bool(0)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	return m.Called(dest...).Error(0)
}

func (m *MockRows) Values() ([]interface{}, error) {
	call := m.Called()
	return call.Get(0).([]interface{}), call.Error(1)
}

func (m *MockRows) RawValues() [][]byte {
	return m.Called().Get(0).([][]byte)
}

func (m *MockRows) Conn() *pgx.Conn {
	return m.Called().Get(0).(*pgx.Conn)
}
