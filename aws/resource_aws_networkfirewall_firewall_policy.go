package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/networkfirewall/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/networkfirewall/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsNetworkFirewallFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsNetworkFirewallFirewallPolicyCreate,
		ReadContext:   resourceAwsNetworkFirewallFirewallPolicyRead,
		UpdateContext: resourceAwsNetworkFirewallFirewallPolicyUpdate,
		DeleteContext: resourceAwsNetworkFirewallFirewallPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"firewall_policy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stateful_rule_group_reference": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
						"stateless_custom_action": customActionSchema(),
						"stateless_default_actions": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"stateless_fragment_default_actions": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"stateless_rule_group_reference": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"priority": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"resource_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsNetworkFirewallFirewallPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn
	name := d.Get("name").(string)
	input := &networkfirewall.CreateFirewallPolicyInput{
		FirewallPolicy:     expandNetworkFirewallFirewallPolicy(d.Get("firewall_policy").([]interface{})),
		FirewallPolicyName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().NetworkfirewallTags()
	}

	log.Printf("[DEBUG] Creating NetworkFirewall Firewall Policy %s", name)

	output, err := conn.CreateFirewallPolicyWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating NetworkFirewall Firewall Policy (%s): %w", name, err))
	}
	if output == nil || output.FirewallPolicyResponse == nil {
		return diag.FromErr(fmt.Errorf("error creating NetworkFirewall Firewall Policy (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.FirewallPolicyResponse.FirewallPolicyArn))

	return resourceAwsNetworkFirewallFirewallPolicyRead(ctx, d, meta)
}

func resourceAwsNetworkFirewallFirewallPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading NetworkFirewall Firewall Policy %s", d.Id())

	output, err := finder.FirewallPolicy(ctx, conn, d.Id())
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] NetworkFirewall Firewall Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s): %w", d.Id(), err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s): empty output", d.Id()))
	}
	if output.FirewallPolicyResponse == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s): empty output.FirewallPolicyResponse", d.Id()))
	}

	resp := output.FirewallPolicyResponse
	policy := output.FirewallPolicy

	d.Set("arn", resp.FirewallPolicyArn)
	d.Set("description", resp.Description)
	d.Set("name", resp.FirewallPolicyName)
	d.Set("update_token", output.UpdateToken)

	if err := d.Set("firewall_policy", flattenNetworkFirewallFirewallPolicy(policy)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting firewall_policy: %w", err))
	}

	if err := d.Set("tags", keyvaluetags.NetworkfirewallKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	return nil
}

func resourceAwsNetworkFirewallFirewallPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn
	arn := d.Id()

	log.Printf("[DEBUG] Updating NetworkFirewall Firewall Policy %s", arn)

	if d.HasChanges("description", "firewall_policy") {
		input := &networkfirewall.UpdateFirewallPolicyInput{
			FirewallPolicy:    expandNetworkFirewallFirewallPolicy(d.Get("firewall_policy").([]interface{})),
			FirewallPolicyArn: aws.String(arn),
			UpdateToken:       aws.String(d.Get("update_token").(string)),
		}
		// Only pass non-empty description values, else API request returns an InternalServiceError
		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}
		_, err := conn.UpdateFirewallPolicyWithContext(ctx, input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall Policy (%s) firewall_policy: %w", arn, err))
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.NetworkfirewallUpdateTags(conn, arn, o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall Policy (%s) tags: %w", arn, err))
		}
	}

	return resourceAwsNetworkFirewallFirewallPolicyRead(ctx, d, meta)
}

func resourceAwsNetworkFirewallFirewallPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn

	log.Printf("[DEBUG] Deleting NetworkFirewall Firewall Policy %s", d.Id())

	input := &networkfirewall.DeleteFirewallPolicyInput{
		FirewallPolicyArn: aws.String(d.Id()),
	}

	err := resource.RetryContext(ctx, waiter.FirewallPolicyTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.DeleteFirewallPolicyWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, networkfirewall.ErrCodeInvalidOperationException, "Unable to delete the object because it is still in use") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteFirewallPolicyWithContext(ctx, input)
	}

	if err != nil {
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting NetworkFirewall Firewall Policy (%s): %w", d.Id(), err))
	}

	if _, err := waiter.FirewallPolicyDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error waiting for NetworkFirewall Firewall Policy (%s) to delete: %w", d.Id(), err))
	}

	return nil
}

