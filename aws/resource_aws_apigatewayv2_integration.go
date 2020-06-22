package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsApiGatewayV2Integration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2IntegrationCreate,
		Read:   resourceAwsApiGatewayV2IntegrationRead,
		Update: resourceAwsApiGatewayV2IntegrationUpdate,
		Delete: resourceAwsApiGatewayV2IntegrationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsApiGatewayV2IntegrationImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"connection_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"connection_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  apigatewayv2.ConnectionTypeInternet,
				ValidateFunc: validation.StringInSlice([]string{
					apigatewayv2.ConnectionTypeInternet,
					apigatewayv2.ConnectionTypeVpcLink,
				}, false),
			},
			"content_handling_strategy": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					apigatewayv2.ContentHandlingStrategyConvertToBinary,
					apigatewayv2.ContentHandlingStrategyConvertToText,
				}, false),
			},
			"credentials_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"integration_method": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateHTTPMethod(),
			},
			"integration_response_selection_expression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"integration_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					apigatewayv2.IntegrationTypeAws,
					apigatewayv2.IntegrationTypeAwsProxy,
					apigatewayv2.IntegrationTypeHttp,
					apigatewayv2.IntegrationTypeHttpProxy,
					apigatewayv2.IntegrationTypeMock,
				}, false),
			},
			"integration_uri": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"passthrough_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  apigatewayv2.PassthroughBehaviorWhenNoMatch,
				ValidateFunc: validation.StringInSlice([]string{
					apigatewayv2.PassthroughBehaviorWhenNoMatch,
					apigatewayv2.PassthroughBehaviorNever,
					apigatewayv2.PassthroughBehaviorWhenNoTemplates,
				}, false),
			},
			"payload_format_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.0",
				ValidateFunc: validation.StringInSlice([]string{
					"1.0",
					"2.0",
				}, false),
			},
			"request_templates": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"template_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"timeout_milliseconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      29000,
				ValidateFunc: validation.IntBetween(50, 29000),
			},
		},
	}
}

