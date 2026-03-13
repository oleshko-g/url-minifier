package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func Test_PanicFatalExitAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), PanicFatalExitAnalyzer, "./...")
}
