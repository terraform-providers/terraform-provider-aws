package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsWafSubscribedRuleGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWafSubscribedRuleGroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"metric_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsWafSubscribedRuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn
	name, nameOk := d.GetOk("name")
	metricName, metricNameOk := d.GetOk("metric_name")

	rules := make([]*waf.SubscribedRuleGroupSummary, 0)
	// ListSubscribedRuleGroupsInput does not have a name parameter for filtering
	input := &waf.ListSubscribedRuleGroupsInput{}
	for {
		output, err := conn.ListSubscribedRuleGroups(input)
		if err != nil {
			return fmt.Errorf("error reading WAF Rules Groups: %s", err)
		}
		for _, rule := range output.RuleGroups {
			if nameOk && aws.StringValue(rule.Name) != name {
				continue
			}
			if metricNameOk && aws.StringValue(rule.MetricName) != metricName {
				continue
			}

			rules = append(rules, rule)
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if len(rules) == 0 {
		return fmt.Errorf("WAF Subscribed Rule Group not found for name %s and metricName %s", name, metricName)
	}

	if len(rules) > 1 {
		return fmt.Errorf("multiple WAF Rule Groups found for name %s and metricName %s", name, metricName)
	}

	rule := rules[0]

	d.SetId(aws.StringValue(rule.RuleGroupId))
	d.Set("metric_name", aws.StringValue(rule.MetricName))
	d.Set("name", aws.StringValue(rule.Name))

	return nil
}
