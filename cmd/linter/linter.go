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
	diags := inspect(pass.Files)
	for _, diag := range diags {
		pass.Report(diag)
	}
	return nil, nil
}

func inspect(files []*ast.File) []analysis.Diagnostic {
	var diags []analysis.Diagnostic
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.CallExpr:
				if diag := checkCallExpr(node); diag != nil {
					diags = append(diags, *diag)
				}
			}
			return true
		})
	}

	return nil
}

func checkCallExpr(call *ast.CallExpr) *analysis.Diagnostic {

	if fun, ok := call.Fun.(*ast.Ident); ok {
		if fun.Name == "panic" {
			return &analysis.Diagnostic{
				Pos:     fun.Pos(),
				Message: "panics",
			}
		}
	}

	return nil
}

func checkBlockStmt(declBody *ast.BlockStmt) *analysis.Diagnostic {
	_ = declBody

	// TODO: if in FuncDecl.Name == "main" AND BlockStmt contains
	// * TODO: write if log.Fatal.Name == "log.Fatal" report the error
	// * TODO: write if log.Fatal.Name == "os.Exit" report the error
	return nil
}
