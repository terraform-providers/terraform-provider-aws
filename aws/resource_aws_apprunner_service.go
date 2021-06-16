package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apprunner/waiter"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsAppRunnerService() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsAppRunnerServiceCreate,
		ReadWithoutTimeout:   resourceAwsAppRunnerServiceRead,
		UpdateWithoutTimeout: resourceAwsAppRunnerServiceUpdate,
		DeleteWithoutTimeout: resourceAwsAppRunnerServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"auto_scaling_configuration_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},

			"encryption_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"health_check_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"healthy_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						"interval": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						"path": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "/",
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(0, 51200),
						},
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      apprunner.HealthCheckProtocolTcp,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(apprunner.HealthCheckProtocol_Values(), false),
						},
						"timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      2,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						"unhealthy_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
					},
				},
			},

			"instance_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "1024",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`1024|2048|(1|2) vCPU`), ""),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// App Runner API always returns the amount in multiples of 1024 units
								return (old == "1024" && new == "1 vCPU") || (old == "2048" && new == "2 vCPU")
							},
						},
						"instance_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"memory": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "2048",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`2048|3072|4096|(2|3|4) GB`), ""),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// App Runner API always returns the amount in MB
								return (old == "2048" && new == "2 GB") || (old == "3072" && new == "3 GB") || (old == "4096" && new == "4 GB")
							},
						},
					},
				},
			},

			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"service_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"source_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_role_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
									"connection_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
						"auto_deployments_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"code_repository": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"code_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"code_configuration_values": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"build_command": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 51200),
															},
															"port": {
																Type:         schema.TypeString,
																Optional:     true,
																Default:      "8080",
																ValidateFunc: validation.StringLenBetween(0, 51200),
															},
															"runtime": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(apprunner.Runtime_Values(), false),
															},
															"runtime_environment_variables": {
																Type:     schema.TypeMap,
																Optional: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: validation.StringLenBetween(0, 51200),
																},
															},
															"start_command": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 51200),
															},
														},
													},
												},
												"configuration_source": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(apprunner.ConfigurationSource_Values(), false),
												},
											},
										},
									},
									"repository_url": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 51200),
									},
									"source_code_version": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(apprunner.SourceCodeVersionType_Values(), false),
												},
												"value": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 51200),
												},
											},
										},
									},
								},
							},
							ExactlyOneOf: []string{"source_configuration.0.code_repository", "source_configuration.0.image_repository"},
						},
						"image_repository": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:         schema.TypeString,
													Optional:     true,
													Default:      "8080",
													ValidateFunc: validation.StringLenBetween(0, 51200),
												},
												"runtime_environment_variables": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringLenBetween(0, 51200),
													},
												},
												"start_command": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 51200),
												},
											},
										},
									},
									"image_identifier": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`([0-9]{12}.dkr.ecr.[a-z\-]+-[0-9]{1}.amazonaws.com\/.*)|(^public\.ecr\.aws\/.+\/.+)`), ""),
									},
									"image_repository_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(apprunner.ImageRepositoryType_Values(), false),
									},
								},
							},
							ExactlyOneOf: []string{"source_configuration.0.image_repository", "source_configuration.0.code_repository"},
						},
					},
				},
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),

			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsAppRunnerServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).apprunnerconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	serviceName := d.Get("service_name").(string)

	input := &apprunner.CreateServiceInput{
		ServiceName:         aws.String(serviceName),
		SourceConfiguration: expandAppRunnerServiceSourceConfiguration(d.Get("source_configuration").([]interface{})),
		Tags:                tags.IgnoreAws().ApprunnerTags(),
	}

	if v, ok := d.GetOk("auto_scaling_configuration_arn"); ok {
		input.AutoScalingConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encryption_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EncryptionConfiguration = expandAppRunnerServiceEncryptionConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("health_check_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.HealthCheckConfiguration = expandAppRunnerServiceHealthCheckConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("instance_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.InstanceConfiguration = expandAppRunnerServiceInstanceConfiguration(v.([]interface{}))
	}

	var output *apprunner.CreateServiceOutput

	err := resource.RetryContext(ctx, iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.CreateServiceWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, apprunner.ErrCodeInvalidRequestException, "Error in assuming instance role") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateServiceWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating App Runner Service (%s): %w", serviceName, err))
	}

	if output == nil || output.Service == nil {
		return diag.FromErr(fmt.Errorf("error creating App Runner Service (%s): empty output", serviceName))
	}

	d.SetId(aws.StringValue(output.Service.ServiceArn))

	if err := waiter.ServiceCreated(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for App Runner Service (%s) creation: %w", d.Id(), err))
	}

	return resourceAwsAppRunnerServiceRead(ctx, d, meta)
}

func resourceAwsAppRunnerServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).apprunnerconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &apprunner.DescribeServiceInput{
		ServiceArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeServiceWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] App Runner Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading App Runner Service (%s): %w", d.Id(), err))
	}

	if output == nil || output.Service == nil {
		return diag.FromErr(fmt.Errorf("error reading App Runner Service (%s): empty output", d.Id()))
	}

	if aws.StringValue(output.Service.Status) == apprunner.ServiceStatusDeleted {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading App Runner Service (%s): %s after creation", d.Id(), aws.StringValue(output.Service.Status)))
		}
		log.Printf("[WARN] App Runner Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	service := output.Service
	arn := aws.StringValue(service.ServiceArn)

	var autoScalingConfigArn string
	if service.AutoScalingConfigurationSummary != nil {
		autoScalingConfigArn = aws.StringValue(service.AutoScalingConfigurationSummary.AutoScalingConfigurationArn)
	}

	d.Set("arn", arn)
	d.Set("auto_scaling_configuration_arn", autoScalingConfigArn)
	d.Set("service_id", service.ServiceId)
	d.Set("service_name", service.ServiceName)
	d.Set("service_url", service.ServiceUrl)
	d.Set("status", service.Status)
	if err := d.Set("encryption_configuration", flattenAppRunnerServiceEncryptionConfiguration(service.EncryptionConfiguration)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting encryption_configuration: %w", err))
	}

	if err := d.Set("health_check_configuration", flattenAppRunnerServiceHealthCheckConfiguration(service.HealthCheckConfiguration)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting health_check_configuration: %w", err))
	}

	if err := d.Set("instance_configuration", flattenAppRunnerServiceInstanceConfiguration(service.InstanceConfiguration)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting instance_configuration: %w", err))
	}

	if err := d.Set("source_configuration", flattenAppRunnerServiceSourceConfiguration(service.SourceConfiguration)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting source_configuration: %w", err))
	}

	tags, err := keyvaluetags.ApprunnerListTags(conn, arn)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for App Runner Service (%s): %s", arn, err))
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceAwsAppRunnerServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).apprunnerconn

	if d.HasChanges(
		"auto_scaling_configuration_arn",
		"instance_configuration",
		"source_configuration",
	) {
		input := &apprunner.UpdateServiceInput{
			ServiceArn: aws.String(d.Id()),
		}

		if d.HasChange("auto_scaling_configuration_arn") {
			input.AutoScalingConfigurationArn = aws.String(d.Get("auto_scaling_configuration_arn").(string))
		}

		if d.HasChange("instance_configuration") {
			input.InstanceConfiguration = expandAppRunnerServiceInstanceConfiguration(d.Get("instance_configuration").([]interface{}))
		}

		if d.HasChange("source_configuration") {
			input.SourceConfiguration = expandAppRunnerServiceSourceConfiguration(d.Get("source_configuration").([]interface{}))
		}

		_, err := conn.UpdateServiceWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating App Runner Service (%s): %w", d.Id(), err))
		}

		if err := waiter.ServiceUpdated(ctx, conn, d.Id()); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for App Runner Service (%s) to update: %w", d.Id(), err))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.ApprunnerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating App Runner Service (%s) tags: %s", d.Get("arn").(string), err))
		}
	}

	return resourceAwsAppRunnerServiceRead(ctx, d, meta)
}

func resourceAwsAppRunnerServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).apprunnerconn

	input := &apprunner.DeleteServiceInput{
		ServiceArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteServiceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting App Runner Service (%s): %w", d.Id(), err))
	}

	if err := waiter.ServiceDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}

		return diag.FromErr(fmt.Errorf("error waiting for App Runner Service (%s) deletion: %w", d.Id(), err))
	}

	return nil
}

func expandAppRunnerServiceEncryptionConfiguration(l []interface{}) *apprunner.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.EncryptionConfiguration{}

	if v, ok := tfMap["kms_key"].(string); ok && v != "" {
		result.KmsKey = aws.String(v)
	}

	return result
}

func expandAppRunnerServiceHealthCheckConfiguration(l []interface{}) *apprunner.HealthCheckConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.HealthCheckConfiguration{}

	if v, ok := tfMap["healthy_threshold"].(int); ok {
		result.HealthyThreshold = aws.Int64(int64(v))
	}

	if v, ok := tfMap["interval"].(int); ok {
		result.Interval = aws.Int64(int64(v))
	}

	if v, ok := tfMap["path"].(string); ok {
		result.Path = aws.String(v)
	}

	if v, ok := tfMap["protocol"].(string); ok {
		result.Protocol = aws.String(v)
	}

	if v, ok := tfMap["timeout"].(int); ok {
		result.Timeout = aws.Int64(int64(v))
	}

	if v, ok := tfMap["unhealthy_threshold"].(int); ok {
		result.UnhealthyThreshold = aws.Int64(int64(v))
	}

	return result
}

