package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	// Maximum amount of time to wait for a Capacity Provider to return INACTIVE
	CapacityProviderInactiveTimeout = 20 * time.Minute
)

// CapacityProviderInactive waits for a Capacity Provider to return INACTIVE
func CapacityProviderInactive(conn *ecs.ECS, capacityProvider string) (*ecs.CapacityProvider, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ecs.CapacityProviderStatusActive},
		Target:  []string{ecs.CapacityProviderStatusInactive},
		Refresh: CapacityProviderStatus(conn, capacityProvider),
		Timeout: CapacityProviderInactiveTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ecs.CapacityProvider); ok {
		return v, err
	}

	return nil, err
}
