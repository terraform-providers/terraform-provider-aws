package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a Capacity Provider to return INACTIVE
	CapacityProviderInactiveTimeout = 20 * time.Minute

	ServiceCreateTimeout      = 2 * time.Minute
	ServiceInactiveTimeout    = 10 * time.Minute
	ServiceInactiveTimeoutMin = 1 * time.Second
	ServiceDescribeTimeout    = 2 * time.Minute
	ServiceUpdateTimeout      = 2 * time.Minute

	ClusterAvailableTimeout = 10 * time.Minute
	ClusterDeleteTimeout    = 10 * time.Minute
	ClusterAvailableDelay   = 10 * time.Second
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

func ServiceStable(conn *ecs.ECS, id, cluster string) error {
	input := &ecs.DescribeServicesInput{
		Services: aws.StringSlice([]string{id}),
	}

	if cluster != "" {
		input.Cluster = aws.String(cluster)
	}

	if err := conn.WaitUntilServicesStable(input); err != nil {
		return err
	}
	return nil
}

func ServiceInactive(conn *ecs.ECS, id, cluster string) error {
	input := &ecs.DescribeServicesInput{
		Services: aws.StringSlice([]string{id}),
	}

	if cluster != "" {
		input.Cluster = aws.String(cluster)
	}

	if err := conn.WaitUntilServicesInactive(input); err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ServiceStatusActive, ServiceStatusDraining},
		Target:     []string{ServiceStatusInactive, ServiceStatusNone},
		Refresh:    ServiceStatus(conn, id, cluster),
		Timeout:    ServiceInactiveTimeout,
		MinTimeout: ServiceInactiveTimeoutMin,
	}

	_, err := stateConf.WaitForState()

	if err != nil {
		return err
	}

	return nil
}

func ServiceDescribeReady(conn *ecs.ECS, id, cluster string) (*ecs.DescribeServicesOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ServiceStatusInactive, ServiceStatusDraining, ServiceStatusNone},
		Target:  []string{ServiceStatusActive},
		Refresh: ServiceStatus(conn, id, cluster),
		Timeout: ServiceDescribeTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ecs.DescribeServicesOutput); ok {
		return v, err
	}

	return nil, err
}

func ClusterAvailable(conn *ecs.ECS, arn string) (*ecs.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"PROVISIONING"},
		Target:  []string{"ACTIVE"},
		Refresh: ClusterStatus(conn, arn),
		Timeout: ClusterAvailableTimeout,
		Delay:   ClusterAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ecs.Cluster); ok {
		return v, err
	}

	return nil, err
}

func ClusterDeleted(conn *ecs.ECS, arn string) (*ecs.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"ACTIVE", "DEPROVISIONING"},
		Target:  []string{"INACTIVE"},
		Refresh: ClusterStatus(conn, arn),
		Timeout: ClusterDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*ecs.Cluster); ok {
		return v, err
	}

	return nil, err
}
