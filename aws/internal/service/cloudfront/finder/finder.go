package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FunctionByNameAndStage(conn *cloudfront.CloudFront, name, stage string) (*cloudfront.DescribeFunctionOutput, error) {
	input := &cloudfront.DescribeFunctionInput{
		Name:  aws.String(name),
		Stage: aws.String(stage),
	}

	output, err := conn.DescribeFunction(input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchFunctionExists) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}

// RealtimeLogConfigByARN returns the real-time log configuration corresponding to the specified ARN.
// Returns nil if no configuration is found.
func RealtimeLogConfigByARN(conn *cloudfront.CloudFront, arn string) (*cloudfront.RealtimeLogConfig, error) {
	input := &cloudfront.GetRealtimeLogConfigInput{
		ARN: aws.String(arn),
	}

	output, err := conn.GetRealtimeLogConfig(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.RealtimeLogConfig, nil
}
