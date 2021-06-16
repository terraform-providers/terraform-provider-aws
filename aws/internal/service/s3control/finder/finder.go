package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
)

func PublicAccessBlockConfiguration(conn *s3control.S3Control, accountID string) (*s3control.PublicAccessBlockConfiguration, error) {
	input := &s3control.GetPublicAccessBlockInput{
		AccountId: aws.String(accountID),
	}

	output, err := conn.GetPublicAccessBlock(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.PublicAccessBlockConfiguration, nil
}
