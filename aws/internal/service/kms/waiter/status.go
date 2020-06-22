package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// KeyState fetches the Key and its State
func KeyState(conn *kms.KMS, keyID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &kms.DescribeKeyInput{
			KeyId: aws.String(keyID),
		}

		output, err := conn.DescribeKey(input)

		if err != nil {
			return nil, kms.KeyStateUnavailable, err
		}

		if output == nil || output.KeyMetadata == nil {
			return output, kms.KeyStateUnavailable, nil
		}

		return output, aws.StringValue(output.KeyMetadata.KeyState), nil
	}
}
