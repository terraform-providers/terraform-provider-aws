package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsConfigOrganizationCustomRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsConfigOrganizationCustomRuleCreate,
		Delete: resourceAwsConfigOrganizationCustomRuleDelete,
		Read:   resourceAwsConfigOrganizationCustomRuleRead,
		Update: resourceAwsConfigOrganizationCustomRuleUpdate,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"excluded_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1000,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateAwsAccountId,
				},
			},
			"input_parameters": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 2048),
					validation.StringIsJSON,
				),
			},
			"lambda_function_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"maximum_execution_frequency": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					configservice.MaximumExecutionFrequencyOneHour,
					configservice.MaximumExecutionFrequencyThreeHours,
					configservice.MaximumExecutionFrequencySixHours,
					configservice.MaximumExecutionFrequencyTwelveHours,
					configservice.MaximumExecutionFrequencyTwentyFourHours,
				}, false),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"resource_id_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 768),
			},
			"resource_types_scope": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 256),
				},
			},
			"tag_key_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"tag_value_scope": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"trigger_types": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 3,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"ConfigurationItemChangeNotification",
						"OversizedConfigurationItemChangeNotification",
						"ScheduledNotification",
					}, false),
				},
			},
		},
	}
}

func resourceAwsConfigOrganizationCustomRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn
	name := d.Get("name").(string)

	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(name),
		OrganizationCustomRuleMetadata: &configservice.OrganizationCustomRuleMetadata{
			LambdaFunctionArn:                  aws.String(d.Get("lambda_function_arn").(string)),
			OrganizationConfigRuleTriggerTypes: expandStringSet(d.Get("trigger_types").(*schema.Set)),
		},
	}

	if v, ok := d.GetOk("description"); ok {
		input.OrganizationCustomRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationCustomRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationCustomRuleMetadata.MaximumExecutionFrequency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationCustomRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationCustomRuleMetadata.ResourceTypesScope = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConfigRule(input)

	if err != nil {
		return fmt.Errorf("error creating Config Organization Custom Rule (%s): %s", name, err)
	}

	d.SetId(name)

	if err := configWaitForOrganizationRuleStatusCreateSuccessful(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Config Organization Custom Rule (%s) creation: %s", d.Id(), err)
	}

	return resourceAwsConfigOrganizationCustomRuleRead(d, meta)
}

func resourceAwsConfigOrganizationCustomRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	rule, err := configDescribeOrganizationConfigRule(conn, d.Id())

	if isAWSErr(err, configservice.ErrCodeNoSuchOrganizationConfigRuleException, "") {
		log.Printf("[WARN] Config Organization Custom Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Config Organization Custom Rule (%s): %s", d.Id(), err)
	}

	if rule == nil {
		log.Printf("[WARN] Config Organization Custom Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if rule.OrganizationManagedRuleMetadata != nil {
		return fmt.Errorf("expected Config Organization Custom Rule, found Config Organization Custom Rule: %s", d.Id())
	}

	if rule.OrganizationCustomRuleMetadata == nil {
		return fmt.Errorf("error describing Config Organization Custom Rule (%s): empty metadata", d.Id())
	}

	d.Set("arn", rule.OrganizationConfigRuleArn)
	d.Set("description", rule.OrganizationCustomRuleMetadata.Description)

	if err := d.Set("excluded_accounts", aws.StringValueSlice(rule.ExcludedAccounts)); err != nil {
		return fmt.Errorf("error setting excluded_accounts: %s", err)
	}

	d.Set("input_parameters", rule.OrganizationCustomRuleMetadata.InputParameters)
	d.Set("lambda_function_arn", rule.OrganizationCustomRuleMetadata.LambdaFunctionArn)
	d.Set("maximum_execution_frequency", rule.OrganizationCustomRuleMetadata.MaximumExecutionFrequency)
	d.Set("name", rule.OrganizationConfigRuleName)
	d.Set("resource_id_scope", rule.OrganizationCustomRuleMetadata.ResourceIdScope)

	if err := d.Set("resource_types_scope", aws.StringValueSlice(rule.OrganizationCustomRuleMetadata.ResourceTypesScope)); err != nil {
		return fmt.Errorf("error setting resource_types_scope: %s", err)
	}

	d.Set("tag_key_scope", rule.OrganizationCustomRuleMetadata.TagKeyScope)
	d.Set("tag_value_scope", rule.OrganizationCustomRuleMetadata.TagValueScope)

	if err := d.Set("trigger_types", aws.StringValueSlice(rule.OrganizationCustomRuleMetadata.OrganizationConfigRuleTriggerTypes)); err != nil {
		return fmt.Errorf("error setting trigger_types: %s", err)
	}

	return nil
}

func resourceAwsConfigOrganizationCustomRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	input := &configservice.PutOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
		OrganizationCustomRuleMetadata: &configservice.OrganizationCustomRuleMetadata{
			LambdaFunctionArn:                  aws.String(d.Get("lambda_function_arn").(string)),
			OrganizationConfigRuleTriggerTypes: expandStringSet(d.Get("trigger_types").(*schema.Set)),
		},
	}

	if v, ok := d.GetOk("description"); ok {
		input.OrganizationCustomRuleMetadata.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excluded_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludedAccounts = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("input_parameters"); ok {
		input.OrganizationCustomRuleMetadata.InputParameters = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_execution_frequency"); ok {
		input.OrganizationCustomRuleMetadata.MaximumExecutionFrequency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_id_scope"); ok {
		input.OrganizationCustomRuleMetadata.ResourceIdScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_types_scope"); ok && v.(*schema.Set).Len() > 0 {
		input.OrganizationCustomRuleMetadata.ResourceTypesScope = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_key_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagKeyScope = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag_value_scope"); ok {
		input.OrganizationCustomRuleMetadata.TagValueScope = aws.String(v.(string))
	}

	_, err := conn.PutOrganizationConfigRule(input)

	if err != nil {
		return fmt.Errorf("error updating Config Organization Custom Rule (%s): %s", d.Id(), err)
	}

	if err := configWaitForOrganizationRuleStatusUpdateSuccessful(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("error waiting for Config Organization Custom Rule (%s) update: %s", d.Id(), err)
	}

	return resourceAwsConfigOrganizationCustomRuleRead(d, meta)
}

func resourceAwsConfigOrganizationCustomRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	input := &configservice.DeleteOrganizationConfigRuleInput{
		OrganizationConfigRuleName: aws.String(d.Id()),
	}

	_, err := conn.DeleteOrganizationConfigRule(input)

	if err != nil {
		return fmt.Errorf("error deleting Config Organization Custom Rule (%s): %s", d.Id(), err)
	}

	if err := configWaitForOrganizationRuleStatusDeleteSuccessful(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Config Organization Custom Rule (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
