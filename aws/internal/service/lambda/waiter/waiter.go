package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	EventSourceMappingCreateTimeout      = 10 * time.Minute
	EventSourceMappingUpdateTimeout      = 10 * time.Minute
	LambdaFunctionCreateTimeout          = 5 * time.Minute
	LambdaFunctionUpdateTimeout          = 5 * time.Minute
	LambdaFunctionPublishTimeout         = 2 * time.Minute
	LambdaFunctionPutConcurrencyTimeout  = 1 * time.Minute
	LambdaFunctionExtraThrottlingTimeout = 9 * time.Minute
)

func EventSourceMappingCreate(conn *lambda.Lambda, id string) (*lambda.EventSourceMappingConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			EventSourceMappingStateCreating,
			EventSourceMappingStateDisabling,
			EventSourceMappingStateEnabling,
		},
		Target: []string{
			EventSourceMappingStateDisabled,
			EventSourceMappingStateEnabled,
		},
		Refresh: EventSourceMappingState(conn, id),
		Timeout: EventSourceMappingCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lambda.EventSourceMappingConfiguration); ok {
		return output, err
	}

	return nil, err
}

func EventSourceMappingUpdate(conn *lambda.Lambda, id string) (*lambda.EventSourceMappingConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			EventSourceMappingStateDisabling,
			EventSourceMappingStateEnabling,
			EventSourceMappingStateUpdating,
		},
		Target: []string{
			EventSourceMappingStateDisabled,
			EventSourceMappingStateEnabled,
		},
		Refresh: EventSourceMappingState(conn, id),
		Timeout: EventSourceMappingUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lambda.EventSourceMappingConfiguration); ok {
		return output, err
	}

	return nil, err
}
