package aws

import (
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// Returns true if the error matches all these conditions:
//  * err is of type awserr.Error
//  * Error.Code() matches code
//  * Error.Message() contains message
func isAWSErr(err error, code string, message string) bool {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		return awsErr.Code() == code && strings.Contains(awsErr.Message(), message)
	}
	return false
}

// Returns true if the error matches all these conditions:
//  * err is of type awserr.RequestFailure
//  * RequestFailure.StatusCode() matches status code
// It is always preferable to use isAWSErr() except in older APIs (e.g. S3)
// that sometimes only respond with status codes.
func isAWSErrRequestFailureStatusCode(err error, statusCode int) bool {
	var awsErr awserr.RequestFailure
	if errors.As(err, &awsErr) {
		return awsErr.StatusCode() == statusCode
	}
	return false
}

func retryOnAwsCode(code string, f func() (interface{}, error)) (interface{}, error) {
	var resp interface{}
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok && awsErr.Code() == code {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = f()
	}

	return resp, err
}

// RetryOnAwsCodes retries AWS error codes for one minute
// Note: This function will be moved out of the aws package in the future.
func RetryOnAwsCodes(codes []string, f func() (interface{}, error)) (interface{}, error) {
	var resp interface{}
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok {
				for _, code := range codes {
					if awsErr.Code() == code {
						return resource.RetryableError(err)
					}
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = f()
	}

	return resp, err
}
