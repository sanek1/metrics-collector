package validation

import (
	"fmt"
	"net/http"
	"strings"

	c "github.com/sanek1/metrics-collector/internal/config"
)

type Middleware func(http.Handler) http.Handler

func Conveyor(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func ValidationOld(next http.Handler) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		splitedPath := strings.Split(r.URL.Path, "/")
		if len(splitedPath) < c.MinPathLen {
			message := "invalid path"

			logValidationMessage(message)
			http.Error(rw, message, http.StatusNotFound)
			return
		}
		next.ServeHTTP(rw, r)
	})
}

func logValidationMessage(message string) {
	fmt.Println(message)
}
