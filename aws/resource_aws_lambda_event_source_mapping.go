package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					lambda.EventSourcePositionAtTimestamp,
					lambda.EventSourcePositionLatest,
					lambda.EventSourcePositionTrimHorizon,
				}, false),
			},
			"starting_position_timestamp": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
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
					case dynamodb.ServiceName, kinesis.ServiceName:
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
				ValidateFunc: validation.IntBetween(0, 10000),
				Computed:     true,
			},
			"maximum_record_age_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(60, 604800),
				Computed:     true,
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

	functionName := d.Get("function_name").(string)
	eventSourceArn := d.Get("event_source_arn").(string)

	log.Printf("[DEBUG] Creating Lambda event source mapping: source %s to function %s", eventSourceArn, functionName)

	params := &lambda.CreateEventSourceMappingInput{
		EventSourceArn: aws.String(eventSourceArn),
		FunctionName:   aws.String(functionName),
		Enabled:        aws.Bool(d.Get("enabled").(bool)),
	}

	if batchSize, ok := d.GetOk("batch_size"); ok {
		params.BatchSize = aws.Int64(int64(batchSize.(int)))
	}

	if batchWindow, ok := d.GetOk("maximum_batching_window_in_seconds"); ok {
		params.MaximumBatchingWindowInSeconds = aws.Int64(int64(batchWindow.(int)))
	}

	if startingPosition, ok := d.GetOk("starting_position"); ok {
		params.StartingPosition = aws.String(startingPosition.(string))
	}

	if startingPositionTimestamp, ok := d.GetOk("starting_position_timestamp"); ok {
		t, _ := time.Parse(time.RFC3339, startingPositionTimestamp.(string))
		params.StartingPositionTimestamp = aws.Time(t)
	}
	if parallelizationFactor, ok := d.GetOk("parallelization_factor"); ok {
		params.ParallelizationFactor = aws.Int64(int64(parallelizationFactor.(int)))
	}

	if maximumRetryAttempts, ok := d.GetOk("maximum_retry_attempts"); ok {
		params.MaximumRetryAttempts = aws.Int64(int64(maximumRetryAttempts.(int)))
	}

	if maximumRecordAgeInSeconds, ok := d.GetOk("maximum_record_age_in_seconds"); ok {
		params.MaximumRecordAgeInSeconds = aws.Int64(int64(maximumRecordAgeInSeconds.(int)))
	}

	if bisectBatchOnFunctionError, ok := d.GetOk("bisect_batch_on_function_error"); ok {
		params.BisectBatchOnFunctionError = aws.Bool(bisectBatchOnFunctionError.(bool))
	}

	if vDest, ok := d.GetOk("destination_config"); ok {
		params.SetDestinationConfig(expandLambdaEventSourceMappingDestinationConfig(vDest.([]interface{})))
	}

	// IAM profiles and roles can take some time to propagate in AWS:
	//  http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#launch-instance-with-role-console
	// Error creating Lambda function: InvalidParameterValueException: The
	// function defined for the task cannot be assumed by Lambda.
	//
	// The role may exist, but the permissions may not have propagated, so we
	// retry
	var eventSourceMappingConfiguration *lambda.EventSourceMappingConfiguration
	var err error
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		eventSourceMappingConfiguration, err = conn.CreateEventSourceMapping(params)
		if err != nil {
			if awserr, ok := err.(awserr.Error); ok {
				if awserr.Code() == "InvalidParameterValueException" {
					return resource.RetryableError(awserr)
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		eventSourceMappingConfiguration, err = conn.CreateEventSourceMapping(params)
	}
	if err != nil {
		return fmt.Errorf("Error creating Lambda event source mapping: %s", err)
	}

	// No error
	d.Set("uuid", eventSourceMappingConfiguration.UUID)
	d.SetId(*eventSourceMappingConfiguration.UUID)
	return resourceAwsLambdaEventSourceMappingRead(d, meta)
}

// resourceAwsLambdaEventSourceMappingRead maps to:
// GetEventSourceMapping in the API / SDK
func resourceAwsLambdaEventSourceMappingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	log.Printf("[DEBUG] Fetching Lambda event source mapping: %s", d.Id())

	params := &lambda.GetEventSourceMappingInput{
		UUID: aws.String(d.Id()),
	}

	eventSourceMappingConfiguration, err := conn.GetEventSourceMapping(params)
	if err != nil {
		if isAWSErr(err, "ResourceNotFoundException", "") {
			log.Printf("[DEBUG] Lambda event source mapping (%s) not found", d.Id())
			d.SetId("")

			return nil
		}
		return err
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
		return fmt.Errorf("error setting destination_config: %s", err)
	}

	state := aws.StringValue(eventSourceMappingConfiguration.State)

	switch state {
	case "Enabled", "Enabling":
		d.Set("enabled", true)
	case "Disabled", "Disabling":
		d.Set("enabled", false)
	default:
		log.Printf("[DEBUG] Lambda event source mapping is neither enabled nor disabled but %s", *eventSourceMappingConfiguration.State)
	}

	return nil
}

