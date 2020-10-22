package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
)

// ApplicationByName returns the application corresponding to the specified name.
func ApplicationByName(conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string) (*kinesisanalyticsv2.ApplicationDetail, error) {
	input := &kinesisanalyticsv2.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}

	output, err := conn.DescribeApplication(input)
	if err != nil {
		return nil, err
	}

	return output.ApplicationDetail, nil
}
