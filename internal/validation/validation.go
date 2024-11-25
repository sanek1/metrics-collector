package validation

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	typeMethod = 1
	typeMetric = 2
	metricName = 3
	metricVal  = 4
	minPathLen = 5
)

func Validation(next http.Handler) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		message := "-- start validation --"
		printValidationMessage(message)

		// check type of request
		if r.Method != http.MethodPost {
			message = "The specified address accepts only POST"
			printValidationMessage(message)
			http.Error(rw, message, http.StatusMethodNotAllowed)
			return
		}

		// check path
		splitedPath := strings.Split(r.URL.Path, "/")
		if len(splitedPath) < minPathLen {
			message = "invalid path"
			printValidationMessage(message)
			http.Error(rw, message, http.StatusNotFound)
			return
		}

		// check correct name metrick
		if splitedPath[typeMethod] != "update" {
			message = "invalid type method name"
			printValidationMessage(message)
			http.Error(rw, message, http.StatusBadRequest)
			return
		}

		// check correct name metrick
		if splitedPath[typeMetric] != "gauge" && splitedPath[typeMetric] != "counter" {
			message = "invalid metrick name"
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
