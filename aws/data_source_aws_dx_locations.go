package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsDxLocations() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDxLocationsRead,

		Schema: map[string]*schema.Schema{
			"location_codes": {
				Type:     schema.TypeList,
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
		return fmt.Errorf("error listing Direct Connect locations: %s", err)
	}

	d.SetId(time.Now().UTC().String())

	locationCodes := []*string{}
	for _, location := range resp.Locations {
		locationCodes = append(locationCodes, location.LocationCode)
	}
	err = d.Set("location_codes", flattenStringList(locationCodes))
	if err != nil {
		return fmt.Errorf("error setting location_codes: %s", err)
	}

	return nil
}
