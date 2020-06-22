package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsApiGatewayV2Api() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2ApiCreate,
		Read:   resourceAwsApiGatewayV2ApiRead,
		Update: resourceAwsApiGatewayV2ApiUpdate,
		Delete: resourceAwsApiGatewayV2ApiDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_key_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "$request.header.x-api-key",
				ValidateFunc: validation.StringInSlice([]string{
					"$context.authorizer.usageIdentifierKey",
					"$request.header.x-api-key",
				}, false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cors_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_credentials": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"allow_headers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
						},
						"allow_methods": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
						},
						"allow_origins": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
						},
						"expose_headers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
						},
						"max_age": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"credentials_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"protocol_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					apigatewayv2.ProtocolTypeHttp,
					apigatewayv2.ProtocolTypeWebsocket,
				}, false),
			},
			"route_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"route_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "$request.method $request.path",
			},
			"tags": tagsSchema(),
			"target": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceAwsApiGatewayV2ApiCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	protocolType := d.Get("protocol_type").(string)
	req := &apigatewayv2.CreateApiInput{
		Name:         aws.String(d.Get("name").(string)),
		ProtocolType: aws.String(protocolType),
		Tags:         keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().Apigatewayv2Tags(),
	}
	if v, ok := d.GetOk("api_key_selection_expression"); ok {
		req.ApiKeySelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("cors_configuration"); ok {
		req.CorsConfiguration = expandApiGateway2CorsConfiguration(v.([]interface{}))
	}
	if v, ok := d.GetOk("credentials_arn"); ok {
		req.CredentialsArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("route_key"); ok {
		req.RouteKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("route_selection_expression"); ok {
		req.RouteSelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target"); ok {
		req.Target = aws.String(v.(string))
	}
	if v, ok := d.GetOk("version"); ok {
		req.Version = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 API: %s", req)
	resp, err := conn.CreateApi(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 API: %s", err)
	}

	d.SetId(aws.StringValue(resp.ApiId))

	return resourceAwsApiGatewayV2ApiRead(d, meta)
}

func resourceAwsApiGatewayV2ApiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetApi(&apigatewayv2.GetApiInput{
		ApiId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 API (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 API (%s): %s", d.Id(), err)
	}

	d.Set("api_endpoint", resp.ApiEndpoint)
	d.Set("api_key_selection_expression", resp.ApiKeySelectionExpression)
	apiArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/apis/%s", d.Id()),
	}.String()
	d.Set("arn", apiArn)
	if err := d.Set("cors_configuration", flattenApiGateway2CorsConfiguration(resp.CorsConfiguration)); err != nil {
		return fmt.Errorf("error setting cors_configuration: %s", err)
	}
	d.Set("description", resp.Description)
	executionArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "execute-api",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  d.Id(),
	}.String()
	d.Set("execution_arn", executionArn)
	d.Set("name", resp.Name)
	d.Set("protocol_type", resp.ProtocolType)
	d.Set("route_selection_expression", resp.RouteSelectionExpression)
	if err := d.Set("tags", keyvaluetags.Apigatewayv2KeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}
	d.Set("version", resp.Version)

	return nil
}

func resourceAwsApiGatewayV2ApiUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	deleteCorsConfiguration := false
	if d.HasChange("cors_configuration") {
		v := d.Get("cors_configuration")
		if len(v.([]interface{})) == 0 {
			deleteCorsConfiguration = true

			log.Printf("[DEBUG] Deleting CORS configuration for API Gateway v2 API (%s)", d.Id())
			_, err := conn.DeleteCorsConfiguration(&apigatewayv2.DeleteCorsConfigurationInput{
				ApiId: aws.String(d.Id()),
			})
			if err != nil {
				return fmt.Errorf("error deleting CORS configuration for API Gateway v2 API (%s): %s", d.Id(), err)
			}
		}
	}

	if d.HasChanges("api_key_selection_expression", "description", "name", "route_selection_expression", "version") ||
		(d.HasChange("cors_configuration") && !deleteCorsConfiguration) {
		req := &apigatewayv2.UpdateApiInput{
			ApiId: aws.String(d.Id()),
		}

		if d.HasChange("api_key_selection_expression") {
			req.ApiKeySelectionExpression = aws.String(d.Get("api_key_selection_expression").(string))
		}
		if d.HasChange("cors_configuration") {
			req.CorsConfiguration = expandApiGateway2CorsConfiguration(d.Get("cors_configuration").([]interface{}))
		}
		if d.HasChange("description") {
			req.Description = aws.String(d.Get("description").(string))
		}
		if d.HasChange("name") {
			req.Name = aws.String(d.Get("name").(string))
		}
		if d.HasChange("route_selection_expression") {
			req.RouteSelectionExpression = aws.String(d.Get("route_selection_expression").(string))
		}
		if d.HasChange("version") {
			req.Version = aws.String(d.Get("version").(string))
		}

		log.Printf("[DEBUG] Updating API Gateway v2 API: %s", req)
		_, err := conn.UpdateApi(req)
		if err != nil {
			return fmt.Errorf("error updating API Gateway v2 API (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Apigatewayv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating API Gateway v2 API (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsApiGatewayV2ApiRead(d, meta)
}

func resourceAwsApiGatewayV2ApiDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 API (%s)", d.Id())
	_, err := conn.DeleteApi(&apigatewayv2.DeleteApiInput{
		ApiId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 API (%s): %s", d.Id(), err)
	}

	return nil
}

func expandApiGateway2CorsConfiguration(vConfiguration []interface{}) *apigatewayv2.Cors {
	configuration := &apigatewayv2.Cors{}

	if len(vConfiguration) == 0 || vConfiguration[0] == nil {
		return configuration
	}
	mConfiguration := vConfiguration[0].(map[string]interface{})

	if vAllowCredentials, ok := mConfiguration["allow_credentials"].(bool); ok {
		configuration.AllowCredentials = aws.Bool(vAllowCredentials)
	}
	if vAllowHeaders, ok := mConfiguration["allow_headers"].(*schema.Set); ok {
		configuration.AllowHeaders = expandStringSet(vAllowHeaders)
	}
	if vAllowMethods, ok := mConfiguration["allow_methods"].(*schema.Set); ok {
		configuration.AllowMethods = expandStringSet(vAllowMethods)
	}
	if vAllowOrigins, ok := mConfiguration["allow_origins"].(*schema.Set); ok {
		configuration.AllowOrigins = expandStringSet(vAllowOrigins)
	}
	if vExposeHeaders, ok := mConfiguration["expose_headers"].(*schema.Set); ok {
		configuration.ExposeHeaders = expandStringSet(vExposeHeaders)
	}
	if vMaxAge, ok := mConfiguration["max_age"].(int); ok {
		configuration.MaxAge = aws.Int64(int64(vMaxAge))
	}

	return configuration
}

func flattenApiGateway2CorsConfiguration(configuration *apigatewayv2.Cors) []interface{} {
	if configuration == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"allow_credentials": aws.BoolValue(configuration.AllowCredentials),
		"allow_headers":     flattenCaseInsensitiveStringSet(configuration.AllowHeaders),
		"allow_methods":     flattenCaseInsensitiveStringSet(configuration.AllowMethods),
		"allow_origins":     flattenCaseInsensitiveStringSet(configuration.AllowOrigins),
		"expose_headers":    flattenCaseInsensitiveStringSet(configuration.ExposeHeaders),
		"max_age":           int(aws.Int64Value(configuration.MaxAge)),
	}}
}
