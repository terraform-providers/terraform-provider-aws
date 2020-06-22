package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2LocalGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2LocalGatewayRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"filter": ec2CustomFiltersSchema(),

			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEc2LocalGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeLocalGatewaysInput{}

	if v, ok := d.GetOk("id"); ok {
		req.LocalGatewayIds = []*string{aws.String(v.(string))}
	}

	req.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"state": d.Get("state").(string),
		},
	)

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

	log.Printf("[DEBUG] Reading AWS LOCAL GATEWAY: %s", req)
	resp, err := conn.DescribeLocalGateways(req)
	if err != nil {
		return fmt.Errorf("error describing EC2 Local Gateways: %w", err)
	}
	if resp == nil || len(resp.LocalGateways) == 0 {
		return fmt.Errorf("no matching Local Gateway found")
	}
	if len(resp.LocalGateways) > 1 {
		return fmt.Errorf("multiple Local Gateways matched; use additional constraints to reduce matches to a single Local Gateway")
	}

	localGateway := resp.LocalGateways[0]

	d.SetId(aws.StringValue(localGateway.LocalGatewayId))
	d.Set("outpost_arn", localGateway.OutpostArn)
	d.Set("owner_id", localGateway.OwnerId)
	d.Set("state", localGateway.State)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(localGateway.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
