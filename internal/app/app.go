package app

import (
	"net/http"
	"time"

	"github.com/sanek1/metrics-collector/internal/controller"
	v "github.com/sanek1/metrics-collector/internal/validation"
	"go.uber.org/zap"
)

type App struct {
	controller *controller.Controller
	addr       string
}

func New(addr string) *App {
	ctrl := controller.New()

	return &App{
		controller: ctrl,
		addr:       addr,
	}
}

func (a *App) Run() error {
	// init zap logger
	if _, err := v.Initialize("test_level"); err != nil {
		return err
	}
	server := &http.Server{
		Addr:              a.addr,
		Handler:           a.controller.Router(),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	v.Loger.Info("Running server", zap.String("address", a.addr))
	return server.ListenAndServe()
}
