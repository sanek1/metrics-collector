package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
	m "github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
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

	testServer := httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(rw, "")
			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("content-encoding", "")
			rw.Header().Set("Accept-Encoding", "")
		}),
	)
	defer testServer.Close()
	ctx := context.Background()
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	logger.InfoCtx(ctx, "agent started ", zap.String("time: ", time.DateTime))
	opt := flags.ParseFlags()
	s := NewServices(opt, logger)

	for _, tt := range tests {
		t.Run(tt.ID, func(t *testing.T) {
			metricURL := fmt.Sprint(testServer.URL, "/update/gauge/"+tt.ID+"/"+fmt.Sprintf("%f", *tt.Value))
			err := s.SendToServerMetric(ctx, &http.Client{}, metricURL, tt)
			assert.Equal(t, err, nil)
		})
	}
}
