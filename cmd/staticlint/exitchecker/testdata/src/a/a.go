// Package main содержит тестовые примеры для анализатора exitchecker
package main

import (
	"os"
)

func main() {
	os.Exit(0)
}

func otherFunc() {
	os.Exit(1)
}
