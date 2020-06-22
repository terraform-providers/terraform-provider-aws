package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2TransitGatewayVpnAttachment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2TransitGatewayVpnAttachmentRead,

		Schema: map[string]*schema.Schema{
			"tags": tagsSchemaComputed(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vpn_connection_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filter": dataSourceFiltersSchema(),
		},
	}
}

func dataSourceAwsEc2TransitGatewayVpnAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	filters, filtersOk := d.GetOk("filter")
	tags, tagsOk := d.GetOk("tags")
	connectionId, connectionIdOk := d.GetOk("vpn_connection_id")
	transitGatewayId, transitGatewayIdOk := d.GetOk("transit_gateway_id")

	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("resource-type"),
				Values: []*string{aws.String(ec2.TransitGatewayAttachmentResourceTypeVpn)},
			},
		},
	}

	if filtersOk {
		input.Filters = append(input.Filters, buildAwsDataSourceFilters(filters.(*schema.Set))...)
	}
	if tagsOk {
		input.Filters = append(input.Filters, ec2TagFiltersFromMap(tags.(map[string]interface{}))...)
	}
	if connectionIdOk {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name:   aws.String("resource-id"),
			Values: []*string{aws.String(connectionId.(string))},
		})
	}

	if transitGatewayIdOk {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name:   aws.String("transit-gateway-id"),
			Values: []*string{aws.String(transitGatewayId.(string))},
		})
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateway VPN Attachments: %s", input)
	output, err := conn.DescribeTransitGatewayAttachments(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway VPN Attachment: %s", err)
	}

	if output == nil || len(output.TransitGatewayAttachments) == 0 || output.TransitGatewayAttachments[0] == nil {
		return errors.New("error reading EC2 Transit Gateway VPN Attachment: no results found")
	}

	if len(output.TransitGatewayAttachments) > 1 {
		return errors.New("error reading EC2 Transit Gateway VPN Attachment: multiple results found, try adjusting search criteria")
	}

	transitGatewayAttachment := output.TransitGatewayAttachments[0]

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(transitGatewayAttachment.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("transit_gateway_id", aws.StringValue(transitGatewayAttachment.TransitGatewayId))
	d.Set("vpn_connection_id", aws.StringValue(transitGatewayAttachment.ResourceId))

	d.SetId(aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId))

	return nil
}
