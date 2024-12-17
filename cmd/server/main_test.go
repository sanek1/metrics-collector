package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	h "github.com/sanek1/metrics-collector/internal/handlers"
	rc "github.com/sanek1/metrics-collector/internal/routing"
	s "github.com/sanek1/metrics-collector/internal/storage"
	v "github.com/sanek1/metrics-collector/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTable struct {
	url    string
	body   string
	want   string
	status int
}

func testRequest(t *testing.T, ts *httptest.Server, method string, str testTable) (*http.Response, string) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, method, ts.URL+str.url, bytes.NewBufferString(str.body))
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	if _, err := v.Initialize("test_level"); err != nil {
		return
	}

	memStorage := s.NewMemoryStorage()
	metricStorage := h.MetricStorage{
		Storage: memStorage,
	}

	ts := httptest.NewServer(rc.InitRouting(metricStorage))
	defer ts.Close()
	var testTable = []testTable{
		{"/update/counter/Mallocs/777", `{"id": "test_Mallocs", "type": "counter", "delta": 5, "value": 123.4}`, `{"id":"test_Mallocs","type":"counter","delta":5}`, http.StatusOK},
		{"/update/counter/Mallocs/777", `{"id": "test_Mallocs", "type": "counter", "delta": 6, "value": 123.4}`, `{"id":"test_Mallocs","type":"counter","delta":11}`, http.StatusOK}, // expected 5+5=10

		{"/update/gauge/Alloc/777", `{"id": "test_Alloc", "type": "gauge", "delta": 6, "value": 123.4}`, `{"id":"test_Alloc","type":"gauge","delta":6,"value":123.4}`, http.StatusOK},
		//bad response
		{"/update/unknown/Mallocs/7", "", "unsupported request type\nunsupported request type", http.StatusBadRequest},
		{"/update1/counter/Mallocs/7", "", "Not Implemented\n", http.StatusBadRequest},
		//{"/update/gauge/Alloc/77.7.", "", "No request body\n", http.StatusBadRequest},
	}
	for _, v := range testTable {
		resp, post := testRequest(t, ts, "POST", v)
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
