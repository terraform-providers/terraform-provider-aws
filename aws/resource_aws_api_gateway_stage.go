package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigateway/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigateway/waiter"
)

func resourceAwsApiGatewayStage() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayStageCreate,
		Read:   resourceAwsApiGatewayStageRead,
		Update: resourceAwsApiGatewayStageUpdate,
		Delete: resourceAwsApiGatewayStageDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/STAGE-NAME", d.Id())
				}
				restApiID := idParts[0]
				stageName := idParts[1]
				d.Set("stage_name", stageName)
				d.Set("rest_api_id", restApiID)
				d.SetId(fmt.Sprintf("ags-%s-%s", restApiID, stageName))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"access_log_settings": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"format": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"cache_cluster_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cache_cluster_size": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(apigateway.CacheClusterSize_Values(), true),
			},
			"client_certificate_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"documentation_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invoke_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
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
			"variables": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"xray_tracing_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"web_acl_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsApiGatewayStageCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	restApiId := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	input := apigateway.CreateStageInput{
		RestApiId:    aws.String(restApiId),
		StageName:    aws.String(stageName),
		DeploymentId: aws.String(d.Get("deployment_id").(string)),
	}

	waitForCache := false
	if v, ok := d.GetOk("cache_cluster_enabled"); ok {
		input.CacheClusterEnabled = aws.Bool(v.(bool))
		waitForCache = true
	}
	if v, ok := d.GetOk("cache_cluster_size"); ok {
		input.CacheClusterSize = aws.String(v.(string))
		waitForCache = true
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("xray_tracing_enabled"); ok {
		input.TracingEnabled = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("documentation_version"); ok {
		input.DocumentationVersion = aws.String(v.(string))
	}
	if vars, ok := d.GetOk("variables"); ok {
		variables := make(map[string]string)
		for k, v := range vars.(map[string]interface{}) {
			variables[k] = v.(string)
		}
		input.Variables = aws.StringMap(variables)
	}
	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().ApigatewayTags()
	}

	out, err := conn.CreateStage(&input)
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Stage: %w", err)
	}

	d.SetId(fmt.Sprintf("ags-%s-%s", restApiId, stageName))

	if waitForCache && out != nil && aws.StringValue(out.CacheClusterStatus) != apigateway.CacheClusterStatusNotAvailable {
		_, err = waiter.StageCacheAvailable(conn, restApiId, stageName)
		if err != nil {
			return fmt.Errorf("error waiting for API Gateway Stage (%s) to be available: %w", d.Id(), err)
		}
	}

	if _, ok := d.GetOk("client_certificate_id"); ok {
		return resourceAwsApiGatewayStageUpdate(d, meta)
	}
	if _, ok := d.GetOk("access_log_settings"); ok {
		return resourceAwsApiGatewayStageUpdate(d, meta)
	}
	return resourceAwsApiGatewayStageRead(d, meta)
}

func resourceAwsApiGatewayStageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading API Gateway Stage %s", d.Id())
	restApiId := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	stage, err := finder.StageByName(conn, restApiId, stageName)

	if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway Stage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting API Gateway REST API (%s) Stage (%s): %w", restApiId, stageName, err)
	}

	log.Printf("[DEBUG] Received API Gateway Stage: %s", stage)

	if err := d.Set("access_log_settings", flattenApiGatewayStageAccessLogSettings(stage.AccessLogSettings)); err != nil {
		return fmt.Errorf("error setting access_log_settings: %w", err)
	}

	d.Set("client_certificate_id", stage.ClientCertificateId)

	if aws.StringValue(stage.CacheClusterStatus) == apigateway.CacheClusterStatusDeleteInProgress {
		d.Set("cache_cluster_enabled", false)
		d.Set("cache_cluster_size", nil)
	} else {
		d.Set("cache_cluster_enabled", stage.CacheClusterEnabled)
		d.Set("cache_cluster_size", stage.CacheClusterSize)
	}

	d.Set("deployment_id", stage.DeploymentId)
	d.Set("description", stage.Description)
	d.Set("documentation_version", stage.DocumentationVersion)
	d.Set("xray_tracing_enabled", stage.TracingEnabled)
	d.Set("web_acl_arn", stage.WebAclArn)

	tags := keyvaluetags.ApigatewayKeyValueTags(stage.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	stageArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "apigateway",
		Resource:  fmt.Sprintf("/restapis/%s/stages/%s", restApiId, stageName),
	}.String()
	d.Set("arn", stageArn)

	if err := d.Set("variables", aws.StringValueMap(stage.Variables)); err != nil {
		return fmt.Errorf("error setting variables: %w", err)
	}

	d.Set("invoke_url", buildApiGatewayInvokeURL(meta.(*AWSClient), restApiId, stageName))

	executionArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "execute-api",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("%s/%s", restApiId, stageName),
	}.String()
	d.Set("execution_arn", executionArn)

	return nil
}

func resourceAwsApiGatewayStageUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.ApigatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	if d.HasChangeExcept("tags") {
		operations := make([]*apigateway.PatchOperation, 0)
		waitForCache := false
		if d.HasChange("cache_cluster_enabled") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/cacheClusterEnabled"),
				Value: aws.String(fmt.Sprintf("%t", d.Get("cache_cluster_enabled").(bool))),
			})
			waitForCache = true
		}
		if d.HasChange("cache_cluster_size") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/cacheClusterSize"),
				Value: aws.String(d.Get("cache_cluster_size").(string)),
			})
			waitForCache = true
		}
		if d.HasChange("client_certificate_id") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/clientCertificateId"),
				Value: aws.String(d.Get("client_certificate_id").(string)),
			})
		}
		if d.HasChange("deployment_id") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/deploymentId"),
				Value: aws.String(d.Get("deployment_id").(string)),
			})
		}
		if d.HasChange("description") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/description"),
				Value: aws.String(d.Get("description").(string)),
			})
		}
		if d.HasChange("xray_tracing_enabled") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/tracingEnabled"),
				Value: aws.String(fmt.Sprintf("%t", d.Get("xray_tracing_enabled").(bool))),
			})
		}
		if d.HasChange("documentation_version") {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/documentationVersion"),
				Value: aws.String(d.Get("documentation_version").(string)),
			})
		}
		if d.HasChange("variables") {
			o, n := d.GetChange("variables")
			oldV := o.(map[string]interface{})
			newV := n.(map[string]interface{})
			operations = append(operations, diffVariablesOps(oldV, newV)...)
		}
		if d.HasChange("access_log_settings") {
			accessLogSettings := d.Get("access_log_settings").([]interface{})
			if len(accessLogSettings) == 1 {
				operations = append(operations,
					&apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpReplace),
						Path:  aws.String("/accessLogSettings/destinationArn"),
						Value: aws.String(d.Get("access_log_settings.0.destination_arn").(string)),
					}, &apigateway.PatchOperation{
						Op:    aws.String(apigateway.OpReplace),
						Path:  aws.String("/accessLogSettings/format"),
						Value: aws.String(d.Get("access_log_settings.0.format").(string)),
					})
			} else if len(accessLogSettings) == 0 {
				operations = append(operations, &apigateway.PatchOperation{
					Op:   aws.String(apigateway.OpRemove),
					Path: aws.String("/accessLogSettings"),
				})
			}
		}

		restApiId := d.Get("rest_api_id").(string)
		stageName := d.Get("stage_name").(string)

		input := apigateway.UpdateStageInput{
			RestApiId:       aws.String(restApiId),
			StageName:       aws.String(stageName),
			PatchOperations: operations,
		}
		log.Printf("[DEBUG] Updating API Gateway Stage: %s", input)
		out, err := conn.UpdateStage(&input)
		if err != nil {
			return fmt.Errorf("Updating API Gateway Stage failed: %w", err)
		}

		if waitForCache && out != nil && aws.StringValue(out.CacheClusterStatus) != apigateway.CacheClusterStatusNotAvailable {
			_, err = waiter.StageCacheUpdated(conn, restApiId, stageName)
			if err != nil {
				return fmt.Errorf("error waiting for API Gateway Stage (%s) to be updated: %w", d.Id(), err)
			}
		}
	}

	return resourceAwsApiGatewayStageRead(d, meta)
}

func diffVariablesOps(oldVars, newVars map[string]interface{}) []*apigateway.PatchOperation {
	ops := make([]*apigateway.PatchOperation, 0)
	prefix := "/variables/"

	for k := range oldVars {
		if _, ok := newVars[k]; !ok {
			ops = append(ops, &apigateway.PatchOperation{
				Op:   aws.String(apigateway.OpRemove),
				Path: aws.String(prefix + k),
			})
		}
	}

	for k, v := range newVars {
		newValue := v.(string)

		if oldV, ok := oldVars[k]; ok {
			oldValue := oldV.(string)
			if oldValue == newValue {
				continue
			}
		}
		ops = append(ops, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String(prefix + k),
			Value: aws.String(newValue),
		})
	}

	return ops
}

func resourceAwsApiGatewayStageDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	log.Printf("[DEBUG] Deleting API Gateway Stage: %s", d.Id())
	restApiId := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	input := apigateway.DeleteStageInput{
		RestApiId: aws.String(restApiId),
		StageName: aws.String(stageName),
	}
	_, err := conn.DeleteStage(&input)

	if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway REST API (%s) Stage (%s): %w", restApiId, stageName, err)
	}

	return nil
}

func flattenApiGatewayStageAccessLogSettings(accessLogSettings *apigateway.AccessLogSettings) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)
	if accessLogSettings != nil {
		result = append(result, map[string]interface{}{
			"destination_arn": aws.StringValue(accessLogSettings.DestinationArn),
			"format":          aws.StringValue(accessLogSettings.Format),
		})
	}
	return result
}
