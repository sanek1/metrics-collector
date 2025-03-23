package validation

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func GzipMiddleware() gin.HandlerFunc {
	return gzip.Gzip(gzip.DefaultCompression)
}
