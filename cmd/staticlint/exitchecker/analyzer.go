// Package exitchecker implements a checker that detects direct calls to os.Exit from
// the main function in main packages. This function should be isolated to allow
// for better testability of the application.
package exitchecker

import (
	"go/ast"
	"path/filepath"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer entry point for the analyzer. Detects direct calls to os.Exit from
// the main function in main packages.
var Analyzer = &analysis.Analyzer{
	Name: "exitchecker",
	Doc:  "check for direct os.Exit calls in main",
	Run:  run,
}

// Регулярное выражение для определения файлов из кэша сборки
var goBuildCachePattern = regexp.MustCompile(`(go-build|AppData[\\/]Local[\\/]go-build|[\\/]tmp[\\/]go-build)`)

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := inspector.New(pass.Files)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		pos := pass.Fset.Position(n.Pos())
		filename := filepath.Clean(pos.Filename)

		if isInBuildCache(filename) {
			return
		}

		if pass.Pkg.Name() != "main" {
			return
		}

		fd := n.(*ast.FuncDecl)
		if fd.Name.Name != "main" {
			return
		}

		ast.Inspect(fd.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok || ident.Name != "os" || sel.Sel.Name != "Exit" {
				return true
			}

			pass.Reportf(call.Pos(), "os.Exit not allowed in main")
			return true
		})
	})

	return nil, nil
}

func isInBuildCache(filename string) bool {
	normalizedPath := filepath.ToSlash(filename)
	return goBuildCachePattern.MatchString(normalizedPath)
}
