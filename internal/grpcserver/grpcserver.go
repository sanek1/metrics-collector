package grpcserver

import (
	"context"
	"log"
	"net"

	"github.com/sanek1/metrics-collector/cmd/grpc/metricsgrpc"
	m "github.com/sanek1/metrics-collector/internal/models"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	"google.golang.org/grpc"
)

type Server struct {
	metricsgrpc.UnimplementedMetricsServiceServer
	Storage storage.Storage
}

func (s *Server) SendMetrics(ctx context.Context, list *metricsgrpc.MetricList) (*metricsgrpc.Empty, error) {
	var models []m.Metrics

	for _, l := range list.Metrics {
		metric := m.Metrics{
			ID:    l.ID,
			MType: l.MType,
			Delta: &l.Delta,
			Value: &l.Value,
		}

		models = append(models, metric)
	}

	_, err := s.Storage.SetGauge(ctx, models...)
	if err != nil {
		return nil, err

	}

	return &metricsgrpc.Empty{}, nil
}

func RunGRPCServer(addr string, storage storage.Storage) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	srv := &Server{Storage: storage}
	metricsgrpc.RegisterMetricsServiceServer(grpcServer, srv)

	log.Printf("gRPC server listening on %s", addr)
	return grpcServer.Serve(lis)
}
