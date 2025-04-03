package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewZapLogger(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		logger, err := NewZapLogger(zap.DebugLevel)
		require.NoError(t, err)
		require.NotNil(t, logger)
		defer logger.Sync()

		assert.Equal(t, zap.DebugLevel, logger.level.Level())
	})
}

func TestWithContextFields(t *testing.T) {
	logger, err := NewZapLogger(zap.DebugLevel)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = logger.WithContextFields(ctx, zap.String("request_id", "123"))

	fields, ok := ctx.Value(zapFieldsKey).(ZapFields)
	require.True(t, ok)
	assert.Contains(t, fields, "request_id")
	assert.Equal(t, "123", fields["request_id"].String)
}

func TestLogLevels(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := &ZapLogger{
		logger: zap.New(core),
		level:  zap.NewAtomicLevelAt(zap.InfoLevel),
	}

	tests := []struct {
		name     string
		logFunc  func()
		expected zapcore.Level
	}{
		{
			name: "info level",
			logFunc: func() {
				logger.InfoCtx(context.Background(), "info message")
			},
			expected: zap.InfoLevel,
		},
		{
			name: "warn level",
			logFunc: func() {
				logger.WarnCtx(context.Background(), "warn message")
			},
			expected: zap.WarnLevel,
		},
		{
			name: "error level",
			logFunc: func() {
				logger.ErrorCtx(context.Background(), "error message")
			},
			expected: zap.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFunc()
			entries := recorded.All()
			require.Len(t, entries, 1)
			assert.Equal(t, tt.expected, entries[0].Level)
			recorded.TakeAll()
		})
	}
}

func TestSyncAndStd(t *testing.T) {
	logger, err := NewZapLogger(zap.InfoLevel)
	require.NoError(t, err)

	stdLogger := logger.Std()
	require.NotNil(t, stdLogger)

	logger.Sync()
}

func TestPanicCtx(t *testing.T) {
	core, recorded := observer.New(zap.DebugLevel)
	logger := &ZapLogger{
		logger: zap.New(core),
		level:  zap.NewAtomicLevelAt(zap.DebugLevel),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic")
		}

		entries := recorded.All()
		require.Len(t, entries, 1, "should have one log entry")
		assert.Equal(t, "panic test", entries[0].Message)
		assert.Equal(t, "***@domain.com", entries[0].ContextMap()["email"])
	}()

	ctx := context.Background()
	ctx = logger.WithContextFields(ctx, zap.String("email", "user@domain.com"))
	logger.PanicCtx(ctx, "panic test", zap.Int("count", 42))
}

func TestSetLevel(t *testing.T) {
	atomicLevel := zap.NewAtomicLevelAt(zap.InfoLevel)
	core, recorded := observer.New(atomicLevel)
	logger := &ZapLogger{
		logger: zap.New(core),
		level:  atomicLevel,
	}

	logger.DebugCtx(context.Background(), "debug message")
	require.Empty(t, recorded.TakeAll(), "debug messages should be filtered")

	logger.SetLevel(zap.DebugLevel)
	logger.DebugCtx(context.Background(), "debug message")
	entries := recorded.TakeAll()
	require.Len(t, entries, 1, "should log debug messages after level change")
	assert.Equal(t, "debug message", entries[0].Message)

	logger.SetLevel(zap.ErrorLevel)
	logger.WarnCtx(context.Background(), "warn message")
	logger.ErrorCtx(context.Background(), "error message")
	entries = recorded.TakeAll()
	require.Len(t, entries, 1, "should only log error messages")
	assert.Equal(t, zap.ErrorLevel, entries[0].Level)
	assert.Equal(t, "error message", entries[0].Message)
}
