package main

import (
	"github.com/sanek1/metrics-collector/cmd/staticlint/exitchecker"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(exitchecker.Analyzer)
}
