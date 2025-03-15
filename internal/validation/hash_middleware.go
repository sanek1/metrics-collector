package validation

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

type Secret struct {
	SecretKey string `json:"key"`
}

func NewHash(key string) *Secret {
	return &Secret{
		SecretKey: key,
	}
}

func (s *Secret) HashMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body", http.StatusBadRequest)
			return
		}

		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		datatoHash := string(body) + s.SecretKey

		hash := sha256.New()
		hash.Write([]byte(datatoHash))
		hashSum := hash.Sum(nil)

		hashHex := hex.EncodeToString(hashSum)
		w.Header().Set("HashSHA256", hashHex)

		h.ServeHTTP(w, r)
	})
}
