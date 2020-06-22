package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2CoipPool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2CoipPoolRead,

		Schema: map[string]*schema.Schema{
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"pool_cidrs": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),

			"filter": ec2CustomFiltersSchema(),
		},
	}
}

func dataSourceAwsEc2CoipPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeCoipPoolsInput{}

	if v, ok := d.GetOk("pool_id"); ok {
		req.PoolIds = []*string{aws.String(v.(string))}
	}

	filters := map[string]string{}

	if v, ok := d.GetOk("local_gateway_route_table_id"); ok {
		filters["coip-pool.local-gateway-route-table-id"] = v.(string)
	}

	req.Filters = buildEC2AttributeFilterList(filters)

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, buildEC2TagFilterList(
			keyvaluetags.New(tags.(map[string]interface{})).Ec2Tags(),
		)...)
	}

	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] Reading AWS COIP Pool: %s", req)
	resp, err := conn.DescribeCoipPools(req)
	if err != nil {
		return fmt.Errorf("error describing EC2 COIP Pools: %w", err)
	}
	if resp == nil || len(resp.CoipPools) == 0 {
		return fmt.Errorf("no matching COIP Pool found")
	}
	if len(resp.CoipPools) > 1 {
		return fmt.Errorf("multiple Coip Pools matched; use additional constraints to reduce matches to a single COIP Pool")
	}

	coip := resp.CoipPools[0]

	d.SetId(aws.StringValue(coip.PoolId))

	d.Set("local_gateway_route_table_id", coip.LocalGatewayRouteTableId)

	if err := d.Set("pool_cidrs", aws.StringValueSlice(coip.PoolCidrs)); err != nil {
		return fmt.Errorf("error setting pool_cidrs: %s", err)
	}

	d.Set("pool_id", coip.PoolId)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(coip.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
