package tfresource

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestNotFound(t *testing.T) {
	testCases := []struct {
		Name     string
		Err      error
		Expected bool
	}{
		{
			Name: "nil error",
			Err:  nil,
		},
		{
			Name: "other error",
			Err:  errors.New("test"),
		},
		{
			Name:     "not found error",
			Err:      &resource.NotFoundError{LastError: errors.New("test")},
			Expected: true,
		},
		{
			Name: "wrapped other error",
			Err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			Name:     "wrapped not found error",
			Err:      fmt.Errorf("test: %w", &resource.NotFoundError{LastError: errors.New("test")}),
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := NotFound(testCase.Err)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestTimedOut(t *testing.T) {
	testCases := []struct {
		Name     string
		Err      error
		Expected bool
	}{
		{
			Name: "nil error",
			Err:  nil,
		},
		{
			Name: "other error",
			Err:  errors.New("test"),
		},
		{
			Name:     "timeout error",
			Err:      &resource.TimeoutError{},
			Expected: true,
		},
		{
			Name: "timeout error non-nil last error",
			Err:  &resource.TimeoutError{LastError: errors.New("test")},
		},
		{
			Name: "wrapped other error",
			Err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			Name: "wrapped timeout error",
			Err:  fmt.Errorf("test: %w", &resource.TimeoutError{}),
		},
		{
			Name: "wrapped timeout error non-nil last error",
			Err:  fmt.Errorf("test: %w", &resource.TimeoutError{LastError: errors.New("test")}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := TimedOut(testCase.Err)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
