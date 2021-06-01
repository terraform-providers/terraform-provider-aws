package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
)

// EndpointGroupByARN returns the endpoint group corresponding to the specified ARN.
func EndpointGroupByARN(conn *globalaccelerator.GlobalAccelerator, arn string) (*globalaccelerator.EndpointGroup, error) {
	input := &globalaccelerator.DescribeEndpointGroupInput{
		EndpointGroupArn: aws.String(arn),
	}

	output, err := conn.DescribeEndpointGroup(input)
	if err != nil {
		return nil, err
	}

	return output.EndpointGroup, nil
}
