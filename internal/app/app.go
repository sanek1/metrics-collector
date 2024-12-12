package app

import (
	"net/http"

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
	if err := v.Initialize("test_level"); err != nil {
		return err
	}
	v.Loger.Info("Running server", zap.String("address", a.addr))
	return http.ListenAndServe(a.addr, a.controller.Router())
}
