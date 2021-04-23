package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2LocalGatewayRouteTables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2LocalGatewayRouteTablesRead,
		Schema: map[string]*schema.Schema{
			"filter": ec2CustomFiltersSchema(),

			"tags": tagsSchemaComputed(),

			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsEc2LocalGatewayRouteTablesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeLocalGatewayRouteTablesInput{}

	req.Filters = append(req.Filters, buildEC2TagFilterList(
		keyvaluetags.New(d.Get("tags").(map[string]interface{})).Ec2Tags(),
	)...)

	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	var localGatewayRouteTables []*ec2.LocalGatewayRouteTable

	err := conn.DescribeLocalGatewayRouteTablesPages(req, func(page *ec2.DescribeLocalGatewayRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		localGatewayRouteTables = append(localGatewayRouteTables, page.LocalGatewayRouteTables...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error describing EC2 Local Gateway Route Tables: %w", err)
	}

	if len(localGatewayRouteTables) == 0 {
		return fmt.Errorf("no matching EC2 Local Gateway Route Tables found")
	}

	var ids []string

	for _, localGatewayRouteTable := range localGatewayRouteTables {
		if localGatewayRouteTable == nil {
			continue
		}

		ids = append(ids, aws.StringValue(localGatewayRouteTable.LocalGatewayRouteTableId))
	}

	d.SetId(meta.(*AWSClient).region)

	if err := d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
