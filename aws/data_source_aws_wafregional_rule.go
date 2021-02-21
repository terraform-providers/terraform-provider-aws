package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsWafRegionalRule() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWafRegionalRuleRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsWafRegionalRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	name := d.Get("name").(string)

	rules := make([]*waf.RuleSummary, 0)
	// ListRulesInput does not have a name parameter for filtering
	input := &waf.ListRulesInput{}
	for {
		output, err := conn.ListRules(input)
		if err != nil {
			return fmt.Errorf("error reading WAF Rule: %w", err)
		}
		for _, rule := range output.Rules {
			if aws.StringValue(rule.Name) == name {
				rules = append(rules, rule)
			}
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if len(rules) == 0 {
		return fmt.Errorf("WAF Rule not found for name: %s", name)
	}

	if len(rules) > 1 {
		return fmt.Errorf("multiple WAF Rules found for name: %s", name)
	}

	rule := rules[0]

	d.SetId(aws.StringValue(rule.RuleId))

	return nil
}
