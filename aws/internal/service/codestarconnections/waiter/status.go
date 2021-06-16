package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// HostStatus fetches the Host and its Status
func HostStatus(conn *codestarconnections.CodeStarConnections, hostARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &codestarconnections.GetHostInput{
			HostArn: aws.String(hostARN),
		}

		output, err := conn.GetHost(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
