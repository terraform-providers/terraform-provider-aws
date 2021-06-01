package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"

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

func resourceAwsNetworkFirewallRuleGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsNetworkFirewallRuleGroupCreate,
		ReadContext:   resourceAwsNetworkFirewallRuleGroupRead,
		UpdateContext: resourceAwsNetworkFirewallRuleGroupUpdate,
		DeleteContext: resourceAwsNetworkFirewallRuleGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rule_group": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_variables": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip_sets": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 32),
														validation.StringMatch(regexp.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
														validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_]+$`), "must contain only alphanumeric and underscore characters"),
													),
												},
												"ip_set": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"definition": {
																Type:     schema.TypeSet,
																Required: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: validateIpv4CIDRNetworkAddress,
																},
															},
														},
													},
												},
											},
										},
									},
									"port_sets": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 32),
														validation.StringMatch(regexp.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
														validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_]+$`), "must contain only alphanumeric and underscore characters"),
													),
												},
												"port_set": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"definition": {
																Type:     schema.TypeSet,
																Required: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
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
						"rules_source": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"rules_source_list": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"generated_rules_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(networkfirewall.GeneratedRulesType_Values(), false),
												},
												"target_types": {
													Type:     schema.TypeSet,
													Required: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(networkfirewall.TargetType_Values(), false),
													},
												},
												"targets": {
													Type:     schema.TypeSet,
													Required: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"rules_string": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"stateful_rule": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"action": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(networkfirewall.StatefulAction_Values(), false),
												},
												"header": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"destination": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.Any(
																	validateIpv4CIDRNetworkAddress,
																	validation.StringInSlice([]string{networkfirewall.StatefulRuleDirectionAny}, false),
																),
															},
															"destination_port": {
																Type:     schema.TypeString,
																Required: true,
															},
															"direction": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(networkfirewall.StatefulRuleDirection_Values(), false),
															},
															"protocol": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(networkfirewall.StatefulRuleProtocol_Values(), false),
															},
															"source": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.Any(
																	validateIpv4CIDRNetworkAddress,
																	validation.StringInSlice([]string{networkfirewall.StatefulRuleDirectionAny}, false),
																),
															},
															"source_port": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
												"rule_option": {
													Type:     schema.TypeSet,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"keyword": {
																Type:     schema.TypeString,
																Required: true,
															},
															"settings": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
														},
													},
												},
											},
										},
									},
									"stateless_rules_and_custom_actions": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"custom_action": customActionSchema(),
												"stateless_rule": {
													Type:     schema.TypeSet,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"priority": {
																Type:     schema.TypeInt,
																Required: true,
															},
															"rule_definition": {
																Type:     schema.TypeList,
																MaxItems: 1,
																Required: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"actions": {
																			Type:     schema.TypeSet,
																			Required: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
																		},
																		"match_attributes": {
																			Type:     schema.TypeList,
																			MaxItems: 1,
																			Required: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"destination": {
																						Type:     schema.TypeSet,
																						Optional: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"address_definition": {
																									Type:         schema.TypeString,
																									Required:     true,
																									ValidateFunc: validateIpv4CIDRNetworkAddress,
																								},
																							},
																						},
																					},
																					"destination_port": {
																						Type:     schema.TypeSet,
																						Optional: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"from_port": {
																									Type:     schema.TypeInt,
																									Required: true,
																								},
																								"to_port": {
																									Type:     schema.TypeInt,
																									Optional: true,
																								},
																							},
																						},
																					},
																					"protocols": {
																						Type:     schema.TypeSet,
																						Optional: true,
																						Elem:     &schema.Schema{Type: schema.TypeInt},
																					},
																					"source": {
																						Type:     schema.TypeSet,
																						Optional: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"address_definition": {
																									Type:         schema.TypeString,
																									Required:     true,
																									ValidateFunc: validateIpv4CIDRNetworkAddress,
																								},
																							},
																						},
																					},
																					"source_port": {
																						Type:     schema.TypeSet,
																						Optional: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"from_port": {
																									Type:     schema.TypeInt,
																									Required: true,
																								},
																								"to_port": {
																									Type:     schema.TypeInt,
																									Optional: true,
																								},
																							},
																						},
																					},
																					"tcp_flag": {
																						Type:     schema.TypeSet,
																						Optional: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"flags": {
																									Type:     schema.TypeSet,
																									Required: true,
																									Elem: &schema.Schema{
																										Type:         schema.TypeString,
																										ValidateFunc: validation.StringInSlice(networkfirewall.TCPFlag_Values(), false),
																									},
																								},
																								"masks": {
																									Type:     schema.TypeSet,
																									Optional: true,
																									Elem: &schema.Schema{
																										Type:         schema.TypeString,
																										ValidateFunc: validation.StringInSlice(networkfirewall.TCPFlag_Values(), false),
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
					},
				},
			},
			"rules": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(networkfirewall.RuleGroupType_Values(), false),
			},
			"update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsNetworkFirewallRuleGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn
	name := d.Get("name").(string)

	input := &networkfirewall.CreateRuleGroupInput{
		Capacity:      aws.Int64(int64(d.Get("capacity").(int))),
		RuleGroupName: aws.String(name),
		Type:          aws.String(d.Get("type").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("rule_group"); ok {
		if vRaw := v.([]interface{}); len(vRaw) > 0 && vRaw[0] != nil {
			input.RuleGroup = expandNetworkFirewallRuleGroup(vRaw)
		}
	}
	if v, ok := d.GetOk("rules"); ok {
		input.Rules = aws.String(v.(string))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().NetworkfirewallTags()
	}

	log.Printf("[DEBUG] Creating NetworkFirewall Rule Group %s", name)

	output, err := conn.CreateRuleGroupWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating NetworkFirewall Rule Group %s: %w", name, err))
	}
	if output == nil || output.RuleGroupResponse == nil {
		return diag.FromErr(fmt.Errorf("error creating NetworkFirewall Rule Group (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.RuleGroupResponse.RuleGroupArn))

	return resourceAwsNetworkFirewallRuleGroupRead(ctx, d, meta)
}

func resourceAwsNetworkFirewallRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading NetworkFirewall Rule Group %s", d.Id())

	output, err := finder.RuleGroup(ctx, conn, d.Id())
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] NetworkFirewall Rule Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Rule Group (%s): %w", d.Id(), err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Rule Group (%s): empty output", d.Id()))
	}
	if output.RuleGroupResponse == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Rule Group (%s): empty output.RuleGroupResponse", d.Id()))
	}
	if output.RuleGroup == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Rule Group (%s): empty output.RuleGroup", d.Id()))
	}

	resp := output.RuleGroupResponse
	ruleGroup := output.RuleGroup

	d.Set("arn", resp.RuleGroupArn)
	d.Set("capacity", resp.Capacity)
	d.Set("description", resp.Description)
	d.Set("name", resp.RuleGroupName)
	d.Set("type", resp.Type)
	d.Set("update_token", output.UpdateToken)

	if err := d.Set("rule_group", flattenNetworkFirewallRuleGroup(ruleGroup)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting rule_group: %w", err))
	}

	if err := d.Set("tags", keyvaluetags.NetworkfirewallKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	return nil
}

func resourceAwsNetworkFirewallRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn
	arn := d.Id()

	log.Printf("[DEBUG] Updating NetworkFirewall Rule Group %s", arn)

	if d.HasChanges("description", "rule_group", "rules", "type") {
		// Provide updated object with the currently configured fields
		input := &networkfirewall.UpdateRuleGroupInput{
			RuleGroupArn: aws.String(arn),
			Type:         aws.String(d.Get("type").(string)),
			UpdateToken:  aws.String(d.Get("update_token").(string)),
		}
		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}
		if v, ok := d.GetOk("rule_group"); ok {
			input.RuleGroup = expandNetworkFirewallRuleGroup(v.([]interface{}))
		}
		if v, ok := d.GetOk("rules"); ok {
			input.Rules = aws.String(v.(string))
		}

		_, err := conn.UpdateRuleGroupWithContext(ctx, input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Rule Group (%s): %w", arn, err))
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.NetworkfirewallUpdateTags(conn, arn, o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Rule Group (%s) tags: %w", arn, err))
		}
	}

	return resourceAwsNetworkFirewallRuleGroupRead(ctx, d, meta)
}

func resourceAwsNetworkFirewallRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn

	log.Printf("[DEBUG] Deleting NetworkFirewall Rule Group %s", d.Id())

	input := &networkfirewall.DeleteRuleGroupInput{
		RuleGroupArn: aws.String(d.Id()),
	}
	err := resource.RetryContext(ctx, waiter.RuleGroupDeleteTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.DeleteRuleGroupWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, networkfirewall.ErrCodeInvalidOperationException, "Unable to delete the object because it is still in use") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteRuleGroupWithContext(ctx, input)
	}

	if err != nil {
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting NetworkFirewall Rule Group (%s): %w", d.Id(), err))
	}

	if _, err := waiter.RuleGroupDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error waiting for NetworkFirewall Rule Group (%s) to delete: %w", d.Id(), err))
	}

	return nil
}

