package cloudformation_test

import (
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation"
)

func TestTypeVersionARNToTypeARNAndVersionID(t *testing.T) {
	testCases := []struct {
		TestName          string
		InputARN          string
		ExpectedError     *regexp.Regexp
		ExpectedTypeARN   string
		ExpectedVersionID string
	}{
		{
			TestName:      "empty ARN",
			InputARN:      "",
			ExpectedError: regexp.MustCompile(`error parsing ARN`),
		},
		{
			TestName:      "unparsable ARN",
			InputARN:      "test",
			ExpectedError: regexp.MustCompile(`error parsing ARN`),
		},
		{
			TestName:      "invalid ARN service",
			InputARN:      "arn:aws:ec2:us-west-2:123456789012:instance/i-12345678",
			ExpectedError: regexp.MustCompile(`expected service cloudformation`),
		},
		{
			TestName:      "invalid ARN resource parts",
			InputARN:      "arn:aws:cloudformation:us-west-2:123456789012:type/resource/HashiCorp-TerraformAwsProvider-TfAccTestzwv6r2i7",
			ExpectedError: regexp.MustCompile(`expected 4 resource parts`),
		},
		{
			TestName:      "invalid ARN resource prefix",
			InputARN:      "arn:aws:cloudformation:us-west-2:123456789012:stack/name/1/2",
			ExpectedError: regexp.MustCompile(`expected resource prefix type`),
		},
		{
			TestName:          "valid ARN",
			InputARN:          "arn:aws:cloudformation:us-west-2:123456789012:type/resource/HashiCorp-TerraformAwsProvider-TfAccTestzwv6r2i7/00000001",
			ExpectedTypeARN:   "arn:aws:cloudformation:us-west-2:123456789012:type/resource/HashiCorp-TerraformAwsProvider-TfAccTestzwv6r2i7",
			ExpectedVersionID: "00000001",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotTypeARN, gotVersionID, err := cloudformation.TypeVersionARNToTypeARNAndVersionID(testCase.InputARN)

			if err == nil && testCase.ExpectedError != nil {
				t.Fatalf("expected error %s, got no error", testCase.ExpectedError.String())
			}

			if err != nil && testCase.ExpectedError == nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if err != nil && !testCase.ExpectedError.MatchString(err.Error()) {
				t.Fatalf("expected error %s, got: %s", testCase.ExpectedError.String(), err)
			}

			if gotTypeARN != testCase.ExpectedTypeARN {
				t.Errorf("got %s, expected %s", gotTypeARN, testCase.ExpectedTypeARN)
			}

			if gotVersionID != testCase.ExpectedVersionID {
				t.Errorf("got %s, expected %s", gotVersionID, testCase.ExpectedVersionID)
			}
		})
	}
}
