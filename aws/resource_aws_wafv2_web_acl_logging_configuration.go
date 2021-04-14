package aws

import (
	"bytes"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
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
	if err != nil {
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAFv2 Logging Configuration for WebACL with ARN %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading WAFv2 Logging Configuration for resource (%s): %w", d.Id(), err)
	}
	if output == nil || output.LoggingConfiguration == nil {
		return fmt.Errorf("error reading WAFv2 Logging Configuration for resource (%s): empty response", d.Id())
	}

	if err := d.Set("log_destination_configs", flattenStringList(output.LoggingConfiguration.LogDestinationConfigs)); err != nil {
		return fmt.Errorf("error setting log_destination_configs: %w", err)
	}

	if err := d.Set("redacted_fields", flattenWafv2RedactedFields(output.LoggingConfiguration.RedactedFields)); err != nil {
		return fmt.Errorf("error setting redacted_fields: %w", err)
	}

	d.Set("resource_arn", output.LoggingConfiguration.ResourceArn)

	return nil
}

func resourceAwsWafv2WebACLLoggingConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	input := &wafv2.DeleteLoggingConfigurationInput{
		ResourceArn: aws.String(d.Id()),
	}
	_, err := conn.DeleteLoggingConfiguration(input)
	if err != nil {
		return fmt.Errorf("error deleting WAFv2 Logging Configuration for resource (%s): %w", d.Id(), err)
	}

	return nil
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
