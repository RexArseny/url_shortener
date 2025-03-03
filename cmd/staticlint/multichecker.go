package main

import (
	exitAnalyzer "multichecker/exit_analyzer"

	"github.com/Antonboom/nilnil/pkg/analyzer"
	"github.com/polyfloyd/go-errorlint/errorlint"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var mychecks []*analysis.Analyzer

	errorlintAnalyzer := errorlint.NewAnalyzer()
	errorlintAnalyzer.Flags.Set("errorf", "true")

	mychecks = append(
		mychecks,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		analyzer.New(),
		errorlintAnalyzer,
		exitAnalyzer.Analyzer)

	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	multichecker.Main(
		mychecks...,
	)
}
