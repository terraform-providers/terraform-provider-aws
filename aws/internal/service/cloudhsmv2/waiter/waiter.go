package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func ClusterActive(conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cloudhsmv2.ClusterStateCreateInProgress,
			cloudhsmv2.ClusterStateInitializeInProgress,
		},
		Target:     []string{cloudhsmv2.ClusterStateActive},
		Refresh:    ClusterState(conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*cloudhsmv2.Cluster); ok {
		return v, err
	}

	return nil, err
}

func ClusterDeleted(conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{cloudhsmv2.ClusterStateDeleteInProgress},
		Target:     []string{cloudhsmv2.ClusterStateDeleted},
		Refresh:    ClusterState(conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*cloudhsmv2.Cluster); ok {
		return v, err
	}

	return nil, err
}

func ClusterUninitialized(conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cloudhsmv2.ClusterStateCreateInProgress,
			cloudhsmv2.ClusterStateInitializeInProgress,
		},
		Target:     []string{cloudhsmv2.ClusterStateUninitialized},
		Refresh:    ClusterState(conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*cloudhsmv2.Cluster); ok {
		return v, err
	}

	return nil, err
}

func HsmActive(conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Hsm, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{cloudhsmv2.HsmStateCreateInProgress},
		Target:     []string{cloudhsmv2.HsmStateActive},
		Refresh:    HsmState(conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*cloudhsmv2.Hsm); ok {
		return v, err
	}

	return nil, err
}

func HsmDeleted(conn *cloudhsmv2.CloudHSMV2, id string, timeout time.Duration) (*cloudhsmv2.Hsm, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{cloudhsmv2.HsmStateDeleteInProgress},
		Target:     []string{},
		Refresh:    HsmState(conn, id),
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*cloudhsmv2.Hsm); ok {
		return v, err
	}

	return nil, err
}
