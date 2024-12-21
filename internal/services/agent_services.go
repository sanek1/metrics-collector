package services

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/sanek1/metrics-collector/internal/validation"
)

func SendToServer(client *http.Client, url string, m validation.Metrics, logger *log.Logger) error {
	ctx := context.Background()
	reqBody, err := buildMetrickBody(m)
	if err != nil {
		logger.Printf("Error building request body: %v", err)
		return err
	}
	/// compress request body
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(reqBody); err != nil {
		panic(err)
	}

	if err := zw.Close(); err != nil {
		panic(err)
	}
	compressedBody := buf.Bytes()
	///
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(compressedBody))
	if err != nil {
		logger.Printf("Error creating request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("content-encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	// compress request body

	cookie := &http.Cookie{
		Name:   "Token",
		Value:  "TEST_TOKEN",
		MaxAge: 360,
	}
	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("Error sending request: %v", err)
		return err
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
		}
		defer gzReader.Close()
		buf2 := new(bytes.Buffer)
		writer := bufio.NewWriter(buf2)
		if _, err := io.Copy(writer, gzReader); err != nil {
			log.Fatalf("error when copying: %v", err)
		}
		writer.Flush()

		os.Stdout.Write(buf2.Bytes())
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Printf("Error reading response body: %v", err)
			return err
		}
		os.Stdout.Write(body)
	}

	fmt.Fprintf(os.Stdout, " Url: %s status: %d\n", url, resp.StatusCode)

	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		logger.Printf("Error discarding response body: %v", err)
	}

	return nil
}

func buildMetrickBody(m validation.Metrics) ([]byte, error) {
	body, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return body, nil
}
