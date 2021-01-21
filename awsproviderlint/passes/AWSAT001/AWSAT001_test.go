package AWSAT001_test

import (
	"testing"

	_ "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
			Name:          "equals kms_key_id",
			AttributeName: "kms_key_id",
			Expected:      true,
		},
		{
			Name:          "arn suffix",
			AttributeName: "some_arn",
			Expected:      true,
		},
		{
			Name:          "kms_key_id suffix",
			AttributeName: "some_kms_key_id",
			Expected:      true,
		},
		{
			Name:          "nested attribute equals arn",
			AttributeName: "config_block.0.arn",
			Expected:      true,
		},
		{
			Name:          "nested attribute equals kms_key_id",
			AttributeName: "config_block.0.kms_key_id",
			Expected:      true,
		},
		{
			Name:          "nested attribute arn suffix",
			AttributeName: "config_block.0.some_arn",
			Expected:      true,
		},
		{
			Name:          "nested attribute kms_key_id suffix",
			AttributeName: "config_block.0.some_kms_key_id",
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
