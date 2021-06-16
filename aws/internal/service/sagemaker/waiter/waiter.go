package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	NotebookInstanceInServiceTimeout  = 60 * time.Minute
	NotebookInstanceStoppedTimeout    = 10 * time.Minute
	NotebookInstanceDeletedTimeout    = 10 * time.Minute
	ModelPackageGroupCompletedTimeout = 10 * time.Minute
	ModelPackageGroupDeletedTimeout   = 10 * time.Minute
	ImageCreatedTimeout               = 10 * time.Minute
	ImageDeletedTimeout               = 10 * time.Minute
	ImageVersionCreatedTimeout        = 10 * time.Minute
	ImageVersionDeletedTimeout        = 10 * time.Minute
	DomainInServiceTimeout            = 10 * time.Minute
	DomainDeletedTimeout              = 10 * time.Minute
	FeatureGroupCreatedTimeout        = 10 * time.Minute
	FeatureGroupDeletedTimeout        = 10 * time.Minute
	UserProfileInServiceTimeout       = 10 * time.Minute
	UserProfileDeletedTimeout         = 10 * time.Minute
	AppInServiceTimeout               = 10 * time.Minute
	AppDeletedTimeout                 = 10 * time.Minute
)

// NotebookInstanceInService waits for a NotebookInstance to return InService
func NotebookInstanceInService(conn *sagemaker.SageMaker, notebookName string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			SagemakerNotebookInstanceStatusNotFound,
			sagemaker.NotebookInstanceStatusUpdating,
			sagemaker.NotebookInstanceStatusPending,
			sagemaker.NotebookInstanceStatusStopped,
		},
		Target:  []string{sagemaker.NotebookInstanceStatusInService},
		Refresh: NotebookInstanceStatus(conn, notebookName),
		Timeout: NotebookInstanceInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		return output, err
	}

	return nil, err
}

// NotebookInstanceStopped waits for a NotebookInstance to return Stopped
func NotebookInstanceStopped(conn *sagemaker.SageMaker, notebookName string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.NotebookInstanceStatusUpdating,
			sagemaker.NotebookInstanceStatusStopping,
		},
		Target:  []string{sagemaker.NotebookInstanceStatusStopped},
		Refresh: NotebookInstanceStatus(conn, notebookName),
		Timeout: NotebookInstanceStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		return output, err
	}

	return nil, err
}

// NotebookInstanceDeleted waits for a NotebookInstance to return Deleted
func NotebookInstanceDeleted(conn *sagemaker.SageMaker, notebookName string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.NotebookInstanceStatusDeleting,
		},
		Target:  []string{},
		Refresh: NotebookInstanceStatus(conn, notebookName),
		Timeout: NotebookInstanceDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		return output, err
	}

	return nil, err
}

// ModelPackageGroupCompleted waits for a ModelPackageGroup to return Created
func ModelPackageGroupCompleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeModelPackageGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.ModelPackageGroupStatusPending,
			sagemaker.ModelPackageGroupStatusInProgress,
		},
		Target:  []string{sagemaker.ModelPackageGroupStatusCompleted},
		Refresh: ModelPackageGroupStatus(conn, name),
		Timeout: ModelPackageGroupCompletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeModelPackageGroupOutput); ok {
		return output, err
	}

	return nil, err
}

// ModelPackageGroupDeleted waits for a ModelPackageGroup to return Created
func ModelPackageGroupDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeModelPackageGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.ModelPackageGroupStatusDeleting,
		},
		Target:  []string{},
		Refresh: ModelPackageGroupStatus(conn, name),
		Timeout: ModelPackageGroupDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeModelPackageGroupOutput); ok {
		return output, err
	}

	return nil, err
}

// ImageCreated waits for a Image to return Created
func ImageCreated(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.ImageStatusCreating,
			sagemaker.ImageStatusUpdating,
		},
		Target:  []string{sagemaker.ImageStatusCreated},
		Refresh: ImageStatus(conn, name),
		Timeout: ImageCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeImageOutput); ok {
		return output, err
	}

	return nil, err
}

// ImageDeleted waits for a Image to return Deleted
func ImageDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.ImageStatusDeleting},
		Target:  []string{},
		Refresh: ImageStatus(conn, name),
		Timeout: ImageDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeImageOutput); ok {
		return output, err
	}

	return nil, err
}

