package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	UserPoolDomainDeleteTimeout = 1 * time.Minute
)

// UserPoolDomainDeleted waits for an Operation to return Success
func UserPoolDomainDeleted(conn *cognitoidentityprovider.CognitoIdentityProvider, domain string) (*cognitoidentityprovider.DescribeUserPoolDomainOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cognitoidentityprovider.DomainStatusTypeUpdating,
			cognitoidentityprovider.DomainStatusTypeDeleting,
		},
		Target:  []string{""},
		Refresh: UserPoolDomainStatus(conn, domain),
		Timeout: UserPoolDomainDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cognitoidentityprovider.DescribeUserPoolDomainOutput); ok {
		return output, err
	}

	return nil, err
}

func UserPoolDomainCreated(conn *cognitoidentityprovider.CognitoIdentityProvider, domain string, timeout time.Duration) (*cognitoidentityprovider.DescribeUserPoolDomainOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cognitoidentityprovider.DomainStatusTypeCreating,
			cognitoidentityprovider.DomainStatusTypeUpdating,
		},
		Target: []string{
			cognitoidentityprovider.DomainStatusTypeActive,
		},
		Refresh: UserPoolDomainStatus(conn, domain),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cognitoidentityprovider.DescribeUserPoolDomainOutput); ok {
		return output, err
	}

	return nil, err
}
