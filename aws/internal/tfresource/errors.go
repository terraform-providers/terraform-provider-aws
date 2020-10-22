package tfresource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// NotFound returns true if the error represents a "resource not found" condition.
// Specifically, NotFound returns true if the error is of type resource.NotFoundError.
func NotFound(err error) bool {
	_, ok := err.(*resource.NotFoundError)
	return ok
}

// TimedOut returns true if the error represents a "wait timed out" condition.
// Specifically, TimedOut returns true if the error matches all these conditions:
//  * err is of type resource.TimeoutError
//  * TimeoutError.LastError is nil
func TimedOut(err error) bool {
	timeoutErr, ok := err.(*resource.TimeoutError)
	return ok && timeoutErr.LastError == nil
}
