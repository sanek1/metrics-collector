package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/sanek1/metrics-collector/internal/crypto"
	m "github.com/sanek1/metrics-collector/internal/models"
	"github.com/sanek1/metrics-collector/internal/storage/server/mocks"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

func TestDecryptionOfAgentData(t *testing.T) {
	// Generate temporary keys for test
	privateKeyPath := "test_private.pem"
	publicKeyPath := "test_public.pem"

	// Generate key pair
	err := crypto.GenerateKeyPair(privateKeyPath, publicKeyPath)
	require.NoError(t, err, "error generating keys")

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}()

	// Load only public key for encryption, private key will be loaded from file by service
	_, err = crypto.LoadPrivateKey(privateKeyPath) // Check only that key is loaded
	require.NoError(t, err, "error loading private key")
	publicKey, err := crypto.LoadPublicKey(publicKeyPath)
	require.NoError(t, err, "error loading public key")

	// Create test data
	gaugeValue := float64(123.45)
	testMetric := m.Metrics{
		ID:    "TestMetric",
		MType: "gauge",
		Value: &gaugeValue,
	}

	// Encode metric to JSON
	jsonData, err := json.Marshal(testMetric)
	require.NoError(t, err, "error encoding metric to JSON")

	// Encrypt data with public key
	encryptedData, err := crypto.EncryptData(publicKey, jsonData)
	require.NoError(t, err, "error encrypting data")

	// Setup Gin for testing
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	mockStorage := mocks.NewStorage(t)

	// Create service with decryption support
	services := NewHandlerServices(mockStorage, nil, privateKeyPath, logger)

	// Check that decryption is enabled
	assert.True(t, services.useDecrypt, "Decryption should be enabled")
	assert.NotNil(t, services.privateKey, "Private key should be loaded")

	// Create test request with encrypted data
	req, _ := http.NewRequest("POST", "/update", bytes.NewBuffer(encryptedData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Encrypted", "true") // Specify that data is encrypted

	// Create test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Parse metrics from request (with decryption)
	metrics, err := services.ParseMetricsServices(c)

	// Check results
	require.NoError(t, err, "error parsing metrics")
	require.Len(t, metrics, 1, "should be one metric")

	// Check that metric is decrypted correctly
	assert.Equal(t, "TestMetric", metrics[0].ID)
	assert.Equal(t, "gauge", metrics[0].MType)
	assert.Equal(t, gaugeValue, *metrics[0].Value)
}

// Test checks processing of unencrypted data
func TestHandlingUnencryptedData(t *testing.T) {
	// Setup Gin for testing
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	mockStorage := mocks.NewStorage(t)

	// Create test data
	gaugeValue := float64(123.45)
	testMetric := m.Metrics{
		ID:    "TestMetric",
		MType: "gauge",
		Value: &gaugeValue,
	}

	// Create service without decryption support
	services := NewHandlerServices(mockStorage, nil, "", logger)

	// Check that decryption is disabled
	assert.False(t, services.useDecrypt, "Decryption should be disabled")
	assert.Nil(t, services.privateKey, "Private key should be nil")

	// Encode metric to JSON (without encryption)
	jsonData, err := json.Marshal(testMetric)
	require.NoError(t, err, "Error encoding metric to JSON")

	// Create test request with unencrypted data
	req, _ := http.NewRequest("POST", "/update", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	// Do not set X-Encrypted header

	// Create test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Parse metrics from request (without decryption)
	metrics, err := services.ParseMetricsServices(c)

	// Check results
	require.NoError(t, err, "error parsing metrics")
	require.Len(t, metrics, 1, "should be one metric")

	// Check that metric is processed correctly
	assert.Equal(t, "TestMetric", metrics[0].ID)
	assert.Equal(t, "gauge", metrics[0].MType)
	assert.Equal(t, gaugeValue, *metrics[0].Value)
}

// Test checks error when decrypting with wrong private key
func TestDecryptionWithWrongKey(t *testing.T) {
	// Generate two key pairs
	privateKeyPath1 := "test_private1.pem"
	publicKeyPath1 := "test_public1.pem"
	privateKeyPath2 := "test_private2.pem"
	publicKeyPath2 := "test_public2.pem"

	// Generate first key pair
	err := crypto.GenerateKeyPair(privateKeyPath1, publicKeyPath1)
	require.NoError(t, err, "error generating first key pair")

	// Generate second key pair
	err = crypto.GenerateKeyPair(privateKeyPath2, publicKeyPath2)
	require.NoError(t, err, "error generating second key pair")

	defer func() {
		os.Remove(privateKeyPath1)
		os.Remove(publicKeyPath1)
		os.Remove(privateKeyPath2)
		os.Remove(publicKeyPath2)
	}()

	// Load public key from first pair (for encryption)
	publicKey, err := crypto.LoadPublicKey(publicKeyPath1)
	require.NoError(t, err, "error loading public key")

	// Setup Gin for testing
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	mockStorage := mocks.NewStorage(t)

	// Create test data
	gaugeValue := float64(123.45)
	testMetric := m.Metrics{
		ID:    "TestMetric",
		MType: "gauge",
		Value: &gaugeValue,
	}

	// Encode metric to JSON
	jsonData, err := json.Marshal(testMetric)
	require.NoError(t, err, "error encoding metric to JSON")

	// Encrypt data with public key from first pair
	encryptedData, err := crypto.EncryptData(publicKey, jsonData)
	require.NoError(t, err, "error encrypting data")

	// Create service with private key from second pair (wrong key for decryption)
	services := NewHandlerServices(mockStorage, nil, privateKeyPath2, logger)

	// Create test request with encrypted data
	req, _ := http.NewRequest("POST", "/update", bytes.NewBuffer(encryptedData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Encrypted", "true") // Specify that data is encrypted

	// Create test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Parse metrics from request (should be decryption error)
	_, err = services.ParseMetricsServices(c)

	// Check that decryption error is received
	require.Error(t, err, "should be error when decrypting with wrong key")
	assert.Contains(t, err.Error(), "decryption", "error should be related to decryption")
}

// Test checks error when decrypting without private key
func TestEncryptedDataWithoutKey(t *testing.T) {
	// Generate temporary keys for test
	privateKeyPath := "test_private.pem"
	publicKeyPath := "test_public.pem"

	// Generate key pair
	err := crypto.GenerateKeyPair(privateKeyPath, publicKeyPath)
	require.NoError(t, err, "error generating keys")

	defer func() {
		os.Remove(privateKeyPath)
		os.Remove(publicKeyPath)
	}()

	// Load public key
	publicKey, err := crypto.LoadPublicKey(publicKeyPath)
	require.NoError(t, err, "error loading public key")

	// Setup Gin for testing
	gin.SetMode(gin.TestMode)
	logger, _ := l.NewZapLogger(zap.InfoLevel)
	mockStorage := mocks.NewStorage(t)

	// Create test data
	gaugeValue := float64(123.45)
	testMetric := m.Metrics{
		ID:    "TestMetric",
		MType: "gauge",
		Value: &gaugeValue,
	}

	// Encode metric to JSON
	jsonData, err := json.Marshal(testMetric)
	require.NoError(t, err, "error encoding metric to JSON")

	// Encrypt data with public key
	encryptedData, err := crypto.EncryptData(publicKey, jsonData)
	require.NoError(t, err, "error encrypting data")

	// Create service without private key
	services := NewHandlerServices(mockStorage, nil, "", logger)

	// Create test request with encrypted data
	req, _ := http.NewRequest("POST", "/update", bytes.NewBuffer(encryptedData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Encrypted", "true") // Specify that data is encrypted

	// Create test Gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Parse metrics from request (should be error due to missing key)
	_, err = services.ParseMetricsServices(c)

	// Check that error due to missing key is received
	require.Error(t, err, "should be error when missing key for decryption")
	assert.Contains(t, err.Error(), "no private key", "error should be related to missing key")
}
