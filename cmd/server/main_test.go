package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	h "github.com/sanek1/metrics-collector/internal/handlers"
	s "github.com/sanek1/metrics-collector/internal/storage"
	v "github.com/sanek1/metrics-collector/internal/validation"
	"github.com/stretchr/testify/require"
)

type testTable struct {
	url    string
	body   string
	want   string
	status int
}

func TestRouter(t *testing.T) {
	if _, err := v.Initialize("test_level"); err != nil {
		return
	}
	memStorage := s.NewMemoryStorage()
	metricStorage := h.MetricStorage{
		Storage: memStorage,
		Logger:  memStorage.Logger,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v.GzipMiddleware(http.HandlerFunc(metricStorage.GetMetricsHandler)).ServeHTTP(w, r)
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	var testTable = []testTable{
		{"/update/counter/Mallocs/777", `{"id": "test_Mallocs", "type": "counter", "delta": 5, "value": 123.4}`, `{"id":"test_Mallocs","type":"counter","delta":5}`, http.StatusOK},
		{"/update/counter/Mallocs/777", `{"id": "test_Mallocs", "type": "counter", "delta": 6, "value": 123.4}`, `{"id":"test_Mallocs","type":"counter","delta":11}`, http.StatusOK}, // expected 5+5=10

		{"/update/gauge/Alloc/777", `{"id": "test_Alloc", "type": "gauge", "delta": 6, "value": 123.4}`, `{"id":"test_Alloc","type":"gauge","delta":6,"value":123.4}`, http.StatusOK},
		//bad response
		//{"/update/unknown/Mallocs/7", "", `{"error": "failed to read body"}`, http.StatusBadRequest},
		//{"/update1/counter/Mallocs/7", "", "Not Implemented\n", http.StatusBadRequest},
		//{"/update/gauge/Alloc/77.7.", "", "No request body\n", http.StatusBadRequest},
	}
	for _, v := range testTable {

		t.Run("sends_gzip", func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			zb := gzip.NewWriter(buf)
			_, err := zb.Write([]byte(v.body))
			require.NoError(t, err)
			err = zb.Close()
			require.NoError(t, err)

			r := httptest.NewRequest("POST", srv.URL, buf)
			r.RequestURI = ""
			r.Header.Set("Content-Encoding", "gzip")
			r.Header.Set("Accept-Encoding", "")

			resp, err := http.DefaultClient.Do(r)
			require.NoError(t, err)
			require.Equal(t, v.status, resp.StatusCode)

			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			str := string(b)
			require.JSONEq(t, v.want, str)
		})
	}
}

func TestGzipCompression(t *testing.T) {
	storageImpl := s.NewMemoryStorage()
	metricStorage := h.MetricStorage{
		Storage: storageImpl,
		Logger:  storageImpl.Logger,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v.GzipMiddleware(http.HandlerFunc(metricStorage.GetMetricsHandler)).ServeHTTP(w, r)
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{"id": "testSetGet33", "type": "gauge", "delta": 1, "value": 123.4}`
	successBody := `{"id":"testSetGet33","type":"gauge","delta":1,"value":123.4}`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		str := string(b)
		require.JSONEq(t, successBody, str)
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}
