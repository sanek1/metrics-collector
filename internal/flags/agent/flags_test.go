package flags

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlags(t *testing.T) {
	oldArgs := os.Args
	oldAddr := os.Getenv("ADDRESS")
	oldReport := os.Getenv("REPORT_INTERVAL")
	oldPoll := os.Getenv("POLL_INTERVAL")
	oldKey := os.Getenv("KEY")
	oldLimit := os.Getenv("RATE_LIMIT")

	defer func() {
		os.Args = oldArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		_ = os.Setenv("ADDRESS", oldAddr)
		_ = os.Setenv("REPORT_INTERVAL", oldReport)
		_ = os.Setenv("POLL_INTERVAL", oldPoll)
		_ = os.Setenv("KEY", oldKey)
		_ = os.Setenv("RATE_LIMIT", oldLimit)
	}()

	t.Run("default values", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test"}
		_ = os.Unsetenv("ADDRESS")
		_ = os.Unsetenv("REPORT_INTERVAL")
		_ = os.Unsetenv("POLL_INTERVAL")
		_ = os.Unsetenv("KEY")
		_ = os.Unsetenv("RATE_LIMIT")

		opt := ParseFlags()
		assert.Equal(t, ":8080", opt.FlagRunAddr)
		assert.Equal(t, "", opt.CryptoKey)
		assert.Equal(t, int64(10), opt.ReportInterval)
		assert.Equal(t, int64(2), opt.PollInterval)
		assert.Equal(t, int64(2), opt.RateLimit)
	})

	t.Run("command line flags", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test", "-a", ":9090", "-k", "testkey", "-r", "20", "-p", "5", "-l", "10"}
		_ = os.Unsetenv("ADDRESS")
		_ = os.Unsetenv("REPORT_INTERVAL")
		_ = os.Unsetenv("POLL_INTERVAL")
		_ = os.Unsetenv("KEY")
		_ = os.Unsetenv("RATE_LIMIT")

		opt := ParseFlags()
		assert.Equal(t, ":9090", opt.FlagRunAddr)
		assert.Equal(t, "testkey", opt.CryptoKey)
		assert.Equal(t, int64(20), opt.ReportInterval)
		assert.Equal(t, int64(5), opt.PollInterval)
		assert.Equal(t, int64(10), opt.RateLimit)
	})

	t.Run("environment variables", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test"}
		_ = os.Setenv("ADDRESS", ":7070")
		_ = os.Setenv("REPORT_INTERVAL", "30")
		_ = os.Setenv("POLL_INTERVAL", "7")
		_ = os.Setenv("KEY", "envkey")
		_ = os.Setenv("RATE_LIMIT", "5")

		opt := ParseFlags()
		assert.Equal(t, ":7070", opt.FlagRunAddr)
		assert.Equal(t, "envkey", opt.CryptoKey)
		assert.Equal(t, int64(30), opt.ReportInterval)
		assert.Equal(t, int64(7), opt.PollInterval)
		assert.Equal(t, int64(5), opt.RateLimit)
	})

	t.Run("env overrides flags", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test", "-a", ":9090", "-k", "testkey", "-r", "20", "-p", "5", "-l", "10"}
		_ = os.Setenv("ADDRESS", ":7070")
		_ = os.Setenv("REPORT_INTERVAL", "30")
		_ = os.Setenv("POLL_INTERVAL", "7")
		_ = os.Setenv("KEY", "envkey")
		_ = os.Setenv("RATE_LIMIT", "5")

		opt := ParseFlags()
		assert.Equal(t, ":7070", opt.FlagRunAddr)
		assert.Equal(t, "envkey", opt.CryptoKey)
		assert.Equal(t, int64(30), opt.ReportInterval)
		assert.Equal(t, int64(7), opt.PollInterval)
		assert.Equal(t, int64(5), opt.RateLimit)
	})
}

