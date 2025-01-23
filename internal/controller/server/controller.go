package controller

import (
	"context"
	"net/http"

	flags "github.com/sanek1/metrics-collector/internal/flags/server"
	"github.com/sanek1/metrics-collector/internal/routing"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type Controller struct {
	storage    storage.Storage
	fieStorage storage.FileStorage
	router     http.Handler
	logger     *l.ZapLogger
}

func NewController(fs storage.FileStorage, s storage.Storage, opt *flags.ServerOptions, logger *l.ZapLogger) *Controller {
	r := routing.New(s, opt, logger)
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
