package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	af "github.com/sanek1/metrics-collector/internal/flags/agent"
)

type MockController struct {
	mock.Mock
}

func (m *MockController) SendingCounterMetrics(ctx context.Context, pollCount *int64, client *http.Client) {
	m.Called(ctx, pollCount, client)
}

func (m *MockController) SendingGaugeMetrics(ctx context.Context, metrics map[string]float64, client *http.Client) {
	m.Called(ctx, metrics, client)
}

func Test_InitDataAgent(t *testing.T) {
	opt := &af.Options{
		PollInterval:   2,
		ReportInterval: 10,
	}
	pollTick, reportTick, metrics, gpMetrics := initDataAgent(opt)

	assert.NotNil(t, pollTick)
	assert.NotNil(t, reportTick)
	assert.NotNil(t, metrics)
	assert.NotNil(t, gpMetrics)
	assert.Equal(t, 0, len(metrics))
	assert.Equal(t, 0, len(gpMetrics))

	time.Sleep(2100 * time.Millisecond)
	select {
	case <-pollTick.C:
	default:
		t.Error("expected pollTick")
	}
}

func TestNewApp(t *testing.T) {
	opt := &af.Options{
		PollInterval:   2,
		ReportInterval: 10,
	}

	app := New(opt)

	assert.NotNil(t, app)
	assert.NotNil(t, app.controller)
	assert.Equal(t, opt, app.opt)
}
