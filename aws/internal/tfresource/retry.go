package tfresource

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// RetryWhenAwsErrCodeEquals retries the specified function when it returns one of the specified AWS error code.
func RetryWhenAwsErrCodeEquals(timeout time.Duration, f func() (interface{}, error), codes ...string) (interface{}, error) {
	var output interface{}

	err := resource.Retry(timeout, func() *resource.RetryError {
		var err error

		output, err = f()

		for _, code := range codes {
			if tfawserr.ErrCodeEquals(err, code) {
				return resource.RetryableError(err)
			}
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if TimedOut(err) {
		output, err = f()
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// RetryConfigContext allows configuration of StateChangeConf's various time arguments.
// This is especially useful for AWS services that are prone to throttling, such as Route53, where
// the default durations cause problems. To not use a StateChangeConf argument and revert to the
// default, pass in a zero value (i.e., 0*time.Second).
func RetryConfigContext(ctx context.Context, delay time.Duration, delayRand time.Duration, minTimeout time.Duration, pollInterval time.Duration, timeout time.Duration, f resource.RetryFunc) error {
	// These are used to pull the error out of the function; need a mutex to
	// avoid a data race.
	var resultErr error
	var resultErrMu sync.Mutex

	c := &resource.StateChangeConf{
		Pending: []string{"retryableerror"},
		Target:  []string{"success"},
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			rerr := f()

			resultErrMu.Lock()
			defer resultErrMu.Unlock()

			if rerr == nil {
				resultErr = nil
				return 42, "success", nil
			}

			resultErr = rerr.Err

			if rerr.Retryable {
				return 42, "retryableerror", nil
			}

			return nil, "quit", rerr.Err
		},
	}

	if delay.Milliseconds() > 0 {
		c.Delay = delay
	}

	if delayRand.Milliseconds() > 0 {
		// Hitting the API at exactly the same time on each iteration of the retry is more likely to
		// cause Throttling problems. We introduce randomness in order to help AWS be happier.
		rand.Seed(time.Now().UTC().UnixNano())

		c.Delay = time.Duration(rand.Int63n(delayRand.Milliseconds())) * time.Millisecond
	}

	if minTimeout.Milliseconds() > 0 {
		c.MinTimeout = minTimeout
	}

	if pollInterval.Milliseconds() > 0 {
		c.PollInterval = pollInterval
	}

	_, waitErr := c.WaitForStateContext(ctx)

	// Need to acquire the lock here to be able to avoid race using resultErr as
	// the return value
	resultErrMu.Lock()
	defer resultErrMu.Unlock()

	// resultErr may be nil because the wait timed out and resultErr was never
	// set; this is still an error
	if resultErr == nil {
		return waitErr
	}
	// resultErr takes precedence over waitErr if both are set because it is
	// more likely to be useful
	return resultErr
}