func expandAppRunnerServiceInstanceConfiguration(l []interface{}) *apprunner.InstanceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.InstanceConfiguration{}

	if v, ok := tfMap["cpu"].(string); ok {
		result.Cpu = aws.String(v)
	}

	if v, ok := tfMap["instance_role_arn"].(string); ok {
		result.InstanceRoleArn = aws.String(v)
	}

	if v, ok := tfMap["memory"].(string); ok {
		result.Memory = aws.String(v)
	}

	return result
}

func expandAppRunnerServiceSourceConfiguration(l []interface{}) *apprunner.SourceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.SourceConfiguration{}

	if v, ok := tfMap["authentication_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.AuthenticationConfiguration = expandAppRunnerServiceAuthenticationConfiguration(v)
	}

	if v, ok := tfMap["auto_deployments_enabled"].(bool); ok {
		result.AutoDeploymentsEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["code_repository"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.CodeRepository = expandAppRunnerServiceCodeRepository(v)
	}

	if v, ok := tfMap["image_repository"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ImageRepository = expandAppRunnerServiceImageRepository(v)
	}

	return result
}

func expandAppRunnerServiceAuthenticationConfiguration(l []interface{}) *apprunner.AuthenticationConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.AuthenticationConfiguration{}

	if v, ok := tfMap["access_role_arn"].(string); ok && v != "" {
		result.AccessRoleArn = aws.String(v)
	}

	if v, ok := tfMap["connection_arn"].(string); ok && v != "" {
		result.ConnectionArn = aws.String(v)
	}

	return result
}

func expandAppRunnerServiceImageConfiguration(l []interface{}) *apprunner.ImageConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.ImageConfiguration{}

	if v, ok := tfMap["port"].(string); ok && v != "" {
		result.Port = aws.String(v)
	}

	if v, ok := tfMap["runtime_environment_variables"].(map[string]interface{}); ok && len(v) > 0 {
		result.RuntimeEnvironmentVariables = expandStringMap(v)
	}

	if v, ok := tfMap["start_command"].(string); ok && v != "" {
		result.StartCommand = aws.String(v)
	}

	return result
}

func expandAppRunnerServiceCodeRepository(l []interface{}) *apprunner.CodeRepository {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.CodeRepository{}

	if v, ok := tfMap["source_code_version"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.SourceCodeVersion = expandAppRunnerServiceSourceCodeVersion(v)
	}

	if v, ok := tfMap["code_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.CodeConfiguration = expandAppRunnerServiceCodeConfiguration(v)
	}

	if v, ok := tfMap["repository_url"].(string); ok && v != "" {
		result.RepositoryUrl = aws.String(v)
	}

	return result
}

func expandAppRunnerServiceImageRepository(l []interface{}) *apprunner.ImageRepository {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.ImageRepository{}

	if v, ok := tfMap["image_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ImageConfiguration = expandAppRunnerServiceImageConfiguration(v)
	}

	if v, ok := tfMap["image_identifier"].(string); ok && v != "" {
		result.ImageIdentifier = aws.String(v)
	}

	if v, ok := tfMap["image_repository_type"].(string); ok && v != "" {
		result.ImageRepositoryType = aws.String(v)
	}

	return result
}

func expandAppRunnerServiceCodeConfiguration(l []interface{}) *apprunner.CodeConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.CodeConfiguration{}

	if v, ok := tfMap["configuration_source"].(string); ok && v != "" {
		result.ConfigurationSource = aws.String(v)
	}

	if v, ok := tfMap["code_configuration_values"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.CodeConfigurationValues = expandAppRunnerServiceCodeConfigurationValues(v)
	}

	return result
}

func expandAppRunnerServiceCodeConfigurationValues(l []interface{}) *apprunner.CodeConfigurationValues {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.CodeConfigurationValues{}

	if v, ok := tfMap["build_command"].(string); ok && v != "" {
		result.BuildCommand = aws.String(v)
	}

	if v, ok := tfMap["port"].(string); ok && v != "" {
		result.Port = aws.String(v)
	}

	if v, ok := tfMap["runtime"].(string); ok && v != "" {
		result.Runtime = aws.String(v)
	}

	if v, ok := tfMap["runtime_environment_variables"].(map[string]interface{}); ok && len(v) > 0 {
		result.RuntimeEnvironmentVariables = expandStringMap(v)
	}

	if v, ok := tfMap["start_command"].(string); ok && v != "" {
		result.StartCommand = aws.String(v)
	}

	return result
}

