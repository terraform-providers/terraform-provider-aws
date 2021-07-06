package tfresource_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestRetryWhenAwsErrCodeEquals(t *testing.T) {
	var retryCount int32

	testCases := []struct {
		Name        string
		F           func() (interface{}, error)
		ExpectError bool
	}{
		{
			Name: "no error",
			F: func() (interface{}, error) {
				return nil, nil
			},
		},
		{
			Name: "non-retryable other error",
			F: func() (interface{}, error) {
				return nil, errors.New("TestCode")
			},
			ExpectError: true,
		},
		{
			Name: "non-retryable AWS error",
			F: func() (interface{}, error) {
				return nil, awserr.New("Testing", "Testing", nil)
			},
			ExpectError: true,
		},
		{
			Name: "retryable AWS error timeout",
			F: func() (interface{}, error) {
				return nil, awserr.New("TestCode1", "TestMessage", nil)
			},
			ExpectError: true,
		},
		{
			Name: "retryable AWS error success",
			F: func() (interface{}, error) {
				if atomic.CompareAndSwapInt32(&retryCount, 0, 1) {
					return nil, awserr.New("TestCode2", "TestMessage", nil)
				}

				return nil, nil
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			retryCount = 0

			_, err := tfresource.RetryWhenAwsErrCodeEquals(5*time.Second, testCase.F, "TestCode1", "TestCode2")

			if testCase.ExpectError && err == nil {
				t.Fatal("expected error")
			} else if !testCase.ExpectError && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestRetryConfigContext_error(t *testing.T) {
	t.Parallel()

	expected := fmt.Errorf("nope")
	f := func() *resource.RetryError {
		return resource.NonRetryableError(expected)
	}

	errCh := make(chan error)
	go func() {
		errCh <- tfresource.RetryConfigContext(context.Background(), 0*time.Second, 0*time.Second, 0*time.Second, 0*time.Second, 1*time.Second, f)
	}()

	select {
	case err := <-errCh:
		if err != expected {
			t.Fatalf("bad: %#v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}
