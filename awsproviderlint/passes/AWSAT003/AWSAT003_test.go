package AWSAT003_test

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/passes/AWSAT003"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT003(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT003.Analyzer, "a")
}
