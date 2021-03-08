package tfresource

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestEmptyResultErrorAsNotFoundError(t *testing.T) {
	lastRequest := 123
	err := NewEmptyResultError(lastRequest)

	var nfe *resource.NotFoundError
	ok := errors.As(err, &nfe)

	if !ok {
		t.Fatal("expected errors.As() to return true")
	}
	if nfe.Message != "empty result" {
		t.Errorf(`expected Message to be "empty result", got %q`, nfe.Message)
	}
	if nfe.LastRequest != lastRequest {
		t.Errorf("unexpected value for LastRequest")
	}
}

func TestEmptyResultErrorIs(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "compare to nil",
			err:  nil,
		},
		{
			name: "other error",
			err:  errors.New("test"),
		},
		{
			name: "EmptyResultError with LastRequest",
			err: &EmptyResultError{
				LastRequest: 123,
			},
			expected: true,
		},
		{
			name:     "ErrEmptyResult",
			err:      ErrEmptyResult,
			expected: true,
		},
		{
			name: "wrapped other error",
			err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			name: "wrapped EmptyResultError with LastRequest",
			err: fmt.Errorf("test: %w", &EmptyResultError{
				LastRequest: 123,
			}),
			expected: true,
		},
		{
			name:     "wrapped ErrEmptyResult",
			err:      fmt.Errorf("test: %w", ErrEmptyResult),
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := &EmptyResultError{}
			ok := errors.Is(testCase.err, err)
			if ok != testCase.expected {
				t.Errorf("got %t, expected %t", ok, testCase.expected)
			}
		})
	}
}

func TestTooManyResultsErrorAsNotFoundError(t *testing.T) {
	count := 2
	lastRequest := 123
	err := NewTooManyResultsError(count, lastRequest)

	var nfe *resource.NotFoundError
	ok := errors.As(err, &nfe)

	if !ok {
		t.Fatal("expected errors.As() to return true")
	}
	if expected := fmt.Sprintf("too many results: wanted 1, got %d", count); nfe.Message != expected {
		t.Errorf(`expected Message to be %q, got %q`, expected, nfe.Message)
	}
	if nfe.LastRequest != lastRequest {
		t.Errorf("unexpected value for LastRequest")
	}
}

func TestTooManyResultsErrorIs(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "compare to nil",
			err:  nil,
		},
		{
			name: "other error",
			err:  errors.New("test"),
		},
		{
			name: "TooManyResultsError with LastRequest",
			err: &TooManyResultsError{
				LastRequest: 123,
			},
			expected: true,
		},
		{
			name:     "ErrTooManyResults",
			err:      ErrTooManyResults,
			expected: true,
		},
		{
			name: "wrapped other error",
			err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			name: "wrapped TooManyResultsError with LastRequest",
			err: fmt.Errorf("test: %w", &TooManyResultsError{
				LastRequest: 123,
			}),
			expected: true,
		},
		{
			name:     "wrapped ErrTooManyResults",
			err:      fmt.Errorf("test: %w", ErrTooManyResults),
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := &TooManyResultsError{}
			ok := errors.Is(testCase.err, err)
			if ok != testCase.expected {
				t.Errorf("got %t, expected %t", ok, testCase.expected)
			}
		})
	}
}
