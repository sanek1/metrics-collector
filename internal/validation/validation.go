package validation

import (
	"net/http"
	"strings"

	"github.com/sanek1/metrics-collector/internal/config"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type Middleware func(http.Handler) http.Handler

type MiddlewareController struct {
	l *l.ZapLogger
}

func New(logger *l.ZapLogger) *MiddlewareController {
	return &MiddlewareController{
		l: logger,
	}
}

func Conveyor(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}
func (c *MiddlewareController) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				c.l.PanicCtx(r.Context(), "recovered from panic", zap.Any("panic", rec))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (c *MiddlewareController) ValidationOld(next http.Handler) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		splitedPath := strings.Split(r.URL.Path, "/")
		if len(splitedPath) < config.MinPathLen {
			message := "invalid path"
			c.l.ErrorCtx(r.Context(), message, zap.Any("invalid path", r.URL.Path))
			http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		next.ServeHTTP(rw, r)
	})
}
