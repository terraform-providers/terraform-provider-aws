package aws

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsOpsworksMysqlLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeDbMaster,
		DefaultLayerName: "MySQL",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"root_password": {
				AttrName:  opsworks.LayerAttributesKeysMysqlRootPassword,
				Type:      schema.TypeString,
				WriteOnly: true,
			},
			"root_password_on_all_instances": {
				AttrName: opsworks.LayerAttributesKeysMysqlRootPasswordUbiquitous,
				Type:     schema.TypeBool,
				Default:  true,
			},
		},
	}

	return layerType.SchemaResource()
}
