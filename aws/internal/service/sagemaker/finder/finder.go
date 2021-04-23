package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
)

// CodeRepositoryByName returns the code repository corresponding to the specified name.
// Returns nil if no code repository is found.
func CodeRepositoryByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeCodeRepositoryOutput, error) {
	input := &sagemaker.DescribeCodeRepositoryInput{
		CodeRepositoryName: aws.String(name),
	}

	output, err := conn.DescribeCodeRepository(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

// ImageByName returns the Image corresponding to the specified name.
// Returns nil if no Image is found.
func ImageByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageOutput, error) {
	input := &sagemaker.DescribeImageInput{
		ImageName: aws.String(name),
	}

	output, err := conn.DescribeImage(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

// ImageVersionByName returns the Image Version corresponding to the specified name.
// Returns nil if no Image Version is found.
func ImageVersionByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageVersionOutput, error) {
	input := &sagemaker.DescribeImageVersionInput{
		ImageName: aws.String(name),
	}

	output, err := conn.DescribeImageVersion(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

// DomainByName returns the domain corresponding to the specified domain id.
// Returns nil if no domain is found.
func DomainByName(conn *sagemaker.SageMaker, domainID string) (*sagemaker.DescribeDomainOutput, error) {
	input := &sagemaker.DescribeDomainInput{
		DomainId: aws.String(domainID),
	}

	output, err := conn.DescribeDomain(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

// FeatureGroupByName returns the feature group corresponding to the specified name.
// Returns nil if no feature group is found.
func FeatureGroupByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	input := &sagemaker.DescribeFeatureGroupInput{
		FeatureGroupName: aws.String(name),
	}

	output, err := conn.DescribeFeatureGroup(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

// UserProfileByName returns the domain corresponding to the specified domain id.
// Returns nil if no domain is found.
func UserProfileByName(conn *sagemaker.SageMaker, domainID, userProfileName string) (*sagemaker.DescribeUserProfileOutput, error) {
	input := &sagemaker.DescribeUserProfileInput{
		DomainId:        aws.String(domainID),
		UserProfileName: aws.String(userProfileName),
	}

	output, err := conn.DescribeUserProfile(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}
