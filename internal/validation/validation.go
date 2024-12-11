package validation

import (
	"log"
	"net/http"
	"strings"

	c "github.com/sanek1/metrics-collector/internal/config"
)

var logger *log.Logger

func init() {
	logger = log.New(log.Writer(), "VALIDATION: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Validation(next http.Handler) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logValidationMessage("-- start validation --")
		rw.Header().Set("Content-Type", "application/json")
		splitedPath := strings.Split(r.URL.Path, "/")
		if len(splitedPath) < c.MinPathLen {
			message := "invalid path"

			logValidationMessage(message)
			http.Error(rw, message, http.StatusNotFound)
			return
		}
		next.ServeHTTP(rw, r)
		logValidationMessage("--validation completed successfully--")
	})
}

func logValidationMessage(message string) {
	logger.Println(message)
}
