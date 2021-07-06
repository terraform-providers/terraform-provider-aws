package tfresource_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
			got := tfresource.NotFound(testCase.Err)

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
			got := tfresource.TimedOut(testCase.Err)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestSetLastError(t *testing.T) {
	testCases := []struct {
		Name     string
		Err      error
		LastErr  error
		Expected bool
	}{
		{
			Name: "nil error",
		},
		{
			Name:    "other error",
			Err:     errors.New("test"),
			LastErr: errors.New("last"),
		},
		{
			Name: "timeout error lastErr is nil",
			Err:  &resource.TimeoutError{},
		},
		{
			Name:     "timeout error",
			Err:      &resource.TimeoutError{},
			LastErr:  errors.New("lasttest"),
			Expected: true,
		},
		{
			Name: "timeout error non-nil last error lastErr is nil",
			Err:  &resource.TimeoutError{LastError: errors.New("test")},
		},
		{
			Name:    "timeout error non-nil last error no overwrite",
			Err:     &resource.TimeoutError{LastError: errors.New("test")},
			LastErr: errors.New("lasttest"),
		},
		{
			Name: "unexpected state error lastErr is nil",
			Err:  &resource.UnexpectedStateError{},
		},
		{
			Name:     "unexpected state error",
			Err:      &resource.UnexpectedStateError{},
			LastErr:  errors.New("lasttest"),
			Expected: true,
		},
		{
			Name: "unexpected state error non-nil last error lastErr is nil",
			Err:  &resource.UnexpectedStateError{LastError: errors.New("test")},
		},
		{
			Name:    "unexpected state error non-nil last error no overwrite",
			Err:     &resource.UnexpectedStateError{LastError: errors.New("test")},
			LastErr: errors.New("lasttest"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tfresource.SetLastError(testCase.Err, testCase.LastErr)

			if testCase.Err != nil {
				got := testCase.Err.Error()
				contains := strings.Contains(got, "lasttest")

				if (testCase.Expected && !contains) || (!testCase.Expected && contains) {
					t.Errorf("got %s", got)
				}
			}
		})
	}
}
