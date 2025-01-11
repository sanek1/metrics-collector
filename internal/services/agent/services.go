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
	"github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

func SendToServer(client *http.Client, url string, m models.Metrics, l *logging.ZapLogger) error {
	ctx := context.Background()
	body, err := json.Marshal(m)
	if err != nil {
		l.WarnCtx(ctx, "", zap.String("", fmt.Sprintf("Error building request body:%v", err)))
		return err
	}
	/// compress request body
	compressedBody, err := compressedBody(ctx, body, l)
	if err != nil {
		l.WarnCtx(ctx, "", zap.String("", fmt.Sprintf("Error compressing request body:%v", err)))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(compressedBody))
	if err != nil {
		l.WarnCtx(ctx, "", zap.String("", fmt.Sprintf("Error creating request:%v", err)))
		return err
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

	resp, err := client.Do(req)
	if err != nil {
		l.ErrorCtx(ctx, "compressedBody11", zap.String("", fmt.Sprintf("Error sending request:%v", err)))
		return err
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		buf, err := decompressReader(ctx, resp.Body, l)
		if err != nil {
			l.ErrorCtx(ctx, "compressedBody1", zap.String("", fmt.Sprintf("Error decompressing response body:%v", err)))
			return err
		}
		os.Stdout.Write(buf.Bytes())
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			l.ErrorCtx(ctx, "compressedBody2", zap.String("", fmt.Sprintf("Error reading response body:%v", err)))
			return err
		}
		os.Stdout.Write(body)
	}

	fmt.Fprintf(os.Stdout, "\nUrl: %s\nStatus: %d\n", url, resp.StatusCode)
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		l.ErrorCtx(ctx, "compressedBody3", zap.String("", fmt.Sprintf("Error discarding response body:%v", err)))
	}

	return nil
}

func compressedBody(ctx context.Context, body []byte, l *logging.ZapLogger) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(body)
	if err != nil {
		l.FatalCtx(ctx, "compressedBody4", zap.String("", fmt.Sprintf("Error compressing request body:%v", err)))
		return nil, err
	}
	err = zw.Close()
	if err != nil {
		l.FatalCtx(ctx, "compressedBody5", zap.String("", fmt.Sprintf("gzip close error:%v", err)))
		return nil, err
	}
	compressedBody := buf.Bytes()
	return compressedBody, err
}

func decompressReader(ctx context.Context, body io.ReadCloser, l *logging.ZapLogger) (*bytes.Buffer, error) {
	const maxDecompressedSize = 20 // 20 MB

	gzReader, err := gzip.NewReader(body)
	if err != nil {
		l.FatalCtx(ctx, "decompressReader", zap.String("", fmt.Sprintf("Error reading response body:%v", err)))
	}
	defer gzReader.Close()

	buf := new(bytes.Buffer)
	writer := bufio.NewWriter(buf)

	_, err = io.CopyN(writer, gzReader, maxDecompressedSize)
	if err != nil {
		if err == io.EOF {
			err = fmt.Errorf("decompressed data exceeds maximum size limit")
		}
		l.FatalCtx(ctx, "decompressReader", zap.String("", fmt.Sprintf("Error when copying:%v", err)))
		return nil, err
	}

	writer.Flush()
	return buf, nil
}
