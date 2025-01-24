package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sanek1/metrics-collector/internal/models"
	"go.uber.org/zap"
)

func (s Services) SendToServer(ctx context.Context, client *http.Client, url string, m models.Metrics) error {
	// prepare request
	req, err := s.preparingMetrics(ctx, url, m)
	if err != nil {
		s.l.ErrorCtx(ctx, "compressedBody10", zap.String("", fmt.Sprintf("Error creating request:%v", err)))
		return err
	}
	//send request to server
	resp, err := client.Do(req)
	if err != nil {
		s.l.ErrorCtx(ctx, "compressedBody11", zap.String("", fmt.Sprintf("Error sending request:%v", err)))
		return err
	}
	//processing response
	if err := s.ProcessingResponseServer(ctx, resp); err != nil {
		s.l.ErrorCtx(ctx, "compressedBody12", zap.String("", fmt.Sprintf("Error processing response:%v", err)))
		return err
	}
	return nil
}
