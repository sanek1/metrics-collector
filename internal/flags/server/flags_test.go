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
		os.Setenv("ADDRESS", oldAddr)
		os.Setenv("STORE_INTERVAL", oldStoreInterval)
		os.Setenv("FILE_STORAGE_PATH", oldFilePath)
		os.Setenv("RESTORE", oldRestore)
		os.Setenv("DATABASE_DSN", oldDBPath)
		os.Setenv("KEY", oldKey)
	}()

	t.Run("default values", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		os.Args = []string{"test"}
		os.Unsetenv("ADDRESS")
		os.Unsetenv("STORE_INTERVAL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("RESTORE")
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("KEY")

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
		os.Unsetenv("ADDRESS")
		os.Unsetenv("STORE_INTERVAL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("RESTORE")
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("KEY")

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
		os.Setenv("ADDRESS", ":7070")
		os.Setenv("STORE_INTERVAL", "180")
		os.Setenv("FILE_STORAGE_PATH", "env.json")
		os.Setenv("RESTORE", "false")
		os.Setenv("DATABASE_DSN", "postgres://env")
		os.Setenv("KEY", "envkey")

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
		os.Setenv("ADDRESS", ":7070")
		os.Setenv("STORE_INTERVAL", "180")
		os.Setenv("FILE_STORAGE_PATH", "env.json")
		os.Setenv("RESTORE", "true")
		os.Setenv("DATABASE_DSN", "postgres://env")
		os.Setenv("KEY", "envkey")

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
		os.Unsetenv("ADDRESS")
		os.Unsetenv("STORE_INTERVAL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("RESTORE")
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("KEY")

		assert.Equal(t, "host=localhost port=5432 user=postgres password=admin dbname=MetricStore sslmode=disable", initDefaulthPathDB())
	})
}
