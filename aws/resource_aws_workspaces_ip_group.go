package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

	d.SetId(aws.StringValue(resp.GroupId))

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
			log.Printf("[WARN] WorkSpaces Ip Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	ipGroups := resp.Result

	if len(ipGroups) == 0 {
		log.Printf("[WARN] WorkSpaces Ip Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	ipGroup := ipGroups[0]

	d.Set("name", ipGroup.GroupName)
	d.Set("description", ipGroup.GroupDesc)
	d.Set("rules", flattenIpGroupRules(ipGroup.UserRules))

	tags, err := keyvaluetags.WorkspacesListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags for WorkSpaces IP Group (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsWorkspacesIpGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	if d.HasChange("rules") {
		rules := d.Get("rules").(*schema.Set).List()

		log.Printf("[INFO] Updating WorkSpaces IP Group Rules")
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
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsWorkspacesIpGroupRead(d, meta)
}

func resourceAwsWorkspacesIpGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).workspacesconn

	var found bool
	var sweeperErrs *multierror.Error
	log.Printf("[DEBUG] Finding directories associated with WorkSpaces IP Group (%s)", d.Id())
	err := conn.DescribeWorkspaceDirectoriesPages(nil, func(page *workspaces.DescribeWorkspaceDirectoriesOutput, lastPage bool) bool {
		for _, dir := range page.Directories {
			for _, ipg := range dir.IpGroupIds {
				groupID := aws.StringValue(ipg)
				if groupID == d.Id() {
					found = true
					log.Printf("[DEBUG] WorkSpaces IP Group (%s) associated with WorkSpaces Directory (%s), disassociating", groupID, aws.StringValue(dir.DirectoryId))
					_, err := conn.DisassociateIpGroups(&workspaces.DisassociateIpGroupsInput{
						DirectoryId: dir.DirectoryId,
						GroupIds:    aws.StringSlice([]string{d.Id()}),
					})
					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error disassociating WorkSpaces IP Group (%s) from WorkSpaces Directory (%s): %w", d.Id(), aws.StringValue(dir.DirectoryId), err))
						continue
					}
					log.Printf("[INFO] WorkSpaces IP Group (%s) disassociated from WorkSpaces Directory (%s)", d.Id(), aws.StringValue(dir.DirectoryId))
				}
			}
		}
		return !lastPage
	})
	if err != nil {
		return multierror.Append(sweeperErrs, fmt.Errorf("error describing WorkSpaces Directories: %w", err))
	}
	if sweeperErrs.ErrorOrNil() != nil {
		return sweeperErrs
	}

	if !found {
		log.Printf("[DEBUG] WorkSpaces IP Group (%s) not associated with any WorkSpaces Directories", d.Id())
	}

	log.Printf("[DEBUG] Deleting WorkSpaces IP Group (%s)", d.Id())
	_, err = conn.DeleteIpGroup(&workspaces.DeleteIpGroupInput{
		GroupId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error deleting WorkSpaces IP Group (%s): %w", d.Id(), err)
	}
	log.Printf("[INFO] WorkSpaces IP Group (%s) deleted", d.Id())

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
