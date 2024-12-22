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

	"github.com/sanek1/metrics-collector/internal/validation"
	"go.uber.org/zap"
)

func SendToServer(client *http.Client, url string, m validation.Metrics, logger *zap.SugaredLogger) error {
	ctx := context.Background()
	body, err := json.Marshal(m)
	if err != nil {
		logger.Warnln("Error building request body: %v", err)
		return err
	}
	/// compress request body
	compressedBody, err := compressedBody(body, logger)
	if err != nil {
		logger.Warnln("Error compressing request body: %v", err)
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(compressedBody))
	if err != nil {
		logger.Warnln("Error creating request: %v", err)
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
		logger.Warnln("Error sending request: %v", err)
		return err
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		buf, err := decompressReader(resp.Body, logger)
		if err != nil {
			logger.Warnln("Error decompressing response body: %v", err)
			return err
		}
		os.Stdout.Write(buf.Bytes())
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Infoln("Error reading response body: %v", err)
			return err
		}
		os.Stdout.Write(body)
	}

	fmt.Fprintf(os.Stdout, "\nUrl: %s\nStatus: %d\n", url, resp.StatusCode)
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		logger.Infoln("Error discarding response body: %v", err)
	}

	return nil
}

func compressedBody(body []byte, logger *zap.SugaredLogger) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(body)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	err = zw.Close()
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	compressedBody := buf.Bytes()
	return compressedBody, err
}

func decompressReader(body io.ReadCloser, logger *zap.SugaredLogger) (*bytes.Buffer, error) {
	const maxDecompressedSize = 10 * 1024 * 1024 // 10 MB

	gzReader, err := gzip.NewReader(body)
	if err != nil {
		logger.Fatalf("Error reading response body: %v", err)
	}
	defer gzReader.Close()

	buf := new(bytes.Buffer)
	writer := bufio.NewWriter(buf)

	_, err = io.CopyN(writer, gzReader, maxDecompressedSize)
	if err != nil {
		if err == io.EOF {
			err = fmt.Errorf("decompressed data exceeds maximum size limit")
		}
		logger.Fatalf("Error when copying: %v", err)
		return nil, err
	}

	writer.Flush()
	return buf, nil
}
