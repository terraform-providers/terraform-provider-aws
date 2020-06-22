package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	Wafv2WebACLCreateTimeout = 5 * time.Minute
	Wafv2WebACLUpdateTimeout = 5 * time.Minute
	Wafv2WebACLDeleteTimeout = 5 * time.Minute
)

func resourceAwsWafv2WebACL() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafv2WebACLCreate,
		Read:   resourceAwsWafv2WebACLRead,
		Update: resourceAwsWafv2WebACLUpdate,
		Delete: resourceAwsWafv2WebACLDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ID/NAME/SCOPE", d.Id())
				}
				id := idParts[0]
				name := idParts[1]
				scope := idParts[2]
				d.SetId(id)
				d.Set("name", name)
				d.Set("scope", scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"default_action": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow": wafv2EmptySchema(),
						"block": wafv2EmptySchema(),
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"lock_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]+$`), "must contain only alphanumeric hyphen and underscore characters"),
				),
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
			"rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allow": wafv2EmptySchema(),
									"block": wafv2EmptySchema(),
									"count": wafv2EmptySchema(),
								},
							},
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"override_action": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"count": wafv2EmptySchema(),
									"none":  wafv2EmptySchema(),
								},
							},
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"statement":         wafv2WebACLRootStatementSchema(3),
						"visibility_config": wafv2VisibilityConfigSchema(),
					},
				},
			},
			"tags":              tagsSchema(),
			"visibility_config": wafv2VisibilityConfigSchema(),
		},
	}
}

func resourceAwsWafv2WebACLCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	var resp *wafv2.CreateWebACLOutput

	params := &wafv2.CreateWebACLInput{
		Name:             aws.String(d.Get("name").(string)),
		Scope:            aws.String(d.Get("scope").(string)),
		DefaultAction:    expandWafv2DefaultAction(d.Get("default_action").([]interface{})),
		Rules:            expandWafv2WebACLRules(d.Get("rule").(*schema.Set).List()),
		VisibilityConfig: expandWafv2VisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(v).IgnoreAws().Wafv2Tags()
	}

	err := resource.Retry(Wafv2WebACLCreateTimeout, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateWebACL(params)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFUnavailableEntityException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateWebACL(params)
	}

	if err != nil {
		return fmt.Errorf("Error creating WAFv2 WebACL: %w", err)
	}

	if resp == nil || resp.Summary == nil {
		return fmt.Errorf("Error creating WAFv2 WebACL")
	}

	d.SetId(aws.StringValue(resp.Summary.Id))

	return resourceAwsWafv2WebACLRead(d, meta)
}

func resourceAwsWafv2WebACLRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &wafv2.GetWebACLInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetWebACL(params)
	if err != nil {
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAFv2 WebACL (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.WebACL == nil {
		return fmt.Errorf("Error getting WAFv2 WebACL")
	}

	d.Set("name", aws.StringValue(resp.WebACL.Name))
	d.Set("capacity", aws.Int64Value(resp.WebACL.Capacity))
	d.Set("description", aws.StringValue(resp.WebACL.Description))
	d.Set("arn", aws.StringValue(resp.WebACL.ARN))
	d.Set("lock_token", aws.StringValue(resp.LockToken))

	if err := d.Set("default_action", flattenWafv2DefaultAction(resp.WebACL.DefaultAction)); err != nil {
		return fmt.Errorf("Error setting default_action: %w", err)
	}

	if err := d.Set("rule", flattenWafv2WebACLRules(resp.WebACL.Rules)); err != nil {
		return fmt.Errorf("Error setting rule: %w", err)
	}

	if err := d.Set("visibility_config", flattenWafv2VisibilityConfig(resp.WebACL.VisibilityConfig)); err != nil {
		return fmt.Errorf("Error setting visibility_config: %w", err)
	}

	arn := aws.StringValue(resp.WebACL.ARN)
	tags, err := keyvaluetags.Wafv2ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("Error listing tags for WAFv2 WebACL (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %w", err)
	}

	return nil
}

