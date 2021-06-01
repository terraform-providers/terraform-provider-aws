package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalytics/finder"
)

const (
	applicationStatusNotFound = "NotFound"
	applicationStatusUnknown  = "Unknown"
)

// ApplicationStatus fetches the Application and its Status
func ApplicationStatus(conn *kinesisanalytics.KinesisAnalytics, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		application, err := finder.ApplicationByName(conn, name)

		if tfawserr.ErrCodeEquals(err, kinesisanalytics.ErrCodeResourceNotFoundException) {
			return nil, applicationStatusNotFound, nil
		}

		if err != nil {
			return nil, applicationStatusUnknown, err
		}

		return application, aws.StringValue(application.ApplicationStatus), nil
	}
}
