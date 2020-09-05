package noeffectmarshaler_test

import (
	"testing"

	"github.com/komeme/noeffectmarshaler"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, noeffectmarshaler.Analyzer, "a")
}
