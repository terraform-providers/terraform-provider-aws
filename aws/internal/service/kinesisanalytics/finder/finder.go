package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// ApplicationDetailByName returns the application corresponding to the specified name.
// Returns NotFoundError if no application is found.
func ApplicationDetailByName(conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) {
	input := &kinesisanalytics.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}

	return ApplicationDetail(conn, input)
}

// ApplicationDetail returns the application details corresponding to the specified name.
// Returns NotFoundError if no application is found.
func ApplicationDetail(conn *kinesisanalytics.KinesisAnalytics, input *kinesisanalytics.DescribeApplicationInput) (*kinesisanalytics.ApplicationDetail, error) {
	output, err := conn.DescribeApplication(input)

	if tfawserr.ErrCodeEquals(err, kinesisanalytics.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationDetail == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.ApplicationDetail, nil
}
