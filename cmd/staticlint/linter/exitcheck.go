package linter

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os.Exit",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	searchFunc := func(fn *ast.FuncDecl) {
		if fn.Name.Name != "main" {
			return
		}

		for _, stmt := range fn.Body.List {
			switch stmt := stmt.(type) {
			case *ast.ExprStmt:
				ast.Inspect(stmt, func(node ast.Node) bool {
					switch s := node.(type) {
					case *ast.SelectorExpr:
						if ident, ok := s.X.(*ast.Ident); ok {
							if ident.Name == "os" && s.Sel.Name == "Exit" {
								pass.Reportf(s.Pos(), "os.Exit is not allowed")
							}
						}

					}
					return true
				})
			}
		}
		return
	}

	for _, f := range pass.Files {
		if f.Name.Name != "main" {
			continue
		}

		ast.Inspect(f, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				searchFunc(x)
			}
			return true
		})
	}
	return nil, nil
}
