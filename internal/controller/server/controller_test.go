package controller

import (
	"context"
	"testing"
	"time"

	m "github.com/sanek1/metrics-collector/internal/models"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	"github.com/sanek1/metrics-collector/pkg/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestController_PeriodicallySaveBackUp(t *testing.T) {
	type args struct {
		filename string
		restore  bool
		interval time.Duration
	}
	tests := []struct {
		name      string
		c         *Controller
		args      args
		expectErr bool
	}{
		{
			name: "Successful backup",
			c:    &Controller{},
			args: args{
				filename: "File_Log_Store.json",
				restore:  false,
				interval: 5 * time.Second,
			},
			expectErr: false,
		},
	}

	logger, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		panic(err)
	}

	ctrl := New(storage.NewMetricsStorage(logger), storage.NewMetricsStorage(logger), nil, logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a new context
			ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
			defer cancel()
			// send metrics to server
			var v = 11.0
			metric := m.NewMetricGauge("test_gauge", &v)
			_, err := ctrl.storage.SetGauge(ctx, *metric)
			if err != nil {
				t.Error("Failed to set gauge in memory storage")
			}
			// create a new file
			go ctrl.PeriodicallySaveBackUp(ctx, tt.args.filename, tt.args.restore, tt.args.interval)
			<-ctx.Done()

			//file save successful
			time.Sleep(15 * time.Second)
			// check file exists
			if err := ctrl.fieStorage.LoadFromFile(tt.args.filename); err != nil {
				t.Errorf("Error loading metrics from file: %v", err)
			}

			assert.Equal(t, err, nil)
			// improvement,
			// check for other types of metrics and what is in the file
		})
	}
}
