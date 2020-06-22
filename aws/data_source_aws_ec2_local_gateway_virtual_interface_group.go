package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2LocalGatewayVirtualInterfaceGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2LocalGatewayVirtualInterfaceGroupRead,

		Schema: map[string]*schema.Schema{
			"filter": ec2CustomFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"local_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

func dataSourceAwsEc2LocalGatewayVirtualInterfaceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput{}

	if v, ok := d.GetOk("id"); ok {
		input.LocalGatewayVirtualInterfaceGroupIds = []*string{aws.String(v.(string))}
	}

	input.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"local-gateway-id": d.Get("local_gateway_id").(string),
		},
	)

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
		return fmt.Errorf("error describing EC2 Local Gateway Virtual Interface Groups: %w", err)
	}

	if output == nil || len(output.LocalGatewayVirtualInterfaceGroups) == 0 {
		return fmt.Errorf("no matching EC2 Local Gateway Virtual Interface Group found")
	}

	if len(output.LocalGatewayVirtualInterfaceGroups) > 1 {
		return fmt.Errorf("multiple EC2 Local Gateway Virtual Interface Groups matched; use additional constraints to reduce matches to a single EC2 Local Gateway Virtual Interface Group")
	}

	localGatewayVirtualInterfaceGroup := output.LocalGatewayVirtualInterfaceGroups[0]

	d.SetId(aws.StringValue(localGatewayVirtualInterfaceGroup.LocalGatewayVirtualInterfaceGroupId))
	d.Set("local_gateway_id", localGatewayVirtualInterfaceGroup.LocalGatewayId)
	d.Set("local_gateway_virtual_interface_group_id", localGatewayVirtualInterfaceGroup.LocalGatewayVirtualInterfaceGroupId)

	if err := d.Set("local_gateway_virtual_interface_ids", aws.StringValueSlice(localGatewayVirtualInterfaceGroup.LocalGatewayVirtualInterfaceIds)); err != nil {
		return fmt.Errorf("error setting local_gateway_virtual_interface_ids: %w", err)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(localGatewayVirtualInterfaceGroup.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
