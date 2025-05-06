package validation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter(subnet string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CheckTrustedSubnetMiddleware(subnet))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	return r
}

func TestTrustedIPAllowed(t *testing.T) {
	router := setupRouter("192.168.1.0/24")
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.42")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestUntrustedIPBlocked(t *testing.T) {
	router := setupRouter("192.168.1.0/24")
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestInvalidIPBlocked(t *testing.T) {
	router := setupRouter("192.168.1.0/24")
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "not-an-ip")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestNoSubnetAllowsAll(t *testing.T) {
	router := setupRouter("")
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}
