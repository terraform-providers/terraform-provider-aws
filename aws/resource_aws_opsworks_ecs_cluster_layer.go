package aws

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsOpsworksECSClusterLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeEcsCluster,
		DefaultLayerName: "Ecs Cluster",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"ecs_cluster_arn": {
				AttrName: opsworks.LayerAttributesKeysEcsClusterArn,
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}

	return layerType.SchemaResource()
}
