package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// ConnectInstanceCreateTimeout Timeout for connect instance creation
	ConnectInstanceCreateTimeout = 5 * time.Minute
)

func InstanceCreated(ctx context.Context, conn *connect.Connect, instanceId string) (*connect.DescribeInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.InstanceStatusCreationInProgress},
		Target:  []string{connect.InstanceStatusActive},
		Refresh: InstanceStatus(ctx, conn, instanceId),
		Timeout: ConnectInstanceCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*connect.DescribeInstanceOutput); ok {
		return v, err
	}

	return nil, err
}