func expandAppRunnerServiceSourceCodeVersion(l []interface{}) *apprunner.SourceCodeVersion {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.SourceCodeVersion{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		result.Type = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		result.Value = aws.String(v)
	}

	return result
}

func flattenAppRunnerServiceEncryptionConfiguration(config *apprunner.EncryptionConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"kms_key": aws.StringValue(config.KmsKey),
	}

	return []interface{}{m}
}

func flattenAppRunnerServiceHealthCheckConfiguration(config *apprunner.HealthCheckConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"healthy_threshold":   aws.Int64Value(config.HealthyThreshold),
		"interval":            aws.Int64Value(config.Interval),
		"path":                aws.StringValue(config.Path),
		"protocol":            aws.StringValue(config.Protocol),
		"timeout":             aws.Int64Value(config.Timeout),
		"unhealthy_threshold": aws.Int64Value(config.UnhealthyThreshold),
	}

	return []interface{}{m}
}

func flattenAppRunnerServiceInstanceConfiguration(config *apprunner.InstanceConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cpu":               aws.StringValue(config.Cpu),
		"instance_role_arn": aws.StringValue(config.InstanceRoleArn),
		"memory":            aws.StringValue(config.Memory),
	}

	return []interface{}{m}
}

func flattenAppRunnerServiceCodeRepository(r *apprunner.CodeRepository) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"code_configuration":  flattenAppRunnerServiceCodeConfiguration(r.CodeConfiguration),
		"repository_url":      aws.StringValue(r.RepositoryUrl),
		"source_code_version": flattenAppRunnerServiceSourceCodeVersion(r.SourceCodeVersion),
	}

	return []interface{}{m}
}

func flattenAppRunnerServiceCodeConfiguration(config *apprunner.CodeConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"code_configuration_values": flattenAppRunnerServiceCodeConfigurationValues(config.CodeConfigurationValues),
		"configuration_source":      aws.StringValue(config.ConfigurationSource),
	}

	return []interface{}{m}
}

func flattenAppRunnerServiceCodeConfigurationValues(values *apprunner.CodeConfigurationValues) []interface{} {
	if values == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"build_command":                 aws.StringValue(values.BuildCommand),
		"port":                          aws.StringValue(values.Port),
		"runtime":                       aws.StringValue(values.Runtime),
		"runtime_environment_variables": aws.StringValueMap(values.RuntimeEnvironmentVariables),
		"start_command":                 aws.StringValue(values.StartCommand),
	}

	return []interface{}{m}
}

func flattenAppRunnerServiceSourceCodeVersion(v *apprunner.SourceCodeVersion) []interface{} {
	if v == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"type":  aws.StringValue(v.Type),
		"value": aws.StringValue(v.Value),
	}

	return []interface{}{m}
}

func flattenAppRunnerServiceSourceConfiguration(config *apprunner.SourceConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"authentication_configuration": flattenAppRunnerServiceAuthenticationConfiguration(config.AuthenticationConfiguration),
		"auto_deployments_enabled":     aws.BoolValue(config.AutoDeploymentsEnabled),
		"code_repository":              flattenAppRunnerServiceCodeRepository(config.CodeRepository),
		"image_repository":             flattenAppRunnerServiceImageRepository(config.ImageRepository),
	}

	return []interface{}{m}
}

func flattenAppRunnerServiceAuthenticationConfiguration(config *apprunner.AuthenticationConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"access_role_arn": aws.StringValue(config.AccessRoleArn),
		"connection_arn":  aws.StringValue(config.ConnectionArn),
	}

	return []interface{}{m}
}

func flattenAppRunnerServiceImageConfiguration(config *apprunner.ImageConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"port":                          aws.StringValue(config.Port),
		"runtime_environment_variables": aws.StringValueMap(config.RuntimeEnvironmentVariables),
		"start_command":                 aws.StringValue(config.StartCommand),
	}

	return []interface{}{m}
}

func flattenAppRunnerServiceImageRepository(r *apprunner.ImageRepository) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"image_configuration":   flattenAppRunnerServiceImageConfiguration(r.ImageConfiguration),
		"image_identifier":      aws.StringValue(r.ImageIdentifier),
		"image_repository_type": aws.StringValue(r.ImageRepositoryType),
	}

	return []interface{}{m}
}