func resourceAwsWafv2WebACLUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Updating WAFv2 WebACL %s", d.Id())

	u := &wafv2.UpdateWebACLInput{
		Id:               aws.String(d.Id()),
		Name:             aws.String(d.Get("name").(string)),
		Scope:            aws.String(d.Get("scope").(string)),
		LockToken:        aws.String(d.Get("lock_token").(string)),
		DefaultAction:    expandWafv2DefaultAction(d.Get("default_action").([]interface{})),
		Rules:            expandWafv2Rules(d.Get("rule").(*schema.Set).List()),
		VisibilityConfig: expandWafv2VisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok && len(v.(string)) > 0 {
		u.Description = aws.String(v.(string))
	}

	err := resource.Retry(Wafv2WebACLUpdateTimeout, func() *resource.RetryError {
		_, err := conn.UpdateWebACL(u)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFUnavailableEntityException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.UpdateWebACL(u)
	}

	if err != nil {
		if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "") {
			return fmt.Errorf("Error updating WAFv2 WebACL, resource has changed since last refresh please run a new plan before applying again: %w", err)
		}
		return fmt.Errorf("Error updating WAFv2 WebACL: %w", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Wafv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsWafv2WebACLRead(d, meta)
}

func resourceAwsWafv2WebACLDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Deleting WAFv2 WebACL %s", d.Id())

	r := &wafv2.DeleteWebACLInput{
		Id:        aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
		LockToken: aws.String(d.Get("lock_token").(string)),
	}

	err := resource.Retry(Wafv2WebACLDeleteTimeout, func() *resource.RetryError {
		_, err := conn.DeleteWebACL(r)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFAssociatedItemException, "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFUnavailableEntityException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteWebACL(r)
	}

	if err != nil {
		return fmt.Errorf("Error deleting WAFv2 WebACL: %w", err)
	}

	return nil
}

func wafv2WebACLRootStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"and_statement":                         wafv2StatementSchema(level - 1),
				"byte_match_statement":                  wafv2ByteMatchStatementSchema(),
				"geo_match_statement":                   wafv2GeoMatchStatementSchema(),
				"ip_set_reference_statement":            wafv2IpSetReferenceStatementSchema(),
				"managed_rule_group_statement":          wafv2ManagedRuleGroupStatementSchema(),
				"not_statement":                         wafv2StatementSchema(level - 1),
				"or_statement":                          wafv2StatementSchema(level - 1),
				"rate_based_statement":                  wafv2RateBasedStatementSchema(level - 1),
				"regex_pattern_set_reference_statement": wafv2RegexPatternSetReferenceStatementSchema(),
				"rule_group_reference_statement":        wafv2RuleGroupReferenceStatementSchema(),
				"size_constraint_statement":             wafv2SizeConstraintSchema(),
				"sqli_match_statement":                  wafv2SqliMatchStatementSchema(),
				"xss_match_statement":                   wafv2XssMatchStatementSchema(),
			},
		},
	}
}

func wafv2ManagedRuleGroupStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"excluded_rule": wafv2ExcludedRuleSchema(),
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				"vendor_name": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
			},
		},
	}
}

func wafv2ExcludedRuleSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
			},
		},
	}
}

func wafv2RateBasedStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Required field but currently only supports "IP"
				"aggregate_key_type": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  wafv2.RateBasedStatementAggregateKeyTypeIp,
					ValidateFunc: validation.StringInSlice([]string{
						wafv2.RateBasedStatementAggregateKeyTypeIp,
					}, false),
				},
				"limit": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(100, 2000000000),
				},
				"scope_down_statement": wafv2ScopeDownStatementSchema(level - 1),
			},
		},
	}
}

