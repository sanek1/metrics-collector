package flags

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseServerFlags(t *testing.T) {
	oldArgs := os.Args
	oldAddr := os.Getenv("ADDRESS")
	oldStoreInterval := os.Getenv("STORE_INTERVAL")
	oldFilePath := os.Getenv("FILE_STORAGE_PATH")
	oldRestore := os.Getenv("RESTORE")
	oldDBPath := os.Getenv("DATABASE_DSN")
	oldKey := os.Getenv("KEY")
	oldCryptoKey := os.Getenv("CRYPTO_KEY")
	oldConfig := os.Getenv("CONFIG")

	defer func() {
		os.Args = oldArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		_ = os.Setenv("ADDRESS", oldAddr)
		_ = os.Setenv("STORE_INTERVAL", oldStoreInterval)
		_ = os.Setenv("FILE_STORAGE_PATH", oldFilePath)
		_ = os.Setenv("RESTORE", oldRestore)
		_ = os.Setenv("DATABASE_DSN", oldDBPath)
		_ = os.Setenv("KEY", oldKey)
		_ = os.Setenv("CRYPTO_KEY", oldCryptoKey)
		_ = os.Setenv("CONFIG", oldConfig)
	}()

	t.Run("default values", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test"}
		_ = os.Unsetenv("ADDRESS")
		_ = os.Unsetenv("STORE_INTERVAL")
		_ = os.Unsetenv("FILE_STORAGE_PATH")
		_ = os.Unsetenv("RESTORE")
		_ = os.Unsetenv("DATABASE_DSN")
		_ = os.Unsetenv("KEY")
		_ = os.Unsetenv("CRYPTO_KEY")
		_ = os.Unsetenv("CONFIG")

		opt := ParseServerFlags()
		assert.Equal(t, ":8080", opt.FlagRunAddr)
		assert.Equal(t, int64(60), opt.StoreInterval)
		assert.Equal(t, "File_Log_Store.json", opt.Path)
		assert.Equal(t, true, opt.Restore)
		assert.Contains(t, opt.DBPath, "host=localhost port=5432")
		assert.Contains(t, opt.DBPath, "user=postgres password=admin")
		assert.Contains(t, opt.DBPath, "dbname=MetricStore sslmode=disable")
		assert.Equal(t, "", opt.CryptoKey)
	})

	t.Run("command line flags", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test", "-a", ":9090", "-i", "120", "-f", "custom.json", "-r=false", "-d", "postgres://test", "-k", "testkey"}
		_ = os.Unsetenv("ADDRESS")
		_ = os.Unsetenv("STORE_INTERVAL")
		_ = os.Unsetenv("FILE_STORAGE_PATH")
		_ = os.Unsetenv("RESTORE")
		_ = os.Unsetenv("DATABASE_DSN")
		_ = os.Unsetenv("KEY")
		_ = os.Unsetenv("CRYPTO_KEY")
		_ = os.Unsetenv("CONFIG")

		opt := ParseServerFlags()
		assert.Equal(t, ":9090", opt.FlagRunAddr)
		assert.Equal(t, int64(120), opt.StoreInterval)
		assert.Equal(t, "custom.json", opt.Path)
		assert.Equal(t, false, opt.Restore)
		assert.Equal(t, "postgres://test", opt.DBPath)
		assert.Equal(t, true, opt.UseDatabase)
		assert.Equal(t, "testkey", opt.CryptoKey)
	})

	t.Run("environment variables", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test"}
		_ = os.Setenv("ADDRESS", ":7070")
		_ = os.Setenv("STORE_INTERVAL", "180")
		_ = os.Setenv("FILE_STORAGE_PATH", "env.json")
		_ = os.Setenv("RESTORE", "false")
		_ = os.Setenv("DATABASE_DSN", "postgres://env")
		_ = os.Setenv("KEY", "envkey")

		opt := ParseServerFlags()
		assert.Equal(t, ":7070", opt.FlagRunAddr)
		assert.Equal(t, int64(180), opt.StoreInterval)
		assert.Equal(t, "env.json", opt.Path)
		assert.Equal(t, false, opt.Restore)
		assert.Equal(t, "postgres://env", opt.DBPath)
		assert.Equal(t, true, opt.UseDatabase)
		assert.Equal(t, "envkey", opt.CryptoKey)
	})

	t.Run("env overrides flags", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test", "-a", ":9090", "-i", "120", "-f", "custom.json", "-r=false", "-d", "postgres://test", "-k", "testkey"}
		_ = os.Setenv("ADDRESS", ":7070")
		_ = os.Setenv("STORE_INTERVAL", "180")
		_ = os.Setenv("FILE_STORAGE_PATH", "env.json")
		_ = os.Setenv("RESTORE", "true")
		_ = os.Setenv("DATABASE_DSN", "postgres://env")
		_ = os.Setenv("KEY", "envkey")

		opt := ParseServerFlags()
		assert.Equal(t, ":7070", opt.FlagRunAddr)
		assert.Equal(t, int64(180), opt.StoreInterval)
		assert.Equal(t, "env.json", opt.Path)
		assert.Equal(t, true, opt.Restore)
		assert.Equal(t, "postgres://env", opt.DBPath)
		assert.Equal(t, true, opt.UseDatabase)
		assert.Equal(t, "envkey", opt.CryptoKey)
	})

	t.Run("database connection test", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test"}
		_ = os.Unsetenv("ADDRESS")
		_ = os.Unsetenv("STORE_INTERVAL")
		_ = os.Unsetenv("FILE_STORAGE_PATH")
		_ = os.Unsetenv("RESTORE")
		_ = os.Unsetenv("DATABASE_DSN")
		_ = os.Unsetenv("KEY")

		assert.Equal(t, "host=localhost port=5432 user=postgres password=admin dbname=MetricStore sslmode=disable", initDefaulthPathDB())
	})

	t.Run("config file test", func(t *testing.T) {
		// Создаем тестовый файл конфигурации
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")
		configContent := `{
			"address": ":5000",
			"restore": false,
			"store_interval": "30s",
			"store_file": "config_store.json",
			"database_dsn": "postgres://config_user:password@localhost/configdb",
			"crypto_key": "config_key"
		}`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test", "-c", configPath}
		_ = os.Unsetenv("ADDRESS")
		_ = os.Unsetenv("STORE_INTERVAL")
		_ = os.Unsetenv("FILE_STORAGE_PATH")
		_ = os.Unsetenv("RESTORE")
		_ = os.Unsetenv("DATABASE_DSN")
		_ = os.Unsetenv("KEY")
		_ = os.Unsetenv("CRYPTO_KEY")
		_ = os.Unsetenv("CONFIG")

		opt := ParseServerFlags()
		assert.Equal(t, ":5000", opt.FlagRunAddr)
		assert.Equal(t, int64(30), opt.StoreInterval)
		assert.Equal(t, "config_store.json", opt.Path)
		assert.Equal(t, false, opt.Restore)
		assert.Equal(t, "postgres://config_user:password@localhost/configdb", opt.DBPath)
		assert.Equal(t, true, opt.UseDatabase)
		assert.Equal(t, "config_key", opt.CryptoKey)
	})

	t.Run("config from env variable", func(t *testing.T) {
		// Создаем тестовый файл конфигурации
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "env_config.json")
		configContent := `{
			"address": ":6000",
			"restore": true,
			"store_interval": "45s",
			"store_file": "env_config_store.json",
			"database_dsn": "postgres://env_config_user:password@localhost/envdb",
			"crypto_key": "env_config_key"
		}`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test"}
		_ = os.Unsetenv("ADDRESS")
		_ = os.Unsetenv("STORE_INTERVAL")
		_ = os.Unsetenv("FILE_STORAGE_PATH")
		_ = os.Unsetenv("RESTORE")
		_ = os.Unsetenv("DATABASE_DSN")
		_ = os.Unsetenv("KEY")
		_ = os.Unsetenv("CRYPTO_KEY")
		_ = os.Setenv("CONFIG", configPath)

		opt := ParseServerFlags()
		assert.Equal(t, ":6000", opt.FlagRunAddr)
		assert.Equal(t, int64(45), opt.StoreInterval)
		assert.Equal(t, "env_config_store.json", opt.Path)
		assert.Equal(t, true, opt.Restore)
		assert.Equal(t, "postgres://env_config_user:password@localhost/envdb", opt.DBPath)
		assert.Equal(t, true, opt.UseDatabase)
		assert.Equal(t, "env_config_key", opt.CryptoKey)
	})

	t.Run("env variables override config file", func(t *testing.T) {
		// Создаем тестовый файл конфигурации
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "override_config.json")
		configContent := `{
			"address": ":6000",
			"restore": true,
			"store_interval": "45s",
			"store_file": "config_store.json",
			"database_dsn": "postgres://config_user:password@localhost/configdb",
			"crypto_key": "config_key"
		}`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test", "-c", configPath}
		_ = os.Setenv("ADDRESS", ":7777")
		_ = os.Setenv("STORE_INTERVAL", "90")
		_ = os.Setenv("FILE_STORAGE_PATH", "env_override.json")
		_ = os.Setenv("RESTORE", "false")
		_ = os.Setenv("DATABASE_DSN", "postgres://env_override_user:password@localhost/overridedb")
		_ = os.Setenv("KEY", "env_override_key")

		opt := ParseServerFlags()
		assert.Equal(t, ":7777", opt.FlagRunAddr)
		assert.Equal(t, int64(90), opt.StoreInterval)
		assert.Equal(t, "env_override.json", opt.Path)
		assert.Equal(t, false, opt.Restore)
		assert.Equal(t, "postgres://env_override_user:password@localhost/overridedb", opt.DBPath)
		assert.Equal(t, true, opt.UseDatabase)
		assert.Equal(t, "env_override_key", opt.CryptoKey)
	})
}
