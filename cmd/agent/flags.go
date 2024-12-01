package main

import (
	"flag"
	"fmt"
	"os"
)

// неэкспортированная переменная flagRunAddr содержит адрес и порт для запуска сервера

var Options struct {
	flagRunAddr    string
	reportInterval int64
	pollInterval   int64
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() {
	// регистрируем переменные
	flag.StringVar(&Options.flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Int64Var(&Options.reportInterval, "r", 10, "report interval in seconds")
	flag.Int64Var(&Options.pollInterval, "p", 2, "poll interval in seconds")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "Unknown flags:", flag.Args())
		os.Exit(1)
	}
}
