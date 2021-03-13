package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for Config service eventual consistency on deletion
	configRemediationConfigurationDeletionTimeout = 2 * time.Minute
)

func resourceAwsConfigRemediationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsConfigRemediationConfigurationPut,
		Read:   resourceAwsConfigRemediationConfigurationRead,
		Update: resourceAwsConfigRemediationConfigurationPut,
		Delete: resourceAwsConfigRemediationConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config_rule_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"resource_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"target_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"target_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(configservice.RemediationTargetType_Values(), false),
			},
			"target_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"parameter": {
				Type:     schema.TypeSet,
				MaxItems: 25,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"resource_value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"static_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"automatic": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"execution_controls": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ssm_controls": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"concurrent_execution_rate_percentage": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
									"error_percentage": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
								},
							},
						},
					},
				},
			},
			"maximum_automatic_attempts": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 25),
			},
			"retry_attempt_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 2678000),
			},
		},
	}
}

func expandConfigRemediationConfigurationParameters(configured *schema.Set) (map[string]*configservice.RemediationParameterValue, error) {
	results := make(map[string]*configservice.RemediationParameterValue)

	for _, item := range configured.List() {
		detail := item.(map[string]interface{})
		rpv := configservice.RemediationParameterValue{}
		resourceName, ok := detail["name"].(string)
		if ok {
			results[resourceName] = &rpv
		} else {
			return nil, fmt.Errorf("Could not extract name from parameter.")
		}
		if resourceValue, ok := detail["resource_value"].(string); ok && len(resourceValue) > 0 {
			rpv.ResourceValue = &configservice.ResourceValue{
				Value: &resourceValue,
			}
		} else if staticValue, ok := detail["static_value"].(string); ok && len(staticValue) > 0 {
			rpv.StaticValue = &configservice.StaticValue{
				Values: []*string{&staticValue},
			}
		} else {
			return nil, fmt.Errorf("Parameter '%s' needs one of resource_value or static_value", resourceName)
		}
	}

	return results, nil
}

