package main

import (
	"flag"
	"fmt"
	"os"
)

// неэкспортированная переменная flagRunAddr содержит адрес и порт для запуска сервера

var Options struct {
	flagRunAddr string
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&Options.flagRunAddr, "a", ":8080", "address and port to run server")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "Unknown flags:", flag.Args())
		os.Exit(1)
	}

	// override flags with env vars
	if addr := os.Getenv("ADDRESS"); addr != "" {
		Options.flagRunAddr = addr
	}
}