// ImageVersionCreated waits for a ImageVersion to return Created
func ImageVersionCreated(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageVersionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.ImageVersionStatusCreating,
		},
		Target:  []string{sagemaker.ImageVersionStatusCreated},
		Refresh: ImageVersionStatus(conn, name),
		Timeout: ImageVersionCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeImageVersionOutput); ok {
		return output, err
	}

	return nil, err
}

// ImageVersionDeleted waits for a ImageVersion to return Deleted
func ImageVersionDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageVersionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.ImageVersionStatusDeleting},
		Target:  []string{},
		Refresh: ImageVersionStatus(conn, name),
		Timeout: ImageVersionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeImageVersionOutput); ok {
		return output, err
	}

	return nil, err
}

// DomainInService waits for a Domain to return InService
func DomainInService(conn *sagemaker.SageMaker, domainID string) (*sagemaker.DescribeDomainOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			SagemakerDomainStatusNotFound,
			sagemaker.DomainStatusPending,
			sagemaker.DomainStatusUpdating,
		},
		Target:  []string{sagemaker.DomainStatusInService},
		Refresh: DomainStatus(conn, domainID),
		Timeout: DomainInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeDomainOutput); ok {
		return output, err
	}

	return nil, err
}

// DomainDeleted waits for a Domain to return Deleted
func DomainDeleted(conn *sagemaker.SageMaker, domainID string) (*sagemaker.DescribeDomainOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.DomainStatusDeleting,
		},
		Target:  []string{},
		Refresh: DomainStatus(conn, domainID),
		Timeout: DomainDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeDomainOutput); ok {
		return output, err
	}

	return nil, err
}

// FeatureGroupCreated waits for a Feature Group to return Created
func FeatureGroupCreated(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.FeatureGroupStatusCreating},
		Target:  []string{sagemaker.FeatureGroupStatusCreated},
		Refresh: FeatureGroupStatus(conn, name),
		Timeout: FeatureGroupCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeFeatureGroupOutput); ok {
		return output, err
	}

	return nil, err
}

// FeatureGroupDeleted waits for a Feature Group to return Deleted
func FeatureGroupDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.FeatureGroupStatusDeleting},
		Target:  []string{},
		Refresh: FeatureGroupStatus(conn, name),
		Timeout: FeatureGroupDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeFeatureGroupOutput); ok {
		return output, err
	}

	return nil, err
}

// UserProfileInService waits for a UserProfile to return InService
func UserProfileInService(conn *sagemaker.SageMaker, domainID, userProfileName string) (*sagemaker.DescribeUserProfileOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			SagemakerUserProfileStatusNotFound,
			sagemaker.UserProfileStatusPending,
			sagemaker.UserProfileStatusUpdating,
		},
		Target:  []string{sagemaker.UserProfileStatusInService},
		Refresh: UserProfileStatus(conn, domainID, userProfileName),
		Timeout: UserProfileInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeUserProfileOutput); ok {
		return output, err
	}

	return nil, err
}

// UserProfileDeleted waits for a UserProfile to return Deleted
func UserProfileDeleted(conn *sagemaker.SageMaker, domainID, userProfileName string) (*sagemaker.DescribeUserProfileOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.UserProfileStatusDeleting,
		},
		Target:  []string{},
		Refresh: UserProfileStatus(conn, domainID, userProfileName),
		Timeout: UserProfileDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeUserProfileOutput); ok {
		return output, err
	}

	return nil, err
}

// AppInService waits for a App to return InService
func AppInService(conn *sagemaker.SageMaker, domainID, userProfileName, appType, appName string) (*sagemaker.DescribeAppOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			SagemakerAppStatusNotFound,
			sagemaker.AppStatusPending,
		},
		Target:  []string{sagemaker.AppStatusInService},
		Refresh: AppStatus(conn, domainID, userProfileName, appType, appName),
		Timeout: AppInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeAppOutput); ok {
		return output, err
	}

	return nil, err
}

// AppDeleted waits for a App to return Deleted
func AppDeleted(conn *sagemaker.SageMaker, domainID, userProfileName, appType, appName string) (*sagemaker.DescribeAppOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.AppStatusDeleting,
		},
		Target: []string{
			sagemaker.AppStatusDeleted,
		},
		Refresh: AppStatus(conn, domainID, userProfileName, appType, appName),
		Timeout: AppDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeAppOutput); ok {
		return output, err
	}

	return nil, err
}
