package flags

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseServerFlags(t *testing.T) {
	oldArgs := os.Args
	oldAddr := os.Getenv("ADDRESS")
	oldStoreInterval := os.Getenv("STORE_INTERVAL")
	oldFilePath := os.Getenv("FILE_STORAGE_PATH")
	oldRestore := os.Getenv("RESTORE")
	oldDBPath := os.Getenv("DATABASE_DSN")
	oldKey := os.Getenv("KEY")

	defer func() {
		os.Args = oldArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		_ = os.Setenv("ADDRESS", oldAddr)
		_ = os.Setenv("STORE_INTERVAL", oldStoreInterval)
		_ = os.Setenv("FILE_STORAGE_PATH", oldFilePath)
		_ = os.Setenv("RESTORE", oldRestore)
		_ = os.Setenv("DATABASE_DSN", oldDBPath)
		_ = os.Setenv("KEY", oldKey)
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
}
