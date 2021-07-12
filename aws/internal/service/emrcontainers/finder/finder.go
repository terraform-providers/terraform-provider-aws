package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
)

// VirtualClusterById returns the EMR containers virtual cluster corresponding to the specified Id.
// Returns nil if no environment is found.
func VirtualClusterById(conn *emrcontainers.EMRContainers, id string) (*emrcontainers.VirtualCluster, error) {
	input := &emrcontainers.DescribeVirtualClusterInput{
		Id: aws.String(id),
	}

	output, err := conn.DescribeVirtualCluster(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.VirtualCluster, nil
}
