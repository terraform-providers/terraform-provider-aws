package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	tfevents "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents/finder"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateCloudWatchEventRuleName,
			},
			"schedule_expression": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
				AtLeastOneOf: []string{"schedule_expression", "event_pattern"},
			},
			"event_bus_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateCloudWatchEventBusName,
				Default:      tfevents.DefaultEventBusName,
			},
			"event_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateEventPatternValue(),
				AtLeastOneOf: []string{"schedule_expression", "event_pattern"},
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

	name := naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))

	input, err := buildPutRuleInputStruct(d, name)
	if err != nil {
		return fmt.Errorf("Creating CloudWatch Events Rule failed: %w", err)
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CloudwatcheventsTags()
	}

	log.Printf("[DEBUG] Creating CloudWatch Events Rule: %s", input)

	// IAM Roles take some time to propagate
	var out *events.PutRuleOutput
	err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		out, err = conn.PutRule(input)

		if isAWSErr(err, "ValidationException", "cannot be assumed by principal") {
			log.Printf("[DEBUG] Retrying update of CloudWatch Events Rule %q", aws.StringValue(input.Name))
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
		return fmt.Errorf("Creating CloudWatch Events Rule failed: %w", err)
	}

	d.Set("arn", out.RuleArn)

	id := tfevents.RuleCreateID(aws.StringValue(input.EventBusName), aws.StringValue(input.Name))
	d.SetId(id)

	log.Printf("[INFO] CloudWatch Events Rule (%s) created", aws.StringValue(out.RuleArn))

	return resourceAwsCloudWatchEventRuleRead(d, meta)
}

func resourceAwsCloudWatchEventRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	out, err := finder.RuleByID(conn, d.Id())
	if tfawserr.ErrCodeEquals(err, events.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Removing CloudWatch Events Rule (%s) because it's gone.", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading CloudWatch Events Rule (%s): %w", d.Id(), err)
	}
	log.Printf("[DEBUG] Found Event Rule: %s", out)

	arn := aws.StringValue(out.Arn)
	d.Set("arn", arn)
	d.Set("description", out.Description)
	if out.EventPattern != nil {
		pattern, err := structure.NormalizeJsonString(aws.StringValue(out.EventPattern))
		if err != nil {
			return fmt.Errorf("event pattern contains an invalid JSON: %w", err)
		}
		d.Set("event_pattern", pattern)
	}
	d.Set("name", out.Name)
	d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(out.Name)))
	d.Set("role_arn", out.RoleArn)
	d.Set("schedule_expression", out.ScheduleExpression)
	d.Set("event_bus_name", out.EventBusName)

	boolState, err := getBooleanStateFromString(*out.State)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Setting boolean state: %t", boolState)
	d.Set("is_enabled", boolState)

	tags, err := keyvaluetags.CloudwatcheventsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for CloudWatch Events Rule (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsCloudWatchEventRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	_, ruleName, err := tfevents.RuleParseID(d.Id())
	if err != nil {
		return err
	}
	input, err := buildPutRuleInputStruct(d, ruleName)
	if err != nil {
		return fmt.Errorf("Updating CloudWatch Events Rule (%s) failed: %w", ruleName, err)
	}
	log.Printf("[DEBUG] Updating CloudWatch Events Rule: %s", input)

	// IAM Roles take some time to propagate
	err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		_, err := conn.PutRule(input)

		if isAWSErr(err, "ValidationException", "cannot be assumed by principal") {
			log.Printf("[DEBUG] Retrying update of CloudWatch Events Rule %q", aws.StringValue(input.Name))
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
		return fmt.Errorf("Updating CloudWatch Events Rule (%s) failed: %w", ruleName, err)
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.CloudwatcheventsUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating CloudwWatch Event Rule (%s) tags: %w", arn, err)
		}
	}

	return resourceAwsCloudWatchEventRuleRead(d, meta)
}

func resourceAwsCloudWatchEventRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	busName, ruleName, err := tfevents.RuleParseID(d.Id())
	if err != nil {
		return err
	}
	input := &events.DeleteRuleInput{
		Name:         aws.String(ruleName),
		EventBusName: aws.String(busName),
	}

	err = resource.Retry(cloudWatchEventRuleDeleteRetryTimeout, func() *resource.RetryError {
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
		return fmt.Errorf("error deleting CloudWatch Events Rule (%s): %w", d.Id(), err)
	}

	return nil
}

func buildPutRuleInputStruct(d *schema.ResourceData, name string) (*events.PutRuleInput, error) {
	input := events.PutRuleInput{
		Name: aws.String(name),
	}
	var eventBusName string
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("event_bus_name"); ok {
		eventBusName = v.(string)
		input.EventBusName = aws.String(eventBusName)
	}
	if v, ok := d.GetOk("event_pattern"); ok {
		pattern, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("event pattern contains an invalid JSON: %w", err)
		}
		input.EventPattern = aws.String(pattern)
	}
	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("schedule_expression"); ok {
		input.ScheduleExpression = aws.String(v.(string))
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
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %w", k, err))

			// Invalid JSON? Return immediately,
			// there is no need to collect other
			// errors.
			return
		}

		// Check whether the normalized JSON is within the given length.
		if len(json) > 2048 {
			errors = append(errors, fmt.Errorf("%q cannot be longer than %d characters: %q", k, 2048, json))
		}
		return
	}
}