func expandNetworkFirewallStatefulRuleGroupReferences(l []interface{}) []*networkfirewall.StatefulRuleGroupReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	references := make([]*networkfirewall.StatefulRuleGroupReference, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		reference := &networkfirewall.StatefulRuleGroupReference{}
		if v, ok := tfMap["resource_arn"].(string); ok && v != "" {
			reference.ResourceArn = aws.String(v)
		}
		references = append(references, reference)
	}
	return references
}

func expandNetworkFirewallStatelessRuleGroupReferences(l []interface{}) []*networkfirewall.StatelessRuleGroupReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	references := make([]*networkfirewall.StatelessRuleGroupReference, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		reference := &networkfirewall.StatelessRuleGroupReference{}
		if v, ok := tfMap["priority"].(int); ok && v > 0 {
			reference.Priority = aws.Int64(int64(v))
		}
		if v, ok := tfMap["resource_arn"].(string); ok && v != "" {
			reference.ResourceArn = aws.String(v)
		}
		references = append(references, reference)
	}
	return references
}

func expandNetworkFirewallFirewallPolicy(l []interface{}) *networkfirewall.FirewallPolicy {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	lRaw := l[0].(map[string]interface{})
	policy := &networkfirewall.FirewallPolicy{
		StatelessDefaultActions:         expandStringSet(lRaw["stateless_default_actions"].(*schema.Set)),
		StatelessFragmentDefaultActions: expandStringSet(lRaw["stateless_fragment_default_actions"].(*schema.Set)),
	}

	if v, ok := lRaw["stateful_rule_group_reference"].(*schema.Set); ok && v.Len() > 0 {
		policy.StatefulRuleGroupReferences = expandNetworkFirewallStatefulRuleGroupReferences(v.List())
	}

	if v, ok := lRaw["stateless_custom_action"].(*schema.Set); ok && v.Len() > 0 {
		policy.StatelessCustomActions = expandNetworkFirewallCustomActions(v.List())
	}

	if v, ok := lRaw["stateless_rule_group_reference"].(*schema.Set); ok && v.Len() > 0 {
		policy.StatelessRuleGroupReferences = expandNetworkFirewallStatelessRuleGroupReferences(v.List())
	}

	return policy
}

func flattenNetworkFirewallFirewallPolicy(policy *networkfirewall.FirewallPolicy) []interface{} {
	if policy == nil {
		return []interface{}{}
	}
	p := map[string]interface{}{}
	if policy.StatefulRuleGroupReferences != nil {
		p["stateful_rule_group_reference"] = flattenNetworkFirewallPolicyStatefulRuleGroupReference(policy.StatefulRuleGroupReferences)
	}
	if policy.StatelessCustomActions != nil {
		p["stateless_custom_action"] = flattenNetworkFirewallCustomActions(policy.StatelessCustomActions)
	}
	if policy.StatelessDefaultActions != nil {
		p["stateless_default_actions"] = flattenStringSet(policy.StatelessDefaultActions)
	}
	if policy.StatelessFragmentDefaultActions != nil {
		p["stateless_fragment_default_actions"] = flattenStringSet(policy.StatelessFragmentDefaultActions)
	}
	if policy.StatelessRuleGroupReferences != nil {
		p["stateless_rule_group_reference"] = flattenNetworkFirewallPolicyStatelessRuleGroupReference(policy.StatelessRuleGroupReferences)
	}

	return []interface{}{p}
}

func flattenNetworkFirewallPolicyStatefulRuleGroupReference(l []*networkfirewall.StatefulRuleGroupReference) []interface{} {
	references := make([]interface{}, 0, len(l))
	for _, ref := range l {
		reference := map[string]interface{}{
			"resource_arn": aws.StringValue(ref.ResourceArn),
		}
		references = append(references, reference)
	}

	return references
}

func flattenNetworkFirewallPolicyStatelessRuleGroupReference(l []*networkfirewall.StatelessRuleGroupReference) []interface{} {
	references := make([]interface{}, 0, len(l))
	for _, ref := range l {
		reference := map[string]interface{}{
			"priority":     int(aws.Int64Value(ref.Priority)),
			"resource_arn": aws.StringValue(ref.ResourceArn),
		}
		references = append(references, reference)
	}
	return references
}