func wafv2ScopeDownStatementSchema(level int) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"and_statement":                         wafv2StatementSchema(level - 1),
				"byte_match_statement":                  wafv2ByteMatchStatementSchema(),
				"geo_match_statement":                   wafv2GeoMatchStatementSchema(),
				"ip_set_reference_statement":            wafv2IpSetReferenceStatementSchema(),
				"not_statement":                         wafv2StatementSchema(level - 1),
				"or_statement":                          wafv2StatementSchema(level - 1),
				"regex_pattern_set_reference_statement": wafv2RegexPatternSetReferenceStatementSchema(),
				"size_constraint_statement":             wafv2SizeConstraintSchema(),
				"sqli_match_statement":                  wafv2SqliMatchStatementSchema(),
				"xss_match_statement":                   wafv2XssMatchStatementSchema(),
			},
		},
	}
}

func wafv2RuleGroupReferenceStatementSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"arn": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateArn,
				},
				"excluded_rule": wafv2ExcludedRuleSchema(),
			},
		},
	}
}

func expandWafv2WebACLRules(l []interface{}) []*wafv2.Rule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]*wafv2.Rule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandWafv2WebACLRule(rule.(map[string]interface{})))
	}

	return rules
}

func expandWafv2WebACLRule(m map[string]interface{}) *wafv2.Rule {
	if m == nil {
		return nil
	}

	return &wafv2.Rule{
		Name:             aws.String(m["name"].(string)),
		Priority:         aws.Int64(int64(m["priority"].(int))),
		Action:           expandWafv2RuleAction(m["action"].([]interface{})),
		OverrideAction:   expandWafv2OverrideAction(m["override_action"].([]interface{})),
		Statement:        expandWafv2WebACLRootStatement(m["statement"].([]interface{})),
		VisibilityConfig: expandWafv2VisibilityConfig(m["visibility_config"].([]interface{})),
	}
}

func expandWafv2OverrideAction(l []interface{}) *wafv2.OverrideAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	action := &wafv2.OverrideAction{}

	if v, ok := m["count"]; ok && len(v.([]interface{})) > 0 {
		action.Count = &wafv2.CountAction{}
	}

	if v, ok := m["none"]; ok && len(v.([]interface{})) > 0 {
		action.None = &wafv2.NoneAction{}
	}

	return action
}

func expandWafv2DefaultAction(l []interface{}) *wafv2.DefaultAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	action := &wafv2.DefaultAction{}

	if v, ok := m["allow"]; ok && len(v.([]interface{})) > 0 {
		action.Allow = &wafv2.AllowAction{}
	}

	if v, ok := m["block"]; ok && len(v.([]interface{})) > 0 {
		action.Block = &wafv2.BlockAction{}
	}

	return action
}

func expandWafv2WebACLRootStatement(l []interface{}) *wafv2.Statement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return expandWafv2WebACLStatement(m)
}

