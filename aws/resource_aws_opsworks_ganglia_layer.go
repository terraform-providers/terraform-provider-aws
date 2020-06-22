package aws

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsOpsworksGangliaLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeMonitoringMaster,
		DefaultLayerName: "Ganglia",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"url": {
				AttrName: opsworks.LayerAttributesKeysGangliaUrl,
				Type:     schema.TypeString,
				Default:  "/ganglia",
			},
			"username": {
				AttrName: opsworks.LayerAttributesKeysGangliaUser,
				Type:     schema.TypeString,
				Default:  "opsworks",
			},
			"password": {
				AttrName:  opsworks.LayerAttributesKeysGangliaPassword,
				Type:      schema.TypeString,
				Required:  true,
				WriteOnly: true,
			},
		},
	}

	return layerType.SchemaResource()
}
