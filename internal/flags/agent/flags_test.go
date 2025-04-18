package flags

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