func expandNetworkFirewallStatefulRuleHeader(l []interface{}) *networkfirewall.Header {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}
	header := &networkfirewall.Header{}
	if v, ok := tfMap["destination"].(string); ok && v != "" {
		header.Destination = aws.String(v)
	}
	if v, ok := tfMap["destination_port"].(string); ok && v != "" {
		header.DestinationPort = aws.String(v)
	}
	if v, ok := tfMap["direction"].(string); ok && v != "" {
		header.Direction = aws.String(v)
	}
	if v, ok := tfMap["protocol"].(string); ok && v != "" {
		header.Protocol = aws.String(v)
	}
	if v, ok := tfMap["source"].(string); ok && v != "" {
		header.Source = aws.String(v)
	}
	if v, ok := tfMap["source_port"].(string); ok && v != "" {
		header.SourcePort = aws.String(v)
	}

	return header
}

func expandNetworkFirewallStatefulRuleOptions(l []interface{}) []*networkfirewall.RuleOption {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	ruleOptions := make([]*networkfirewall.RuleOption, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		keyword := tfMap["keyword"].(string)
		option := &networkfirewall.RuleOption{
			Keyword: aws.String(keyword),
		}
		if v, ok := tfMap["settings"].(*schema.Set); ok && v.Len() > 0 {
			option.Settings = expandStringSet(v)
		}
		ruleOptions = append(ruleOptions, option)
	}

	return ruleOptions
}

func expandNetworkFirewallRulesSourceList(l []interface{}) *networkfirewall.RulesSourceList {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}
	rulesSourceList := &networkfirewall.RulesSourceList{}
	if v, ok := tfMap["generated_rules_type"].(string); ok && v != "" {
		rulesSourceList.GeneratedRulesType = aws.String(v)
	}
	if v, ok := tfMap["target_types"].(*schema.Set); ok && v.Len() > 0 {
		rulesSourceList.TargetTypes = expandStringSet(v)
	}
	if v, ok := tfMap["targets"].(*schema.Set); ok && v.Len() > 0 {
		rulesSourceList.Targets = expandStringSet(v)
	}

	return rulesSourceList
}

func expandNetworkFirewallStatefulRules(l []interface{}) []*networkfirewall.StatefulRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]*networkfirewall.StatefulRule, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		rule := &networkfirewall.StatefulRule{}
		if v, ok := tfMap["action"].(string); ok && v != "" {
			rule.Action = aws.String(v)
		}
		if v, ok := tfMap["header"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			rule.Header = expandNetworkFirewallStatefulRuleHeader(v)
		}
		if v, ok := tfMap["rule_option"].(*schema.Set); ok && v.Len() > 0 {
			rule.RuleOptions = expandNetworkFirewallStatefulRuleOptions(v.List())
		}
		rules = append(rules, rule)
	}

	return rules
}

func expandNetworkFirewallRuleGroup(l []interface{}) *networkfirewall.RuleGroup {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}
	ruleGroup := &networkfirewall.RuleGroup{}
	if tfList, ok := tfMap["rule_variables"].([]interface{}); ok && len(tfList) > 0 && tfList[0] != nil {
		ruleVariables := &networkfirewall.RuleVariables{}
		rvMap, ok := tfList[0].(map[string]interface{})
		if ok {
			if v, ok := rvMap["ip_sets"].(*schema.Set); ok && v.Len() > 0 {
				ruleVariables.IPSets = expandNetworkFirewallIPSets(v.List())
			}
			if v, ok := rvMap["port_sets"].(*schema.Set); ok && v.Len() > 0 {
				ruleVariables.PortSets = expandNetworkFirewallPortSets(v.List())
			}
			ruleGroup.RuleVariables = ruleVariables
		}
	}
	if tfList, ok := tfMap["rules_source"].([]interface{}); ok && len(tfList) > 0 && tfList[0] != nil {
		rulesSource := &networkfirewall.RulesSource{}
		rsMap, ok := tfList[0].(map[string]interface{})
		if ok {
			if v, ok := rsMap["rules_source_list"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				rulesSource.RulesSourceList = expandNetworkFirewallRulesSourceList(v)
			}
			if v, ok := rsMap["rules_string"].(string); ok && v != "" {
				rulesSource.RulesString = aws.String(v)
			}
			if v, ok := rsMap["stateful_rule"].(*schema.Set); ok && v.Len() > 0 {
				rulesSource.StatefulRules = expandNetworkFirewallStatefulRules(v.List())
			}
			if v, ok := rsMap["stateless_rules_and_custom_actions"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				rulesSource.StatelessRulesAndCustomActions = expandNetworkFirewallStatelessRulesAndCustomActions(v)
			}
			ruleGroup.RulesSource = rulesSource
		}
	}

	return ruleGroup
}

