package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	h "github.com/sanek1/metrics-collector/internal/handlers"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	v "github.com/sanek1/metrics-collector/internal/validation"
	"github.com/sanek1/metrics-collector/pkg/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type testTable struct {
	url    string
	body   string
	want   string
	status int
}

func TestRouter(t *testing.T) {
	l, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		log.Panic(err)
	}
	s := storage.GetStorage(false, nil, l)
	memStorage := h.NewStorage(s, l)

	router := gin.New()
	router.Use(v.GzipMiddleware())
	router.POST("/update/counter/:metricName/:metricValue", memStorage.MetricHandler)
	router.POST("/update/gauge/:metricName/:metricValue", memStorage.MetricHandler)
	router.POST("/", memStorage.MetricHandler)

	var testTable = []testTable{
		{"/update/counter/Mallocs/777", `{"id": "test_Mallocs", "type": "counter", "delta": 5, "value": 123.4}`, `{"id":"test_Mallocs","type":"counter","delta":5}`, http.StatusOK},
		{"/update/counter/Mallocs/777", `{"id": "test_Mallocs", "type": "counter", "delta": 6, "value": 123.4}`, `{"id":"test_Mallocs","type":"counter","delta":11}`, http.StatusOK},
		{"/update/gauge/Alloc/777", `{"id": "test_Alloc", "type": "gauge", "value": 123.4}`, `{"id":"test_Alloc","type":"gauge","value":123.4}`, http.StatusOK},
	}

	for _, tt := range testTable {
		t.Run("sends_gzip_"+tt.url, func(t *testing.T) {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
			_, err := zw.Write([]byte(tt.body))
			require.NoError(t, err)
			err = zw.Close()
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, tt.url, &buf)
			req.Header.Set("Content-Encoding", "gzip")

			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			require.Equal(t, tt.status, w.Code)

			require.JSONEq(t, tt.want, w.Body.String())
		})
	}
}

func TestGzipCompression(t *testing.T) {
	l, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		log.Panic(err)
	}

	s := storage.GetStorage(false, nil, l)
	memStorage := h.NewStorage(s, l)

	router := gin.New()
	router.Use(v.GzipMiddleware())
	router.POST("/", memStorage.MetricHandler)

	requestBody := `{"id": "testSetGet33", "type": "gauge", "value": 123.4}`
	successBody := `{"id":"testSetGet33","type":"gauge","value":123.4}`

	t.Run("sends_gzip", func(t *testing.T) {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, err := zw.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zw.Close()
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.JSONEq(t, successBody, w.Body.String())
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(requestBody))
		req.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		require.Equal(t, "gzip", w.Header().Get("Content-Encoding"))

		zr, err := gzip.NewReader(w.Body)
		require.NoError(t, err)

		decompressedBody, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(decompressedBody))
	})
}

func TestPrintBuildInfo(t *testing.T) {
	originalBuildVersion := buildVersion
	originalBuildDate := buildDate
	originalBuildCommit := buildCommit

	defer func() {
		buildVersion = originalBuildVersion
		buildDate = originalBuildDate
		buildCommit = originalBuildCommit
	}()

	buildVersion = "test-version"
	buildDate = "test-date"
	buildCommit = "test-commit"

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	printBuildInfo()

	w.Close()
	out, _ := io.ReadAll(r)
	output := string(out)

	require.Contains(t, output, "Build version: test-version")
	require.Contains(t, output, "Build date: test-date")
	require.Contains(t, output, "Build commit: test-commit")

	buildVersion = ""
	buildDate = ""
	buildCommit = ""

	r, w, _ = os.Pipe()
	os.Stdout = w

	printBuildInfo()

	w.Close()
	out, _ = io.ReadAll(r)
	output = string(out)

	require.Contains(t, output, "Build version: N/A")
	require.Contains(t, output, "Build date: N/A")
	require.Contains(t, output, "Build commit: N/A")
}

func TestRun_Error(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	defer func() {
		os.Stderr = oldStderr
	}()

	errCh := make(chan error, 1)
	errCh <- io.EOF
	code := run()

	w.Close()
	errOut, _ := io.ReadAll(r)
	errorOutput := string(errOut)

	require.Equal(t, 1, code)
	require.Contains(t, errorOutput, "critical error:")
}
