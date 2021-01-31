package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
)

// ApplicationByName returns the application corresponding to the specified name.
func ApplicationByName(conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) {
	input := &kinesisanalytics.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}

	output, err := conn.DescribeApplication(input)
	if err != nil {
		return nil, err
	}

	return output.ApplicationDetail, nil
}
