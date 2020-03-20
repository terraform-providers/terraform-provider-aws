package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsLambdaFunctionEventInvokeConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLambdaFunctionEventInvokeConfigCreate,
		Read:   resourceAwsLambdaFunctionEventInvokeConfigRead,
		Update: resourceAwsLambdaFunctionEventInvokeConfigUpdate,
		Delete: resourceAwsLambdaFunctionEventInvokeConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
									"destination": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
						"on_success": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination": {
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
			"function_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"maximum_event_age_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(60, 21600),
			},
			"maximum_retry_attempts": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				ValidateFunc: validation.IntBetween(0, 2),
			},
			"qualifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceAwsLambdaFunctionEventInvokeConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn
	functionName := d.Get("function_name").(string)
	qualifier := d.Get("qualifier").(string)

	id := functionName

	if qualifier != "" {
		id = fmt.Sprintf("%s:%s", functionName, qualifier)
	}

	input := &lambda.PutFunctionEventInvokeConfigInput{
		DestinationConfig:    expandLambdaFunctionEventInvokeConfigDestinationConfig(d.Get("destination_config").([]interface{})),
		FunctionName:         aws.String(functionName),
		MaximumRetryAttempts: aws.Int64(int64(d.Get("maximum_retry_attempts").(int))),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	if v, ok := d.GetOk("maximum_event_age_in_seconds"); ok {
		input.MaximumEventAgeInSeconds = aws.Int64(int64(v.(int)))
	}

	// Retry for destination validation eventual consistency errors
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.PutFunctionEventInvokeConfig(input)

		// InvalidParameterValueException: The destination ARN arn:PARTITION:SERVICE:REGION:ACCOUNT:RESOURCE is invalid.
		if isAWSErr(err, lambda.ErrCodeInvalidParameterValueException, "destination ARN") {
			return resource.RetryableError(err)
		}

		// InvalidParameterValueException: The function's execution role does not have permissions to call Publish on arn:...
		if isAWSErr(err, lambda.ErrCodeInvalidParameterValueException, "does not have permissions") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.PutFunctionEventInvokeConfig(input)
	}

	if err != nil {
		return fmt.Errorf("error putting Lambda Function Event Invoke Config (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceAwsLambdaFunctionEventInvokeConfigRead(d, meta)
}

func resourceAwsLambdaFunctionEventInvokeConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	functionName, qualifier, err := resourceAwsLambdaFunctionEventInvokeConfigParseId(d.Id())

	if err != nil {
		return err
	}

	input := &lambda.GetFunctionEventInvokeConfigInput{
		FunctionName: aws.String(functionName),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	output, err := conn.GetFunctionEventInvokeConfig(input)

	if isAWSErr(err, lambda.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Lambda Function Event Invoke Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Lambda Function Event Invoke Config (%s): %s", d.Id(), err)
	}

	if err := d.Set("destination_config", flattenLambdaFunctionEventInvokeConfigDestinationConfig(output.DestinationConfig)); err != nil {
		return fmt.Errorf("error setting destination_config: %s", err)
	}

	d.Set("function_name", functionName)
	d.Set("maximum_event_age_in_seconds", aws.Int64Value(output.MaximumEventAgeInSeconds))
	d.Set("maximum_retry_attempts", aws.Int64Value(output.MaximumRetryAttempts))
	d.Set("qualifier", qualifier)

	return nil
}

func resourceAwsLambdaFunctionEventInvokeConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	functionName, qualifier, err := resourceAwsLambdaFunctionEventInvokeConfigParseId(d.Id())

	if err != nil {
		return err
	}

	input := &lambda.PutFunctionEventInvokeConfigInput{
		DestinationConfig:    expandLambdaFunctionEventInvokeConfigDestinationConfig(d.Get("destination_config").([]interface{})),
		FunctionName:         aws.String(functionName),
		MaximumRetryAttempts: aws.Int64(int64(d.Get("maximum_retry_attempts").(int))),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	if v, ok := d.GetOk("maximum_event_age_in_seconds"); ok {
		input.MaximumEventAgeInSeconds = aws.Int64(int64(v.(int)))
	}

	// Retry for destination validation eventual consistency errors
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.PutFunctionEventInvokeConfig(input)

		// InvalidParameterValueException: The destination ARN arn:PARTITION:SERVICE:REGION:ACCOUNT:RESOURCE is invalid.
		if isAWSErr(err, lambda.ErrCodeInvalidParameterValueException, "destination ARN") {
			return resource.RetryableError(err)
		}

		// InvalidParameterValueException: The function's execution role does not have permissions to call Publish on arn:...
		if isAWSErr(err, lambda.ErrCodeInvalidParameterValueException, "does not have permissions") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.PutFunctionEventInvokeConfig(input)
	}

	if err != nil {
		return fmt.Errorf("error putting Lambda Function Event Invoke Config (%s): %s", d.Id(), err)
	}

	return resourceAwsLambdaFunctionEventInvokeConfigRead(d, meta)
}

