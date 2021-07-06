package aws

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lambda"
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

					var serviceName string
					if v, ok := d.GetOk("event_source_arn"); ok {
						eventSourceARN, err := arn.Parse(v.(string))
						if err != nil {
							return false
						}

						serviceName = eventSourceARN.Service
					} else if _, ok := d.GetOk("self_managed_event_source"); ok {
						serviceName = "kafka"
					}

					switch serviceName {
					case "dynamodb", "kinesis", "kafka", "mq":
						return old == "100"
					case "sqs":
						return old == "10"
					}

					return old == new
				},
			},

			"bisect_batch_on_function_error": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"destination_config": {
				Type:     schema.TypeList,
				Optional: true,
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
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"event_source_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"event_source_arn", "self_managed_event_source"},
			},

			"function_arn": {
				Type:     schema.TypeString,
				Computed: true,
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

			"function_response_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(lambda.FunctionResponseType_Values(), false),
				},
			},

			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_processing_result": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"maximum_batching_window_in_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
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

			"maximum_retry_attempts": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(-1, 10_000),
			},

			"parallelization_factor": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 10),
				Computed:     true,
			},

			"queues": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 1000),
				},
			},

			"self_managed_event_source": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoints": {
							Type:     schema.TypeMap,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if k == "self_managed_event_source.0.endpoints.KAFKA_BOOTSTRAP_SERVERS" {
									// AWS returns the bootstrap brokers in sorted order.
									olds := strings.Split(old, ",")
									sort.Strings(olds)
									news := strings.Split(new, ",")
									sort.Strings(news)

									return reflect.DeepEqual(olds, news)
								}

								return old == new
							},
						},
					},
				},
				ExactlyOneOf: []string{"event_source_arn", "self_managed_event_source"},
			},

			"source_access_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 22,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lambda.SourceAccessType_Values(), false),
						},
						"uri": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
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

			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"state_transition_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"topics": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 249),
				},
			},

			"tumbling_window_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 900),
			},

			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsLambdaEventSourceMappingCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	functionName := d.Get("function_name").(string)
	input := &lambda.CreateEventSourceMappingInput{
		Enabled:      aws.Bool(d.Get("enabled").(bool)),
		FunctionName: aws.String(functionName),
	}

	var target string

	if v, ok := d.GetOk("batch_size"); ok {
		input.BatchSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("bisect_batch_on_function_error"); ok {
		input.BisectBatchOnFunctionError = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("destination_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DestinationConfig = expandLambdaDestinationConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("event_source_arn"); ok {
		v := v.(string)

		input.EventSourceArn = aws.String(v)
		target = v
	}

	if v, ok := d.GetOk("function_response_types"); ok && v.(*schema.Set).Len() > 0 {
		input.FunctionResponseTypes = expandStringSet(v.(*schema.Set))
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

	if v, ok := d.GetOk("queues"); ok && v.(*schema.Set).Len() > 0 {
		input.Queues = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("self_managed_event_source"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SelfManagedEventSource = expandLambdaSelfManagedEventSource(v.([]interface{})[0].(map[string]interface{}))

		target = "Self-Managed Apache Kafka"
	}

	if v, ok := d.GetOk("source_access_configuration"); ok && v.(*schema.Set).Len() > 0 {
		input.SourceAccessConfigurations = expandLambdaSourceAccessConfigurations(v.(*schema.Set).List())
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

	if v, ok := d.GetOk("tumbling_window_in_seconds"); ok {
		input.TumblingWindowInSeconds = aws.Int64(int64(v.(int)))
	}

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

		if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "cannot be assumed by Lambda") {
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "execution role does not have permissions") {
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
		return fmt.Errorf("error creating Lambda Event Source Mapping (%s): %w", target, err)
	}

	d.SetId(aws.StringValue(eventSourceMappingConfiguration.UUID))

	if _, err := waiter.EventSourceMappingCreate(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Lambda Event Source Mapping (%s) to create: %w", d.Id(), err)
	}

	return resourceAwsLambdaEventSourceMappingRead(d, meta)
}

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
	d.Set("bisect_batch_on_function_error", eventSourceMappingConfiguration.BisectBatchOnFunctionError)
	if eventSourceMappingConfiguration.DestinationConfig != nil {
		if err := d.Set("destination_config", []interface{}{flattenLambdaDestinationConfig(eventSourceMappingConfiguration.DestinationConfig)}); err != nil {
			return fmt.Errorf("error setting destination_config: %w", err)
		}
	} else {
		d.Set("destination_config", nil)
	}
	d.Set("event_source_arn", eventSourceMappingConfiguration.EventSourceArn)
	d.Set("function_arn", eventSourceMappingConfiguration.FunctionArn)
	d.Set("function_name", eventSourceMappingConfiguration.FunctionArn)
	d.Set("function_response_types", aws.StringValueSlice(eventSourceMappingConfiguration.FunctionResponseTypes))
	if eventSourceMappingConfiguration.LastModified != nil {
		d.Set("last_modified", aws.TimeValue(eventSourceMappingConfiguration.LastModified).Format(time.RFC3339))
	} else {
		d.Set("last_modified", nil)
	}
	d.Set("last_processing_result", eventSourceMappingConfiguration.LastProcessingResult)
	d.Set("maximum_batching_window_in_seconds", eventSourceMappingConfiguration.MaximumBatchingWindowInSeconds)
	d.Set("maximum_record_age_in_seconds", eventSourceMappingConfiguration.MaximumRecordAgeInSeconds)
	d.Set("maximum_retry_attempts", eventSourceMappingConfiguration.MaximumRetryAttempts)
	d.Set("parallelization_factor", eventSourceMappingConfiguration.ParallelizationFactor)
	d.Set("queues", aws.StringValueSlice(eventSourceMappingConfiguration.Queues))
	if eventSourceMappingConfiguration.SelfManagedEventSource != nil {
		if err := d.Set("self_managed_event_source", []interface{}{flattenLambdaSelfManagedEventSource(eventSourceMappingConfiguration.SelfManagedEventSource)}); err != nil {
			return fmt.Errorf("error setting self_managed_event_source: %w", err)
		}
	} else {
		d.Set("self_managed_event_source", nil)
	}
	if err := d.Set("source_access_configuration", flattenLambdaSourceAccessConfigurations(eventSourceMappingConfiguration.SourceAccessConfigurations)); err != nil {
		return fmt.Errorf("error setting source_access_configuration: %w", err)
	}
	d.Set("starting_position", eventSourceMappingConfiguration.StartingPosition)
	if eventSourceMappingConfiguration.StartingPositionTimestamp != nil {
		d.Set("starting_position_timestamp", aws.TimeValue(eventSourceMappingConfiguration.StartingPositionTimestamp).Format(time.RFC3339))
	} else {
		d.Set("starting_position_timestamp", nil)
	}
	d.Set("state", eventSourceMappingConfiguration.State)
	d.Set("state_transition_reason", eventSourceMappingConfiguration.StateTransitionReason)
	d.Set("topics", aws.StringValueSlice(eventSourceMappingConfiguration.Topics))
	d.Set("tumbling_window_in_seconds", eventSourceMappingConfiguration.TumblingWindowInSeconds)
	d.Set("uuid", eventSourceMappingConfiguration.UUID)

	switch state := d.Get("state").(string); state {
	case waiter.EventSourceMappingStateEnabled, waiter.EventSourceMappingStateEnabling:
		d.Set("enabled", true)
	case waiter.EventSourceMappingStateDisabled, waiter.EventSourceMappingStateDisabling:
		d.Set("enabled", false)
	default:
		log.Printf("[WARN] Lambda Event Source Mapping (%s) is neither enabled nor disabled, but %s", d.Id(), state)
		d.Set("enabled", nil)
	}

	return nil
}

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
		if v, ok := d.GetOk("destination_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.DestinationConfig = expandLambdaDestinationConfig(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	if d.HasChange("enabled") {
		input.Enabled = aws.Bool(d.Get("enabled").(bool))
	}

	if d.HasChange("function_name") {
		input.FunctionName = aws.String(d.Get("function_name").(string))
	}

	if d.HasChange("function_response_types") {
		input.FunctionResponseTypes = expandStringSet(d.Get("function_response_types").(*schema.Set))
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

	if d.HasChange("source_access_configuration") {
		if v, ok := d.GetOk("source_access_configuration"); ok && v.(*schema.Set).Len() > 0 {
			input.SourceAccessConfigurations = expandLambdaSourceAccessConfigurations(v.(*schema.Set).List())
		}
	}

	if d.HasChange("tumbling_window_in_seconds") {
		input.TumblingWindowInSeconds = aws.Int64(int64(d.Get("tumbling_window_in_seconds").(int)))
	}

	err := resource.Retry(waiter.EventSourceMappingPropagationTimeout, func() *resource.RetryError {
		_, err := conn.UpdateEventSourceMapping(input)

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

func resourceAwsLambdaEventSourceMappingDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	log.Printf("[INFO] Deleting Lambda Event Source Mapping: %s", d.Id())

	input := &lambda.DeleteEventSourceMappingInput{
		UUID: aws.String(d.Id()),
	}

	err := resource.Retry(waiter.EventSourceMappingPropagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteEventSourceMapping(input)

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

func expandLambdaDestinationConfig(tfMap map[string]interface{}) *lambda.DestinationConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.DestinationConfig{}

	if v, ok := tfMap["on_failure"].([]interface{}); ok && len(v) > 0 {
		apiObject.OnFailure = expandLambdaOnFailure(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandLambdaOnFailure(tfMap map[string]interface{}) *lambda.OnFailure {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.OnFailure{}

	if v, ok := tfMap["destination_arn"].(string); ok {
		apiObject.Destination = aws.String(v)
	}

	return apiObject
}

func flattenLambdaDestinationConfig(apiObject *lambda.DestinationConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.OnFailure; v != nil {
		tfMap["on_failure"] = []interface{}{flattenLambdaOnFailure(v)}
	}

	return tfMap
}

func flattenLambdaOnFailure(apiObject *lambda.OnFailure) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Destination; v != nil {
		tfMap["destination_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func expandLambdaSelfManagedEventSource(tfMap map[string]interface{}) *lambda.SelfManagedEventSource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.SelfManagedEventSource{}

	if v, ok := tfMap["endpoints"].(map[string]interface{}); ok && len(v) > 0 {
		m := map[string][]*string{}

		for k, v := range v {
			m[k] = aws.StringSlice(strings.Split(v.(string), ","))
		}

		apiObject.Endpoints = m
	}

	return apiObject
}

func flattenLambdaSelfManagedEventSource(apiObject *lambda.SelfManagedEventSource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Endpoints; v != nil {
		m := map[string]string{}

		for k, v := range v {
			m[k] = strings.Join(aws.StringValueSlice(v), ",")
		}

		tfMap["endpoints"] = m
	}

	return tfMap
}

func expandLambdaSourceAccessConfiguration(tfMap map[string]interface{}) *lambda.SourceAccessConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.SourceAccessConfiguration{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	if v, ok := tfMap["uri"].(string); ok && v != "" {
		apiObject.URI = aws.String(v)
	}

	return apiObject
}

func expandLambdaSourceAccessConfigurations(tfList []interface{}) []*lambda.SourceAccessConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*lambda.SourceAccessConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLambdaSourceAccessConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenLambdaSourceAccessConfiguration(apiObject *lambda.SourceAccessConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	if v := apiObject.URI; v != nil {
		tfMap["uri"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLambdaSourceAccessConfigurations(apiObjects []*lambda.SourceAccessConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLambdaSourceAccessConfiguration(apiObject))
	}

	return tfList
}