func expandNetworkFirewallIPSets(l []interface{}) map[string]*networkfirewall.IPSet {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := make(map[string]*networkfirewall.IPSet)
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		if key, ok := tfMap["key"].(string); ok && key != "" {
			if tfList, ok := tfMap["ip_set"].([]interface{}); ok && len(tfList) > 0 && tfList[0] != nil {
				tfMap, ok := tfList[0].(map[string]interface{})
				if ok {
					if tfSet, ok := tfMap["definition"].(*schema.Set); ok && tfSet.Len() > 0 {
						ipSet := &networkfirewall.IPSet{
							Definition: expandStringSet(tfSet),
						}
						m[key] = ipSet
					}
				}
			}
		}
	}

	return m
}

func expandNetworkFirewallPortSets(l []interface{}) map[string]*networkfirewall.PortSet {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := make(map[string]*networkfirewall.PortSet)
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		if key, ok := tfMap["key"].(string); ok && key != "" {
			if tfList, ok := tfMap["port_set"].([]interface{}); ok && len(tfList) > 0 && tfList[0] != nil {
				tfMap, ok := tfList[0].(map[string]interface{})
				if ok {
					if tfSet, ok := tfMap["definition"].(*schema.Set); ok && tfSet.Len() > 0 {
						ipSet := &networkfirewall.PortSet{
							Definition: expandStringSet(tfSet),
						}
						m[key] = ipSet
					}
				}
			}
		}
	}

	return m
}

func expandNetworkFirewallAddresses(l []interface{}) []*networkfirewall.Address {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	destinations := make([]*networkfirewall.Address, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		destination := &networkfirewall.Address{}
		if v, ok := tfMap["address_definition"].(string); ok && v != "" {
			destination.AddressDefinition = aws.String(v)
		}
		destinations = append(destinations, destination)
	}
	return destinations
}

func expandNetworkFirewallPortRanges(l []interface{}) []*networkfirewall.PortRange {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	ports := make([]*networkfirewall.PortRange, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		port := &networkfirewall.PortRange{}
		if v, ok := tfMap["from_port"].(int); ok {
			port.FromPort = aws.Int64(int64(v))
		}
		if v, ok := tfMap["to_port"].(int); ok {
			port.ToPort = aws.Int64(int64(v))
		}
		ports = append(ports, port)
	}
	return ports
}

func expandNetworkFirewallTCPFlags(l []interface{}) []*networkfirewall.TCPFlagField {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	tcpFlags := make([]*networkfirewall.TCPFlagField, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		tcpFlag := &networkfirewall.TCPFlagField{}
		if v, ok := tfMap["flags"].(*schema.Set); ok && v.Len() > 0 {
			tcpFlag.Flags = expandStringSet(v)
		}
		if v, ok := tfMap["masks"].(*schema.Set); ok && v.Len() > 0 {
			tcpFlag.Masks = expandStringSet(v)
		}
		tcpFlags = append(tcpFlags, tcpFlag)
	}
	return tcpFlags
}

func expandNetworkFirewallMatchAttributes(l []interface{}) *networkfirewall.MatchAttributes {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}
	matchAttributes := &networkfirewall.MatchAttributes{}
	if v, ok := tfMap["destination"].(*schema.Set); ok && v.Len() > 0 {
		matchAttributes.Destinations = expandNetworkFirewallAddresses(v.List())
	}
	if v, ok := tfMap["destination_port"].(*schema.Set); ok && v.Len() > 0 {
		matchAttributes.DestinationPorts = expandNetworkFirewallPortRanges(v.List())
	}
	if v, ok := tfMap["protocols"].(*schema.Set); ok && v.Len() > 0 {
		matchAttributes.Protocols = expandInt64Set(v)
	}
	if v, ok := tfMap["source"].(*schema.Set); ok && v.Len() > 0 {
		matchAttributes.Sources = expandNetworkFirewallAddresses(v.List())
	}
	if v, ok := tfMap["source_port"].(*schema.Set); ok && v.Len() > 0 {
		matchAttributes.SourcePorts = expandNetworkFirewallPortRanges(v.List())
	}
	if v, ok := tfMap["tcp_flag"].(*schema.Set); ok && v.Len() > 0 {
		matchAttributes.TCPFlags = expandNetworkFirewallTCPFlags(v.List())
	}

	return matchAttributes
}

