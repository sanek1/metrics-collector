package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
	"github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

func TestCompressedBody(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	s := &Services{
		options:    &flags.Options{},
		l:          logger,
		publicKey:  nil,
		useEncrypt: false,
	}

	ctx := context.Background()
	testData := []byte("test data for compression")

	// Test compression of data
	compressed, err := s.compressedBody(ctx, testData)
	require.NoError(t, err)
	assert.NotNil(t, compressed)
	assert.NotEmpty(t, compressed)

	assert.Less(t, len(compressed), len(testData)*2)

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	require.NoError(t, err)

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, testData, decompressed)
}

func TestDecompressBody(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	s := &Services{
		options:    &flags.Options{},
		l:          logger,
		publicKey:  nil,
		useEncrypt: false,
	}

	ctx := context.Background()
	testData := []byte("test data for decompression")

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write(testData)
	require.NoError(t, err)
	err = gzipWriter.Close()
	require.NoError(t, err)

	readCloser := io.NopCloser(bytes.NewReader(buf.Bytes()))

	decompressed, err := s.decompressBody(ctx, readCloser)
	require.NoError(t, err)
	assert.NotNil(t, decompressed)

	assert.Equal(t, testData, decompressed.Bytes())
}

func TestPreparingMetrics(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	t.Run("WithoutEncryption", func(t *testing.T) {
		s := &Services{
			options:    &flags.Options{},
			l:          logger,
			publicKey:  nil,
			useEncrypt: false,
			encryptData: func(ctx context.Context, data []byte) ([]byte, error) {
				return data, nil
			},
		}

		ctx := context.Background()
		url := "http://example.com/metrics"
		body := []byte(`{"id":"test","type":"gauge","value":123.45}`)

		req, err := s.preparingMetrics(ctx, url, body)
		require.NoError(t, err)
		assert.NotNil(t, req)

		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, url, req.URL.String())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", req.Header.Get("content-encoding"))
		assert.Equal(t, "gzip", req.Header.Get("Accept-Encoding"))
		assert.Empty(t, req.Header.Get("X-Encrypted"), "Заголовок X-Encrypted не должен быть установлен")

		assert.NotNil(t, req.Body)
	})

	t.Run("WithEncryption", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		s := &Services{
			options:    &flags.Options{},
			l:          logger,
			publicKey:  &privateKey.PublicKey,
			useEncrypt: true,
			encryptData: func(ctx context.Context, data []byte) ([]byte, error) {
				return data, nil
			},
		}

		ctx := context.Background()
		url := "http://example.com/metrics"
		body := []byte(`{"id":"test","type":"gauge","value":123.45}`)

		req, err := s.preparingMetrics(ctx, url, body)
		require.NoError(t, err)
		assert.NotNil(t, req)

		assert.Equal(t, "true", req.Header.Get("X-Encrypted"))
	})
}

func TestProcessingResponseServer(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	s := &Services{
		options:    &flags.Options{},
		l:          logger,
		publicKey:  nil,
		useEncrypt: false,
	}

	ctx := context.Background()

	t.Run("WithoutGzip", func(t *testing.T) {
		body := []byte(`{"status":"ok"}`)
		resp := &http.Response{
			Body:   io.NopCloser(bytes.NewReader(body)),
			Header: make(http.Header),
		}

		err := s.processingResponseServer(ctx, resp)
		require.NoError(t, err)
	})

	t.Run("WithGzip", func(t *testing.T) {
		body := []byte(`{"status":"ok"}`)

		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)
		_, err := gzipWriter.Write(body)
		require.NoError(t, err)
		err = gzipWriter.Close()
		require.NoError(t, err)

		resp := &http.Response{
			Body:   io.NopCloser(bytes.NewReader(buf.Bytes())),
			Header: make(http.Header),
		}
		resp.Header.Set("Content-Encoding", "gzip")

		err = s.processingResponseServer(ctx, resp)
		require.NoError(t, err)
	})
}

