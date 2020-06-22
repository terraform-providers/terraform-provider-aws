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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigatewayv2/waiter"
)

func resourceAwsApiGatewayV2VpcLink() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2VpcLinkCreate,
		Read:   resourceAwsApiGatewayV2VpcLinkRead,
		Update: resourceAwsApiGatewayV2VpcLinkUpdate,
		Delete: resourceAwsApiGatewayV2VpcLinkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsApiGatewayV2VpcLinkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateVpcLinkInput{
		Name:             aws.String(d.Get("name").(string)),
		SecurityGroupIds: expandStringSet(d.Get("security_group_ids").(*schema.Set)),
		SubnetIds:        expandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:             keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().Apigatewayv2Tags(),
	}

	log.Printf("[DEBUG] Creating API Gateway v2 VPC Link: %s", req)
	resp, err := conn.CreateVpcLink(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 VPC Link: %s", err)
	}

	d.SetId(aws.StringValue(resp.VpcLinkId))

	if _, err := waiter.VpcLinkAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for API Gateway v2 deployment (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsApiGatewayV2VpcLinkRead(d, meta)
}

func resourceAwsApiGatewayV2VpcLinkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	outputRaw, _, err := waiter.VpcLinkStatus(conn, d.Id())()
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 VPC Link (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 VPC Link (%s): %s", d.Id(), err)
	}

	output := outputRaw.(*apigatewayv2.GetVpcLinkOutput)
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/vpclinks/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("name", output.Name)
	if err := d.Set("security_group_ids", flattenStringSet(output.SecurityGroupIds)); err != nil {
		return fmt.Errorf("error setting security_group_ids: %s", err)
	}
	if err := d.Set("subnet_ids", flattenStringSet(output.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}
	if err := d.Set("tags", keyvaluetags.Apigatewayv2KeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayV2VpcLinkUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	if d.HasChange("name") {
		req := &apigatewayv2.UpdateVpcLinkInput{
			VpcLinkId: aws.String(d.Id()),
			Name:      aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating API Gateway v2 VPC Link: %s", req)
		_, err := conn.UpdateVpcLink(req)
		if err != nil {
			return fmt.Errorf("error updating API Gateway v2 VPC Link (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Apigatewayv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating API Gateway v2 VPC Link (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsApiGatewayV2VpcLinkRead(d, meta)
}

func resourceAwsApiGatewayV2VpcLinkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 VPC Link (%s)", d.Id())
	_, err := conn.DeleteVpcLink(&apigatewayv2.DeleteVpcLinkInput{
		VpcLinkId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 VPC Link (%s): %s", d.Id(), err)
	}

	_, err = waiter.VpcLinkDeleted(conn, d.Id())
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error waiting for API Gateway v2 VPC Link (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
