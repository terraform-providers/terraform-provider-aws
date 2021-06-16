package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a Canary to return Ready
	CanaryCreatedTimeout = 5 * time.Minute
	CanaryRunningTimeout = 5 * time.Minute
	CanaryStoppedTimeout = 5 * time.Minute
	CanaryDeletedTimeout = 5 * time.Minute
)

// CanaryReady waits for a Canary to return Ready
func CanaryReady(conn *synthetics.Synthetics, name string) (*synthetics.GetCanaryOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{synthetics.CanaryStateCreating, synthetics.CanaryStateUpdating},
		Target:  []string{synthetics.CanaryStateReady},
		Refresh: CanaryStatus(conn, name),
		Timeout: CanaryCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*synthetics.GetCanaryOutput); ok {
		return v, err
	}

	return nil, err
}

// CanaryReady waits for a Canary to return Stopped
func CanaryStopped(conn *synthetics.Synthetics, name string) (*synthetics.GetCanaryOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{synthetics.CanaryStateStopping, synthetics.CanaryStateUpdating,
			synthetics.CanaryStateRunning, synthetics.CanaryStateReady, synthetics.CanaryStateStarting},
		Target:  []string{synthetics.CanaryStateStopped},
		Refresh: CanaryStatus(conn, name),
		Timeout: CanaryStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*synthetics.GetCanaryOutput); ok {
		return v, err
	}

	return nil, err
}

// CanaryReady waits for a Canary to return Running
func CanaryRunning(conn *synthetics.Synthetics, name string) (*synthetics.GetCanaryOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{synthetics.CanaryStateStarting, synthetics.CanaryStateUpdating,
			synthetics.CanaryStateStarting, synthetics.CanaryStateReady},
		Target:  []string{synthetics.CanaryStateRunning},
		Refresh: CanaryStatus(conn, name),
		Timeout: CanaryRunningTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*synthetics.GetCanaryOutput); ok {
		return v, err
	}

	return nil, err
}

// CanaryReady waits for a Canary to return Ready
func CanaryDeleted(conn *synthetics.Synthetics, name string) (*synthetics.GetCanaryOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{synthetics.CanaryStateDeleting, synthetics.CanaryStateStopped},
		Target:  []string{},
		Refresh: CanaryStatus(conn, name),
		Timeout: CanaryDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*synthetics.GetCanaryOutput); ok {
		return v, err
	}

	return nil, err
}
