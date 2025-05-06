package services

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"io"
	"net"
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/sanek1/metrics-collector/internal/crypto"
	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
	"github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

const (
	batchSizeBuffer     = 10
	maxDecompressedSize = 10 * 1024 * 1024 // 10 MB
)

// EncryptFunc - тип функции для шифрования данных
type EncryptFunc func(ctx context.Context, data []byte) ([]byte, error)

type Services struct {
	options    *flags.Options
	l          *l.ZapLogger
	publicKey  *rsa.PublicKey
	useEncrypt bool
	// encryptData - функция шифрования данных, может быть заменена в тестах
	encryptData EncryptFunc
}

// NewServices создает новый экземпляр Services
// Возвращает *Services и ошибку, если не удалось загрузить ключ
func NewServices(options *flags.Options, zl *l.ZapLogger) *Services {
	var publicKey *rsa.PublicKey
	useEncrypt := false

	s := &Services{
		options:    options,
		l:          zl,
		publicKey:  nil,
		useEncrypt: false,
	}

	// Загружаем публичный ключ, если указан путь
	if options.CryptoKey != "" {
		var err error
		publicKey, err = crypto.LoadPublicKey(options.CryptoKey)
		if err != nil {
			zl.ErrorCtx(context.Background(), "Failed to load public key", zap.Error(err))
			// Возвращаем сервис без шифрования в случае ошибки
			s.encryptData = s.defaultEncryptData
			return s
		}
		useEncrypt = true
		zl.InfoCtx(context.Background(), "Public key loaded successfully, encryption enabled")

		// Обновляем поля после успешной загрузки ключа
		s.publicKey = publicKey
		s.useEncrypt = useEncrypt
	}

	// Устанавливаем функцию шифрования по умолчанию
	s.encryptData = s.defaultEncryptData

	return s
}

// defaultEncryptData - стандартная реализация шифрования данных
func (s Services) defaultEncryptData(ctx context.Context, data []byte) ([]byte, error) {
	if !s.useEncrypt || s.publicKey == nil {
		return data, nil
	}

	encrypted, err := crypto.EncryptData(s.publicKey, data)
	if err != nil {
		s.l.ErrorCtx(ctx, "Error encrypting data", zap.Error(err))
		return nil, err
	}

	s.l.InfoCtx(ctx, "Data encrypted successfully", zap.Int("encrypted_size", len(encrypted)))
	return encrypted, nil
}

func (s Services) preparingMetrics(ctx context.Context, url string, body []byte) (*http.Request, error) {
	var processedBody []byte
	var err error

	// Шифруем данные, если публичный ключ доступен
	processedBody, err = s.encryptData(ctx, body)
	if err != nil {
		return nil, err
	}

	// Сжимаем данные
	compressedBody, err := s.compressedBody(ctx, processedBody)
	if err != nil {
		s.l.WarnCtx(ctx, "Error compressing request body", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(compressedBody))
	if err != nil {
		s.l.WarnCtx(ctx, "Error creating request", zap.Error(err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("content-encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Real-IP", getLocalIP())

	// Добавляем заголовок, указывающий что данные зашифрованы
	if s.useEncrypt {
		req.Header.Set("X-Encrypted", "true")
	}

	return req, nil
}
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return ""
}

func (s Services) processingResponseServer(ctx context.Context, resp *http.Response) error {
	if resp.Header.Get("Content-Encoding") == "gzip" {
		buf, err := s.decompressBody(ctx, resp.Body)
		if err != nil {
			s.l.ErrorCtx(ctx, "Error decompressing response body", zap.Error(err))
			return err
		}
		_, err = os.Stdout.Write(buf.Bytes())
		if err != nil {
			s.l.ErrorCtx(ctx, "Error writing response body", zap.Error(err))
		}
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			s.l.ErrorCtx(ctx, "Error reading response body", zap.Error(err))
			return err
		}
		_, err = os.Stdout.Write(body)
		if err != nil {
			s.l.ErrorCtx(ctx, "Error writing response body", zap.Error(err))
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		s.l.ErrorCtx(ctx, "Error discarding response body", zap.Error(err))
	}
	return nil
}

func (s Services) compressedBody(ctx context.Context, body []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(body); err != nil {
		s.l.ErrorCtx(ctx, "error compressing body content", zap.Error(err))
		return nil, err
	}
	if err := zw.Close(); err != nil {
		s.l.ErrorCtx(ctx, "error closing gzip", zap.Error(err))
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s Services) decompressBody(ctx context.Context, body io.ReadCloser) (*bytes.Buffer, error) {
	gzReader, err := gzip.NewReader(body)
	if err != nil {
		s.l.FatalCtx(ctx, "Error reading response body", zap.Error(err))
	}
	defer func() {
		_ = gzReader.Close()
	}()

	buf := new(bytes.Buffer)
	writer := bufio.NewWriter(buf)
	limitedReader := io.LimitReader(gzReader, maxDecompressedSize)

	_, err = io.Copy(writer, limitedReader)
	if err != nil && err != io.EOF {
		s.l.FatalCtx(ctx, "Error when copying", zap.Error(err))
		return nil, err
	}

	if err := writer.Flush(); err != nil {
		s.l.FatalCtx(ctx, "Error flushing writer", zap.Error(err))
		return nil, err
	}

	return buf, nil
}

func (s Services) worker(ctx context.Context, client *http.Client, url string, jobs <-chan models.Metrics) {
	var metrics []models.Metrics
	for j := range jobs {
		metrics = append(metrics, j)
		if len(metrics) == batchSizeBuffer {
			s.sendMetricsBatch(ctx, client, url, metrics)
			metrics = nil
		}
	}
	s.sendMetricsBatch(ctx, client, url, metrics)
}

func (s Services) sendMetricsBatch(ctx context.Context, client *http.Client, url string, metrics []models.Metrics) {
	if len(metrics) > 0 {
		if err := s.SendToServerBatchMetrics(ctx, client, url, metrics); err != nil {
			s.l.ErrorCtx(ctx, "SendToServer3 failed", zap.Error(err))
		}
	}
}
