// Package exitchecker содержит анализатор, который запрещает использование прямого вызова os.Exit
// в функции main пакета main.
package exitchecker

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "exitchecker",
	Doc:  "check for direct os.Exit calls in main function of main package",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		mainFunc, ok := n.(*ast.FuncDecl)
		if !ok || mainFunc.Name.Name != "main" {
			return
		}

		ast.Inspect(mainFunc.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			pkg, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			if pkg.Name == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(), "os.Exit not allowed in main")
			}

			return true
		})
	})

	return nil, nil
}
