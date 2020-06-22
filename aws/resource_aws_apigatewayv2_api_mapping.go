package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsApiGatewayV2ApiMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2ApiMappingCreate,
		Read:   resourceAwsApiGatewayV2ApiMappingRead,
		Update: resourceAwsApiGatewayV2ApiMappingUpdate,
		Delete: resourceAwsApiGatewayV2ApiMappingDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsApiGatewayV2ApiMappingImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"api_mapping_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stage": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsApiGatewayV2ApiMappingCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateApiMappingInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
		Stage:      aws.String(d.Get("stage").(string)),
	}
	if v, ok := d.GetOk("api_mapping_key"); ok {
		req.ApiMappingKey = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 API mapping: %s", req)
	resp, err := conn.CreateApiMapping(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 API mapping: %s", err)
	}

	d.SetId(aws.StringValue(resp.ApiMappingId))

	return resourceAwsApiGatewayV2ApiMappingRead(d, meta)
}

func resourceAwsApiGatewayV2ApiMappingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetApiMapping(&apigatewayv2.GetApiMappingInput{
		ApiMappingId: aws.String(d.Id()),
		DomainName:   aws.String(d.Get("domain_name").(string)),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 API mapping (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 API mapping: %s", err)
	}

	d.Set("api_id", resp.ApiId)
	d.Set("api_mapping_key", resp.ApiMappingKey)
	d.Set("stage", resp.Stage)

	return nil
}

func resourceAwsApiGatewayV2ApiMappingUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.UpdateApiMappingInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		ApiMappingId: aws.String(d.Id()),
		DomainName:   aws.String(d.Get("domain_name").(string)),
	}
	if d.HasChange("api_mapping_key") {
		req.ApiMappingKey = aws.String(d.Get("api_mapping_key").(string))
	}
	if d.HasChange("stage") {
		req.Stage = aws.String(d.Get("stage").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 API mapping: %s", req)
	_, err := conn.UpdateApiMapping(req)
	if err != nil {
		return fmt.Errorf("error updating API Gateway v2 API mapping: %s", err)
	}

	return resourceAwsApiGatewayV2ApiMappingRead(d, meta)
}

func resourceAwsApiGatewayV2ApiMappingDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 API mapping (%s)", d.Id())
	_, err := conn.DeleteApiMapping(&apigatewayv2.DeleteApiMappingInput{
		ApiMappingId: aws.String(d.Id()),
		DomainName:   aws.String(d.Get("domain_name").(string)),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 API mapping: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayV2ApiMappingImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-mapping-id/domain-name'", d.Id())
	}

	d.SetId(parts[0])
	d.Set("domain_name", parts[1])

	return []*schema.ResourceData{d}, nil
}
