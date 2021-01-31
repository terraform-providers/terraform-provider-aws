package aws

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSsmMaintenanceWindowTask() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmMaintenanceWindowTaskCreate,
		Read:   resourceAwsSsmMaintenanceWindowTaskRead,
		Update: resourceAwsSsmMaintenanceWindowTaskUpdate,
		Delete: resourceAwsSsmMaintenanceWindowTaskDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsSsmMaintenanceWindowTaskImport,
		},

		Schema: map[string]*schema.Schema{
			"window_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"max_concurrency": {
				Type:     schema.TypeString,
				Required: true,
			},

			"max_errors": {
				Type:     schema.TypeString,
				Required: true,
			},

			"task_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"task_arn": {
				Type:     schema.TypeString,
				Required: true,
			},

			"service_role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},

			"targets": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateAwsSSMMaintenanceWindowTaskName,
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},

			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"task_invocation_parameters": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automation_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"document_version": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile("([$]LATEST|[$]DEFAULT|^[1-9][0-9]*$)"), "see https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_MaintenanceWindowAutomationParameters.html"),
									},
									"parameter": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},

												"values": {
													Type:     schema.TypeList,
													Required: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},

						"lambda_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_context": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 8000),
									},

									"payload": {
										Type:         schema.TypeString,
										Optional:     true,
										Sensitive:    true,
										ValidateFunc: validation.StringLenBetween(0, 4096),
									},

									"qualifier": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
								},
							},
						},

						"run_command_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"comment": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"document_hash": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"document_hash_type": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											ssm.DocumentHashTypeSha256,
											ssm.DocumentHashTypeSha1,
										}, false),
									},

									"notification_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"notification_arn": {
													Type:     schema.TypeString,
													Optional: true,
												},

												"notification_events": {
													Type:     schema.TypeList,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},

												"notification_type": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														ssm.NotificationTypeCommand,
														ssm.NotificationTypeInvocation,
													}, false),
												},
											},
										},
									},

									"output_s3_bucket": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"output_s3_key_prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"parameter": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},

												"values": {
													Type:     schema.TypeList,
													Required: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},

									"service_role_arn": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"timeout_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},

						"step_functions_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input": {
										Type:         schema.TypeString,
										Optional:     true,
										Sensitive:    true,
										ValidateFunc: validation.StringLenBetween(0, 4096),
									},

									"name": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 80),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func expandAwsSsmTaskInvocationParameters(config []interface{}) *ssm.MaintenanceWindowTaskInvocationParameters {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.MaintenanceWindowTaskInvocationParameters{}
	for _, v := range config {
		paramConfig := v.(map[string]interface{})
		if attr, ok := paramConfig["automation_parameters"]; ok && len(attr.([]interface{})) > 0 && attr.([]interface{})[0] != nil {
			params.Automation = expandAwsSsmTaskInvocationAutomationParameters(attr.([]interface{}))
		}
		if attr, ok := paramConfig["lambda_parameters"]; ok && len(attr.([]interface{})) > 0 && attr.([]interface{})[0] != nil {
			params.Lambda = expandAwsSsmTaskInvocationLambdaParameters(attr.([]interface{}))
		}
		if attr, ok := paramConfig["run_command_parameters"]; ok && len(attr.([]interface{})) > 0 && attr.([]interface{})[0] != nil {
			params.RunCommand = expandAwsSsmTaskInvocationRunCommandParameters(attr.([]interface{}))
		}
		if attr, ok := paramConfig["step_functions_parameters"]; ok && len(attr.([]interface{})) > 0 && attr.([]interface{})[0] != nil {
			params.StepFunctions = expandAwsSsmTaskInvocationStepFunctionsParameters(attr.([]interface{}))
		}
	}
	return params
}

func flattenAwsSsmTaskInvocationParameters(parameters *ssm.MaintenanceWindowTaskInvocationParameters) []interface{} {
	result := make(map[string]interface{})
	if parameters.Automation != nil {
		result["automation_parameters"] = flattenAwsSsmTaskInvocationAutomationParameters(parameters.Automation)
	}

	if parameters.Lambda != nil {
		result["lambda_parameters"] = flattenAwsSsmTaskInvocationLambdaParameters(parameters.Lambda)
	}

	if parameters.RunCommand != nil {
		result["run_command_parameters"] = flattenAwsSsmTaskInvocationRunCommandParameters(parameters.RunCommand)
	}

	if parameters.StepFunctions != nil {
		result["step_functions_parameters"] = flattenAwsSsmTaskInvocationStepFunctionsParameters(parameters.StepFunctions)
	}

	return []interface{}{result}
}

