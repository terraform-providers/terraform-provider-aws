package s3

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#pkg-constants

const (
	ErrCodeNoSuchPublicAccessBlockConfiguration = "NoSuchPublicAccessBlockConfiguration"
	ErrCodeNoSuchTagSet                         = "NoSuchTagSet"
)
