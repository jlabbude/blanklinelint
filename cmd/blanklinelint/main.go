package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"
	"linelint/pkg/analyzer"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
