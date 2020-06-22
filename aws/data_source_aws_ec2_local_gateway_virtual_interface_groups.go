package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2LocalGatewayVirtualInterfaceGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsRead,

		Schema: map[string]*schema.Schema{
			"filter": ec2CustomFiltersSchema(),
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"local_gateway_virtual_interface_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput{}

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

	output, err := conn.DescribeLocalGatewayVirtualInterfaceGroups(input)

	if err != nil {
		return fmt.Errorf("error describing EC2 Virtual Interface Groups: %w", err)
	}

	if output == nil || len(output.LocalGatewayVirtualInterfaceGroups) == 0 {
		return fmt.Errorf("no matching Virtual Interface Group found")
	}

	var ids, localGatewayVirtualInterfaceIds []*string

	for _, group := range output.LocalGatewayVirtualInterfaceGroups {
		ids = append(ids, group.LocalGatewayVirtualInterfaceGroupId)
		localGatewayVirtualInterfaceIds = append(localGatewayVirtualInterfaceIds, group.LocalGatewayVirtualInterfaceIds...)
	}

	d.SetId(resource.UniqueId())

	if err := d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	if err := d.Set("local_gateway_virtual_interface_ids", localGatewayVirtualInterfaceIds); err != nil {
		return fmt.Errorf("error setting local_gateway_virtual_interface_ids: %w", err)
	}

	return nil
}
