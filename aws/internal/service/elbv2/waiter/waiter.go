package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Default maximum amount of time to wait for a Load Balancer to be created
	LoadBalancerCreateTimeout = 10 * time.Minute

	// Default maximum amount of time to wait for a Load Balancer to be updated
	LoadBalancerUpdateTimeout = 10 * time.Minute

	// Default maximum amount of time to wait for a Load Balancer to be deleted
	LoadBalancerDeleteTimeout = 10 * time.Minute
)

// LoadBalancerActive waits for a Load Balancer to return active
func LoadBalancerActive(conn *elbv2.ELBV2, arn string, timeout time.Duration) (*elbv2.LoadBalancer, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{elbv2.LoadBalancerStateEnumProvisioning, elbv2.LoadBalancerStateEnumFailed},
		Target:     []string{elbv2.LoadBalancerStateEnumActive},
		Refresh:    LoadBalancerState(conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}
	output, err := stateConf.WaitForState()

	if v, ok := output.(*elbv2.LoadBalancer); ok {
		return v, err
	}
	return nil, err
}
