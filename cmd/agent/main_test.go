package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_reportClient(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value gauge
		url   string
		want  string
	}{
		{
			name:  "Alloc",
			key:   "Alloc",
			value: gauge(123),
			url:   "/update1/gauge/Alloc/123",
		},
		{
			name:  "BuckHashSys",
			key:   "BuckHashSys",
			value: gauge(-123),
			url:   "/update/gauge/BuckHashSys/-123",
		},
		{
			name:  "Frees",
			key:   "Frees",
			value: gauge(0),
			url:   "/update/gauge/Frees/0",
		},
		{
			name:  "GCCPUFraction",
			key:   "GCCPUFraction",
			value: gauge(1.1),
			url:   "/update/gauge/GCCPUFraction/0",
		},
		{
			name:  "wrong path",
			key:   "GCCPUFraction",
			value: gauge(1.1),
			url:   "/update111/gauge/GCCPUFraction/0",
		},
	}

	testServer := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(rw, "")
		}),
	)
	defer testServer.Close()

	logger := log.New(io.Discard, "", log.LstdFlags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metricURL := fmt.Sprint(testServer.URL, tt.url)
			err := reportClient(&http.Client{}, metricURL, logger)
			assert.Equal(t, err, nil)
		})
	}
}