func expandAwsSsmTaskInvocationAutomationParameters(config []interface{}) *ssm.MaintenanceWindowAutomationParameters {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.MaintenanceWindowAutomationParameters{}
	configParam := config[0].(map[string]interface{})
	if attr, ok := configParam["document_version"]; ok && len(attr.(string)) != 0 {
		params.DocumentVersion = aws.String(attr.(string))
	}
	if attr, ok := configParam["parameter"]; ok && len(attr.(*schema.Set).List()) > 0 {
		params.Parameters = expandAwsSsmTaskInvocationCommonParameters(attr.(*schema.Set).List())
	}

	return params
}

func flattenAwsSsmTaskInvocationAutomationParameters(parameters *ssm.MaintenanceWindowAutomationParameters) []interface{} {
	result := make(map[string]interface{})

	if parameters.DocumentVersion != nil {
		result["document_version"] = aws.StringValue(parameters.DocumentVersion)
	}
	if parameters.Parameters != nil {
		result["parameter"] = flattenAwsSsmTaskInvocationCommonParameters(parameters.Parameters)
	}

	return []interface{}{result}
}

func expandAwsSsmTaskInvocationLambdaParameters(config []interface{}) *ssm.MaintenanceWindowLambdaParameters {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.MaintenanceWindowLambdaParameters{}
	configParam := config[0].(map[string]interface{})
	if attr, ok := configParam["client_context"]; ok && len(attr.(string)) != 0 {
		params.ClientContext = aws.String(attr.(string))
	}
	if attr, ok := configParam["payload"]; ok && len(attr.(string)) != 0 {
		params.Payload = []byte(attr.(string))
	}
	if attr, ok := configParam["qualifier"]; ok && len(attr.(string)) != 0 {
		params.Qualifier = aws.String(attr.(string))
	}
	return params
}

func flattenAwsSsmTaskInvocationLambdaParameters(parameters *ssm.MaintenanceWindowLambdaParameters) []interface{} {
	result := make(map[string]interface{})

	if parameters.ClientContext != nil {
		result["client_context"] = aws.StringValue(parameters.ClientContext)
	}
	if parameters.Payload != nil {
		result["payload"] = string(parameters.Payload)
	}
	if parameters.Qualifier != nil {
		result["qualifier"] = aws.StringValue(parameters.Qualifier)
	}
	return []interface{}{result}
}

func expandAwsSsmTaskInvocationRunCommandParameters(config []interface{}) *ssm.MaintenanceWindowRunCommandParameters {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.MaintenanceWindowRunCommandParameters{}
	configParam := config[0].(map[string]interface{})
	if attr, ok := configParam["comment"]; ok && len(attr.(string)) != 0 {
		params.Comment = aws.String(attr.(string))
	}
	if attr, ok := configParam["document_hash"]; ok && len(attr.(string)) != 0 {
		params.DocumentHash = aws.String(attr.(string))
	}
	if attr, ok := configParam["document_hash_type"]; ok && len(attr.(string)) != 0 {
		params.DocumentHashType = aws.String(attr.(string))
	}
	if attr, ok := configParam["notification_config"]; ok && len(attr.([]interface{})) > 0 {
		params.NotificationConfig = expandAwsSsmTaskInvocationRunCommandParametersNotificationConfig(attr.([]interface{}))
	}
	if attr, ok := configParam["output_s3_bucket"]; ok && len(attr.(string)) != 0 {
		params.OutputS3BucketName = aws.String(attr.(string))
	}
	if attr, ok := configParam["output_s3_key_prefix"]; ok && len(attr.(string)) != 0 {
		params.OutputS3KeyPrefix = aws.String(attr.(string))
	}
	if attr, ok := configParam["parameter"]; ok && len(attr.(*schema.Set).List()) > 0 {
		params.Parameters = expandAwsSsmTaskInvocationCommonParameters(attr.(*schema.Set).List())
	}
	if attr, ok := configParam["service_role_arn"]; ok && len(attr.(string)) != 0 {
		params.ServiceRoleArn = aws.String(attr.(string))
	}
	if attr, ok := configParam["timeout_seconds"]; ok && attr.(int) != 0 {
		params.TimeoutSeconds = aws.Int64(int64(attr.(int)))
	}
	return params
}

