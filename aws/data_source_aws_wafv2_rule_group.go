package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsWafv2RuleGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWafv2RuleGroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
		},
	}
}

func dataSourceAwsWafv2RuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	name := d.Get("name").(string)

	var foundRuleGroup *wafv2.RuleGroupSummary
	input := &wafv2.ListRuleGroupsInput{
		Scope: aws.String(d.Get("scope").(string)),
		Limit: aws.Int64(100),
	}

	for {
		resp, err := conn.ListRuleGroups(input)
		if err != nil {
			return fmt.Errorf("Error reading WAFv2 RuleGroups: %s", err)
		}

		if resp == nil || resp.RuleGroups == nil {
			return fmt.Errorf("Error reading WAFv2 RuleGroups")
		}

		for _, ruleGroup := range resp.RuleGroups {
			if aws.StringValue(ruleGroup.Name) == name {
				foundRuleGroup = ruleGroup
				break
			}
		}

		if resp.NextMarker == nil || foundRuleGroup != nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundRuleGroup == nil {
		return fmt.Errorf("WAFv2 RuleGroup not found for name: %s", name)
	}

	d.SetId(aws.StringValue(foundRuleGroup.Id))
	d.Set("arn", aws.StringValue(foundRuleGroup.ARN))
	d.Set("description", aws.StringValue(foundRuleGroup.Description))

	return nil
}
