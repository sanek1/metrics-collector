// package services
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"go.uber.org/zap"

	"github.com/sanek1/metrics-collector/internal/models"
)

func (s Services) SendToServerAsync(ctx context.Context, client *http.Client, url string, m []models.Metrics) error {
	var numJobs = s.options.RateLimit
	jobs := make(chan models.Metrics, numJobs)

	var wg sync.WaitGroup
	wg.Add(int(numJobs))
	for i := 0; i < int(numJobs); i++ {
		go func(id int) {
			defer wg.Done()
			s.l.InfoCtx(ctx, "Starting worker"+fmt.Sprint(id), zap.Int("worker_id", id))
			s.worker(ctx, client, url, jobs)
		}(i)
	}

	go func() {
		for _, j := range m {
			jobs <- j
		}
		close(jobs)
	}()

	wg.Wait()
	return nil
}

func (s Services) SendToServerBatchMetrics(ctx context.Context, client *http.Client, url string, m []models.Metrics) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		s.l.ErrorCtx(ctx, "Error encoding metrics to JSON",
			zap.Error(err),
		)
		return fmt.Errorf("error encoding metrics: %w", err)
	}

	req, err := s.preparingMetrics(ctx, url, buf.Bytes())
	if err != nil {
		return err
	}

	return s.sendToServer(ctx, client, req)
}
func (s Services) SendToServerMetric(ctx context.Context, client *http.Client, url string, model models.Metrics) error {
	body, err := json.Marshal(model)
	if err != nil {
		s.l.ErrorCtx(ctx, "Error encoding metric to JSON",
			zap.Error(err),
		)
		return fmt.Errorf("error encoding metric: %w", err)
	}

	req, err := s.preparingMetrics(ctx, url, body)
	if err != nil {
		return err
	}

	return s.sendToServer(ctx, client, req)
}

func (s Services) sendToServer(ctx context.Context, client *http.Client, req *http.Request) error {
	resp, err := client.Do(req)
	if err != nil {
		s.l.ErrorCtx(ctx, "sendToServer Request sending failed",
			zap.String("method", req.Method),
			zap.String("url", req.URL.String()),
			zap.Error(err),
		)
		return fmt.Errorf("request sending error: %w", err)
	}

	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if err := s.processingResponseServer(ctx, resp); err != nil {
		s.l.ErrorCtx(ctx, "Response processing failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
			zap.Error(err),
		)
		return fmt.Errorf("response processing error: %w", err)
	}

	s.l.InfoCtx(ctx, "Metrics batch successfully sent")
	return nil
}
