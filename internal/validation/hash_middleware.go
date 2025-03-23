package validation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Secret struct {
	SecretKey string `json:"key"`
}

func NewHash(key string) *Secret {
	return &Secret{
		SecretKey: key,
	}
}

func (s *Secret) HashMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Unable to read request body",
			})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		dataToHash := string(body) + s.SecretKey
		hash := sha256.New()
		hash.Write([]byte(dataToHash))
		hashSum := hash.Sum(nil)
		hashHex := hex.EncodeToString(hashSum)
		c.Header("HashSHA256", hashHex)
		c.Next()
	}
}

func (s *Secret) VerifyHash(body []byte, providedHash string) bool {
	dataToHash := string(body) + s.SecretKey

	hash := sha256.New()
	hash.Write([]byte(dataToHash))
	hashSum := hash.Sum(nil)

	hashHex := hex.EncodeToString(hashSum)
	return hashHex == providedHash
}
