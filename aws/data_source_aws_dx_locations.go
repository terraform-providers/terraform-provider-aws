package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directconnect/finder"
)

func dataSourceAwsDxLocations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDxLocationsRead,

		Schema: map[string]*schema.Schema{
			"location_codes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsDxLocationsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	locations, err := finder.Locations(conn, &directconnect.DescribeLocationsInput{})

	if err != nil {
		return fmt.Errorf("error reading Direct Connect locations: %w", err)
	}

	var locationCodes []*string

	for _, location := range locations {
		locationCodes = append(locationCodes, location.LocationCode)
	}

	d.SetId(meta.(*AWSClient).region)
	d.Set("location_codes", aws.StringValueSlice(locationCodes))

	return nil
}
