package AWSAT001_test

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/passes/AWSAT001"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT001(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT001.Analyzer, "a")
}

func TestAttributeNameAppearsArnRelated(t *testing.T) {
	testCases := []struct {
		Name          string
		AttributeName string
		Expected      bool
	}{
		{
			Name:          "empty",
			AttributeName: "",
			Expected:      false,
		},
		{
			Name:          "not arn",
			AttributeName: "test",
			Expected:      false,
		},
		{
			Name:          "equals arn",
			AttributeName: "arn",
			Expected:      true,
		},
		{
			Name:          "arn suffix",
			AttributeName: "some_arn",
			Expected:      true,
		},
		{
			Name:          "nested attribute equals arn",
			AttributeName: "config_block.0.arn",
			Expected:      true,
		},
		{
			Name:          "nested attribute arn suffix",
			AttributeName: "config_block.0.some_arn",
			Expected:      true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Name, func(t *testing.T) {
			got := AWSAT001.AttributeNameAppearsArnRelated(testCase.AttributeName)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
