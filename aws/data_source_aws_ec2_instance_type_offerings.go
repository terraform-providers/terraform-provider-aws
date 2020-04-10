package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsEc2InstanceTypeOfferings() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEc2InstanceTypeOfferingsRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"instance_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"location_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.LocationTypeAvailabilityZone,
					ec2.LocationTypeAvailabilityZoneId,
					ec2.LocationTypeRegion,
				}, false),
			},
		},
	}
}

func dataSourceAwsEc2InstanceTypeOfferingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DescribeInstanceTypeOfferingsInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = buildAwsDataSourceFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("location_type"); ok {
		input.LocationType = aws.String(v.(string))
	}

	var instanceTypes []string

	for {
		output, err := conn.DescribeInstanceTypeOfferings(input)

		if err != nil {
			return fmt.Errorf("error reading EC2 Instance Type Offerings: %w", err)
		}

		if output == nil {
			break
		}

		for _, instanceTypeOffering := range output.InstanceTypeOfferings {
			if instanceTypeOffering == nil {
				continue
			}

			instanceTypes = append(instanceTypes, aws.StringValue(instanceTypeOffering.InstanceType))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := d.Set("instance_types", instanceTypes); err != nil {
		return fmt.Errorf("error setting instance_types: %s", err)
	}

	d.SetId(resource.UniqueId())

	return nil
}
