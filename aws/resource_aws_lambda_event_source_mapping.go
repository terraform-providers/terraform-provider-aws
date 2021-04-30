package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lambda/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lambda/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsLambdaEventSourceMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLambdaEventSourceMappingCreate,
		Read:   resourceAwsLambdaEventSourceMappingRead,
		Update: resourceAwsLambdaEventSourceMappingUpdate,
		Delete: resourceAwsLambdaEventSourceMappingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"event_source_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Using function name or ARN should not be shown as a diff.
					// Try to convert the old and new values from ARN to function name
					oldFunctionName, oldFunctionNameErr := getFunctionNameFromLambdaArn(old)
					newFunctionName, newFunctionNameErr := getFunctionNameFromLambdaArn(new)
					return (oldFunctionName == new && oldFunctionNameErr == nil) || (newFunctionName == old && newFunctionNameErr == nil)
				},
			},
			"starting_position": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(lambda.EventSourcePosition_Values(), false),
			},
			"starting_position_timestamp": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"topics": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"batch_size": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// When AWS repurposed EventSourceMapping for use with SQS they kept
					// the default for BatchSize at 100 for Kinesis and DynamoDB, but made
					// the default 10 for SQS.  As such, we had to make batch_size optional.
					// Because of this, we need to ensure that if someone doesn't have
					// batch_size specified that it is not treated as a diff for those
					if new != "" && new != "0" {
						return false
					}

					eventSourceARN, err := arn.Parse(d.Get("event_source_arn").(string))
					if err != nil {
						return false
					}
					switch eventSourceARN.Service {
					// kafka.ServiceName is "Kafka".
					case dynamodb.ServiceName, kinesis.ServiceName, "kafka":
						if old == "100" {
							return true
						}
					case sqs.ServiceName:
						if old == "10" {
							return true
						}
					}
					return false
				},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"maximum_batching_window_in_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"parallelization_factor": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 10),
				Computed:     true,
			},
			"maximum_retry_attempts": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(-1, 10_000),
			},
			"maximum_record_age_in_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.Any(
					validation.IntInSlice([]int{-1}),
					validation.IntBetween(60, 604_800),
				),
			},
			"bisect_batch_on_function_error": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"destination_config": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"on_failure": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
					},
				},
			},
			"function_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_processing_result": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_transition_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// resourceAwsLambdaEventSourceMappingCreate maps to:
