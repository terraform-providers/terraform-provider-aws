package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsCloudWatchEventConnection() *schema.Resource {
	connectionHttpParameters := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"body": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"is_value_secret": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"header": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"is_value_secret": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"query_string": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"is_value_secret": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}

	return &schema.Resource{
		Create: resourceAwsCloudWatchEventConnectionCreate,
		Read:   resourceAwsCloudWatchEventConnectionRead,
		Update: resourceAwsCloudWatchEventConnectionUpdate,
		Delete: resourceAwsCloudWatchEventConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+`), ""),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"authorization_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(events.ConnectionAuthorizationType_Values(), true),
			},
			"auth_parameters": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_key": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"auth_parameters.0.api_key",
								"auth_parameters.0.basic",
								"auth_parameters.0.oauth",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
									},
								},
							},
						},
						"basic": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"auth_parameters.0.api_key",
								"auth_parameters.0.basic",
								"auth_parameters.0.oauth",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"username": {
										Type:     schema.TypeString,
										Required: true,
									},
									"password": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
									},
								},
							},
						},
						"oauth": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ExactlyOneOf: []string{
								"auth_parameters.0.api_key",
								"auth_parameters.0.basic",
								"auth_parameters.0.oauth",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authorization_endpoint": {
										Type:     schema.TypeString,
										Required: true,
									},
									"http_method": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(events.ConnectionOAuthHttpMethod_Values(), true),
									},
									"oauth_http_parameters": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem:     connectionHttpParameters,
									},
									"client_parameters": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"client_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"client_secret": {
													Type:      schema.TypeString,
													Required:  true,
													Sensitive: true,
												},
											},
										},
									},
								},
							},
						},
						"invocation_http_parameters": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem:     connectionHttpParameters,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCloudWatchEventConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	name := d.Get("name").(string)
	input := &events.CreateConnectionInput{
		AuthorizationType: aws.String(d.Get("authorization_type").(string)),
		AuthParameters:    expandAwsCloudWatchEventCreateConnectionAuthRequestParameters(d.Get("auth_parameters").([]interface{})),
		Name:              aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating CloudWatch Events connection: %s", input)

	_, err := conn.CreateConnection(input)

	if err != nil {
		return fmt.Errorf("error creating CloudWatch Events connection (%s): %w", name, err)
	}

	d.SetId(name)

	_, err = waiter.ConnectionCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for CloudWatch Events connection (%s) to create: %w", d.Id(), err)
	}

	return resourceAwsCloudWatchEventConnectionRead(d, meta)
}

func resourceAwsCloudWatchEventConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	output, err := finder.ConnectionByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Events connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudWatch Events connection (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.ConnectionArn)
	d.Set("authorization_type", output.AuthorizationType)
	d.Set("description", output.Description)
	d.Set("name", output.Name)
	d.Set("secret_arn", output.SecretArn)

	if output.AuthParameters != nil {
		authParameters := flattenAwsCloudWatchEventConnectionAuthParameters(output.AuthParameters, d)
		if err := d.Set("auth_parameters", authParameters); err != nil {
			return fmt.Errorf("error setting auth_parameters error: %w", err)
		}
	}

	return nil
}

func resourceAwsCloudWatchEventConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	input := &events.UpdateConnectionInput{
		Name: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("authorization_type"); ok {
		input.AuthorizationType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("auth_parameters"); ok {
		input.AuthParameters = expandAwsCloudWatchEventUpdateConnectionAuthRequestParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating CloudWatch Events connection: %s", input)
	_, err := conn.UpdateConnection(input)

	if err != nil {
		return fmt.Errorf("error updating CloudWatch Events connection (%s): %w", d.Id(), err)
	}

	_, err = waiter.ConnectionUpdated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for CloudWatch Events connection (%s) to update: %w", d.Id(), err)
	}

	return resourceAwsCloudWatchEventConnectionRead(d, meta)
}

func resourceAwsCloudWatchEventConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	log.Printf("[INFO] Deleting CloudWatch Events connection (%s)", d.Id())
	_, err := conn.DeleteConnection(&events.DeleteConnectionInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, events.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudWatch Events connection (%s): %w", d.Id(), err)
	}

	_, err = waiter.ConnectionDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for CloudWatch Events connection (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func expandAwsCloudWatchEventCreateConnectionAuthRequestParameters(config []interface{}) *events.CreateConnectionAuthRequestParameters {
	authParameters := &events.CreateConnectionAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["api_key"]; ok {
			authParameters.ApiKeyAuthParameters = expandAwsCreateConnectionApiKeyAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["basic"]; ok {
			authParameters.BasicAuthParameters = expandAwsCreateConnectionBasicAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["oauth"]; ok {
			authParameters.OAuthParameters = expandAwsCreateConnectionOAuthAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["invocation_http_parameters"]; ok {
			authParameters.InvocationHttpParameters = expandAwsConnectionHttpParameters(val.([]interface{}))
		}
	}

	return authParameters
}

func expandAwsCreateConnectionApiKeyAuthRequestParameters(config []interface{}) *events.CreateConnectionApiKeyAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	apiKeyAuthParameters := &events.CreateConnectionApiKeyAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["key"].(string); ok && val != "" {
			apiKeyAuthParameters.ApiKeyName = aws.String(val)
		}
		if val, ok := param["value"].(string); ok && val != "" {
			apiKeyAuthParameters.ApiKeyValue = aws.String(val)
		}
	}
	return apiKeyAuthParameters
}

func expandAwsCreateConnectionBasicAuthRequestParameters(config []interface{}) *events.CreateConnectionBasicAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	basicAuthParameters := &events.CreateConnectionBasicAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["username"].(string); ok && val != "" {
			basicAuthParameters.Username = aws.String(val)
		}
		if val, ok := param["password"].(string); ok && val != "" {
			basicAuthParameters.Password = aws.String(val)
		}
	}
	return basicAuthParameters
}

func expandAwsCreateConnectionOAuthAuthRequestParameters(config []interface{}) *events.CreateConnectionOAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	oAuthParameters := &events.CreateConnectionOAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["authorization_endpoint"].(string); ok && val != "" {
			oAuthParameters.AuthorizationEndpoint = aws.String(val)
		}
		if val, ok := param["http_method"].(string); ok && val != "" {
			oAuthParameters.HttpMethod = aws.String(val)
		}
		if val, ok := param["oauth_http_parameters"]; ok {
			oAuthParameters.OAuthHttpParameters = expandAwsConnectionHttpParameters(val.([]interface{}))
		}
		if val, ok := param["client_parameters"]; ok {
			oAuthParameters.ClientParameters = expandAwsCreateConnectionOAuthClientRequestParameters(val.([]interface{}))
		}
	}
	return oAuthParameters
}

func expandAwsCreateConnectionOAuthClientRequestParameters(config []interface{}) *events.CreateConnectionOAuthClientRequestParameters {
	oAuthClientRequestParameters := &events.CreateConnectionOAuthClientRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["client_id"].(string); ok && val != "" {
			oAuthClientRequestParameters.ClientID = aws.String(val)
		}
		if val, ok := param["client_secret"].(string); ok && val != "" {
			oAuthClientRequestParameters.ClientSecret = aws.String(val)
		}
	}
	return oAuthClientRequestParameters
}

func expandAwsConnectionHttpParameters(config []interface{}) *events.ConnectionHttpParameters {
	if len(config) == 0 {
		return nil
	}
	httpParameters := &events.ConnectionHttpParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["body"]; ok {
			httpParameters.BodyParameters = expandAwsConnectionHttpParametersBody(val.([]interface{}))
		}
		if val, ok := param["header"]; ok {
			httpParameters.HeaderParameters = expandAwsConnectionHttpParametersHeader(val.([]interface{}))
		}
		if val, ok := param["query_string"]; ok {
			httpParameters.QueryStringParameters = expandAwsConnectionHttpParametersQueryString(val.([]interface{}))
		}
	}
	return httpParameters
}

func expandAwsConnectionHttpParametersBody(config []interface{}) []*events.ConnectionBodyParameter {
	if len(config) == 0 {
		return nil
	}
	var parameters []*events.ConnectionBodyParameter
	for _, c := range config {
		parameter := events.ConnectionBodyParameter{}

		input := c.(map[string]interface{})
		if val, ok := input["key"].(string); ok && val != "" {
			parameter.Key = aws.String(val)
		}
		if val, ok := input["value"].(string); ok && val != "" {
			parameter.Value = aws.String(val)
		}
		if val, ok := input["is_value_secret"].(bool); ok {
			parameter.IsValueSecret = aws.Bool(val)
		}
		parameters = append(parameters, &parameter)
	}
	return parameters
}

func expandAwsConnectionHttpParametersHeader(config []interface{}) []*events.ConnectionHeaderParameter {
	if len(config) == 0 {
		return nil
	}
	var parameters []*events.ConnectionHeaderParameter
	for _, c := range config {
		parameter := events.ConnectionHeaderParameter{}

		input := c.(map[string]interface{})
		if val, ok := input["key"].(string); ok && val != "" {
			parameter.Key = aws.String(val)
		}
		if val, ok := input["value"].(string); ok && val != "" {
			parameter.Value = aws.String(val)
		}
		if val, ok := input["is_value_secret"].(bool); ok {
			parameter.IsValueSecret = aws.Bool(val)
		}
		parameters = append(parameters, &parameter)
	}
	return parameters
}

func expandAwsConnectionHttpParametersQueryString(config []interface{}) []*events.ConnectionQueryStringParameter {
	if len(config) == 0 {
		return nil
	}
	var parameters []*events.ConnectionQueryStringParameter
	for _, c := range config {
		parameter := events.ConnectionQueryStringParameter{}

		input := c.(map[string]interface{})
		if val, ok := input["key"].(string); ok && val != "" {
			parameter.Key = aws.String(val)
		}
		if val, ok := input["value"].(string); ok && val != "" {
			parameter.Value = aws.String(val)
		}
		if val, ok := input["is_value_secret"].(bool); ok {
			parameter.IsValueSecret = aws.Bool(val)
		}
		parameters = append(parameters, &parameter)
	}
	return parameters
}

func flattenAwsCloudWatchEventConnectionAuthParameters(
	authParameters *events.ConnectionAuthResponseParameters,
	resourceData *schema.ResourceData,
) []map[string]interface{} {
	config := make(map[string]interface{})

	if authParameters.ApiKeyAuthParameters != nil {
		config["api_key"] = flattenAwsConnectionApiKeyAuthParameters(authParameters.ApiKeyAuthParameters, resourceData)
	}

	if authParameters.BasicAuthParameters != nil {
		config["basic"] = flattenAwsConnectionBasicAuthParameters(authParameters.BasicAuthParameters, resourceData)
	}

	if authParameters.OAuthParameters != nil {
		config["oauth"] = flattenAwsConnectionOAuthParameters(authParameters.OAuthParameters, resourceData)
	}

	if authParameters.InvocationHttpParameters != nil {
		config["invocation_http_parameters"] = flattenAwsConnectionHttpParameters(authParameters.InvocationHttpParameters, resourceData, "auth_parameters.0.invocation_http_parameters")
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenAwsConnectionApiKeyAuthParameters(apiKeyAuthParameters *events.ConnectionApiKeyAuthResponseParameters, resourceData *schema.ResourceData) []map[string]interface{} {
	if apiKeyAuthParameters == nil {
		return nil
	}

	config := make(map[string]interface{})
	if apiKeyAuthParameters.ApiKeyName != nil {
		config["key"] = aws.StringValue(apiKeyAuthParameters.ApiKeyName)
	}

	if v, ok := resourceData.GetOk("auth_parameters.0.api_key.0.value"); ok {
		config["value"] = v.(string)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenAwsConnectionBasicAuthParameters(basicAuthParameters *events.ConnectionBasicAuthResponseParameters, resourceData *schema.ResourceData) []map[string]interface{} {
	if basicAuthParameters == nil {
		return nil
	}

	config := make(map[string]interface{})
	if basicAuthParameters.Username != nil {
		config["username"] = aws.StringValue(basicAuthParameters.Username)
	}

	if v, ok := resourceData.GetOk("auth_parameters.0.basic.0.password"); ok {
		config["password"] = v.(string)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenAwsConnectionOAuthParameters(oAuthParameters *events.ConnectionOAuthResponseParameters, resourceData *schema.ResourceData) []map[string]interface{} {
	if oAuthParameters == nil {
		return nil
	}

	config := make(map[string]interface{})
	if oAuthParameters.AuthorizationEndpoint != nil {
		config["authorization_endpoint"] = aws.StringValue(oAuthParameters.AuthorizationEndpoint)
	}
	if oAuthParameters.HttpMethod != nil {
		config["http_method"] = aws.StringValue(oAuthParameters.HttpMethod)
	}
	config["oauth_http_parameters"] = flattenAwsConnectionHttpParameters(oAuthParameters.OAuthHttpParameters, resourceData, "auth_parameters.0.oauth.0.oauth_http_parameters")
	config["client_parameters"] = flattenAwsConnectionOAuthClientResponseParameters(oAuthParameters.ClientParameters, resourceData)

	result := []map[string]interface{}{config}
	return result
}

func flattenAwsConnectionOAuthClientResponseParameters(oAuthClientRequestParameters *events.ConnectionOAuthClientResponseParameters, resourceData *schema.ResourceData) []map[string]interface{} {
	if oAuthClientRequestParameters == nil {
		return nil
	}

	config := make(map[string]interface{})
	if oAuthClientRequestParameters.ClientID != nil {
		config["client_id"] = aws.StringValue(oAuthClientRequestParameters.ClientID)
	}

	if v, ok := resourceData.GetOk("auth_parameters.0.oauth.0.client_parameters.0.client_secret"); ok {
		config["client_secret"] = v.(string)
	}

	result := []map[string]interface{}{config}
	return result
}

func flattenAwsConnectionHttpParameters(
	httpParameters *events.ConnectionHttpParameters,
	resourceData *schema.ResourceData,
	path string,
) []map[string]interface{} {
	if httpParameters == nil {
		return nil
	}

	var bodyParameters []map[string]interface{}
	for i, param := range httpParameters.BodyParameters {
		config := make(map[string]interface{})
		config["is_value_secret"] = aws.BoolValue(param.IsValueSecret)
		config["key"] = aws.StringValue(param.Key)

		if param.Value != nil {
			config["value"] = aws.StringValue(param.Value)
		} else if v, ok := resourceData.GetOk(fmt.Sprintf("%s.0.body.%d.value", path, i)); ok {
			config["value"] = v.(string)
		}
		bodyParameters = append(bodyParameters, config)
	}

	var headerParameters []map[string]interface{}
	for i, param := range httpParameters.HeaderParameters {
		config := make(map[string]interface{})
		config["is_value_secret"] = aws.BoolValue(param.IsValueSecret)
		config["key"] = aws.StringValue(param.Key)

		if param.Value != nil {
			config["value"] = aws.StringValue(param.Value)
		} else if v, ok := resourceData.GetOk(fmt.Sprintf("%s.0.header.%d.value", path, i)); ok {
			config["value"] = v.(string)
		}
		headerParameters = append(headerParameters, config)
	}

	var queryStringParameters []map[string]interface{}
	for i, param := range httpParameters.QueryStringParameters {
		config := make(map[string]interface{})
		config["is_value_secret"] = aws.BoolValue(param.IsValueSecret)
		config["key"] = aws.StringValue(param.Key)

		if param.Value != nil {
			config["value"] = aws.StringValue(param.Value)
		} else if v, ok := resourceData.GetOk(fmt.Sprintf("%s.0.query_string.%d.value", path, i)); ok {
			config["value"] = v.(string)
		}
		queryStringParameters = append(queryStringParameters, config)
	}

	parameters := make(map[string]interface{})
	parameters["body"] = bodyParameters
	parameters["header"] = headerParameters
	parameters["query_string"] = queryStringParameters

	result := []map[string]interface{}{parameters}
	return result
}

func expandAwsCloudWatchEventUpdateConnectionAuthRequestParameters(config []interface{}) *events.UpdateConnectionAuthRequestParameters {
	authParameters := &events.UpdateConnectionAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["api_key"]; ok {
			authParameters.ApiKeyAuthParameters = expandAwsUpdateConnectionApiKeyAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["basic"]; ok {
			authParameters.BasicAuthParameters = expandAwsUpdateConnectionBasicAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["oauth"]; ok {
			authParameters.OAuthParameters = expandAwsUpdateConnectionOAuthAuthRequestParameters(val.([]interface{}))
		}
		if val, ok := param["invocation_http_parameters"]; ok {
			authParameters.InvocationHttpParameters = expandAwsConnectionHttpParameters(val.([]interface{}))
		}
	}

	return authParameters
}

func expandAwsUpdateConnectionApiKeyAuthRequestParameters(config []interface{}) *events.UpdateConnectionApiKeyAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	apiKeyAuthParameters := &events.UpdateConnectionApiKeyAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["key"].(string); ok && val != "" {
			apiKeyAuthParameters.ApiKeyName = aws.String(val)
		}
		if val, ok := param["value"].(string); ok && val != "" {
			apiKeyAuthParameters.ApiKeyValue = aws.String(val)
		}
	}
	return apiKeyAuthParameters
}

func expandAwsUpdateConnectionBasicAuthRequestParameters(config []interface{}) *events.UpdateConnectionBasicAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	basicAuthParameters := &events.UpdateConnectionBasicAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["username"].(string); ok && val != "" {
			basicAuthParameters.Username = aws.String(val)
		}
		if val, ok := param["password"].(string); ok && val != "" {
			basicAuthParameters.Password = aws.String(val)
		}
	}
	return basicAuthParameters
}

func expandAwsUpdateConnectionOAuthAuthRequestParameters(config []interface{}) *events.UpdateConnectionOAuthRequestParameters {
	if len(config) == 0 {
		return nil
	}
	oAuthParameters := &events.UpdateConnectionOAuthRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["authorization_endpoint"].(string); ok && val != "" {
			oAuthParameters.AuthorizationEndpoint = aws.String(val)
		}
		if val, ok := param["http_method"].(string); ok && val != "" {
			oAuthParameters.HttpMethod = aws.String(val)
		}
		if val, ok := param["oauth_http_parameters"]; ok {
			oAuthParameters.OAuthHttpParameters = expandAwsConnectionHttpParameters(val.([]interface{}))
		}
		if val, ok := param["client_parameters"]; ok {
			oAuthParameters.ClientParameters = expandAwsUpdateConnectionOAuthClientRequestParameters(val.([]interface{}))
		}
	}
	return oAuthParameters
}

func expandAwsUpdateConnectionOAuthClientRequestParameters(config []interface{}) *events.UpdateConnectionOAuthClientRequestParameters {
	oAuthClientRequestParameters := &events.UpdateConnectionOAuthClientRequestParameters{}
	for _, c := range config {
		param := c.(map[string]interface{})
		if val, ok := param["client_id"].(string); ok && val != "" {
			oAuthClientRequestParameters.ClientID = aws.String(val)
		}
		if val, ok := param["client_secret"].(string); ok && val != "" {
			oAuthClientRequestParameters.ClientSecret = aws.String(val)
		}
	}
	return oAuthClientRequestParameters
}
