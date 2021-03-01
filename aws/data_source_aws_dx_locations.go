package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	log.Printf("[DEBUG] Listing Direct Connect locations")
	resp, err := conn.DescribeLocations(&directconnect.DescribeLocationsInput{})
	if err != nil {
		return fmt.Errorf("error listing Direct Connect locations: %w", err)
	}

	d.SetId(meta.(*AWSClient).region)

	locationCodes := []*string{}
	for _, location := range resp.Locations {
		locationCodes = append(locationCodes, location.LocationCode)
	}
	err = d.Set("location_codes", flattenStringList(locationCodes))
	if err != nil {
		return fmt.Errorf("error setting location_codes: %w", err)
	}

	return nil
}
