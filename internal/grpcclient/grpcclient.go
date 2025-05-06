package grpcclient

import (
	"context"

	"github.com/sanek1/metrics-collector/cmd/grpc/metricsgrpc"
	m "github.com/sanek1/metrics-collector/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client metricsgrpc.MetricsServiceClient
}

func New(address string) (*Client, error) {
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	client := metricsgrpc.NewMetricsServiceClient(conn)
	return &Client{conn: conn, client: client}, nil
}

func (c *Client) SendMetrics(ctx context.Context, metrics []*metricsgrpc.Metric) error {
	_, err := c.client.SendMetrics(ctx, &metricsgrpc.MetricList{Metrics: metrics})
	return err
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func ConvertToGRPC(metrics []m.Metrics) []*metricsgrpc.Metric {
	var result []*metricsgrpc.Metric
	for _, m := range metrics {
		g := &metricsgrpc.Metric{
			ID:    m.ID,
			MType: m.MType,
		}
		if m.MType == "gauge" && m.Value != nil {
			g.Value = *m.Value
		}
		if m.MType == "counter" && m.Delta != nil {
			g.Delta = *m.Delta
		}
		result = append(result, g)
	}
	return result
}
