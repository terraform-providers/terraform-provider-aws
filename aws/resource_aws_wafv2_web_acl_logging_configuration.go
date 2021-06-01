package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    100,
				Elem:        wafv2FieldToMatchBaseSchema(),
				Description: "Parts of the request to exclude from logs",
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

	if v, ok := d.GetOk("redacted_fields"); ok && v.(*schema.Set).Len() > 0 {
		config.RedactedFields = expandWafv2RedactedFields(v.(*schema.Set).List())
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

func flattenWafv2RedactedFields(fields []*wafv2.FieldToMatch) []map[string]interface{} {
	redactedFields := make([]map[string]interface{}, 0, len(fields))
	for _, field := range fields {
		redactedFields = append(redactedFields, flattenWafv2FieldToMatch(field).([]interface{})[0].(map[string]interface{}))
	}
	return redactedFields
}

func expandWafv2RedactedFields(fields []interface{}) []*wafv2.FieldToMatch {
	redactedFields := make([]*wafv2.FieldToMatch, 0, len(fields))
	for _, field := range fields {
		redactedFields = append(redactedFields, expandWafv2FieldToMatch([]interface{}{field}))
	}
	return redactedFields
}