func flattenAwsSsmTaskInvocationRunCommandParameters(parameters *ssm.MaintenanceWindowRunCommandParameters) []interface{} {
	result := make(map[string]interface{})

	if parameters.Comment != nil {
		result["comment"] = aws.StringValue(parameters.Comment)
	}
	if parameters.DocumentHash != nil {
		result["document_hash"] = aws.StringValue(parameters.DocumentHash)
	}
	if parameters.DocumentHashType != nil {
		result["document_hash_type"] = aws.StringValue(parameters.DocumentHashType)
	}
	if parameters.NotificationConfig != nil {
		result["notification_config"] = flattenAwsSsmTaskInvocationRunCommandParametersNotificationConfig(parameters.NotificationConfig)
	}
	if parameters.OutputS3BucketName != nil {
		result["output_s3_bucket"] = aws.StringValue(parameters.OutputS3BucketName)
	}
	if parameters.OutputS3KeyPrefix != nil {
		result["output_s3_key_prefix"] = aws.StringValue(parameters.OutputS3KeyPrefix)
	}
	if parameters.Parameters != nil {
		result["parameter"] = flattenAwsSsmTaskInvocationCommonParameters(parameters.Parameters)
	}
	if parameters.ServiceRoleArn != nil {
		result["service_role_arn"] = aws.StringValue(parameters.ServiceRoleArn)
	}
	if parameters.TimeoutSeconds != nil {
		result["timeout_seconds"] = aws.Int64Value(parameters.TimeoutSeconds)
	}

	return []interface{}{result}
}

func expandAwsSsmTaskInvocationStepFunctionsParameters(config []interface{}) *ssm.MaintenanceWindowStepFunctionsParameters {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.MaintenanceWindowStepFunctionsParameters{}
	configParam := config[0].(map[string]interface{})

	if attr, ok := configParam["input"]; ok && len(attr.(string)) != 0 {
		params.Input = aws.String(attr.(string))
	}
	if attr, ok := configParam["name"]; ok && len(attr.(string)) != 0 {
		params.Name = aws.String(attr.(string))
	}

	return params
}

func flattenAwsSsmTaskInvocationStepFunctionsParameters(parameters *ssm.MaintenanceWindowStepFunctionsParameters) []interface{} {
	result := make(map[string]interface{})

	if parameters.Input != nil {
		result["input"] = aws.StringValue(parameters.Input)
	}
	if parameters.Name != nil {
		result["name"] = aws.StringValue(parameters.Name)
	}
	return []interface{}{result}
}

func expandAwsSsmTaskInvocationRunCommandParametersNotificationConfig(config []interface{}) *ssm.NotificationConfig {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := &ssm.NotificationConfig{}
	configParam := config[0].(map[string]interface{})

	if attr, ok := configParam["notification_arn"]; ok && len(attr.(string)) != 0 {
		params.NotificationArn = aws.String(attr.(string))
	}
	if attr, ok := configParam["notification_events"]; ok && len(attr.([]interface{})) > 0 {
		params.NotificationEvents = expandStringList(attr.([]interface{}))
	}
	if attr, ok := configParam["notification_type"]; ok && len(attr.(string)) != 0 {
		params.NotificationType = aws.String(attr.(string))
	}

	return params
}

func flattenAwsSsmTaskInvocationRunCommandParametersNotificationConfig(config *ssm.NotificationConfig) []interface{} {
	result := make(map[string]interface{})

	if config.NotificationArn != nil {
		result["notification_arn"] = aws.StringValue(config.NotificationArn)
	}
	if config.NotificationEvents != nil {
		result["notification_events"] = flattenStringList(config.NotificationEvents)
	}
	if config.NotificationType != nil {
		result["notification_type"] = aws.StringValue(config.NotificationType)
	}

	return []interface{}{result}
}

func expandAwsSsmTaskInvocationCommonParameters(config []interface{}) map[string][]*string {
	if len(config) == 0 || config[0] == nil {
		return nil
	}

	params := make(map[string][]*string)

	for _, v := range config {
		paramConfig := v.(map[string]interface{})
		params[paramConfig["name"].(string)] = expandStringList(paramConfig["values"].([]interface{}))
	}

	return params
}

func flattenAwsSsmTaskInvocationCommonParameters(parameters map[string][]*string) []interface{} {
	attributes := make([]interface{}, 0, len(parameters))

	keys := make([]string, 0, len(parameters))
	for k := range parameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		values := make([]string, 0)
		for _, value := range parameters[key] {
			values = append(values, aws.StringValue(value))
		}
		params := map[string]interface{}{
			"name":   key,
			"values": values,
		}
		attributes = append(attributes, params)
	}

	return attributes
}

