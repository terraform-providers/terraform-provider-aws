package apprunner_test

import (
	"fmt"
	"testing"

	tfapprunner "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apprunner"
)

func TestCustomDomainAssociationParseID(t *testing.T) {
	testCases := []struct {
		TestName      string
		InputID       string
		ExpectedError bool
		ExpectedPart0 string
		ExpectedPart1 string
	}{
		{
			TestName:      "empty ID",
			InputID:       "",
			ExpectedError: true,
		},
		{
			TestName:      "single part",
			InputID:       "example.com",
			ExpectedError: true,
		},
		{
			TestName:      "two parts",
			InputID:       fmt.Sprintf("%s,%s", "example.com", "arn:aws:apprunner:us-east-1:1234567890:service/example/0a03292a89764e5882c41d8f991c82fe"), //lintignore:AWSAT005
			ExpectedPart0: "example.com",
			ExpectedPart1: "arn:aws:apprunner:us-east-1:1234567890:service/example/0a03292a89764e5882c41d8f991c82fe", //lintignore:AWSAT005
		},

		{
			TestName:      "empty both parts",
			InputID:       ",",
			ExpectedError: true,
		},
		{
			TestName:      "empty first part",
			InputID:       ",arn:aws:apprunner:us-east-1:1234567890:service/example/0a03292a89764e5882c41d8f991c82fe", //lintignore:AWSAT005
			ExpectedError: true,
		},
		{
			TestName:      "empty second part",
			InputID:       "example.com,",
			ExpectedError: true,
		},
		{
			TestName:      "three parts",
			InputID:       "example.com,arn:aws:apprunner:us-east-1:1234567890:service/example/0a03292a89764e5882c41d8f991c82fe,example", //lintignore:AWSAT005
			ExpectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotPart0, gotPart1, err := tfapprunner.CustomDomainAssociationParseID(testCase.InputID)

			if err == nil && testCase.ExpectedError {
				t.Fatalf("expected error, got no error")
			}

			if err != nil && !testCase.ExpectedError {
				t.Fatalf("got unexpected error: %s", err)
			}

			if gotPart0 != testCase.ExpectedPart0 {
				t.Errorf("got part 0 %s, expected %s", gotPart0, testCase.ExpectedPart0)
			}

			if gotPart1 != testCase.ExpectedPart1 {
				t.Errorf("got part 1 %s, expected %s", gotPart1, testCase.ExpectedPart1)
			}
		})
	}
}
