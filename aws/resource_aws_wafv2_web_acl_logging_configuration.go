package aws

import (
	"bytes"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
)

func resourceAwsWafv2WebACLLoggingConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafv2WebACLLoggingConfigurationPut,
		Read:   resourceAwsWafv2WebACLLoggingConfigurationRead,
		Update: resourceAwsWafv2WebACLLoggingConfigurationPut,
		Delete: resourceAwsWafv2WebACLLoggingConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"log_destination_configs": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
				Description: "AWS Kinesis Firehose Delivery Stream ARNs",
			},
			"logging_filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_behavior": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(wafv2.FilterBehavior_Values(), false),
						},
						"filter": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"behavior": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(wafv2.FilterBehavior_Values(), false),
									},
									"condition": {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"action_condition": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"action": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(wafv2.ActionValue_Values(), false),
															},
														},
													},
												},
												"label_name_condition": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"label_name": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(1, 1024),
																	validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z_\-:]+$`), "must contain only alphanumeric characters, underscores, hyphens, and colons"),
																),
															},
														},
													},
												},
											},
										},
									},
									"requirement": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(wafv2.FilterRequirement_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			"redacted_fields": {
				// To allow this argument and its nested fields with Empty Schemas (e.g. "method")
				// to be correctly interpreted, this argument must be of type List,
				// otherwise, at apply-time a field configured as an empty block
				// (e.g. body {}) will result in a nil redacted_fields attribute
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// TODO: remove attributes marked as Deprecated
						// as they are not supported by the WAFv2 API
						// in the context of WebACL Logging Configurations
						"all_query_arguments": wafv2EmptySchemaDeprecated(),
						"body":                wafv2EmptySchemaDeprecated(),
						"method":              wafv2EmptySchema(),
						"query_string":        wafv2EmptySchema(),
						"single_header": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 40),
											// The value is returned in lower case by the API.
											// Trying to solve it with StateFunc and/or DiffSuppressFunc resulted in hash problem of the rule field or didn't work.
											validation.StringMatch(regexp.MustCompile(`^[a-z0-9-_]+$`), "must contain only lowercase alphanumeric characters, underscores, and hyphens"),
										),
									},
								},
							},
						},
						"single_query_argument": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 30),
											// The value is returned in lower case by the API.
											// Trying to solve it with StateFunc and/or DiffSuppressFunc resulted in hash problem of the rule field or didn't work.
											validation.StringMatch(regexp.MustCompile(`^[a-z0-9-_]+$`), "must contain only lowercase alphanumeric characters, underscores, and hyphens"),
										),
										Deprecated: "Not supported by WAFv2 API",
									},
								},
							},
							Deprecated: "Not supported by WAFv2 API",
						},
						"uri_path": wafv2EmptySchema(),
					},
				},
				Description:      "Parts of the request to exclude from logs",
				DiffSuppressFunc: suppressRedactedFieldsDiff,
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
				Description:  "AWS WebACL ARN",
			},
		},
	}
}

func resourceAwsWafv2WebACLLoggingConfigurationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	resourceArn := d.Get("resource_arn").(string)

	config := &wafv2.LoggingConfiguration{
		LogDestinationConfigs: expandStringSet(d.Get("log_destination_configs").(*schema.Set)),
		ResourceArn:           aws.String(resourceArn),
	}

	if v, ok := d.GetOk("logging_filter"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.LoggingFilter = expandWafv2LoggingFilter(v.([]interface{}))
	}

	if v, ok := d.GetOk("redacted_fields"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.RedactedFields = expandWafv2RedactedFields(v.([]interface{}))
	} else {
		config.RedactedFields = []*wafv2.FieldToMatch{}
	}

	input := &wafv2.PutLoggingConfigurationInput{
		LoggingConfiguration: config,
	}

	output, err := conn.PutLoggingConfiguration(input)

	if err != nil {
		return fmt.Errorf("error putting WAFv2 Logging Configuration for resource (%s): %w", resourceArn, err)
	}

	if output == nil || output.LoggingConfiguration == nil {
		return fmt.Errorf("error putting WAFv2 Logging Configuration for resource (%s): empty response", resourceArn)
	}

	d.SetId(aws.StringValue(output.LoggingConfiguration.ResourceArn))

	return resourceAwsWafv2WebACLLoggingConfigurationRead(d, meta)
}

func resourceAwsWafv2WebACLLoggingConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	input := &wafv2.GetLoggingConfigurationInput{
		ResourceArn: aws.String(d.Id()),
	}

	output, err := conn.GetLoggingConfiguration(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		log.Printf("[WARN] WAFv2 Logging Configuration for WebACL with ARN %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading WAFv2 Logging Configuration for resource (%s): %w", d.Id(), err)
	}

	if output == nil || output.LoggingConfiguration == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading WAFv2 Logging Configuration for resource (%s): empty response after creation", d.Id())
		}
		log.Printf("[WARN] WAFv2 Logging Configuration for WebACL with ARN %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	loggingConfig := output.LoggingConfiguration

	if err := d.Set("log_destination_configs", flattenStringList(loggingConfig.LogDestinationConfigs)); err != nil {
		return fmt.Errorf("error setting log_destination_configs: %w", err)
	}

	if err := d.Set("logging_filter", flattenWafv2LoggingFilter(loggingConfig.LoggingFilter)); err != nil {
		return fmt.Errorf("error setting logging_filter: %w", err)
	}

	if err := d.Set("redacted_fields", flattenWafv2RedactedFields(loggingConfig.RedactedFields)); err != nil {
		return fmt.Errorf("error setting redacted_fields: %w", err)
	}

	d.Set("resource_arn", loggingConfig.ResourceArn)

	return nil
}

func resourceAwsWafv2WebACLLoggingConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	input := &wafv2.DeleteLoggingConfigurationInput{
		ResourceArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteLoggingConfiguration(input)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting WAFv2 Logging Configuration for resource (%s): %w", d.Id(), err)
	}

	return nil
}

func expandWafv2LoggingFilter(l []interface{}) *wafv2.LoggingFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	loggingFilter := &wafv2.LoggingFilter{}

	if v, ok := tfMap["default_behavior"].(string); ok && v != "" {
		loggingFilter.DefaultBehavior = aws.String(v)
	}

	if v, ok := tfMap["filter"].(*schema.Set); ok && v.Len() > 0 {
		loggingFilter.Filters = expandWafv2Filters(v.List())
	}

	return loggingFilter
}

func expandWafv2Filters(l []interface{}) []*wafv2.Filter {
	if len(l) == 0 {
		return nil
	}

	var filters []*wafv2.Filter

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		filter := &wafv2.Filter{}

		if v, ok := tfMap["behavior"].(string); ok && v != "" {
			filter.Behavior = aws.String(v)
		}

		if v, ok := tfMap["condition"].(*schema.Set); ok && v.Len() > 0 {
			filter.Conditions = expandWafv2FilterConditions(v.List())
		}

		if v, ok := tfMap["requirement"].(string); ok && v != "" {
			filter.Requirement = aws.String(v)
		}

		filters = append(filters, filter)
	}

	return filters
}

func expandWafv2FilterConditions(l []interface{}) []*wafv2.Condition {
	if len(l) == 0 {
		return nil
	}

	var conditions []*wafv2.Condition

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		condition := &wafv2.Condition{}

		if v, ok := tfMap["action_condition"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			condition.ActionCondition = expandWafv2ActionCondition(v)
		}

		if v, ok := tfMap["label_name_condition"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			condition.LabelNameCondition = expandWafv2LabelNameCondition(v)
		}

		conditions = append(conditions, condition)
	}

	return conditions
}

func expandWafv2ActionCondition(l []interface{}) *wafv2.ActionCondition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	condition := &wafv2.ActionCondition{}

	if v, ok := tfMap["action"].(string); ok && v != "" {
		condition.Action = aws.String(v)
	}

	return condition
}

func expandWafv2LabelNameCondition(l []interface{}) *wafv2.LabelNameCondition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	condition := &wafv2.LabelNameCondition{}

	if v, ok := tfMap["label_name"].(string); ok && v != "" {
		condition.LabelName = aws.String(v)
	}

	return condition
}

func expandWafv2RedactedFields(fields []interface{}) []*wafv2.FieldToMatch {
	redactedFields := make([]*wafv2.FieldToMatch, 0, len(fields))
	for _, field := range fields {
		redactedFields = append(redactedFields, expandWafv2RedactedField(field))
	}
	return redactedFields
}

func expandWafv2RedactedField(field interface{}) *wafv2.FieldToMatch {
	m := field.(map[string]interface{})

	f := &wafv2.FieldToMatch{}

	// While the FieldToMatch struct allows more than 1 of its fields to be set,
	// the WAFv2 API does not. In addition, in the context of Logging Configuration requests,
	// the WAFv2 API only supports the following redacted fields.
	// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/14244
	if v, ok := m["method"]; ok && len(v.([]interface{})) > 0 {
		f.Method = &wafv2.Method{}
	}

	if v, ok := m["query_string"]; ok && len(v.([]interface{})) > 0 {
		f.QueryString = &wafv2.QueryString{}
	}

	if v, ok := m["single_header"]; ok && len(v.([]interface{})) > 0 {
		f.SingleHeader = expandWafv2SingleHeader(m["single_header"].([]interface{}))
	}

	if v, ok := m["uri_path"]; ok && len(v.([]interface{})) > 0 {
		f.UriPath = &wafv2.UriPath{}
	}

	return f
}

func flattenWafv2LoggingFilter(filter *wafv2.LoggingFilter) []interface{} {
	if filter == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"default_behavior": aws.StringValue(filter.DefaultBehavior),
		"filter":           flattenWafv2Filters(filter.Filters),
	}

	return []interface{}{m}
}

func flattenWafv2Filters(f []*wafv2.Filter) []interface{} {
	if len(f) == 0 {
		return []interface{}{}
	}

	var filters []interface{}

	for _, filter := range f {
		if filter == nil {
			continue
		}

		m := map[string]interface{}{
			"behavior":    aws.StringValue(filter.Behavior),
			"condition":   flattenWafv2FilterConditions(filter.Conditions),
			"requirement": aws.StringValue(filter.Requirement),
		}

		filters = append(filters, m)
	}

	return filters
}

func flattenWafv2FilterConditions(c []*wafv2.Condition) []interface{} {
	if len(c) == 0 {
		return []interface{}{}
	}

	var conditions []interface{}

	for _, condition := range c {
		if condition == nil {
			continue
		}

		m := map[string]interface{}{
			"action_condition":     flattenWafv2ActionCondition(condition.ActionCondition),
			"label_name_condition": flattenWafv2LabelNameCondition(condition.LabelNameCondition),
		}

		conditions = append(conditions, m)
	}

	return conditions
}

func flattenWafv2ActionCondition(a *wafv2.ActionCondition) []interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"action": aws.StringValue(a.Action),
	}

	return []interface{}{m}
}

func flattenWafv2LabelNameCondition(l *wafv2.LabelNameCondition) []interface{} {
	if l == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"label_name": aws.StringValue(l.LabelName),
	}

	return []interface{}{m}
}

func flattenWafv2RedactedFields(fields []*wafv2.FieldToMatch) []interface{} {
	redactedFields := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		redactedFields = append(redactedFields, flattenWafv2RedactedField(field))
	}
	return redactedFields
}

func flattenWafv2RedactedField(f *wafv2.FieldToMatch) map[string]interface{} {
	m := map[string]interface{}{}

	if f == nil {
		return m
	}

	// In the context of Logging Configuration requests,
	// the WAFv2 API only supports the following redacted fields.
	// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/14244
	if f.Method != nil {
		m["method"] = make([]map[string]interface{}, 1)
	}

	if f.QueryString != nil {
		m["query_string"] = make([]map[string]interface{}, 1)
	}

	if f.SingleHeader != nil {
		m["single_header"] = flattenWafv2SingleHeader(f.SingleHeader)
	}

	if f.UriPath != nil {
		m["uri_path"] = make([]map[string]interface{}, 1)
	}

	return m
}

// redactedFieldsHash takes a map[string]interface{} as input and generates a
// unique hashcode, taking into account keys defined in the resource's schema
// are present even if not explicitly configured
func redactedFieldsHash(v interface{}) int {
	var buf bytes.Buffer
	m, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}
	if v, ok := m["method"].([]interface{}); ok && len(v) > 0 {
		buf.WriteString("method-")
	}
	if v, ok := m["query_string"].([]interface{}); ok && len(v) > 0 {
		buf.WriteString("query_string-")
	}
	if v, ok := m["uri_path"].([]interface{}); ok && len(v) > 0 {
		buf.WriteString("uri_path-")
	}
	if v, ok := m["single_header"].([]interface{}); ok && len(v) > 0 {
		sh, ok := v[0].(map[string]interface{})
		if ok {
			if name, ok := sh["name"].(string); ok {
				buf.WriteString(fmt.Sprintf("%s-", name))
			}
		}
	}

	return hashcode.String(buf.String())
}

func suppressRedactedFieldsDiff(k, old, new string, d *schema.ResourceData) bool {
	o, n := d.GetChange("redacted_fields")
	oList := o.([]interface{})
	nList := n.([]interface{})

	if len(oList) == 0 && len(nList) == 0 {
		return true
	}

	if len(oList) == 0 && len(nList) != 0 {
		// account for empty block
		return nList[0] == nil
	}

	if len(oList) != 0 && len(nList) == 0 {
		// account for empty block
		return oList[0] == nil
	}

	oldSet := schema.NewSet(redactedFieldsHash, oList)
	newSet := schema.NewSet(redactedFieldsHash, nList)
	return oldSet.Equal(newSet)
}
