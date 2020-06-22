package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsApiGatewayMethodSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayMethodSettingsUpdate,
		Read:   resourceAwsApiGatewayMethodSettingsRead,
		Update: resourceAwsApiGatewayMethodSettingsUpdate,
		Delete: resourceAwsApiGatewayMethodSettingsDelete,

		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stage_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"method_path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metrics_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"logging_level": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"data_trace_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"throttling_burst_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"throttling_rate_limit": {
							Type:     schema.TypeFloat,
							Optional: true,
							Computed: true,
						},
						"caching_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"cache_ttl_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"cache_data_encrypted": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"require_authorization_for_cache_control": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"unauthorized_cache_control_header_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func flattenAwsApiGatewayMethodSettings(settings *apigateway.MethodSetting) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"metrics_enabled":                            settings.MetricsEnabled,
			"logging_level":                              settings.LoggingLevel,
			"data_trace_enabled":                         settings.DataTraceEnabled,
			"throttling_burst_limit":                     settings.ThrottlingBurstLimit,
			"throttling_rate_limit":                      settings.ThrottlingRateLimit,
			"caching_enabled":                            settings.CachingEnabled,
			"cache_ttl_in_seconds":                       settings.CacheTtlInSeconds,
			"cache_data_encrypted":                       settings.CacheDataEncrypted,
			"require_authorization_for_cache_control":    settings.RequireAuthorizationForCacheControl,
			"unauthorized_cache_control_header_strategy": settings.UnauthorizedCacheControlHeaderStrategy,
		},
	}
}

func resourceAwsApiGatewayMethodSettingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn

	log.Printf("[DEBUG] Reading API Gateway Method Settings %s", d.Id())
	input := apigateway.GetStageInput{
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
		StageName: aws.String(d.Get("stage_name").(string)),
	}
	stage, err := conn.GetStage(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			log.Printf("[WARN] API Gateway Stage (%s) not found, removing method settings", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	log.Printf("[DEBUG] Received API Gateway Stage: %s", stage)

	methodPath := d.Get("method_path").(string)
	settings, ok := stage.MethodSettings[methodPath]
	if !ok {
		log.Printf("[WARN] API Gateway Method Settings for %q not found, removing", methodPath)
		d.SetId("")
		return nil
	}

	if err := d.Set("settings", flattenAwsApiGatewayMethodSettings(settings)); err != nil {
		return fmt.Errorf("error setting settings: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayMethodSettingsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn

	methodPath := d.Get("method_path").(string)
	prefix := fmt.Sprintf("/%s/", methodPath)

	ops := make([]*apigateway.PatchOperation, 0)
	if d.HasChange("settings.0.metrics_enabled") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String(prefix + "metrics/enabled"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.metrics_enabled").(bool))),
		})
	}
	if d.HasChange("settings.0.logging_level") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String(prefix + "logging/loglevel"),
			Value: aws.String(d.Get("settings.0.logging_level").(string)),
		})
	}
	if d.HasChange("settings.0.data_trace_enabled") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String(prefix + "logging/dataTrace"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.data_trace_enabled").(bool))),
		})
	}

	if d.HasChange("settings.0.throttling_burst_limit") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String(prefix + "throttling/burstLimit"),
			Value: aws.String(fmt.Sprintf("%d", d.Get("settings.0.throttling_burst_limit").(int))),
		})
	}
	if d.HasChange("settings.0.throttling_rate_limit") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String(prefix + "throttling/rateLimit"),
			Value: aws.String(fmt.Sprintf("%f", d.Get("settings.0.throttling_rate_limit").(float64))),
		})
	}
	if d.HasChange("settings.0.caching_enabled") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String(prefix + "caching/enabled"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.caching_enabled").(bool))),
		})
	}
	if d.HasChange("settings.0.cache_ttl_in_seconds") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String(prefix + "caching/ttlInSeconds"),
			Value: aws.String(fmt.Sprintf("%d", d.Get("settings.0.cache_ttl_in_seconds").(int))),
		})
	}
	if d.HasChange("settings.0.cache_data_encrypted") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String(prefix + "caching/dataEncrypted"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.cache_data_encrypted").(bool))),
		})
	}
	if d.HasChange("settings.0.require_authorization_for_cache_control") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String(prefix + "caching/requireAuthorizationForCacheControl"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("settings.0.require_authorization_for_cache_control").(bool))),
		})
	}
	if d.HasChange("settings.0.unauthorized_cache_control_header_strategy") {
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String(prefix + "caching/unauthorizedCacheControlHeaderStrategy"),
			Value: aws.String(d.Get("settings.0.unauthorized_cache_control_header_strategy").(string)),
		})
	}

	restApiId := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	input := apigateway.UpdateStageInput{
		RestApiId:       aws.String(restApiId),
		StageName:       aws.String(stageName),
		PatchOperations: ops,
	}
	log.Printf("[DEBUG] Updating API Gateway Stage: %s", input)
	_, err := conn.UpdateStage(&input)
	if err != nil {
		return fmt.Errorf("Updating API Gateway Stage failed: %s", err)
	}

	d.SetId(restApiId + "-" + stageName + "-" + methodPath)

	return resourceAwsApiGatewayMethodSettingsRead(d, meta)
}

func resourceAwsApiGatewayMethodSettingsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	log.Printf("[DEBUG] Deleting API Gateway Method Settings: %s", d.Id())

	input := apigateway.UpdateStageInput{
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
		StageName: aws.String(d.Get("stage_name").(string)),
		PatchOperations: []*apigateway.PatchOperation{
			{
				Op:   aws.String("remove"),
				Path: aws.String(fmt.Sprintf("/%s", d.Get("method_path").(string))),
			},
		},
	}
	log.Printf("[DEBUG] Updating API Gateway Stage: %s", input)
	_, err := conn.UpdateStage(&input)
	if err != nil {
		return fmt.Errorf("Updating API Gateway Stage failed: %s", err)
	}

	return nil
}
