package waiter

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ChangeTimeout = 30 * time.Minute

	HostedZoneDnssecStatusTimeout = 5 * time.Minute

	KeySigningKeyStatusTimeout = 5 * time.Minute
)

func ChangeInfoStatusInsync(conn *route53.Route53, changeID string) (*route53.ChangeInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{route53.ChangeStatusPending},
		Target:     []string{route53.ChangeStatusInsync},
		Refresh:    ChangeInfoStatus(conn, changeID),
		Delay:      30 * time.Second,
		MinTimeout: 5 * time.Second,
		Timeout:    ChangeTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53.ChangeInfo); ok {
		return output, err
	}

	return nil, err
}

func HostedZoneDnssecStatusUpdated(conn *route53.Route53, hostedZoneID string, status string) (*route53.DNSSECStatus, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{status},
		Refresh:    HostedZoneDnssecStatus(conn, hostedZoneID),
		MinTimeout: 5 * time.Second,
		Timeout:    HostedZoneDnssecStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53.DNSSECStatus); ok {
		if err != nil && output != nil && output.ServeSignature != nil && output.StatusMessage != nil {
			newErr := fmt.Errorf("%s: %s", aws.StringValue(output.ServeSignature), aws.StringValue(output.StatusMessage))

			switch e := err.(type) {
			case *resource.TimeoutError:
				if e.LastError == nil {
					e.LastError = newErr
				}
			case *resource.UnexpectedStateError:
				if e.LastError == nil {
					e.LastError = newErr
				}
			}
		}

		return output, err
	}

	return nil, err
}

func KeySigningKeyStatusUpdated(conn *route53.Route53, hostedZoneID string, name string, status string) (*route53.KeySigningKey, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{status},
		Refresh:    KeySigningKeyStatus(conn, hostedZoneID, name),
		MinTimeout: 5 * time.Second,
		Timeout:    KeySigningKeyStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53.KeySigningKey); ok {
		if err != nil && output != nil && output.Status != nil && output.StatusMessage != nil {
			newErr := fmt.Errorf("%s: %s", aws.StringValue(output.Status), aws.StringValue(output.StatusMessage))

			var te *resource.TimeoutError
			var use *resource.UnexpectedStateError
			if ok := errors.As(err, &te); ok && te.LastError == nil {
				te.LastError = newErr
			} else if ok := errors.As(err, &use); ok && use.LastError == nil {
				use.LastError = newErr
			}
		}

		return output, err
	}

	return nil, err
}