func TestParseDuration(t *testing.T) {
	testCases := []struct {
		name     string
		duration string
		expected int64
		hasError bool
	}{
		{
			name:     "Seconds",
			duration: "10s",
			expected: 10,
			hasError: false,
		},
		{
			name:     "Minutes",
			duration: "2m",
			expected: 120,
			hasError: false,
		},
		{
			name:     "Hours",
			duration: "1h",
			expected: 3600,
			hasError: false,
		},
		{
			name:     "Complex",
			duration: "1m30s",
			expected: 90,
			hasError: false,
		},
		{
			name:     "Invalid",
			duration: "invalid",
			expected: 0,
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseDuration(tc.duration)

			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Создаем временный файл конфигурации
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "agent_config.json")

	// Пишем тестовую конфигурацию в файл
	configData := `{
		"address": "localhost:9090",
		"report_interval": "5s",
		"poll_interval": "2s",
		"crypto_key": "/path/to/key.pem"
	}`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Загружаем конфигурацию
	config, err := LoadConfig(configPath)
	require.NoError(t, err)

	// Проверяем загруженные значения
	assert.Equal(t, "localhost:9090", config.Address)
	assert.Equal(t, "5s", config.ReportInterval)
	assert.Equal(t, "2s", config.PollInterval)
	assert.Equal(t, "/path/to/key.pem", config.CryptoKey)

	// Проверяем ошибку при отсутствующем файле
	_, err = LoadConfig(filepath.Join(tempDir, "nonexistent.json"))
	assert.Error(t, err)

	// Проверяем ошибку при неверном JSON
	invalidConfigPath := filepath.Join(tempDir, "invalid_config.json")
	err = os.WriteFile(invalidConfigPath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(invalidConfigPath)
	assert.Error(t, err)
}

func TestApplyFileConfig(t *testing.T) {
	// Создаем тестовую конфигурацию
	config := &AgentFileConfig{
		Address:        "localhost:9090",
		ReportInterval: "5s",
		PollInterval:   "2s",
		CryptoKey:      "/path/to/key.pem",
	}

	// Создаем опции агента
	opt := &Options{
		FlagRunAddr:    ":8080",
		ReportInterval: 10,
		PollInterval:   2,
		CryptoKey:      "",
	}

	// Применяем конфигурацию
	err := ApplyFileConfig(opt, config)
	require.NoError(t, err)

	// Проверяем результаты
	assert.Equal(t, "localhost:9090", opt.FlagRunAddr)
	assert.Equal(t, int64(5), opt.ReportInterval)
	assert.Equal(t, int64(2), opt.PollInterval)
	assert.Equal(t, "/path/to/key.pem", opt.CryptoKey)

	// Проверяем ошибку при неверном формате длительности
	invalidConfig := &AgentFileConfig{
		ReportInterval: "invalid",
	}

	err = ApplyFileConfig(opt, invalidConfig)
	assert.Error(t, err)

	invalidConfig = &AgentFileConfig{
		ReportInterval: "5s",
		PollInterval:   "invalid",
	}

	err = ApplyFileConfig(opt, invalidConfig)
	assert.Error(t, err)
}

func TestParseFlags_WithConfigFile(t *testing.T) {
	// Сохраняем оригинальные аргументы
	oldArgs := os.Args
	oldEnv := os.Getenv("CONFIG")
	defer func() {
		os.Args = oldArgs
		os.Setenv("CONFIG", oldEnv)
	}()

	// Создаем временный файл конфигурации
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "agent_config.json")

	// Пишем тестовую конфигурацию в файл
	configData := `{
		"address": "localhost:9090",
		"report_interval": "5s",
		"poll_interval": "2s",
		"crypto_key": "/path/to/key.pem"
	}`

	err := os.WriteFile(configPath, []byte(configData), 0644)
	require.NoError(t, err)

	// Тест с конфигурацией через флаг -config
	t.Run("config from command line flag", func(t *testing.T) {
		// Сбрасываем флаги для избежания ошибки повторного определения
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"app", "-config=" + configPath}

		// Вызываем ParseFlags
		opt := ParseFlags()

		// Проверяем, что значения из конфигурации были применены
		assert.Equal(t, "localhost:9090", opt.FlagRunAddr)
		assert.Equal(t, int64(5), opt.ReportInterval)
		assert.Equal(t, int64(2), opt.PollInterval)
		assert.Equal(t, "/path/to/key.pem", opt.CryptoKey)
	})

	// Тест с конфигурацией через переменную окружения
	t.Run("config from environment variable", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"app"}
		os.Setenv("CONFIG", configPath)

		opt := ParseFlags()

		// Проверяем, что значения из конфигурации были применены
		assert.Equal(t, "localhost:9090", opt.FlagRunAddr)
		assert.Equal(t, int64(5), opt.ReportInterval)
		assert.Equal(t, int64(2), opt.PollInterval)
		assert.Equal(t, "/path/to/key.pem", opt.CryptoKey)
	})

	// Тест, что переменные окружения имеют приоритет над конфигурацией
	t.Run("environment variables override config", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"app"}
		os.Setenv("CONFIG", configPath)
		os.Setenv("ADDRESS", "localhost:8888")
		os.Setenv("REPORT_INTERVAL", "30")
		os.Setenv("POLL_INTERVAL", "15")
		os.Setenv("CRYPTO_KEY", "/other/key.pem")

		opt := ParseFlags()

		// Проверяем, что значения из переменных окружения имеют приоритет
		assert.Equal(t, "localhost:8888", opt.FlagRunAddr)
		assert.Equal(t, int64(30), opt.ReportInterval)
		assert.Equal(t, int64(15), opt.PollInterval)
		assert.Equal(t, "/other/key.pem", opt.CryptoKey)
	})
}