func TestDefaultEncryptData(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	t.Run("WithoutEncryption", func(t *testing.T) {
		s := &Services{
			options:    &flags.Options{},
			l:          logger,
			publicKey:  nil,
			useEncrypt: false,
		}

		ctx := context.Background()
		data := []byte("test data")

		result, err := s.defaultEncryptData(ctx, data)
		require.NoError(t, err)
		assert.Equal(t, data, result, "Данные не должны быть изменены при отключенном шифровании")
	})

	t.Run("WithEncryption", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		s := &Services{
			options:    &flags.Options{},
			l:          logger,
			publicKey:  &privateKey.PublicKey,
			useEncrypt: true,
		}

		ctx := context.Background()
		data := []byte("test data")

		result, err := s.defaultEncryptData(ctx, data)
		require.NoError(t, err)
		assert.NotEqual(t, data, result, "Данные должны быть изменены при включенном шифровании")

		assert.Greater(t, len(result), len(data))
	})
}

func TestNewServices(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	t.Run("WithoutCryptoKey", func(t *testing.T) {
		options := &flags.Options{
			CryptoKey: "",
		}

		s := NewServices(options, logger)
		assert.NotNil(t, s)
		assert.False(t, s.useEncrypt)
		assert.Nil(t, s.publicKey)
		assert.NotNil(t, s.encryptData)
	})

	t.Run("WithInvalidCryptoKey", func(t *testing.T) {
		invalidKeyFile, err := os.CreateTemp("", "invalid_key_*.pem")
		require.NoError(t, err)
		defer os.Remove(invalidKeyFile.Name())

		_, err = invalidKeyFile.WriteString("invalid key data")
		require.NoError(t, err)
		err = invalidKeyFile.Close()
		require.NoError(t, err)

		options := &flags.Options{
			CryptoKey: invalidKeyFile.Name(),
		}

		s := NewServices(options, logger)
		assert.NotNil(t, s)
		assert.False(t, s.useEncrypt)
		assert.Nil(t, s.publicKey)
		assert.NotNil(t, s.encryptData)
	})

	t.Run("WithValidCryptoKey", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		require.NoError(t, err)

		publicKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		})

		validKeyFile, err := os.CreateTemp("", "valid_key_*.pem")
		require.NoError(t, err)
		defer os.Remove(validKeyFile.Name())

		_, err = validKeyFile.Write(publicKeyPEM)
		require.NoError(t, err)
		err = validKeyFile.Close()
		require.NoError(t, err)

		options := &flags.Options{
			CryptoKey: validKeyFile.Name(),
		}

		s := NewServices(options, logger)
		assert.NotNil(t, s)
		assert.True(t, s.useEncrypt)
		assert.NotNil(t, s.publicKey)
		assert.NotNil(t, s.encryptData)
	})
}

func TestWorkerAndSendMetricsBatch(t *testing.T) {
	logger, _ := l.NewZapLogger(zap.InfoLevel)

	s := &Services{
		options:    &flags.Options{},
		l:          logger,
		publicKey:  nil,
		useEncrypt: false,
		encryptData: func(ctx context.Context, data []byte) ([]byte, error) {
			return data, nil
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()

	t.Run("SendMetricsBatchEmpty", func(t *testing.T) {
		var metrics []models.Metrics
		s.sendMetricsBatch(ctx, server.Client(), server.URL, metrics)
	})

	t.Run("SendMetricsBatchNonEmpty", func(t *testing.T) {
		gaugeValue := float64(123.45)
		metrics := []models.Metrics{
			{
				ID:    "test_gauge",
				MType: "gauge",
				Value: &gaugeValue,
			},
		}
		s.sendMetricsBatch(ctx, server.Client(), server.URL, metrics)
	})

	t.Run("Worker", func(t *testing.T) {
		gaugeValue := float64(123.45)
		jobs := make(chan models.Metrics, 2)

		jobs <- models.Metrics{
			ID:    "test_gauge1",
			MType: "gauge",
			Value: &gaugeValue,
		}
		jobs <- models.Metrics{
			ID:    "test_gauge2",
			MType: "gauge",
			Value: &gaugeValue,
		}
		close(jobs)

		s.worker(ctx, server.Client(), server.URL, jobs)
	})
}
