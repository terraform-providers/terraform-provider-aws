package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TransitGatewayVpcAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayVpcAttachmentAccepterCreate,
		Read:   resourceAwsEc2TransitGatewayVpcAttachmentAccepterRead,
		Update: resourceAwsEc2TransitGatewayVpcAttachmentAccepterUpdate,
		Delete: resourceAwsEc2TransitGatewayVpcAttachmentAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"dns_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
			"transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_default_route_table_association": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"transit_gateway_default_route_table_propagation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2TransitGatewayVpcAttachmentAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.AcceptTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Get("transit_gateway_attachment_id").(string)),
	}

	log.Printf("[DEBUG] Accepting EC2 Transit Gateway VPC Attachment: %s", input)
	output, err := conn.AcceptTransitGatewayVpcAttachment(input)
	if err != nil {
		return fmt.Errorf("error accepting EC2 Transit Gateway VPC Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))
	transitGatewayID := aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayId)

	if err := waitForEc2TransitGatewayVpcAttachmentAcceptance(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway VPC Attachment (%s) availability: %s", d.Id(), err)
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), v); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway VPC Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	if err := ec2TransitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
		return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %s", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
	}

	if err := ec2TransitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
		return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) propagation: %s", d.Id(), aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), err)
	}

	return resourceAwsEc2TransitGatewayVpcAttachmentAccepterRead(d, meta)
}

func resourceAwsEc2TransitGatewayVpcAttachmentAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	transitGatewayVpcAttachment, err := ec2DescribeTransitGatewayVpcAttachment(conn, d.Id())

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachment: %s", err)
	}

	if transitGatewayVpcAttachment == nil {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGatewayVpcAttachment.State) == ec2.TransitGatewayAttachmentStateDeleting || aws.StringValue(transitGatewayVpcAttachment.State) == ec2.TransitGatewayAttachmentStateDeleted {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(transitGatewayVpcAttachment.State))
		d.SetId("")
		return nil
	}

	transitGatewayID := aws.StringValue(transitGatewayVpcAttachment.TransitGatewayId)
	transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if transitGateway.Options == nil {
		return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
	}

	transitGatewayAssociationDefaultRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
	transitGatewayDefaultRouteTableAssociation, err := ec2DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayAssociationDefaultRouteTableID, d.Id())
	if err != nil {
		return fmt.Errorf("error determining EC2 Transit Gateway Attachment (%s) association to Route Table (%s): %s", d.Id(), transitGatewayAssociationDefaultRouteTableID, err)
	}

	transitGatewayPropagationDefaultRouteTableID := aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId)
	transitGatewayDefaultRouteTablePropagation, err := ec2DescribeTransitGatewayRouteTablePropagation(conn, transitGatewayPropagationDefaultRouteTableID, d.Id())
	if err != nil {
		return fmt.Errorf("error determining EC2 Transit Gateway Attachment (%s) propagation to Route Table (%s): %s", d.Id(), transitGatewayPropagationDefaultRouteTableID, err)
	}

	if transitGatewayVpcAttachment.Options == nil {
		return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachment (%s): missing options", d.Id())
	}

	d.Set("dns_support", transitGatewayVpcAttachment.Options.DnsSupport)
	d.Set("ipv6_support", transitGatewayVpcAttachment.Options.Ipv6Support)

	if err := d.Set("subnet_ids", aws.StringValueSlice(transitGatewayVpcAttachment.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(transitGatewayVpcAttachment.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("transit_gateway_attachment_id", aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId))
	d.Set("transit_gateway_default_route_table_association", (transitGatewayDefaultRouteTableAssociation != nil))
	d.Set("transit_gateway_default_route_table_propagation", (transitGatewayDefaultRouteTablePropagation != nil))
	d.Set("transit_gateway_id", aws.StringValue(transitGatewayVpcAttachment.TransitGatewayId))
	d.Set("vpc_id", aws.StringValue(transitGatewayVpcAttachment.VpcId))
	d.Set("vpc_owner_id", aws.StringValue(transitGatewayVpcAttachment.VpcOwnerId))

	return nil
}

func resourceAwsEc2TransitGatewayVpcAttachmentAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChanges("transit_gateway_default_route_table_association", "transit_gateway_default_route_table_propagation") {
		transitGatewayID := d.Get("transit_gateway_id").(string)

		transitGateway, err := ec2DescribeTransitGateway(conn, transitGatewayID)
		if err != nil {
			return fmt.Errorf("error describing EC2 Transit Gateway (%s): %s", transitGatewayID, err)
		}

		if transitGateway.Options == nil {
			return fmt.Errorf("error describing EC2 Transit Gateway (%s): missing options", transitGatewayID)
		}

		if d.HasChange("transit_gateway_default_route_table_association") {
			if err := ec2TransitGatewayRouteTableAssociationUpdate(conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
				return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) association: %s", d.Id(), aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), err)
			}
		}

		if d.HasChange("transit_gateway_default_route_table_propagation") {
			if err := ec2TransitGatewayRouteTablePropagationUpdate(conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
				return fmt.Errorf("error updating EC2 Transit Gateway Attachment (%s) Route Table (%s) propagation: %s", d.Id(), aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway VPC Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEc2TransitGatewayVpcAttachmentAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayVpcAttachment(input)

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway VPC Attachment: %s", err)
	}

	if err := waitForEc2TransitGatewayVpcAttachmentDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway VPC Attachment (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
