package validation

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(GzipMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test response")
	})

	t.Run("with gzip compression", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))
		gr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)
		defer gr.Close()

		body, err := io.ReadAll(gr)
		require.NoError(t, err)

		assert.Equal(t, "test response", string(body))
	})

	t.Run("without gzip compression", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Empty(t, resp.Header().Get("Content-Encoding"))
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, "test response", string(body))
	})
}

func TestGzipMiddleware_ContentType(t *testing.T) {
	router := gin.New()
	router.Use(GzipMiddleware())

	router.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/json", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, "application/json; charset=utf-8", resp.Header().Get("Content-Type"))
	assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))
}
