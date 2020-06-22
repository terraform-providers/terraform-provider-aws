package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

	log.Printf("[DEBUG] DescribeLocalGatewayRouteTables %s\n", req)
	resp, err := conn.DescribeLocalGatewayRouteTables(req)
	if err != nil {
		return fmt.Errorf("error describing EC2 Local Gateway Route Tables: %w", err)
	}

	if resp == nil || len(resp.LocalGatewayRouteTables) == 0 {
		return fmt.Errorf("no matching Local Gateway Route Table found")
	}

	localgatewayroutetables := make([]string, 0)

	for _, localgatewayroutetable := range resp.LocalGatewayRouteTables {
		localgatewayroutetables = append(localgatewayroutetables, aws.StringValue(localgatewayroutetable.LocalGatewayRouteTableId))
	}

	d.SetId(time.Now().UTC().String())
	if err := d.Set("ids", localgatewayroutetables); err != nil {
		return fmt.Errorf("Error setting local gateway route table ids: %s", err)
	}

	return nil
}
