package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsVpcEndpointPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcEndpointPolicyPut,
		Read:   resourceAwsVpcEndpointPolicyRead,
		Update: resourceAwsVpcEndpointPolicyPut,
		Delete: resourceAwsVpcEndpointPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceAwsVpcEndpointPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpceRaw, state, err := vpcEndpointStateRefresh(conn, d.Id())()
	if err != nil && state != "failed" {
		return fmt.Errorf("error reading VPC Endpoint Policy (%s): %w", d.Id(), err)
	}

	terminalStates := map[string]bool{
		"deleted":  true,
		"deleting": true,
		"failed":   true,
		"expired":  true,
		"rejected": true,
	}
	if _, ok := terminalStates[state]; ok {
		log.Printf("[WARN] VPC Endpoint Policy (%s) in state (%s), removing from state", d.Id(), state)
		d.SetId("")
		return nil
	}

	vpce := vpceRaw.(*ec2.VpcEndpoint)
	policy, err := structure.NormalizeJsonString(aws.StringValue(vpce.PolicyDocument))
	if err != nil {
		return fmt.Errorf("policy contains an invalid JSON: %w", err)
	}
	d.Set("vpc_endpoint_id", d.Id())
	d.Set("policy", policy)

	return nil
}

func resourceAwsVpcEndpointPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	endpointID := d.Get("vpc_endpoint_id").(string)
	req := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(endpointID),
	}

	policy, err := structure.NormalizeJsonString(d.Get("policy"))
	if err != nil {
		return fmt.Errorf("policy contains an invalid JSON: %w", err)
	}

	if policy == "" {
		req.ResetPolicy = aws.Bool(true)
	} else {
		req.PolicyDocument = aws.String(policy)
	}

	log.Printf("[DEBUG] Updating VPC Endpoint Policy: %#v", req)
	if _, err := conn.ModifyVpcEndpoint(req); err != nil {
		return fmt.Errorf("Error updating VPC Endpoint Policy: %w", err)
	}
	d.SetId(endpointID)

	return resourceAwsVpcEndpointPolicyRead(d, meta)
}

func resourceAwsVpcEndpointPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(d.Id()),
		ResetPolicy:   aws.Bool(true),
	}

	log.Printf("[DEBUG] Resetting VPC Endpoint Policy: %#v", req)
	if _, err := conn.ModifyVpcEndpoint(req); err != nil {
		return fmt.Errorf("Error Resetting VPC Endpoint Policy: %w", err)
	}

	return nil
}
