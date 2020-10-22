package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// ApplicationStatus NotFound
	ApplicationStatusNotFound = "NotFound"

	// ApplicationStatus Unknown
	ApplicationStatusUnknown = "Unknown"
)

// ApplicationStatus fetches the Application and its Status
func ApplicationStatus(conn *kinesisanalytics.KinesisAnalytics, applicationName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &kinesisanalytics.DescribeApplicationInput{
			ApplicationName: aws.String(applicationName),
		}

		output, err := conn.DescribeApplication(input)

		if err != nil {
			return nil, ApplicationStatusUnknown, err
		}

		application := output.ApplicationDetail

		if application == nil {
			return application, ApplicationStatusNotFound, nil
		}

		return application, aws.StringValue(application.ApplicationStatus), nil
	}
}
