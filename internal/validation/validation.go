package validation

import (
	"fmt"
	"net/http"
	"strings"

	c "github.com/sanek1/metrics-collector/internal/config"
)

func Validation(next http.Handler) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		message := "-- start validation --"
		printValidationMessage(message)
		rw.Header().Set("Content-Type", "application/json")

		// check type of request
		if r.Method != http.MethodPost {
			message = "The specified address accepts only POST"
			printValidationMessage(message)
			http.Error(rw, message, http.StatusMethodNotAllowed)
			return
		}

		// check path
		splitedPath := strings.Split(r.URL.Path, "/")
		if len(splitedPath) < c.MinPathLen {
			message = "invalid path"

			printValidationMessage(message)
			http.Error(rw, message, http.StatusBadRequest)
			return
		}

		next.ServeHTTP(rw, r)
		printValidationMessage("--validation completed successfully--")
	})
}

func printValidationMessage(message string) {
	fmt.Println(message)
}
