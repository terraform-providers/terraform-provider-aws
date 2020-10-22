package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Application to be deleted
	ApplicationDeletedTimeout = 20 * time.Minute
)

// ApplicationDeleted waits for an Application to be deleted
func ApplicationDeleted(conn *kinesisanalytics.KinesisAnalytics, applicationName string) (*kinesisanalytics.ApplicationSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalytics.ApplicationStatusRunning, kinesisanalytics.ApplicationStatusDeleting},
		Target:  []string{ApplicationStatusNotFound},
		Refresh: ApplicationStatus(conn, applicationName),
		Timeout: ApplicationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalytics.ApplicationSummary); ok {
		return v, err
	}

	return nil, err
}
