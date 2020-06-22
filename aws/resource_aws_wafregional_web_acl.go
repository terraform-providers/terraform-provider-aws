package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsWafRegionalWebAcl() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafRegionalWebAclCreate,
		Read:   resourceAwsWafRegionalWebAclRead,
		Update: resourceAwsWafRegionalWebAclUpdate,
		Delete: resourceAwsWafRegionalWebAclDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_action": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								waf.WafActionTypeAllow,
								waf.WafActionTypeBlock,
								waf.WafActionTypeCount,
							}, false),
						},
					},
				},
			},
			"logging_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_destination": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"redacted_fields": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field_to_match": {
										Type:     schema.TypeSet,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"data": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"type": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.StringInSlice([]string{
														waf.MatchFieldTypeAllQueryArgs,
														waf.MatchFieldTypeBody,
														waf.MatchFieldTypeHeader,
														waf.MatchFieldTypeMethod,
														waf.MatchFieldTypeQueryString,
														waf.MatchFieldTypeSingleQueryArg,
														waf.MatchFieldTypeUri,
													}, false),
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
			"metric_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											waf.WafActionTypeAllow,
											waf.WafActionTypeBlock,
											waf.WafActionTypeCount,
										}, false),
									},
								},
							},
						},
						"override_action": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											waf.WafOverrideActionTypeCount,
											waf.WafOverrideActionTypeNone,
										}, false),
									},
								},
							},
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  waf.WafRuleTypeRegular,
							ValidateFunc: validation.StringInSlice([]string{
								waf.WafRuleTypeRegular,
								waf.WafRuleTypeRateBased,
								waf.WafRuleTypeGroup,
							}, false),
						},
						"rule_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsWafRegionalWebAclCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().WafregionalTags()

	wr := newWafRegionalRetryer(conn, region)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateWebACLInput{
			ChangeToken:   token,
			DefaultAction: expandWafAction(d.Get("default_action").([]interface{})),
			MetricName:    aws.String(d.Get("metric_name").(string)),
			Name:          aws.String(d.Get("name").(string)),
		}

		if len(tags) > 0 {
			params.Tags = tags
		}

		return conn.CreateWebACL(params)
	})
	if err != nil {
		return err
	}
	resp := out.(*waf.CreateWebACLOutput)
	d.SetId(*resp.WebACL.WebACLId)

	// The WAF API currently omits this, but use it when it becomes available
	webACLARN := aws.StringValue(resp.WebACL.WebACLArn)
	if webACLARN == "" {
		webACLARN = arn.ARN{
			AccountID: meta.(*AWSClient).accountid,
			Partition: meta.(*AWSClient).partition,
			Region:    meta.(*AWSClient).region,
			Resource:  fmt.Sprintf("webacl/%s", d.Id()),
			Service:   "waf-regional",
		}.String()
	}

	loggingConfiguration := d.Get("logging_configuration").([]interface{})

	if len(loggingConfiguration) == 1 {
		input := &waf.PutLoggingConfigurationInput{
			LoggingConfiguration: expandWAFRegionalLoggingConfiguration(loggingConfiguration, webACLARN),
		}

		log.Printf("[DEBUG] Updating WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), input)
		if _, err := conn.PutLoggingConfiguration(input); err != nil {
			return fmt.Errorf("error Updating WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), err)
		}
	}

	rules := d.Get("rule").(*schema.Set).List()
	if len(rules) > 0 {
		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateWebACLInput{
				ChangeToken:   token,
				DefaultAction: expandWafAction(d.Get("default_action").([]interface{})),
				Updates:       diffWafWebAclRules([]interface{}{}, rules),
				WebACLId:      aws.String(d.Id()),
			}
			return conn.UpdateWebACL(req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF Regional ACL: %s", err)
		}
	}

	return resourceAwsWafRegionalWebAclRead(d, meta)
}

func resourceAwsWafRegionalWebAclRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &waf.GetWebACLInput{
		WebACLId: aws.String(d.Id()),
	}

	resp, err := conn.GetWebACL(params)
	if err != nil {
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAF Regional ACL (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	if resp == nil || resp.WebACL == nil {
		log.Printf("[WARN] WAF Regional ACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	// The WAF API currently omits this, but use it when it becomes available
	webACLARN := aws.StringValue(resp.WebACL.WebACLArn)
	if webACLARN == "" {
		webACLARN = arn.ARN{
			AccountID: meta.(*AWSClient).accountid,
			Partition: meta.(*AWSClient).partition,
			Region:    meta.(*AWSClient).region,
			Resource:  fmt.Sprintf("webacl/%s", d.Id()),
			Service:   "waf-regional",
		}.String()
	}
	d.Set("arn", webACLARN)

	if err := d.Set("default_action", flattenWafAction(resp.WebACL.DefaultAction)); err != nil {
		return fmt.Errorf("error setting default_action: %s", err)
	}
	d.Set("name", resp.WebACL.Name)
	d.Set("metric_name", resp.WebACL.MetricName)
	if err := d.Set("rule", flattenWafWebAclRules(resp.WebACL.Rules)); err != nil {
		return fmt.Errorf("error setting rule: %s", err)
	}

	tags, err := keyvaluetags.WafregionalListTags(conn, webACLARN)
	if err != nil {
		return fmt.Errorf("error listing tags for WAF Regional ACL (%s): %s", webACLARN, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	getLoggingConfigurationInput := &waf.GetLoggingConfigurationInput{
		ResourceArn: aws.String(d.Get("arn").(string)),
	}
	loggingConfiguration := []interface{}{}

	log.Printf("[DEBUG] Getting WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), getLoggingConfigurationInput)
	getLoggingConfigurationOutput, err := conn.GetLoggingConfiguration(getLoggingConfigurationInput)

	if err != nil && !isAWSErr(err, waf.ErrCodeNonexistentItemException, "") {
		return fmt.Errorf("error getting WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), err)
	}

	if getLoggingConfigurationOutput != nil {
		loggingConfiguration = flattenWAFRegionalLoggingConfiguration(getLoggingConfigurationOutput.LoggingConfiguration)
	}

	if err := d.Set("logging_configuration", loggingConfiguration); err != nil {
		return fmt.Errorf("error setting logging_configuration: %s", err)
	}

	return nil
}

func resourceAwsWafRegionalWebAclUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	if d.HasChanges("default_action", "rule") {
		o, n := d.GetChange("rule")
		oldR, newR := o.(*schema.Set).List(), n.(*schema.Set).List()

		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateWebACLInput{
				ChangeToken:   token,
				DefaultAction: expandWafAction(d.Get("default_action").([]interface{})),
				Updates:       diffWafWebAclRules(oldR, newR),
				WebACLId:      aws.String(d.Id()),
			}
			return conn.UpdateWebACL(req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF Regional ACL: %s", err)
		}
	}

	if d.HasChange("logging_configuration") {
		loggingConfiguration := d.Get("logging_configuration").([]interface{})

		if len(loggingConfiguration) == 1 {
			input := &waf.PutLoggingConfigurationInput{
				LoggingConfiguration: expandWAFRegionalLoggingConfiguration(loggingConfiguration, d.Get("arn").(string)),
			}

			log.Printf("[DEBUG] Updating WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), input)
			if _, err := conn.PutLoggingConfiguration(input); err != nil {
				return fmt.Errorf("error updating WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), err)
			}
		} else {
			input := &waf.DeleteLoggingConfigurationInput{
				ResourceArn: aws.String(d.Get("arn").(string)),
			}

			log.Printf("[DEBUG] Deleting WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), input)
			if _, err := conn.DeleteLoggingConfiguration(input); err != nil {
				return fmt.Errorf("error deleting WAF Regional Web ACL (%s) Logging Configuration: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.WafregionalUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWafRegionalWebAclRead(d, meta)
}

func resourceAwsWafRegionalWebAclDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	// First, need to delete all rules
	rules := d.Get("rule").(*schema.Set).List()
	if len(rules) > 0 {
		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateWebACLInput{
				ChangeToken:   token,
				DefaultAction: expandWafAction(d.Get("default_action").([]interface{})),
				Updates:       diffWafWebAclRules(rules, []interface{}{}),
				WebACLId:      aws.String(d.Id()),
			}
			return conn.UpdateWebACL(req)
		})
		if err != nil {
			return fmt.Errorf("Error Removing WAF Regional ACL Rules: %s", err)
		}
	}

	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteWebACLInput{
			ChangeToken: token,
			WebACLId:    aws.String(d.Id()),
		}

		log.Printf("[INFO] Deleting WAF ACL")
		return conn.DeleteWebACL(req)
	})
	if err != nil {
		return fmt.Errorf("Error Deleting WAF Regional ACL: %s", err)
	}
	return nil
}

func expandWAFRegionalLoggingConfiguration(l []interface{}, resourceARN string) *waf.LoggingConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	loggingConfiguration := &waf.LoggingConfiguration{
		LogDestinationConfigs: []*string{
			aws.String(m["log_destination"].(string)),
		},
		RedactedFields: expandWAFRegionalRedactedFields(m["redacted_fields"].([]interface{})),
		ResourceArn:    aws.String(resourceARN),
	}

	return loggingConfiguration
}

func expandWAFRegionalRedactedFields(l []interface{}) []*waf.FieldToMatch {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	if m["field_to_match"] == nil {
		return nil
	}

	redactedFields := make([]*waf.FieldToMatch, 0)

	for _, fieldToMatch := range m["field_to_match"].(*schema.Set).List() {
		if fieldToMatch == nil {
			continue
		}

		redactedFields = append(redactedFields, expandFieldToMatch(fieldToMatch.(map[string]interface{})))
	}

	return redactedFields
}

func flattenWAFRegionalLoggingConfiguration(loggingConfiguration *waf.LoggingConfiguration) []interface{} {
	if loggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"log_destination": "",
		"redacted_fields": flattenWAFRegionalRedactedFields(loggingConfiguration.RedactedFields),
	}

	if len(loggingConfiguration.LogDestinationConfigs) > 0 {
		m["log_destination"] = aws.StringValue(loggingConfiguration.LogDestinationConfigs[0])
	}

	return []interface{}{m}
}

func flattenWAFRegionalRedactedFields(fieldToMatches []*waf.FieldToMatch) []interface{} {
	if len(fieldToMatches) == 0 {
		return []interface{}{}
	}

	fieldToMatchResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
	l := make([]interface{}, len(fieldToMatches))

	for i, fieldToMatch := range fieldToMatches {
		l[i] = flattenFieldToMatch(fieldToMatch)[0]
	}

	m := map[string]interface{}{
		"field_to_match": schema.NewSet(schema.HashResource(fieldToMatchResource), l),
	}

	return []interface{}{m}
}
