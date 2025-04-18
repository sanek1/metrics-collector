package controller

import (
	"testing"

	sf "github.com/sanek1/metrics-collector/internal/flags/server"
	"github.com/sanek1/metrics-collector/internal/storage/server/mocks"
	"github.com/sanek1/metrics-collector/pkg/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewController(t *testing.T) {
	mockStorage := mocks.NewStorage(t)
	mockFileStorage := mocks.NewFileStorage(t)
	logger, _ := logging.NewZapLogger(zap.InfoLevel)
	options := &sf.ServerOptions{
		FlagRunAddr:   ":8080",
		StoreInterval: 60,
		Path:          "test.json",
		Restore:       true,
		DBPath:        "",
		UseDatabase:   false,
		CryptoKey:     "test-key",
	}

	controller := NewController(mockFileStorage, mockStorage, options, logger)

	assert.NotNil(t, controller)
	assert.Equal(t, mockStorage, controller.storage)
	assert.Equal(t, mockFileStorage, controller.fieStorage)
	assert.NotNil(t, controller.router)
	assert.Equal(t, logger, controller.logger)
}
