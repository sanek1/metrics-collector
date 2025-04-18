package app

import (
	"context"
	"testing"
	"time"

	sf "github.com/sanek1/metrics-collector/internal/flags/server"
	"github.com/stretchr/testify/assert"
)

type mockServer struct {
	ListenAndServeFunc func() error
	Addr               string
}

func (m *mockServer) ListenAndServe() error {
	return m.ListenAndServeFunc()
}

func TestNew(t *testing.T) {
	options := &sf.ServerOptions{
		FlagRunAddr:    ":8080",
		Path:           "./tmp/metrics-db.json",
		StoreInterval:  300,
		Restore:        true,
	}

	app := New(options, false)

	assert.NotNil(t, app)
	assert.Equal(t, options, app.options)
	assert.False(t, app.useDatabase)
}

func TestApp_Run_MemoryStorage(t *testing.T) {
	options := &sf.ServerOptions{
		FlagRunAddr:    ":18080",
		Path:           "./tmp/test-metrics.json",
		StoreInterval:  1,
		Restore:        false,
	}
	app := New(options, false)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()
	select {
	case err := <-done:
		t.Fatalf("Run() completed with error: %v", err)
	case <-ctx.Done():
	}
}
