package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsEc2TransitGatewayPrefixListReference() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayPrefixListReferenceCreate,
		Read:   resourceAwsEc2TransitGatewayPrefixListReferenceRead,
		Update: resourceAwsEc2TransitGatewayPrefixListReferenceUpdate,
		Delete: resourceAwsEc2TransitGatewayPrefixListReferenceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"blackhole": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"prefix_list_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_attachment_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"transit_gateway_route_table_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceAwsEc2TransitGatewayPrefixListReferenceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.CreateTransitGatewayPrefixListReferenceInput{}

	if v, ok := d.GetOk("blackhole"); ok {
		input.Blackhole = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_attachment_id"); ok {
		input.TransitGatewayAttachmentId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_route_table_id"); ok {
		input.TransitGatewayRouteTableId = aws.String(v.(string))
	}

	output, err := conn.CreateTransitGatewayPrefixListReference(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Prefix List Reference: %w", err)
	}

	if output == nil || output.TransitGatewayPrefixListReference == nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Prefix List Reference: empty response")
	}

	d.SetId(tfec2.TransitGatewayPrefixListReferenceCreateID(aws.StringValue(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.StringValue(output.TransitGatewayPrefixListReference.PrefixListId)))

	if _, err := waiter.TransitGatewayPrefixListReferenceStateCreated(conn, aws.StringValue(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.StringValue(output.TransitGatewayPrefixListReference.PrefixListId)); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Prefix List Reference (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsEc2TransitGatewayPrefixListReferenceRead(d, meta)
}

func resourceAwsEc2TransitGatewayPrefixListReferenceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	transitGatewayPrefixListReference, err := finder.TransitGatewayPrefixListReferenceByID(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteTableIDNotFound) {
		log.Printf("[WARN] EC2 Transit Gateway Prefix List Reference (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Prefix List Reference (%s): %w", d.Id(), err)
	}

	if transitGatewayPrefixListReference == nil {
		log.Printf("[WARN] EC2 Transit Gateway Prefix List Reference (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGatewayPrefixListReference.State) == ec2.TransitGatewayPrefixListReferenceStateDeleting {
		log.Printf("[WARN] EC2 Transit Gateway Prefix List Reference (%s) deleting, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("blackhole", transitGatewayPrefixListReference.Blackhole)
	d.Set("prefix_list_id", transitGatewayPrefixListReference.PrefixListId)
	d.Set("prefix_list_owner_id", transitGatewayPrefixListReference.PrefixListOwnerId)

	if transitGatewayPrefixListReference.TransitGatewayAttachment == nil {
		d.Set("transit_gateway_attachment_id", nil)
	} else {
		d.Set("transit_gateway_attachment_id", transitGatewayPrefixListReference.TransitGatewayAttachment.TransitGatewayAttachmentId)
	}

	d.Set("transit_gateway_route_table_id", transitGatewayPrefixListReference.TransitGatewayRouteTableId)

	return nil
}

func resourceAwsEc2TransitGatewayPrefixListReferenceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.ModifyTransitGatewayPrefixListReferenceInput{}

	if v, ok := d.GetOk("blackhole"); ok {
		input.Blackhole = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_attachment_id"); ok {
		input.TransitGatewayAttachmentId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_route_table_id"); ok {
		input.TransitGatewayRouteTableId = aws.String(v.(string))
	}

	output, err := conn.ModifyTransitGatewayPrefixListReference(input)

	if err != nil {
		return fmt.Errorf("error updating EC2 Transit Gateway Prefix List Reference (%s): %w", d.Id(), err)
	}

	if output == nil || output.TransitGatewayPrefixListReference == nil {
		return fmt.Errorf("error updating EC2 Transit Gateway Prefix List Reference (%s): empty response", d.Id())
	}

	if _, err := waiter.TransitGatewayPrefixListReferenceStateUpdated(conn, aws.StringValue(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.StringValue(output.TransitGatewayPrefixListReference.PrefixListId)); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Prefix List Reference (%s) update: %w", d.Id(), err)
	}

	return resourceAwsEc2TransitGatewayPrefixListReferenceRead(d, meta)
}

func resourceAwsEc2TransitGatewayPrefixListReferenceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	transitGatewayRouteTableID, prefixListID, err := tfec2.TransitGatewayPrefixListReferenceParseID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing EC2 Transit Gateway Prefix List Reference (%s) idenfitier: %w", d.Id(), err)
	}

	input := &ec2.DeleteTransitGatewayPrefixListReferenceInput{
		PrefixListId:               aws.String(prefixListID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	_, err = conn.DeleteTransitGatewayPrefixListReference(input)

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Prefix List Reference (%s): %w", d.Id(), err)
	}

	if _, err := waiter.TransitGatewayPrefixListReferenceStateDeleted(conn, transitGatewayRouteTableID, prefixListID); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Prefix List Reference (%s) deletion: %w", d.Id(), err)
	}

	return nil
}
