package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEgressOnlyInternetGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEgressOnlyInternetGatewayCreate,
		Read:   resourceAwsEgressOnlyInternetGatewayRead,
		Update: resourceAwsEgressOnlyInternetGatewayUpdate,
		Delete: resourceAwsEgressOnlyInternetGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEgressOnlyInternetGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resp, err := conn.CreateEgressOnlyInternetGateway(&ec2.CreateEgressOnlyInternetGatewayInput{
		VpcId: aws.String(d.Get("vpc_id").(string)),
	})
	if err != nil {
		return fmt.Errorf("Error creating egress internet gateway: %s", err)
	}

	d.SetId(aws.StringValue(resp.EgressOnlyInternetGateway.EgressOnlyInternetGatewayId))

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), v); err != nil {
			return fmt.Errorf("error adding tags: %s", err)
		}
	}

	return resourceAwsEgressOnlyInternetGatewayRead(d, meta)
}

func resourceAwsEgressOnlyInternetGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	var req = &ec2.DescribeEgressOnlyInternetGatewaysInput{
		EgressOnlyInternetGatewayIds: []*string{aws.String(d.Id())},
	}

	var resp *ec2.DescribeEgressOnlyInternetGatewaysOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.DescribeEgressOnlyInternetGateways(req)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		igw := getEc2EgressOnlyInternetGateway(d.Id(), resp)
		if d.IsNewResource() && igw == nil {
			return resource.RetryableError(fmt.Errorf("Egress Only Internet Gateway (%s) not found.", d.Id()))
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		resp, err = conn.DescribeEgressOnlyInternetGateways(req)
	}

	if err != nil {
		return fmt.Errorf("Error describing egress internet gateway: %s", err)
	}

	igw := getEc2EgressOnlyInternetGateway(d.Id(), resp)
	if igw == nil {
		log.Printf("[Error] Cannot find Egress Only Internet Gateway: %q", d.Id())
		d.SetId("")
		return nil
	}

	if len(igw.Attachments) == 1 && aws.StringValue(igw.Attachments[0].State) == ec2.AttachmentStatusAttached {
		d.Set("vpc_id", igw.Attachments[0].VpcId)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(igw.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func getEc2EgressOnlyInternetGateway(id string, resp *ec2.DescribeEgressOnlyInternetGatewaysOutput) *ec2.EgressOnlyInternetGateway {
	if resp != nil && len(resp.EgressOnlyInternetGateways) > 0 {
		for _, igw := range resp.EgressOnlyInternetGateways {
			if aws.StringValue(igw.EgressOnlyInternetGatewayId) == id {
				return igw
			}
		}
	}
	return nil
}

func resourceAwsEgressOnlyInternetGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Egress Only Internet Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsEgressOnlyInternetGatewayRead(d, meta)
}

func resourceAwsEgressOnlyInternetGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.DeleteEgressOnlyInternetGateway(&ec2.DeleteEgressOnlyInternetGatewayInput{
		EgressOnlyInternetGatewayId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting egress internet gateway: %s", err)
	}

	return nil
}