// CreateEventSourceMapping in the API / SDK
func resourceAwsLambdaEventSourceMappingCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	input := &lambda.CreateEventSourceMappingInput{
		Enabled:      aws.Bool(d.Get("enabled").(bool)),
		FunctionName: aws.String(d.Get("function_name").(string)),
	}

	if v, ok := d.GetOk("batch_size"); ok {
		input.BatchSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("bisect_batch_on_function_error"); ok {
		input.BisectBatchOnFunctionError = aws.Bool(v.(bool))
	}

	if vDest, ok := d.GetOk("destination_config"); ok {
		input.DestinationConfig = expandLambdaEventSourceMappingDestinationConfig(vDest.([]interface{}))
	}

	if v, ok := d.GetOk("event_source_arn"); ok {
		input.EventSourceArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maximum_batching_window_in_seconds"); ok {
		input.MaximumBatchingWindowInSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("maximum_record_age_in_seconds"); ok {
		input.MaximumRecordAgeInSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOkExists("maximum_retry_attempts"); ok {
		input.MaximumRetryAttempts = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("parallelization_factor"); ok {
		input.ParallelizationFactor = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("starting_position"); ok {
		input.StartingPosition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("starting_position_timestamp"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))

		input.StartingPositionTimestamp = aws.Time(t)
	}

	if v, ok := d.GetOk("topics"); ok && v.(*schema.Set).Len() > 0 {
		input.Topics = expandStringSet(v.(*schema.Set))
	}

	// When non-ARN targets are supported, set target to the non-nil value.
	target := input.EventSourceArn

	log.Printf("[DEBUG] Creating Lambda Event Source Mapping: %s", input)

	// IAM profiles and roles can take some time to propagate in AWS:
	//  http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#launch-instance-with-role-console
	// Error creating Lambda function: InvalidParameterValueException: The
	// function defined for the task cannot be assumed by Lambda.
	//
	// The role may exist, but the permissions may not have propagated, so we
	// retry
	var eventSourceMappingConfiguration *lambda.EventSourceMappingConfiguration
	var err error
	err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		eventSourceMappingConfiguration, err = conn.CreateEventSourceMapping(input)

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeInvalidParameterValueException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		eventSourceMappingConfiguration, err = conn.CreateEventSourceMapping(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Lambda Event Source Mapping (%s): %w", aws.StringValue(target), err)
	}

	d.SetId(aws.StringValue(eventSourceMappingConfiguration.UUID))

	if _, err := waiter.EventSourceMappingCreate(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Lambda Event Source Mapping (%s) to create: %w", d.Id(), err)
	}

	return resourceAwsLambdaEventSourceMappingRead(d, meta)
}

// resourceAwsLambdaEventSourceMappingRead maps to:
// GetEventSourceMapping in the API / SDK
func resourceAwsLambdaEventSourceMappingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	eventSourceMappingConfiguration, err := finder.EventSourceMappingConfigurationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[DEBUG] Lambda Event Source Mapping (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Lambda Event Source Mapping (%s): %w", d.Id(), err)
	}

	d.Set("batch_size", eventSourceMappingConfiguration.BatchSize)
	d.Set("maximum_batching_window_in_seconds", eventSourceMappingConfiguration.MaximumBatchingWindowInSeconds)
	d.Set("event_source_arn", eventSourceMappingConfiguration.EventSourceArn)
	d.Set("function_arn", eventSourceMappingConfiguration.FunctionArn)
	d.Set("last_modified", aws.TimeValue(eventSourceMappingConfiguration.LastModified).Format(time.RFC3339))
	d.Set("last_processing_result", eventSourceMappingConfiguration.LastProcessingResult)
	d.Set("state", eventSourceMappingConfiguration.State)
	d.Set("state_transition_reason", eventSourceMappingConfiguration.StateTransitionReason)
	d.Set("uuid", eventSourceMappingConfiguration.UUID)
	d.Set("function_name", eventSourceMappingConfiguration.FunctionArn)
	d.Set("parallelization_factor", eventSourceMappingConfiguration.ParallelizationFactor)
	d.Set("maximum_retry_attempts", eventSourceMappingConfiguration.MaximumRetryAttempts)
	d.Set("maximum_record_age_in_seconds", eventSourceMappingConfiguration.MaximumRecordAgeInSeconds)
	d.Set("bisect_batch_on_function_error", eventSourceMappingConfiguration.BisectBatchOnFunctionError)
	if err := d.Set("destination_config", flattenLambdaEventSourceMappingDestinationConfig(eventSourceMappingConfiguration.DestinationConfig)); err != nil {
		return fmt.Errorf("error setting destination_config: %w", err)
	}
	if err := d.Set("topics", flattenStringSet(eventSourceMappingConfiguration.Topics)); err != nil {
		return fmt.Errorf("error setting topics: %w", err)
	}

	state := aws.StringValue(eventSourceMappingConfiguration.State)

	switch state {
	case waiter.EventSourceMappingStateEnabled, waiter.EventSourceMappingStateEnabling:
		d.Set("enabled", true)
	case waiter.EventSourceMappingStateDisabled, waiter.EventSourceMappingStateDisabling:
		d.Set("enabled", false)
	default:
		log.Printf("[WARN] Lambda Event Source Mapping is neither enabled nor disabled but %s", state)
	}

	return nil
}

// resourceAwsLambdaEventSourceMappingDelete maps to:
// DeleteEventSourceMapping in the API / SDK
func resourceAwsLambdaEventSourceMappingDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	log.Printf("[INFO] Deleting Lambda Event Source Mapping: %s", d.Id())

	input := &lambda.DeleteEventSourceMappingInput{
		UUID: aws.String(d.Id()),
	}

	err := resource.Retry(waiter.EventSourceMappingPropagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteEventSourceMapping(input)

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			return nil
		}

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceInUseException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteEventSourceMapping(input)
	}

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Lambda Event Source Mapping (%s): %w", d.Id(), err)
	}

	if _, err := waiter.EventSourceMappingDelete(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Lambda Event Source Mapping (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

// resourceAwsLambdaEventSourceMappingUpdate maps to:
// UpdateEventSourceMapping in the API / SDK
func resourceAwsLambdaEventSourceMappingUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	log.Printf("[DEBUG] Updating Lambda Event Source Mapping: %s", d.Id())

	input := &lambda.UpdateEventSourceMappingInput{
		UUID: aws.String(d.Id()),
	}

	if d.HasChange("batch_size") {
		input.BatchSize = aws.Int64(int64(d.Get("batch_size").(int)))
	}

	if d.HasChange("bisect_batch_on_function_error") {
		input.BisectBatchOnFunctionError = aws.Bool(d.Get("bisect_batch_on_function_error").(bool))
	}

	if d.HasChange("destination_config") {
		input.DestinationConfig = expandLambdaEventSourceMappingDestinationConfig(d.Get("destination_config").([]interface{}))
	}

	if d.HasChange("enabled") {
		input.Enabled = aws.Bool(d.Get("enabled").(bool))
	}

	if d.HasChange("function_name") {
		input.FunctionName = aws.String(d.Get("function_name").(string))
	}

	if d.HasChange("maximum_batching_window_in_seconds") {
		input.MaximumBatchingWindowInSeconds = aws.Int64(int64(d.Get("maximum_batching_window_in_seconds").(int)))
	}

	if d.HasChange("maximum_record_age_in_seconds") {
		input.MaximumRecordAgeInSeconds = aws.Int64(int64(d.Get("maximum_record_age_in_seconds").(int)))
	}

	if d.HasChange("maximum_retry_attempts") {
		input.MaximumRetryAttempts = aws.Int64(int64(d.Get("maximum_retry_attempts").(int)))
	}

	if d.HasChange("parallelization_factor") {
		input.ParallelizationFactor = aws.Int64(int64(d.Get("parallelization_factor").(int)))
	}

	err := resource.Retry(waiter.EventSourceMappingPropagationTimeout, func() *resource.RetryError {
		_, err := conn.UpdateEventSourceMapping(input)

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeInvalidParameterValueException) {
			return resource.RetryableError(err)
		}

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceInUseException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateEventSourceMapping(input)
	}

	if err != nil {
		return fmt.Errorf("error updating Lambda Event Source Mapping (%s): %w", d.Id(), err)
	}

	if _, err := waiter.EventSourceMappingUpdate(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Lambda Event Source Mapping (%s) to update: %w", d.Id(), err)
	}

	return resourceAwsLambdaEventSourceMappingRead(d, meta)
}
