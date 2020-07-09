package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func ClusterStatus(conn *kafka.Kafka, clusterArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &kafka.DescribeClusterInput{
			ClusterArn: aws.String(clusterArn),
		}

		output, err := conn.DescribeCluster(input)

		if err != nil {
			return nil, kafka.ClusterStateFailed, err
		}

		if output == nil || output.ClusterInfo == nil {
			return output, kafka.ClusterStateFailed, nil
		}

		return output.ClusterInfo, aws.StringValue(output.ClusterInfo.State), nil
	}
}
