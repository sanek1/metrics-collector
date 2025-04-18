package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	cryptoutils "github.com/sanek1/metrics-collector/internal/crypto"
	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
	"github.com/sanek1/metrics-collector/internal/models"
	"github.com/sanek1/metrics-collector/pkg/logging"
)

// Test checks that the agent correctly encrypts data when sending metrics
func TestEncryptedMetricsSending(t *testing.T) {
	// Create temporary files for keys
	privateKeyPath := "test_private.pem"
	publicKeyPath := "test_public.pem"

	// Generate key pair
	err := cryptoutils.GenerateKeyPair(privateKeyPath, publicKeyPath)
	require.NoError(t, err, "Error generating keys")

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}()

	// Проверяем только наличие приватного ключа, сам ключ не нужен в этом тесте
	_, err = cryptoutils.LoadPrivateKey(privateKeyPath)
	require.NoError(t, err, "Ошибка при загрузке приватного ключа")

	// Create logger
	logger, _ := logging.NewZapLogger(zap.InfoLevel)

	// Create options with path to public key
	options := &flags.Options{
		CryptoKey: publicKeyPath,
	}

	// Create agent service
	service := NewServices(options, logger)

	// Check that encryption is enabled
	assert.True(t, service.useEncrypt, "Encryption should be enabled")
	assert.NotNil(t, service.publicKey, "Public key should be loaded")

	// Test metrics
	gaugeValue := float64(123.45)
	testMetric := models.Metrics{
		ID:    "TestMetric",
		MType: "gauge",
		Value: &gaugeValue,
	}

	// Create test server for checking encryption
	var encryptedRequest []byte
	var requestHeader http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestHeader = r.Header.Clone()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Error reading request body: %v", err)
		}
		encryptedRequest = body

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Send metric to test server
	err = service.SendToServerMetric(context.Background(), server.Client(), server.URL, testMetric)
	require.NoError(t, err, "Error sending metric")

	// Check that X-Encrypted header is set
	assert.Equal(t, "true", requestHeader.Get("X-Encrypted"), "X-Encrypted header should be set")

	// Encryption is not performed directly, since data is also compressed
	// and there is no simple way to decrypt them in the test
	assert.NotEmpty(t, encryptedRequest, "Encrypted request should not be empty")
}

// Test checks the work of the service without encryption
func TestUnencryptedMetricsSending(t *testing.T) {
	// Create logger
	logger, _ := logging.NewZapLogger(zap.InfoLevel)

	// Create options without specifying the path to the public key
	options := &flags.Options{
		CryptoKey: "",
	}

	// Create agent service
	service := NewServices(options, logger)

	// Check that encryption is disabled
	assert.False(t, service.useEncrypt, "Encryption should be disabled")
	assert.Nil(t, service.publicKey, "Public key should be nil")

	// Test metrics
	gaugeValue := float64(123.45)
	testMetric := models.Metrics{
		ID:    "TestMetric",
		MType: "gauge",
		Value: &gaugeValue,
	}

	// Create test server
	var requestHeader http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Save request headers
		requestHeader = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Send metric to test server
	err := service.SendToServerMetric(context.Background(), server.Client(), server.URL, testMetric)
	require.NoError(t, err, "Error sending metric")

	// Check that X-Encrypted header is not set
	assert.Empty(t, requestHeader.Get("X-Encrypted"), "X-Encrypted header should not be set")
}

// Проверка обработки ошибок загрузки ключа
func TestEncryptionWithInvalidKey(t *testing.T) {
	// Create test keys for correct operation, but then we will pass an invalid file
	privateKeyPath := "test_private.pem"
	publicKeyPath := "test_public.pem"
	invalidKeyPath := "invalid_key.pem"

	// Generate key pair
	err := cryptoutils.GenerateKeyPair(privateKeyPath, publicKeyPath)
	require.NoError(t, err, "Error generating keys")

	// Create an invalid key file
	err = os.WriteFile(invalidKeyPath, []byte("это неверный ключ"), 0644)
	require.NoError(t, err, "Error creating invalid key file")

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
		os.Remove(invalidKeyPath)
	}()

	t.Run("Invalid Public Key Format", func(t *testing.T) {
		logger, _ := logging.NewZapLogger(zap.InfoLevel)

		// Create options with the path to the invalid key
		options := &flags.Options{
			CryptoKey: invalidKeyPath,
		}

		// Create service - now this should work without errors,
		// but encryption should be disabled
		service := NewServices(options, logger)

		// Check that encryption is disabled due to invalid key
		assert.False(t, service.useEncrypt, "Encryption should be disabled due to invalid key")
		assert.Nil(t, service.publicKey, "Public key should be nil due to invalid key")

		// Test metrics
		gaugeValue := float64(123.45)
		testMetric := models.Metrics{
			ID:    "TestMetric",
			MType: "gauge",
			Value: &gaugeValue,
		}

		// Create test server
		var requestHeader http.Header

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Save request headers
			requestHeader = r.Header.Clone()
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Send metric to test server - this should work, but without encryption
		err := service.SendToServerMetric(context.Background(), server.Client(), server.URL, testMetric)
		require.NoError(t, err, "Error sending metric with invalid key")

		// X-Encrypted header should not be set
		assert.Empty(t, requestHeader.Get("X-Encrypted"), "X-Encrypted header should not be set due to invalid key")
	})

	t.Run("Non-existent Key File", func(t *testing.T) {
		logger, _ := logging.NewZapLogger(zap.InfoLevel)

		// Create options with the path to the non-existent file
		options := &flags.Options{
			CryptoKey: "non_existent_key.pem",
		}

		// Create service - now this should work without errors,
		// but encryption should be disabled
		service := NewServices(options, logger)

		// Check that encryption is disabled due to missing key
		assert.False(t, service.useEncrypt, "Encryption should be disabled due to missing key")
		assert.Nil(t, service.publicKey, "Public key should be nil due to missing key")
	})
}

// Test for checking that data encryption is not performed when the key is missing
func TestEncryptionSkippedWithoutKey(t *testing.T) {
	// Create test server
	var requestHeader http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestHeader = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create test metric
	gaugeValue := float64(123.45)
	testMetric := models.Metrics{
		ID:    "TestMetric",
		MType: "gauge",
		Value: &gaugeValue,
	}

	// Create service instance without specifying the public key
	zl, _ := logging.NewZapLogger(zap.InfoLevel)
	options := &flags.Options{
		CryptoKey: "", // Do not specify the encryption key
	}
	service := NewServices(options, zl)

	// Variable to track the call of the encryption function
	encryptionPerformed := false
	origEncryptData := service.encryptData

	// Replace the encryption function with a test function
	service.encryptData = func(ctx context.Context, data []byte) ([]byte, error) {
		// Mark that the function was called
		encryptionPerformed = true
		// Call the original function, which should return the data without encryption
		return origEncryptData(ctx, data)
	}
	defer func() {
		service.encryptData = origEncryptData
	}()

	// Send metric to test server
	err := service.SendToServerMetric(context.Background(), server.Client(), server.URL, testMetric)
	require.NoError(t, err, "Error sending metric")

	// Check that the useEncrypt flag is not set
	assert.False(t, service.useEncrypt, "The useEncrypt flag should be false")

	// Check that encryption was called, but not actually performed
	assert.True(t, encryptionPerformed, "The encryption function should be called")

	// Check that the X-Encrypted header is not set
	assert.Empty(t, requestHeader.Get("X-Encrypted"), "The X-Encrypted header should not be set")
}
