package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var PanicFatalExitAnalyzer *analysis.Analyzer = &analysis.Analyzer{
	Name: "PanicFatalExitAnalyzer",
	Doc:  "PanicFatalExit\n\nreports: [panic], [log.fatal] OR [os.exit]",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	inspect(pass.Files)
	return nil, nil
}

func inspect(files []*ast.File) {
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.CallExpr:
				check(node)
			}
			return true
		})

	}
}

func check(call *ast.CallExpr) *analysis.Diagnostic {
	// TODO: write if CallExpr.Name == "panic" report the error

	// TODO: if in FuncDecl.Name == "main" AND BlockStmt contains
	// * TODO: write if log.Fatal.Name == "log.Fatal" report the error
	// * TODO: write if log.Fatal.Name == "os.Exit" report the error
	// # TODO: write if log.Fatal.Name == "os.Exit" report the error

	if ok := true; !ok {
		return &analysis.Diagnostic{}
	}

	return nil
}