func resourceAwsLambdaFunctionEventInvokeConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	functionName, qualifier, err := resourceAwsLambdaFunctionEventInvokeConfigParseId(d.Id())

	if err != nil {
		return err
	}

	input := &lambda.DeleteFunctionEventInvokeConfigInput{
		FunctionName: aws.String(functionName),
	}

	if qualifier != "" {
		input.Qualifier = aws.String(qualifier)
	}

	_, err = conn.DeleteFunctionEventInvokeConfig(input)

	if isAWSErr(err, lambda.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error putting Lambda Function Event Invoke Config (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsLambdaFunctionEventInvokeConfigParseId(id string) (string, string, error) {
	if arn.IsARN(id) {
		parsedARN, err := arn.Parse(id)

		if err != nil {
			return "", "", fmt.Errorf("error parsing ARN (%s): %s", id, err)
		}

		function := strings.TrimPrefix(parsedARN.Resource, "function:")

		if !strings.Contains(function, ":") {
			// Return ARN for function name to match configuration
			return id, "", nil
		}

		functionParts := strings.Split(id, ":")

		if len(functionParts) != 2 || functionParts[0] == "" || functionParts[1] == "" {
			return "", "", fmt.Errorf("unexpected format of function resource (%s), expected name:qualifier", id)
		}

		qualifier := functionParts[1]
		// Return ARN minus qualifier for function name to match configuration
		functionName := strings.TrimSuffix(id, fmt.Sprintf(":%s", qualifier))

		return functionName, qualifier, nil
	}

	if !strings.Contains(id, ":") {
		return id, "", nil
	}

	idParts := strings.Split(id, ":")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected name or name:qualifier", id)
	}

	return idParts[0], idParts[1], nil
}

func expandLambdaFunctionEventInvokeConfigDestinationConfig(l []interface{}) *lambda.DestinationConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	destinationConfig := &lambda.DestinationConfig{}

	if v, ok := m["on_failure"].([]interface{}); ok {
		destinationConfig.OnFailure = expandLambdaFunctionEventInvokeConfigDestinationConfigOnFailure(v)
	}

	if v, ok := m["on_success"].([]interface{}); ok {
		destinationConfig.OnSuccess = expandLambdaFunctionEventInvokeConfigDestinationConfigOnSuccess(v)
	}

	return destinationConfig
}

func expandLambdaFunctionEventInvokeConfigDestinationConfigOnFailure(l []interface{}) *lambda.OnFailure {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	onFailure := &lambda.OnFailure{}

	if v, ok := m["destination"].(string); ok {
		onFailure.Destination = aws.String(v)
	}

	return onFailure
}

func expandLambdaFunctionEventInvokeConfigDestinationConfigOnSuccess(l []interface{}) *lambda.OnSuccess {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	onSuccess := &lambda.OnSuccess{}

	if v, ok := m["destination"].(string); ok {
		onSuccess.Destination = aws.String(v)
	}

	return onSuccess
}

func flattenLambdaFunctionEventInvokeConfigDestinationConfig(destinationConfig *lambda.DestinationConfig) []interface{} {
	// The API will respond with empty OnFailure and OnSuccess destinations when unconfigured:
	// "DestinationConfig":{"OnFailure":{"Destination":null},"OnSuccess":{"Destination":null}}
	// Return no destination configuration to prevent Terraform state difference

	if destinationConfig == nil {
		return []interface{}{}
	}

	if destinationConfig.OnFailure == nil && destinationConfig.OnSuccess == nil {
		return []interface{}{}
	}

	if (destinationConfig.OnFailure != nil && destinationConfig.OnFailure.Destination == nil) && (destinationConfig.OnSuccess != nil && destinationConfig.OnSuccess.Destination == nil) {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"on_failure": flattenLambdaFunctionEventInvokeConfigDestinationConfigOnFailure(destinationConfig.OnFailure),
		"on_success": flattenLambdaFunctionEventInvokeConfigDestinationConfigOnSuccess(destinationConfig.OnSuccess),
	}

	return []interface{}{m}
}

func flattenLambdaFunctionEventInvokeConfigDestinationConfigOnFailure(onFailure *lambda.OnFailure) []interface{} {
	// The API will respond with empty OnFailure destination when unconfigured:
	// "DestinationConfig":{"OnFailure":{"Destination":null},"OnSuccess":{"Destination":null}}
	// Return no on failure configuration to prevent Terraform state difference

	if onFailure == nil || onFailure.Destination == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"destination": aws.StringValue(onFailure.Destination),
	}

	return []interface{}{m}
}

func flattenLambdaFunctionEventInvokeConfigDestinationConfigOnSuccess(onSuccess *lambda.OnSuccess) []interface{} {
	// The API will respond with empty OnSuccess destination when unconfigured:
	// "DestinationConfig":{"OnFailure":{"Destination":null},"OnSuccess":{"Destination":null}}
	// Return no on success configuration to prevent Terraform state difference

	if onSuccess == nil || onSuccess.Destination == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"destination": aws.StringValue(onSuccess.Destination),
	}

	return []interface{}{m}
}
