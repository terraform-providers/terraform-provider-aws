package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsApiGatewayV2RouteResponse() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2RouteResponseCreate,
		Read:   resourceAwsApiGatewayV2RouteResponseRead,
		Update: resourceAwsApiGatewayV2RouteResponseUpdate,
		Delete: resourceAwsApiGatewayV2RouteResponseDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsApiGatewayV2RouteResponseImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"model_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"response_models": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"route_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route_response_key": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsApiGatewayV2RouteResponseCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateRouteResponseInput{
		ApiId:            aws.String(d.Get("api_id").(string)),
		RouteId:          aws.String(d.Get("route_id").(string)),
		RouteResponseKey: aws.String(d.Get("route_response_key").(string)),
	}
	if v, ok := d.GetOk("model_selection_expression"); ok {
		req.ModelSelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("response_models"); ok {
		req.ResponseModels = stringMapToPointers(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 route response: %s", req)
	resp, err := conn.CreateRouteResponse(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 route response: %s", err)
	}

	d.SetId(aws.StringValue(resp.RouteResponseId))

	return resourceAwsApiGatewayV2RouteResponseRead(d, meta)
}

func resourceAwsApiGatewayV2RouteResponseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetRouteResponse(&apigatewayv2.GetRouteResponseInput{
		ApiId:           aws.String(d.Get("api_id").(string)),
		RouteId:         aws.String(d.Get("route_id").(string)),
		RouteResponseId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 route response (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 route response: %s", err)
	}

	d.Set("model_selection_expression", resp.ModelSelectionExpression)
	if err := d.Set("response_models", pointersMapToStringList(resp.ResponseModels)); err != nil {
		return fmt.Errorf("error setting response_models: %s", err)
	}
	d.Set("route_response_key", resp.RouteResponseKey)

	return nil
}

func resourceAwsApiGatewayV2RouteResponseUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.UpdateRouteResponseInput{
		ApiId:           aws.String(d.Get("api_id").(string)),
		RouteId:         aws.String(d.Get("route_id").(string)),
		RouteResponseId: aws.String(d.Id()),
	}
	if d.HasChange("model_selection_expression") {
		req.ModelSelectionExpression = aws.String(d.Get("model_selection_expression").(string))
	}
	if d.HasChange("response_models") {
		req.ResponseModels = stringMapToPointers(d.Get("response_models").(map[string]interface{}))
	}
	if d.HasChange("route_response_key") {
		req.RouteResponseKey = aws.String(d.Get("route_response_key").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 route response: %s", req)
	_, err := conn.UpdateRouteResponse(req)
	if err != nil {
		return fmt.Errorf("error updating API Gateway v2 route response: %s", err)
	}

	return resourceAwsApiGatewayV2RouteResponseRead(d, meta)
}

func resourceAwsApiGatewayV2RouteResponseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 route response (%s)", d.Id())
	_, err := conn.DeleteRouteResponse(&apigatewayv2.DeleteRouteResponseInput{
		ApiId:           aws.String(d.Get("api_id").(string)),
		RouteId:         aws.String(d.Get("route_id").(string)),
		RouteResponseId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 route response: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayV2RouteResponseImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/route-id/route-response-id'", d.Id())
	}

	d.SetId(parts[2])
	d.Set("api_id", parts[0])
	d.Set("route_id", parts[1])

	return []*schema.ResourceData{d}, nil
}
