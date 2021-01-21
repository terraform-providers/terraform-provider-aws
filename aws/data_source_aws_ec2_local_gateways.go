package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	var localGateways []*ec2.LocalGateway

	err := conn.DescribeLocalGatewaysPages(req, func(page *ec2.DescribeLocalGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		localGateways = append(localGateways, page.LocalGateways...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error describing EC2 Local Gateways: %w", err)
	}

	if len(localGateways) == 0 {
		return fmt.Errorf("no matching EC2 Local Gateways found")
	}

	var ids []string

	for _, localGateway := range localGateways {
		if localGateway == nil {
			continue
		}

		ids = append(ids, aws.StringValue(localGateway.LocalGatewayId))
	}

	d.SetId(meta.(*AWSClient).region)

	if err := d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
