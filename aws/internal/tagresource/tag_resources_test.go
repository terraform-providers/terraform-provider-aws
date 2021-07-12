package tagresource

import (
	"testing"
)

func TestGetResourceId(t *testing.T) {
	testCases := []struct {
		Description        string
		ResourceIdentifier string
		ExpectedIdentifier string
		ExpectedKey        string
		ExpectedError      func(err error) bool
	}{
		{
			Description:        "empty resource identifier",
			ResourceIdentifier: "",
			ExpectedError: func(err error) bool {
				return err.Error() == "invalid resource identifier (), expected ID,KEY"
			},
		},
		{
			Description:        "missing identifier",
			ResourceIdentifier: ",testkey",
			ExpectedError: func(err error) bool {
				return err.Error() == "invalid resource identifier (,testkey), expected ID,KEY"
			},
		},
		{
			Description:        "missing key",
			ResourceIdentifier: "testidentifier,",
			ExpectedError: func(err error) bool {
				return err.Error() == "invalid resource identifier (testidentifier,), expected ID,KEY"
			},
		},
		{
			Description:        "incorrect separator",
			ResourceIdentifier: "testidentifier;testkey",
			ExpectedError: func(err error) bool {
				return err.Error() == "invalid resource identifier (testidentifier;testkey), expected ID,KEY"
			},
		},
		{
			Description:        "correct",
			ResourceIdentifier: "testidentifier,testkey",
			ExpectedIdentifier: "testidentifier",
			ExpectedKey:        "testkey",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Description, func(t *testing.T) {
			gotIdentifier, gotKey, err := GetResourceId(testCase.ResourceIdentifier)

			if err != nil && !testCase.ExpectedError(err) {
				t.Fatalf("unexpected error: %s", err)
			}

			if testCase.ExpectedError == nil && err != nil {
				t.Fatalf("expected no error, got error: %s", err)
			}

			if testCase.ExpectedError != nil && err == nil {
				t.Fatalf("expected error, got no error")
			}

			if gotIdentifier != testCase.ExpectedIdentifier {
				t.Errorf("got identifier %s, expected identifier %s", gotIdentifier, testCase.ExpectedIdentifier)
			}

			if gotKey != testCase.ExpectedKey {
				t.Errorf("got key %s, expected key %s", gotKey, testCase.ExpectedKey)
			}
		})
	}
}

func TestSetResourceId(t *testing.T) {
	testCases := []struct {
		Description                string
		Identifier                 string
		Key                        string
		ExpectedResourceIdentifier string
	}{
		{
			Description:                "correct",
			Identifier:                 "testidentifier",
			Key:                        "testkey",
			ExpectedResourceIdentifier: "testidentifier,testkey",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Description, func(t *testing.T) {
			got := SetResourceId(testCase.Identifier, testCase.Key)

			if got != testCase.ExpectedResourceIdentifier {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedResourceIdentifier)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	testCases := []struct {
		Input    string
		Expected string
	}{
		{
			Input:    "ARN",
			Expected: "arn",
		},
		{
			Input:    "LogGroupName",
			Expected: "log_group_name",
		},
		{
			Input:    "ResourceId",
			Expected: "resource_id",
		},
		{
			Input:    "ResourceArn",
			Expected: "resource_arn",
		},
		{
			Input:    "ResourceARN",
			Expected: "resource_arn",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Input, func(t *testing.T) {
			got := toSnakeCase(testCase.Input)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}
