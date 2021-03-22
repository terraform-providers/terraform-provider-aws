package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

const (
	SagemakerNotebookInstanceStatusNotFound  = "NotFound"
	SagemakerImageStatusNotFound             = "NotFound"
	SagemakerImageStatusFailed               = "Failed"
	SagemakerImageVersionStatusNotFound      = "NotFound"
	SagemakerImageVersionStatusFailed        = "Failed"
	SagemakerDomainStatusNotFound            = "NotFound"
	SagemakerFeatureGroupStatusNotFound      = "NotFound"
	SagemakerFeatureGroupStatusUnknown       = "Unknown"
	SagemakerUserProfileStatusNotFound       = "NotFound"
	SagemakerModelPackageGroupStatusNotFound = "NotFound"
	SagemakerAppStatusNotFound               = "NotFound"
)

// NotebookInstanceStatus fetches the NotebookInstance and its Status
func NotebookInstanceStatus(conn *sagemaker.SageMaker, notebookName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(notebookName),
		}

		output, err := conn.DescribeNotebookInstance(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
			return nil, SagemakerNotebookInstanceStatusNotFound, nil
		}

		if err != nil {
			return nil, sagemaker.NotebookInstanceStatusFailed, err
		}

		if output == nil {
			return nil, SagemakerNotebookInstanceStatusNotFound, nil
		}

		return output, aws.StringValue(output.NotebookInstanceStatus), nil
	}
}

// ModelPackageGroupStatus fetches the ModelPackageGroup and its Status
func ModelPackageGroupStatus(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeModelPackageGroupInput{
			ModelPackageGroupName: aws.String(name),
		}

		output, err := conn.DescribeModelPackageGroup(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			return nil, SagemakerModelPackageGroupStatusNotFound, nil
		}

		if err != nil {
			return nil, sagemaker.ModelPackageGroupStatusFailed, err
		}

		if output == nil {
			return nil, SagemakerModelPackageGroupStatusNotFound, nil
		}

		return output, aws.StringValue(output.ModelPackageGroupStatus), nil
	}
}

// ImageStatus fetches the Image and its Status
func ImageStatus(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeImageInput{
			ImageName: aws.String(name),
		}

		output, err := conn.DescribeImage(input)

		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "No Image with the name") {
			return nil, SagemakerImageStatusNotFound, nil
		}

		if err != nil {
			return nil, SagemakerImageStatusFailed, err
		}

		if output == nil {
			return nil, SagemakerImageStatusNotFound, nil
		}

		if aws.StringValue(output.ImageStatus) == sagemaker.ImageStatusCreateFailed {
			return output, sagemaker.ImageStatusCreateFailed, fmt.Errorf("%s", aws.StringValue(output.FailureReason))
		}

		return output, aws.StringValue(output.ImageStatus), nil
	}
}

// ImageVersionStatus fetches the ImageVersion and its Status
func ImageVersionStatus(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeImageVersionInput{
			ImageName: aws.String(name),
		}

		output, err := conn.DescribeImageVersion(input)

		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "No ImageVersion with the name") {
			return nil, SagemakerImageVersionStatusNotFound, nil
		}

		if err != nil {
			return nil, SagemakerImageVersionStatusFailed, err
		}

		if output == nil {
			return nil, SagemakerImageVersionStatusNotFound, nil
		}

		if aws.StringValue(output.ImageVersionStatus) == sagemaker.ImageVersionStatusCreateFailed {
			return output, sagemaker.ImageVersionStatusCreateFailed, fmt.Errorf("%s", aws.StringValue(output.FailureReason))
		}

		return output, aws.StringValue(output.ImageVersionStatus), nil
	}
}

// DomainStatus fetches the Domain and its Status
func DomainStatus(conn *sagemaker.SageMaker, domainID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeDomainInput{
			DomainId: aws.String(domainID),
		}

		output, err := conn.DescribeDomain(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
			return nil, sagemaker.UserProfileStatusFailed, nil
		}

		if err != nil {
			return nil, sagemaker.DomainStatusFailed, err
		}

		if output == nil {
			return nil, SagemakerDomainStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// FeatureGroupStatus fetches the Feature Group and its Status
func FeatureGroupStatus(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FeatureGroupByName(conn, name)
		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return nil, SagemakerFeatureGroupStatusNotFound, nil
		}

		if err != nil {
			return nil, SagemakerFeatureGroupStatusUnknown, err
		}

		if output == nil {
			return nil, SagemakerFeatureGroupStatusNotFound, nil
		}

		return output, aws.StringValue(output.FeatureGroupStatus), nil
	}
}

// UserProfileStatus fetches the UserProfile and its Status
func UserProfileStatus(conn *sagemaker.SageMaker, domainID, userProfileName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeUserProfileInput{
			DomainId:        aws.String(domainID),
			UserProfileName: aws.String(userProfileName),
		}

		output, err := conn.DescribeUserProfile(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
			return nil, SagemakerUserProfileStatusNotFound, nil
		}

		if err != nil {
			return nil, sagemaker.UserProfileStatusFailed, err
		}

		if output == nil {
			return nil, SagemakerUserProfileStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// AppStatus fetches the App and its Status
func AppStatus(conn *sagemaker.SageMaker, domainID, userProfileName, appType, appName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeAppInput{
			DomainId:        aws.String(domainID),
			UserProfileName: aws.String(userProfileName),
			AppType:         aws.String(appType),
			AppName:         aws.String(appName),
		}

		output, err := conn.DescribeApp(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
			return nil, SagemakerAppStatusNotFound, nil
		}

		if err != nil {
			return nil, sagemaker.AppStatusFailed, err
		}

		if output == nil {
			return nil, SagemakerAppStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
