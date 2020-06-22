package aws

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsOpsworksMemcachedLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeMemcached,
		DefaultLayerName: "Memcached",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"allocated_memory": {
				AttrName: opsworks.LayerAttributesKeysMemcachedMemory,
				Type:     schema.TypeInt,
				Default:  512,
			},
		},
	}

	return layerType.SchemaResource()
}
