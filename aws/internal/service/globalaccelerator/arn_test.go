package globalaccelerator_test

import (
	"regexp"
	"testing"

	tfglobalaccelerator "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/globalaccelerator"
)

func TestEndpointGroupARNToListenerARN(t *testing.T) {
	testCases := []struct {
		TestName      string
		InputARN      string
		ExpectedError *regexp.Regexp
		ExpectedARN   string
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
			InputARN:      "arn:aws:ec2::123456789012:accelerator/a-123/listener/l-456/endpoint-group/eg-789",
			ExpectedError: regexp.MustCompile(`expected service globalaccelerator`),
		},
		{
			TestName:      "invalid ARN resource parts",
			InputARN:      "arn:aws:globalaccelerator::123456789012:accelerator/a-123/listener/l-456",
			ExpectedError: regexp.MustCompile(`expected at least 6 resource parts`),
		},
		{
			TestName:    "valid ARN",
			InputARN:    "arn:aws:globalaccelerator::123456789012:accelerator/a-123/listener/l-456/endpoint-group/eg-789",
			ExpectedARN: "arn:aws:globalaccelerator::123456789012:accelerator/a-123/listener/l-456",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got, err := tfglobalaccelerator.EndpointGroupARNToListenerARN(testCase.InputARN)

			if err == nil && testCase.ExpectedError != nil {
				t.Fatalf("expected error %s, got no error", testCase.ExpectedError.String())
			}

			if err != nil && testCase.ExpectedError == nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if err != nil && !testCase.ExpectedError.MatchString(err.Error()) {
				t.Fatalf("expected error %s, got: %s", testCase.ExpectedError.String(), err)
			}

			if got != testCase.ExpectedARN {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedARN)
			}
		})
	}
}

func TestListenerOrEndpointGroupARNToAcceleratorARN(t *testing.T) {
	testCases := []struct {
		TestName      string
		InputARN      string
		ExpectedError *regexp.Regexp
		ExpectedARN   string
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
			InputARN:      "arn:aws:ec2::123456789012:accelerator/a-123/listener/l-456",
			ExpectedError: regexp.MustCompile(`expected service globalaccelerator`),
		},
		{
			TestName:      "invalid ARN resource parts",
			InputARN:      "arn:aws:globalaccelerator::123456789012:accelerator/a-123",
			ExpectedError: regexp.MustCompile(`expected at least 4 resource parts`),
		},
		{
			TestName:    "valid listener ARN",
			InputARN:    "arn:aws:globalaccelerator::123456789012:accelerator/a-123/listener/l-456",
			ExpectedARN: "arn:aws:globalaccelerator::123456789012:accelerator/a-123",
		},
		{
			TestName:    "valid endpoint group ARN",
			InputARN:    "arn:aws:globalaccelerator::123456789012:accelerator/a-123/listener/l-456/endpoint-group/eg-789",
			ExpectedARN: "arn:aws:globalaccelerator::123456789012:accelerator/a-123",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got, err := tfglobalaccelerator.ListenerOrEndpointGroupARNToAcceleratorARN(testCase.InputARN)

			if err == nil && testCase.ExpectedError != nil {
				t.Fatalf("expected error %s, got no error", testCase.ExpectedError.String())
			}

			if err != nil && testCase.ExpectedError == nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if err != nil && !testCase.ExpectedError.MatchString(err.Error()) {
				t.Fatalf("expected error %s, got: %s", testCase.ExpectedError.String(), err)
			}

			if got != testCase.ExpectedARN {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedARN)
			}
		})
	}
}
