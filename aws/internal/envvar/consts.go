package envvar

// Standard AWS environment variables used in the Terraform AWS Provider testing.
// These are not provided as constants in the AWS Go SDK currently.
const (
	// Default static credential identifier for tests (AWS Go SDK does not provide this as constant)
	// See also AWS_SECRET_ACCESS_KEY and AWS_PROFILE
	AwsAccessKeyId = "AWS_ACCESS_KEY_ID"

	// Container credentials endpoint
	// See also AWS_ACCESS_KEY_ID and AWS_PROFILE
	AwsContainerCredentialsFullUri = "AWS_CONTAINER_CREDENTIALS_FULL_URI"

	// Default AWS region for tests (AWS Go SDK does not provide this as constant)
	AwsDefaultRegion = "AWS_DEFAULT_REGION"

	// Default AWS shared configuration profile for tests (AWS Go SDK does not provide this as constant)
	AwsProfile = "AWS_PROFILE"

	// Default static credential value for tests (AWS Go SDK does not provide this as constant)
	// See also AWS_ACCESS_KEY_ID and AWS_PROFILE
	AwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
)

// Custom environment variables used in the Terraform AWS Provider testing.
// Additions should also be documented in the Environment Variable Dictionary
// of the Maintainers Guide: docs/MAINTAINING.md
const (
	// For tests using an alternate AWS account, the equivalent of AWS_ACCESS_KEY_ID for that account
	AwsAlternateAccessKeyId = "AWS_ALTERNATE_ACCESS_KEY_ID"

	// For tests using an alternate AWS account, the equivalent of AWS_PROFILE for that account
	AwsAlternateProfile = "AWS_ALTERNATE_PROFILE"

	// For tests using an alternate AWS region, the equivalent of AWS_DEFAULT_REGION for that account
	AwsAlternateRegion = "AWS_ALTERNATE_REGION"

	// For tests using an alternate AWS account, the equivalent of AWS_SECRET_ACCESS_KEY for that account
	AwsAlternateSecretAccessKey = "AWS_ALTERNATE_SECRET_ACCESS_KEY"

	// For tests using a third AWS region, the equivalent of AWS_DEFAULT_REGION for that region
	AwsThirdRegion = "AWS_THIRD_REGION"

	// For tests requiring GitHub permissions
	GithubToken = "GITHUB_TOKEN"

	// For tests requiring restricted IAM permissions, an existing IAM Role to assume
	// An inline assume role policy is then used to deny actions for the test
	TfAccAssumeRoleArn = "TF_ACC_ASSUME_ROLE_ARN"
)
