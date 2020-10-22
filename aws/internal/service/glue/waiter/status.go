package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	MLTransformStatusUnknown = "Unknown"
)

// MLTransformStatus fetches the Operation and its Status
func MLTransformStatus(conn *glue.Glue, transformId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &glue.GetMLTransformInput{
			TransformId: aws.String(transformId),
		}

		output, err := conn.GetMLTransform(input)

		if err != nil {
			return nil, MLTransformStatusUnknown, err
		}

		if output == nil {
			return output, MLTransformStatusUnknown, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
