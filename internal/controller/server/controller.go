package controller

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/sanek1/metrics-collector/internal/routing"
	storage "github.com/sanek1/metrics-collector/internal/storage/server"
	l "github.com/sanek1/metrics-collector/pkg/logging"
	"go.uber.org/zap"
)

type Controller struct {
	storage storage.Storage
	router  http.Handler
	logger  *l.ZapLogger
}

func New(s storage.Storage, db *sql.DB, logger *l.ZapLogger) *Controller {
	r := routing.New(s, db, logger)
	return &Controller{
		storage: s,
		router:  r.InitRouting(),
		logger:  logger,
	}
}

func (c *Controller) Router() http.Handler {
	c.logger.InfoCtx(context.Background(), "init router")
	return c.router
}

func (c *Controller) PeriodicallySaveBackUp(ctx context.Context, filename string, restore bool, interval time.Duration) {
	ctx = c.logger.WithContextFields(ctx,
		zap.String("app", "logging"),
		zap.String("service", "main"))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	if restore {
		err := c.storage.LoadFromFile(filename)
		if err != nil {
			c.logger.ErrorCtx(ctx, "Error loading metrics from file")
		}
	}

	for range ticker.C {
		err := c.storage.SaveToFile(filename)
		c.logger.InfoCtx(ctx, "saving to file was successful")
		if err != nil {
			c.logger.ErrorCtx(ctx, "Error saving metrics to file"+err.Error())
		}
	}

	for {
		select {
		case <-ticker.C:
			err := c.storage.SaveToFile(filename)
			c.logger.InfoCtx(ctx, "saving to file was successful")
			if err != nil {
				c.logger.ErrorCtx(ctx, "Error saving metrics to file"+err.Error())
			}

		case <-ctx.Done():
			c.logger.InfoCtx(ctx, "Backup process stopped.")
			return
		}
	}
}
