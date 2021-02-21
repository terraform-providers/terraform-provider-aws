package naming

import (
	"regexp"
	"testing"
)

func strPtr(str string) *string {
	return &str
}

func TestGenerate(t *testing.T) {
	testCases := []struct {
		TestName              string
		Name                  string
		NamePrefix            string
		ExpectedRegexpPattern string
	}{
		{
			TestName:              "name",
			Name:                  "test",
			NamePrefix:            "",
			ExpectedRegexpPattern: "^test$",
		},
		{
			TestName:              "name prefix",
			Name:                  "",
			NamePrefix:            "test",
			ExpectedRegexpPattern: resourcePrefixedUniqueIDRegexpPattern("test"),
		},
		{
			TestName:              "fully generated",
			Name:                  "",
			NamePrefix:            "",
			ExpectedRegexpPattern: resourceUniqueIDRegexpPattern,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got := Generate(testCase.Name, testCase.NamePrefix)

			expectedRegexp, err := regexp.Compile(testCase.ExpectedRegexpPattern)

			if err != nil {
				t.Errorf("unable to compile regular expression pattern %s: %s", testCase.ExpectedRegexpPattern, err)
			}

			if !expectedRegexp.MatchString(got) {
				t.Errorf("got %s, expected to match regular expression pattern %s", got, testCase.ExpectedRegexpPattern)
			}
		})
	}
}

func TestHasResourceUniqueIdSuffix(t *testing.T) {
	testCases := []struct {
		TestName string
		Input    string
		Expected bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: false,
		},
		{
			TestName: "incorrect suffix",
			Input:    "test-123",
			Expected: false,
		},
		{
			TestName: "correct suffix with numbers",
			Input:    "test-20060102150405000000000001",
			Expected: true,
		},
		{
			TestName: "correct suffix with hex",
			Input:    "test-200601021504050000000000a1",
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got := HasResourceUniqueIdSuffix(testCase.Input)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestNamePrefixFromName(t *testing.T) {
	testCases := []struct {
		TestName string
		Input    string
		Expected *string
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: nil,
		},
		{
			TestName: "incorrect suffix",
			Input:    "test-123",
			Expected: nil,
		},
		{
			TestName: "prefix without hyphen, correct suffix",
			Input:    "test20060102150405000000000001",
			Expected: strPtr("test"),
		},
		{
			TestName: "prefix with hyphen, correct suffix",
			Input:    "test-20060102150405000000000001",
			Expected: strPtr("test-"),
		},
		{
			TestName: "prefix with hyphen, correct suffix with hex",
			Input:    "test-200601021504050000000000f1",
			Expected: strPtr("test-"),
		},
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17017
		{
			TestName: "terraform prefix, correct suffix",
			Input:    "terraform-20060102150405000000000001",
			Expected: strPtr("terraform-"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			expected := testCase.Expected
			got := NamePrefixFromName(testCase.Input)

			if expected == nil && got != nil {
				t.Errorf("got %s, expected nil", *got)
			}

			if expected != nil && got == nil {
				t.Errorf("got nil, expected %s", *expected)
			}

			if expected != nil && got != nil && *expected != *got {
				t.Errorf("got %s, expected %s", *got, *expected)
			}
		})
	}

	t.Run("extracting prefix from generated name", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			prefix := "test-"
			input := Generate("", prefix)
			got := NamePrefixFromName(input)

			if got == nil {
				t.Errorf("run%d: got nil, expected %s for input %s", i, prefix, input)
			}

			if got != nil && prefix != *got {
				t.Errorf("run%d: got %s, expected %s for input %s", i, *got, prefix, input)
			}
		}
	})
}
