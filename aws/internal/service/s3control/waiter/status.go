package waiter

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/s3control/finder"
)

// PublicAccessBlockConfigurationBlockPublicAcls fetches the PublicAccessBlockConfiguration and its BlockPublicAcls
func PublicAccessBlockConfigurationBlockPublicAcls(conn *s3control.S3Control, accountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		publicAccessBlockConfiguration, err := finder.PublicAccessBlockConfiguration(conn, accountID)

		if err != nil {
			return nil, "false", err
		}

		if publicAccessBlockConfiguration == nil {
			return nil, "false", nil
		}

		return publicAccessBlockConfiguration, strconv.FormatBool(aws.BoolValue(publicAccessBlockConfiguration.BlockPublicAcls)), nil
	}
}

// PublicAccessBlockConfigurationBlockPublicPolicy fetches the PublicAccessBlockConfiguration and its BlockPublicPolicy
func PublicAccessBlockConfigurationBlockPublicPolicy(conn *s3control.S3Control, accountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		publicAccessBlockConfiguration, err := finder.PublicAccessBlockConfiguration(conn, accountID)

		if err != nil {
			return nil, "false", err
		}

		if publicAccessBlockConfiguration == nil {
			return nil, "false", nil
		}

		return publicAccessBlockConfiguration, strconv.FormatBool(aws.BoolValue(publicAccessBlockConfiguration.BlockPublicPolicy)), nil
	}
}

// PublicAccessBlockConfigurationIgnorePublicAcls fetches the PublicAccessBlockConfiguration and its IgnorePublicAcls
func PublicAccessBlockConfigurationIgnorePublicAcls(conn *s3control.S3Control, accountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		publicAccessBlockConfiguration, err := finder.PublicAccessBlockConfiguration(conn, accountID)

		if err != nil {
			return nil, "false", err
		}

		if publicAccessBlockConfiguration == nil {
			return nil, "false", nil
		}

		return publicAccessBlockConfiguration, strconv.FormatBool(aws.BoolValue(publicAccessBlockConfiguration.IgnorePublicAcls)), nil
	}
}

// PublicAccessBlockConfigurationRestrictPublicBuckets fetches the PublicAccessBlockConfiguration and its RestrictPublicBuckets
func PublicAccessBlockConfigurationRestrictPublicBuckets(conn *s3control.S3Control, accountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		publicAccessBlockConfiguration, err := finder.PublicAccessBlockConfiguration(conn, accountID)

		if err != nil {
			return nil, "false", err
		}

		if publicAccessBlockConfiguration == nil {
			return nil, "false", nil
		}

		return publicAccessBlockConfiguration, strconv.FormatBool(aws.BoolValue(publicAccessBlockConfiguration.RestrictPublicBuckets)), nil
	}
}
