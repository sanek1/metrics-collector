package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	app "github.com/sanek1/metrics-collector/internal/app/server"
	flags "github.com/sanek1/metrics-collector/internal/flags/server"
)

func main() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
	// Открываем файл для записи профиля
	//f, err := os.Create("../../profiles/cpu.pprof")
	//if err != nil {
	//	panic(err)
	//}
	//defer f.Close()

	// Начать профилирование CPU
	//if err := pprof.StartCPUProfile(f); err != nil {
	//	panic(err)
	//}
	//defer pprof.StopCPUProfile()

	opt := flags.ParseServerFlags()
	if err := app.New(opt, opt.UseDatabase).Run(); err != nil {
		log.Fatal(err)
	}

}
