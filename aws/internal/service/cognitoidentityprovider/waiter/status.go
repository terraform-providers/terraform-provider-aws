package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	userPoolDomainStatusNotFound = "NotFound"
	userPoolDomainStatusUnknown  = "Unknown"
)

// UserPoolDomainStatus fetches the Operation and its Status
func UserPoolDomainStatus(conn *cognitoidentityprovider.CognitoIdentityProvider, domain string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &cognitoidentityprovider.DescribeUserPoolDomainInput{
			Domain: aws.String(domain),
		}

		output, err := conn.DescribeUserPoolDomain(input)

		if err != nil {
			return nil, userPoolDomainStatusUnknown, err
		}

		if output == nil {
			return nil, userPoolDomainStatusNotFound, nil
		}

		return output, aws.StringValue(output.DomainDescription.Status), nil
	}
}
