package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsVpcPeeringConnections() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsVpcPeeringConnectionsRead,

		Schema: map[string]*schema.Schema{
			"filter": ec2CustomFiltersSchema(),
			"tags":   tagsSchemaComputed(),
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsVpcPeeringConnectionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeVpcPeeringConnectionsInput{}

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

	resp, err := conn.DescribeVpcPeeringConnections(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.VpcPeeringConnections) == 0 {
		return fmt.Errorf("no matching VPC peering connections found")
	}

	var ids []string
	for _, pcx := range resp.VpcPeeringConnections {
		ids = append(ids, aws.StringValue(pcx.VpcPeeringConnectionId))
	}

	d.SetId(meta.(*AWSClient).region)

	err = d.Set("ids", ids)
	if err != nil {
		return err
	}

	return nil
}
