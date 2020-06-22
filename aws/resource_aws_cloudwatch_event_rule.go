package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	cloudWatchEventRuleDeleteRetryTimeout = 5 * time.Minute
)

func resourceAwsCloudWatchEventRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchEventRuleCreate,
		Read:   resourceAwsCloudWatchEventRuleRead,
		Update: resourceAwsCloudWatchEventRuleUpdate,
		Delete: resourceAwsCloudWatchEventRuleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateCloudWatchEventRuleName,
			},
			"name_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateCloudWatchEventRuleName,
			},
			"schedule_expression": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"event_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateEventPatternValue(),
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v.(string))
					return json
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsCloudWatchEventRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	input, err := buildPutRuleInputStruct(d, name)
	if err != nil {
		return fmt.Errorf("Creating CloudWatch Event Rule failed: %s", err)
	}
	log.Printf("[DEBUG] Creating CloudWatch Event Rule: %s", input)

	// IAM Roles take some time to propagate
	var out *events.PutRuleOutput
	err = resource.Retry(30*time.Second, func() *resource.RetryError {
		out, err = conn.PutRule(input)

		if isAWSErr(err, "ValidationException", "cannot be assumed by principal") {
			log.Printf("[DEBUG] Retrying update of CloudWatch Event Rule %q", *input.Name)
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.PutRule(input)
	}

	if err != nil {
		return fmt.Errorf("Updating CloudWatch Event Rule failed: %s", err)
	}

	d.Set("arn", out.RuleArn)
	d.SetId(*input.Name)

	log.Printf("[INFO] CloudWatch Event Rule %q created", *out.RuleArn)

	return resourceAwsCloudWatchEventRuleRead(d, meta)
}

func resourceAwsCloudWatchEventRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := events.DescribeRuleInput{
		Name: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading CloudWatch Event Rule: %s", input)
	out, err := conn.DescribeRule(&input)
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == events.ErrCodeResourceNotFoundException {
			log.Printf("[WARN] Removing CloudWatch Event Rule %q because it's gone.", d.Id())
			d.SetId("")
			return nil
		}
	}
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Found Event Rule: %s", out)

	arn := *out.Arn
	d.Set("arn", arn)
	d.Set("description", out.Description)
	if out.EventPattern != nil {
		pattern, err := structure.NormalizeJsonString(*out.EventPattern)
		if err != nil {
			return fmt.Errorf("event pattern contains an invalid JSON: %s", err)
		}
		d.Set("event_pattern", pattern)
	}
	d.Set("name", out.Name)
	d.Set("role_arn", out.RoleArn)
	d.Set("schedule_expression", out.ScheduleExpression)

	boolState, err := getBooleanStateFromString(*out.State)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Setting boolean state: %t", boolState)
	d.Set("is_enabled", boolState)

	tags, err := keyvaluetags.CloudwatcheventsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for CloudWatch Event Rule (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsCloudWatchEventRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	input, err := buildPutRuleInputStruct(d, d.Id())
	if err != nil {
		return fmt.Errorf("Updating CloudWatch Event Rule failed: %s", err)
	}
	log.Printf("[DEBUG] Updating CloudWatch Event Rule: %s", input)

	// IAM Roles take some time to propagate
	err = resource.Retry(30*time.Second, func() *resource.RetryError {
		_, err := conn.PutRule(input)

		if isAWSErr(err, "ValidationException", "cannot be assumed by principal") {
			log.Printf("[DEBUG] Retrying update of CloudWatch Event Rule %q", *input.Name)
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.PutRule(input)
	}

	if err != nil {
		return fmt.Errorf("Updating CloudWatch Event Rule failed: %s", err)
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.CloudwatcheventsUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating CloudwWatch Event Rule (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsCloudWatchEventRuleRead(d, meta)
}

func resourceAwsCloudWatchEventRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	input := &events.DeleteRuleInput{
		Name: aws.String(d.Id()),
	}

	err := resource.Retry(cloudWatchEventRuleDeleteRetryTimeout, func() *resource.RetryError {
		_, err := conn.DeleteRule(input)

		if isAWSErr(err, "ValidationException", "Rule can't be deleted since it has targets") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteRule(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudWatch Event Rule (%s): %s", d.Id(), err)
	}

	return nil
}

func buildPutRuleInputStruct(d *schema.ResourceData, name string) (*events.PutRuleInput, error) {
	input := events.PutRuleInput{
		Name: aws.String(name),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("event_pattern"); ok {
		pattern, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("event pattern contains an invalid JSON: %s", err)
		}
		input.EventPattern = aws.String(pattern)
	}
	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("schedule_expression"); ok {
		input.ScheduleExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CloudwatcheventsTags()
	}

	input.State = aws.String(getStringStateFromBoolean(d.Get("is_enabled").(bool)))

	return &input, nil
}

// State is represented as (ENABLED|DISABLED) in the API
func getBooleanStateFromString(state string) (bool, error) {
	if state == events.RuleStateEnabled {
		return true, nil
	} else if state == events.RuleStateDisabled {
		return false, nil
	}
	// We don't just blindly trust AWS as they tend to return
	// unexpected values in similar cases (different casing etc.)
	return false, fmt.Errorf("Failed converting state %q into boolean", state)
}

// State is represented as (ENABLED|DISABLED) in the API
func getStringStateFromBoolean(isEnabled bool) string {
	if isEnabled {
		return events.RuleStateEnabled
	}
	return events.RuleStateDisabled
}

func validateEventPatternValue() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		json, err := structure.NormalizeJsonString(v)
		if err != nil {
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))

			// Invalid JSON? Return immediately,
			// there is no need to collect other
			// errors.
			return
		}

		// Check whether the normalized JSON is within the given length.
		if len(json) > 2048 {
			errors = append(errors, fmt.Errorf(
				"%q cannot be longer than %d characters: %q", k, 2048, json))
		}
		return
	}
}
