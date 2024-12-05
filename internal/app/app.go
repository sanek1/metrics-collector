package app

import (
	"log"
	"net/http"

	"github.com/sanek1/metrics-collector/internal/controller"
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
	log.Println("Server start on", a.addr)
	return http.ListenAndServe(a.addr, a.controller.Router())
}
