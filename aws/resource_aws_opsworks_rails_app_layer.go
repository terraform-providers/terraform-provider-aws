package aws

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsOpsworksRailsAppLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeRailsApp,
		DefaultLayerName: "Rails App Server",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"ruby_version": {
				AttrName: opsworks.LayerAttributesKeysRubyVersion,
				Type:     schema.TypeString,
				Default:  "2.0.0",
			},
			"app_server": {
				AttrName: opsworks.LayerAttributesKeysRailsStack,
				Type:     schema.TypeString,
				Default:  "apache_passenger",
			},
			"passenger_version": {
				AttrName: opsworks.LayerAttributesKeysPassengerVersion,
				Type:     schema.TypeString,
				Default:  "4.0.46",
			},
			"rubygems_version": {
				AttrName: opsworks.LayerAttributesKeysRubygemsVersion,
				Type:     schema.TypeString,
				Default:  "2.2.2",
			},
			"manage_bundler": {
				AttrName: opsworks.LayerAttributesKeysManageBundler,
				Type:     schema.TypeBool,
				Default:  true,
			},
			"bundler_version": {
				AttrName: opsworks.LayerAttributesKeysBundlerVersion,
				Type:     schema.TypeString,
				Default:  "1.5.3",
			},
		},
	}

	return layerType.SchemaResource()
}
