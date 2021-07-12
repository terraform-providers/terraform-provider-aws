package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsEc2TransitGatewayVpcAttachments() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2TransitGatewayVpcAttachmentsRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsEc2TransitGatewayVpcAttachmentsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{}
	var items []string

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Reading EC2 Transit Gateways: %s", input)
	output, err := conn.DescribeTransitGatewayVpcAttachments(input)

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachments: %s", err)
	}

	if output == nil || len(output.TransitGatewayVpcAttachments) == 0 {
		return errors.New("error reading EC2 Transit Gateway VPC Attachments: no results found")
	}

	for _, transitGatewayVpcAttachment := range output.TransitGatewayVpcAttachments {
		if transitGatewayVpcAttachment != nil {
			if transitGatewayVpcAttachment.Options == nil {
				return fmt.Errorf("error reading EC2 Transit Gateway VPC Attachment (%s): missing options", aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId))
			}

			items = append(items, *transitGatewayVpcAttachment.TransitGatewayAttachmentId)
		}
	}
	d.SetId(resource.UniqueId())
	d.Set("ids", items)
	return nil
}
