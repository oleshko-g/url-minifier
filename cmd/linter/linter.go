package main

import (
	_ "go/ast"
	_ "go/token"

	"golang.org/x/tools/go/analysis"
)

var PanicFatalExitAnalyzer *analysis.Analyzer = &analysis.Analyzer{
	Name: "PanicFatalExitAnalyzer",
	Doc:  "PanicFatalExit\n\nreports: [panic], [log.fatal] OR [os.exit]",
	Run:  nil,
}

func run(*analysis.Pass) (any, error) {
	// TODO: feed packages
	// TODO: parse files into AST
	// TODO: traverse AST
	// TODO: write if CallExpr.Name == "panic" report the error
	// TODO: if in FuncDecl.Name == "main" AND BlockStmt contains
	// * TODO: write if log.Fatal.Name == "log.Fatal" report the error
	// * TODO: write if log.Fatal.Name == "os.Exit" report the error
	// # TODO: write if log.Fatal.Name == "os.Exit" report the error
	return nil, nil
}
