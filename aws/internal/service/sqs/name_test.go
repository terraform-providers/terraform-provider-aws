package sqs

import (
	"testing"
)

func TestQueueNameFromURL(t *testing.T) {
	testCases := []struct {
		Name              string
		URL               string
		ExpectedQueueName string
		ExpectError       bool
	}{
		{
			Name:        "empty URL",
			ExpectError: true,
		},
		{
			Name:        "invalid URL",
			URL:         "---",
			ExpectError: true,
		},
		{
			Name:        "too few path parts",
			URL:         "http://sqs.us-west-2.amazonaws.com",
			ExpectError: true,
		},
		{
			Name:        "too many path parts",
			URL:         "http://sqs.us-west-2.amazonaws.com/123456789012/queueName/extra",
			ExpectError: true,
		},
		{
			Name:              "valid URL",
			URL:               "http://sqs.us-west-2.amazonaws.com/123456789012/queueName",
			ExpectedQueueName: "queueName",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got, err := QueueNameFromURL(testCase.URL)

			if err != nil && !testCase.ExpectError {
				t.Errorf("got unexpected error: %s", err)
			}

			if err == nil && testCase.ExpectError {
				t.Errorf("expected error, but received none")
			}

			if got != testCase.ExpectedQueueName {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedQueueName)
			}
		})
	}
}
