package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53/finder"
)

func ChangeInfoStatus(conn *route53.Route53, changeID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &route53.GetChangeInput{
			Id: aws.String(changeID),
		}

		output, err := conn.GetChange(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.ChangeInfo == nil {
			return nil, "", nil
		}

		return output.ChangeInfo, aws.StringValue(output.ChangeInfo.Status), nil
	}
}

func KeySigningKeyStatus(conn *route53.Route53, hostedZoneID string, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		keySigningKey, err := finder.KeySigningKey(conn, hostedZoneID, name)

		if err != nil {
			return nil, "", err
		}

		if keySigningKey == nil {
			return nil, "", nil
		}

		return keySigningKey, aws.StringValue(keySigningKey.Status), nil
	}
}
