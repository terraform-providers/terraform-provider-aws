package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ssm/finder"
)

func resourceAwsSsmPatchGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmPatchGroupCreate,
		Read:   resourceAwsSsmPatchGroupRead,
		Delete: resourceAwsSsmPatchGroupDelete,

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceAwsSsmPatchGroupV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceAwsSsmPatchGroupStateUpgradeV0,
				Version: 0,
			},
		},

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

func resourceAwsSsmPatchGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	baselineId := d.Get("baseline_id").(string)
	patchGroup := d.Get("patch_group").(string)

	params := &ssm.RegisterPatchBaselineForPatchGroupInput{
		BaselineId: aws.String(baselineId),
		PatchGroup: aws.String(patchGroup),
	}

	resp, err := conn.RegisterPatchBaselineForPatchGroup(params)
	if err != nil {
		return fmt.Errorf("error registering SSM Patch Baseline (%s) for Patch Group (%s): %w", baselineId, patchGroup, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", aws.StringValue(resp.PatchGroup), aws.StringValue(resp.BaselineId)))

	return resourceAwsSsmPatchGroupRead(d, meta)
}

func resourceAwsSsmPatchGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	patchGroup, baselineId, err := parseSsmPatchGroupId(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSM Patch Group ID (%s): %w", d.Id(), err)
	}

	group, err := finder.PatchGroup(conn, patchGroup, baselineId)

	if err != nil {
		return fmt.Errorf("error reading SSM Patch Group (%s): %w", d.Id(), err)
	}

	if group == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading SSM Patch Group (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] SSM Patch Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var groupBaselineId string
	if group.BaselineIdentity != nil {
		groupBaselineId = aws.StringValue(group.BaselineIdentity.BaselineId)
	}

	d.Set("baseline_id", groupBaselineId)
	d.Set("patch_group", group.PatchGroup)

	return nil

}

func resourceAwsSsmPatchGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn

	patchGroup, baselineId, err := parseSsmPatchGroupId(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSM Patch Group ID (%s): %w", d.Id(), err)
	}

	params := &ssm.DeregisterPatchBaselineForPatchGroupInput{
		BaselineId: aws.String(baselineId),
		PatchGroup: aws.String(patchGroup),
	}

	_, err = conn.DeregisterPatchBaselineForPatchGroup(params)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ssm.ErrCodeDoesNotExistException) {
			return nil
		}
		return fmt.Errorf("error deregistering SSM Patch Baseline (%s) for Patch Group (%s): %w", baselineId, patchGroup, err)
	}

	return nil
}

func parseSsmPatchGroupId(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("please make sure ID is in format PATCH_GROUP,BASELINE_ID")
	}

	return parts[0], parts[1], nil
}
