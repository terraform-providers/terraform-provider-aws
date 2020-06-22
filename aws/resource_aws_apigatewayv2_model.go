package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsApiGatewayV2Model() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2ModelCreate,
		Read:   resourceAwsApiGatewayV2ModelRead,
		Update: resourceAwsApiGatewayV2ModelUpdate,
		Delete: resourceAwsApiGatewayV2ModelDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsApiGatewayV2ModelImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), "must be alphanumeric"),
				),
			},
			"schema": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 32768),
					validation.StringIsJSON,
				),
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceAwsApiGatewayV2ModelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateModelInput{
		ApiId:       aws.String(d.Get("api_id").(string)),
		ContentType: aws.String(d.Get("content_type").(string)),
		Name:        aws.String(d.Get("name").(string)),
		Schema:      aws.String(d.Get("schema").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 model: %s", req)
	resp, err := conn.CreateModel(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 model: %s", err)
	}

	d.SetId(aws.StringValue(resp.ModelId))

	return resourceAwsApiGatewayV2ModelRead(d, meta)
}

func resourceAwsApiGatewayV2ModelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	resp, err := conn.GetModel(&apigatewayv2.GetModelInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		ModelId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 model (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 model: %s", err)
	}

	d.Set("content_type", resp.ContentType)
	d.Set("description", resp.Description)
	d.Set("name", resp.Name)
	d.Set("schema", resp.Schema)

	return nil
}

func resourceAwsApiGatewayV2ModelUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.UpdateModelInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		ModelId: aws.String(d.Id()),
	}
	if d.HasChange("content_type") {
		req.ContentType = aws.String(d.Get("content_type").(string))
	}
	if d.HasChange("description") {
		req.Description = aws.String(d.Get("description").(string))
	}
	if d.HasChange("name") {
		req.Name = aws.String(d.Get("name").(string))
	}
	if d.HasChange("schema") {
		req.Schema = aws.String(d.Get("schema").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 model: %s", req)
	_, err := conn.UpdateModel(req)
	if err != nil {
		return fmt.Errorf("error updating API Gateway v2 model: %s", err)
	}

	return resourceAwsApiGatewayV2ModelRead(d, meta)
}

func resourceAwsApiGatewayV2ModelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 model (%s)", d.Id())
	_, err := conn.DeleteModel(&apigatewayv2.DeleteModelInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		ModelId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 model: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayV2ModelImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/model-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("api_id", parts[0])

	return []*schema.ResourceData{d}, nil
}
