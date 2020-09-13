package AWSV001

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSV001(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}
