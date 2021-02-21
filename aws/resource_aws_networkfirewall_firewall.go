package aws

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/networkfirewall/waiter"
)

func resourceAwsNetworkFirewallFirewall() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsNetworkFirewallFirewallCreate,
		ReadContext:   resourceAwsNetworkFirewallFirewallRead,
		UpdateContext: resourceAwsNetworkFirewallFirewallUpdate,
		DeleteContext: resourceAwsNetworkFirewallFirewallDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("firewall_status", func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("subnet_mapping")
			}),
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delete_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"firewall_policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"firewall_policy_change_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"firewall_status": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sync_states": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"attachment": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"endpoint_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"subnet_id": {
													Type:     schema.TypeString,
													Computed: true,
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_change_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"subnet_mapping": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
			"update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsNetworkFirewallFirewallCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn
	name := d.Get("name").(string)
	input := &networkfirewall.CreateFirewallInput{
		FirewallName:      aws.String(name),
		FirewallPolicyArn: aws.String(d.Get("firewall_policy_arn").(string)),
		SubnetMappings:    expandNetworkFirewallSubnetMappings(d.Get("subnet_mapping").(*schema.Set).List()),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	if v, ok := d.GetOk("delete_protection"); ok {
		input.DeleteProtection = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("firewall_policy_change_protection"); ok {
		input.FirewallPolicyChangeProtection = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("subnet_change_protection"); ok {
		input.SubnetChangeProtection = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().NetworkfirewallTags()
	}

	log.Printf("[DEBUG] Creating NetworkFirewall Firewall %s", name)

	output, err := conn.CreateFirewallWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating NetworkFirewall Firewall (%s): %w", name, err))
	}

	d.SetId(aws.StringValue(output.Firewall.FirewallArn))

	if _, err := waiter.FirewallCreated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for NetworkFirewall Firewall (%s) to be created: %w", d.Id(), err))
	}

	return resourceAwsNetworkFirewallFirewallRead(ctx, d, meta)
}

func resourceAwsNetworkFirewallFirewallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading NetworkFirewall Firewall %s", d.Id())

	input := &networkfirewall.DescribeFirewallInput{
		FirewallArn: aws.String(d.Id()),
	}
	output, err := conn.DescribeFirewallWithContext(ctx, input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] NetworkFirewall Firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall (%s): %w", d.Id(), err))
	}

	if output == nil || output.Firewall == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall (%s): empty output", d.Id()))
	}

	firewall := output.Firewall

	d.Set("arn", firewall.FirewallArn)
	d.Set("delete_protection", firewall.DeleteProtection)
	d.Set("description", firewall.Description)
	d.Set("name", firewall.FirewallName)
	d.Set("firewall_policy_arn", firewall.FirewallPolicyArn)
	d.Set("firewall_policy_change_protection", firewall.FirewallPolicyChangeProtection)
	d.Set("firewall_status", flattenNetworkFirewallFirewallStatus(output.FirewallStatus))
	d.Set("subnet_change_protection", firewall.SubnetChangeProtection)
	d.Set("update_token", output.UpdateToken)
	d.Set("vpc_id", firewall.VpcId)

	if err := d.Set("subnet_mapping", flattenNetworkFirewallSubnetMappings(firewall.SubnetMappings)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting subnet_mappings: %w", err))
	}

	if err := d.Set("tags", keyvaluetags.NetworkfirewallKeyValueTags(firewall.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	return nil
}

func resourceAwsNetworkFirewallFirewallUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn
	arn := d.Id()
	updateToken := aws.String(d.Get("update_token").(string))

	if d.HasChange("description") {
		input := &networkfirewall.UpdateFirewallDescriptionInput{
			Description: aws.String(d.Get("description").(string)),
			FirewallArn: aws.String(arn),
			UpdateToken: updateToken,
		}
		resp, err := conn.UpdateFirewallDescriptionWithContext(ctx, input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) description: %w", d.Id(), err))
		}
		if resp == nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) description: empty update_token", arn))
		}
		updateToken = resp.UpdateToken
	}

	if d.HasChange("delete_protection") {
		input := &networkfirewall.UpdateFirewallDeleteProtectionInput{
			DeleteProtection: aws.Bool(d.Get("delete_protection").(bool)),
			FirewallArn:      aws.String(arn),
			UpdateToken:      updateToken,
		}
		resp, err := conn.UpdateFirewallDeleteProtectionWithContext(ctx, input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) delete_protection: %w", arn, err))
		}
		if resp == nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) delete_protection: empty update_token", arn))
		}
		updateToken = resp.UpdateToken
	}

	// Note: The *_change_protection fields below are handled before their respective fields
	// to account for disabling and subsequent changes

	if d.HasChange("firewall_policy_change_protection") {
		input := &networkfirewall.UpdateFirewallPolicyChangeProtectionInput{
			FirewallArn:                    aws.String(arn),
			FirewallPolicyChangeProtection: aws.Bool(d.Get("firewall_policy_change_protection").(bool)),
			UpdateToken:                    updateToken,
		}
		resp, err := conn.UpdateFirewallPolicyChangeProtection(input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) firewall_policy_change_protection: %w", arn, err))
		}
		if resp == nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) firewall_policy_change_protection: empty update_token", arn))
		}
		updateToken = resp.UpdateToken
	}

	if d.HasChange("firewall_policy_arn") {
		input := &networkfirewall.AssociateFirewallPolicyInput{
			FirewallArn:       aws.String(arn),
			FirewallPolicyArn: aws.String(d.Get("firewall_policy_arn").(string)),
			UpdateToken:       updateToken,
		}
		resp, err := conn.AssociateFirewallPolicy(input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) firewall_policy_arn: %w", arn, err))
		}
		if resp == nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) firewall_policy_arn: empty update_token", arn))
		}
		updateToken = resp.UpdateToken
	}

	if d.HasChange("subnet_change_protection") {
		input := &networkfirewall.UpdateSubnetChangeProtectionInput{
			FirewallArn:            aws.String(arn),
			SubnetChangeProtection: aws.Bool(d.Get("subnet_change_protection").(bool)),
			UpdateToken:            updateToken,
		}
		resp, err := conn.UpdateSubnetChangeProtection(input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) subnet_change_protection: %w", arn, err))
		}
		if resp == nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) subnet_change_protection: empty update_token", arn))
		}
		updateToken = resp.UpdateToken
	}

	if d.HasChange("subnet_mapping") {
		o, n := d.GetChange("subnet_mapping")
		subnetsToRemove, subnetsToAdd := networkFirewallSubnetMappingsDiff(o.(*schema.Set), n.(*schema.Set))
		// Ensure we add before removing a SubnetMapping if there is only 1
		if len(subnetsToAdd) > 0 {
			input := &networkfirewall.AssociateSubnetsInput{
				FirewallArn:    aws.String(arn),
				SubnetMappings: subnetsToAdd,
				UpdateToken:    updateToken,
			}

			_, err := conn.AssociateSubnetsWithContext(ctx, input)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error associating NetworkFirewall Firewall (%s) subnet: %w", arn, err))
			}

			respToken, err := waiter.FirewallUpdated(ctx, conn, arn)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error waiting for NetworkFirewall Firewall (%s) to be updated: %w", d.Id(), err))

			}
			if respToken == nil {
				return diag.FromErr(fmt.Errorf("error associating NetworkFirewall Firewall (%s) subnet: empty update_token", arn))
			}

			updateToken = respToken
		}
		if len(subnetsToRemove) > 0 {
			input := &networkfirewall.DisassociateSubnetsInput{
				FirewallArn: aws.String(arn),
				SubnetIds:   aws.StringSlice(subnetsToRemove),
				UpdateToken: updateToken,
			}

			_, err := conn.DisassociateSubnetsWithContext(ctx, input)
			if err != nil && !tfawserr.ErrMessageContains(err, networkfirewall.ErrCodeInvalidRequestException, "inaccessible") {
				return diag.FromErr(fmt.Errorf("error disassociating NetworkFirewall Firewall (%s) subnet: %w", arn, err))
			}

			_, err = waiter.FirewallUpdated(ctx, conn, arn)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error waiting for NetworkFirewall Firewall (%s) to be updated: %w", d.Id(), err))

			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.NetworkfirewallUpdateTags(conn, arn, o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall (%s) tags: %w", arn, err))
		}
	}

	return resourceAwsNetworkFirewallFirewallRead(ctx, d, meta)
}

func resourceAwsNetworkFirewallFirewallDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn

	log.Printf("[DEBUG] Deleting NetworkFirewall Firewall %s", d.Id())

	input := &networkfirewall.DeleteFirewallInput{
		FirewallArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteFirewallWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting NetworkFirewall Firewall (%s): %w", d.Id(), err))
	}

	if _, err := waiter.FirewallDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error waiting for NetworkFirewall Firewall (%s) to delete: %w", d.Id(), err))
	}

	return nil
}

func expandNetworkFirewallSubnetMappings(l []interface{}) []*networkfirewall.SubnetMapping {
	mappings := make([]*networkfirewall.SubnetMapping, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		mapping := &networkfirewall.SubnetMapping{
			SubnetId: aws.String(tfMap["subnet_id"].(string)),
		}
		mappings = append(mappings, mapping)
	}

	return mappings
}

func expandNetworkFirewallSubnetMappingIds(l []interface{}) []string {
	var ids []string
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := tfMap["subnet_id"].(string); ok && id != "" {
			ids = append(ids, id)
		}
	}

	return ids
}

func flattenNetworkFirewallFirewallStatus(status *networkfirewall.FirewallStatus) []interface{} {
	if status == nil {
		return nil
	}

	m := map[string]interface{}{
		"sync_states": flattenNetworkFirewallSyncStates(status.SyncStates),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallSyncStates(s map[string]*networkfirewall.SyncState) []interface{} {
	if s == nil {
		return nil
	}

	syncStates := make([]interface{}, 0, len(s))
	for k, v := range s {
		m := map[string]interface{}{
			"availability_zone": k,
			"attachment":        flattenNetworkFirewallSyncStateAttachment(v.Attachment),
		}
		syncStates = append(syncStates, m)
	}

	return syncStates
}

func flattenNetworkFirewallSyncStateAttachment(a *networkfirewall.Attachment) []interface{} {
	if a == nil {
		return nil
	}

	m := map[string]interface{}{
		"endpoint_id": aws.StringValue(a.EndpointId),
		"subnet_id":   aws.StringValue(a.SubnetId),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallSubnetMappings(sm []*networkfirewall.SubnetMapping) []interface{} {
	mappings := make([]interface{}, 0, len(sm))
	for _, s := range sm {
		m := map[string]interface{}{
			"subnet_id": aws.StringValue(s.SubnetId),
		}
		mappings = append(mappings, m)
	}

	return mappings
}

func networkFirewallSubnetMappingsHash(v interface{}) int {
	var buf bytes.Buffer

	tfMap, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}
	if id, ok := tfMap["subnet_id"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", id))
	}

	return hashcode.String(buf.String())
}

func networkFirewallSubnetMappingsDiff(old, new *schema.Set) ([]string, []*networkfirewall.SubnetMapping) {
	if old.Len() == 0 {
		return nil, expandNetworkFirewallSubnetMappings(new.List())
	}
	if new.Len() == 0 {
		return expandNetworkFirewallSubnetMappingIds(old.List()), nil
	}

	oldHashedSet := schema.NewSet(networkFirewallSubnetMappingsHash, old.List())
	newHashedSet := schema.NewSet(networkFirewallSubnetMappingsHash, new.List())

	toRemove := oldHashedSet.Difference(newHashedSet)
	toAdd := new.Difference(old)

	subnetsToRemove := expandNetworkFirewallSubnetMappingIds(toRemove.List())
	subnetsToAdd := expandNetworkFirewallSubnetMappings(toAdd.List())

	return subnetsToRemove, subnetsToAdd
}
