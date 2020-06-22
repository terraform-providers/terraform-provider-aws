package aws

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsOpsworksNodejsAppLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeNodejsApp,
		DefaultLayerName: "Node.js App Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"nodejs_version": {
				AttrName: opsworks.LayerAttributesKeysNodejsVersion,
				Type:     schema.TypeString,
				Default:  "0.10.38",
			},
		},
	}

	return layerType.SchemaResource()
}
