package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/sanek1/metrics-collector/cmd/staticlint/exitchecker"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// main - точка входа для multichecker, который объединяет несколько статических анализаторов
// для проверки кода проекта.
//
// Этот multichecker включает в себя:
// 1. Стандартные статические анализаторы из пакета golang.org/x/tools/go/analysis/passes
// 2. Все анализаторы класса SA из пакета staticcheck.io
// 3. Некоторые анализаторы из других классов staticcheck.io
// 4. Публичные анализаторы: bodyclose и errcheck
// 5. Собственный анализатор exitchecker, запрещающий использование os.Exit в функции main пакета main
//
// Использование:
//
//	go run ./cmd/staticlint/... ./...   - проверка всего проекта
//	go run ./cmd/staticlint/... ./pkg/... - проверка только конкретных пакетов
//
//	поддеживаемые флаги
//	-explain - показать подробное объяснение для каждой проблемы
//	-fix - автоматически исправить некоторые проблемы, если возможно
//	-json - форматировать вывод в формате JSON
func main() {
	_ = os.Setenv("GODEBUG", "analysisnoverify=1")

	var exitcheckerOnly bool
	flag.BoolVar(&exitcheckerOnly, "exitchecker", false, "Run only exitchecker analyzer")
	flag.Parse()

	if exitcheckerOnly {
		fmt.Println("running only exitchecker...")
		mychecks := []*analysis.Analyzer{exitchecker.Analyzer}
		multichecker.Main(mychecks...)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "exitchecker" && len(os.Args) > 2 {
		fmt.Println("running only exitchecker (legacy mode)...")
		mychecks := []*analysis.Analyzer{exitchecker.Analyzer}
		os.Args = append(os.Args[:1], os.Args[2:]...)
		multichecker.Main(mychecks...)
		return
	}

	var mychecks []*analysis.Analyzer

	mychecks = append(mychecks,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
	)

	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1000" || v.Analyzer.Name == "ST1001" {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	mychecks = append(mychecks, bodyclose.Analyzer)
	mychecks = append(mychecks, errcheck.Analyzer)

	mychecks = append(mychecks, exitchecker.Analyzer)
	fmt.Println("analyzer exitchecker added:", exitchecker.Analyzer.Name)

	multichecker.Main(
		mychecks...,
	)
}