func resourceAwsSsmMaintenanceWindowTaskCreate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Registering SSM Maintenance Window Task")

	params := &ssm.RegisterTaskWithMaintenanceWindowInput{
		WindowId:       aws.String(d.Get("window_id").(string)),
		MaxConcurrency: aws.String(d.Get("max_concurrency").(string)),
		MaxErrors:      aws.String(d.Get("max_errors").(string)),
		TaskType:       aws.String(d.Get("task_type").(string)),
		ServiceRoleArn: aws.String(d.Get("service_role_arn").(string)),
		TaskArn:        aws.String(d.Get("task_arn").(string)),
		Targets:        expandAwsSsmTargets(d.Get("targets").([]interface{})),
	}

	if v, ok := d.GetOk("name"); ok {
		params.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("priority"); ok {
		params.Priority = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("task_invocation_parameters"); ok {
		params.TaskInvocationParameters = expandAwsSsmTaskInvocationParameters(v.([]interface{}))
	}

	resp, err := ssmconn.RegisterTaskWithMaintenanceWindow(params)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.WindowTaskId))

	return resourceAwsSsmMaintenanceWindowTaskRead(d, meta)
}

func resourceAwsSsmMaintenanceWindowTaskRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn
	windowID := d.Get("window_id").(string)

	params := &ssm.GetMaintenanceWindowTaskInput{
		WindowId:     aws.String(windowID),
		WindowTaskId: aws.String(d.Id()),
	}
	resp, err := ssmconn.GetMaintenanceWindowTask(params)
	if isAWSErr(err, ssm.ErrCodeDoesNotExistException, "") {
		log.Printf("[WARN] Maintenance Window (%s) Task (%s) not found, removing from state", windowID, d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting Maintenance Window (%s) Task (%s): %s", windowID, d.Id(), err)
	}

	d.Set("window_id", resp.WindowId)
	d.Set("max_concurrency", resp.MaxConcurrency)
	d.Set("max_errors", resp.MaxErrors)
	d.Set("task_type", resp.TaskType)
	d.Set("service_role_arn", resp.ServiceRoleArn)
	d.Set("task_arn", resp.TaskArn)
	d.Set("priority", resp.Priority)
	d.Set("name", resp.Name)
	d.Set("description", resp.Description)

	if resp.TaskInvocationParameters != nil {
		if err := d.Set("task_invocation_parameters", flattenAwsSsmTaskInvocationParameters(resp.TaskInvocationParameters)); err != nil {
			return fmt.Errorf("Error setting task_invocation_parameters error: %#v", err)
		}
	}

	if err := d.Set("targets", flattenAwsSsmTargets(resp.Targets)); err != nil {
		return fmt.Errorf("Error setting targets error: %#v", err)
	}

	return nil
}

func resourceAwsSsmMaintenanceWindowTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn
	windowID := d.Get("window_id").(string)

	params := &ssm.UpdateMaintenanceWindowTaskInput{
		WindowId:       aws.String(windowID),
		WindowTaskId:   aws.String(d.Id()),
		MaxConcurrency: aws.String(d.Get("max_concurrency").(string)),
		MaxErrors:      aws.String(d.Get("max_errors").(string)),
		ServiceRoleArn: aws.String(d.Get("service_role_arn").(string)),
		TaskArn:        aws.String(d.Get("task_arn").(string)),
		Targets:        expandAwsSsmTargets(d.Get("targets").([]interface{})),
		Replace:        aws.Bool(true),
	}

	if v, ok := d.GetOk("name"); ok {
		params.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("priority"); ok {
		params.Priority = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("task_invocation_parameters"); ok {
		params.TaskInvocationParameters = expandAwsSsmTaskInvocationParameters(v.([]interface{}))
	}

	_, err := ssmconn.UpdateMaintenanceWindowTask(params)
	if isAWSErr(err, ssm.ErrCodeDoesNotExistException, "") {
		log.Printf("[WARN] Maintenance Window (%s) Task (%s) not found, removing from state", windowID, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error updating Maintenance Window (%s) Task (%s): %s", windowID, d.Id(), err)
	}

	return resourceAwsSsmMaintenanceWindowTaskRead(d, meta)
}

func resourceAwsSsmMaintenanceWindowTaskDelete(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Deregistering SSM Maintenance Window Task: %s", d.Id())

	params := &ssm.DeregisterTaskFromMaintenanceWindowInput{
		WindowId:     aws.String(d.Get("window_id").(string)),
		WindowTaskId: aws.String(d.Id()),
	}

	_, err := ssmconn.DeregisterTaskFromMaintenanceWindow(params)
	if isAWSErr(err, ssm.ErrCodeDoesNotExistException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deregistering SSM Maintenance Window Task (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsSsmMaintenanceWindowTaskImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <window-id>/<window-task-id>", d.Id())
	}

	windowID := idParts[0]
	windowTaskID := idParts[1]

	d.Set("window_id", windowID)
	d.SetId(windowTaskID)

	return []*schema.ResourceData{d}, nil
}
