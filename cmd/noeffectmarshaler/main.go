package main

import (
	"github.com/komeme/noeffectmarshaler"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(noeffectmarshaler.Analyzer) }