func expandWafv2WebACLStatement(m map[string]interface{}) *wafv2.Statement {
	if m == nil {
		return nil
	}

	statement := &wafv2.Statement{}

	if v, ok := m["and_statement"]; ok {
		statement.AndStatement = expandWafv2AndStatement(v.([]interface{}))
	}

	if v, ok := m["byte_match_statement"]; ok {
		statement.ByteMatchStatement = expandWafv2ByteMatchStatement(v.([]interface{}))
	}

	if v, ok := m["ip_set_reference_statement"]; ok {
		statement.IPSetReferenceStatement = expandWafv2IpSetReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["geo_match_statement"]; ok {
		statement.GeoMatchStatement = expandWafv2GeoMatchStatement(v.([]interface{}))
	}

	if v, ok := m["managed_rule_group_statement"]; ok {
		statement.ManagedRuleGroupStatement = expandWafv2ManagedRuleGroupStatement(v.([]interface{}))
	}

	if v, ok := m["not_statement"]; ok {
		statement.NotStatement = expandWafv2NotStatement(v.([]interface{}))
	}

	if v, ok := m["or_statement"]; ok {
		statement.OrStatement = expandWafv2OrStatement(v.([]interface{}))
	}

	if v, ok := m["rate_based_statement"]; ok {
		statement.RateBasedStatement = expandWafv2RateBasedStatement(v.([]interface{}))
	}

	if v, ok := m["regex_pattern_set_reference_statement"]; ok {
		statement.RegexPatternSetReferenceStatement = expandWafv2RegexPatternSetReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["rule_group_reference_statement"]; ok {
		statement.RuleGroupReferenceStatement = expandWafv2RuleGroupReferenceStatement(v.([]interface{}))
	}

	if v, ok := m["size_constraint_statement"]; ok {
		statement.SizeConstraintStatement = expandWafv2SizeConstraintStatement(v.([]interface{}))
	}

	if v, ok := m["sqli_match_statement"]; ok {
		statement.SqliMatchStatement = expandWafv2SqliMatchStatement(v.([]interface{}))
	}

	if v, ok := m["xss_match_statement"]; ok {
		statement.XssMatchStatement = expandWafv2XssMatchStatement(v.([]interface{}))
	}

	return statement
}

func expandWafv2ManagedRuleGroupStatement(l []interface{}) *wafv2.ManagedRuleGroupStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	return &wafv2.ManagedRuleGroupStatement{
		ExcludedRules: expandWafv2ExcludedRules(m["excluded_rule"].([]interface{})),
		Name:          aws.String(m["name"].(string)),
		VendorName:    aws.String(m["vendor_name"].(string)),
	}
}

func expandWafv2RateBasedStatement(l []interface{}) *wafv2.RateBasedStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	r := &wafv2.RateBasedStatement{
		AggregateKeyType: aws.String(m["aggregate_key_type"].(string)),
		Limit:            aws.Int64(int64(m["limit"].(int))),
	}

	s := m["scope_down_statement"].([]interface{})
	if len(s) > 0 && s[0] != nil {
		r.ScopeDownStatement = expandWafv2Statement(s[0].(map[string]interface{}))
	}

	return r
}

func expandWafv2RuleGroupReferenceStatement(l []interface{}) *wafv2.RuleGroupReferenceStatement {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &wafv2.RuleGroupReferenceStatement{
		ARN:           aws.String(m["arn"].(string)),
		ExcludedRules: expandWafv2ExcludedRules(m["excluded_rule"].([]interface{})),
	}
}

func expandWafv2ExcludedRules(l []interface{}) []*wafv2.ExcludedRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	rules := make([]*wafv2.ExcludedRule, 0)

	for _, rule := range l {
		if rule == nil {
			continue
		}
		rules = append(rules, expandWafv2ExcludedRule(rule.(map[string]interface{})))
	}

	return rules
}

func expandWafv2ExcludedRule(m map[string]interface{}) *wafv2.ExcludedRule {
	if m == nil {
		return nil
	}

	return &wafv2.ExcludedRule{
		Name: aws.String(m["name"].(string)),
	}
}

func flattenWafv2WebACLRootStatement(s *wafv2.Statement) interface{} {
	if s == nil {
		return []interface{}{}
	}

	return []interface{}{flattenWafv2WebACLStatement(s)}
}

