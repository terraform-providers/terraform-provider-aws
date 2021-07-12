package AWSAT007_test

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/passes/AWSAT007"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT007(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT007.Analyzer, "a")
}
