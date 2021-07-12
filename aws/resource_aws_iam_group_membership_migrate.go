package aws

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsIamGroupMembershipV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"users": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsIamGroupMembershipStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		rawState = map[string]interface{}{}
	}

	if group, ok := rawState["group"]; ok {
		rawState["id"] = group
	}
	delete(rawState, "name")
	return rawState, nil
}