func resourceAwsApiGatewayV2IntegrationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateIntegrationInput{
		ApiId:           aws.String(d.Get("api_id").(string)),
		IntegrationType: aws.String(d.Get("integration_type").(string)),
	}
	if v, ok := d.GetOk("connection_id"); ok {
		req.ConnectionId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("connection_type"); ok {
		req.ConnectionType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("content_handling_strategy"); ok {
		req.ContentHandlingStrategy = aws.String(v.(string))
	}
	if v, ok := d.GetOk("credentials_arn"); ok {
		req.CredentialsArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("integration_method"); ok {
		req.IntegrationMethod = aws.String(v.(string))
	}
	if v, ok := d.GetOk("integration_uri"); ok {
		req.IntegrationUri = aws.String(v.(string))
	}
	if v, ok := d.GetOk("passthrough_behavior"); ok {
		req.PassthroughBehavior = aws.String(v.(string))
	}
	if v, ok := d.GetOk("payload_format_version"); ok {
		req.PayloadFormatVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("request_templates"); ok {
		req.RequestTemplates = stringMapToPointers(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("template_selection_expression"); ok {
		req.TemplateSelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("timeout_milliseconds"); ok {
		req.TimeoutInMillis = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 integration: %s", req)
	resp, err := conn.CreateIntegration(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 integration: %s", err)
	}

	d.SetId(aws.StringValue(resp.IntegrationId))

	return resourceAwsApiGatewayV2IntegrationRead(d, meta)
}

func resourceAwsApiGatewayV2IntegrationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetIntegration(&apigatewayv2.GetIntegrationInput{
		ApiId:         aws.String(d.Get("api_id").(string)),
		IntegrationId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 integration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 integration: %s", err)
	}

	d.Set("connection_id", resp.ConnectionId)
	d.Set("connection_type", resp.ConnectionType)
	d.Set("content_handling_strategy", resp.ContentHandlingStrategy)
	d.Set("credentials_arn", resp.CredentialsArn)
	d.Set("description", resp.Description)
	d.Set("integration_method", resp.IntegrationMethod)
	d.Set("integration_response_selection_expression", resp.IntegrationResponseSelectionExpression)
	d.Set("integration_type", resp.IntegrationType)
	d.Set("integration_uri", resp.IntegrationUri)
	d.Set("passthrough_behavior", resp.PassthroughBehavior)
	d.Set("payload_format_version", resp.PayloadFormatVersion)
	err = d.Set("request_templates", pointersMapToStringList(resp.RequestTemplates))
	if err != nil {
		return fmt.Errorf("error setting request_templates: %s", err)
	}
	d.Set("template_selection_expression", resp.TemplateSelectionExpression)
	d.Set("timeout_milliseconds", resp.TimeoutInMillis)

	return nil
}

func resourceAwsApiGatewayV2IntegrationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.UpdateIntegrationInput{
		ApiId:         aws.String(d.Get("api_id").(string)),
		IntegrationId: aws.String(d.Id()),
	}
	if d.HasChange("connection_id") {
		req.ConnectionId = aws.String(d.Get("connection_id").(string))
	}
	if d.HasChange("connection_type") {
		req.ConnectionType = aws.String(d.Get("connection_type").(string))
	}
	if d.HasChange("content_handling_strategy") {
		req.ContentHandlingStrategy = aws.String(d.Get("content_handling_strategy").(string))
	}
	if d.HasChange("credentials_arn") {
		req.CredentialsArn = aws.String(d.Get("credentials_arn").(string))
	}
	if d.HasChange("description") {
		req.Description = aws.String(d.Get("description").(string))
	}
	if d.HasChange("integration_method") {
		req.IntegrationMethod = aws.String(d.Get("integration_method").(string))
	}
	if d.HasChange("integration_uri") {
		req.IntegrationUri = aws.String(d.Get("integration_uri").(string))
	}
	if d.HasChange("passthrough_behavior") {
		req.PassthroughBehavior = aws.String(d.Get("passthrough_behavior").(string))
	}
	if d.HasChange("payload_format_version") {
		req.PayloadFormatVersion = aws.String(d.Get("payload_format_version").(string))
	}
	if d.HasChange("request_templates") {
		req.RequestTemplates = stringMapToPointers(d.Get("request_templates").(map[string]interface{}))
	}
	if d.HasChange("template_selection_expression") {
		req.TemplateSelectionExpression = aws.String(d.Get("template_selection_expression").(string))
	}
	if d.HasChange("timeout_milliseconds") {
		req.TimeoutInMillis = aws.Int64(int64(d.Get("timeout_milliseconds").(int)))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 integration: %s", req)
	_, err := conn.UpdateIntegration(req)
	if err != nil {
		return fmt.Errorf("error updating API Gateway v2 integration: %s", err)
	}

	return resourceAwsApiGatewayV2IntegrationRead(d, meta)
}

func resourceAwsApiGatewayV2IntegrationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 integration (%s)", d.Id())
	_, err := conn.DeleteIntegration(&apigatewayv2.DeleteIntegrationInput{
		ApiId:         aws.String(d.Get("api_id").(string)),
		IntegrationId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 integration: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayV2IntegrationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/integration-id'", d.Id())
	}

	apiId := parts[0]
	integrationId := parts[1]

	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetIntegration(&apigatewayv2.GetIntegrationInput{
		ApiId:         aws.String(apiId),
		IntegrationId: aws.String(integrationId),
	})
	if err != nil {
		return nil, err
	}

	if aws.BoolValue(resp.ApiGatewayManaged) {
		return nil, fmt.Errorf("API Gateway v2 integration (%s) was created via quick create", integrationId)
	}

	d.SetId(integrationId)
	d.Set("api_id", apiId)

	return []*schema.ResourceData{d}, nil
}