func expandConfigRemediationConfigurationExecutionControlsConfig(v map[string]interface{}) (ret *configservice.ExecutionControls, err error) {
	if w, ok := v["ssm_controls"]; ok {
		x := w.([]interface{})
		if len(x) > 0 {
			ssmControls, err := expandConfigRemediationConfigurationSsmControlsConfig(x[0].(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			ret = &configservice.ExecutionControls{
				SsmControls: ssmControls,
			}
			return ret, nil
		}
	}
	return nil, fmt.Errorf("expected 'ssm_controls' in execution controls configuration")
}

func expandConfigRemediationConfigurationSsmControlsConfig(v map[string]interface{}) (ret *configservice.SsmControls, err error) {
	ret = &configservice.SsmControls{}
	p := false
	if concurrentExecutionRatePercentage, ok := v["concurrent_execution_rate_percentage"]; ok {
		p = true
		ret.ConcurrentExecutionRatePercentage = aws.Int64(int64(concurrentExecutionRatePercentage.(int)))
	}
	if errorPercentage, ok := v["error_percentage"]; ok {
		p = true
		ret.ErrorPercentage = aws.Int64(int64(errorPercentage.(int)))
	}
	if !p {
		return nil, fmt.Errorf("'concurrent_execution_rate_percentage' or 'error_percentage' must be provided in ssm_controls")
	}
	return ret, nil
}

func flattenRemediationConfigurationParameters(parameters map[string]*configservice.RemediationParameterValue) []interface{} {
	var items []interface{}

	for key, value := range parameters {
		item := make(map[string]interface{})
		item["name"] = key
		if value.ResourceValue != nil {
			item["resource_value"] = *value.ResourceValue.Value
		}
		if value.StaticValue != nil && len(value.StaticValue.Values) > 0 {
			item["static_value"] = *value.StaticValue.Values[0]
		}

		items = append(items, item)
	}

	return items
}

func flattenConfigRemediationConfigurationExecutionControlsConfig(controls *configservice.ExecutionControls) []interface{} {
	if controls == nil {
		return nil
	}
	return []interface{}{map[string]interface{}{
		"ssm_controls": flattenConfigRemediationConfigurationSsmControlsConfig(controls.SsmControls),
	}}
}

func flattenConfigRemediationConfigurationSsmControlsConfig(controls *configservice.SsmControls) []interface{} {
	if controls == nil {
		return nil
	}
	m := make(map[string]interface{})
	if controls.ConcurrentExecutionRatePercentage != nil {
		m["concurrent_execution_rate_percentage"] = controls.ConcurrentExecutionRatePercentage
	}
	if controls.ErrorPercentage != nil {
		m["error_percentage"] = controls.ErrorPercentage
	}
	return []interface{}{m}
}

func resourceAwsConfigRemediationConfigurationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	name := d.Get("config_rule_name").(string)
	remediationConfigurationInput := configservice.RemediationConfiguration{
		ConfigRuleName: aws.String(name),
	}

	if v, ok := d.GetOk("parameter"); ok {
		params, err := expandConfigRemediationConfigurationParameters(v.(*schema.Set))
		if err != nil {
			return err
		}
		remediationConfigurationInput.Parameters = params
	}
	if v, ok := d.GetOk("resource_type"); ok {
		remediationConfigurationInput.ResourceType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target_id"); ok {
		remediationConfigurationInput.TargetId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target_type"); ok {
		remediationConfigurationInput.TargetType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target_version"); ok {
		remediationConfigurationInput.TargetVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("automatic"); ok {
		remediationConfigurationInput.Automatic = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("maximum_automatic_attempts"); ok {
		remediationConfigurationInput.MaximumAutomaticAttempts = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("retry_attempt_seconds"); ok {
		remediationConfigurationInput.RetryAttemptSeconds = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("execution_controls"); ok {
		executionControlsConfigs := v.([]interface{})
		if len(executionControlsConfigs) == 1 {
			w := executionControlsConfigs[0].(map[string]interface{})
			controls, err := expandConfigRemediationConfigurationExecutionControlsConfig(w)
			if err != nil {
				return err
			}
			remediationConfigurationInput.ExecutionControls = controls
		}
	}
	input := configservice.PutRemediationConfigurationsInput{
		RemediationConfigurations: []*configservice.RemediationConfiguration{&remediationConfigurationInput},
	}
	log.Printf("[DEBUG] Creating AWSConfig remediation configuration: %s", input)
	_, err := conn.PutRemediationConfigurations(&input)
	if err != nil {
		return fmt.Errorf("Failed to create AWSConfig remediation configuration: %w", err)
	}

	d.SetId(name)

	log.Printf("[DEBUG] AWSConfig config remediation configuration for rule %q created", name)

	return resourceAwsConfigRemediationConfigurationRead(d, meta)
}

func resourceAwsConfigRemediationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn
	out, err := conn.DescribeRemediationConfigurations(&configservice.DescribeRemediationConfigurationsInput{
		ConfigRuleNames: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if isAWSErr(err, configservice.ErrCodeNoSuchConfigRuleException, "") {
			log.Printf("[WARN] Config Rule %q is gone (NoSuchConfigRuleException)", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	numberOfRemediationConfigurations := len(out.RemediationConfigurations)
	if numberOfRemediationConfigurations < 1 {
		log.Printf("[WARN] No Remediation Configuration for Config Rule %q (no remediation configuration found)", d.Id())
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] AWS Config remediation configurations received: %s", out)

	remediationConfiguration := out.RemediationConfigurations[0]
	d.Set("arn", remediationConfiguration.Arn)
	d.Set("config_rule_name", remediationConfiguration.ConfigRuleName)
	d.Set("resource_type", remediationConfiguration.ResourceType)
	d.Set("target_id", remediationConfiguration.TargetId)
	d.Set("target_type", remediationConfiguration.TargetType)
	d.Set("target_version", remediationConfiguration.TargetVersion)
	d.Set("parameter", flattenRemediationConfigurationParameters(remediationConfiguration.Parameters))
	d.Set("automatic", remediationConfiguration.Automatic)
	d.Set("maximum_automatic_attempts", remediationConfiguration.MaximumAutomaticAttempts)
	d.Set("retry_attempt_seconds", remediationConfiguration.RetryAttemptSeconds)
	d.Set("maximum_automatic_attempts", remediationConfiguration.MaximumAutomaticAttempts)
	d.Set("execution_controls", flattenConfigRemediationConfigurationExecutionControlsConfig(remediationConfiguration.ExecutionControls))
	d.SetId(aws.StringValue(remediationConfiguration.ConfigRuleName))

	return nil
}

func resourceAwsConfigRemediationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	name := d.Get("config_rule_name").(string)

	input := &configservice.DeleteRemediationConfigurationInput{
		ConfigRuleName: aws.String(name),
	}

	if v, ok := d.GetOk("resource_type"); ok {
		input.ResourceType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting AWS Config remediation configurations for rule %q", name)
	err := resource.Retry(configRemediationConfigurationDeletionTimeout, func() *resource.RetryError {
		_, err := conn.DeleteRemediationConfiguration(input)

		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeResourceInUseException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteRemediationConfiguration(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting Config Remediation Configuration (%s): %w", d.Id(), err)
	}

	return nil
}
