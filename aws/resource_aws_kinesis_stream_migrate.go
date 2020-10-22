package aws

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsKinesisStreamResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"shard_count": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"retention_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  24,
			},

			"shard_level_metrics": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"encryption_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "NONE",
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsKinesisStreamStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	rawState["enforce_consumer_deletion"] = false

	return rawState, nil
}
