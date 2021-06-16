package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2TransitGatewayRouteTables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2TransitGatewayRouteTablesRead,

		Schema: map[string]*schema.Schema{
			"filter": ec2CustomFiltersSchema(),

			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsEc2TransitGatewayRouteTablesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DescribeTransitGatewayRouteTablesInput{}

	input.Filters = append(input.Filters, buildEC2TagFilterList(
		keyvaluetags.New(d.Get("tags").(map[string]interface{})).Ec2Tags(),
	)...)

	input.Filters = append(input.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	var transitGatewayRouteTables []*ec2.TransitGatewayRouteTable

	err := conn.DescribeTransitGatewayRouteTablesPages(input, func(page *ec2.DescribeTransitGatewayRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		transitGatewayRouteTables = append(transitGatewayRouteTables, page.TransitGatewayRouteTables...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error describing EC2 Transit Gateway Route Tables: %w", err)
	}

	if len(transitGatewayRouteTables) == 0 {
		return fmt.Errorf("no matching EC2 Transit Gateway Route Tables found")
	}

	var ids []string

	for _, transitGatewayRouteTable := range transitGatewayRouteTables {
		if transitGatewayRouteTable == nil {
			continue
		}

		ids = append(ids, aws.StringValue(transitGatewayRouteTable.TransitGatewayRouteTableId))
	}

	d.SetId(meta.(*AWSClient).region)

	if err = d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
