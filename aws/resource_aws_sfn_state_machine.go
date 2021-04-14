package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sfn/waiter"
)

func resourceAwsSfnStateMachine() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSfnStateMachineCreate,
		Read:   resourceAwsSfnStateMachineRead,
		Update: resourceAwsSfnStateMachineUpdate,
		Delete: resourceAwsSfnStateMachineDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"definition": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024*1024), // 1048576
			},

			"logging_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_destination": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"include_execution_data": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"level": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								sfn.LogLevelAll,
								sfn.LogLevelError,
								sfn.LogLevelFatal,
								sfn.LogLevelOff,
							}, false),
						},
					},
				},
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateSfnStateMachineName,
			},

			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},

			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  sfn.StateMachineTypeStandard,
				ValidateFunc: validation.StringInSlice([]string{
					sfn.StateMachineTypeStandard,
					sfn.StateMachineTypeExpress,
				}, false),
			},
		},
	}
}

func resourceAwsSfnStateMachineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn
	log.Print("[DEBUG] Creating Step Function State Machine")
	params := &sfn.CreateStateMachineInput{
		Definition:           aws.String(d.Get("definition").(string)),
		LoggingConfiguration: expandAwsSfnLoggingConfiguration(d.Get("logging_configuration").([]interface{})),
		Name:                 aws.String(d.Get("name").(string)),
		RoleArn:              aws.String(d.Get("role_arn").(string)),
		Tags:                 keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().SfnTags(),
		Type:                 aws.String(d.Get("type").(string)),
	}

	var stateMachine *sfn.CreateStateMachineOutput

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var err error
		stateMachine, err = conn.CreateStateMachine(params)

		if err != nil {
			// Note: the instance may be in a deleting mode, hence the retry
			// when creating the step function. This can happen when we are
			// updating the resource (since there is no update API call).
			if isAWSErr(err, sfn.ErrCodeStateMachineDeleting, "") {
				return resource.RetryableError(err)
			}
			//This is done to deal with IAM eventual consistency
			if isAWSErr(err, "AccessDeniedException", "") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		stateMachine, err = conn.CreateStateMachine(params)
	}

	if err != nil {
		return fmt.Errorf("Error creating Step Function State Machine: %s", err)
	}

	d.SetId(aws.StringValue(stateMachine.StateMachineArn))

	return resourceAwsSfnStateMachineRead(d, meta)
}

func resourceAwsSfnStateMachineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading Step Function State Machine: %s", d.Id())
	sm, err := conn.DescribeStateMachine(&sfn.DescribeStateMachineInput{
		StateMachineArn: aws.String(d.Id()),
	})
	if err != nil {

		if isAWSErr(err, sfn.ErrCodeStateMachineDoesNotExist, "") {
			log.Printf("[WARN] SFN State Machine (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("arn", sm.StateMachineArn)
	d.Set("definition", sm.Definition)
	d.Set("name", sm.Name)
	d.Set("role_arn", sm.RoleArn)
	d.Set("type", sm.Type)
	d.Set("status", sm.Status)

	loggingConfiguration := flattenAwsSfnLoggingConfiguration(sm.LoggingConfiguration)

	if err := d.Set("logging_configuration", loggingConfiguration); err != nil {
		log.Printf("[DEBUG] Error setting logging_configuration %s", err)
	}

	if err := d.Set("creation_date", sm.CreationDate.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting creation_date: %s", err)
	}

	tags, err := keyvaluetags.SfnListTags(conn, d.Id())

	if err != nil && !isAWSErr(err, "UnknownOperationException", "") {
		return fmt.Errorf("error listing tags for SFN State Machine (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSfnStateMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn

	if d.HasChanges("definition", "role_arn", "logging_configuration") {
		params := &sfn.UpdateStateMachineInput{
			StateMachineArn: aws.String(d.Id()),
			Definition:      aws.String(d.Get("definition").(string)),
			RoleArn:         aws.String(d.Get("role_arn").(string)),
		}

		if d.HasChange("logging_configuration") {
			params.LoggingConfiguration = expandAwsSfnLoggingConfiguration(d.Get("logging_configuration").([]interface{}))
		}

		_, err := conn.UpdateStateMachine(params)

		log.Printf("[DEBUG] Updating Step Function State Machine: %#v", params)

		if err != nil {
			if isAWSErr(err, sfn.ErrCodeStateMachineDoesNotExist, "State Machine Does Not Exist") {
				return fmt.Errorf("Error updating Step Function State Machine: %s", err)
			}
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.SfnUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsSfnStateMachineRead(d, meta)
}

func resourceAwsSfnStateMachineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn
	log.Printf("[DEBUG] Deleting Step Function State Machine: %s", d.Id())

	input := &sfn.DeleteStateMachineInput{
		StateMachineArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteStateMachine(input)

	if err != nil {
		return fmt.Errorf("Error deleting SFN state machine: %s", err)
	}

	if _, err := waiter.StateMachineDeleted(conn, d.Id()); err != nil {
		if isAWSErr(err, sfn.ErrCodeStateMachineDoesNotExist, "") {
			return nil
		}
		return fmt.Errorf("error waiting for SFN State Machine (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func expandAwsSfnLoggingConfiguration(l []interface{}) *sfn.LoggingConfiguration {

	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	loggingConfiguration := &sfn.LoggingConfiguration{
		Destinations: []*sfn.LogDestination{
			{
				CloudWatchLogsLogGroup: &sfn.CloudWatchLogsLogGroup{
					LogGroupArn: aws.String(m["log_destination"].(string)),
				},
			},
		},
		IncludeExecutionData: aws.Bool(m["include_execution_data"].(bool)),
		Level:                aws.String(m["level"].(string)),
	}

	return loggingConfiguration
}

func flattenAwsSfnLoggingConfiguration(loggingConfiguration *sfn.LoggingConfiguration) []interface{} {

	if loggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"log_destination":        "",
		"include_execution_data": aws.BoolValue(loggingConfiguration.IncludeExecutionData),
		"level":                  aws.StringValue(loggingConfiguration.Level),
	}

	if len(loggingConfiguration.Destinations) > 0 {
		m["log_destination"] = aws.StringValue(loggingConfiguration.Destinations[0].CloudWatchLogsLogGroup.LogGroupArn)
	}

	return []interface{}{m}
}
