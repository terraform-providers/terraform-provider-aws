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

func resourceAwsApiGatewayV2Route() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2RouteCreate,
		Read:   resourceAwsApiGatewayV2RouteRead,
		Update: resourceAwsApiGatewayV2RouteUpdate,
		Delete: resourceAwsApiGatewayV2RouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsApiGatewayV2RouteImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"api_key_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"authorization_scopes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"authorization_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  apigatewayv2.AuthorizationTypeNone,
				ValidateFunc: validation.StringInSlice([]string{
					apigatewayv2.AuthorizationTypeNone,
					apigatewayv2.AuthorizationTypeAwsIam,
					apigatewayv2.AuthorizationTypeCustom,
					apigatewayv2.AuthorizationTypeJwt,
				}, false),
			},
			"authorizer_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"model_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"operation_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"request_models": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"route_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"route_response_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"target": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceAwsApiGatewayV2RouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateRouteInput{
		ApiId:             aws.String(d.Get("api_id").(string)),
		ApiKeyRequired:    aws.Bool(d.Get("api_key_required").(bool)),
		AuthorizationType: aws.String(d.Get("authorization_type").(string)),
		RouteKey:          aws.String(d.Get("route_key").(string)),
	}
	if v, ok := d.GetOk("authorization_scopes"); ok {
		req.AuthorizationScopes = expandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("authorizer_id"); ok {
		req.AuthorizerId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("model_selection_expression"); ok {
		req.ModelSelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("operation_name"); ok {
		req.OperationName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("request_models"); ok {
		req.RequestModels = stringMapToPointers(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("route_response_selection_expression"); ok {
		req.RouteResponseSelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target"); ok {
		req.Target = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 route: %s", req)
	resp, err := conn.CreateRoute(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 route: %s", err)
	}

	d.SetId(aws.StringValue(resp.RouteId))

	return resourceAwsApiGatewayV2RouteRead(d, meta)
}

func resourceAwsApiGatewayV2RouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetRoute(&apigatewayv2.GetRouteInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		RouteId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 route: %s", err)
	}

	d.Set("api_key_required", resp.ApiKeyRequired)
	if err := d.Set("authorization_scopes", flattenStringSet(resp.AuthorizationScopes)); err != nil {
		return fmt.Errorf("error setting authorization_scopes: %s", err)
	}
	d.Set("authorization_type", resp.AuthorizationType)
	d.Set("authorizer_id", resp.AuthorizerId)
	d.Set("model_selection_expression", resp.ModelSelectionExpression)
	d.Set("operation_name", resp.OperationName)
	if err := d.Set("request_models", pointersMapToStringList(resp.RequestModels)); err != nil {
		return fmt.Errorf("error setting request_models: %s", err)
	}
	d.Set("route_key", resp.RouteKey)
	d.Set("route_response_selection_expression", resp.RouteResponseSelectionExpression)
	d.Set("target", resp.Target)

	return nil
}

func resourceAwsApiGatewayV2RouteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.UpdateRouteInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		RouteId: aws.String(d.Id()),
	}
	if d.HasChange("api_key_required") {
		req.ApiKeyRequired = aws.Bool(d.Get("api_key_required").(bool))
	}
	if d.HasChange("authorization_scopes") {
		req.AuthorizationScopes = expandStringSet(d.Get("authorization_scopes").(*schema.Set))
	}
	if d.HasChange("authorization_type") {
		req.AuthorizationType = aws.String(d.Get("authorization_type").(string))
	}
	if d.HasChange("authorizer_id") {
		req.AuthorizerId = aws.String(d.Get("authorizer_id").(string))
	}
	if d.HasChange("model_selection_expression") {
		req.ModelSelectionExpression = aws.String(d.Get("model_selection_expression").(string))
	}
	if d.HasChange("operation_name") {
		req.OperationName = aws.String(d.Get("operation_name").(string))
	}
	if d.HasChange("request_models") {
		req.RequestModels = stringMapToPointers(d.Get("request_models").(map[string]interface{}))
	}
	if d.HasChange("route_response_selection_expression") {
		req.RouteResponseSelectionExpression = aws.String(d.Get("route_response_selection_expression").(string))
	}
	if d.HasChange("target") {
		req.Target = aws.String(d.Get("target").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 route: %s", req)
	_, err := conn.UpdateRoute(req)
	if err != nil {
		return fmt.Errorf("error updating API Gateway v2 route: %s", err)
	}

	return resourceAwsApiGatewayV2RouteRead(d, meta)
}

func resourceAwsApiGatewayV2RouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 route (%s)", d.Id())
	_, err := conn.DeleteRoute(&apigatewayv2.DeleteRouteInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		RouteId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 route: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayV2RouteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/route-id'", d.Id())
	}

	apiId := parts[0]
	routeId := parts[1]

	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetRoute(&apigatewayv2.GetRouteInput{
		ApiId:   aws.String(apiId),
		RouteId: aws.String(routeId),
	})
	if err != nil {
		return nil, err
	}

	if aws.BoolValue(resp.ApiGatewayManaged) {
		return nil, fmt.Errorf("API Gateway v2 route (%s) was created via quick create", routeId)
	}

	d.SetId(routeId)
	d.Set("api_id", apiId)

	return []*schema.ResourceData{d}, nil
}
