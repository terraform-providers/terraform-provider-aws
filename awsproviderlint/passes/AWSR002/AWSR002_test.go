package AWSR002

import (
	"testing"

	"golang.org/x/tools/go/analysis"
)

// analysistest testing with the actual keyvaluetags internal package requires
// self-referencing an internal package. Vendoring via symlinks would need to
// be manually constructed and error prone. Using Go Modules to assemble the
// testdata vendor directory would re-vendor thousands of source code files.
// func TestAnalyzer(t *testing.T) {
// 	testdata := analysistest.TestData()
// 	analysistest.Run(t, testdata, Analyzer, "a")
// }

func TestValidate(t *testing.T) {
	err := analysis.Validate([]*analysis.Analyzer{Analyzer})

	if err != nil {
		t.Fatal(err)
	}
}
