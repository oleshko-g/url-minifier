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
		var inMainFunc bool
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.FuncDecl:
				if node.Name.Name == "main" {
					inMainFunc = true
				}
			case *ast.SelectorExpr:
				if !inMainFunc {
					if diag := checkBlockStmt(inMainFunc, node); diag != nil {
						diags = append(diags, *diag)
					}
				}
			case *ast.CallExpr:
				if diag := checkCallExpr(node); diag != nil {
					diags = append(diags, *diag)
				}
			}
			return true
		})
	}

	return diags
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

func checkBlockStmt(inMainFunc bool, expr *ast.SelectorExpr) *analysis.Diagnostic {
	if x, ok := expr.X.(*ast.Ident); ok {
		if !inMainFunc && isOSExit(x.Name, expr.Sel.Name) {
			return &analysis.Diagnostic{
				Pos:     expr.Pos(),
				Message: "calls os.Exit outside of main func",
			}
		}

		if !inMainFunc && isLogFatal(x.Name, expr.Sel.Name) {
			return &analysis.Diagnostic{
				Pos:     expr.Pos(),
				Message: "calls log.Fatal outside of main func",
			}
		}
	}
	return nil
}

func isOSExit(xName, selName string) bool {
	if xName == "os" && selName == "Exit" {
		return true
	}
	return false
}

func isLogFatal(xName, selName string) bool {
	if xName == "log" && selName == "Fatal" {
		return true
	}

	return false
}

// targetName, targetSelName,
