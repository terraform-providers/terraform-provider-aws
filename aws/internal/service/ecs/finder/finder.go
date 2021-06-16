package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func ClusterByARN(conn *ecs.ECS, arn string) (*ecs.DescribeClustersOutput, error) {
	input := &ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(arn)},
		Include:  []*string{aws.String(ecs.ClusterFieldTags), aws.String(ecs.ClusterFieldConfigurations)},
	}

	output, err := conn.DescribeClusters(input)

	if err != nil {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}
