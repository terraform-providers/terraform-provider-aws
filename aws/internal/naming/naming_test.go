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

func TestHasResourceUniqueIdPrefix(t *testing.T) {
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
			TestName: "incorrect prefix",
			Input:    "test-20060102150405000000000001",
			Expected: false,
		},
		{
			TestName: "correct prefix",
			Input:    "terraform-20060102150405000000000001",
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got := HasResourceUniqueIdPrefix(testCase.Input)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
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
			TestName: "correct suffix, incorrect prefix",
			Input:    "test-20060102150405000000000001",
			Expected: true,
		},
		{
			TestName: "correct suffix, correct prefix",
			Input:    "terraform-20060102150405000000000001",
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
			TestName: "correct prefix, incorrect suffix",
			Input:    "test-123",
			Expected: nil,
		},
		{
			TestName: "correct prefix without hyphen, correct suffix",
			Input:    "test20060102150405000000000001",
			Expected: strPtr("test"),
		},
		{
			TestName: "correct prefix with hyphen, correct suffix",
			Input:    "test-20060102150405000000000001",
			Expected: strPtr("test-"),
		},
		{
			TestName: "incorrect prefix, correct suffix",
			Input:    "terraform-20060102150405000000000001",
			Expected: nil,
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
}
