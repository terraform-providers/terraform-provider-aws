package aws

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsOpsworksHaproxyLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeLb,
		DefaultLayerName: "HAProxy",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"stats_enabled": {
				AttrName: opsworks.LayerAttributesKeysEnableHaproxyStats,
				Type:     schema.TypeBool,
				Default:  true,
			},
			"stats_url": {
				AttrName: opsworks.LayerAttributesKeysHaproxyStatsUrl,
				Type:     schema.TypeString,
				Default:  "/haproxy?stats",
			},
			"stats_user": {
				AttrName: opsworks.LayerAttributesKeysHaproxyStatsUser,
				Type:     schema.TypeString,
				Default:  "opsworks",
			},
			"stats_password": {
				AttrName:  opsworks.LayerAttributesKeysHaproxyStatsPassword,
				Type:      schema.TypeString,
				WriteOnly: true,
				Required:  true,
			},
			"healthcheck_url": {
				AttrName: opsworks.LayerAttributesKeysHaproxyHealthCheckUrl,
				Type:     schema.TypeString,
				Default:  "/",
			},
			"healthcheck_method": {
				AttrName: opsworks.LayerAttributesKeysHaproxyHealthCheckMethod,
				Type:     schema.TypeString,
				Default:  "OPTIONS",
			},
		},
	}

	return layerType.SchemaResource()
}
