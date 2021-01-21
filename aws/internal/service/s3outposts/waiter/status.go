package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/s3outposts/finder"
)

const (
	EndpointStatusNotFound = "NotFound"
	EndpointStatusUnknown  = "Unknown"
)

// EndpointStatus fetches the Endpoint and its Status
func EndpointStatus(conn *s3outposts.S3Outposts, endpointArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		endpoint, err := finder.Endpoint(conn, endpointArn)

		if err != nil {
			return nil, EndpointStatusUnknown, err
		}

		if endpoint == nil {
			return nil, EndpointStatusNotFound, nil
		}

		return endpoint, aws.StringValue(endpoint.Status), nil
	}
}
