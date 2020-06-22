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

func dataSourceAwsEc2CoipPools() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2CoipPoolsRead,
		Schema: map[string]*schema.Schema{
			"filter": ec2CustomFiltersSchema(),

			"tags": tagsSchemaComputed(),

			"pool_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsEc2CoipPoolsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeCoipPoolsInput{}

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

	log.Printf("[DEBUG] DescribeCoipPools %s\n", req)
	resp, err := conn.DescribeCoipPools(req)
	if err != nil {
		return fmt.Errorf("error describing EC2 COIP Pools: %w", err)
	}

	if resp == nil || len(resp.CoipPools) == 0 {
		return fmt.Errorf("no matching Coip Pool found")
	}

	coippools := make([]string, 0)

	for _, coippool := range resp.CoipPools {
		coippools = append(coippools, aws.StringValue(coippool.PoolId))
	}

	d.SetId(time.Now().UTC().String())
	if err := d.Set("pool_ids", coippools); err != nil {
		return fmt.Errorf("Error setting coip pool ids: %s", err)
	}

	return nil
}
