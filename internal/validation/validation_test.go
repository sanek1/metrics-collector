package validation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/sanek1/metrics-collector/internal/handlers"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type MockZapLogger struct {
	mock.Mock
}

func (m *MockZapLogger) PanicCtx(ctx context.Context, msg string, fields ...zap.Field) {
	m.Called(ctx, msg, fields)
}

func (m *MockZapLogger) WithContextFields(ctx context.Context, fields ...zap.Field) *l.ZapLogger {
	args := m.Called(ctx, fields)
	return args.Get(0).(*l.ZapLogger)
}

func (m *MockZapLogger) Sync() error {
	return m.Called().Error(0)
}

func TestRecoverWithoutPanic_Recover(t *testing.T) {
	mockLogger := new(MockZapLogger)
	mockStorage := new(handlers.Storage)
	l, _ := l.NewZapLogger(zap.InfoLevel)
	router := gin.New()
	mc := NewValidation(mockStorage, l)

	router.Use(mc.Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	router.GET("/ok", func(c *gin.Context) {})

	t.Run("normal request", func(t *testing.T) {
		mockLogger.AssertNotCalled(t, "PanicCtx")

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/ok", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
