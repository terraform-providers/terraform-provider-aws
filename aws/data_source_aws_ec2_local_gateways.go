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

func dataSourceAwsEc2LocalGateways() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2LocalGatewaysRead,
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

func dataSourceAwsEc2LocalGatewaysRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeLocalGatewaysInput{}

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, buildEC2TagFilterList(
			keyvaluetags.New(tags.(map[string]interface{})).Ec2Tags(),
		)...)
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		req.Filters = append(req.Filters, buildEC2CustomFilterList(
			filters.(*schema.Set),
		)...)
	}
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] DescribeLocalGateways %s\n", req)
	resp, err := conn.DescribeLocalGateways(req)
	if err != nil {
		return fmt.Errorf("error describing EC2 Local Gateways: %w", err)
	}

	if resp == nil || len(resp.LocalGateways) == 0 {
		return fmt.Errorf("no matching Local Gateways found")
	}

	localgateways := make([]string, 0)

	for _, localgateway := range resp.LocalGateways {
		localgateways = append(localgateways, aws.StringValue(localgateway.LocalGatewayId))
	}

	d.SetId(time.Now().UTC().String())
	if err := d.Set("ids", localgateways); err != nil {
		return fmt.Errorf("Error setting local gateway ids: %s", err)
	}

	return nil
}