func expandNetworkFirewallRuleDefinition(l []interface{}) *networkfirewall.RuleDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}
	rd := &networkfirewall.RuleDefinition{}
	if v, ok := tfMap["actions"].(*schema.Set); ok && v.Len() > 0 {
		rd.Actions = expandStringSet(v)
	}
	if v, ok := tfMap["match_attributes"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		rd.MatchAttributes = expandNetworkFirewallMatchAttributes(v)
	}
	return rd
}

func expandNetworkFirewallStatelessRules(l []interface{}) []*networkfirewall.StatelessRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	statelessRules := make([]*networkfirewall.StatelessRule, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		statelessRule := &networkfirewall.StatelessRule{}
		if v, ok := tfMap["priority"].(int); ok && v > 0 {
			statelessRule.Priority = aws.Int64(int64(v))
		}
		if v, ok := tfMap["rule_definition"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			statelessRule.RuleDefinition = expandNetworkFirewallRuleDefinition(v)
		}
		statelessRules = append(statelessRules, statelessRule)
	}

	return statelessRules
}

func expandNetworkFirewallStatelessRulesAndCustomActions(l []interface{}) *networkfirewall.StatelessRulesAndCustomActions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	s := &networkfirewall.StatelessRulesAndCustomActions{}
	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}
	if v, ok := tfMap["custom_action"].(*schema.Set); ok && v.Len() > 0 {
		s.CustomActions = expandNetworkFirewallCustomActions(v.List())
	}
	if v, ok := tfMap["stateless_rule"].(*schema.Set); ok && v.Len() > 0 {
		s.StatelessRules = expandNetworkFirewallStatelessRules(v.List())
	}

	return s
}

