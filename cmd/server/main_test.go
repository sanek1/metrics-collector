// model_test.go
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	h "github.com/sanek1/metrics-collector/internal/handlers"
	s "github.com/sanek1/metrics-collector/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {

	memStorage := s.NewMemStorage()
	metricStorage := h.MetricStorage{
		Storage: memStorage,
	}

	ts := httptest.NewServer(InitRouting(metricStorage))
	defer ts.Close()
	var testTable = []struct {
		url    string
		want   string
		status int
	}{
		{"/update/counter/Mallocs/777", `{"value":"Metric Name: Mallocs, Metric Value:777"}`, http.StatusOK},
		{"/update/gauge/Alloc/777", `{"value":"Metric Name: Alloc, Metric Value:777"}`, http.StatusOK},
		//bad response
		{"/update/unknown/Mallocs/7", "404 page not found\n", http.StatusNotFound},
		{"/update1/counter/Mallocs/7", "404 page not found\n", http.StatusNotFound},
		{"/update/gauge/Alloc/77.7.", "The value does not match the expected type.\n", http.StatusBadRequest},
	}
	for _, v := range testTable {
		resp, post := testRequest(t, ts, "POST", v.url)
		defer resp.Body.Close()
		assert.Equal(t, v.status, resp.StatusCode)
		if isValidJSON(v.want) {
			assert.JSONEq(t, v.want, post)
		} else {
			assert.Equal(t, v.want, post)
		}

	}
}

func isValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
