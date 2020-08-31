package main

import (
	"github.com/komeme/myanalyzer"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(myanalyzer.Analyzer) }
