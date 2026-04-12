package main

import (
	"multichecker/internal/staticlint"
	"strings"

	gocritic "github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/kkHAIKE/contextcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {

	analyzers := []*analysis.Analyzer{
		// Package printf defines an Analyzer that checks consistency of Printf format strings and arguments
		printf.Analyzer,
		// Package shadow defines an Analyzer that checks for shadowed variables
		shadow.Analyzer,
		// Package structtag defines an Analyzer that checks struct field tags are well formed
		structtag.Analyzer,
	}

	for _, a := range staticcheck.Analyzers {
		// Add all analyzers for SA-class
		if strings.HasPrefix(a.Analyzer.Name, "SA") {
			analyzers = append(analyzers, a.Analyzer)
		}
		switch a.Analyzer.Name {
		// Use plain channel send or receive instead of single-case select
		case "S1000":
			analyzers = append(analyzers, a.Analyzer)
		// Incorrect or missing package comment
		case "ST1000":
			analyzers = append(analyzers, a.Analyzer)
		// Convert slice of bytes to string when printing it
		case "QF1010":
			analyzers = append(analyzers, a.Analyzer)
		}
	}

	// Add all from Go-critic
	analyzers = append(analyzers, gocritic.Analyzer)
	// Add all from contextcheck
	analyzers = append(analyzers, contextcheck.NewAnalyzer(contextcheck.Configuration{}))
	// Add our own analyzer
	analyzers = append(analyzers, staticlint.Analyzer)

	multichecker.Main(
		analyzers...,
	)
}
