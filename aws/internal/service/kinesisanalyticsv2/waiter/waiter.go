package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// ApplicationDeleted waits for an Application to return Deleted
func ApplicationDeleted(conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string, timeout time.Duration) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusDeleting},
		Target:  []string{},
		Refresh: ApplicationStatus(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}
