package finder

import (
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func CallerIdentity(conn *sts.STS) (*sts.GetCallerIdentityOutput, error) {
	input := &sts.GetCallerIdentityInput{}

	output, err := conn.GetCallerIdentity(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}
