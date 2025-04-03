package validation

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/sanek1/metrics-collector/internal/handlers"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type Middleware func(http.Handler) http.Handler

type MiddlewareController struct {
	l *l.ZapLogger
	s *handlers.Storage
}

func NewValidation(s *handlers.Storage, logger *l.ZapLogger) *MiddlewareController {
	return &MiddlewareController{
		l: logger,
		s: s,
	}
}

func (mc *MiddlewareController) Recover(next http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				mc.l.PanicCtx(c.Request.Context(), "recovered from panic", zap.Any("panic", rec))
				http.Error(c.Writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(c.Writer, c.Request)
	}
}