func flattenWafv2WebACLStatement(s *wafv2.Statement) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if s.AndStatement != nil {
		m["and_statement"] = flattenWafv2AndStatement(s.AndStatement)
	}

	if s.ByteMatchStatement != nil {
		m["byte_match_statement"] = flattenWafv2ByteMatchStatement(s.ByteMatchStatement)
	}

	if s.IPSetReferenceStatement != nil {
		m["ip_set_reference_statement"] = flattenWafv2IpSetReferenceStatement(s.IPSetReferenceStatement)
	}

	if s.GeoMatchStatement != nil {
		m["geo_match_statement"] = flattenWafv2GeoMatchStatement(s.GeoMatchStatement)
	}

	if s.ManagedRuleGroupStatement != nil {
		m["managed_rule_group_statement"] = flattenWafv2ManagedRuleGroupStatement(s.ManagedRuleGroupStatement)
	}

	if s.NotStatement != nil {
		m["not_statement"] = flattenWafv2NotStatement(s.NotStatement)
	}

	if s.OrStatement != nil {
		m["or_statement"] = flattenWafv2OrStatement(s.OrStatement)
	}

	if s.RateBasedStatement != nil {
		m["rate_based_statement"] = flattenWafv2RateBasedStatement(s.RateBasedStatement)
	}

	if s.RegexPatternSetReferenceStatement != nil {
		m["regex_pattern_set_reference_statement"] = flattenWafv2RegexPatternSetReferenceStatement(s.RegexPatternSetReferenceStatement)
	}

	if s.RuleGroupReferenceStatement != nil {
		m["rule_group_reference_statement"] = flattenWafv2RuleGroupReferenceStatement(s.RuleGroupReferenceStatement)
	}

	if s.SizeConstraintStatement != nil {
		m["size_constraint_statement"] = flattenWafv2SizeConstraintStatement(s.SizeConstraintStatement)
	}

	if s.SqliMatchStatement != nil {
		m["sqli_match_statement"] = flattenWafv2SqliMatchStatement(s.SqliMatchStatement)
	}

	if s.XssMatchStatement != nil {
		m["xss_match_statement"] = flattenWafv2XssMatchStatement(s.XssMatchStatement)
	}

	return m
}

func flattenWafv2WebACLRules(r []*wafv2.Rule) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, rule := range r {
		m := make(map[string]interface{})
		m["action"] = flattenWafv2RuleAction(rule.Action)
		m["override_action"] = flattenWafv2OverrideAction(rule.OverrideAction)
		m["name"] = aws.StringValue(rule.Name)
		m["priority"] = int(aws.Int64Value(rule.Priority))
		m["statement"] = flattenWafv2WebACLRootStatement(rule.Statement)
		m["visibility_config"] = flattenWafv2VisibilityConfig(rule.VisibilityConfig)
		out[i] = m
	}

	return out
}

func flattenWafv2OverrideAction(a *wafv2.OverrideAction) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.Count != nil {
		m["count"] = make([]map[string]interface{}, 1)
	}

	if a.None != nil {
		m["none"] = make([]map[string]interface{}, 1)
	}

	return []interface{}{m}
}

func flattenWafv2DefaultAction(a *wafv2.DefaultAction) interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if a.Allow != nil {
		m["allow"] = make([]map[string]interface{}, 1)
	}

	if a.Block != nil {
		m["block"] = make([]map[string]interface{}, 1)
	}

	return []interface{}{m}
}

func flattenWafv2ManagedRuleGroupStatement(r *wafv2.ManagedRuleGroupStatement) interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"excluded_rule": flattenWafv2ExcludedRules(r.ExcludedRules),
		"name":          aws.StringValue(r.Name),
		"vendor_name":   aws.StringValue(r.VendorName),
	}

	return []interface{}{m}
}

func flattenWafv2RateBasedStatement(r *wafv2.RateBasedStatement) interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"limit":                int(aws.Int64Value(r.Limit)),
		"aggregate_key_type":   aws.StringValue(r.AggregateKeyType),
		"scope_down_statement": nil,
	}

	if r.ScopeDownStatement != nil {
		m["scope_down_statement"] = []interface{}{flattenWafv2Statement(r.ScopeDownStatement)}
	}

	return []interface{}{m}
}

func flattenWafv2RuleGroupReferenceStatement(r *wafv2.RuleGroupReferenceStatement) interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"excluded_rule": flattenWafv2ExcludedRules(r.ExcludedRules),
		"arn":           aws.StringValue(r.ARN),
	}

	return []interface{}{m}
}

func flattenWafv2ExcludedRules(r []*wafv2.ExcludedRule) interface{} {
	out := make([]map[string]interface{}, len(r))
	for i, rule := range r {
		m := make(map[string]interface{})
		m["name"] = aws.StringValue(rule.Name)
		out[i] = m
	}

	return out
}
