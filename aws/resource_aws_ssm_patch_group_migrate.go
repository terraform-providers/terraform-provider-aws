package aws

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSsmPatchGroupV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"baseline_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"patch_group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsSsmPatchGroupStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		rawState = map[string]interface{}{}
	}

	rawState["id"] = fmt.Sprintf("%s,%s", rawState["patch_group"], rawState["baseline_id"])

	return rawState, nil
}
