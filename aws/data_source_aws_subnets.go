package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsSubnets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSubnetsRead,
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

func dataSourceAwsSubnetsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeSubnetsInput{}

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
		req.Filters = nil
	}

	log.Printf("[DEBUG] DescribeSubnets %s\n", req)
	resp, err := conn.DescribeSubnets(req)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.Subnets) == 0 {
		return fmt.Errorf("no matching subnets found")
	}

	subnets := make([]string, 0)

	for _, subnet := range resp.Subnets {
		subnets = append(subnets, *subnet.SubnetId)
	}

	d.SetId(meta.(*AWSClient).region)
	d.Set("ids", subnets)

	return nil
}
