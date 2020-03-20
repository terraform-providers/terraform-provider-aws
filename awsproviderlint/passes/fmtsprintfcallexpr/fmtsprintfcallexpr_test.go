package fmtsprintfcallexpr

import (
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestValidateAnalyzer(t *testing.T) {
	err := analysis.Validate([]*analysis.Analyzer{Analyzer})

	if err != nil {
		t.Fatal(err)
	}
}
