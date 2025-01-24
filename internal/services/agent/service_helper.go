package services

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sanek1/metrics-collector/internal/models"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type Services struct {
	l *l.ZapLogger
}

func NewServices(l *l.ZapLogger) *Services {
	return &Services{l: l}
}

func (s Services) preparingMetrics(ctx context.Context, url string, m models.Metrics) (*http.Request, error) {
	body, err := json.Marshal(m)
	if err != nil {
		s.l.WarnCtx(ctx, "", zap.String("", fmt.Sprintf("Error building request body:%v", err)))
		return nil, err
	}
	/// compress request body
	compressedBody, err := s.compressedBody(ctx, body)
	if err != nil {
		s.l.WarnCtx(ctx, "", zap.String("", fmt.Sprintf("Error compressing request body:%v", err)))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(compressedBody))
	if err != nil {
		s.l.WarnCtx(ctx, "", zap.String("", fmt.Sprintf("Error creating request:%v", err)))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("content-encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	cookie := &http.Cookie{
		Name:   "Token",
		Value:  "TEST_TOKEN",
		MaxAge: 360,
	}
	req.AddCookie(cookie)
	return req, nil
}

func (s Services) ProcessingResponseServer(ctx context.Context, resp *http.Response) error {
	if resp.Header.Get("Content-Encoding") == "gzip" {
		buf, err := s.decompressReader(ctx, resp.Body)
		if err != nil {
			s.l.ErrorCtx(ctx, "compressedBody1", zap.String("", fmt.Sprintf("Error decompressing response body:%v", err)))
			return err
		}
		os.Stdout.Write(buf.Bytes())
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			s.l.ErrorCtx(ctx, "compressedBody2", zap.String("", fmt.Sprintf("Error reading response body:%v", err)))
			return err
		}
		os.Stdout.Write(body)
	}

	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		s.l.ErrorCtx(ctx, "compressedBody3", zap.String("", fmt.Sprintf("Error discarding response body:%v", err)))
	}
	return nil
}

func (s Services) compressedBody(ctx context.Context, body []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(body); err != nil {
		s.l.FatalCtx(ctx, "compressedBody4", zap.String("", fmt.Sprintf("Error compressing request body:%v", err)))
		return nil, err
	}
	if err := zw.Close(); err != nil {
		s.l.FatalCtx(ctx, "compressedBody5", zap.String("", fmt.Sprintf("gzip close error:%v", err)))
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s Services) decompressReader(ctx context.Context, body io.ReadCloser) (*bytes.Buffer, error) {
	const maxDecompressedSize = 20 // 20 MB

	gzReader, err := gzip.NewReader(body)
	if err != nil {
		s.l.FatalCtx(ctx, "decompressReader", zap.String("", fmt.Sprintf("Error reading response body:%v", err)))
	}
	defer gzReader.Close()

	buf := new(bytes.Buffer)
	writer := bufio.NewWriter(buf)

	_, err = io.Copy(writer, gzReader)
	if err != nil && err != io.EOF {
		s.l.FatalCtx(ctx, "decompressReader", zap.String("", fmt.Sprintf("Error when copying:%v", err)))
		return nil, err
	}

	writer.Flush()
	return buf, nil
}
