package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3outposts"
)

// Endpoint returns matching Endpoint by ARN.
func Endpoint(conn *s3outposts.S3Outposts, endpointArn string) (*s3outposts.Endpoint, error) {
	input := &s3outposts.ListEndpointsInput{}
	var result *s3outposts.Endpoint

	err := conn.ListEndpointsPages(input, func(page *s3outposts.ListEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, endpoint := range page.Endpoints {
			if aws.StringValue(endpoint.EndpointArn) == endpointArn {
				result = endpoint
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
