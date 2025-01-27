package controller

import (
	"context"
	"net/http"

	sf "github.com/sanek1/metrics-collector/internal/flags/server"
	"github.com/sanek1/metrics-collector/internal/routing"
	ss "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type Controller struct {
	storage    ss.Storage
	fieStorage ss.FileStorage
	router     http.Handler
	logger     *l.ZapLogger
}

func NewController(fs ss.FileStorage, s ss.Storage, opt *sf.ServerOptions, logger *l.ZapLogger) *Controller {
	r := routing.NewRouting(s, opt, logger)
	return &Controller{
		storage:    s,
		fieStorage: fs,
		router:     r.InitRouting(),
		logger:     logger,
	}
}

func (c *Controller) Router() http.Handler {
	c.logger.InfoCtx(context.Background(), "init router")
	return c.router
}
