package validation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_VerifyHash(t *testing.T) {
	secret := NewHash("test-secret")

	t.Run("valid hash", func(t *testing.T) {
		body := []byte("test-data")
		data := string(body) + "test-secret"

		hash := sha256.Sum256([]byte(data))
		validHash := hex.EncodeToString(hash[:])

		assert.True(t, secret.VerifyHash(body, validHash))
	})

	t.Run("invalid hash", func(t *testing.T) {
		body := []byte("test-data")
		assert.False(t, secret.VerifyHash(body, "invalid-hash"))
	})

	t.Run("empty body", func(t *testing.T) {
		body := []byte{}
		data := string(body) + "test-secret"

		hash := sha256.Sum256([]byte(data))
		validHash := hex.EncodeToString(hash[:])

		assert.True(t, secret.VerifyHash(body, validHash))
	})
}

func Test_HashMiddleware(t *testing.T) {
	router := gin.New()
	secret := NewHash("test-key")

	router.Use(secret.HashMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("adds valid hash header", func(t *testing.T) {
		body := []byte(`{"data":"test"}`)

		data := string(body) + "test-key"
		hash := sha256.Sum256([]byte(data))
		expectedHash := hex.EncodeToString(hash[:])

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedHash, w.Header().Get("HashSHA256"))
	})

	t.Run("handles empty body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("HashSHA256"))
	})

	t.Run("handles read error", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", &errorReader{})
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"Unable to read request body"}`, w.Body.String())
	})
}

func TestNewHash(t *testing.T) {
	t.Run("creates secret with key", func(t *testing.T) {
		secret := NewHash("test-key")
		assert.Equal(t, "test-key", secret.SecretKey)
	})

	t.Run("creates secret with empty key", func(t *testing.T) {
		secret := NewHash("")
		assert.Equal(t, "", secret.SecretKey)
	})
}

type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
