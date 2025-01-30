package services

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"os"

	flags "github.com/sanek1/metrics-collector/internal/flags/agent"
	"github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

const (
	batchSizeBuffer     = 10
	maxDecompressedSize = 10 * 1024 * 1024 // 10 MB
)

type Services struct {
	options *flags.Options
	l       *l.ZapLogger
}

func NewServices(options *flags.Options, zl *l.ZapLogger) *Services {
	return &Services{
		options: options,
		l:       zl,
	}
}

func (s Services) preparingMetrics(ctx context.Context, url string, body []byte) (*http.Request, error) {
	compressedBody, err := s.compressedBody(ctx, body)
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
	return req, nil
}

func (s Services) processingResponseServer(ctx context.Context, resp *http.Response) error {
	if resp.Header.Get("Content-Encoding") == "gzip" {
		buf, err := s.decompressBody(ctx, resp.Body)
		if err != nil {
			s.l.ErrorCtx(ctx, "Error decompressing response body", zap.Error(err))
			return err
		}
		os.Stdout.Write(buf.Bytes())
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			s.l.ErrorCtx(ctx, "Error reading response body", zap.Error(err))
			return err
		}
		os.Stdout.Write(body)
	}

	defer resp.Body.Close()

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
	defer gzReader.Close()

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
