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

func resourceAwsApiGatewayV2IntegrationResponse() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2IntegrationResponseCreate,
		Read:   resourceAwsApiGatewayV2IntegrationResponseRead,
		Update: resourceAwsApiGatewayV2IntegrationResponseUpdate,
		Delete: resourceAwsApiGatewayV2IntegrationResponseDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsApiGatewayV2IntegrationResponseImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content_handling_strategy": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					apigatewayv2.ContentHandlingStrategyConvertToBinary,
					apigatewayv2.ContentHandlingStrategyConvertToText,
				}, false),
			},
			"integration_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"integration_response_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"response_templates": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"template_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsApiGatewayV2IntegrationResponseCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateIntegrationResponseInput{
		ApiId:                  aws.String(d.Get("api_id").(string)),
		IntegrationId:          aws.String(d.Get("integration_id").(string)),
		IntegrationResponseKey: aws.String(d.Get("integration_response_key").(string)),
	}
	if v, ok := d.GetOk("content_handling_strategy"); ok {
		req.ContentHandlingStrategy = aws.String(v.(string))
	}
	if v, ok := d.GetOk("response_templates"); ok {
		req.ResponseTemplates = stringMapToPointers(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("template_selection_expression"); ok {
		req.TemplateSelectionExpression = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 integration response: %s", req)
	resp, err := conn.CreateIntegrationResponse(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 integration response: %s", err)
	}

	d.SetId(aws.StringValue(resp.IntegrationResponseId))

	return resourceAwsApiGatewayV2IntegrationResponseRead(d, meta)
}

func resourceAwsApiGatewayV2IntegrationResponseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetIntegrationResponse(&apigatewayv2.GetIntegrationResponseInput{
		ApiId:                 aws.String(d.Get("api_id").(string)),
		IntegrationId:         aws.String(d.Get("integration_id").(string)),
		IntegrationResponseId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 integration response (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 integration response: %s", err)
	}

	d.Set("content_handling_strategy", resp.ContentHandlingStrategy)
	d.Set("integration_response_key", resp.IntegrationResponseKey)
	err = d.Set("response_templates", pointersMapToStringList(resp.ResponseTemplates))
	if err != nil {
		return fmt.Errorf("error setting response_templates: %s", err)
	}
	d.Set("template_selection_expression", resp.TemplateSelectionExpression)

	return nil
}

func resourceAwsApiGatewayV2IntegrationResponseUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.UpdateIntegrationResponseInput{
		ApiId:                 aws.String(d.Get("api_id").(string)),
		IntegrationId:         aws.String(d.Get("integration_id").(string)),
		IntegrationResponseId: aws.String(d.Id()),
	}
	if d.HasChange("content_handling_strategy") {
		req.ContentHandlingStrategy = aws.String(d.Get("content_handling_strategy").(string))
	}
	if d.HasChange("integration_response_key") {
		req.IntegrationResponseKey = aws.String(d.Get("integration_response_key").(string))
	}
	if d.HasChange("response_templates") {
		req.ResponseTemplates = stringMapToPointers(d.Get("response_templates").(map[string]interface{}))
	}
	if d.HasChange("template_selection_expression") {
		req.TemplateSelectionExpression = aws.String(d.Get("template_selection_expression").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 integration response: %s", req)
	_, err := conn.UpdateIntegrationResponse(req)
	if err != nil {
		return fmt.Errorf("error updating API Gateway v2 integration response: %s", err)
	}

	return resourceAwsApiGatewayV2IntegrationResponseRead(d, meta)
}

func resourceAwsApiGatewayV2IntegrationResponseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 integration response (%s)", d.Id())
	_, err := conn.DeleteIntegrationResponse(&apigatewayv2.DeleteIntegrationResponseInput{
		ApiId:                 aws.String(d.Get("api_id").(string)),
		IntegrationId:         aws.String(d.Get("integration_id").(string)),
		IntegrationResponseId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 integration response: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayV2IntegrationResponseImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/integration-id/integration-response-id'", d.Id())
	}

	d.SetId(parts[2])
	d.Set("api_id", parts[0])
	d.Set("integration_id", parts[1])

	return []*schema.ResourceData{d}, nil
}
