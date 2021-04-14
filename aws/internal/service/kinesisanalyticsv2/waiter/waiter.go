package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	ApplicationDeletedTimeout = 5 * time.Minute
	ApplicationStartedTimeout = 5 * time.Minute
	ApplicationStoppedTimeout = 5 * time.Minute
	ApplicationUpdatedTimeout = 5 * time.Minute

	SnapshotCreatedTimeout = 5 * time.Minute
	SnapshotDeletedTimeout = 5 * time.Minute
)

// ApplicationDeleted waits for an Application to return Deleted
func ApplicationDeleted(conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusDeleting},
		Target:  []string{},
		Refresh: ApplicationStatus(conn, name),
		Timeout: ApplicationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// ApplicationStarted waits for an Application to start
func ApplicationStarted(conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusStarting},
		Target:  []string{kinesisanalyticsv2.ApplicationStatusRunning},
		Refresh: ApplicationStatus(conn, name),
		Timeout: ApplicationStartedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// ApplicationStopped waits for an Application to stop
func ApplicationStopped(conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusForceStopping, kinesisanalyticsv2.ApplicationStatusStopping},
		Target:  []string{kinesisanalyticsv2.ApplicationStatusReady},
		Refresh: ApplicationStatus(conn, name),
		Timeout: ApplicationStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// ApplicationUpdated waits for an Application to return Deleted
func ApplicationUpdated(conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string) (*kinesisanalyticsv2.ApplicationDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.ApplicationStatusUpdating},
		Target:  []string{kinesisanalyticsv2.ApplicationStatusReady, kinesisanalyticsv2.ApplicationStatusRunning},
		Refresh: ApplicationStatus(conn, name),
		Timeout: ApplicationUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.ApplicationDetail); ok {
		return v, err
	}

	return nil, err
}

// IAMPropagation retries the specified function if the returned error indicates an IAM eventual consistency issue.
// If the retries time out the specified function is called one last time.
func IAMPropagation(f func() (interface{}, error)) (interface{}, error) {
	var output interface{}

	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = f()

		// Kinesis Stream: https://github.com/hashicorp/terraform-provider-aws/issues/7032
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics service doesn't have sufficient privileges") {
			return resource.RetryableError(err)
		}

		// Kinesis Firehose: https://github.com/hashicorp/terraform-provider-aws/issues/7394
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Kinesis Analytics doesn't have sufficient privileges") {
			return resource.RetryableError(err)
		}

		// InvalidArgumentException: Given IAM role arn : arn:aws:iam::123456789012:role/xxx does not provide Invoke permissions on the Lambda resource : arn:aws:lambda:us-west-2:123456789012:function:yyy
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "does not provide Invoke permissions on the Lambda resource") {
			return resource.RetryableError(err)
		}

		// S3: https://github.com/hashicorp/terraform-provider-aws/issues/16104
		if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "Please check the role provided or validity of S3 location you provided") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = f()
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// SnapshotCreated waits for a Snapshot to return Created
func SnapshotCreated(conn *kinesisanalyticsv2.KinesisAnalyticsV2, applicationName, snapshotName string) (*kinesisanalyticsv2.SnapshotDetails, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.SnapshotStatusCreating},
		Target:  []string{kinesisanalyticsv2.SnapshotStatusReady},
		Refresh: SnapshotDetailsStatus(conn, applicationName, snapshotName),
		Timeout: SnapshotCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.SnapshotDetails); ok {
		return v, err
	}

	return nil, err
}

// SnapshotDeleted waits for a Snapshot to return Deleted
func SnapshotDeleted(conn *kinesisanalyticsv2.KinesisAnalyticsV2, applicationName, snapshotName string) (*kinesisanalyticsv2.SnapshotDetails, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesisanalyticsv2.SnapshotStatusDeleting},
		Target:  []string{},
		Refresh: SnapshotDetailsStatus(conn, applicationName, snapshotName),
		Timeout: SnapshotDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesisanalyticsv2.SnapshotDetails); ok {
		return v, err
	}

	return nil, err
}
