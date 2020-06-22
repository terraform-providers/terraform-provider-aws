package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TransitGatewayPeeringAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayPeeringAttachmentAccepterCreate,
		Read:   resourceAwsEc2TransitGatewayPeeringAttachmentAccepterRead,
		Update: resourceAwsEc2TransitGatewayPeeringAttachmentAccepterUpdate,
		Delete: resourceAwsEc2TransitGatewayPeeringAttachmentAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"peer_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2TransitGatewayPeeringAttachmentAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.AcceptTransitGatewayPeeringAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Get("transit_gateway_attachment_id").(string)),
	}

	log.Printf("[DEBUG] Accepting EC2 Transit Gateway Peering Attachment: %s", input)
	output, err := conn.AcceptTransitGatewayPeeringAttachment(input)
	if err != nil {
		return fmt.Errorf("error accepting EC2 Transit Gateway Peering Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayPeeringAttachment.TransitGatewayAttachmentId))

	if err := waitForEc2TransitGatewayPeeringAttachmentAcceptance(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Peering Attachment (%s) availability: %s", d.Id(), err)
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), v); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Peering Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsEc2TransitGatewayPeeringAttachmentAccepterRead(d, meta)
}

func resourceAwsEc2TransitGatewayPeeringAttachmentAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	transitGatewayPeeringAttachment, err := ec2DescribeTransitGatewayPeeringAttachment(conn, d.Id())

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Peering Attachment: %s", err)
	}

	if transitGatewayPeeringAttachment == nil {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	recreationStates := map[string]bool{
		ec2.TransitGatewayAttachmentStateDeleted:   true,
		ec2.TransitGatewayAttachmentStateDeleting:  true,
		ec2.TransitGatewayAttachmentStateFailed:    true,
		ec2.TransitGatewayAttachmentStateFailing:   true,
		ec2.TransitGatewayAttachmentStateRejected:  true,
		ec2.TransitGatewayAttachmentStateRejecting: true,
	}
	if _, ok := recreationStates[aws.StringValue(transitGatewayPeeringAttachment.State)]; ok {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) in state (%s), removing from state", d.Id(), aws.StringValue(transitGatewayPeeringAttachment.State))
		d.SetId("")
		return nil
	}

	transitGatewayID := aws.StringValue(transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)
	transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	d.Set("peer_account_id", transitGatewayPeeringAttachment.RequesterTgwInfo.OwnerId)
	d.Set("peer_region", transitGatewayPeeringAttachment.RequesterTgwInfo.Region)
	d.Set("peer_transit_gateway_id", transitGatewayPeeringAttachment.RequesterTgwInfo.TransitGatewayId)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(transitGatewayPeeringAttachment.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("transit_gateway_attachment_id", transitGatewayPeeringAttachment.TransitGatewayAttachmentId)
	d.Set("transit_gateway_id", transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)

	return nil
}

func resourceAwsEc2TransitGatewayPeeringAttachmentAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Peering Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEc2TransitGatewayPeeringAttachmentAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteTransitGatewayPeeringAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Peering Attachment (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayPeeringAttachment(input)

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Peering Attachment: %s", err)
	}

	if err := waitForEc2TransitGatewayPeeringAttachmentDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Peering Attachment (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
