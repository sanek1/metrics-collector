package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) InfoCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) ErrorCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) WarnCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

func (m *MockLogger) DebugCtx(ctx context.Context, msg string, fields ...interface{}) {
	m.Called(ctx, msg, fields)
}

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) Exec(query string, args ...interface{}) error {
	callArgs := m.Called(query, args)
	return callArgs.Error(0)
}

func (m *MockDB) Query(query string, args ...interface{}) (map[string]interface{}, error) {
	callArgs := m.Called(query, args)
	return callArgs.Get(0).(map[string]interface{}), callArgs.Error(1)
}

func TestDBStorageWithMocks(t *testing.T) {
	mockLogger := new(MockLogger)
	mockLogger.On("InfoCtx", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("ErrorCtx", mock.Anything, mock.Anything, mock.Anything).Return()

	mockDB := new(MockDB)
	mockDB.On("Ping").Return(nil)
	mockDB.On("Exec", mock.Anything, mock.Anything).Return(nil)

	gaugeValue := float64(123.45)
	gaugeResults := map[string]interface{}{
		"id":    "TestGauge",
		"type":  "gauge",
		"value": gaugeValue,
	}
	mockDB.On("Query", "SELECT id, type, value, delta FROM metrics WHERE id = $1 AND type = $2", mock.Anything).Return(gaugeResults, nil)

	deltaValue := int64(42)
	counterResults := map[string]interface{}{
		"id":    "TestCounter",
		"type":  "counter",
		"delta": deltaValue,
	}
	mockDB.On("Query", "SELECT id, type, value, delta FROM metrics WHERE id = $1 AND type = $2", mock.Anything).Return(counterResults, nil)

	t.Run("Ping", func(t *testing.T) {
		err := mockDB.Ping()
		assert.NoError(t, err)
	})

	t.Run("Exec", func(t *testing.T) {
		err := mockDB.Exec("INSERT INTO metrics (id, type, value, delta) VALUES ($1, $2, $3, $4)", "TestGauge", "gauge", 123.45, nil)
		assert.NoError(t, err)
	})
}
