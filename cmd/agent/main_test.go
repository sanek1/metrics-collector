package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	m "github.com/sanek1/metrics-collector/internal/validation"
	"github.com/stretchr/testify/assert"
)

func Test_reportClient(t *testing.T) {
	value1 := float64(123)
	value2 := float64(-123)
	value3 := float64(0)
	value4 := float64(1.1)

	tests := []m.Metrics{
		{
			ID:    "Alloc",
			MType: "gauge",
			Value: &value1,
		},
		{
			ID:    "BuckHashSys",
			MType: "gauge",
			Delta: nil,
			Value: &value2,
		},
		{
			ID:    "Frees",
			MType: "gauge",
			Delta: nil,
			Value: &value3,
		},
		{
			ID:    "GCCPUFraction",
			MType: "gauge",
			Delta: nil,
			Value: &value4,
		},
		{
			ID:    "wrong path",
			MType: "gauge",
			Delta: nil,
			Value: &value4,
		},
	}

	logger := log.New(io.Discard, "", log.LstdFlags)

	for _, tt := range tests {
		t.Run(tt.ID, func(t *testing.T) {
			metricURL := fmt.Sprint("http://localhost:8080", "/update/gauge/"+tt.ID+"/"+fmt.Sprintf("%f", *tt.Value))
			err := reportClient(&http.Client{}, metricURL, tt, logger)
			assert.Equal(t, err, nil)
		})
	}
}
