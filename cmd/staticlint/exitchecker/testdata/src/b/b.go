// Package main содержит тестовые примеры для анализатора exitchecker
package main

import (
	"fmt"
	"os"
)

func main() {
	code := run()
	os.Exit(code) // want "os.Exit not allowed in main"
}

func run() int {
	fmt.Println("Hello, world!")
	return 0
}
