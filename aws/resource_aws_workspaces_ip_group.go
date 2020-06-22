package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsWorkspacesIpGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWorkspacesIpGroupCreate,
		Read:   resourceAwsWorkspacesIpGroupRead,
		Update: resourceAwsWorkspacesIpGroupUpdate,
		Delete: resourceAwsWorkspacesIpGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"rules": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsWorkspacesIpGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	rules := d.Get("rules").(*schema.Set).List()

	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().WorkspacesTags()

	resp, err := conn.CreateIpGroup(&workspaces.CreateIpGroupInput{
		GroupName: aws.String(d.Get("name").(string)),
		GroupDesc: aws.String(d.Get("description").(string)),
		UserRules: expandIpGroupRules(rules),
		Tags:      tags,
	})
	if err != nil {
		return err
	}

	d.SetId(*resp.GroupId)

	return resourceAwsWorkspacesIpGroupRead(d, meta)
}

func resourceAwsWorkspacesIpGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeIpGroups(&workspaces.DescribeIpGroupsInput{
		GroupIds: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if len(resp.Result) == 0 {
			log.Printf("[WARN] Workspaces Ip Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	ipGroups := resp.Result

	if len(ipGroups) == 0 {
		log.Printf("[WARN] Workspaces Ip Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	ipGroup := ipGroups[0]

	d.Set("name", ipGroup.GroupName)
	d.Set("description", ipGroup.GroupDesc)
	d.Set("rules", flattenIpGroupRules(ipGroup.UserRules))

	tags, err := keyvaluetags.WorkspacesListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags for Workspaces IP Group (%q): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsWorkspacesIpGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	if d.HasChange("rules") {
		rules := d.Get("rules").(*schema.Set).List()

		log.Printf("[INFO] Updating Workspaces IP Group Rules")
		_, err := conn.UpdateRulesOfIpGroup(&workspaces.UpdateRulesOfIpGroupInput{
			GroupId:   aws.String(d.Id()),
			UserRules: expandIpGroupRules(rules),
		})
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.WorkspacesUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWorkspacesIpGroupRead(d, meta)
}

func resourceAwsWorkspacesIpGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	log.Printf("[INFO] Deleting Workspaces IP Group")
	_, err := conn.DeleteIpGroup(&workspaces.DeleteIpGroupInput{
		GroupId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error Deleting Workspaces IP Group: %s", err)
	}

	return nil
}

func expandIpGroupRules(rules []interface{}) []*workspaces.IpRuleItem {
	var result []*workspaces.IpRuleItem
	for _, rule := range rules {
		r := rule.(map[string]interface{})

		result = append(result, &workspaces.IpRuleItem{
			IpRule:   aws.String(r["source"].(string)),
			RuleDesc: aws.String(r["description"].(string)),
		})
	}
	return result
}

func flattenIpGroupRules(rules []*workspaces.IpRuleItem) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		r := map[string]interface{}{
			"source": *rule.IpRule,
		}
		if rule.RuleDesc != nil {
			r["description"] = *rule.RuleDesc
		}
		result = append(result, r)
	}
	return result
}
