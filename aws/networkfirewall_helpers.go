package aws

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func customActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"action_definition": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"publish_metric_action": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"dimension": {
											Type:     schema.TypeSet,
											Required: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"value": {
														Type:     schema.TypeString,
														Required: true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"action_name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), "must contain only alphanumeric characters"),
				},
			},
		},
	}
}

func expandNetworkFirewallCustomActions(l []interface{}) []*networkfirewall.CustomAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	customActions := make([]*networkfirewall.CustomAction, 0, len(l))
	for _, tfMapRaw := range l {
		customAction := &networkfirewall.CustomAction{}
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		if v, ok := tfMap["action_definition"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			customAction.ActionDefinition = expandNetworkFirewallActionDefinition(v)
		}
		if v, ok := tfMap["action_name"].(string); ok && v != "" {
			customAction.ActionName = aws.String(v)
		}
		customActions = append(customActions, customAction)
	}

	return customActions
}

func expandNetworkFirewallActionDefinition(l []interface{}) *networkfirewall.ActionDefinition {
	if l == nil || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}
	customAction := &networkfirewall.ActionDefinition{}

	if v, ok := tfMap["publish_metric_action"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		customAction.PublishMetricAction = expandNetworkFirewallCustomActionPublishMetricAction(v)
	}

	return customAction
}

func expandNetworkFirewallCustomActionPublishMetricAction(l []interface{}) *networkfirewall.PublishMetricAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}
	action := &networkfirewall.PublishMetricAction{}
	if tfSet, ok := tfMap["dimension"].(*schema.Set); ok && tfSet.Len() > 0 {
		tfList := tfSet.List()
		dimensions := make([]*networkfirewall.Dimension, 0, len(tfList))
		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}
			dimension := &networkfirewall.Dimension{
				Value: aws.String(tfMap["value"].(string)),
			}
			dimensions = append(dimensions, dimension)
		}
		action.Dimensions = dimensions
	}
	return action
}

func flattenNetworkFirewallCustomActions(c []*networkfirewall.CustomAction) []interface{} {
	if c == nil {
		return []interface{}{}
	}

	customActions := make([]interface{}, 0, len(c))
	for _, elem := range c {
		m := map[string]interface{}{
			"action_definition": flattenNetworkFirewallActionDefinition(elem.ActionDefinition),
			"action_name":       aws.StringValue(elem.ActionName),
		}
		customActions = append(customActions, m)
	}

	return customActions
}

func flattenNetworkFirewallActionDefinition(v *networkfirewall.ActionDefinition) []interface{} {
	if v == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"publish_metric_action": flattenNetworkFirewallPublishMetricAction(v.PublishMetricAction),
	}
	return []interface{}{m}
}

func flattenNetworkFirewallPublishMetricAction(m *networkfirewall.PublishMetricAction) []interface{} {
	if m == nil {
		return []interface{}{}
	}

	metrics := map[string]interface{}{
		"dimension": flattenNetworkFirewallDimensions(m.Dimensions),
	}

	return []interface{}{metrics}
}

func flattenNetworkFirewallDimensions(d []*networkfirewall.Dimension) []interface{} {
	dimensions := make([]interface{}, 0, len(d))
	for _, v := range d {
		dimension := map[string]interface{}{
			"value": aws.StringValue(v.Value),
		}
		dimensions = append(dimensions, dimension)
	}

	return dimensions
}
