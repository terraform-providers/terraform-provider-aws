package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsVpnGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpnGatewayAttachmentCreate,
		Read:   resourceAwsVpnGatewayAttachmentRead,
		Delete: resourceAwsVpnGatewayAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpn_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsVpnGatewayAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcId := d.Get("vpc_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)

	input := &ec2.AttachVpnGatewayInput{
		VpcId:        aws.String(vpcId),
		VpnGatewayId: aws.String(vgwId),
	}

	log.Printf("[DEBUG] Creating VPN Gateway Attachment: %s", input)
	_, err := conn.AttachVpnGateway(input)

	if err != nil {
		return fmt.Errorf("error creating VPN Gateway (%s) Attachment (%s): %w", vgwId, vpcId, err)
	}

	d.SetId(tfec2.VpnGatewayVpcAttachmentCreateID(vgwId, vpcId))

	_, err = waiter.VpnGatewayVpcAttachmentAttached(conn, vgwId, vpcId)

	if err != nil {
		return fmt.Errorf("error waiting for VPN Gateway (%s) Attachment (%s) to become attached: %w", vgwId, vpcId, err)
	}

	return resourceAwsVpnGatewayAttachmentRead(d, meta)
}

func resourceAwsVpnGatewayAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcId := d.Get("vpc_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)

	vpcAttachment, err := finder.VpnGatewayVpcAttachment(conn, vgwId, vpcId)

	if isAWSErr(err, tfec2.InvalidVpnGatewayIDNotFound, "") {
		log.Printf("[WARN] VPN Gateway (%s) Attachment (%s) not found, removing from state", vgwId, vpcId)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPN Gateway (%s) Attachment (%s): %w", vgwId, vpcId, err)
	}

	if vpcAttachment == nil || aws.StringValue(vpcAttachment.State) == ec2.AttachmentStatusDetached {
		log.Printf("[WARN] VPN Gateway (%s) Attachment (%s) not found, removing from state", vgwId, vpcId)
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsVpnGatewayAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcId := d.Get("vpc_id").(string)
	vgwId := d.Get("vpn_gateway_id").(string)

	log.Printf("[INFO] Deleting VPN Gateway (%s) Attachment (%s)", vgwId, vpcId)
	_, err := conn.DetachVpnGateway(&ec2.DetachVpnGatewayInput{
		VpcId:        aws.String(vpcId),
		VpnGatewayId: aws.String(vgwId),
	})

	if isAWSErr(err, tfec2.InvalidVpnGatewayAttachmentNotFound, "") || isAWSErr(err, tfec2.InvalidVpnGatewayIDNotFound, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPN Gateway (%s) Attachment (%s): %w", vgwId, vpcId, err)
	}

	_, err = waiter.VpnGatewayVpcAttachmentDetached(conn, vgwId, vpcId)

	if err != nil {
		return fmt.Errorf("error waiting for VPN Gateway (%s) Attachment (%s) to become detached: %w", vgwId, vpcId, err)
	}

	return nil
}
