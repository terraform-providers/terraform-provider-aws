package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type WafRetryer struct {
	Connection *waf.WAF
}

type withTokenFunc func(token *string) (interface{}, error)

func (t *WafRetryer) RetryWithToken(f withTokenFunc) (interface{}, error) {
	awsMutexKV.Lock("WafRetryer")
	defer awsMutexKV.Unlock("WafRetryer")

	var out interface{}
	var tokenOut *waf.GetChangeTokenOutput
	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		var err error
		tokenOut, err = t.Connection.GetChangeToken(&waf.GetChangeTokenInput{})
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Failed to acquire change token: %s", err))
		}

		out, err = f(tokenOut.ChangeToken)
		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok && awsErr.Code() == "WAFStaleDataException" {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		tokenOut, err = t.Connection.GetChangeToken(&waf.GetChangeTokenInput{})

		if err != nil {
			return nil, fmt.Errorf("error getting WAF change token: %s", err)
		}

		out, err = f(tokenOut.ChangeToken)
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func newWafRetryer(conn *waf.WAF) *WafRetryer {
	return &WafRetryer{Connection: conn}
}
