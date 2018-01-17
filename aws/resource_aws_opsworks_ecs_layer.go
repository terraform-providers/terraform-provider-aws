package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsOpsworksECSClusterLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         "ecs-cluster",
		DefaultLayerName: "Ecs Cluster",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"ecs_cluster_arn": {
				AttrName: "EcsClusterArn",
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}

	return layerType.SchemaResource()
}
