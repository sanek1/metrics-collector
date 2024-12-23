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

	storeInterval int64
	path          string
	restore       bool
}

func New(addr string, storeInterval int64, path string, restore bool) *App {
	ctrl := controller.New()

	return &App{
		controller: ctrl,
		addr:       addr,

		storeInterval: storeInterval,
		path:          path,
		restore:       restore,
	}
}

func (a *App) Run() error {
	// init zap logger
	loger, err := v.Initialize("info")
	if err != nil {
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
	loger.Logger.Info("Running server ", zap.String("address", a.addr))
	go a.controller.PeriodicallySaveBackUp(a.path, a.restore, time.Duration(a.storeInterval)*time.Second)
	return server.ListenAndServe()
}
