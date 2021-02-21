package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsEc2CarrierGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2CarrierGatewayCreate,
		Read:   resourceAwsEc2CarrierGatewayRead,
		Update: resourceAwsEc2CarrierGatewayUpdate,
		Delete: resourceAwsEc2CarrierGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsEc2CarrierGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.CreateCarrierGatewayInput{
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), "carrier-gateway"),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Carrier Gateway: %s", input)
	output, err := conn.CreateCarrierGateway(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Carrier Gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.CarrierGateway.CarrierGatewayId))

	_, err = waiter.CarrierGatewayAvailable(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Carrier Gateway (%s) to become available: %w", d.Id(), err)
	}

	return resourceAwsEc2CarrierGatewayRead(d, meta)
}

func resourceAwsEc2CarrierGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	carrierGateway, err := finder.CarrierGatewayByID(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidCarrierGatewayIDNotFound) {
		log.Printf("[WARN] EC2 Carrier Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Carrier Gateway (%s): %w", d.Id(), err)
	}

	if carrierGateway == nil || aws.StringValue(carrierGateway.State) == ec2.CarrierGatewayStateDeleted {
		log.Printf("[WARN] EC2 Carrier Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("carrier-gateway/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("owner_id", carrierGateway.OwnerId)
	d.Set("vpc_id", carrierGateway.VpcId)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(carrierGateway.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsEc2CarrierGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Carrier Gateway (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsEc2CarrierGatewayRead(d, meta)
}

func resourceAwsEc2CarrierGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[INFO] Deleting EC2 Carrier Gateway (%s)", d.Id())
	_, err := conn.DeleteCarrierGateway(&ec2.DeleteCarrierGatewayInput{
		CarrierGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidCarrierGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Carrier Gateway (%s): %w", d.Id(), err)
	}

	_, err = waiter.CarrierGatewayDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Carrier Gateway (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