func flattenNetworkFirewallRuleGroup(r *networkfirewall.RuleGroup) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"rule_variables": flattenNetworkFirewallRuleVariables(r.RuleVariables),
		"rules_source":   flattenNetworkFirewallRulesSource(r.RulesSource),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallRuleVariables(rv *networkfirewall.RuleVariables) []interface{} {
	if rv == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"ip_sets":   flattenNetworkFirewallIPSets(rv.IPSets),
		"port_sets": flattenNetworkFirewallPortSets(rv.PortSets),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallIPSets(m map[string]*networkfirewall.IPSet) []interface{} {
	if m == nil {
		return []interface{}{}
	}
	sets := make([]interface{}, 0, len(m))
	for k, v := range m {
		tfMap := map[string]interface{}{
			"key":    k,
			"ip_set": flattenNetworkFirewallIPSet(v),
		}
		sets = append(sets, tfMap)
	}

	return sets
}

func flattenNetworkFirewallPortSets(m map[string]*networkfirewall.PortSet) []interface{} {
	if m == nil {
		return []interface{}{}
	}
	sets := make([]interface{}, 0, len(m))
	for k, v := range m {
		tfMap := map[string]interface{}{
			"key":      k,
			"port_set": flattenNetworkFirewallPortSet(v),
		}
		sets = append(sets, tfMap)
	}

	return sets
}

func flattenNetworkFirewallIPSet(i *networkfirewall.IPSet) []interface{} {
	if i == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"definition": flattenStringSet(i.Definition),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallPortSet(p *networkfirewall.PortSet) []interface{} {
	if p == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"definition": flattenStringSet(p.Definition),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallRulesSource(rs *networkfirewall.RulesSource) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"rules_source_list":                  flattenNetworkFirewallRulesSourceList(rs.RulesSourceList),
		"rules_string":                       aws.StringValue(rs.RulesString),
		"stateful_rule":                      flattenNetworkFirewallStatefulRules(rs.StatefulRules),
		"stateless_rules_and_custom_actions": flattenNetworkFirewallStatelessRulesAndCustomActions(rs.StatelessRulesAndCustomActions),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallRulesSourceList(r *networkfirewall.RulesSourceList) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"generated_rules_type": aws.StringValue(r.GeneratedRulesType),
		"target_types":         flattenStringSet(r.TargetTypes),
		"targets":              flattenStringSet(r.Targets),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallStatefulRules(sr []*networkfirewall.StatefulRule) []interface{} {
	if sr == nil {
		return []interface{}{}
	}
	rules := make([]interface{}, 0, len(sr))
	for _, s := range sr {
		m := map[string]interface{}{
			"action":      aws.StringValue(s.Action),
			"header":      flattenNetworkFirewallHeader(s.Header),
			"rule_option": flattenNetworkFirewallRuleOptions(s.RuleOptions),
		}
		rules = append(rules, m)
	}
	return rules
}

func flattenNetworkFirewallHeader(h *networkfirewall.Header) []interface{} {
	if h == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"destination":      aws.StringValue(h.Destination),
		"destination_port": aws.StringValue(h.DestinationPort),
		"direction":        aws.StringValue(h.Direction),
		"protocol":         aws.StringValue(h.Protocol),
		"source":           aws.StringValue(h.Source),
		"source_port":      aws.StringValue(h.SourcePort),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallRuleOptions(o []*networkfirewall.RuleOption) []interface{} {
	if o == nil {
		return []interface{}{}
	}

	options := make([]interface{}, 0, len(o))
	for _, option := range o {
		m := map[string]interface{}{
			"keyword":  aws.StringValue(option.Keyword),
			"settings": aws.StringValueSlice(option.Settings),
		}
		options = append(options, m)
	}

	return options
}

func flattenNetworkFirewallStatelessRulesAndCustomActions(sr *networkfirewall.StatelessRulesAndCustomActions) []interface{} {
	if sr == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"custom_action":  flattenNetworkFirewallCustomActions(sr.CustomActions),
		"stateless_rule": flattenNetworkFirewallStatelessRules(sr.StatelessRules),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallStatelessRules(sr []*networkfirewall.StatelessRule) []interface{} {
	if sr == nil {
		return []interface{}{}
	}

	rules := make([]interface{}, 0, len(sr))
	for _, s := range sr {
		rule := map[string]interface{}{
			"priority":        int(aws.Int64Value(s.Priority)),
			"rule_definition": flattenNetworkFirewallRuleDefinition(s.RuleDefinition),
		}
		rules = append(rules, rule)
	}

	return rules
}

func flattenNetworkFirewallRuleDefinition(r *networkfirewall.RuleDefinition) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"actions":          flattenStringSet(r.Actions),
		"match_attributes": flattenNetworkFirewallMatchAttributes(r.MatchAttributes),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallMatchAttributes(ma *networkfirewall.MatchAttributes) []interface{} {
	if ma == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"destination":      flattenNetworkFirewallAddresses(ma.Destinations),
		"destination_port": flattenNetworkFirewallPortRanges(ma.DestinationPorts),
		"protocols":        flattenInt64Set(ma.Protocols),
		"source":           flattenNetworkFirewallAddresses(ma.Sources),
		"source_port":      flattenNetworkFirewallPortRanges(ma.SourcePorts),
		"tcp_flag":         flattenNetworkFirewallTCPFlags(ma.TCPFlags),
	}

	return []interface{}{m}
}

func flattenNetworkFirewallAddresses(d []*networkfirewall.Address) []interface{} {
	if d == nil {
		return []interface{}{}
	}

	destinations := make([]interface{}, 0, len(d))
	for _, addr := range d {
		m := map[string]interface{}{
			"address_definition": aws.StringValue(addr.AddressDefinition),
		}
		destinations = append(destinations, m)
	}

	return destinations
}

func flattenNetworkFirewallPortRanges(pr []*networkfirewall.PortRange) []interface{} {
	if pr == nil {
		return []interface{}{}
	}

	portRanges := make([]interface{}, 0, len(pr))
	for _, r := range pr {
		m := map[string]interface{}{
			"from_port": int(aws.Int64Value(r.FromPort)),
			"to_port":   int(aws.Int64Value(r.ToPort)),
		}
		portRanges = append(portRanges, m)
	}

	return portRanges
}

func flattenNetworkFirewallTCPFlags(t []*networkfirewall.TCPFlagField) []interface{} {
	if t == nil {
		return []interface{}{}
	}
	flagFields := make([]interface{}, 0, len(t))
	for _, v := range t {
		m := map[string]interface{}{
			"flags": flattenStringSet(v.Flags),
			"masks": flattenStringSet(v.Masks),
		}
		flagFields = append(flagFields, m)
	}

	return flagFields
}
