package aws

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsOpsworksJavaAppLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeJavaApp,
		DefaultLayerName: "Java App Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"jvm_type": {
				AttrName: opsworks.LayerAttributesKeysJvm,
				Type:     schema.TypeString,
				Default:  "openjdk",
			},
			"jvm_version": {
				AttrName: opsworks.LayerAttributesKeysJvmVersion,
				Type:     schema.TypeString,
				Default:  "7",
			},
			"jvm_options": {
				AttrName: opsworks.LayerAttributesKeysJvmOptions,
				Type:     schema.TypeString,
				Default:  "",
			},
			"app_server": {
				AttrName: opsworks.LayerAttributesKeysJavaAppServer,
				Type:     schema.TypeString,
				Default:  "tomcat",
			},
			"app_server_version": {
				AttrName: opsworks.LayerAttributesKeysJavaAppServerVersion,
				Type:     schema.TypeString,
				Default:  "7",
			},
		},
	}

	return layerType.SchemaResource()
}
