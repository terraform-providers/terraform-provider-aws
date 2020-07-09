package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	ClusterActiveTimeout  = 60 * time.Minute
	ClusterDeletedTimeout = 60 * time.Minute
)

func ClusterActive(conn *kafka.Kafka, clusterArn string) (*kafka.DescribeClusterOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ClusterStateCreating},
		Target:  []string{kafka.ClusterStateActive},
		Refresh: ClusterStatus(conn, clusterArn),
		Timeout: ClusterActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kafka.DescribeClusterOutput); ok {
		return v, err
	}

	return nil, err
}

func ClusterDeleted(conn *kafka.Kafka, clusterArn string) (*kafka.DescribeClusterOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafka.ClusterStateDeleting},
		Target:  []string{kafka.ClusterStateFailed},
		Refresh: ClusterStatus(conn, clusterArn),
		Timeout: ClusterDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kafka.DescribeClusterOutput); ok {
		return v, err
	}

	return nil, err
}