// resourceAwsLambdaEventSourceMappingDelete maps to:
// DeleteEventSourceMapping in the API / SDK
func resourceAwsLambdaEventSourceMappingDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	log.Printf("[INFO] Deleting Lambda event source mapping: %s", d.Id())

	params := &lambda.DeleteEventSourceMappingInput{
		UUID: aws.String(d.Id()),
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteEventSourceMapping(params)
		if err != nil {
			if isAWSErr(err, lambda.ErrCodeResourceInUseException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteEventSourceMapping(params)
	}
	if err != nil {
		return fmt.Errorf("Error deleting Lambda event source mapping: %s", err)
	}

	return nil
}

// resourceAwsLambdaEventSourceMappingUpdate maps to:
// UpdateEventSourceMapping in the API / SDK
func resourceAwsLambdaEventSourceMappingUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	log.Printf("[DEBUG] Updating Lambda event source mapping: %s", d.Id())

	params := &lambda.UpdateEventSourceMappingInput{
		UUID:         aws.String(d.Id()),
		BatchSize:    aws.Int64(int64(d.Get("batch_size").(int))),
		FunctionName: aws.String(d.Get("function_name").(string)),
		Enabled:      aws.Bool(d.Get("enabled").(bool)),
	}

	// AWS API will fail if this parameter is set (even as default value) for sqs event source.  Ideally this should be implemented in GO SDK or AWS API itself.
	eventSourceArn, err := arn.Parse(d.Get("event_source_arn").(string))
	if err != nil {
		return fmt.Errorf("Error updating event source mapping: %s", err)
	}

	if eventSourceArn.Service != "sqs" {
		params.MaximumBatchingWindowInSeconds = aws.Int64(int64(d.Get("maximum_batching_window_in_seconds").(int)))

		if parallelizationFactor, ok := d.GetOk("parallelization_factor"); ok {
			params.SetParallelizationFactor(int64(parallelizationFactor.(int)))
		}

		if maximumRetryAttempts, ok := d.GetOk("maximum_retry_attempts"); ok {
			params.SetMaximumRetryAttempts(int64(maximumRetryAttempts.(int)))
		}

		params.MaximumRecordAgeInSeconds = aws.Int64(int64(d.Get("maximum_record_age_in_seconds").(int)))

		if bisectBatchOnFunctionError, ok := d.GetOk("bisect_batch_on_function_error"); ok {
			params.SetBisectBatchOnFunctionError(bisectBatchOnFunctionError.(bool))
		}

		if vDest, ok := d.GetOk("destination_config"); ok {
			params.SetDestinationConfig(expandLambdaEventSourceMappingDestinationConfig(vDest.([]interface{})))
		}
	}

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.UpdateEventSourceMapping(params)
		if err != nil {
			if isAWSErr(err, lambda.ErrCodeInvalidParameterValueException, "") ||
				isAWSErr(err, lambda.ErrCodeResourceInUseException, "") {

				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.UpdateEventSourceMapping(params)
	}
	if err != nil {
		return fmt.Errorf("Error updating Lambda event source mapping: %s", err)
	}

	return resourceAwsLambdaEventSourceMappingRead(d, meta)
}
