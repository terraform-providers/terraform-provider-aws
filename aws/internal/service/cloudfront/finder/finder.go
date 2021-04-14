package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

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
