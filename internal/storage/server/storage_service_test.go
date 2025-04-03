package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	flags "github.com/sanek1/metrics-collector/internal/flags/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

func TestGetStorage(t *testing.T) {
	opt := &flags.ServerOptions{}
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	t.Run("Database storage", func(t *testing.T) {
		storage := GetStorage(true, opt, logger)
		_, ok := storage.(*DBStorage)
		assert.True(t, ok, "Should return DBStorage instance")
	})

	t.Run("In-memory storage", func(t *testing.T) {
		storage := GetStorage(false, opt, logger)
		_, ok := storage.(*MetricsStorage)
		assert.True(t, ok, "Should return MetricsStorage instance")
	})
}
