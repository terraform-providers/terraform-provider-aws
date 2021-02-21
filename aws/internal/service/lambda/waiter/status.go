package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	EventSourceMappingStateCreating  = "Creating"
	EventSourceMappingStateDisabled  = "Disabled"
	EventSourceMappingStateDisabling = "Disabling"
	EventSourceMappingStateEnabled   = "Enabled"
	EventSourceMappingStateEnabling  = "Enabling"
	EventSourceMappingStateUpdating  = "Updating"
)

func EventSourceMappingState(conn *lambda.Lambda, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lambda.GetEventSourceMappingInput{
			UUID: aws.String(id),
		}

		output, err := conn.GetEventSourceMapping(input)

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.State), nil
	}
}
