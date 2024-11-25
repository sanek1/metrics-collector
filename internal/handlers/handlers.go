package handlers

import (
	"fmt"
	"net/http"

	"github.com/sanek1/metrics-collector/internal/storage"
)

type MetricStorage struct {
	Storage storage.IMetricStorage
}

func MainPage(res http.ResponseWriter, req *http.Request) {
	data := []byte(" ---------   Main Page ---------")
	res.Write(data)
}
func (ms MetricStorage) GetHandler(res http.ResponseWriter, r *http.Request) {

	//ms.Storage.GetGauge("test")

	if r.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body := fmt.Sprintf("Method: %s\r\n", r.Method)
	body += "header =========\r\n"
	for k, v := range r.Header {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}

	body += "query params ===========================\r\n"
	for k, v := range r.URL.Query() {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}

	// return answer
	body += "close ===========================\r\n"
	res.Write([]byte(body))
}

func MiddleWare(next http.Handler) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("Before")
		next.ServeHTTP(res, req)
		fmt.Println("After")
	})
}
